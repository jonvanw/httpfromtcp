package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jonvanw/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Failed to close listener: %s", err.Error())
		} else {
			fmt.Println("Listener closed successfully.")
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %s", err.Error())
			continue
		}
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("Failed to parse request: %s", err.Error())
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %+v\n", request.RequestLine.Method)
		fmt.Printf("- Target: %+v\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %+v\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		conn.Close()
	}
}