package client

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/prefetch"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func createClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)

	// Check that client exists.
	if _, err := api.Client.Read(ctx, clientID, management.IncludeFields("client_id")); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(clientID)

	authenticationMethod := data.Get("authentication_method").(string)
	if len(authenticationMethod) > 0 {
		switch authenticationMethod {
		case "private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth":
			if diagnostics := createAuthenticationMethodCredentials(ctx, api, data, authenticationMethod); diagnostics.HasError() {
				return diagnostics
			}
		case "client_secret_post", "client_secret_basic":
			if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}

			if err := updateSecret(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}
		case "none":
			if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if data.GetRawConfig().GetAttr("signed_request_object").LengthInt() > 0 {
		diagnostics := createSignedRequestObject(ctx, api, data)
		if diagnostics.HasError() {
			return diagnostics
		}
	}

	return readClientCredentials(ctx, data, meta)
}

func readClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cfg := meta.(*config.Config)
	api := cfg.GetAPI()

	var (
		client *management.Client
		err    error
	)

	if cache := cfg.GetPrefetchCache(); cache != nil {
		client, err = prefetch.GetClient(ctx, cache, api, data.Id())
	} else {
		client, err = api.Client.Read(ctx, data.Id())
	}

	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenClientCredentials(ctx, api, data, client))
}

func updateClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	// Check that client exists.
	if _, err := api.Client.Read(ctx, data.Id(), management.IncludeFields("client_id")); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	// When switching away from a credential-based auth method, detach and
	// delete existing credentials before changing the auth method.
	if data.HasChange("authentication_method") {
		oldVal, _ := data.GetChange("authentication_method")
		oldMethod, _ := oldVal.(string)
		newMethod := data.Get("authentication_method").(string)

		isOldCredentialBased := oldMethod == "private_key_jwt" || oldMethod == "tls_client_auth" || oldMethod == "self_signed_tls_client_auth"
		isNewCredentialBased := newMethod == "private_key_jwt" || newMethod == "tls_client_auth" || newMethod == "self_signed_tls_client_auth"

		if isOldCredentialBased && !isNewCredentialBased {
			clientID := data.Get("client_id").(string)
			credentials, err := api.Client.ListCredentials(ctx, clientID)
			if err != nil {
				return diag.FromErr(err)
			}
			if len(credentials) > 0 {
				if err := detachClientCredentials(ctx, api, clientID, newMethod); err != nil {
					return diag.FromErr(err)
				}
				for _, cred := range credentials {
					if err := api.Client.DeleteCredential(ctx, clientID, cred.GetID()); err != nil {
						return diag.FromErr(err)
					}
				}
			}
		}
	}

	authenticationMethod := data.Get("authentication_method").(string)
	switch authenticationMethod {
	case "private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth":
		if diagnostics := modifyAuthenticationMethodCredentials(ctx, api, data, authenticationMethod); diagnostics.HasError() {
			return diagnostics
		}
	case "client_secret_post", "client_secret_basic":
		if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}

		if err := updateSecret(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}
	case "none":
		if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.GetRawConfig().GetAttr("signed_request_object").LengthInt() > 0 {
		diagnostics := modifySignedRequestObject(ctx, api, data)
		if diagnostics.HasError() {
			return diagnostics
		}
	}

	return readClientCredentials(ctx, data, meta)
}

func deleteClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id(), management.IncludeFields("client_id", "app_type"))
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	tokenEndpointAuthMethod := ""
	switch client.GetAppType() {
	case "native", "spa":
		tokenEndpointAuthMethod = "none"
	case "regular_web", "non_interactive":
		tokenEndpointAuthMethod = "client_secret_post"
	default:
		tokenEndpointAuthMethod = "client_secret_basic"
	}

	credentials, err := api.Client.ListCredentials(ctx, client.GetClientID())
	if err != nil {
		return diag.FromErr(err)
	}

	if len(credentials) > 0 {
		if err := detachClientCredentials(ctx, api, client.GetClientID(), tokenEndpointAuthMethod); err != nil {
			return diag.FromErr(err)
		}

		for _, credential := range credentials {
			if err := api.Client.DeleteCredential(ctx, client.GetClientID(), credential.GetID()); err != nil {
				return diag.FromErr(err)
			}
		}

		return nil
	}

	if err := api.Client.Update(ctx, client.GetClientID(), &management.Client{
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func createAuthenticationMethodCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, authenticationMethod string) diag.Diagnostics {
	credentials, diagnostics := expandAuthenticationMethodCredentials(data.GetRawConfig(), authenticationMethod)
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	credentialsToAttach := make([]management.Credential, 0)
	for _, credential := range credentials {
		if err := api.Client.CreateCredential(ctx, clientID, credential); err != nil {
			return diag.FromErr(err)
		}

		credentialsToAttach = append(credentialsToAttach, management.Credential{
			ID: credential.ID,
		})
	}

	err := attachAuthenticationMethodCredentials(ctx, api, clientID, authenticationMethod, credentialsToAttach)

	return diag.FromErr(err)
}

func modifyAuthenticationMethodCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, authenticationMethod string) diag.Diagnostics {
	if authenticationMethod == "private_key_jwt" {
		return modifyPrivateKeyJWTCredentials(ctx, api, data)
	}

	return modifyListBasedCredentials(ctx, api, data, authenticationMethod)
}

