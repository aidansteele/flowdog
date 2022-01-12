package mytls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

var (
	intermediateCert   *x509.Certificate
	intermediateSigner crypto.Signer
)

func SetIntermediateCA(commonName string, expiry time.Time, caCertificate *x509.Certificate, signer crypto.Signer) error {
	limit := (&big.Int{}).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return errors.WithStack(err)
	}

	tmpl := x509.Certificate{
		Subject:               pkix.Name{CommonName: commonName},
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              expiry,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	intermediateSigner, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.WithStack(err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, caCertificate, intermediateSigner.Public(), signer)
	if err != nil {
		return errors.WithStack(err)
	}

	intermediateCert, err = x509.ParseCertificate(derBytes)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
