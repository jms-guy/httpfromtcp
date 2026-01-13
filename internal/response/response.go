package response

import (
	"fmt"
	"io"

	"github.com/jms-guy/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Code200 StatusCode = iota
	Code400
	Code500
)

type Writer struct {
	ResponseWriter io.Writer
	Status         StatusCode
	Headers        headers.Headers
	Body           []byte
}

func (w *Writer) WriteStatusLine() error {
	statusLine := ""
	switch w.Status {
	case Code200:
		statusLine = "HTTP/1.1 200 OK\r\n"
	case Code400:
		statusLine = "HTTP/1.1 400 Bad Request\r\n"
	case Code500:
		statusLine = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
	}

	_, err := w.ResponseWriter.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, val := range headers {
		_, err := w.ResponseWriter.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, val)))
		if err != nil {
			return fmt.Errorf("error writing header: %s %s: %s", key, val, err)
		}
	}
	_, err := w.ResponseWriter.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing final CLRF: %s", err)
	}

	return nil
}

func (w *Writer) WriteTrailers(headers headers.Headers) error {
	for key, val := range headers {
		_, err := w.ResponseWriter.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, val)))
		if err != nil {
			return fmt.Errorf("error writing trailer: %s %s: %s", key, val, err)
		}
	}
	_, err := w.ResponseWriter.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing final CLRF: %s", err)
	}

	return nil
}

func (w *Writer) WriteBody() (int, error) {
	numBytes, err := w.ResponseWriter.Write(w.Body)
	if err != nil {
		return numBytes, err
	}
	return numBytes, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	chunkHex := fmt.Sprintf("%X\r\n", len(p))

	numBytesHex, err := w.ResponseWriter.Write([]byte(chunkHex))
	if err != nil {
		return 0, fmt.Errorf("error writing byte chunk to response writer")
	}
	numBytesMain, err := w.ResponseWriter.Write(p)
	if err != nil {
		return 0, fmt.Errorf("error writing byte chunk to response writer")
	}
	numBytesCLRF, err := w.ResponseWriter.Write([]byte("\r\n"))
	if err != nil {
		return 0, fmt.Errorf("error writing byte chunk to response writer")
	}
	w.Body = append(w.Body, p...)
	return numBytesHex + numBytesMain + numBytesCLRF, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	numBytes, err := w.ResponseWriter.Write([]byte("0\r\n"))
	if err != nil {
		return 0, fmt.Errorf("error writing final 0 chunk to response")
	}
	return numBytes, nil
}
