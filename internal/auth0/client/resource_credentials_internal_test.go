package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// generateTestRSAPEMs returns the same RSA public key encoded in the three PEM
// formats jwkThumbprint supports: SPKI ("PUBLIC KEY"), PKCS#1 ("RSA PUBLIC
// KEY"), and a self-signed X.509 certificate ("CERTIFICATE").
func generateTestRSAPEMs(t *testing.T) (spkiPEM, pkcs1PEM, certPEM string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	spkiDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	spki := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: spkiDER})

	pkcs1DER := x509.MarshalPKCS1PublicKey(&key.PublicKey)
	pkcs1 := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pkcs1DER})

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return string(spki), string(pkcs1), string(cert)
}

func TestJWKThumbprint_AllFormatsMatch(t *testing.T) {
	spkiPEM, pkcs1PEM, certPEM := generateTestRSAPEMs(t)

	spkiThumb := jwkThumbprint(spkiPEM)
	pkcs1Thumb := jwkThumbprint(pkcs1PEM)
	certThumb := jwkThumbprint(certPEM)

	assert.NotEmpty(t, spkiThumb, "SPKI thumbprint should be computed")
	assert.NotEmpty(t, pkcs1Thumb, "PKCS#1 thumbprint should be computed")
	assert.NotEmpty(t, certThumb, "X.509 certificate thumbprint should be computed")

	assert.Equal(t, spkiThumb, pkcs1Thumb,
		"same key in SPKI and PKCS#1 form must produce the same kid")
	assert.Equal(t, spkiThumb, certThumb,
		"same key in SPKI and X.509 certificate form must produce the same kid")
}

func TestJWKThumbprint_UnsupportedInputReturnsEmpty(t *testing.T) {
	assert.Empty(t, jwkThumbprint(""), "empty input returns empty")
	assert.Empty(t, jwkThumbprint("not a pem"), "garbage input returns empty")
	assert.Empty(t, jwkThumbprint(
		"-----BEGIN EC PRIVATE KEY-----\nMHcCAQE=\n-----END EC PRIVATE KEY-----\n"),
		"unsupported block type returns empty")
}

// formatRotation renders the planned steps as "-<removed id>" and "+<added name>"
// in order, so a case can state the whole expected sequence in one string.
func formatRotation(rotationSteps []rotationStep) string {
	rendered := make([]string, 0, len(rotationSteps))
	for _, step := range rotationSteps {
		switch step.kind {
		case detachAndDelete:
			rendered = append(rendered, "-"+step.credentialID)
		case createAndAttach:
			name, _ := step.newCredential["name"].(string)
			rendered = append(rendered, "+"+name)
		}
	}
	return strings.Join(rendered, " ")
}

// highestAttachedCount replays the steps and returns the largest attached count
// reached at any boundary.
func highestAttachedCount(rotationSteps []rotationStep, startingCount int) int {
	current, highest := startingCount, startingCount
	for _, step := range rotationSteps {
		switch step.kind {
		case detachAndDelete:
			current--
		case createAndAttach:
			current++
		}
		if current > highest {
			highest = current
		}
	}
	return highest
}

func TestPlanCredentialRotation(t *testing.T) {
	removals := func(ids ...string) []interface{} {
		entries := make([]interface{}, 0, len(ids))
		for _, id := range ids {
			entries = append(entries, map[string]interface{}{"id": id})
		}
		return entries
	}
	additions := func(names ...string) []interface{} {
		entries := make([]interface{}, 0, len(names))
		for _, name := range names {
			entries = append(entries, map[string]interface{}{"name": name})
		}
		return entries
	}

	cases := []struct {
		name           string
		toRemove       []interface{}
		toAdd          []interface{}
		attachedCount  int
		poolCount      int
		expectedSteps  string
		expectedReason string
	}{
		{
			name: "swap of two at the slot cap removes before each add",
			// Adding first here would ask the slot to hold 3.
			toRemove: removals("old-1", "old-2"), toAdd: additions("new-1", "new-2"),
			attachedCount: maxSlotCredentials, poolCount: 2,
			expectedSteps: "-old-1 +new-1 -old-2 +new-2",
		},
		{
			name: "full pool removes first even when the slot has room",
			// The remaining pool records belong to other features. Creating first
			// fails with "A client can have a maximum of 4 credentials".
			toRemove: removals("old-1"), toAdd: additions("new-1"),
			attachedCount: 1, poolCount: maxPoolCredentials,
			expectedSteps:  "-old-1 +new-1",
			expectedReason: "both counts are consulted, not just the slot",
		},
		{
			name:     "room in both containers adds first",
			toRemove: removals("old-1"), toAdd: additions("new-1"),
			attachedCount: 1, poolCount: maxPoolCredentials - 1,
			expectedSteps:  "+new-1 -old-1",
			expectedReason: "a usable credential stays attached throughout the swap",
		},
		{
			name:          "pure addition",
			toAdd:         additions("new-1"),
			expectedSteps: "+new-1",
		},
		{
			name:          "pure removal",
			toRemove:      removals("old-1"),
			attachedCount: 1, poolCount: 1,
			expectedSteps: "-old-1",
		},
		{
			name:     "more removals than additions pairs one, then trails the rest",
			toRemove: removals("old-1", "old-2"), toAdd: additions("new-1"),
			attachedCount: maxSlotCredentials, poolCount: 2,
			expectedSteps: "-old-1 +new-1 -old-2",
		},
		{
			name:     "more additions than removals pairs one, then trails the rest",
			toRemove: removals("old-1"), toAdd: additions("new-1", "new-2"),
			attachedCount: 1, poolCount: 1,
			expectedSteps: "+new-1 -old-1 +new-2",
		},
		{
			name:          "a removal without an ID is skipped",
			toRemove:      removals(""),
			attachedCount: 1, poolCount: 1,
			expectedSteps: "",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			rotationSteps := planCredentialRotation(
				credentialDiff{toAdd: testCase.toAdd, toRemove: testCase.toRemove},
				testCase.attachedCount, testCase.poolCount,
			)

			assert.Equal(t, testCase.expectedSteps, formatRotation(rotationSteps), testCase.expectedReason)
			assert.LessOrEqual(t,
				highestAttachedCount(rotationSteps, testCase.attachedCount), maxSlotCredentials,
				"the attached count must never exceed the slot cap mid-rotation")
		})
	}
}

