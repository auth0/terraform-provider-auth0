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

func TestPlanCredentialRotation_InterleavesSwaps(t *testing.T) {
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

	steps := planCredentialRotation(diff)

	require.Len(t, steps, 4)
	assert.Equal(t, detachAndDelete, steps[0].kind)
	assert.Equal(t, "old-1", steps[0].credentialID)

	assert.Equal(t, createAndAttach, steps[1].kind)
	assert.Equal(t, "new-1", steps[1].newCredential["name"])

	assert.Equal(t, detachAndDelete, steps[2].kind)
	assert.Equal(t, "old-2", steps[2].credentialID)
	
	assert.Equal(t, createAndAttach, steps[3].kind)
	assert.Equal(t, "new-2", steps[3].newCredential["name"])
}

func TestPlanCredentialRotation_NeverExceedsStartingCountForBalancedSwap(t *testing.T) {
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

	steps := planCredentialRotation(diff)

	// Simulate the credential-collection count as the steps execute. A
	// detach+delete frees a slot (-1) before a create fills one (+1), so the
	// collection count must never rise above the starting count of 2 — this
	// is what keeps us within the tenant's (unknown) credential cap.
	const start = 2
	count, peak := start, start
	for _, step := range steps {
		switch step.kind {
		case detachAndDelete:
			count--
		case createAndAttach:
			count++
		}
		if count > peak {
			peak = count
		}
	}
	assert.Equal(t, start, peak, "collection count must never exceed the starting count during a balanced swap")
	assert.Equal(t, start, count, "collection must return to the starting count after a balanced swap")
}

func TestPlanCredentialRotation_HandlesUnevenAndPureChanges(t *testing.T) {
	add := planCredentialRotation(credentialDiff{
		toAdd: []interface{}{map[string]interface{}{"name": "new-1"}},
	})
	require.Len(t, add, 1)
	assert.Equal(t, createAndAttach, add[0].kind)

	remove := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": "old-1"}},
	})
	require.Len(t, remove, 1)
	assert.Equal(t, detachAndDelete, remove[0].kind)

	// More removals than additions: one interleaved pair, then a trailing removal.
	uneven1 := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{
			map[string]interface{}{"id": "old-1"},
			map[string]interface{}{"id": "old-2"},
		},
		toAdd: []interface{}{map[string]interface{}{"name": "new-1"}},
	})
	require.Len(t, uneven1, 3)
	assert.Equal(t, detachAndDelete, uneven1[0].kind)
	assert.Equal(t, createAndAttach, uneven1[1].kind)
	assert.Equal(t, detachAndDelete, uneven1[2].kind)

	// More additions than removals: one interleaved pair, then a trailing addition.
	uneven2 := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{
			map[string]interface{}{"id": "old-1"},
		},
		toAdd: []interface{}{
			map[string]interface{}{"name": "new-1"}, 
			map[string]interface{}{"name": "new-2"},
		},
	})
	require.Len(t, uneven2, 3)
	assert.Equal(t, detachAndDelete, uneven2[0].kind)
	assert.Equal(t, createAndAttach, uneven2[1].kind)
	assert.Equal(t, createAndAttach, uneven2[2].kind)
}

func TestPlanCredentialRotation_SkipsRemovalsWithoutID(t *testing.T) {
	steps := planCredentialRotation(credentialDiff{
		toRemove: []interface{}{map[string]interface{}{"id": ""}},
	})
	assert.Empty(t, steps)
}

func TestRemoveAttachedCredential(t *testing.T) {
	id1, id2 := "cred-1", "cred-2"
	creds := []management.Credential{{ID: &id1}, {ID: &id2}}

	result := removeAttachedCredential(creds, "cred-1")

	require.Len(t, result, 1)
	assert.Equal(t, "cred-2", result[0].GetID())
}
