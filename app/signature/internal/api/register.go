package api

import (
	"SignatureService/rest"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup, rst rest.Rest) {
	registerSignature(r.Group("signature"), rst)
}