func TestIsCredentialStillAttachedError(t *testing.T) {
	assert.False(t, isCredentialStillAttachedError(nil), "no error is not an attachment error")

	assert.True(t, isCredentialStillAttachedError(
		fmt.Errorf("400 Bad Request: Cannot delete credential still associated with a client")),
		"the API's refusal message must be recognised")

	assert.True(t, isCredentialStillAttachedError(
		fmt.Errorf("wrapped: %w",
			fmt.Errorf("400 Bad Request: Cannot delete credential still associated with a client"))),
		"a wrapped refusal must still be recognised")

	assert.False(t, isCredentialStillAttachedError(
		fmt.Errorf("404 Not Found: The credential does not exist")),
		"unrelated errors must not be swallowed")
}

// credentialsDataWithState builds a ResourceData whose prior state holds the
// given attributes, which is what the ownership helpers read via GetChange.
//
// The state is produced by writing the attributes through Set and serialising
// the result, so the SDK computes the TypeSet hashes. TestResourceDataRaw is not
// usable here: it diffs against a nil state, leaving everything in the diff and
// the prior value empty.
func credentialsDataWithState(t *testing.T, attributes map[string]interface{}) *schema.ResourceData {
	t.Helper()

	resource := NewCredentialsResource()

	writer := resource.Data(nil)
	writer.SetId("test-client-id")
	for attribute, valueToSet := range attributes {
		require.NoError(t, writer.Set(attribute, valueToSet))
	}

	data := resource.Data(writer.State())
	require.Equal(t, "test-client-id", data.Id())

	return data
}

func TestStateCredentialIDs_ReadsOnlyTheRequestedSlot(t *testing.T) {
	data := credentialsDataWithState(t, map[string]interface{}{
		"private_key_jwt": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-pkjwt-1", "credential_type": "public_key", "pem": "pem-1"},
					map[string]interface{}{"id": "cred-pkjwt-2", "credential_type": "public_key", "pem": "pem-2"},
				},
			},
		},
		"signed_request_object": []interface{}{
			map[string]interface{}{
				"required": true,
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-sro-1", "credential_type": "public_key", "pem": "pem-3"},
				},
			},
		},
	})

	assert.ElementsMatch(t, []string{"cred-pkjwt-1", "cred-pkjwt-2"},
		stateCredentialIDs(data, "private_key_jwt"),
		"a slot read must not include another slot's credentials")

	assert.Equal(t, []string{"cred-sro-1"},
		stateCredentialIDs(data, "signed_request_object"))

	assert.Empty(t, stateCredentialIDs(data, "tls_client_auth"),
		"an unpopulated slot has no credentials")
}

func TestStateCredentialIDs_SkipsEntriesWithoutAnID(t *testing.T) {
	// A credential added by the current config has no ID until it is created.
	data := credentialsDataWithState(t, map[string]interface{}{
		"private_key_jwt": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-1", "credential_type": "public_key", "pem": "pem-1"},
					map[string]interface{}{"credential_type": "public_key", "pem": "pem-2"},
				},
			},
		},
	})

	assert.Equal(t, []string{"cred-1"}, stateCredentialIDs(data, "private_key_jwt"))
}

func TestOwnedCredentialIDs_SpansEveryCredentialBearingAttribute(t *testing.T) {
	data := credentialsDataWithState(t, map[string]interface{}{
		"private_key_jwt": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-pkjwt", "credential_type": "public_key", "pem": "pem-1"},
				},
			},
		},
		"signed_request_object": []interface{}{
			map[string]interface{}{
				"required": false,
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-sro", "credential_type": "public_key", "pem": "pem-2"},
				},
			},
		},
	})

	ownedIDs := ownedCredentialIDs(data)

	assert.True(t, ownedIDs["cred-pkjwt"], "private_key_jwt credentials are owned")
	assert.True(t, ownedIDs["cred-sro"], "signed_request_object credentials are owned")
	assert.False(t, ownedIDs["cred-created-elsewhere"],
		"a credential absent from state is not owned and must never be deleted")
	assert.Len(t, ownedIDs, 2)
}

func TestAttachedCredentialsFromState_ReturnsIDReferencesForOneSlot(t *testing.T) {
	data := credentialsDataWithState(t, map[string]interface{}{
		"private_key_jwt": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-pkjwt", "credential_type": "public_key", "pem": "pem-1"},
				},
			},
		},
		"signed_request_object": []interface{}{
			map[string]interface{}{
				"required": false,
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-sro", "credential_type": "public_key", "pem": "pem-2"},
				},
			},
		},
	})

	attached := attachedCredentialsFromState(data, "private_key_jwt")

	require.Len(t, attached, 1,
		"the attach payload must carry only this slot's credentials")
	assert.Equal(t, "cred-pkjwt", attached[0].GetID())
	assert.Nil(t, attached[0].PEM, "an attach payload carries references, not key material")
}

func TestRemoveAttachedCredential(t *testing.T) {
	firstCredentialID, secondCredentialID := "cred-1", "cred-2"
	attachedCredentials := []management.Credential{{ID: &firstCredentialID}, {ID: &secondCredentialID}}

	remainingCredentials := removeAttachedCredential(attachedCredentials, "cred-1")

	require.Len(t, remainingCredentials, 1)
	assert.Equal(t, "cred-2", remainingCredentials[0].GetID())
}

func pkjwtBlock(creds ...map[string]interface{}) []interface{} {
	list := make([]interface{}, 0, len(creds))
	for _, c := range creds {
		list = append(list, c)
	}
	return []interface{}{map[string]interface{}{"credentials": list}}
}

func diffData(t *testing.T, prior map[string]interface{}, config map[string]interface{}) *schema.ResourceData {
	t.Helper()
	resource := NewCredentialsResource()
	writer := resource.Data(nil)
	writer.SetId("test-client-id")
	for k, v := range prior {
		require.NoError(t, writer.Set(k, v))
	}
	priorState := writer.State()
	priorState.ID = "test-client-id"

	sm := schema.InternalMap(resource.SchemaMap())
	diff, err := sm.Diff(context.Background(), priorState, terraform.NewResourceConfigRaw(config), nil, nil, true)
	require.NoError(t, err)
	data, err := sm.Data(priorState, diff)
	require.NoError(t, err)
	return data
}

