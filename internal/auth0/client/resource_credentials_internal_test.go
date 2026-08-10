package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
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
