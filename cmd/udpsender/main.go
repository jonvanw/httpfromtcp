package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connecting to %s\n", raddr)
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for { 
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Failed to read from console: %s\n", err.Error())
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("Failed to send message: %s\n", err.Error())
		}
	}
}