// Switching authentication_method away from private_key_jwt. The config no
// longer declares the block, so the PLANNED value is empty. Ownership must still
// resolve from prior state, or the owned credential leaks.
func TestOwnedCredentialIDs_SurvivesAuthMethodSwitch(t *testing.T) {
	data := diffData(t,
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"id": "cred-mine", "name": "k1", "credential_type": "public_key",
				"pem": "PEM-ONE", "algorithm": "RS256",
			}),
		},
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "client_secret_post",
		},
	)

	owned := ownedCredentialIDs(data)
	assert.True(t, owned["cred-mine"],
		"an owned credential must still be deletable after the block leaves the config")
	assert.Len(t, owned, 1)
}

// Switching authentication_method while the Token Vault block stays in the
// configuration. The detach that precedes the deletion clears only the
// authentication method and signed_request_object, so a Token Vault credential is
// still attached and the API refuses to delete it. The switch-away deletion set
// must therefore exclude the block.
func TestDetachedCredentialIDs_ExcludesTokenVaultPrivilegedAccess(t *testing.T) {
	tokenVault := []interface{}{map[string]interface{}{
		"credentials": []interface{}{map[string]interface{}{
			"id": "cred-tvpa", "name": "tv", "credential_type": "public_key",
			"pem": "PEM-TVPA", "algorithm": "RS256",
		}},
		"ip_allowlist": []interface{}{"10.0.0.1"},
		"grants": []interface{}{map[string]interface{}{
			"connection": "google-oauth2",
			"scopes":     []interface{}{"calendar.readonly"},
		}},
	}}

	data := diffData(t,
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"id": "cred-pkjwt", "name": "k1", "credential_type": "public_key",
				"pem": "PEM-ONE", "algorithm": "RS256",
			}),
			"token_vault_privileged_access": tokenVault,
		},
		map[string]interface{}{
			"client_id":                     "test-client-id",
			"authentication_method":         "client_secret_post",
			"token_vault_privileged_access": tokenVault,
		},
	)

	// Ownership is unchanged: the worker's credential is still ours to manage, and
	// the destroy path relies on that to tear it down.
	assert.True(t, ownedCredentialIDs(data)["cred-tvpa"],
		"the Token Vault credential is still owned by this resource")

	detached := detachedCredentialIDs(data)
	assert.True(t, detached["cred-pkjwt"],
		"the authentication method's credential is released by the detach and must be deleted")
	assert.False(t, detached["cred-tvpa"],
		"the Token Vault credential is still attached, so deleting it here would fail the update")
}

// The spurious-rotation trigger: private_key_jwt credentials are IDENTICAL
// in state and config; only signed_request_object.required changed. The rotation
// must plan nothing.
func TestCredentialChanges_IgnoresUnchangedSetOnUnrelatedApply(t *testing.T) {
	data := diffData(t,
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"id": "cred-pkjwt", "name": "k1", "credential_type": "public_key",
				"pem": "PEM-ONE", "algorithm": "RS256",
			}),
			"signed_request_object": []interface{}{map[string]interface{}{
				"required": false,
				"credentials": []interface{}{map[string]interface{}{
					"id": "cred-sro", "name": "sro", "credential_type": "public_key",
					"pem": "PEM-SRO", "algorithm": "RS256",
				}},
			}},
		},
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"name": "k1", "credential_type": "public_key", "pem": "PEM-ONE", "algorithm": "RS256",
			}),
			"signed_request_object": []interface{}{map[string]interface{}{
				"required": true,
				"credentials": []interface{}{map[string]interface{}{
					"name": "sro", "credential_type": "public_key", "pem": "PEM-SRO", "algorithm": "RS256",
				}},
			}},
		},
	)

	require.False(t, data.HasChange("private_key_jwt.0.credentials"),
		"the credential set is identical")

	// The trap the guard exists for: on an unchanged key Difference has no prior
	// value to subtract, so it reports every entry as an addition.
	rawAdd, _ := value.Difference(data, "private_key_jwt.0.credentials")
	require.NotEmpty(t, rawAdd)

	diff := credentialChanges(data, "private_key_jwt.0.credentials")
	assert.Empty(t, planCredentialRotation(diff, 1, 2),
		"an unchanged credential set must not create a duplicate credential")
}

