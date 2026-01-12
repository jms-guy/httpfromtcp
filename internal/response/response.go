package response

import (
	"fmt"
	"io"
	"strconv"

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

func (w *Writer) WriteHeaders() error {
	defaultHeaders := GetDefaultHeaders(len(w.Body))
	for key, val := range defaultHeaders {
		_, err := w.ResponseWriter.Write([]byte(key + val))
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

func (w *Writer) WriteBody() (int, error) {
	numBytes, err := w.ResponseWriter.Write(w.Body)
	if err != nil {
		return numBytes, err
	}
	return numBytes, nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length: ": fmt.Sprintf("%s\r\n", strconv.Itoa(contentLen)),
		"Connection: ":     "close\r\n",
		"Content-Type: ":   "text/html\r\n",
	}
}
