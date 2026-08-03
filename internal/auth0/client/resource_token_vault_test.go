package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uniqueScopes(prefix string, count int) []string {
	scopes := make([]string, count)
	for i := range scopes {
		scopes[i] = fmt.Sprintf("%s-%d", prefix, i)
	}
	return scopes
}

// diffTokenVaultConfig runs NewResource().Diff against a config holding the
// given token_vault_privileged_access value, returning the error (if any)
// that validateTokenVaultPrivilegedAccess produced as the CustomizeDiff.
// PlanResourceChange sets RawConfig on the prior state before calling Diff,
// which is what CustomizeDiff's GetRawConfig() reads; res.Diff alone leaves
// it null, so this mirrors that step.
func diffTokenVaultConfig(t *testing.T, tokenVaultPrivilegedAccess cty.Value) error {
	t.Helper()

	res := NewResource()
	sm := schema.InternalMap(res.Schema)
	impliedType := sm.CoreConfigSchema().ImpliedType()
	attrTypes := impliedType.AttributeTypes()

	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	vals["token_vault_privileged_access"] = tokenVaultPrivilegedAccess
	configVal := cty.ObjectVal(vals)

	priorState, err := res.ShimInstanceStateFromValue(cty.NullVal(impliedType))
	require.NoError(t, err)
	priorState.RawConfig = configVal

	resourceConfig := terraform.NewResourceConfigShimmed(configVal, sm.CoreConfigSchema())

	_, err = res.Diff(context.Background(), priorState, resourceConfig, nil)
	return err
}

func tokenVaultBlockWithGrants(grants ...cty.Value) cty.Value {
	return cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"credentials":  cty.ListValEmpty(tokenVaultCredentialsType().ElementType()),
			"ip_allowlist": cty.SetValEmpty(cty.String),
			"grants":       cty.ListVal(grants),
		}),
	})
}

func tokenVaultGrant(connection string, scopes ...string) cty.Value {
	scopeVals := make([]cty.Value, 0, len(scopes))
	for _, s := range scopes {
		scopeVals = append(scopeVals, cty.StringVal(s))
	}
	return cty.ObjectVal(map[string]cty.Value{
		"connection": cty.StringVal(connection),
		"scopes":     cty.SetVal(scopeVals),
	})
}

func TestValidateTokenVaultPrivilegedAccess(t *testing.T) {
	t.Run("passes with distinct connections and scopes within the cap", func(t *testing.T) {
		err := diffTokenVaultConfig(t, tokenVaultBlockWithGrants(
			tokenVaultGrant("conn-a", "s1", "s2"),
			tokenVaultGrant("conn-b", "s3"),
		))

		assert.NoError(t, err)
	})

	t.Run("passes with no block configured", func(t *testing.T) {
		err := diffTokenVaultConfig(t, cty.NullVal(tokenVaultBlockType()))

		assert.NoError(t, err)
	})

	t.Run("rejects more than 20 scopes across all grants", func(t *testing.T) {
		err := diffTokenVaultConfig(t, tokenVaultBlockWithGrants(
			tokenVaultGrant("conn-a", uniqueScopes("a", 11)...),
			tokenVaultGrant("conn-b", uniqueScopes("b", 10)...),
		))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not exceed 20")
	})

	t.Run("passes with exactly 20 scopes across all grants", func(t *testing.T) {
		err := diffTokenVaultConfig(t, tokenVaultBlockWithGrants(
			tokenVaultGrant("conn-a", uniqueScopes("a", 10)...),
			tokenVaultGrant("conn-b", uniqueScopes("b", 10)...),
		))

		assert.NoError(t, err)
	})

	t.Run("rejects a duplicate connection and names it in the error", func(t *testing.T) {
		err := diffTokenVaultConfig(t, tokenVaultBlockWithGrants(
			tokenVaultGrant("conn-a", "s1"),
			tokenVaultGrant("conn-a", "s2"),
		))

		require.Error(t, err)
		assert.Contains(t, err.Error(), `"conn-a"`)
		assert.Contains(t, err.Error(), "more than one grant")
	})
}

