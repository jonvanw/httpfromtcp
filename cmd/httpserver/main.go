package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jonvanw/httpfromtcp/internal/headers"
	"github.com/jonvanw/httpfromtcp/internal/request"
	"github.com/jonvanw/httpfromtcp/internal/response"
	"github.com/jonvanw/httpfromtcp/internal/server"
)

const port = 42069
const bufferSize = 1024

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	// SIGINT is sent when the user presses Ctrl+C, SIGTERM is sent by the OS when it wants to terminate the process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	switch {
	case target == "/yourproblem":
		handleYourProblem(w)
	case target == "/myproblem":
		handleMyProblem(w)
	case strings.HasPrefix(target, "/httpbin/"):
		handleHttpBin(w, target)
	default:
		handleOK(w)
	}
}

func handleYourProblem(w *response.Writer) {
	writeSimpleResponse(w, response.StatusBadRequest, BAD_REQUEST_RESPONSE_BODY)
}

func handleMyProblem(w *response.Writer) {
	writeSimpleResponse(w, response.StatusInternalServerError, INTERNAL_SERVER_ERROR_RESPONSE_BODY)
}

func handleOK(w *response.Writer) {
	writeSimpleResponse(w, response.StatusOK, OK_RESPONSE_BODY)
}

func writeSimpleResponse(w *response.Writer, statusCode response.StatusCode, body string) {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(response.GetDefaultHeaders(len(body)))
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("Error writing body: %v", err)
	}
}

func handleHttpBin(w *response.Writer, target string) {
	err := w.WriteStatusLine(response.StatusOK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	resHeaders := response.GetDefaultHeaders(0)
	resHeaders.Remove("content-length")
	resHeaders.Append("transfer-encoding", "chunked")
	resHeaders.Append("trailer", "X-Content-SHA256, X-Content-Length")
	err = w.WriteHeaders(resHeaders)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	x := strings.TrimPrefix(target, "/httpbin/")
	httpBinUrl := fmt.Sprintf("https://httpbin.org/%s", x)
	resp, err := http.Get(httpBinUrl)
	if err != nil {
		log.Printf("Error making request to httpbin: %v", err)
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, bufferSize)
	fullBody := []byte{}
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fullBody = append(fullBody, buf[:n]...)
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Printf("Error writing chunked body: %v", err)
				return
			}
		}
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error reading response body: %v", err)
			}
			break
		}
	}
	w.WriteChunkedBodyDone()
	trailers := headers.NewHeaders()
	sh := sha256.Sum256(fullBody)
	trailers.Append("X-Content-SHA256", fmt.Sprintf("%x", sh))
	trailers.Append("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		log.Printf("Error writing trailers: %v", err)
	}
}

const BAD_REQUEST_RESPONSE_BODY = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const INTERNAL_SERVER_ERROR_RESPONSE_BODY = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const OK_RESPONSE_BODY = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

