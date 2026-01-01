package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}

	requestBytes, err := io.ReadAll(reader)
	if err != nil {
		return request, err
	}

	requestLine, err := parseRequestLine(string(requestBytes))
	if err != nil {
		return request, err
	}

	request.RequestLine = requestLine

	return request, nil
}

func parseRequestLine(request string) (RequestLine, error) {
	lines := strings.Split(request, "\r\n")

	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return RequestLine{}, fmt.Errorf("invalid number of parts in request line")
	}

	method := parts[0]
	target := parts[1]
	fullVersion := parts[2]

	versionParts := strings.Split(fullVersion, "/")
	versionNumber := versionParts[1]

	if versionNumber != "1.1" {
		return RequestLine{}, fmt.Errorf("bad http version")
	}

	for _, r := range method {
		if !unicode.IsUpper(r) {
			return RequestLine{}, fmt.Errorf("method contains bad character")
		}
	}

	return RequestLine{HttpVersion: versionNumber, RequestTarget: target, Method: method}, nil
}
