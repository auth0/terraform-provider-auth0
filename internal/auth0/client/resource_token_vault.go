package client

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// validateTokenVaultPrivilegedAccess enforces the two token_vault_privileged_access
// constraints that MaxItems/TypeSet can't express: the total number of scopes across
// all grants must not exceed 20, and no two grants may target the same connection.
// Both are plan-time checks; everything else about the block (caps the SDK already
// covers, IP/CIDR format, scope non-emptiness) is left to the API. See spec §2.6.
func validateTokenVaultPrivilegedAccess(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	rawConfig := diff.GetRawConfig()
	if rawConfig.IsNull() {
		return nil
	}

	tokenVaultConfig := rawConfig.GetAttr("token_vault_privileged_access")
	if tokenVaultConfig.IsNull() || tokenVaultConfig.LengthInt() == 0 {
		return nil
	}

	var validationErr error

	tokenVaultConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		grantsConfig := cfg.GetAttr("grants")
		if grantsConfig.IsNull() {
			return false
		}

		totalScopes := 0
		seenConnections := make(map[string]bool)

		grantsConfig.ForEachElement(func(_ cty.Value, grant cty.Value) (stopInner bool) {
			connectionValue := grant.GetAttr("connection")
			if !connectionValue.IsNull() {
				connection := connectionValue.AsString()
				if seenConnections[connection] {
					validationErr = fmt.Errorf(
						"`token_vault_privileged_access.grants` contains more than one grant for connection %q; "+
							"each connection may appear at most once",
						connection,
					)
					return true
				}
				seenConnections[connection] = true
			}

			scopesValue := grant.GetAttr("scopes")
			if !scopesValue.IsNull() {
				totalScopes += scopesValue.LengthInt()
			}

			return false
		})

		if validationErr != nil {
			return true
		}

		if totalScopes > 20 {
			validationErr = fmt.Errorf(
				"`token_vault_privileged_access.grants` scopes must not exceed 20 in total across all grants, got %d",
				totalScopes,
			)
			return true
		}

		return false
	})

	return validationErr
}

// diffTokenVaultCredentials matches the configured token_vault_privileged_access
// credentials against those attached in the prior state.
func diffTokenVaultCredentials(data *schema.ResourceData) (keepIDs []string, toCreate []*management.Credential, toDelete []string) {
	oldRaw, newRaw := data.GetChange("token_vault_privileged_access.0.credentials")
	oldList, _ := oldRaw.([]interface{})
	newList, _ := newRaw.([]interface{})

	return matchTokenVaultCredentials(oldList, newList)
}

// matchTokenVaultCredentials matches configured credentials against those
// attached in the prior state, by (credential_type, pem) — the only stable
// identifier available before creation, since id is server-assigned and
// credential material is immutable (a change requires create+attach+delete,
// never an in-place update; see spec §2.2/§2.8). Unmatched configured
// credentials must be created; unmatched previously attached ids are stale
// and should be deleted once the update detaching them succeeds.
func matchTokenVaultCredentials(oldList, newList []interface{}) (keepIDs []string, toCreate []*management.Credential, toDelete []string) {
	type credentialKey struct {
		credentialType string
		pem            string
	}

	availableIDs := make(map[credentialKey][]string)
	for _, item := range oldList {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := entry["id"].(string)
		if id == "" {
			continue
		}

		credentialType, _ := entry["credential_type"].(string)
		pem, _ := entry["pem"].(string)
		key := credentialKey{credentialType, pem}
		availableIDs[key] = append(availableIDs[key], id)
	}

	matchedIDs := make(map[string]bool)

	for _, item := range newList {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		credentialType, _ := entry["credential_type"].(string)
		pem, _ := entry["pem"].(string)
		key := credentialKey{credentialType, pem}

		matched := ""
		for _, id := range availableIDs[key] {
			if !matchedIDs[id] {
				matched = id
				break
			}
		}

		if matched != "" {
			matchedIDs[matched] = true
			keepIDs = append(keepIDs, matched)
			continue
		}

		toCreate = append(toCreate, &management.Credential{
			CredentialType: auth0.String(credentialType),
			PEM:            auth0.String(pem),
		})
	}

	for _, ids := range availableIDs {
		for _, id := range ids {
			if !matchedIDs[id] {
				toDelete = append(toDelete, id)
			}
		}
	}

	return keepIDs, toCreate, toDelete
}

// modifyTokenVaultPrivilegedAccessCredentials resolves the raw credential
// material expandTokenVaultPrivilegedAccess produced (credential_type + pem,
// the shape POST /clients accepts inline) against the credentials attached in
// the prior state, since PATCH /clients/{id} accepts only id references.
//
// Mutates client.TokenVaultPrivilegedAccess.Credentials in place when the
// block is being written on this update. Returns the ids of credentials no
// longer configured; the caller must delete these only after the PATCH that
// detaches them has succeeded (never before), bounding partial-failure damage.
func modifyTokenVaultPrivilegedAccessCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, client *management.Client) ([]string, error) {
	keepIDs, toCreate, toDelete := diffTokenVaultCredentials(data)

	if client.TokenVaultPrivilegedAccess == nil {
		return toDelete, nil
	}

	resolved := make([]management.ClientCredentialID, 0, len(keepIDs)+len(toCreate))
	for _, id := range keepIDs {
		resolved = append(resolved, management.ClientCredentialID{ID: auth0.String(id)})
	}

	for _, credential := range toCreate {
		if err := api.Client.CreateCredential(ctx, data.Id(), credential); err != nil {
			return nil, err
		}

		resolved = append(resolved, management.ClientCredentialID{ID: credential.ID})
	}

	client.TokenVaultPrivilegedAccess.Credentials = &resolved

	return toDelete, nil
}
