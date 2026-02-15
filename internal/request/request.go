package request

import (
	"fmt"
	"io"
	"strings"
)

type requestState int

const (
	initialized requestState = iota
	done
)

const crlf = "\r\n"
const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}
	data := []byte{}
	buf := make([]byte, bufferSize)
	for {
		bytes, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error reading from reader: %w", err)
			}
		}
		data = append(data, buf[:bytes]...)
		bytes, err = request.parse(data)
		if err != nil {
			return nil, err
		}
		if request.state == done {
			break
		}
		data = data[bytes:]
	}
	
	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	var requestLine RequestLine
	var err error
	bytes := 0
	bytes, requestLine, err = parseRequestLine(data)
	if err != nil {
		return 0, err
	}
	if bytes > 0 {
		r.RequestLine = requestLine
		r.state = done
	}	
	return bytes, nil
}

func parseRequestLine(data []byte) (int, RequestLine, error) {
	text := string(data)
	lineEndIndex := strings.Index(text, crlf)
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