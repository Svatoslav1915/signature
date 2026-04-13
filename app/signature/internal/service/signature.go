package service

import (
	"SignatureService/crypto"
	"SignatureService/domain"
	"SignatureService/domain/signature"

	"github.com/google/uuid"
)

type signatureService struct{}

type CreateSignatureDeviceRequest struct {
	ID        string  `json:"id"`
	Algorithm string  `json:"algorithm"`
	Label     *string `json:"label,omitempty"`
}

type CreateSignatureDeviceResponse struct {
	ID        string  `json:"id"`
	Algorithm string  `json:"algorithm"`
	Label     *string `json:"label,omitempty"`
}

func (signatureService) CreateSignatureDevice(req CreateSignatureDeviceRequest) (CreateSignatureDeviceResponse, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	r := signature.CreateSignatureDeviceRequest{
		DeviceID:  id,
		Algorithm: crypto.Algorithm(req.Algorithm),
		Label:     req.Label,
	}
	if err = r.Validate(); err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	response, err := domain.Signature.CreateSignatureDevice(r)
	if err != nil {
		return CreateSignatureDeviceResponse{}, err
	}

	return CreateSignatureDeviceResponse{
		ID:        response.DeviceID.String(),
		Algorithm: string(response.Algorithm),
		Label:     response.Label,
	}, nil
}

type SignatureDevice struct {
	ID        string  `json:"id"`
	Algorithm string  `json:"algorithm"`
	Label     *string `json:"label,omitempty"`
}

type GetSignatureDevicesResponse struct {
	Devices []SignatureDevice `json:"devices"`
}

func (signatureService) GetSignatureDevices() (GetSignatureDevicesResponse, error) {
	response, err := domain.Signature.GetSignatureDevices()
	if err != nil {
		return GetSignatureDevicesResponse{}, err
	}

	serviceResponse := GetSignatureDevicesResponse{}
	for _, device := range response.Devices {
		serviceResponse.Devices = append(serviceResponse.Devices, SignatureDevice{
			ID:        device.ID.String(),
			Algorithm: string(device.Algorithm),
			Label:     device.Label,
		})
	}

	return serviceResponse, nil
}

type SignTransactionRequest struct {
	DeviceId string `json:"deviceId"`
	Data     string `json:"data"`
}

type SignTransactionResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

func (signatureService) SignTransaction(req SignTransactionRequest) (SignTransactionResponse, error) {
	id, err := uuid.Parse(req.DeviceId)
	if err != nil {
		return SignTransactionResponse{}, err
	}

	r := signature.SignTransactionRequest{
		DeviceID: id,
		Data:     req.Data,
	}
	if err = r.Validate(); err != nil {
		return SignTransactionResponse{}, err
	}

	response, err := domain.Signature.SignTransaction(r)
	if err != nil {
		return SignTransactionResponse{}, err
	}

	serviceResponse := SignTransactionResponse{
		Signature:  string(response.Signature),
		SignedData: response.SecuredData.String(),
	}
	return serviceResponse, nil
}

type DeviceSignedTransactions struct {
	DeviceId           string              `json:"deviceId"`
	SignedTransactions []SignedTransaction `json:"signedTransactions"`
}

type SignedTransaction struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

type GetSignedTransactionsResponse struct {
	DeviceID           string              `json:"deviceId"`
	SignedTransactions []SignedTransaction `json:"signed_transactions"`
}

type GetSignedTransactionResponses []GetSignedTransactionsResponse

func (signatureService) GetTransactions() (GetSignedTransactionResponses, error) {
	r, err := domain.Signature.GetSignedTransactions()
	if err != nil {
		return nil, err
	}

	deviceTransactions := make(map[uuid.UUID][]SignedTransaction)
	for _, signedTransaction := range r.SignedTransactions {
		deviceTransactions[signedTransaction.DeviceID] = append(deviceTransactions[signedTransaction.DeviceID], SignedTransaction{
			Signature:  string(signedTransaction.Signature),
			SignedData: signedTransaction.SecuredData.String(),
		})
	}

	serviceResponse := make(GetSignedTransactionResponses, 0)
	for k, v := range deviceTransactions {
		serviceResponse = append(serviceResponse, GetSignedTransactionsResponse{
			DeviceID:           k.String(),
			SignedTransactions: v,
		})
	}

	return serviceResponse, nil
}
