package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

const (
	maxMetadataBytes = 10240
	maxNameLength    = 255
	maxExternalIDLen = 128
)

// NewResource will return a new auth0_agent resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createAgent,
		ReadContext:   readAgent,
		UpdateContext: updateAgent,
		DeleteContext: deleteAgent,
		Importer: &schema.ResourceImporter{
			StateContext: importAgent,
		},
		Description: "With this resource, you can manage an agent identity (EA only). Link an agent to an " +
			"application with the `agent_id` attribute of the `auth0_client` resource. Deleting an agent does " +
			"not delete the clients linked to it and does not clear their `agent_id`, so reference the agent " +
			"as `auth0_agent.<name>.id` to make Terraform unlink the client before the agent is deleted. " +
			"The Management API credentials need the `*:agents` scopes. Tenants without the feature enabled " +
			"respond with `403 not_enabled`.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateName,
				Description:  "The name of the agent. Must be 1-255 characters and cannot contain `<`, `>` or null characters.",
			},
			"external_agent_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateExternalAgentID,
				Description: "An external identifier for the agent. Must be unique within the tenant, " +
					"1-128 characters and cannot contain null characters. Changing it forces a new agent to be created.",
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateMetadata,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description: "Arbitrary key-value metadata for the agent, as a JSON object of at most 10240 bytes. " +
					"Removing the attribute clears the metadata.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time the agent was created, in RFC 3339 format.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time the agent was last updated, in RFC 3339 format.",
			},
		},
	}
}

func validateName(i interface{}, key string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", key)}
	}

	if len(v) < 1 || len(v) > maxNameLength {
		return nil, []error{fmt.Errorf("expected %q to be between 1 and %d characters", key, maxNameLength)}
	}

	if strings.ContainsAny(v, "<>\x00") {
		return nil, []error{fmt.Errorf("expected %q not to contain '<', '>' or null characters", key)}
	}

	return nil, nil
}

func validateExternalAgentID(i interface{}, key string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", key)}
	}

	if len(v) < 1 || len(v) > maxExternalIDLen {
		return nil, []error{fmt.Errorf("expected %q to be between 1 and %d characters", key, maxExternalIDLen)}
	}

	if strings.ContainsRune(v, 0) {
		return nil, []error{fmt.Errorf("expected %q not to contain null characters", key)}
	}

	return nil, nil
}

func validateMetadata(i interface{}, key string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", key)}
	}

	if warns, errs := validation.StringIsJSON(i, key); len(errs) > 0 {
		return warns, errs
	}

	var object map[string]interface{}
	if err := json.Unmarshal([]byte(v), &object); err != nil || object == nil {
		return nil, []error{fmt.Errorf("expected %q to be a JSON object", key)}
	}

	if len(v) > maxMetadataBytes {
		return nil, []error{fmt.Errorf("expected %q to be at most %d bytes, got %d", key, maxMetadataBytes, len(v))}
	}

	return nil, nil
}

func createAgent(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	request, err := expandAgentCreate(data)
	if err != nil {
		return diag.FromErr(err)
	}

	agent, err := apiv3.Agents.Create(ctx, request)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(agent.GetAgentID())

	return readAgent(ctx, data, meta)
}

func readAgent(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	agent, err := apiv3.Agents.Read(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_agent", data, err)
	}

	return diag.FromErr(flattenAgent(data, agent))
}

func updateAgent(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	request, err := expandAgentUpdate(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := apiv3.Agents.Update(ctx, data.Id(), request); err != nil {
		return diag.FromErr(err)
	}

	return readAgent(ctx, data, meta)
}

func deleteAgent(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	return diag.FromErr(internalError.HandleAPIError(data, apiv3.Agents.Delete(ctx, data.Id())))
}

func importAgent(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if _, errs := internalValidation.IsAgentID(data.Id(), "id"); len(errs) > 0 {
		return nil, errs[0]
	}

	return []*schema.ResourceData{data}, nil
}
