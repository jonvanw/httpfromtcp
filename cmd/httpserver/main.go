package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonvanw/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
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