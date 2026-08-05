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
	"strings"
	"time"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_client_credentials", data, err)
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
				if err := detachClientCredentials(ctx, api, clientID, newMethod, credentialBlockDeclared(data, "signed_request_object")); err != nil {
					return diag.FromErr(err)
				}

				// Delete only the credentials this resource owns. Anything else in
				// the pool belongs to another slot or to no slot at all, and was
				// never created here.
				ownedIDs := ownedCredentialIDs(data)
				for _, cred := range credentials {
					if !ownedIDs[cred.GetID()] {
						continue
					}
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
		// The detach also resets token_endpoint_auth_method, so the update below
		// is only needed when there was nothing attached to begin with.
		if err := detachClientCredentials(
			ctx, api, client.GetClientID(), tokenEndpointAuthMethod,
			credentialBlockDeclared(data, "signed_request_object"),
		); err != nil {
			return diag.FromErr(err)
		}

		// Delete only the credentials this resource owns. The rest of the pool was
		// created elsewhere and is not ours to remove.
		var diagnostics diag.Diagnostics
		ownedIDs := ownedCredentialIDs(data)
		for _, credential := range credentials {
			if !ownedIDs[credential.GetID()] {
				continue
			}

			err := deleteCredentialIgnoringNotFound(ctx, api, client.GetClientID(), credential.GetID())
			if err == nil {
				continue
			}

			// The detach above has already been applied and is not rolled back. A
			// hard error here would leave the client unable to use its previous
			// authentication method while the resource stayed in state, and every
			// retry would fail the same way. Report it and let the destroy finish.
			if isCredentialStillAttachedError(err) {
				diagnostics = append(diagnostics, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Client credential left in place",
					Detail: fmt.Sprintf(
						"The credential with ID %q could not be deleted because it is still "+
							"attached to client %q through a feature this resource does not "+
							"manage. The resource has been removed from the Terraform state.\n\n"+
							"Detach the credential from that feature and delete it directly if it "+
							"is no longer needed.",
						credential.GetID(), client.GetClientID(),
					),
				})
				continue
			}

			return append(diagnostics, diag.FromErr(err)...)
		}

		return diagnostics
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

// maxSlotCredentials is the most credentials Auth0 allows attached to a single
// authentication slot, such as private_key_jwt or signed_request_object. The
// limit is not queryable; exceeding it fails with
// "Array is too long (3), maximum 2".
const maxSlotCredentials = 2

// maxPoolCredentials is the most credential records a client can hold in total.
// The limit is not queryable; exceeding it fails with
// "A client can have a maximum of 4 credentials.".
//
// This is a different container from maxSlotCredentials. The pool holds every
// credential on the client: those attached to any slot, plus any that are
// attached to nothing. Because the slots can address more credentials in total
// than the pool can hold, headroom in a slot does not imply headroom in the
// pool, and a create-then-delete rotation is impossible once the pool is full.
const maxPoolCredentials = 4

// planCredentialRotation orders a credential change into interleaved steps,
// pairing each removal with an addition. Within a pair it chooses the order
// from the live counts: while both the slot and the pool have headroom it adds
// first, keeping a credential attached for zero downtime; once either is at its
// limit it removes first, so neither count ever overshoots. Both counts matter
// because they bound different containers — a slot holding one credential on a
// client whose pool is full still cannot accept a new one.
func planCredentialRotation(diff credentialDiff, attachedCount, poolCount int) []rotationStep {
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

		if attachedCount < maxSlotCredentials && poolCount < maxPoolCredentials {
			// Headroom in both containers: add first so a valid credential
			// stays attached for the duration of the swap.
			rotationSteps = append(rotationSteps, addition)
			attachedCount++
			poolCount++
			if hasID {
				rotationSteps = append(rotationSteps, removal)
				attachedCount--
				poolCount--
			}
		} else {
			// Either container is at its limit: remove first so neither count
			// overshoots. A removal frees a slot entry and a pool record.
			if hasID {
				rotationSteps = append(rotationSteps, removal)
				attachedCount--
				poolCount--
			}
			rotationSteps = append(rotationSteps, addition)
			attachedCount++
			poolCount++
		}
	}
	for _, removed := range diff.toRemove[pairs:] {
		if removal, hasID := newRemoval(removed); hasID {
			rotationSteps = append(rotationSteps, removal)
			attachedCount--
			poolCount--
		}
	}
	for _, added := range diff.toAdd[pairs:] {
		rotationSteps = append(rotationSteps, newAddition(added))
		attachedCount++
		poolCount++
	}

	return rotationSteps
}

