package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		fmt.Println("a message has been accepted.")
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("end of message")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		buf := make([]byte, 8)
		linesChan := []string{}
		for {
			n, err := f.Read(buf)
			if err != nil {
				if err == io.EOF {
					ch <- strings.Join(linesChan, "")
					break
				}
				log.Printf("Unexpected error reading file: %s", err.Error())
				break
			}

			parts := strings.Split(string(buf[:n]), "\n")
				for i := 0; i < len(parts); i++ {
					if i > 0 {
						ch <- strings.Join(linesChan, "")
						linesChan = []string{}
					}
					linesChan = append(linesChan, parts[i])
				}
		}
	}()

	return ch
}