package server

import (
	"www.github.com/jaychillin2607/httpfromtcp/internal/request"
	"www.github.com/jaychillin2607/httpfromtcp/internal/response"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Handler func(w *response.Writer, req *request.Request) *HandlerError