func TestFilterOwnedCredentials(t *testing.T) {
	spkiPEM, _, _ := generateTestRSAPEMs(t)
	ownKeyID := jwkThumbprint(spkiPEM)
	require.NotEmpty(t, ownKeyID)

	foreign := management.Credential{
		ID:    auth0.String("cred-foreign"),
		Name:  auth0.String("set-up-elsewhere"),
		KeyID: auth0.String("some-other-thumbprint"),
	}

	t.Run("a credential state records is kept, one it does not is dropped", func(t *testing.T) {
		data := credentialsDataWithState(t, map[string]interface{}{
			"signed_request_object": []interface{}{map[string]interface{}{
				"required": false,
				"credentials": []interface{}{map[string]interface{}{
					"id": "cred-mine", "credential_type": "public_key", "pem": spkiPEM,
				}},
			}},
		})

		owned, skipped := filterOwnedCredentials(data, "signed_request_object", []management.Credential{
			{ID: auth0.String("cred-mine"), KeyID: auth0.String(ownKeyID)},
			foreign,
		})

		require.Len(t, owned, 1)
		assert.Equal(t, "cred-mine", owned[0].GetID())
		assert.Equal(t, []string{"cred-foreign"}, skipped,
			"the dropped credential must be reported so it can be logged")
	})

	t.Run("a credential just created is matched by its key thumbprint", func(t *testing.T) {
		// The read after a create is what puts the ID into state, so ownership
		// cannot come from state here. The configured PEM is the only signal.
		data := credentialsDataWithState(t, map[string]interface{}{
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"credential_type": "public_key", "pem": spkiPEM, "algorithm": "RS256",
			}),
		})

		owned, skipped := filterOwnedCredentials(data, "private_key_jwt", []management.Credential{
			{ID: auth0.String("cred-brand-new"), KeyID: auth0.String(ownKeyID)},
			foreign,
		})

		require.Len(t, owned, 1, "the new credential must be recorded, not filtered away")
		assert.Equal(t, "cred-brand-new", owned[0].GetID())
		assert.Equal(t, []string{"cred-foreign"}, skipped)
	})

	t.Run("import records the listing whole", func(t *testing.T) {
		// Import carries neither prior state nor a configuration, so nothing can be
		// attributed and everything the slot holds is recorded.
		data := credentialsDataWithState(t, map[string]interface{}{})

		owned, skipped := filterOwnedCredentials(data, "private_key_jwt",
			[]management.Credential{foreign})

		require.Len(t, owned, 1)
		assert.Empty(t, skipped)
	})

	t.Run("a credential with no recognisable field disables filtering", func(t *testing.T) {
		// Nothing here identifies the declared credential, so ours cannot be told
		// from anyone else's. Returning less than was declared fails the apply with
		// "provider produced inconsistent result after apply", so the listing stands.
		data := credentialsDataWithState(t, map[string]interface{}{
			"private_key_jwt": pkjwtBlock(map[string]interface{}{
				"credential_type": "public_key",
			}),
		})

		owned, skipped := filterOwnedCredentials(data, "private_key_jwt",
			[]management.Credential{foreign})

		require.Len(t, owned, 1, "an unidentifiable declaration must not shrink state")
		assert.Empty(t, skipped)
	})

	t.Run("a managed credential deleted out of band does not adopt a slot mate", func(t *testing.T) {
		// Two credentials declared, one deleted elsewhere. The shortfall is drift,
		// not a matching failure, so the foreign credential sharing the slot must
		// still be left alone.
		data := credentialsDataWithState(t, map[string]interface{}{
			"private_key_jwt": []interface{}{map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{
						"id": "cred-mine", "name": "tf-a",
						"credential_type": "public_key", "pem": spkiPEM, "algorithm": "RS256",
					},
					map[string]interface{}{
						"id": "cred-deleted-elsewhere", "name": "tf-b",
						"credential_type": "public_key", "pem": spkiPEM, "algorithm": "RS256",
					},
				},
			}},
		})

		owned, skipped := filterOwnedCredentials(data, "private_key_jwt", []management.Credential{
			{ID: auth0.String("cred-mine"), KeyID: auth0.String(ownKeyID)},
			foreign,
		})

		require.Len(t, owned, 1, "only the surviving managed credential belongs in state")
		assert.Equal(t, "cred-mine", owned[0].GetID())
		assert.Equal(t, []string{"cred-foreign"}, skipped)
	})

	t.Run("a cert_subject_dn credential is matched by its subject", func(t *testing.T) {
		data := credentialsDataWithState(t, map[string]interface{}{
			"tls_client_auth": []interface{}{map[string]interface{}{
				"credentials": []interface{}{map[string]interface{}{
					"credential_type": "cert_subject_dn", "subject_dn": "CN=mine",
				}},
			}},
		})

		owned, skipped := filterOwnedCredentials(data, "tls_client_auth", []management.Credential{
			{ID: auth0.String("cred-mine"), SubjectDN: auth0.String("CN=mine")},
			{ID: auth0.String("cred-theirs"), SubjectDN: auth0.String("CN=theirs")},
		})

		require.Len(t, owned, 1)
		assert.Equal(t, "cred-mine", owned[0].GetID())
		assert.Equal(t, []string{"cred-theirs"}, skipped)
	})

	t.Run("an x509_cert credential without a name is matched by its certificate thumbprint", func(t *testing.T) {
		// The name field is optional on this block, and the API returns neither a
		// key ID nor a subject for an x509_cert credential. Only thumbprint_sha256
		// identifies it, and that is computable from the configured certificate.
		// Without this rung a self-signed mTLS user who omitted the name loses
		// their credential from state on the read that follows create.
		_, _, certPEM := generateTestRSAPEMs(t)

		data := credentialsDataWithState(t, map[string]interface{}{
			"self_signed_tls_client_auth": []interface{}{map[string]interface{}{
				"credentials": []interface{}{map[string]interface{}{
					"credential_type": "x509_cert", "pem": certPEM,
				}},
			}},
		})

		owned, skipped := filterOwnedCredentials(data, "self_signed_tls_client_auth", []management.Credential{
			{
				ID:               auth0.String("cred-mine"),
				CredentialType:   auth0.String("x509_cert"),
				ThumbprintSHA256: auth0.String(certificateThumbprint(certPEM)),
			},
			{
				ID:               auth0.String("cred-theirs"),
				CredentialType:   auth0.String("x509_cert"),
				ThumbprintSHA256: auth0.String("a-different-certificate-digest"),
			},
		})

		require.Len(t, owned, 1, "the user's own certificate must stay in state")
		assert.Equal(t, "cred-mine", owned[0].GetID())
		assert.Equal(t, []string{"cred-theirs"}, skipped)
	})

	t.Run("a cert_subject_dn credential declared by pem alone disables filtering", func(t *testing.T) {
		// The schema lets pem stand in for subject_dn here, and name is optional.
		// The API returns no key_id and no thumbprint_sha256 for this type, so a
		// digest of the certificate matches nothing it sends back. The entry must
		// count as unmatchable rather than look identifiable, or the credential is
		// dropped from state.
		_, _, certPEM := generateTestRSAPEMs(t)

		data := credentialsDataWithState(t, map[string]interface{}{
			"tls_client_auth": []interface{}{map[string]interface{}{
				"credentials": []interface{}{map[string]interface{}{
					"credential_type": "cert_subject_dn", "pem": certPEM,
				}},
			}},
		})

		matchers := desiredCredentialMatchers(data, "tls_client_auth")
		assert.Equal(t, 1, matchers.unmatchable,
			"a pem-only cert_subject_dn entry carries nothing the API echoes back")

		owned, skipped := filterOwnedCredentials(data, "tls_client_auth", []management.Credential{
			{
				ID:             auth0.String("cred-mine"),
				CredentialType: auth0.String("cert_subject_dn"),
				SubjectDN:      auth0.String("CN=derived-by-the-api"),
			},
		})

		require.Len(t, owned, 1, "an unidentifiable declaration must not shrink state")
		assert.Empty(t, skipped)
	})
}

