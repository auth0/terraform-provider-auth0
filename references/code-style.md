# Code Style Reference

## Resource Package Structure

Every resource and data source lives in its own package under `internal/auth0/<name>/`:

```
internal/auth0/<name>/
├── resource.go      # NewResource() → *schema.Resource with CRUD funcs and schema
├── expand.go        # expandX(data *schema.ResourceData) *management.X
├── flatten.go       # flattenX(data *schema.ResourceData, x *management.X) error
├── data_source.go   # NewDataSource() — only if a data source exists for this resource
└── resource_test.go # package <name>_test — acceptance tests
```

## Resource Skeleton

```go
func NewResource() *schema.Resource {
    return &schema.Resource{
        CreateContext: createFoo,
        ReadContext:   readFoo,
        UpdateContext: updateFoo,
        DeleteContext: deleteFoo,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: resourceSchema,
    }
}

func createFoo(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
    api := meta.(*config.Config).GetAPI()
    foo := expandFoo(data)
    if err := api.Foo.Create(ctx, foo); err != nil {
        return diag.FromErr(err)
    }
    data.SetId(foo.GetID())
    return readFoo(ctx, data, meta)  // create delegates to read
}

func readFoo(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
    api := meta.(*config.Config).GetAPI()
    foo, err := api.Foo.Read(ctx, data.Id())
    if err != nil {
        return internalError.HandleReadAPIError("Foo", data, err)
    }
    return diag.FromErr(flattenFoo(data, foo))
}
```

## Expand / Flatten Pattern

```go
// expand.go — Terraform state → API request body
func expandFoo(data *schema.ResourceData) *management.Foo {
    cfg := data.GetRawConfig()
    return &management.Foo{
        Name:    value.String(cfg.GetAttr("name")),
        Enabled: value.Bool(cfg.GetAttr("enabled")),
    }
}

// flatten.go — API response → Terraform state
func flattenFoo(data *schema.ResourceData, foo *management.Foo) error {
    result := multierror.Append(
        data.Set("name", foo.GetName()),
        data.Set("enabled", foo.GetEnabled()),
    )
    return result.ErrorOrNil()
}
```

Use `data.GetRawConfig()` + `value.*()` helpers in expand when you need to distinguish "field not set" from "field set to zero value." `data.Get("field")` returns the zero value for unset fields, which causes spurious diffs.

## Naming Conventions

- **Package names:** lowercase, matching the Auth0 resource name — `action`, `client`, `role`, `connection`.
- **Exported constructors:** `NewResource()` and `NewDataSource()`.
- **CRUD functions:** unexported, camelCase — `createFoo`, `readFoo`, `updateFoo`, `deleteFoo`.
- **Schema variable:** package-level `var resourceSchema = map[string]*schema.Schema{...}` or `dataSourceSchema`.
- **Test config constants:** `const testAccFooCreate = ...`, `const testAccFooUpdate = ...`.
- **Test functions:** `TestAcc<ResourceName>` — always prefixed with `TestAcc`.

## Schema Descriptions

All schema fields must have a `Description`. The `godot` linter enforces this in CI:

```go
// ✅ Good — description present, ends with period, starts with capital
"name": {
    Type:        schema.TypeString,
    Required:    true,
    Description: "Name of the action.",
},

// ❌ Bad — missing description
"name": {
    Type:     schema.TypeString,
    Required: true,
},

// ❌ Bad — does not end with period
"name": {
    Type:        schema.TypeString,
    Required:    true,
    Description: "Name of the action",
},
```

Descriptions support Markdown (`` `code` ``, `**bold**`, links) because `schema.DescriptionKind = schema.StringMarkdown` is set in `main.go`.

## Error Handling

```go
// In Read functions — removes from state with a warning on 404
return internalError.HandleReadAPIError("ResourceType", data, err)

// In Create/Update/Delete — returns the error directly
return diag.FromErr(err)

// In flatten — collect multiple set errors
result := multierror.Append(
    data.Set("field1", val1),
    data.Set("field2", val2),
)
return result.ErrorOrNil()
```

Never use `internalError.HandleAPIError` in `Read` functions — it is a legacy function that silently removes resources from state without a warning.

## API Client Access

```go
// v1 SDK — most resources
api := meta.(*config.Config).GetAPI()

// v3 SDK — newer resources
apiv3 := meta.(*config.Config).GetAPIv3()
```

Always use `config.Config.GetAPI()` or `GetAPIv3()`. Never call `management.New()` directly.

## Change Detection in Updates

```go
func updateFoo(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
    api := meta.(*config.Config).GetAPI()
    foo := expandFoo(data)
    if err := api.Foo.Update(ctx, data.Id(), foo); err != nil {
        return diag.FromErr(err)
    }
    return readFoo(ctx, data, meta)
}
```

Use `data.HasChange("field")` when you need to conditionally include an optional field in the update struct to avoid sending unintended zero values to the API.

## Importer

Every new resource must include an importer unless it genuinely cannot be imported (document why if omitted):

```go
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```
