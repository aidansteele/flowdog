package kmssigner

import (
	"crypto"
	"crypto/x509"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/pkg/errors"
	"io"
)

var signatureMap = map[crypto.Hash]string{
	crypto.SHA256: kms.SigningAlgorithmSpecEcdsaSha256,
	crypto.SHA384: kms.SigningAlgorithmSpecEcdsaSha384,
	crypto.SHA512: kms.SigningAlgorithmSpecEcdsaSha512,
}

type signer struct {
	api   kmsiface.KMSAPI
	keyId string
	pub   crypto.PublicKey
}

func New(api kmsiface.KMSAPI, keyId string) (crypto.Signer, error) {
	getPub, err := api.GetPublicKey(&kms.GetPublicKeyInput{KeyId: &keyId})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pub, err := x509.ParsePKIXPublicKey(getPub.PublicKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &signer{api: api, keyId: keyId, pub: pub}, nil
}

func (k *signer) Public() crypto.PublicKey {
	return k.pub
}

func (k *signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	sign, err := k.api.Sign(&kms.SignInput{
		KeyId:            &k.keyId,
		Message:          digest,
		MessageType:      aws.String(kms.MessageTypeDigest),
		SigningAlgorithm: aws.String(signatureMap[opts.HashFunc()]),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return sign.Signature, nil
}
