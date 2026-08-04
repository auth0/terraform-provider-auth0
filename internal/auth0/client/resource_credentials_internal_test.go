package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"

	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
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

// simulateAttachedCount replays the rotation steps against a starting attached
// count and returns the lowest and highest attached count seen at any step
// boundary.
func simulateAttachedCount(rotationSteps []rotationStep, startingAttachedCount int) (lowestAttachedCount, highestAttachedCount int) {
	currentAttachedCount := startingAttachedCount
	lowestAttachedCount, highestAttachedCount = startingAttachedCount, startingAttachedCount
	for _, step := range rotationSteps {
		switch step.kind {
		case detachAndDelete:
			currentAttachedCount--
		case createAndAttach:
			currentAttachedCount++
		}
		if currentAttachedCount < lowestAttachedCount {
			lowestAttachedCount = currentAttachedCount
		}
		if currentAttachedCount > highestAttachedCount {
			highestAttachedCount = currentAttachedCount
		}
	}
	return lowestAttachedCount, highestAttachedCount
}

func TestPlanCredentialRotation_AtCapacityRemovesFirst(t *testing.T) {
	// A full 2-for-2 swap starting at the cap: each pair must remove before it
	// adds so the count never overshoots 2.
	diff := credentialDiff{
		toRemove: []interface{}{
			map[string]interface{}{"id": "old-1"},
			map[string]interface{}{"id": "old-2"},
		},
		toAdd: []interface{}{
			map[string]interface{}{"name": "new-1"},
			map[string]interface{}{"name": "new-2"},
		},
	}

	const startingAttachedCount = 2
	rotationSteps := planCredentialRotation(diff, startingAttachedCount, startingAttachedCount)

	require.Len(t, rotationSteps, 4)
	assert.Equal(t, detachAndDelete, rotationSteps[0].kind)
	assert.Equal(t, "old-1", rotationSteps[0].credentialID)
	assert.Equal(t, createAndAttach, rotationSteps[1].kind)
	assert.Equal(t, "new-1", rotationSteps[1].newCredential["name"])
	assert.Equal(t, detachAndDelete, rotationSteps[2].kind)
	assert.Equal(t, "old-2", rotationSteps[2].credentialID)
	assert.Equal(t, createAndAttach, rotationSteps[3].kind)
	assert.Equal(t, "new-2", rotationSteps[3].newCredential["name"])

	lowestAttachedCount, highestAttachedCount := simulateAttachedCount(rotationSteps, startingAttachedCount)
	assert.Equal(t, 2, highestAttachedCount, "count must never exceed the cap during a swap at capacity")
	assert.Equal(t, 1, lowestAttachedCount, "at least one credential stays attached throughout")
}

func TestPlanCredentialRotation_WithHeadroomAddsFirst(t *testing.T) {
	// A 1-for-1 swap on a client holding a single credential. With headroom
	// below the minimum cap the new credential must be added before the old one
	// is removed, so the client is never left with zero attached credentials.
	diff := credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
		toAdd:    []interface{}{map[string]interface{}{"name": "new-1"}},
	}

	const startingAttachedCount = 1
	rotationSteps := planCredentialRotation(diff, startingAttachedCount, startingAttachedCount)

	require.Len(t, rotationSteps, 2)
	assert.Equal(t, createAndAttach, rotationSteps[0].kind)
	assert.Equal(t, "new-1", rotationSteps[0].newCredential["name"])
	assert.Equal(t, detachAndDelete, rotationSteps[1].kind)
	assert.Equal(t, "old-1", rotationSteps[1].credentialID)

	lowestAttachedCount, highestAttachedCount := simulateAttachedCount(rotationSteps, startingAttachedCount)
	assert.GreaterOrEqual(t, lowestAttachedCount, 1, "a valid credential must stay attached throughout a 1-for-1 rotation")
	assert.Equal(t, 2, highestAttachedCount, "adding first briefly holds 2, which every tenant allows")
}