func TestCertificateThumbprint(t *testing.T) {
	spkiPEM, pkcs1PEM, certPEM := generateTestRSAPEMs(t)

	t.Run("matches the digest the API reports for a recorded certificate", func(t *testing.T) {
		// The file creds-cert-1.pem is the certificate used by the
		// self_signed_tls_client_auth acceptance test, and the value below is the
		// thumbprint_sha256 the API answered with in the recorded cassette. It pins
		// the digest to real API behaviour rather than to our own restatement of it.
		recorded, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
		require.NoError(t, err)

		assert.Equal(t,
			"irPW_52cQKIRPeblpAs2LjaUELUxnh5Vi7_1idgHX7w",
			certificateThumbprint(string(recorded)))
	})

	t.Run("is empty for anything that is not a certificate", func(t *testing.T) {
		// A public key carries no certificate digest. Returning one anyway would
		// make a public_key entry look matchable on a field the API never sends.
		assert.Empty(t, certificateThumbprint(spkiPEM))
		assert.Empty(t, certificateThumbprint(pkcs1PEM))
		assert.Empty(t, certificateThumbprint(""))
		assert.Empty(t, certificateThumbprint("not a pem at all"))
	})

	t.Run("differs from the jwk thumbprint of the same certificate", func(t *testing.T) {
		// The two digests cover different bytes, so they must never be interchanged:
		// key_id is the JWK thumbprint of the public key, thumbprint_sha256 is the
		// digest of the whole certificate.
		assert.NotEqual(t, jwkThumbprint(certPEM), certificateThumbprint(certPEM))
	})
}

func TestExpandClientCredential(t *testing.T) {
	// The helper below builds the object shape expandClientCredential reads. A raw
	// configuration carries every attribute of the block, null where the
	// configuration said nothing, which is what makes an omitted credential_type
	// arrive as a null string rather than as an absent key.
	credentialRawConfig := func(credentialType cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"credential_type":        credentialType,
			"pem":                    cty.StringVal("-----BEGIN CERTIFICATE-----\nAA==\n-----END CERTIFICATE-----\n"),
			"name":                   cty.NullVal(cty.String),
			"algorithm":              cty.NullVal(cty.String),
			"parse_expiry_from_cert": cty.NullVal(cty.Bool),
			"expires_at":             cty.NullVal(cty.String),
			"subject_dn":             cty.NullVal(cty.String),
		})
	}

	t.Run("an omitted credential_type takes the one its block accepts", func(t *testing.T) {
		// Only self_signed_tls_client_auth makes credential_type Optional, so this
		// configuration is valid. Dereferencing the absent value used to panic the
		// provider, and sending it empty had the API answer 400 'Missing required
		// property: credential_type'.
		for _, credentialType := range []cty.Value{cty.NullVal(cty.String), cty.StringVal("")} {
			credential := expandClientCredential(credentialRawConfig(credentialType), "x509_cert")

			require.NotNil(t, credential.CredentialType)
			assert.Equal(t, "x509_cert", credential.GetCredentialType())
			assert.NotEmpty(t, credential.GetPEM(), "the certificate must still be sent")
		}
	})

	t.Run("a configured credential_type is never overridden", func(t *testing.T) {
		credential := expandClientCredential(credentialRawConfig(cty.StringVal("cert_subject_dn")), "x509_cert")

		assert.Equal(t, "cert_subject_dn", credential.GetCredentialType())
	})

	t.Run("every authentication method has a credential type to fall back on", func(t *testing.T) {
		// A missing entry would default to the empty string, which is the 400 the
		// default exists to prevent. The signed_request_object block is absent by
		// design: it requires credential_type, and it is not an authentication method.
		for _, authenticationMethod := range []string{"private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth"} {
			assert.NotEmpty(t, credentialTypeByAuthenticationMethod[authenticationMethod], authenticationMethod)
		}
	})
}

// rawSource builds the object shape the SDK hands to GetRawConfig/GetRawState:
// client_id plus one list per credential-bearing block, each holding declared
// entries. A blockLengths entry of -1 makes that block null.
func rawSource(clientID interface{}, blockLengths map[string]int) cty.Value {
	attributes := map[string]cty.Value{}

	if clientID == nil {
		attributes["client_id"] = cty.NullVal(cty.String)
	} else {
		attributes["client_id"] = cty.StringVal(clientID.(string))
	}

	for _, attribute := range credentialBearingAttributes {
		blockType := cty.Object(map[string]cty.Type{"placeholder": cty.String})
		length, ok := blockLengths[attribute]
		if !ok || length == 0 {
			attributes[attribute] = cty.ListValEmpty(blockType)
			continue
		}
		if length < 0 {
			attributes[attribute] = cty.NullVal(cty.List(blockType))
			continue
		}

		entries := make([]cty.Value, 0, length)
		for range length {
			entries = append(entries, cty.ObjectVal(map[string]cty.Value{
				"placeholder": cty.StringVal("x"),
			}))
		}
		attributes[attribute] = cty.ListVal(entries)
	}

	return cty.ObjectVal(attributes)
}

