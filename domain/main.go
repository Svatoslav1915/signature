package domain

import (
	"SignatureService/crypto"
	"SignatureService/db/repository"
	"SignatureService/domain/signature"
)

var (
	Signature signature.Service
)

func Init() {
	signersMap := crypto.NewSignerFactory()
	repo := repository.NewService()

	Signature = signature.CreateService(signersMap, repo)
}
