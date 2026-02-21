package headers

import (
	"fmt"
	"strings"

	"github.com/jonvanw/httpfromtcp/internal"
)

type Headers map[string]string

func (h Headers) Get(key string) (string, bool) {
	value, ok := h[strings.ToLower(key)]
	return value, ok
}

func (h Headers) Append(key, value string) {
	originalValue, ok := h[strings.ToLower(key)]
	if ok {
		h[strings.ToLower(key)] = originalValue + ", " + value
	} else {
		h[strings.ToLower(key)] = value
	}
}

func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Remove(key string) { 
	if _, ok := h[strings.ToLower(key)]; !ok {
		return
	}
	delete(h, strings.ToLower(key))
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	text := string(data)
	lineEndIndex := strings.Index(text, internal.CRLF)
	if lineEndIndex == -1 {
		return 0, false, nil
	}

	if lineEndIndex == 0 {
		return 2, true, nil
	}

	line := text[:lineEndIndex]
	parts := strings.SplitN(line, ":", 2)
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key != parts[0] {
		return 0, false, fmt.Errorf("invalid header line: key cannot contain spaces: '%s'", parts[0])
	}
	// ensure key contains only allowed characters per RFC7230 token:
	// alphanumeric and !#$%&'*+-.^_`|~
	if !isValidKey(key) {
		return 0, false, fmt.Errorf("invalid header line: key contains invalid characters: '%s'", parts[0])
	}
	
	key = strings.ToLower(key)
	if existingValue, ok := h[key]; ok {
		value = existingValue + ", " + value
	}

	h[key] = value
	
	return lineEndIndex + 2, false, nil
}

// validKeyChars reports whether r is a character permitted in header field names.
func validKeyChars(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	switch r {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// isValidKey checks that the entire key string only consists of allowed characters.
func isValidKey(k string) bool {
	if k == "" {
		return false
	}
	for _, r := range k {
		if !validKeyChars(r) {
			return false
		}
	}
	return true
}

func NewHeaders() Headers {
	return Headers{}
}