func TestBlockDeclaredIn(t *testing.T) {
	nullObject := cty.NullVal(cty.EmptyObject)

	t.Run("the configuration decides during an apply", func(t *testing.T) {
		config := rawSource("test-client-id", map[string]int{"private_key_jwt": 1})
		state := rawSource("test-client-id", map[string]int{"private_key_jwt": 1})

		assert.True(t, blockDeclaredIn("private_key_jwt", config, state))
		assert.False(t, blockDeclaredIn("signed_request_object", config, state),
			"a block absent from the configuration belongs to whoever set it up")
	})

	t.Run("a block being removed is decided by the configuration, not prior state", func(t *testing.T) {
		// The apply that drops a block must still see it as undeclared, or the read
		// at the end of it re-adopts what the configuration just gave up.
		config := rawSource("test-client-id", nil)
		state := rawSource("test-client-id", map[string]int{"signed_request_object": 1})

		assert.False(t, blockDeclaredIn("signed_request_object", config, state))
	})

	t.Run("prior state decides during a refresh", func(t *testing.T) {
		// ReadResource populates only RawState; GetRawConfig() returns a null object.
		state := rawSource("test-client-id", map[string]int{"signed_request_object": 1})

		assert.True(t, blockDeclaredIn("signed_request_object", nullObject, state))
		assert.False(t, blockDeclaredIn("private_key_jwt", nullObject, state))
	})

	t.Run("import records every block", func(t *testing.T) {
		// Import does not present as a null raw state: the SDK supplies an object
		// with every block empty, so client_id is what distinguishes it. Import has
		// not read one yet and every other phase carries it.
		state := rawSource(nil, nil)

		for _, attribute := range credentialBearingAttributes {
			assert.True(t, blockDeclaredIn(attribute, nullObject, state),
				"%s must be recorded on import, which has nothing to compare against", attribute)
		}
	})

	t.Run("a null block falls through to the next source", func(t *testing.T) {
		config := rawSource("test-client-id", map[string]int{"signed_request_object": -1})
		state := rawSource("test-client-id", map[string]int{"signed_request_object": 1})

		assert.True(t, blockDeclaredIn("signed_request_object", config, state))
	})

	t.Run("with no usable source the response is recorded", func(t *testing.T) {
		assert.True(t, blockDeclaredIn("private_key_jwt", nullObject, nullObject))
		assert.True(t, blockDeclaredIn("private_key_jwt",
			cty.UnknownVal(cty.EmptyObject), cty.NullVal(cty.String)))
	})
}

func TestRetainExistingCredentials(t *testing.T) {
	fromState := []management.Credential{
		{ID: auth0.String("cred-live")},
		{ID: auth0.String("cred-deleted-out-of-band")},
	}
	pool := []*management.Credential{
		{ID: auth0.String("cred-live")},
		{ID: auth0.String("cred-someone-elses")},
	}

	retained := retainExistingCredentials(fromState, pool)

	require.Len(t, retained, 1,
		"an ID state remembers but the client no longer holds must be dropped")
	assert.Equal(t, "cred-live", retained[0].GetID())

	assert.Empty(t, retainExistingCredentials(fromState, nil),
		"an empty pool retains nothing")
}

// TestRollbackCreatedCredentials covers the create path's cleanup. A failed
// attach used to leave the credentials it had just created sitting in the pool,
// where they consumed the four-record cap and made every later apply fail with
// "credentials contains public keys that already exist in client".
func TestRollbackCreatedCredentials(t *testing.T) {
	newAPI := func(t *testing.T, handler http.Handler) *management.Management {
		t.Helper()
		server := httptest.NewServer(handler)
		t.Cleanup(server.Close)

		api, err := management.New(strings.TrimPrefix(server.URL, "http://"),
			management.WithStaticToken("test-token"), management.WithInsecure())
		require.NoError(t, err)

		return api
	}

	t.Run("it deletes every credential the failed apply created", func(t *testing.T) {
		var deleted []string
		api := newAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				deleted = append(deleted, path.Base(r.URL.Path))
			}
			w.WriteHeader(http.StatusNoContent)
		}))

		created := []management.Credential{
			{ID: auth0.String("cred-one")},
			{ID: auth0.String("cred-two")},
		}
		cause := fmt.Errorf("attach failed")

		err := rollbackCreatedCredentials(context.Background(), api, "client-id", created, cause)

		assert.ErrorContains(t, err, "attach failed",
			"the original failure must still reach the user")
		assert.Equal(t, []string{"cred-one", "cred-two"}, deleted,
			"both created credentials must be removed from the pool")
	})

	t.Run("it reports a rollback failure alongside the original error", func(t *testing.T) {
		api := newAPI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}))

		err := rollbackCreatedCredentials(context.Background(), api, "client-id",
			[]management.Credential{{ID: auth0.String("cred-one")}},
			fmt.Errorf("attach failed"))

		assert.ErrorContains(t, err, "attach failed")
		assert.ErrorContains(t, err, "rollback of credential cred-one also failed")
	})

	t.Run("a credential with no ID is skipped and nothing is created to roll back", func(t *testing.T) {
		calls := 0
		api := newAPI(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			w.WriteHeader(http.StatusNoContent)
		}))

		cause := fmt.Errorf("create failed")
		err := rollbackCreatedCredentials(context.Background(), api, "client-id",
			[]management.Credential{{}}, cause)

		assert.ErrorContains(t, err, "create failed")
		assert.Zero(t, calls, "an empty ID must not produce a delete call")

		require.ErrorContains(t,
			rollbackCreatedCredentials(context.Background(), api, "client-id", nil, cause),
			"create failed")
		assert.Zero(t, calls, "nothing created means nothing to roll back")
	})
}

// tokenVaultConfig builds a cty value shaped like the resource's raw
// configuration, carrying only the token_vault_privileged_access block. The
// expanders reach for that one attribute, so the rest of the schema is omitted.
func tokenVaultConfig(t *testing.T, block cty.Value) cty.Value {
	t.Helper()

	return cty.ObjectVal(map[string]cty.Value{
		"token_vault_privileged_access": block,
	})
}

func TestExpandTokenVaultPrivilegedAccess_AlwaysSendsBothSubFields(t *testing.T) {
	// The API rejects a write that omits any sub-field, and the SDK omits the key
	// when the pointer is nil, so an empty block must still marshal both keys.
	tokenVault := expandTokenVaultPrivilegedAccess(tokenVaultConfig(t, cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"credentials":  cty.SetValEmpty(cty.Object(map[string]cty.Type{"pem": cty.String})),
			"ip_allowlist": cty.SetValEmpty(cty.String),
			"grants":       cty.SetValEmpty(cty.Object(map[string]cty.Type{"connection": cty.String, "scopes": cty.Set(cty.String)})),
		}),
	})))

	require.NotNil(t, tokenVault)
	require.NotNil(t, tokenVault.IPAllowlist, "a nil ip_allowlist would omit the key and 400")
	require.NotNil(t, tokenVault.Grants, "a nil grants would omit the key and 400")
	assert.Empty(t, *tokenVault.IPAllowlist)
	assert.Empty(t, *tokenVault.Grants)

	assert.Nil(t, tokenVault.Credentials,
		"credentials are attached by ID once created, not expanded from config")
}