func TestPlanCredentialRotation_HandlesUnevenAndPureChanges(t *testing.T) {
	pureAdditionSteps := planCredentialRotation(credentialDiff{
		toAdd: []interface{}{map[string]interface{}{"name": "new-1"}},
	}, 0, 0)
	require.Len(t, pureAdditionSteps, 1)
	assert.Equal(t, createAndAttach, pureAdditionSteps[0].kind)

	pureRemovalSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
	}, 1, 1)
	require.Len(t, pureRemovalSteps, 1)
	assert.Equal(t, detachAndDelete, pureRemovalSteps[0].kind)

	// More removals than additions, starting at capacity: one interleaved pair
	// (remove-first), then a trailing removal.
	moreRemovalsThanAdditionsSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{
			map[string]interface{}{"id": "old-1"},
			map[string]interface{}{"id": "old-2"},
		},
		toAdd: []interface{}{map[string]interface{}{"name": "new-1"}},
	}, 2, 2)
	require.Len(t, moreRemovalsThanAdditionsSteps, 3)
	assert.Equal(t, detachAndDelete, moreRemovalsThanAdditionsSteps[0].kind)
	assert.Equal(t, createAndAttach, moreRemovalsThanAdditionsSteps[1].kind)
	assert.Equal(t, detachAndDelete, moreRemovalsThanAdditionsSteps[2].kind)

	// More additions than removals, starting with headroom: the pair adds first,
	// then a trailing addition.
	moreAdditionsThanRemovalsSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{
			map[string]interface{}{"id": "old-1"},
		},
		toAdd: []interface{}{
			map[string]interface{}{"name": "new-1"},
			map[string]interface{}{"name": "new-2"},
		},
	}, 1, 1)
	require.Len(t, moreAdditionsThanRemovalsSteps, 3)
	assert.Equal(t, createAndAttach, moreAdditionsThanRemovalsSteps[0].kind)
	assert.Equal(t, detachAndDelete, moreAdditionsThanRemovalsSteps[1].kind)
	assert.Equal(t, createAndAttach, moreAdditionsThanRemovalsSteps[2].kind)
}

func TestPlanCredentialRotation_SkipsRemovalsWithoutID(t *testing.T) {
	rotationSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": ""}},
	}, 1, 1)
	assert.Empty(t, rotationSteps)
}

func TestPlanCredentialRotation_FullPoolWithSlotHeadroomRemovesFirst(t *testing.T) {
	// The slot holds one credential, so it has headroom, but the client's pool is
	// full because other features hold the remaining records. Creating first would
	// fail with "A client can have a maximum of 4 credentials", so the rotation
	// must remove before it adds even though the slot looks like it has room.
	diff := credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
		toAdd:    []interface{}{map[string]interface{}{"name": "new-1"}},
	}

	rotationSteps := planCredentialRotation(diff, 1, maxPoolCredentials)

	require.Len(t, rotationSteps, 2)
	assert.Equal(t, detachAndDelete, rotationSteps[0].kind,
		"a full pool must free a record before a new credential is created")
	assert.Equal(t, "old-1", rotationSteps[0].credentialID)
	assert.Equal(t, createAndAttach, rotationSteps[1].kind)
	assert.Equal(t, "new-1", rotationSteps[1].newCredential["name"])
}

func TestPlanCredentialRotation_SlotHeadroomWithPoolHeadroomAddsFirst(t *testing.T) {
	// Same slot occupancy as above, but the pool has room. Zero-downtime ordering
	// is available here, so the new credential is added before the old is removed.
	diff := credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
		toAdd:    []interface{}{map[string]interface{}{"name": "new-1"}},
	}

	rotationSteps := planCredentialRotation(diff, 1, maxPoolCredentials-1)

	require.Len(t, rotationSteps, 2)
	assert.Equal(t, createAndAttach, rotationSteps[0].kind,
		"pool headroom allows the zero-downtime add-first ordering")
	assert.Equal(t, detachAndDelete, rotationSteps[1].kind)
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

	t.Logf("planned pkjwt = %#v", data.Get("private_key_jwt"))
	owned := ownedCredentialIDs(data)
	t.Logf("owned = %v", owned)
	assert.True(t, owned["cred-mine"],
		"an owned credential must still be deletable after the block leaves the config")
	assert.Len(t, owned, 1)
}

