package rest

type errorResponse struct {
	Message string `json:"message"`
}

type defaultError struct{}

type APIError interface {
	GetError(e error) interface{}
}

func (defaultError) GetError(err error) interface{} {
	return errorResponse{
		Message: err.Error(),
	}
}
