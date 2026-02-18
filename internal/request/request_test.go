package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoodRequestGetLine(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestGoodRequestGetLineWithPath(t *testing.T) {
	// Test: Good GET Request line with path
	r, err := RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestGoodRequestPostLineWithPath(t *testing.T) {
	// Test: Good POST Request line with path
	r, err := RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nContent-Length: 22\r\n\r\n{\"flavor\":\"dark mode\"}"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestGoodRequestWithChunkedReader(t *testing.T) {
	// Test: Good GET Request line
	data := "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	for chunkSize := 1; chunkSize <= len(data); chunkSize++ {
		reader := &chunkReader{
			data:            data,
			numBytesPerRead: chunkSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}
}

func TestGoodRequestWithPathWithChunkedReader(t *testing.T) {
	// Test: Good GET Request line with path
	data := "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	for chunkSize := 1; chunkSize <= len(data); chunkSize++ {
		reader := &chunkReader{
			data:            data,
			numBytesPerRead: chunkSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}
}

func TestBadRequestMissingMethod(t *testing.T) {
	// Test: Invalid number of parts in request line
	_, err := RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestBadRequestOutOfOrderRequestLine(t *testing.T) {
	// Test: Invalid number of parts in request line
	_, err := RequestFromReader(strings.NewReader("/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestBadRequestUnsupportedVersion(t *testing.T) {
	// Test: Unsupported HTTP version
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/3.0\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestBadRequestLowercaseMethod(t *testing.T) {
	// Test: method not all capital letters
	_, err := RequestFromReader(strings.NewReader("get / HTTP/1.1\r\n\r\n"))
	require.Error(t, err)
}

func TestBadRequestHTTP10Version(t *testing.T) {
	// Test: HTTP/1.0 is not supported
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/1.0\r\n\r\n"))
	require.Error(t, err)
}

func TestBadRequestTooManyPartsInRequestLine(t *testing.T) {
	_, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1 EXTRA\r\n\r\n"))
	require.Error(t, err)
}

func TestRequestWithHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])
}

func TestREquestWithoutHeadrs(t *testing.T) {
	// Test: No headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Headers)
}

func TestDuplicateHeaders(t *testing.T) {
	// Test: Duplicate headers should be combined with a comma and space separating values
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: example.com\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, example.com", r.Headers["host"])
}

func TestBadRequestInvalidHeader(t *testing.T) {
	// Test: Malformed Header name
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestMissingEndOfHeaders(t *testing.T) {
	// Test: Missing end of headers (CRLF CRLF)
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestBadRequestInvalidContentLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: abc\r\n" +
			"\r\n" +
			"hello",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestBadRequestBodyLongerThanContentLength(t *testing.T) {
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 3\r\n" +
			"\r\n" +
			"abcDEF", // extra bytes after the declared length
		numBytesPerRead: 2,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestGoodRequestWithStandardBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"Hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "Hello world!\n", string(r.Body))
}

func TestGoodRequestWithExplicitEmptyBody(t *testing.T) {
	// Test: Explicitly empty body with Content-Length: 0
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}

func TestGoodRequestNoContentLengthButWithBodyIgnored(t *testing.T) {
	// Test: Body with no Content-Length header (should be treated as empty body)
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}

func TestGoodRequestWithoutBody(t *testing.T) {
	// Test: No body with no Content-Length header
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}

func TestBadRequestWithTooShortBody(t *testing.T) {
	// Test: Body shorter than reported content length
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestGoodChunkedBody(t *testing.T) {
	// Test: make sure body reading works with chunked reader
	data := "POST /submit HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"Content-Length: 5\r\n" +
		"\r\n" +
		"hello"
	for chunkSize := 1; chunkSize <= len(data); chunkSize++ {
		reader := &chunkReader{
			data:            data,
			numBytesPerRead: chunkSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello", string(r.Body))
	}
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}