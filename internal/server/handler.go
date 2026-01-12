package server

import (
	"io"

	"github.com/jms-guy/httpfromtcp/internal/request"
	"github.com/jms-guy/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Msg        string
}

func WriteError(w io.Writer, handlerErr HandlerError) {
	code := ""
	switch handlerErr.StatusCode {
	case response.Code400:
		code = "400"
	case response.Code500:
		code = "500"
	default:
	}
	w.Write([]byte(code))
	w.Write([]byte(" "))
	w.Write([]byte(handlerErr.Msg))
}
