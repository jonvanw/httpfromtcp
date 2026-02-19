package server

import (
	"io"
	"log"

	"github.com/jonvanw/httpfromtcp/internal"
	"github.com/jonvanw/httpfromtcp/internal/request"
	"github.com/jonvanw/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (e *HandlerError) Write(w io.Writer) {
	err := response.WriteStatusLine(w, response.StatusCode(e.StatusCode))
	if err != nil {
		// If we can't write the status line, there's not much we can do, so we just log the error and return
		log.Printf("Error sending error response: %v", err)
		return
	}
	_, err = io.WriteString(w, internal.CRLF)
	if err != nil {
		log.Printf("Error writing error response body: %v", err)
		return
	}
	_, err = io.WriteString(w, e.Message)
	if err != nil {
		log.Printf("Error writing error response body: %v", err)
	}
} 