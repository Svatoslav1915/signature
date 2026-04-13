package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"errors"
)

type ECCSigner struct{}

func (e *ECCSigner) Algorithm() Algorithm {
	return AlgorithmECC
}

func (e *ECCSigner) Sign(privateKey, data []byte) ([]byte, error) {
	priv, err := x509.ParseECPrivateKey(privateKey)
	if err != nil {
		return nil, errors.New("private key is not ECDSA")
	}

	hash := sha256.Sum256(data)

	signature, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func (e *ECCSigner) GenerateKeyPair() ([]byte, []byte, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privBytes, pubBytes, nil
}