type expiryUpdate struct {
	credentialID string
	expiresAt    string
}

type credentialDiff struct {
	toAdd         []interface{}
	toRemove      []interface{}
	expiryUpdates []expiryUpdate
}

// rotationStepKind holds the type of a rotation step.
type rotationStepKind int

const (
	// Detaches a credential from the client, then deletes it.
	detachAndDelete rotationStepKind = iota
	// Creates a credential, then attaches it to the client.
	createAndAttach
)

// rotationStep is a single ordered action in a credential rotation.
type rotationStep struct {
	kind          rotationStepKind
	credentialID  string                 // Set when kind == detachAndDelete.
	newCredential map[string]interface{} // Set when kind == createAndAttach.
}

// minCredentialCap is the smallest per-client credential limit Auth0 enforces
// on any tenant. The cap is not queryable, but every tenant allows at least
// this many, so holding this many attached credentials transiently is always
// safe.
const minCredentialCap = 2

// planCredentialRotation orders a credential change into interleaved steps,
// pairing each removal with an addition. Within a pair it chooses the order
// from the live attached count: with headroom below minCredentialCap it adds
// first (keeping a credential attached for zero downtime); at or above the cap
// it removes first (so the count never overshoots the tenant limit).
func planCredentialRotation(diff credentialDiff, attachedCount int) []rotationStep {
	rotationSteps := make([]rotationStep, 0, len(diff.toRemove)+len(diff.toAdd))

	newRemoval := func(entry interface{}) (rotationStep, bool) {
		credMap, _ := entry.(map[string]interface{})
		id, _ := credMap["id"].(string)
		return rotationStep{kind: detachAndDelete, credentialID: id}, id != ""
	}
	newAddition := func(entry interface{}) rotationStep {
		credMap, _ := entry.(map[string]interface{})
		return rotationStep{kind: createAndAttach, newCredential: credMap}
	}

	pairs := min(len(diff.toAdd), len(diff.toRemove))

	for i := range pairs {
		removal, hasID := newRemoval(diff.toRemove[i])
		addition := newAddition(diff.toAdd[i])

		if attachedCount < minCredentialCap {
			// Headroom: add first so a valid credential stays attached.
			rotationSteps = append(rotationSteps, addition)
			attachedCount++
			if hasID {
				rotationSteps = append(rotationSteps, removal)
				attachedCount--
			}
		} else {
			// At capacity: remove first so the count never overshoots.
			if hasID {
				rotationSteps = append(rotationSteps, removal)
				attachedCount--
			}
			rotationSteps = append(rotationSteps, addition)
			attachedCount++
		}
	}
	for _, removed := range diff.toRemove[pairs:] {
		if removal, hasID := newRemoval(removed); hasID {
			rotationSteps = append(rotationSteps, removal)
			attachedCount--
		}
	}
	for _, added := range diff.toAdd[pairs:] {
		rotationSteps = append(rotationSteps, newAddition(added))
		attachedCount++
	}

	return rotationSteps
}

