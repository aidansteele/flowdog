package main

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/aidansteele/flowdog/kmssigner"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"math/big"
	"os"
	"time"
)

func main() {
	keyId := os.Args[1]

	sess, err := session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable, Profile: "ge"})
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	tmpl := x509.Certificate{
		Subject:               pkix.Name{CommonName: "Aidan KMS TLS Co."},
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	signer, err := kmssigner.New(kms.New(sess), keyId)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, signer.Public(), signer)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	fmt.Println(string(pemBytes))
}
