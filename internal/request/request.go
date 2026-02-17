package request

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jonvanw/httpfromtcp/internal"
	"github.com/jonvanw/httpfromtcp/internal/headers"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}
	data := []byte{}
	buf := make([]byte, internal.BUFFSIZE)
	for request.state != requestStateDone {
		bytes, err := reader.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				// if we reach EOF before the request is fully parsed, that's an error
				if request.state != requestStateDone {
					return nil, fmt.Errorf("reader ended before request was fully parsed")
				}
				break
			}	
			return nil, fmt.Errorf("failed to read from reader: %w", err)
		}
		data = append(data, buf[:bytes]...)
		bytes, err = request.parse(data)
		if err != nil {
			return nil, err
		}
		data = data[bytes:]
	}
	
	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytes := 0
	for r.state != requestStateDone{
		n, err := r.parseSingleItem(data[totalBytes:])
		if err != nil {
			return 0, err
		}
		totalBytes += n
		if n == 0 {
			break
		}
	}
	return totalBytes, nil
}

func (r *Request) parseSingleItem(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		var requestLine RequestLine
		var err error
		bytes := 0
		bytes, requestLine, err = parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytes > 0 {
			r.RequestLine = requestLine
			r.state = requestStateParsingHeaders 
		}	
		return bytes, nil
	case requestStateParsingHeaders:
		if r.Headers == nil {
			r.Headers = make(headers.Headers)
		}
		bytes, isDone, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if isDone {
			r.state = requestStateDone
		}
		return bytes, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: attempting to parser request after it is already done")
	default:
		return 0, fmt.Errorf("Invalid request parser state: %v", r.state)
	}
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	text := string(data)
	lineEndIndex := strings.Index(text, internal.CRLF)
	if lineEndIndex == -1 {
		return 0, RequestLine{}, nil
	}
	line := text[:lineEndIndex]
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return 0,RequestLine{}, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(parts))
	}
	method := parts[0]
	if !isAllCaps(method) {
		return 0, RequestLine{}, fmt.Errorf("invalid method: expected all uppercase, got %s", method)
	}

	target := parts[1]

	version := strings.TrimPrefix(parts[2], "HTTP/")
	if version != "1.1" {
		return 0,RequestLine{}, fmt.Errorf("unsupported HTTP version: %s; only HTTP/1.1 is supported", version)
	}

	return lineEndIndex + 2, RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}

func isAllCaps(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}