func classifyCredentialChanges(toAdd, toRemove []interface{}) credentialDiff {
	var expiryUpdates []expiryUpdate
	remainingAdd := make([]interface{}, 0, len(toAdd))
	remainingRemove := make([]interface{}, 0, len(toRemove))

	usedRemoveIndexes := make(map[int]bool)
	for _, addedCred := range toAdd {
		addMap := addedCred.(map[string]interface{})
		addPEM, _ := addMap["pem"].(string)
		addAlgo, _ := addMap["algorithm"].(string)
		addExpiry, _ := addMap["expires_at"].(string)

		matched := false
		for i, removedCred := range toRemove {
			if usedRemoveIndexes[i] {
				continue
			}
			rmMap := removedCred.(map[string]interface{})
			rmPEM, _ := rmMap["pem"].(string)
			rmAlgo, _ := rmMap["algorithm"].(string)
			rmID, _ := rmMap["id"].(string)
			rmKeyID, _ := rmMap["key_id"].(string)

			if rmID == "" {
				continue
			}

			var pemMatch bool
			if rmPEM == addPEM {
				pemMatch = true
			} else if rmPEM == "" && rmKeyID != "" && addPEM != "" {
				pemMatch = jwkThumbprint(addPEM) == rmKeyID
			}

			if pemMatch && rmAlgo == addAlgo {
				rmParseExpiry, _ := rmMap["parse_expiry_from_cert"].(bool)
				if rmParseExpiry && rmPEM != "" {
					continue
				}

				rmExpiry, _ := rmMap["expires_at"].(string)
				if addExpiry != "" && addExpiry != rmExpiry && !rmParseExpiry {
					expiryUpdates = append(expiryUpdates, expiryUpdate{
						credentialID: rmID,
						expiresAt:    addExpiry,
					})
				}

				usedRemoveIndexes[i] = true
				matched = true
				break
			}
		}
		if !matched {
			remainingAdd = append(remainingAdd, addedCred)
		}
	}
	for i, removedCred := range toRemove {
		if !usedRemoveIndexes[i] {
			remainingRemove = append(remainingRemove, removedCred)
		}
	}

	return credentialDiff{
		toAdd:         remainingAdd,
		toRemove:      remainingRemove,
		expiryUpdates: expiryUpdates,
	}
}

func modifyPrivateKeyJWTCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	clientID := data.Get("client_id").(string)
	credentialsKey := "private_key_jwt.0.credentials" //nolint:gosec // This is a Terraform schema key, not a credential.

	toAdd, toRemove := value.Difference(data, credentialsKey)
	diff := classifyCredentialChanges(toAdd, toRemove)

	var result *multierror.Error

	if len(diff.toAdd) > 0 || len(diff.toRemove) > 0 {
		// Snapshot the currently attached credentials so we can mutate the set
		// incrementally, one slot at a time, without ever exceeding the cap. The
		// live count also drives the per-pair add-first vs remove-first ordering.
		existingCreds, err := api.Client.ListCredentials(ctx, clientID)
		if err != nil {
			return diag.FromErr(err)
		}

		attachedCreds := make([]management.Credential, 0, len(existingCreds))
		for _, cred := range existingCreds {
			attachedCreds = append(attachedCreds, management.Credential{ID: cred.ID})
		}

		for _, step := range planCredentialRotation(diff, len(attachedCreds)) {
			switch step.kind {
			case detachAndDelete:
				attachedCreds = removeAttachedCredential(attachedCreds, step.credentialID)
				if err := attachAuthenticationMethodCredentials(ctx, api, clientID, "private_key_jwt", attachedCreds); err != nil {
					return diag.FromErr(err)
				}
				if err := deleteCredentialIgnoringNotFound(ctx, api, clientID, step.credentialID); err != nil {
					return diag.FromErr(err)
				}
			case createAndAttach:
				credential := expandClientCredentialFromMap(step.newCredential)
				if err := api.Client.CreateCredential(ctx, clientID, credential); err != nil {
					return diag.FromErr(err)
				}
				attachedCreds = append(attachedCreds, management.Credential{ID: credential.ID})
				if err := attachAuthenticationMethodCredentials(ctx, api, clientID, "private_key_jwt", attachedCreds); err != nil {
					// Roll back the just-created credential so it does not linger
					// unattached and consume a cap slot on the next apply.
					if deleteErr := deleteCredentialIgnoringNotFound(ctx, api, clientID, credential.GetID()); deleteErr != nil {
						return diag.Errorf("failed to attach credential (rollback delete also failed: %v): %v", deleteErr, err)
					}
					return diag.FromErr(err)
				}
			}
		}
	}

	// Apply expires_at PATCH updates for credentials that only changed expiry.
	for _, update := range diff.expiryUpdates {
		t, parseErr := time.Parse(time.RFC3339, update.expiresAt)
		if parseErr != nil {
			t, parseErr = time.Parse(timeRFC3339WithMilliseconds, update.expiresAt)
			if parseErr != nil {
				continue
			}
		}

		if err := api.Client.UpdateCredential(ctx, clientID, update.credentialID, &management.Credential{
			ExpiresAt: &t,
		}); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return diag.FromErr(result.ErrorOrNil())
}