// TestAttachTokenVaultPrivilegedAccessPayload_EmptyListsClearTheFields pins the
// wire format for the clear case, which is what `ip_allowlist = []` means on
// PATCH. The omitempty tag on a pointer field keys off the pointer, not the
// length, so a non-nil pointer to an empty slice sends `[]` and clears them. The
// map round-trip is asserted too because updateClientInternal marshals to a map
// before sending, and a naive shape there would drop the keys again.
func TestAttachTokenVaultPrivilegedAccessPayload_EmptyListsClearTheFields(t *testing.T) {
	emptyAllowlist := make([]string, 0)
	emptyGrants := make([]management.ClientTokenVaultPrivilegedGrant, 0)

	payload := clientWithTokenVaultPrivilegedAccess{
		ID: "client-id",
		TokenVaultPrivilegedAccess: &management.ClientTokenVaultPrivilegedAccess{
			Credentials: &[]management.Credential{},
			IPAllowlist: &emptyAllowlist,
			Grants:      &emptyGrants,
		},
	}

	encoded, err := json.Marshal(payload)
	require.NoError(t, err)

	assert.JSONEq(t,
		`{"token_vault_privileged_access":{"credentials":[],"ip_allowlist":[],"grants":[]}}`,
		string(encoded),
		"an empty list must reach the API as [], which is what clears the field")

	var asMap map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &asMap))
	roundTripped, err := json.Marshal(asMap)
	require.NoError(t, err)

	assert.JSONEq(t,
		`{"token_vault_privileged_access":{"credentials":[],"ip_allowlist":[],"grants":[]}}`,
		string(roundTripped),
		"the map round-trip updateClientInternal performs must preserve the empty lists")
}

// TestRemoveTokenVaultPrivilegedAccessPayload_SendsNull covers the other half of
// the contract: clearing a sub-field is not the same as removing the block, and
// the two must not collapse into one payload.
func TestRemoveTokenVaultPrivilegedAccessPayload_SendsNull(t *testing.T) {
	encoded, err := json.Marshal(clientWithTokenVaultPrivilegedAccess{ID: "client-id"})
	require.NoError(t, err)

	assert.JSONEq(t, `{"token_vault_privileged_access":null}`, string(encoded),
		"a nil pointer must marshal to a literal null, which removes the whole object")
}

func TestExpandTokenVaultPrivilegedAccess_AbsentBlockIsNil(t *testing.T) {
	nullBlock := cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
		"credentials":  cty.Set(cty.Object(map[string]cty.Type{"pem": cty.String})),
		"ip_allowlist": cty.Set(cty.String),
		"grants":       cty.Set(cty.Object(map[string]cty.Type{"connection": cty.String, "scopes": cty.Set(cty.String)})),
	})))

	assert.Nil(t, expandTokenVaultPrivilegedAccess(tokenVaultConfig(t, nullBlock)),
		"an absent block must not be written, which is what makes removal detectable")
}

func TestExpandTokenVaultPrivilegedAccess_ReadsAllowlistAndGrants(t *testing.T) {
	tokenVault := expandTokenVaultPrivilegedAccess(tokenVaultConfig(t, cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"credentials": cty.SetValEmpty(cty.Object(map[string]cty.Type{"pem": cty.String})),
			"ip_allowlist": cty.SetVal([]cty.Value{
				cty.StringVal("10.0.0.1"),
				cty.StringVal("192.168.1.0/24"),
			}),
			"grants": cty.SetVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"connection": cty.StringVal("google-oauth2"),
					"scopes":     cty.SetVal([]cty.Value{cty.StringVal("calendar.readonly")}),
				}),
			}),
		}),
	})))

	require.NotNil(t, tokenVault)
	assert.ElementsMatch(t, []string{"10.0.0.1", "192.168.1.0/24"}, *tokenVault.IPAllowlist)

	require.Len(t, *tokenVault.Grants, 1)
	grant := (*tokenVault.Grants)[0]
	assert.Equal(t, "google-oauth2", grant.GetConnection())
	assert.Equal(t, []string{"calendar.readonly"}, grant.GetScopes())
}

func TestOwnedCredentialIDs_IncludesTokenVaultPrivilegedAccess(t *testing.T) {
	// The destroy path deletes only what this map contains. Were the Token Vault
	// block missing from credentialBearingAttributes, its credentials would be
	// detached and then left behind in the pool.
	data := credentialsDataWithState(t, map[string]interface{}{
		"token_vault_privileged_access": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"id": "cred-tvpa", "credential_type": "public_key", "pem": "pem-1"},
				},
				"ip_allowlist": []interface{}{"10.0.0.1"},
				"grants": []interface{}{
					map[string]interface{}{
						"connection": "google-oauth2",
						"scopes":     []interface{}{"calendar.readonly"},
					},
				},
			},
		},
	})

	assert.True(t, ownedCredentialIDs(data)["cred-tvpa"],
		"a Token Vault credential recorded in state is owned by this resource")

	assert.Equal(t, []string{"cred-tvpa"},
		stateCredentialIDs(data, "token_vault_privileged_access"),
		"the block's credentials must be reachable as its own slot")
}

// credentialsDataWithRawState returns resource data whose RawState carries the
// given block, which is the source credentialBlockDeclared reads during a refresh
// or destroy. Set() alone does not populate it, and with both raw sources null the
// helper defaults to declared.
func credentialsDataWithRawState(t *testing.T, rawState cty.Value) *schema.ResourceData {
	t.Helper()

	resource := NewCredentialsResource()

	return resource.Data(&terraform.InstanceState{
		ID:       "test-client-id",
		RawState: rawState,
	})
}

