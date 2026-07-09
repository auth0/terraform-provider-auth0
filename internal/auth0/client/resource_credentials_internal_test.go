package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	rotationSteps := planCredentialRotation(diff, startingAttachedCount)

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
	rotationSteps := planCredentialRotation(diff, startingAttachedCount)

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
	}, 0)
	require.Len(t, pureAdditionSteps, 1)
	assert.Equal(t, createAndAttach, pureAdditionSteps[0].kind)

	pureRemovalSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
	}, 1)
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
	}, 2)
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
	}, 1)
	require.Len(t, moreAdditionsThanRemovalsSteps, 3)
	assert.Equal(t, createAndAttach, moreAdditionsThanRemovalsSteps[0].kind)
	assert.Equal(t, detachAndDelete, moreAdditionsThanRemovalsSteps[1].kind)
	assert.Equal(t, createAndAttach, moreAdditionsThanRemovalsSteps[2].kind)
}

func TestPlanCredentialRotation_SkipsRemovalsWithoutID(t *testing.T) {
	rotationSteps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": ""}},
	}, 1)
	assert.Empty(t, rotationSteps)
}

func TestRemoveAttachedCredential(t *testing.T) {
	firstCredentialID, secondCredentialID := "cred-1", "cred-2"
	attachedCredentials := []management.Credential{{ID: &firstCredentialID}, {ID: &secondCredentialID}}

	remainingCredentials := removeAttachedCredential(attachedCredentials, "cred-1")

	require.Len(t, remainingCredentials, 1)
	assert.Equal(t, "cred-2", remainingCredentials[0].GetID())
}
