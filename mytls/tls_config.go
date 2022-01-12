package mytls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"time"
)

var certificateCache map[string]*tls.Certificate
var certificateCacheLock sync.Mutex

func TlsConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			certificateCacheLock.Lock()
			defer certificateCacheLock.Unlock()

			domain := info.ServerName

			if certificateCache == nil {
				certificateCache = map[string]*tls.Certificate{}
			}

			cert := certificateCache[domain]
			if cert != nil {
				return cert, nil
			}

			cert, err := generateTlsCertificate(domain)
			certificateCache[domain] = cert
			return cert, err
		},
	}
}

func generateTlsCertificate(domain string) (*tls.Certificate, error) {
	limit := (&big.Int{}).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, limit)

	tmpl := x509.Certificate{
		Subject:               pkix.Name{CommonName: domain},
		DNSNames:              []string{domain},
		SerialNumber:          serialNumber,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, intermediateCert, key.Public(), intermediateSigner)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &tls.Certificate{
		Certificate: [][]byte{derBytes, intermediateCert.Raw},
		PrivateKey:  key,
	}, nil
}
