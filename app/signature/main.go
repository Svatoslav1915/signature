package main

import (
	signature "SignatureService/app/signature/internal/api"
	"SignatureService/conf"
	"SignatureService/domain"
	"SignatureService/rest"

	"github.com/gin-gonic/gin"
)

func main() {
	//Когда возникнет необходимость auth на пользователя - аттачнем мидлвару с r.Use
	r := gin.New()
	rst := rest.Create()

	domain.Init()

	signature.Register(r.Group("/api"), rst)

	err := r.Run(conf.STARTUPURI)
	if err != nil {
		panic(err)
	}
}
