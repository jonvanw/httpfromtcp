package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jonvanw/httpfromtcp/internal/response"
)

type Server struct {
	Port 	 int
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{Port: port, listener: listener}
	go s.listen()
	
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	
	return s.listener.Close()
}

func (s *Server) listen()  { 
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			// if the error is not because the server is closed, we log it and continue accepting new connections
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) { 
	defer conn.Close()
	bw := bufio.NewWriter(conn)
	defer bw.Flush()
	
	err := response.WriteStatusLine(bw, response.StatusOK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(bw, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	} 
}