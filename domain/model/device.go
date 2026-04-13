package model

import (
	"SignatureService/crypto"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Device struct {
	ID     uuid.UUID
	Signer crypto.Signer
	Label  *string

	privateKey []byte
	publicKey  []byte

	counter       int64
	lastSignature Encoded64

	mu sync.Mutex
}

func NewDevice(id uuid.UUID, signer crypto.Signer, label *string) (*Device, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("device id is not valid %s", id)
	}

	privateKey, publicKey, err := signer.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:     id,
		Signer: signer,
		Label:  label,

		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (d *Device) increaseCounter() {
	d.counter++
}

func (d *Device) WithLock(handler func() error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return handler()
}

func (d *Device) GetCounter() int64 {
	return d.counter
}

func (d *Device) GetLastSignature() Encoded64 {
	if d.lastSignature == "" {
		d.lastSignature = Encoded64(base64.StdEncoding.EncodeToString([]byte(d.ID.String())))
	}

	return d.lastSignature
}

func (d *Device) Sign(data SecuredData) (Encoded64, error) {
	if d.Signer == nil {
		return "", fmt.Errorf("device signer is nil %s", d.ID)
	}

	signedBytes, err := d.Signer.Sign(d.privateKey, []byte(data.String()))
	if err != nil {
		return "", err
	}
	defer d.increaseCounter()

	d.lastSignature = Encoded64(base64.StdEncoding.EncodeToString(signedBytes))

	return d.lastSignature, nil
}