// removeAttachedCredential returns creds without the entry matching id.
func removeAttachedCredential(creds []management.Credential, id string) []management.Credential {
	filtered := make([]management.Credential, 0, len(creds))
	for _, cred := range creds {
		if cred.GetID() != id {
			filtered = append(filtered, cred)
		}
	}
	return filtered
}

// deleteCredentialIgnoringNotFound deletes a credential, treats a 404 as success.
func deleteCredentialIgnoringNotFound(ctx context.Context, api *management.Management, clientID, credentialID string) error {
	err := api.Client.DeleteCredential(ctx, clientID, credentialID)
	if internalError.IsStatusNotFound(err) {
		return nil
	}
	return err
}

// modifyListBasedCredentials handles update for tls_client_auth and self_signed_tls_client_auth
// which still use TypeList.
func modifyListBasedCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, authenticationMethod string) diag.Diagnostics {
	credentials, diagnostics := expandAuthenticationMethodCredentials(data.GetRawConfig(), authenticationMethod)
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	for index, credential := range credentials {
		configAddress := fmt.Sprintf("%s.0.credentials.%d", authenticationMethod, index)
		if !data.HasChange(configAddress) {
			continue
		}

		credentialID := data.Get(fmt.Sprintf("%s.id", configAddress)).(string)
		stateExpiresAt := data.Get(fmt.Sprintf("%s.expires_at", configAddress)).(string)
		if stateExpiresAt == "" {
			continue
		}

		expiresAt, _ := time.Parse(time.RFC3339, stateExpiresAt)
		credential.ExpiresAt = &expiresAt

		if err := api.Client.UpdateCredential(ctx, clientID, credentialID, credential); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func expandClientCredentialFromMap(m map[string]interface{}) *management.Credential {
	credentialType, _ := m["credential_type"].(string)
	credential := &management.Credential{
		CredentialType: &credentialType,
	}

	if name, ok := m["name"].(string); ok && name != "" {
		credential.Name = &name
	}

	if credentialType == "public_key" {
		if pem, ok := m["pem"].(string); ok && pem != "" {
			credential.PEM = &pem
		}
		if algo, ok := m["algorithm"].(string); ok && algo != "" {
			credential.Algorithm = &algo
		}
		if parseExpiry, ok := m["parse_expiry_from_cert"].(bool); ok {
			credential.ParseExpiryFromCert = &parseExpiry
		}
		if expiresAt, ok := m["expires_at"].(string); ok && expiresAt != "" {
			t, err := time.Parse(time.RFC3339, expiresAt)
			if err == nil {
				credential.ExpiresAt = &t
			}
		}
	}

	return credential
}

func createSignedRequestObject(ctx context.Context, api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	signedRequestObject, diagnostics := expandSignedRequestObject(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	if signedRequestObject.GetCredentials() != nil {
		credentialsToAttach := make([]management.Credential, 0)
		for _, credential := range signedRequestObject.GetCredentials() {
			if err := api.Client.CreateCredential(ctx, clientID, &credential); err != nil {
				return diag.FromErr(err)
			}

			credentialsToAttach = append(credentialsToAttach, management.Credential{
				ID: credential.ID,
			})
		}

		return diag.FromErr(attachSignedRequestObjectCredentials(ctx, api, clientID, signedRequestObject.Required, credentialsToAttach))
	}

	return nil
}

func modifySignedRequestObject(ctx context.Context, api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	signedRequestObject, diagnostics := expandSignedRequestObject(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	if signedRequestObject.GetCredentials() != nil {
		for index, credential := range signedRequestObject.GetCredentials() {
			configAddress := fmt.Sprintf("signed_request_object.0.credentials.%d", index)
			if !data.HasChange(configAddress) {
				continue
			}

			credentialID := data.Get(fmt.Sprintf("%s.id", configAddress)).(string)
			stateExpiresAt := data.Get(fmt.Sprintf("%s.expires_at", configAddress)).(string)
			if stateExpiresAt == "" {
				continue
			}

			// The error can be ignored, the schema validates the type.
			expiresAt, _ := time.Parse(time.RFC3339, stateExpiresAt)
			credential.ExpiresAt = &expiresAt

			// Limitation: Unable to update the credential to never expire. Needs to get deleted and recreated if needed.
			if err := api.Client.UpdateCredential(ctx, clientID, credentialID, &credential); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if data.HasChange("signed_request_object.0.required") {
		return diag.FromErr(attachSignedRequestObjectNoCredentials(ctx, api, clientID, signedRequestObject.Required))
	}

	return nil
}

type clientWithAuthMethod struct {
	ID                          string                                  `json:"-"`
	ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
	TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
}

type clientWithSignedRequestObject struct {
	ID                  string                                `json:"-"`
	SignedRequestObject *management.ClientSignedRequestObject `json:"signed_request_object"`
}

type clientWithAuthMethodAndSignedRequestObject struct {
	ID                          string                                  `json:"-"`
	ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
	TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
	SignedRequestObject         *management.ClientSignedRequestObject   `json:"signed_request_object"`
}

func attachAuthenticationMethodCredentials(ctx context.Context, api *management.Management, clientID string, authenticationMethod string, credentials []management.Credential) error {
	client := clientWithAuthMethod{
		ID:                          clientID,
		ClientAuthenticationMethods: &management.ClientAuthenticationMethods{},
		TokenEndpointAuthMethod:     nil,
	}

	switch authenticationMethod {
	case "private_key_jwt":
		client.ClientAuthenticationMethods.PrivateKeyJWT = &management.PrivateKeyJWT{
			Credentials: &credentials,
		}
	case "tls_client_auth":
		client.ClientAuthenticationMethods.TLSClientAuth = &management.TLSClientAuth{
			Credentials: &credentials,
		}
	case "self_signed_tls_client_auth":
		client.ClientAuthenticationMethods.SelfSignedTLSClientAuth = &management.SelfSignedTLSClientAuth{
			Credentials: &credentials,
		}
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func attachSignedRequestObjectCredentials(ctx context.Context, api *management.Management, clientID string, required *bool, credentials []management.Credential) error {
	client := clientWithSignedRequestObject{
		ID: clientID,
		SignedRequestObject: &management.ClientSignedRequestObject{
			Required:    required,
			Credentials: &credentials,
		},
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func attachSignedRequestObjectNoCredentials(ctx context.Context, api *management.Management, clientID string, required *bool) error {
	client := clientWithSignedRequestObject{
		ID: clientID,
		SignedRequestObject: &management.ClientSignedRequestObject{
			Required: required,
		},
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func detachClientCredentials(ctx context.Context, api *management.Management, clientID, tokenEndpointAuthMethod string) error {
	client := clientWithAuthMethodAndSignedRequestObject{
		ID:                          clientID,
		SignedRequestObject:         nil,
		ClientAuthenticationMethods: nil,
		// API doesn't accept nil on both of these, so we temporarily set this to a default.
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func updateClientInternal(ctx context.Context, api *management.Management, clientID string, client interface{}) error {
	c, err := api.Client.Read(ctx, clientID, management.IncludeFields("client_id", "app_type"))
	if err != nil {
		return err
	}

	var payloadMap map[string]interface{}
	jsonBytes, _ := json.Marshal(client)
	_ = json.Unmarshal(jsonBytes, &payloadMap)

	if c.GetAppType() == "express_configuration" {
		// Go's delete is safe even if the key doesn't exist.
		delete(payloadMap, "signed_request_object")
		delete(payloadMap, "token_endpoint_auth_method")
		if payloadMap["client_authentication_methods"] == nil {
			payloadMap["client_authentication_methods"] = management.ClientAuthenticationMethods{
				PrivateKeyJWT: &management.PrivateKeyJWT{
					Credentials: &[]management.Credential{},
				},
			}
		}
	}

	request, err := api.NewRequest(ctx, http.MethodPatch, api.URI("clients", clientID), payloadMap)
	if err != nil {
		return err
	}

	response, err := api.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func updateTokenEndpointAuthMethod(ctx context.Context, api *management.Management, data *schema.ResourceData) error {
	if !data.HasChange("authentication_method") {
		return nil
	}

	clientID := data.Get("client_id").(string)
	tokenEndpointAuthMethod := data.Get("authentication_method").(string)

	return api.Client.Update(ctx, clientID, &management.Client{
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	})
}

func updateSecret(ctx context.Context, api *management.Management, data *schema.ResourceData) error {
	clientID := data.Get("client_id").(string)

	// Write-only values are not available via data.Get(); read them from the raw config.
	secretWO := data.GetRawConfig().GetAttr("client_secret_wo")
	if !secretWO.IsNull() && (data.IsNewResource() || data.HasChange("client_secret_wo_version")) {
		clientSecret := secretWO.AsString()

		return api.Client.Update(ctx, clientID, &management.Client{
			ClientSecret: &clientSecret,
		})
	}

	if !data.HasChange("client_secret") {
		return nil
	}

	clientSecret := data.Get("client_secret").(string)

	return api.Client.Update(ctx, clientID, &management.Client{
		ClientSecret: &clientSecret,
	})
}

func expandAuthenticationMethodCredentials(rawConfig cty.Value, authenticationMethod string) ([]*management.Credential, diag.Diagnostics) {
	credentials := make([]*management.Credential, 0)

	rawConfig.GetAttr(authenticationMethod).ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
			credentials = append(credentials, expandClientCredential(config))
			return stop
		})
		return stop
	})

	if len(credentials) == 0 {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Client Credentials Missing",
				Detail:        fmt.Sprintf("You must define client credentials when setting the authentication method as %q.", authenticationMethod),
				AttributePath: cty.Path{cty.GetAttrStep{Name: fmt.Sprintf("%s.credentials", authenticationMethod)}},
			},
		}
	} else if authenticationMethod == "tls_client_auth" {
		for _, credential := range credentials {
			if (credential.PEM != nil && credential.SubjectDN != nil) || (credential.PEM == nil && credential.SubjectDN == nil) {
				return nil, diag.Diagnostics{
					diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Client Credentials Invalid",
						Detail:        fmt.Sprintf("Exactly one of pem and subject_dn must be set when setting the authentication method as %q.", authenticationMethod),
						AttributePath: cty.Path{cty.GetAttrStep{Name: fmt.Sprintf("%s.credentials", authenticationMethod)}},
					},
				}
			}
		}
	}

	return credentials, nil
}

func expandSignedRequestObject(rawConfig cty.Value) (*management.ClientSignedRequestObject, diag.Diagnostics) {
	signedRequestObjectConfig := rawConfig.GetAttr("signed_request_object")
	if signedRequestObjectConfig.IsNull() {
		return nil, nil
	}

	var signedRequestObject management.ClientSignedRequestObject

	signedRequestObjectConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		credentials := make([]management.Credential, 0)
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
			credentials = append(credentials, *expandClientCredential(config))
			return stop
		})
		signedRequestObject.Credentials = &credentials
		signedRequestObject.Required = value.Bool(config.GetAttr("required"))
		return stop
	})

	if signedRequestObject == (management.ClientSignedRequestObject{}) {
		return nil, nil
	}

	if len(*signedRequestObject.Credentials) == 0 {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Client Credentials Missing",
				Detail:        "You must define client credentials when using JWT-secured Authorization Requests.",
				AttributePath: cty.Path{cty.GetAttrStep{Name: "signed_request_object.credentials"}},
			},
		}
	}

	return &signedRequestObject, nil
}

func expandClientCredential(rawConfig cty.Value) *management.Credential {
	clientCredential := management.Credential{
		Name:           value.String(rawConfig.GetAttr("name")),
		CredentialType: value.String(rawConfig.GetAttr("credential_type")),
	}

	switch *clientCredential.CredentialType {
	case "public_key":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
		clientCredential.Algorithm = value.String(rawConfig.GetAttr("algorithm"))
		clientCredential.ParseExpiryFromCert = value.Bool(rawConfig.GetAttr("parse_expiry_from_cert"))
		clientCredential.ExpiresAt = value.Time(rawConfig.GetAttr("expires_at"))
	case "cert_subject_dn":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
		clientCredential.SubjectDN = value.String(rawConfig.GetAttr("subject_dn"))
	case "x509_cert":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
	}

	return &clientCredential
}

// jwkThumbprint computes the RFC 7638 JWK thumbprint from a PEM-encoded
// certificate or public key. Returns empty string if the PEM cannot be parsed
// or the key type is not RSA.
func jwkThumbprint(pemData string) string {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return ""
	}

	var pub *rsa.PublicKey

	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return ""
		}
		var ok bool
		pub, ok = cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return ""
		}
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return ""
		}
		var ok bool
		pub, ok = key.(*rsa.PublicKey)
		if !ok {
			return ""
		}
	case "RSA PUBLIC KEY":
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return ""
		}
		pub = key
	default:
		return ""
	}

	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	canonical := fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`, e, n)

	h := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
