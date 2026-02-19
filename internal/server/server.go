package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jonvanw/httpfromtcp/internal"
	"github.com/jonvanw/httpfromtcp/internal/request"
	"github.com/jonvanw/httpfromtcp/internal/response"
)

type Server struct {
	Port 	 int
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{Port: port, listener: listener}
	go s.listen(handler)
	
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	
	return s.listener.Close()
}

func (s *Server) listen(handler Handler) { 
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

		go s.handle(conn, handler)
	}
}

func (s *Server) handle(conn net.Conn, handler Handler) { 
	defer conn.Close()
	bw := bufio.NewWriter(conn)
	defer bw.Flush()
	
	var buf bytes.Buffer 
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		err = response.WriteStatusLine(bw, response.StatusBadRequest)
		if err != nil {
			log.Printf("Error sending error response: %v", err)
		}
		return
	}
	
	handlerError := handler(&buf, req)
	if handlerError != nil {
		log.Printf("Handler error: %s", handlerError.Message)
		err = response.WriteStatusLine(bw, response.StatusCode(handlerError.StatusCode))
		if err != nil {
			log.Printf("Error sending error response: %v", err)
			return
		}
		_, err = bw.WriteString(internal.CRLF)
		if err != nil {
			log.Printf("Error writing error response body: %v", err)
			return
		}
		_, err = bw.WriteString(handlerError.Message)
		if err != nil {
			log.Printf("Error writing error response body: %v", err)
		}
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