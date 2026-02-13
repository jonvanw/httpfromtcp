package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	filename := "messages.txt"
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error loading %s: %v", filename, err)
	}
	
	linesChan := getLinesChannel(f)
	for line := range linesChan {
		fmt.Printf("read: %s\n", line)
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