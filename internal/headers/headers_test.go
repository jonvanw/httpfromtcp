package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoodSingleHeader(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	length := len(data)
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, length - 2, n)
	assert.False(t, done)
}

func TestGoodSingleHeaderWithExtraWhitespace(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Foo:    bar   \r\n\r\n")
	length := len(data)
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "bar", headers["foo"])
	assert.Equal(t, length - 2, n)
	assert.False(t, done)
}

func TestGoodMultipleHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	headers["foo"] = "bar"
	data := []byte("Fiz:    baz  \r\nHot: dog   \r\n\r\n")
	length := len(data)
	total := 0
	n, done, err := headers.Parse(data)
	data = data[n:]
	total += n
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "bar", headers["foo"])
	assert.Equal(t, "baz", headers["fiz"])
	assert.Equal(t, 15, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data)
	data = data[n:]
	total += n
	assert.Equal(t, "dog", headers["hot"])
	assert.Equal(t, 13, n)
	assert.False(t, done)
	n, done, err = headers.Parse(data)
	total += n
	assert.Equal(t, 2, n)
	assert.True(t, done)
	assert.Equal(t, length, total)
}

func TestGoodNoHeadrs(t *testing.T) {
	// Test: No headers
	headers := NewHeaders()
	data := []byte("\r\n")
	length := len(data)
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, length, n)
	assert.True(t, done)
}


func TestBadInvalidSpacingHeader(t *testing.T) {
	// Test: Invalid spacing header
	headers := NewHeaders()
	data := []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestBadInvalidKeyCharacters(t *testing.T) {
	// keys may only contain alphanumeric and the symbols !#$%&'*+-.^_`|~
	headers := NewHeaders()
	data := []byte("Bad?Key: value\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestGAllowedSymbolInKey(t *testing.T) {
	// this key uses a handful of permitted symbols, including hyphen
	headers := NewHeaders()
	data := []byte("X!#$%&'*+-.^_`|~Header: 123\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "123", headers["x!#$%&'*+-.^_`|~header"])
	assert.Equal(t, len(data)-2, n)
	assert.False(t, done)
}

func TestGoodDuplicateHeaderKey(t *testing.T) {
	// Test: Duplicate header keys should result in values being concatenated with a comma, per RFC7230 sec 3.2.2
	headers := NewHeaders()
	data := []byte("Set-Person: lane-loves-go \r\nSet-Person: prime-loves-zig \r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	l := len(data)
	n, done, err := headers.Parse(data)
	total := n
	data = data[n:]
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go", headers["set-person"])
	assert.False(t, done)

	n, done, err = headers.Parse(data)
	total += n
	data = data[n:]
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.False(t, done)

	n, done, err = headers.Parse(data)
	total += n
	data = data[n:]
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
	assert.False(t, done)
	require.NoError(t, err)

	n, done, err = headers.Parse(data)
	total += n
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
	assert.True(t, done)
	assert.Equal(t, l, total)
}
