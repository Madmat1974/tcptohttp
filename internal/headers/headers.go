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

	k1 := strings.TrimSpace(string(keyBytes))
	key := strings.ToLower(k1) //ensure key is all lowercase
	value := strings.TrimSpace(string(valueBytes))

	for i := 0; i < len(key); i++ {
		if !isValidHeaderChar(key[i]) {
			return 0, false, fmt.Errorf("invalid character in key")
		}
	}

	//check if key already exists in the map
	v, ok := h[key]
	if ok {
		h[key] = v + ", " + value
	} else{
		h[key] = value
	}

	n = index + 2
	return n, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	k1 := strings.TrimSpace(key)
	k2 := strings.ToLower(k1)

	val, ok := h[k2]
	return val, ok
}

func isValidHeaderChar(ch byte) bool {
	//check for letters
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	//check for digits
	if ch >= '0' && ch <= '9' {
		return true
	}
	//check for special characters
	switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~' :
			return true
	}
	return false
}