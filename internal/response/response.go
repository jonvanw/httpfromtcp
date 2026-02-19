package response

import (
	"fmt"
	"io"

	"github.com/jonvanw/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, status StatusCode) error { 
	var reasonPhrase string
	switch status {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}
	
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, reasonPhrase)
	_, err := io.WriteString(w, statusLine)
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := io.WriteString(w, headerLine)
		if err != nil {
			return fmt.Errorf("error writing header %s: %v", key, err)
		}
	}
	// Write the blank line to indicate the end of headers
	_, err := io.WriteString(w, "\r\n")
	if err != nil {
		return fmt.Errorf("error closing headers: %v", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-type": "text/plain",
		"content-length": fmt.Sprintf("%d", contentLen),
		"connection": "close",
	}
}