// State records 2 attached, the rotation replaces 1. Snapshot must be both
// (so the surviving one stays attached), and pool headroom must drive ordering.
func TestRotation_TwoAttachedRotatingOneKeepsTheOther(t *testing.T) {
	data := diffData(t,
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(
				map[string]interface{}{"id": "cred-a", "name": "a", "credential_type": "public_key", "pem": "PEM-A", "algorithm": "RS256"},
				map[string]interface{}{"id": "cred-b", "name": "b", "credential_type": "public_key", "pem": "PEM-B", "algorithm": "RS256"},
			),
		},
		map[string]interface{}{
			"client_id":             "test-client-id",
			"authentication_method": "private_key_jwt",
			"private_key_jwt": pkjwtBlock(
				map[string]interface{}{"name": "a", "credential_type": "public_key", "pem": "PEM-A", "algorithm": "RS256"},
				map[string]interface{}{"name": "b", "credential_type": "public_key", "pem": "PEM-B-NEW", "algorithm": "RS256"},
			),
		},
	)

	ids := stateCredentialIDs(data, "private_key_jwt")
	t.Logf("snapshot ids = %v", ids)
	assert.ElementsMatch(t, []string{"cred-a", "cred-b"}, ids)

	toAdd, toRemove := value.Difference(data, "private_key_jwt.0.credentials")
	d := classifyCredentialChanges(toAdd, toRemove)
	t.Logf("classified toAdd=%d toRemove=%d", len(d.toAdd), len(d.toRemove))

	// Slot is FULL (2) -> must remove first regardless of pool.
	steps := planCredentialRotation(d, len(ids), 2)
	require.Len(t, steps, 2)
	t.Logf("step0=%v id=%q  step1=%v", steps[0].kind, steps[0].credentialID, steps[1].kind)
	assert.Equal(t, detachAndDelete, steps[0].kind, "slot at cap must remove first")

	// Simulate the live sequence to prove the slot list never exceeds 2 and never
	// contains a foreign credential.
	attached := attachedCredentialsFromState(data, "private_key_jwt")
	maxSeen := len(attached)
	for _, s := range steps {
		switch s.kind {
		case detachAndDelete:
			attached = removeAttachedCredential(attached, s.credentialID)
		case createAndAttach:
			newID := "cred-new"
			attached = append(attached, management.Credential{ID: &newID})
		}
		if len(attached) > maxSeen {
			maxSeen = len(attached)
		}
		ids := []string{}
		for _, c := range attached {
			ids = append(ids, c.GetID())
		}
		t.Logf("  after step: slot=%v", ids)
	}
	assert.LessOrEqual(t, maxSeen, maxSlotCredentials, "slot list must never exceed the slot cap")

	final := []string{}
	for _, c := range attached {
		final = append(final, c.GetID())
	}
	assert.ElementsMatch(t, []string{"cred-a", "cred-new"}, final,
		"the untouched credential survives and only the rotated one is replaced")
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

	rawAdd, rawRemove := value.Difference(data, "private_key_jwt.0.credentials")
	t.Logf("raw Difference toAdd=%d toRemove=%d (the trap)", len(rawAdd), len(rawRemove))

	d := credentialChanges(data, "private_key_jwt.0.credentials")
	t.Logf("credentialChanges toAdd=%d toRemove=%d", len(d.toAdd), len(d.toRemove))
	steps := planCredentialRotation(d, 1, 2)
	for i, s := range steps {
		t.Logf("  step %d kind=%v pem=%v", i, s.kind, s.newCredential["pem"])
	}
	assert.Empty(t, steps,
		"an unchanged credential set must not create a duplicate credential")
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
