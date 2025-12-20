package chttp

import "errors"

var (
	ErrRequest        = errors.New("the server couldn't handle the current request")
	ErrRequestInfo    = errors.New("the wrong request data was got")
	ErrRequestPath    = errors.New("try to request to unknown resource")
	ErrServerHandling = errors.New("the server couldn't handle the response")
	ErrExternalServer = errors.New("the external server couldn't handle the response")
	ErrRequestMode    = errors.New("the server currently can't handle the request for this mode")
)

// ResponseErr is the wrapper for the errors' response.
type ResponseErr struct {
	Error string `json:"error"`
}
