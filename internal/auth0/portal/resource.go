package portal

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

const resourceType = "auth0_portal"

var portalSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,99}$`)

// NewResource will return a new auth0_portal resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPortal,
		ReadContext:   readPortal,
		UpdateContext: updatePortal,
		DeleteContext: deletePortal,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validatePortal,
		Description:   "With this resource, you can manage a Universal Portal. (EA only)",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the portal.",
			},
			"name": requiredString("The name of the portal.", 1, 150),
			"slug": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(portalSlugPattern, "must be a valid portal slug"),
				Description:  "The unique slug of the portal.",
			},
			"client": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The client used by the portal.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"client_secret_post": {
						Type:         schema.TypeList,
						Optional:     true,
						MaxItems:     1,
						ExactlyOneOf: []string{"client.0.client_secret_post"},
						Description:  "Client credentials using client_secret_post authentication.",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"client_id": requiredString("The ID of the client.", 5, 64),
							"client_secret": {
								Type:          schema.TypeString,
								Optional:      true,
								Sensitive:     true,
								ConflictsWith: []string{"client.0.client_secret_post.0.client_secret_wo"},
								Description:   "The client secret. This value is stored in Terraform state.",
							},
							"client_secret_wo": {
								Type:          schema.TypeString,
								Optional:      true,
								Sensitive:     true,
								WriteOnly:     true,
								ConflictsWith: []string{"client.0.client_secret_post.0.client_secret"},
								RequiredWith:  []string{"client.0.client_secret_post.0.client_secret_wo_version"},
								Description: "The write-only client secret. This value is not stored in Terraform state. " +
									"It is also sent when any other attribute in `client_secret_post` changes.",
							},
							"client_secret_wo_version": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Version of the write-only client secret. Increment this value to send a new secret.",
							},
						}},
					},
				}},
			},
			"custom_code": optionalBlock("Custom code for the portal.", map[string]*schema.Schema{
				"head": optionalString("Code added to the page head.", 10000),
			}),
			"navigation": optionalBlock("Navigation for the portal.", map[string]*schema.Schema{
				"sidebar": optionalBlock("Sidebar navigation for the portal.", map[string]*schema.Schema{
					"component": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Sidebar components in display order. Only V2 links are supported.",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"external_link_v2": optionalBlock("A V2 link to an external URL.", map[string]*schema.Schema{
								"label": requiredString("The label of the link.", 1, 50),
								"href":  requiredString("The URL of the link.", 1, 200),
								"icon":  optionalString("The icon of the link.", 50),
							}),
							"internal_link_v2": optionalBlock("A V2 link to a portal page.", map[string]*schema.Schema{
								"label":     requiredString("The label of the link.", 1, 50),
								"page_slug": requiredString("The slug of the linked page.", 1, 50),
								"icon":      optionalString("The icon of the link.", 50),
							}),
						}},
					},
				}),
			}),
			"pages": optionalBlock("Pages of the portal.", map[string]*schema.Schema{
				"default": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringMatch(portalSlugPattern, "must be a valid portal slug"),
					Description:  "The slug of the default page.",
				},
				"standard": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Standard pages in display order.",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"title": requiredString("The title of the page.", 1, 150),
						"slug": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(portalSlugPattern, "must be a valid portal slug"),
							Description:  "The slug of the page.",
						},
						"component": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Page components in display order.",
							Elem:        &schema.Resource{Schema: componentTypes(true)},
						},
					}},
				},
			}),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the portal was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the portal was last updated.",
			},
		},
	}
}

func componentTypes(includeStructureSection bool) map[string]*schema.Schema {
	types := map[string]*schema.Schema{
		"form": optionalBlock("A form component.", map[string]*schema.Schema{
			"form_id":            requiredString("The ID of the form.", 1, 50),
			"completion_message": requiredString("The message shown after form completion.", 1, 200),
		}),
		"typography_heading": optionalBlock("A heading component.", map[string]*schema.Schema{
			"title":       requiredString("The heading text.", 1, 50),
			"description": optionalString("The description below the heading.", 200),
		}),
		"typography_rich_text": optionalBlock("A rich text component.", map[string]*schema.Schema{
			"content": requiredString("The HTML content.", 1, 10000),
		}),
		"media_image": optionalBlock("An image component.", map[string]*schema.Schema{
			"src": requiredString("The URL of the image.", 1, 500),
			"position": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"left", "center", "right"}, false),
				Description:  "The position of the image.",
			},
			"height": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The height of the image.",
			},
		}),
		"structure_separator": optionalBlock("A separator component.", map[string]*schema.Schema{
			"variant": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"dashed", "none", "solid"}, false),
				Description:  "The separator style.",
			},
			"text": optionalString("The text of the separator.", 50),
		}),
		"my_organization_domain_table":      optionalBlockWithEmptySchema("An organization domain table component."),
		"my_organization_member_management": optionalBlockWithEmptySchema("An organization member management component."),
		"my_organization_details_edit":      optionalBlockWithEmptySchema("An organization details edit component."),
		"my_account_passkey_management":     optionalBlockWithEmptySchema("A passkey management component."),
		"my_account_mfa_management":         optionalBlockWithEmptySchema("An MFA management component."),
	}

	if includeStructureSection {
		types["structure_section"] = optionalBlock("A section containing up to 20 components.", map[string]*schema.Schema{
			"variant": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"card", "none"}, false),
				Description:  "The section style. Can be `card` or `none`.",
			},
			"title":       optionalString("The title of the section.", 50),
			"description": optionalString("The description of the section.", 200),
			"child": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    20,
				Description: "Components in the section. Sections cannot be nested.",
				Elem:        &schema.Resource{Schema: componentTypes(false)},
			},
		})
	}

	return types
}

func optionalBlock(description string, fields map[string]*schema.Schema) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: description,
		Elem:        &schema.Resource{Schema: fields},
	}
}

func optionalBlockWithEmptySchema(description string) *schema.Schema {
	return optionalBlock(description, map[string]*schema.Schema{})
}

func requiredString(description string, minLen, maxLen int) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(minLen, maxLen),
		Description:  description,
	}
}

func optionalString(description string, maxLen int) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, maxLen),
		Description:  description,
	}
}

func createPortal(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	portal, err := apiv3.Portals.Create(ctx, expandPortalCreate(data))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(portal.GetID())

	return readPortal(ctx, data, meta)
}

func readPortal(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	portal, err := apiv3.Portals.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError(resourceType, data, err)
	}

	return flattenPortal(data, portal)
}

func updatePortal(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	if _, err := apiv3.Portals.Update(ctx, data.Id(), expandPortalUpdate(data)); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readPortal(ctx, data, meta)
}

func deletePortal(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	if err := apiv3.Portals.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

// validatePortal performs plan-time validation for UX only, as the API enforces the same rules.
// It is needed because ExactlyOneOf only accepts fixed-index paths, so it cannot check
// each entry of a component list. Values that are unknown at plan time are skipped, as they cannot be judged yet.
func validatePortal(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	raw := diff.GetRawConfig()
	if err := validateSidebarComponents(raw); err != nil {
		return err
	}

	return validatePages(raw)
}

// validateSidebarComponents ensures each sidebar component sets exactly one link block,
// as the API accepts one link type per component.
func validateSidebarComponents(raw cty.Value) error {
	navigation := value.FirstBlock(value.Attr(raw, "navigation"))
	sidebar := value.FirstBlock(value.Attr(navigation, "sidebar"))
	component := value.Attr(sidebar, "component")

	// Validate exactly one of external_link_v2 or internal_link_v2 is set.
	return validateExactlyOneOf(component, "navigation.sidebar.component")
}

// validatePages ensures each page component and section child sets exactly one component type block,
// as the API accepts one component type per entry.
func validatePages(raw cty.Value) error {
	pages := value.FirstBlock(value.Attr(raw, "pages"))
	for i, page := range blocks(value.Attr(pages, "standard")) {
		components := value.Attr(page, "component")
		path := fmt.Sprintf("pages.standard[%d].component", i)
		// Validate exactly one of component types is set.
		if err := validateExactlyOneOf(components, path); err != nil {
			return err
		}

		// Validate structure_section.
		for j, component := range blocks(components) {
			section := value.FirstBlock(value.Attr(component, "structure_section"))
			childPath := fmt.Sprintf("%s[%d].structure_section.child", path, j)
			// Validate exactly one of child component types is set.
			if err := validateExactlyOneOf(value.Attr(section, "child"), childPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateExactlyOneOf reports the first entry of a block list that sets zero or more than one types.
func validateExactlyOneOf(list cty.Value, path string) error {
	for i, block := range blocks(list) {
		if !value.IsSet(block) {
			continue
		}

		var names, set []string
		unknown := false
		for name, typeBlock := range block.AsValueMap() {
			names = append(names, name)
			switch {
			case !typeBlock.IsKnown():
				unknown = true
			case !typeBlock.IsNull() && typeBlock.LengthInt() > 0:
				set = append(set, name)
			}
		}

		// An unknown component type block can still resolve to the single set block, but two set blocks are already invalid.
		if len(set) == 1 || (unknown && len(set) == 0) {
			continue
		}

		sort.Strings(names)
		sort.Strings(set)
		found := "none"
		if len(set) > 0 {
			found = "`" + strings.Join(set, "`, `") + "`"
		}

		return fmt.Errorf("`%s[%d]` must set exactly one of `%s`, but sets %s",
			path, i, strings.Join(names, "`, `"), found)
	}

	return nil
}

// blocks returns the elements of a nested block list, keeping unknown elements so that indexes match the config,
// or nil when the list itself is not set.
func blocks(list cty.Value) []cty.Value {
	if !value.IsSet(list) {
		return nil
	}

	return list.AsValueSlice()
}
