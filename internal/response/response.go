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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Code200:
		statusLine := "HTTP/1.1 200 OK\r\n"
		_, err := w.Write([]byte(statusLine))
		if err != nil {
			return fmt.Errorf("error writing OK status")
		}
	case Code400:
		statusLine := "HTTP/1.1 400 Bad Request\r\n"
		_, err := w.Write([]byte(statusLine))
		if err != nil {
			return fmt.Errorf("error writing Bad Request status")
		}
	case Code500:
		statusLine := "HTTP/1.1 500 Internal Server Error\r\n"
		_, err := w.Write([]byte(statusLine))
		if err != nil {
			return fmt.Errorf("error writing Internal Server Error status")
		}
	default:
	}
	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, val := range headers {
		_, err := w.Write([]byte(key + val))
		if err != nil {
			return fmt.Errorf("error writing header: %s %s: %s", key, val, err)
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("error writing final CLRF: %s", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length: ": fmt.Sprintf("%s\r\n", strconv.Itoa(contentLen)),
		"Connection: ":     "close\r\n",
		"Content-Type: ":   "text/plain\r\n",
	}
}
