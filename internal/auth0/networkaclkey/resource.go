package networkaclkey

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource returns a new auth0_network_acl_key resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createNetworkACLKey,
		ReadContext:   readNetworkACLKey,
		UpdateContext: updateNetworkACLKey,
		DeleteContext: deleteNetworkACLKey,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			customizeDiffNetworkACLKey,
		),
		Description: "Manages HMAC-SHA256 signing keys used to verify HTTP Message Signatures on Network ACL rules. " +
			"The key material (`value`) is write-only: it is not returned by the API and is not stored in Terraform " +
			"state. Changes are detected by comparing SHA-256 fingerprints. (EA Only)",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "User-supplied label for the key. Must be unique across all Network ACL keys for the tenant. Max 255 characters.",
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"alg": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Algorithm used for the key. Currently only `hmac-sha256` is supported.",
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"hmac-sha256"}, false),
				),
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				Description: "Base64-encoded raw key material. The decoded value must be between 32 and 512 bytes. " +
					"This field is write-only: it is not returned by the API and not stored in Terraform state. " +
					"Changes are detected by comparing SHA-256 fingerprints.",
				ValidateDiagFunc: validateNetworkACLKeyValue,
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA-256 fingerprint of the decoded key material (lowercase hex). Used for drift detection.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time when the key was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time when the key was last updated.",
			},
		},
	}
}

// fingerprintForValue computes SHA-256 of the base64-decoded key material.
func fingerprintForValue(b64value string) (string, error) {
	decoded, err := decodeKeyValue(b64value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(decoded)
	return fmt.Sprintf("%x", sum), nil
}

// decodeKeyValue decodes a base64-encoded key value, trying standard then raw (no-padding) encoding.
func decodeKeyValue(b64value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(b64value)
		if err != nil {
			return nil, fmt.Errorf("value must be valid base64-encoded data: %w", err)
		}
	}
	return decoded, nil
}

// validateNetworkACLKeyValue validates that value is base64 and decodes to 32–512 bytes.
func validateNetworkACLKeyValue(v interface{}, _ cty.Path) diag.Diagnostics {
	s, ok := v.(string)
	if !ok {
		return diag.Errorf("expected string")
	}

	decoded, err := decodeKeyValue(s)
	if err != nil {
		return diag.Errorf("value must be valid base64-encoded data: %s", err)
	}

	if len(decoded) < 32 {
		return diag.Errorf("decoded key material must be at least 32 bytes, got %d", len(decoded))
	}

	if len(decoded) > 512 {
		return diag.Errorf("decoded key material must be at most 512 bytes, got %d", len(decoded))
	}

	return nil
}

func customizeDiffNetworkACLKey(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	rawValue := d.Get("value").(string)
	if rawValue == "" {
		return nil
	}

	newFP, err := fingerprintForValue(rawValue)
	if err != nil {
		return err
	}

	if err := d.SetNew("fingerprint", newFP); err != nil {
		return err
	}

	oldFP, _ := d.GetChange("fingerprint")
	if oldFP.(string) != "" && oldFP.(string) != newFP {
		return d.ForceNew("value")
	}

	return nil
}

func createNetworkACLKey(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	req := expandNetworkACLKey(data)

	result, err := apiv3.Keys.NetworkACLs.Create(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(result.GetID())

	return readNetworkACLKey(ctx, data, meta)
}

func readNetworkACLKey(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	key, err := apiv3.Keys.NetworkACLs.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_network_acl_key", data, err)
	}

	// If the API fingerprint differs from what is in state, the key was rotated out-of-band.
	if stateFP := data.Get("fingerprint").(string); stateFP != "" && stateFP != key.GetFingerprint() {
		data.SetId("")
		return nil
	}

	return diag.FromErr(flattenNetworkACLKey(data, key))
}

// updateNetworkACLKey exists to satisfy the Terraform SDK requirement that every non-ForceNew field
// must have an UpdateContext defined. All meaningful changes to this resource (fingerprint mismatch)
// are handled by CustomizeDiff via d.ForceNew, so this function is never reached in practice.
func updateNetworkACLKey(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return readNetworkACLKey(ctx, data, meta)
}

func deleteNetworkACLKey(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	err := apiv3.Keys.NetworkACLs.Delete(ctx, data.Id())
	if err == nil {
		return nil
	}

	// 404 means already deleted — treat as success.
	if internalError.IsStatusNotFound(err) {
		return nil
	}

	// 409 means the key is still referenced by one or more ACLs.
	var conflictErr *managementv3.ConflictError
	if errors.As(err, &conflictErr) {
		return diag.Errorf(
			"cannot delete network ACL key %q: it is still referenced by one or more ACL rules. "+
				"Remove the key references from all ACL rules before deleting. API response: %v",
			data.Id(), conflictErr.Body,
		)
	}

	if isHTTPStatus(err, http.StatusConflict) {
		return diag.FromErr(fmt.Errorf(
			"cannot delete network ACL key %q: it is still in use by one or more ACL rules: %w",
			data.Id(), err,
		))
	}

	return diag.FromErr(err)
}

// isHTTPStatus reports whether err wraps an HTTP response with the given status code.
func isHTTPStatus(err error, status int) bool {
	type httpStatus interface {
		StatusCode() int
	}

	var sc httpStatus
	if errors.As(err, &sc) {
		return sc.StatusCode() == status
	}
	return false
}
