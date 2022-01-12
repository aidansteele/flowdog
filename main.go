package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/aidansteele/flowdog/examples/account_id_emf"
	"github.com/aidansteele/flowdog/examples/cloudfront_functions"
	"github.com/aidansteele/flowdog/examples/geneve_headers"
	"github.com/aidansteele/flowdog/examples/lambda_acceptor"
	"github.com/aidansteele/flowdog/examples/sts_rickroll"
	"github.com/aidansteele/flowdog/gwlb"
	"github.com/aidansteele/flowdog/kmssigner"
	"github.com/aidansteele/flowdog/mytls"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	_ "github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	err = setupTls(kms.New(sess))
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 6081})
	if err != nil {
		panic(err)
	}

	chain := gwlb.Chain{
		&geneve_headers.GeneveHeaders{},
		&account_id_emf.AccountIdEmf{},
		&sts_rickroll.StsRickroll{},
		cloudfront_functions.NewRickroll(),
	}

	acceptor, err := lambda_acceptor.New(
		lambda.New(sess),
		os.Getenv("LAMBDA_ACCEPTOR_ARN"),
		ec2.New(sess),
	)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	server := &gwlb.Server{
		Handler:  gwlb.DefaultHandler(chain),
		Acceptor: acceptor,
	}

	go healthChecks()

	err = server.Serve(context.Background(), conn)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
}

func setupTls(kmsapi kmsiface.KMSAPI) error {
	// TODO: env var or something
	certBlock, _ := pem.Decode([]byte(`
-----BEGIN CERTIFICATE-----
MIIBfjCCASSgAwIBAgIBATAKBggqhkjOPQQDAjAcMRowGAYDVQQDExFBaWRhbiBL
TVMgVExTIENvLjAeFw0yMjAxMTEwMzU5NTBaFw0zMjAxMTEwMzU5NTBaMBwxGjAY
BgNVBAMTEUFpZGFuIEtNUyBUTFMgQ28uMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcD
QgAE7nApN4JApU+uocRZiefgFoHwpeTTHZkcqVUUTfJp67GmmV4WJ2MTXAH2W+MP
2M8d5cHmgcrSAi6vIzLjvRZBuKNXMFUwDgYDVR0PAQH/BAQDAgKEMBMGA1UdJQQM
MAoGCCsGAQUFBwMBMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFIAvKHJ7WiTA
vlpCx+VMI/+lwgfuMAoGCCqGSM49BAMCA0gAMEUCIFwVxZTneQupaLH2Cunk7FdE
nca45vDEVkjEZw7ERb7SAiEA5Sv3PbIBQSGihtWG4SOJ4tm8US29wM81w9Okl0vR
qpw=
-----END CERTIFICATE-----
	`))

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return errors.WithStack(err)
	}

	signer, err := kmssigner.New(kmsapi, os.Getenv("KEY_ID"))
	if err != nil {
		return errors.WithStack(err)
	}

	return mytls.SetIntermediateCA(
		os.Getenv("INTERMEDIATE_CA_NAME"),
		time.Now().AddDate(1, 0, 0),
		cert,
		signer,
	)
}

func healthChecks() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("health checks")
		w.Write([]byte("ok"))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
}
