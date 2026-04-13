package signature

import "SignatureService/crypto"

type Service struct {
	signersMap *crypto.SignersMap
	repo       Repository
}

func CreateService(signersMap *crypto.SignersMap, repository Repository) Service {
	return Service{
		signersMap: signersMap,
		repo:       repository,
	}
}
