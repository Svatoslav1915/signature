package in_memory

import (
	"SignatureService/domain/model"
	"SignatureService/domain/signature"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type service struct {
	devicesMap map[uuid.UUID]*model.Device
	signatures []model.SignatureRecord

	mu sync.Mutex
}

func NewService() signature.Repository {
	return &service{
		devicesMap: make(map[uuid.UUID]*model.Device),
		signatures: make([]model.SignatureRecord, 0),
		mu:         sync.Mutex{},
	}
}

func (s *service) NewDevice(device *model.Device) error {
	_, ok := s.devicesMap[device.ID]
	if ok {
		return fmt.Errorf("device %s already exists", device.ID.String())
	}

	s.devicesMap[device.ID] = device

	return nil
}

func (s *service) GetDeviceByID(deviceID uuid.UUID) (*model.Device, error) {
	device, ok := s.devicesMap[deviceID]
	if !ok {
		return nil, fmt.Errorf("device %s not found", deviceID.String())
	}

	return device, nil
}

func (s *service) GetDevices() ([]signature.Device, error) {
	if len(s.devicesMap) == 0 {
		return nil, fmt.Errorf("no devices found")
	}

	devices := make([]signature.Device, 0, len(s.devicesMap))
	for _, device := range s.devicesMap {
		err := device.WithLock(func() error {
			devices = append(devices, signature.Device{
				ID:        device.ID,
				Algorithm: device.Signer.Algorithm(),
				Label:     device.Label,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return devices, nil
}

func (s *service) NewSignature(signature model.SignatureRecord) error {
	if signature.DeviceID == uuid.Nil {
		return fmt.Errorf("deviceID is required")
	}
	if signature.Signature == "" {
		return fmt.Errorf("signature is required")
	}

	s.signatures = append(s.signatures, signature)
	return nil
}

func (s *service) GetSignatures() []model.SignatureRecord {
	signLen := len(s.signatures)

	result := make([]model.SignatureRecord, signLen)
	copy(result, s.signatures[:signLen])

	return result
}

func (s *service) WithTransaction(handler func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return handler()
}
