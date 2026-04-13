package rest

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Rest interface {
	POST(group *gin.RouterGroup, path string, h Handler)
	GET(group *gin.RouterGroup, path string, h Handler)
}

type Handler func(ctx *gin.Context) (any, error)

type rest struct {
	error APIError
}

func Create() Rest {
	return &rest{
		error: defaultError{},
	}
}

func (r *rest) POST(group *gin.RouterGroup, path string, h Handler) {
	group.POST(path, r.loggingHandler(h))
}

func (r *rest) GET(group *gin.RouterGroup, path string, h Handler) {
	group.GET(path, r.loggingHandler(h))
}

func (r *rest) loggingHandler(h Handler) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				err := fmt.Errorf("%v", e)
				r.sendError(ctx, err)
			}
		}()

		result, err := h(ctx)
		if err != nil {
			r.sendError(ctx, err)
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func (r *rest) sendError(c *gin.Context, err error) {
	code := c.Writer.Status()

	if code < http.StatusBadRequest {
		code = http.StatusBadRequest
	}

	log.Println(err.Error())

	c.JSON(code, r.error.GetError(err))
}
