package form

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_form resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createForm,
		ReadContext:   readForm,
		UpdateContext: updateForm,
		DeleteContext: deleteForm,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can create and manage Forms for a tenant.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the form.",
			},
			"translations": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "Translations of the form.",
			},
			"style": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "Style specific configuration for the form.",
			},
			"start": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "Input setup of the form.",
			},
			"nodes": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "Nodes of the form.",
			},
			"ending": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "Submission configuration of the form.",
				DefaultFunc: func() (interface{}, error) {
					return `{"resume_flow":true}`, nil
				},
			},
			"languages": formLanguageSchema,
			"messages":  formMessagesSchema,
		},
	}
}

var formLanguageSchema = &schema.Schema{
	Type:        schema.TypeList,
	Optional:    true,
	Description: "Language specific configuration for the form.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"default": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default language for the form.",
			},
			"primary": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Primary language for the form.",
			},
		},
	},
}

var formMessagesSchema = &schema.Schema{
	Type:        schema.TypeList,
	Optional:    true,
	Description: "Message specific configuration for the form.",
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"errors": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Error message for the form.",
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			"custom": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Custom message for the form.",
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	},
}

func createForm(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	form, err := expandForm(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Form.Create(ctx, form); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(form.GetID())

	return readForm(ctx, data, meta)
}

func readForm(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	form, err := api.Form.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenForm(data, form))
}

func updateForm(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	form, err := expandForm(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Form.Update(ctx, data.Id(), form); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if isNodeEmpty(data) {
		time.Sleep(200 * time.Millisecond)

		if err := api.Request(ctx, http.MethodPatch, api.URI("forms", data.Id()), map[string]interface{}{
			"nodes": []interface{}{},
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	// TODO:
	// _ = data.Set("ending", nil).

	return readForm(ctx, data, meta)
}

func deleteForm(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Form.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
