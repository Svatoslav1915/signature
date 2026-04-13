package signature_test

import (
	"SignatureService/crypto"
	"SignatureService/db/repository"
	"SignatureService/domain/signature"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() signature.Service {
	return signature.CreateService(crypto.NewSignerFactory(), repository.NewService())
}

// Валидация корректности алгоритма на уровне доменной модели
func Test_CreateSignatureDeviceRequestValidateSupportsOnlyRSAAndECC(t *testing.T) {
	validAlgorithms := []crypto.Algorithm{crypto.AlgorithmRSA, crypto.AlgorithmECC}
	for _, algorithm := range validAlgorithms {
		req := signature.CreateSignatureDeviceRequest{
			DeviceID:  uuid.New(),
			Algorithm: algorithm,
		}
		assert.NoError(t, req.Validate())
	}

	invalidReq := signature.CreateSignatureDeviceRequest{
		DeviceID:  uuid.New(),
		Algorithm: crypto.Algorithm("DSA"),
	}
	assert.Error(t, invalidReq.Validate())
}

// Каунтер монотонно увеличивается, подписи расположены в корректном порядке
func Test_SignTransactionConcurrentCounterStrictlyMonotonicWithoutGaps(t *testing.T) {
	svc := newTestService()

	//Создаём устройство подписи
	deviceID := uuid.New()
	_, err := svc.CreateSignatureDevice(signature.CreateSignatureDeviceRequest{
		DeviceID:  deviceID,
		Algorithm: crypto.AlgorithmRSA,
	})
	require.NoError(t, err)

	const signaturesToCreate = 64
	var wg sync.WaitGroup
	errorsCh := make(chan error, signaturesToCreate)

	//Нагружаем устройство конкурентным подписанием
	for i := range signaturesToCreate {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.SignTransaction(signature.SignTransactionRequest{
				DeviceID: deviceID,
				Data:     fmt.Sprintf("payload-%d", i),
			})
			errorsCh <- err
		}(i)
	}

	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		require.NoError(t, err)
	}

	transactionsResponse, err := svc.GetSignedTransactions()
	require.NoError(t, err)

	deviceTransactions := make([]signature.SignedTransaction, 0, signaturesToCreate)
	for _, tx := range transactionsResponse.SignedTransactions {
		if tx.DeviceID == deviceID {
			deviceTransactions = append(deviceTransactions, tx)
		}
	}

	//Все подписи созданы
	require.Len(t, deviceTransactions, signaturesToCreate)

	//Каунтер всех подписей монотонно увеличивается, тело подписанной строки содержит предыдущую подпись или изначальную
	expectedInitialLastSignature := base64.StdEncoding.EncodeToString([]byte(deviceID.String()))
	for i, tx := range deviceTransactions {
		assert.Equal(t, int64(i), tx.SecuredData.SignatureCounter)

		if i == 0 {
			assert.Equal(t, expectedInitialLastSignature, string(tx.SecuredData.LastSignature))
			continue
		}

		assert.Equal(t, deviceTransactions[i-1].Signature, tx.SecuredData.LastSignature)
	}
}

func Test_SignTransactionSecuredDataFormat(t *testing.T) {
	svc := newTestService()

	//Создаём устройство
	deviceID := uuid.New()
	_, err := svc.CreateSignatureDevice(signature.CreateSignatureDeviceRequest{
		DeviceID:  deviceID,
		Algorithm: crypto.AlgorithmRSA,
	})
	require.NoError(t, err)

	//Подписываем транзакцию
	resp, err := svc.SignTransaction(signature.SignTransactionRequest{
		DeviceID: deviceID,
		Data:     "arbitrary_payload_with_underscores",
	})
	require.NoError(t, err)

	//Проверяем формат расширенной подписываемой data
	signedData := resp.SecuredData.String()
	firstUnderscore := strings.IndexByte(signedData, '_')
	lastUnderscore := strings.LastIndexByte(signedData, '_')
	require.True(t, firstUnderscore != -1 && lastUnderscore != -1 && lastUnderscore > firstUnderscore)

	counterPart := signedData[:firstUnderscore]
	dataPart := signedData[firstUnderscore+1 : lastUnderscore]
	lastSignaturePart := signedData[lastUnderscore+1:]

	assert.Equal(t, "0", counterPart)
	assert.Equal(t, "arbitrary_payload_with_underscores", dataPart)

	expectedInitialLastSignature := base64.StdEncoding.EncodeToString([]byte(deviceID.String()))
	assert.Equal(t, expectedInitialLastSignature, lastSignaturePart)
}
