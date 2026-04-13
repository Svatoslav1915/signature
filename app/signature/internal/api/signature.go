package api

import (
	"SignatureService/app/signature/internal/service"
	"SignatureService/rest"

	"github.com/gin-gonic/gin"
)

type signatureController struct {
}

func registerSignature(r *gin.RouterGroup, rst rest.Rest) {
	s := signatureController{}

	rst.POST(r, "/CreateSignatureDevice", s.createSignatureDevice)
	rst.GET(r, "/GetSignatureDevices", s.getSignatureDevices)

	rst.POST(r, "/SignTransaction", s.signTransaction)
	rst.GET(r, "/GetTransactions", s.getTransactions)
}

// @Summary Создать устройство подписания транзакции
// @Router /api/signature/CreateSignatureDevice [post]
func (s signatureController) createSignatureDevice(c *gin.Context) (any, error) {
	request := service.CreateSignatureDeviceRequest{}
	if err := c.BindJSON(&request); err != nil {
		return nil, err
	}

	response, err := service.Signature.CreateSignatureDevice(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// @Summary Получить устройства подписаний транзакций
// @Router /api/signature/SignTransaction [get]
func (s signatureController) getSignatureDevices(_ *gin.Context) (any, error) {
	return service.Signature.GetSignatureDevices()
}

// @Summary Подписать транзакцию
// @Router /api/signature/SignTransaction [post]
func (s signatureController) signTransaction(c *gin.Context) (any, error) {
	request := service.SignTransactionRequest{}
	if err := c.BindJSON(&request); err != nil {
		return nil, err
	}

	response, err := service.Signature.SignTransaction(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// @Summary Получить список транзакций
// @Router /api/signature/GetTransactions [post]
func (s signatureController) getTransactions(_ *gin.Context) (any, error) {
	response, err := service.Signature.GetTransactions()
	if err != nil {
		return nil, err
	}

	return response, nil
}
