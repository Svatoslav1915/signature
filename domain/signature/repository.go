package signature

import (
	"SignatureService/domain/model"

	"github.com/google/uuid"
)

type Repository interface {
	NewDevice(device *model.Device) error
	GetDeviceByID(deviceID uuid.UUID) (*model.Device, error)
	GetDevices() ([]Device, error)

	NewSignature(signature model.SignatureRecord) error
	GetSignatures() []model.SignatureRecord

	WithTransaction(handler func() error) error
}