func tokenVaultCredentialState(id, credentialType, pem string) map[string]interface{} {
	return map[string]interface{}{
		"id":              id,
		"credential_type": credentialType,
		"pem":             pem,
	}
}

func tokenVaultCredentialConfig(credentialType, pem string) map[string]interface{} {
	return map[string]interface{}{
		"credential_type": credentialType,
		"pem":             pem,
	}
}

func TestMatchTokenVaultCredentials(t *testing.T) {
	t.Run("keeps an unchanged credential and creates or deletes nothing", func(t *testing.T) {
		old := []interface{}{tokenVaultCredentialState("cred_1", "public_key", "pem-a")}
		newCfg := []interface{}{tokenVaultCredentialConfig("public_key", "pem-a")}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(old, newCfg)

		assert.Equal(t, []string{"cred_1"}, keepIDs)
		assert.Empty(t, toCreate)
		assert.Empty(t, toDelete)
	})

	t.Run("creates a credential absent from the prior state and deletes nothing", func(t *testing.T) {
		newCfg := []interface{}{tokenVaultCredentialConfig("public_key", "pem-a")}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(nil, newCfg)

		assert.Empty(t, keepIDs)
		require.Len(t, toCreate, 1)
		assert.Equal(t, "public_key", toCreate[0].GetCredentialType())
		assert.Equal(t, "pem-a", toCreate[0].GetPEM())
		assert.Empty(t, toDelete)
	})

	t.Run("deletes a credential dropped from config and creates nothing", func(t *testing.T) {
		old := []interface{}{tokenVaultCredentialState("cred_1", "public_key", "pem-a")}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(old, nil)

		assert.Empty(t, keepIDs)
		assert.Empty(t, toCreate)
		assert.Equal(t, []string{"cred_1"}, toDelete)
	})

	t.Run("swaps material by creating the new credential and deleting the old one", func(t *testing.T) {
		old := []interface{}{tokenVaultCredentialState("cred_1", "public_key", "pem-old")}
		newCfg := []interface{}{tokenVaultCredentialConfig("public_key", "pem-new")}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(old, newCfg)

		assert.Empty(t, keepIDs)
		require.Len(t, toCreate, 1)
		assert.Equal(t, "pem-new", toCreate[0].GetPEM())
		assert.Equal(t, []string{"cred_1"}, toDelete)
	})

	t.Run("keeps one, creates one, when adding a second credential", func(t *testing.T) {
		old := []interface{}{tokenVaultCredentialState("cred_1", "public_key", "pem-a")}
		newCfg := []interface{}{
			tokenVaultCredentialConfig("public_key", "pem-a"),
			tokenVaultCredentialConfig("public_key", "pem-b"),
		}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(old, newCfg)

		assert.Equal(t, []string{"cred_1"}, keepIDs)
		require.Len(t, toCreate, 1)
		assert.Equal(t, "pem-b", toCreate[0].GetPEM())
		assert.Empty(t, toDelete)
	})

	t.Run("does not conflate two distinct credentials sharing identical material", func(t *testing.T) {
		old := []interface{}{
			tokenVaultCredentialState("cred_1", "public_key", "pem-a"),
			tokenVaultCredentialState("cred_2", "public_key", "pem-a"),
		}
		newCfg := []interface{}{tokenVaultCredentialConfig("public_key", "pem-a")}

		keepIDs, toCreate, toDelete := matchTokenVaultCredentials(old, newCfg)

		require.Len(t, keepIDs, 1)
		assert.Contains(t, []string{"cred_1", "cred_2"}, keepIDs[0])
		assert.Empty(t, toCreate)
		require.Len(t, toDelete, 1)
	})
}
