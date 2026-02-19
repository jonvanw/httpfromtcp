package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonvanw/httpfromtcp/internal/request"
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

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message: "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message: "Woopsie, my bad\n",
		}
	}

	_, err := io.WriteString(w, "All good, frfr\n")
	if err != nil {
		return &server.HandlerError{
			StatusCode: 500,
			Message: "Error writing response: " + err.Error(),
		}
	}

	return nil
}