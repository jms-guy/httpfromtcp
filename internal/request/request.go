package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/jms-guy/httpfromtcp/internal/headers"
)

var bufferSize int = 8

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	ParserState requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := &Request{Headers: make(headers.Headers), ParserState: requestStateInitialized}
	readerEmpty := false
	bytesRead := 0
	var err error

	for {
		bytesRead = 0
		err = nil

		if request.ParserState == requestStateDone {
			break
		}
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		if !readerEmpty {
			bytesRead, err = reader.Read(buf[readToIndex:])
			if err != nil {
				if err == io.EOF {
					readerEmpty = true
				} else {
					return request, err
				}
			}
		}
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return request, err
		}
		if readerEmpty && bytesParsed == 0 && request.ParserState == requestStateInitialized {
			return request, fmt.Errorf("error: incomplete data at EOF")
		}
		if readerEmpty && bytesParsed == 0 && request.ParserState == requestStateParsingBody {
			return request, fmt.Errorf("error: request body is shorter than content-length")
		}
		copy(buf, buf[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	totalBytesParsed := 0
	switch r.ParserState {
	case requestStateInitialized:
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		} else {
			r.RequestLine = requestLine
			r.ParserState = requestStateParsingHeaders
			return numBytes, nil
		}
	case requestStateParsingHeaders:
		bytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = requestStateParsingBody
			return bytesParsed, nil
		}
		totalBytesParsed += bytesParsed
	case requestStateParsingBody:
		contentLength := r.Headers.Get("content-length")
		if contentLength == "" && len(data) == 0 {
			r.ParserState = requestStateDone
			return 0, nil
		} else if contentLength == "" && len(data) != 0 {
			r.ParserState = requestStateDone
			return 0, fmt.Errorf("error: no content-length but body present")
		}
		r.Body = append(r.Body, data...)
		contentLengthInt, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, err
		}

		if len(r.Body) > contentLengthInt {
			return 0, fmt.Errorf("error: request body is greater than content-length")
		}
		if len(r.Body) == contentLengthInt {
			r.ParserState = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
	return totalBytesParsed, nil
}

func parseRequestLine(requestBytes []byte) (RequestLine, int, error) {
	if !strings.Contains(string(requestBytes), "\r\n") {
		return RequestLine{}, 0, nil
	}
	lines := strings.Split(string(requestBytes), "\r\n")

	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return RequestLine{}, 0, fmt.Errorf("invalid number of parts in request line")
	}

	method := parts[0]
	target := parts[1]
	fullVersion := parts[2]

	versionParts := strings.Split(fullVersion, "/")
	versionNumber := versionParts[1]

	if versionNumber != "1.1" {
		return RequestLine{}, 0, fmt.Errorf("bad http version")
	}

	for _, r := range method {
		if !unicode.IsUpper(r) {
			return RequestLine{}, 0, fmt.Errorf("method contains bad character")
		}
	}

	return RequestLine{HttpVersion: versionNumber, RequestTarget: target, Method: method}, len([]byte(requestBytes[:len(requestLine)+2])), nil
}
