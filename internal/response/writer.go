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
	WroteBody
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