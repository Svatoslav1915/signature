package repository

import (
	"SignatureService/db/internal/repository/in_memory"
	"SignatureService/domain/signature"
)

// Фасад, инкапсулирующий внутреннее устройство
func NewService() signature.Repository {
	return in_memory.NewService()
}
