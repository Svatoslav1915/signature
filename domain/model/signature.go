package model

import (
	"fmt"

	"github.com/google/uuid"
)

type SecuredData struct {
	SignatureCounter int64
	Data             string

	LastSignature Encoded64
}

func (sd SecuredData) String() string {
	return fmt.Sprintf("%d_%s_%s", sd.SignatureCounter, sd.Data, sd.LastSignature)
}

func (sd SecuredData) SetSignatureCounter(counter int64) SecuredData {
	sd.SignatureCounter = counter

	return sd
}

func (sd SecuredData) SetData(data string) SecuredData {
	sd.Data = data

	return sd
}

func (sd SecuredData) SetLastSignature(lastSignature Encoded64) SecuredData {
	sd.LastSignature = lastSignature

	return sd
}

type Encoded64 string

type SignatureRecord struct {
	DeviceID   uuid.UUID
	Signature  Encoded64
	SignedData SecuredData
}
