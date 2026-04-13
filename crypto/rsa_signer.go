package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
)

type RSASigner struct{}

func (r *RSASigner) Algorithm() Algorithm {
	return AlgorithmRSA
}

func (r *RSASigner) Sign(privateKey, data []byte) ([]byte, error) {
	priv, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, errors.New("private key is not RSA")
	}

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (r *RSASigner) GenerateKeyPair() ([]byte, []byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privBytes, pubBytes, nil
}
