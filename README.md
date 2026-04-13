# Signature Service API
Сервис выполнен в DDD архитектуре так, как если бы я его интегрировал в существующий проект.

# Запуск сервиса 
go run ./app/signature/main.go

# Примеры обращений к API, выполняемые в Windows среде PowerShell

## CreateSignatureDevice - создать устройство подписи
curl.exe --location "localhost:9191/api/signature/CreateSignatureDevice" --header "Content-Type: application/json" --data '{\"id\": \"bd02ce12-86ab-4ce0-8ec2-0fca88dfd999\",\"algorithm\": \"RSA\",\"label\": \"RSASignatureDevice\"}'


## GetSignatureDevices - получить все устройства
curl.exe --location "localhost:9191/api/signature/GetSignatureDevices"

## SignTransaction - подписать устройством
curl.exe --location 'localhost:9191/api/signature/SignTransaction' --header 'Content-Type: application/json'--data '{\"deviceId\": \"bd02ce12-86ab-4ce0-8ec2-0fca88dfd999\",\"data\":\"DATA_TO_BE_SIGNED\"}'

## GetTransactions - получить все транзакции для каждого устройства
curl.exe --location 'localhost:9191/api/signature/GetTransactions'

# AI tools
1. Скелет тестов оформлен агентами с Cursor, доработаны руками
2. Реализации алгоритмов подписей, создания ключей с GPT5.0 