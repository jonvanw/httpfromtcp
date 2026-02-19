package response

import (
	"fmt"
	"io"

	"github.com/jonvanw/httpfromtcp/internal/headers"
)

type WriterStatus int

const (
	WriterError WriterStatus = -1
	WriterInitialized = iota
	WroteStatusLine
	WroterHeaders
	WritingBody
	WroteBody
	WroteTrailers
)

type Writer struct {
	status WriterStatus
	IOWriter io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		status: WriterInitialized,
		IOWriter: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.status != WriterInitialized {
		return fmt.Errorf("already wrote status line, current status: %d", w.status)
	}
	err := WriteStatusLine(w.IOWriter, statusCode)
	if err != nil {
		w.status = WriterError
		return err
	}
	w.status = WroteStatusLine
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error { 
	if w.status != WroteStatusLine {
		return fmt.Errorf("cannot write headers before writing status line, current status: %d", w.status)
	}
	err := WriteHeaders(w.IOWriter, headers)
	if err != nil {
		w.status = WriterError
		return err
	}
	w.status = WroterHeaders
	return nil
}

func (w *Writer) WriteBody(p []byte) (n int, err error) {
	if w.status != WroterHeaders {
		return 0, fmt.Errorf("cannot write body before writing headers, current status: %d", w.status)
	}
	n, err = w.IOWriter.Write(p)
	if err != nil {
		w.status = WriterError
		return n, err
	}
	w.status = WroteBody
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.status != WroterHeaders && w.status != WritingBody {
		return 0, fmt.Errorf("can only call WriteChunkedBody() after calling WriteHeaders() or WriteChunkedBody(), current status: %d", w.status)
	}
	total := 0
	n, err := w.IOWriter.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		w.status = WriterError
		return n, err
	}
	total += n
	n, err = w.IOWriter.Write(p)
	if err != nil {
		w.status = WriterError
		return n, err
	}
	total += n
	w.IOWriter.Write([]byte("\r\n"))
	w.status = WritingBody
	return total, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.status != WroterHeaders && w.status != WritingBody {
		return 0, fmt.Errorf("can only call WriteChunkedBodyDone() after calling WriteChunkedBody() (or after WriteHeaders() for empty chunked body), current status: %d", w.status)
	}
	n, err := w.IOWriter.Write([]byte("0\r\n"))
	if err != nil {
		w.status = WriterError
		return n, err
	}
	w.status = WroteBody
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.status != WroteBody {
		return fmt.Errorf("can only call WriteTrailers() after body was written, current status: %d", w.status)
	}
	err := WriteHeaders(w.IOWriter, h)
	if err != nil {
		w.status = WriterError
		return fmt.Errorf("error writing trailers: %v", err)
	}
	w.status = WroteTrailers
	return nil
}