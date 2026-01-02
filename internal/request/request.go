package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

var bufferSize int = 8

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
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
	request := &Request{ParserState: 0}
	readerEmpty := false
	bytesRead := 0
	var err error

	for {
		bytesRead = 0
		err = nil

		if request.ParserState == 1 {
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
		if readerEmpty && bytesParsed == 0 && request.ParserState == 0 {
			return request, fmt.Errorf("error: incomplete data at EOF")
		}
		copy(buf, buf[bytesParsed:readToIndex])
		readToIndex -= bytesParsed
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == 0 {
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		} else {
			r.RequestLine = requestLine
			r.ParserState = 1
			return numBytes, nil
		}
	} else if r.ParserState == 1 {
		return 0, fmt.Errorf("error:trying to read data in a done state")
	} else {
		return 0, fmt.Errorf("error: unknown state")
	}
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
