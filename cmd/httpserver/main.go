package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonvanw/httpfromtcp/internal/request"
	"github.com/jonvanw/httpfromtcp/internal/response"
	"github.com/jonvanw/httpfromtcp/internal/server"
)

const port = 42069

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
	var statusCode response.StatusCode
	var body string
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.StatusBadRequest
		body = BAD_REQUEST_RESPONSE_BODY
	case "/myproblem":
		statusCode = response.StatusInternalServerError
		body = INTERNAL_SERVER_ERROR_RESPONSE_BODY
	default:
		statusCode = response.StatusOK
		body = OK_RESPONSE_BODY
	}
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(len(body))
	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}
	_, err = w.WriteBody([]byte(body))
	if err != nil {
		log.Printf("Error writing body: %v", err)
		return
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

