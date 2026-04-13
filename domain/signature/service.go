package signature

import (
	"SignatureService/crypto"
	"SignatureService/domain/model"
	"errors"

	"github.com/google/uuid"
)

type CreateSignatureDeviceRequest struct {
	DeviceID  uuid.UUID
	Algorithm crypto.Algorithm
	Label     *string
}

func (req CreateSignatureDeviceRequest) Validate() error {
	if req.DeviceID.String() == "" {
		return errors.New("DeviceID is not valid")
	}

	if err := req.Algorithm.Validate(); err != nil {
		return err
	}

	return nil
}

type CreateSignatureDeviceResponse struct {
	DeviceID  uuid.UUID
	Algorithm crypto.Algorithm
	Label     *string
}

func (s Service) CreateSignatureDevice(request CreateSignatureDeviceRequest) (CreateSignatureDeviceResponse, error) {
	signer, err := s.signersMap.GetSigner(request.Algorithm)
	if err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	device, err := model.NewDevice(request.DeviceID, signer, request.Label)
	if err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	err = s.repo.WithTransaction(func() error {
		return s.repo.NewDevice(device)
	})
	if err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	return CreateSignatureDeviceResponse{
		DeviceID:  device.ID,
		Algorithm: signer.Algorithm(),
		Label:     request.Label,
	}, err
}

type Device struct {
	ID        uuid.UUID
	Algorithm crypto.Algorithm
	Label     *string
}

type GetSignatureDevicesResponse struct {
	Devices []Device
}

func (s Service) GetSignatureDevices() (GetSignatureDevicesResponse, error) {
	response := GetSignatureDevicesResponse{}

	err := s.repo.WithTransaction(func() error {
		devices, err := s.repo.GetDevices()
		if len(devices) == 0 {
			return errors.New("No devices found")
		}
		if err != nil {
			return err
		}

		response.Devices = devices

		return nil
	})
	if err != nil {
		return GetSignatureDevicesResponse{}, err
	}

	return response, nil
}

type SignTransactionRequest struct {
	DeviceID uuid.UUID
	Data     string
}

func (req SignTransactionRequest) Validate() error {
	if req.DeviceID.String() == "" {
		return errors.New("DeviceID is not valid")
	}

	if req.Data == "" {
		return errors.New("Data can not be empty")
	}

	return nil
}

type SignTransactionResponse struct {
	Signature   model.Encoded64
	SecuredData model.SecuredData
}

func (s Service) SignTransaction(req SignTransactionRequest) (SignTransactionResponse, error) {
	var (
		encodedString model.Encoded64
		securedData   model.SecuredData
	)

	err := s.repo.WithTransaction(func() error {
		device, err := s.repo.GetDeviceByID(req.DeviceID)
		if err != nil {
			return err
		}

		err = device.WithLock(func() error {
			securedData = securedData.
				SetSignatureCounter(device.GetCounter()).
				SetData(req.Data).
				SetLastSignature(device.GetLastSignature())

			encodedString, err = device.Sign(securedData)
			if err != nil {
				return err
			}

			return s.repo.NewSignature(model.SignatureRecord{
				DeviceID:   device.ID,
				Signature:  encodedString,
				SignedData: securedData,
			})
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return SignTransactionResponse{}, err
	}

	return SignTransactionResponse{
		Signature:   encodedString,
		SecuredData: securedData,
	}, nil
}

type SignedTransaction struct {
	DeviceID    uuid.UUID
	Signature   model.Encoded64
	SecuredData model.SecuredData
}

type GetSignedTransactionsResponse struct {
	SignedTransactions []SignedTransaction
}

func (s Service) GetSignedTransactions() (GetSignedTransactionsResponse, error) {
	signatures := make([]model.SignatureRecord, 0)

	err := s.repo.WithTransaction(func() error {
		signatures = s.repo.GetSignatures()

		return nil
	})
	if err != nil {
		return GetSignedTransactionsResponse{}, err
	}

	if len(signatures) == 0 {
		return GetSignedTransactionsResponse{}, errors.New("No signatures found")
	}

	response := GetSignedTransactionsResponse{
		SignedTransactions: make([]SignedTransaction, len(signatures)),
	}
	for i, signature := range signatures {
		response.SignedTransactions[i] = SignedTransaction{
			DeviceID:    signature.DeviceID,
			Signature:   signature.Signature,
			SecuredData: signature.SignedData,
		}
	}

	return response, nil
}
