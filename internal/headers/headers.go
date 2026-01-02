package headers

import (
	"fmt"
	"strings"
	"unicode"
)

var specialTchars = []rune("!#$%&'*+-.^_`|~")

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// No CRLF found
	if !strings.Contains(string(data), "\r\n") {
		return 0, false, nil
	}
	// CRLF at start of data, end of headers
	if string(data[:2]) == "\r\n" {
		return 0, true, nil
	}

	headerLines := strings.Split(string(data), "\r\n")
	header := headerLines[0]

	key, value, yes := strings.Cut(header, ":")
	if !yes {
		return 0, false, fmt.Errorf("error: malformed header")
	}

	if len(key) > len(strings.TrimRightFunc(key, unicode.IsSpace)) {
		return 0, false, fmt.Errorf("error: header key not formatted correctly")
	}

	lowerKey := strings.ToLower(key)
	isInvalid := checkForInvalidKeyChar(lowerKey)
	if isInvalid {
		return 0, false, fmt.Errorf("error: invalid character in header")
	}
	finalKey := strings.TrimSpace(lowerKey)
	finalValue := strings.TrimSpace(value)

	if value, ok := h[finalKey]; ok {
		h[finalKey] = fmt.Sprintf("%s, %s", value, finalValue)
	} else {
		h[finalKey] = finalValue
	}

	return len([]byte(header)) + 2, false, nil
}

func checkForInvalidKeyChar(s string) bool {
	invalid := false
	found := false
	for _, char := range s {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			for _, specialChar := range specialTchars {
				if char == specialChar {
					found = true
				}
			}
			if !found {
				invalid = true
			}
		}
	}

	return invalid
}
