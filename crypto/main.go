package crypto

import (
	"errors"
	"fmt"
)

type Signer interface {
	GenerateKeyPair() ([]byte, []byte, error)
	Sign(privateKey, data []byte) ([]byte, error)
	Algorithm() Algorithm
}

type Algorithm string

func (a Algorithm) Validate() error {
	if a == AlgorithmECC || a == AlgorithmRSA {
		return nil
	}

	return errors.New("Algorithm is not valid")
}

const (
	AlgorithmECC Algorithm = "ECC"
	AlgorithmRSA Algorithm = "RSA"
)

type SignersMap struct {
	signers map[Algorithm]Signer
}

func NewSignerFactory() *SignersMap {
	return &SignersMap{signers: map[Algorithm]Signer{
		AlgorithmRSA: &RSASigner{},
		AlgorithmECC: &ECCSigner{},
	}}
}

func (sf *SignersMap) GetSigner(algorithm Algorithm) (Signer, error) {
	s, ok := sf.signers[algorithm]
	if !ok {
		return nil, fmt.Errorf("signer for algorithm %s not exist", algorithm)
	}

	return s, nil
}