// credentialChanges returns the credential additions and removals this apply
// must perform for one credential set.
//
// The change is reported as empty unless the set itself changed. Value.Difference
// reports every entry as an addition when the key has not changed, because it has
// no prior value to subtract, and the update path runs on every change to the
// resource — including ones driven by an unrelated attribute such as
// signed_request_object.required. Acting on that would create a second copy of a
// credential the client already holds, which the API refuses with "credentials
// contains public keys that already exist in client", failing an apply that
// should have left the set alone.
func credentialChanges(data *schema.ResourceData, credentialsKey string) credentialDiff {
	if !data.HasChange(credentialsKey) {
		return credentialDiff{}
	}

	toAdd, toRemove := value.Difference(data, credentialsKey)

	return classifyCredentialChanges(toAdd, toRemove)
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

	diff := credentialChanges(data, credentialsKey)

	var result *multierror.Error

	if len(diff.toAdd) > 0 || len(diff.toRemove) > 0 {
		// The pool is still read, but never treated as this resource's own list.
		// It serves two purposes here: counting records, and confirming which of
		// the credentials state remembers the client actually still holds.
		existingCreds, err := api.Client.ListCredentials(ctx, clientID)
		if err != nil {
			return diag.FromErr(err)
		}

		// Snapshot the credentials attached to this slot so we can mutate the set
		// incrementally without ever exceeding the slot limit. The snapshot comes
		// from state, not from a pool listing: the pool also holds credentials
		// owned by other slots and by nothing at all, and attaching those here
		// would grant them this slot's privileges.
		//
		// Entries state remembers but the client no longer holds are dropped. Such
		// a credential was deleted outside Terraform, and sending its ID back in an
		// attach payload would fail every step of the rotation.
		attachedCreds := retainExistingCredentials(
			attachedCredentialsFromState(data, "private_key_jwt"),
			existingCreds,
		)

		for _, step := range planCredentialRotation(diff, len(attachedCreds), len(existingCreds)) {
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

// credentialBearingAttributes are the schema attributes of this resource that
// hold credential records. Every credential this resource owns is reachable
// through one of them.
var credentialBearingAttributes = []string{
	"private_key_jwt",
	"tls_client_auth",
	"self_signed_tls_client_auth",
	"signed_request_object",
}

// credentialBlockDeclared reports whether this resource declares the named
// credential-bearing block, and so whether the read path may record what the
// API returns for it.
//
// Without this test the read path adopts every slot the client happens to hold.
// A credential attached out of band to a slot the configuration never declared
// is written into state as this resource's own, and the next plan then compares a
// block that is in state against a configuration that does not have it. The
// credential fields are ForceNew, so Terraform proposes a replacement, and the
// destroy leg of that replacement deletes a credential this resource did not
// create. Declining to record the block leaves it as it was found.
//
// Two sources are consulted because neither covers every phase on its own.
// During an apply the configuration is authoritative and present. During a
// refresh it is not: ReadResource populates only RawState, leaving GetRawConfig()
// a null object on which GetAttr would panic. Prior state stands in for the
// configuration there, which holds because the read path is this block's only
// writer and this test gates it.
//
// When both sources are null or unknown there is nothing to compare against — an
// import carries no prior state, and the read following one holds only the ID —
// so the API response is recorded as it was before.
func credentialBlockDeclared(data *schema.ResourceData, attribute string) bool {
	for _, source := range []cty.Value{data.GetRawConfig(), data.GetRawState()} {
		if source.IsNull() || !source.IsKnown() {
			continue
		}

		block := source.GetAttr(attribute)
		if block.IsNull() || !block.IsKnown() {
			continue
		}

		return block.LengthInt() > 0
	}

	return true
}

// ownedCredentialIDs returns the IDs of the credentials this resource manages,
// read from state across every credential-bearing attribute.
//
// The credential pool is deliberately not used as the source. A pool listing
// cannot attribute ownership: GET /clients/{id}/credentials returns an identical
// field set for every credential regardless of which slot holds it, or whether
// any slot does. Treating the pool as this resource's own list makes the
// provider attach and delete credentials it never created, so ownership must
// come from state, which flattenCredentials populates from the client's slot
// lists.
func ownedCredentialIDs(data *schema.ResourceData) map[string]bool {
	ownedIDs := make(map[string]bool)

	for _, attribute := range credentialBearingAttributes {
		for _, credentialID := range stateCredentialIDs(data, attribute) {
			ownedIDs[credentialID] = true
		}
	}

	return ownedIDs
}

// stateCredentialIDs returns the credential IDs recorded in state for one
// attribute, in state order. It handles both the TypeSet and TypeList shapes
// the credential-bearing attributes use.
//
// The prior state is read rather than the planned value, because the planned
// value describes the desired end state: a credential being removed by this
// apply is already absent from it, yet is still attached on the client until its
// removal step runs. Newly added credentials have no ID yet and are skipped.
func stateCredentialIDs(data *schema.ResourceData, attribute string) []string {
	priorState, _ := data.GetChange(fmt.Sprintf("%s.0.credentials", attribute))

	var entries []interface{}
	switch credentials := priorState.(type) {
	case *schema.Set:
		entries = credentials.List()
	case []interface{}:
		entries = credentials
	default:
		return nil
	}

	credentialIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		credential, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if credentialID, _ := credential["id"].(string); credentialID != "" {
			credentialIDs = append(credentialIDs, credentialID)
		}
	}

	return credentialIDs
}

// attachedCredentialsFromState returns the credentials attached to one slot as
// bare {id} references ready for a PATCH, taken from state rather than from the
// pool so that credentials belonging to another slot are never pulled in.
func attachedCredentialsFromState(data *schema.ResourceData, attribute string) []management.Credential {
	credentialIDs := stateCredentialIDs(data, attribute)

	attachedCredentials := make([]management.Credential, 0, len(credentialIDs))
	for _, credentialID := range credentialIDs {
		attachedCredentials = append(attachedCredentials, management.Credential{
			ID: auth0.String(credentialID),
		})
	}

	return attachedCredentials
}

// retainExistingCredentials drops entries from attachedCredentials that are not
// present in the client's credential pool.
//
// State can name a credential the client no longer holds, because a credential
// deleted outside Terraform leaves the slot list in state untouched until the
// next refresh. Carrying such an ID into an attach payload makes the API reject
// every step of the rotation, so it is filtered out instead. The pool is used
// only to test existence here, never as a source of ownership.
func retainExistingCredentials(attachedCredentials []management.Credential, poolCredentials []*management.Credential) []management.Credential {
	inPool := make(map[string]bool, len(poolCredentials))
	for _, credential := range poolCredentials {
		inPool[credential.GetID()] = true
	}

	retained := make([]management.Credential, 0, len(attachedCredentials))
	for _, credential := range attachedCredentials {
		if inPool[credential.GetID()] {
			retained = append(retained, credential)
		}
	}

	return retained
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

// credentialStillAttachedMessage is the message the Management API returns when
// a credential cannot be deleted because a client feature still references it.
const credentialStillAttachedMessage = "still associated with a client"

// isCredentialStillAttachedError reports whether err is the API's refusal to
// delete a credential that is still attached to the client. The API exposes no
// error code for this case, so the message is the only signal available.
func isCredentialStillAttachedError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), credentialStillAttachedMessage)
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

func detachClientCredentials(ctx context.Context, api *management.Management, clientID, tokenEndpointAuthMethod string, detachSignedRequestObject bool) error {
	client := clientWithAuthMethodAndSignedRequestObject{
		ID:                          clientID,
		SignedRequestObject:         nil,
		ClientAuthenticationMethods: nil,
		// API doesn't accept nil on both of these, so we temporarily set this to a default.
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	}

	// Clearing signed_request_object is only correct when this resource manages
	// it. A client can hold a JAR configuration set up elsewhere while this
	// resource declares only an authentication method, and detaching then erases
	// a feature it was never asked to touch.
	if !detachSignedRequestObject {
		return updateClientInternal(ctx, api, client.ID, clientWithAuthMethod{
			ID:                          client.ID,
			ClientAuthenticationMethods: nil,
			TokenEndpointAuthMethod:     client.TokenEndpointAuthMethod,
		})
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
