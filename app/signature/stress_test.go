package main

import (
	signatureAPI "SignatureService/app/signature/internal/api"
	"SignatureService/domain"
	"SignatureService/rest"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StressConcurrentDeviceCreationAndSigning(t *testing.T) {
	client, baseURL := setupStressHTTP(t)

	// Сначала создаём набор устройств, затем направляем основную нагрузку подписи на малую их часть.
	deviceIDs := createDevices(t, client, baseURL, 20)
	hotIDs := deviceIDs[:5]
	signRequests := 1000

	// Имитируем много клиентов, которые параллельно подписывают через одни и те же hot-устройства.
	runConcurrent(t, signRequests, func(i int) error {
		targetID := hotIDs[i%len(hotIDs)]
		return postJSON(client, baseURL+"/api/signature/SignTransaction", map[string]any{
			"deviceId": targetID,
			"data":     fmt.Sprintf("hot-payload-%d", i),
		})
	})

	// Проверяем, что каждое hot-устройство получило ожидаемое число подписей.
	recordsByDevice := getRecordsByDevice(t)
	expectedPerDevice := signRequests / len(hotIDs)
	for _, id := range hotIDs {
		assert.Len(t, recordsByDevice[id], expectedPerDevice)
	}
}

func Test_StressConcurrentSignAndReadMonotonicity(t *testing.T) {
	client, baseURL := setupStressHTTP(t)

	// Создаём одно hot-устройство, чтобы максимизировать конкуренцию за обновление счётчика и цепочки подписи.
	deviceID := createDevice(t, client, baseURL, "RSA", "")

	// Запускаем параллельно запись (подписи) и чтение (list-endpoints), чтобы нагрузить пересечение read/write.
	runMixedConcurrent(t, 700, 250,
		func(i int) error {
			return postJSON(client, baseURL+"/api/signature/SignTransaction", map[string]any{
				"deviceId": deviceID,
				"data":     fmt.Sprintf("tx-%d", i),
			})
		},
		func() error {
			if err := getJSON(client, baseURL+"/api/signature/GetTransactions", true); err != nil {
				return err
			}
			return getJSON(client, baseURL+"/api/signature/GetSignatureDevices", false)
		},
	)

	// Проверяем строгую монотонность счётчика и корректную цепочку last_signature под конкурентной нагрузкой.
	records := getRecordsByDevice(t)[deviceID]
	require.Len(t, records, 700)

	assertMonotonicChain(t, deviceID, records)
}

func setupStressHTTP(t *testing.T) (*http.Client, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	domain.Init()
	router := gin.New()
	signatureAPI.Register(router.Group("/api"), rest.Create())
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{MaxIdleConns: 100, MaxIdleConnsPerHost: 100, MaxConnsPerHost: 150},
	}
	return client, server.URL
}

func createDevices(t *testing.T, client *http.Client, baseURL string, count int) []string {
	t.Helper()
	deviceIDs := make([]string, 0, count)
	for i := range count {
		algorithm := "ECC"
		if i%2 == 0 {
			algorithm = "RSA"
		}
		deviceIDs = append(deviceIDs, createDevice(t, client, baseURL, algorithm, fmt.Sprintf("device-%d", i)))
	}
	return deviceIDs
}

func createDevice(t *testing.T, client *http.Client, baseURL, algorithm, label string) string {
	t.Helper()
	id := uuid.New().String()

	payload := map[string]any{"id": id, "algorithm": algorithm}
	if label != "" {
		payload["label"] = label
	}

	if err := postJSON(client, baseURL+"/api/signature/CreateSignatureDevice", payload); err != nil {
		require.NoError(t, err)
	}

	return id
}

func runConcurrent(t *testing.T, workers int, fn func(i int) error) {
	t.Helper()
	var wg sync.WaitGroup
	start, errCh := make(chan struct{}), make(chan error, workers)
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			//Блокируем все созданные горутины на этом моменте, чтобы создать пик нагрузки
			<-start
			if err := fn(i); err != nil {
				errCh <- err
			}
		}(i)
	}

	close(start)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func runMixedConcurrent(t *testing.T, writers, readers int, writeFn func(i int) error, readFn func() error) {
	t.Helper()
	runConcurrent(t, writers+readers, func(i int) error {
		if i < writers {
			return writeFn(i)
		}
		return readFn()
	})
}

type signatureRecord struct {
	counter       int64
	lastSignature string
	signature     string
}

func getRecordsByDevice(t *testing.T) map[string][]signatureRecord {
	t.Helper()
	response, err := domain.Signature.GetSignedTransactions()
	require.NoError(t, err)

	records := map[string][]signatureRecord{}
	for _, tx := range response.SignedTransactions {
		id := tx.DeviceID.String()
		records[id] = append(records[id], signatureRecord{
			counter: tx.SecuredData.SignatureCounter, lastSignature: string(tx.SecuredData.LastSignature), signature: string(tx.Signature),
		})
	}

	return records
}

func assertMonotonicChain(t *testing.T, deviceID string, records []signatureRecord) {
	t.Helper()
	for i, record := range records {
		if !assert.Equal(t, int64(i), record.counter) {
			return
		}
		if i == 0 && !assert.Equal(t, encodeBase64(deviceID), record.lastSignature) {
			return
		}
		if i > 0 && !assert.Equal(t, records[i-1].signature, record.lastSignature) {
			return
		}
	}
}

func postJSON(client *http.Client, url string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	response, err := doRequestWithRetry(client, http.MethodPost, url, bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	message, hasMessage := extractErrorMessage(responseBody)
	if hasMessage {
		return fmt.Errorf("api returned error: %s", message)
	}

	return nil
}

func getJSON(client *http.Client, url string, allowNoData bool) error {
	response, err := doRequestWithRetry(client, http.MethodGet, url, nil, "")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	message, hasMessage := extractErrorMessage(responseBody)
	if hasMessage {
		if allowNoData && strings.Contains(message, "No signatures found") {
			return nil
		}
		return fmt.Errorf("api returned error: %s", message)
	}

	return nil
}

func extractErrorMessage(body []byte) (string, bool) {
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", false
	}

	message, hasMessage := parsed["message"].(string)
	return message, hasMessage
}

func doRequestWithRetry(client *http.Client, method, url string, body io.Reader, contentType string) (*http.Response, error) {
	const attempts = 8
	var lastErr error

	var payload []byte
	if body != nil {
		var err error
		payload, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < attempts; i++ {
		var reqBody io.Reader
		if payload != nil {
			reqBody = bytes.NewReader(payload)
		}

		req, err := http.NewRequest(method, url, reqBody)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}

		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isTransientNetworkError(err) {
			return nil, err
		}

		time.Sleep(time.Duration(i+1) * 50 * time.Millisecond)
	}

	return nil, lastErr
}

func isTransientNetworkError(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}

	message := strings.ToLower(opErr.Err.Error())
	return strings.Contains(message, "refused") || strings.Contains(message, "reset") || strings.Contains(message, "timeout")
}

func encodeBase64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}
