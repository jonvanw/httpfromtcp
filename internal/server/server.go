package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jonvanw/httpfromtcp/internal/request"
	"github.com/jonvanw/httpfromtcp/internal/response"
)

type Server struct {
	Port 	 int
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Port: port, 
		handler: handler,
		listener: listener,
	}
	go s.listen()
	
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	
	return s.listener.Close()
}

func (s *Server) listen() { 
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
	
	var buf bytes.Buffer 
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message: err.Error(),
		}
		hErr.Write(bw)
		return
	}
	
	handlerError := s.handler(&buf, req)
	if handlerError != nil {
		log.Printf("Handler error: %s", handlerError.Message)
		handlerError.Write(bw)
		return
	}

	err = response.WriteStatusLine(bw, response.StatusOK)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}
	headers := response.GetDefaultHeaders(buf.Len())
	err = response.WriteHeaders(bw, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	} 
	
	_, err = bw.Write(buf.Bytes())
	if err != nil {
		log.Printf("Error writing response body: %v", err)
	}
}