func TestCredentialBlockDeclared_TokenVaultGatesTheDestroyRemoval(t *testing.T) {
	// The destroy path only sends the removal PATCH when the block is declared, so
	// a client holding a worker configured elsewhere is left alone.
	implied := NewCredentialsResource().CoreConfigSchema().ImpliedType()

	nullState := make(map[string]cty.Value, len(implied.AttributeTypes()))
	for name, attributeType := range implied.AttributeTypes() {
		nullState[name] = cty.NullVal(attributeType)
	}

	// The client_id attribute is what tells a real refresh apart from an import,
	// and a source without it is skipped entirely. Every case below is a
	// post-import phase, so it has to be present or block lengths are never read.
	nullState["client_id"] = cty.StringVal("test-client-id")

	tokenVaultType := implied.AttributeType("token_vault_privileged_access")

	undeclaredState := make(map[string]cty.Value, len(nullState))
	for name, attributeValue := range nullState {
		undeclaredState[name] = attributeValue
	}
	undeclaredState["token_vault_privileged_access"] = cty.ListValEmpty(tokenVaultType.ElementType())

	assert.False(t,
		credentialBlockDeclared(credentialsDataWithRawState(t, cty.ObjectVal(undeclaredState)),
			"token_vault_privileged_access"),
		"a worker this resource never declared must not be torn down")

	// With no raw config or state to compare against — import, and the read after
	// it — the block is treated as declared so the response is still recorded.
	assert.True(t,
		credentialBlockDeclared(credentialsDataWithRawState(t, cty.NullVal(implied)),
			"token_vault_privileged_access"),
		"import carries no prior state, so the response is recorded as before")

	// State upgraded from a provider version that predates this attribute holds
	// null for it rather than an empty list, so credentialBlockDeclared falls
	// through to its declared default. That is why the destroy gate keys off the
	// credential IDs in state instead: no IDs, no worker, no PATCH. Gating on
	// credentialBlockDeclared here would PATCH the field on every destroy and 403
	// on any tenant without the entitlement.
	upgraded := credentialsDataWithRawState(t, cty.ObjectVal(nullState))

	assert.True(t,
		credentialBlockDeclared(upgraded, "token_vault_privileged_access"),
		"a null block reads as declared, which is exactly why it cannot be the gate")
	assert.Empty(t,
		stateCredentialIDs(upgraded, "token_vault_privileged_access"),
		"upgraded state names no token vault credential, so the destroy must not touch the field")
}

func TestExpandTokenVaultPrivilegedAccess_NullRawConfigIsNil(t *testing.T) {
	// During a destroy there is no configuration at all and the raw config is a
	// null object. GetAttr panics on one, so the nil check has to come first.
	implied := NewCredentialsResource().CoreConfigSchema().ImpliedType()

	assert.Nil(t, expandTokenVaultPrivilegedAccess(cty.NullVal(implied)))
	assert.Nil(t, expandTokenVaultPrivilegedAccess(cty.UnknownVal(implied)))
}

func TestValidateIPAddressOrCIDR(t *testing.T) {
	for _, allowed := range []string{"10.0.0.1", "192.168.1.0/24", "::1", "2001:db8::/32"} {
		_, errs := validateIPAddressOrCIDR(allowed, "ip_allowlist")
		assert.Empty(t, errs, "%q is a valid address or CIDR range", allowed)
	}

	for _, rejected := range []string{"", "not-an-ip", "10.0.0.256", "10.0.0.0/33"} {
		_, errs := validateIPAddressOrCIDR(rejected, "ip_allowlist")
		assert.NotEmpty(t, errs, "%q is neither an address nor a CIDR range", rejected)
	}
}

// tokenVaultGrantsConfig builds a resource configuration carrying one
// token_vault_privileged_access block whose grants hold the given
// connection-to-scope-count mapping. The connections are listed rather than
// mapped so a repeated one can be expressed.
func tokenVaultGrantsConfig(grants [][2]interface{}) map[string]interface{} {
	grantBlocks := make([]interface{}, 0, len(grants))
	for _, grant := range grants {
		connection, scopeCount := grant[0].(string), grant[1].(int)

		scopes := make([]interface{}, 0, scopeCount)
		for i := range scopeCount {
			scopes = append(scopes, fmt.Sprintf("scope-%d", i))
		}

		grantBlocks = append(grantBlocks, map[string]interface{}{
			"connection": connection,
			"scopes":     scopes,
		})
	}

	return map[string]interface{}{
		"client_id":             "test-client-id",
		"authentication_method": "client_secret_post",
		"token_vault_privileged_access": []interface{}{
			map[string]interface{}{
				"credentials": []interface{}{
					map[string]interface{}{"credential_type": "public_key", "pem": "pem-1"},
				},
				"ip_allowlist": []interface{}{"10.0.0.1"},
				"grants":       grantBlocks,
			},
		},
	}
}

// TestValidateTokenVaultPrivilegedAccess covers the two grant constraints that
// span more than one element, which MaxItems cannot express. The API answers both
// with a 400, so catching them in the diff turns a failed apply into a plan error.
func TestValidateTokenVaultPrivilegedAccess(t *testing.T) {
	testCases := []struct {
		name          string
		grants        [][2]interface{}
		expectedError string
	}{
		{
			name:   "accepts exactly twenty scopes across several grants",
			grants: [][2]interface{}{{"google-oauth2", 10}, {"slack", 6}, {"github", 4}},
		},
		{
			name:          "rejects more than twenty scopes in total",
			grants:        [][2]interface{}{{"google-oauth2", 11}, {"slack", 10}},
			expectedError: "maximum of 20 scopes in total across all grants, but 21 were configured",
		},
		{
			name:          "rejects a repeated connection",
			grants:        [][2]interface{}{{"google-oauth2", 1}, {"google-oauth2", 2}},
			expectedError: `must not repeat a connection, but "google-oauth2" appears more than once`,
		},
		{
			name:   "a single grant well under the cap is fine",
			grants: [][2]interface{}{{"google-oauth2", 1}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewCredentialsResource().Diff(
				context.Background(),
				nil,
				terraform.NewResourceConfigRaw(tokenVaultGrantsConfig(testCase.grants)),
				nil,
			)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		})
	}
}

func TestValidateTokenVaultPrivilegedAccess_NoBlockIsAccepted(t *testing.T) {
	// Every other configuration of this resource omits the block entirely, so the
	// validator has to be a no-op when there is nothing to check.
	_, err := NewCredentialsResource().Diff(
		context.Background(),
		nil,
		terraform.NewResourceConfigRaw(map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "client_secret_post",
		}),
		nil,
	)

	require.NoError(t, err)
}
