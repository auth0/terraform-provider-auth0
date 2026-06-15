package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"

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
