package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	//find first occurance of "\r\n" and get its starting index
	index := bytes.Index(data, []byte("\r\n"))
	if index == -1 {
		return 0, false, nil
	}
		
	line := data[:index]
	if len(line) == 0 {
		//bytes consumed are the /r/n bytes
		return 2, true, nil
	}

	//find : location and verify presence
	colon := bytes.IndexByte(line, ':')
	if colon == -1 {
		//no colon found and header is invalid
		return 0, false, fmt.Errorf("missing colon")
	}
	//split the line into key/values
	keyBytes := line[:colon]
	valueBytes := line[colon+1:]

	if len(keyBytes) == 0 {
		return 0, false, fmt.Errorf("empty header name")
	}

	//need to check for no spaces before colon in keyBytes
	if keyBytes[len(keyBytes)-1] == ' ' || keyBytes[len(keyBytes)-1] == '\t' {
		return 0, false, fmt.Errorf("space before colon in header name")
	}

	key := strings.TrimSpace(string(keyBytes))
	value := strings.TrimSpace(string(valueBytes))
	h[key] = value

	n = index + 2
	return n, false, nil
}