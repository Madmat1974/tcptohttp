package request

import (
	"io"
	"strings"
	"fmt"
	"unicode"
	"bytes"
	"errors"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil //need more data
		}
		r.RequestLine = *rl
		r.state = requestStateDone
		return n, nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	//buffer read into index
	bufferSize := 8
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
	}
	
	//main loop and filling the buffer
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		//reading from the io.reader
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("error: incomplete or malformed request")
				} else {
				req.state = requestStateDone
				break
				}
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLineString(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}
	method := parts[0]
	for _, r := range method {
		//checking for all capatilized letters
		if !unicode.IsUpper(r) {
			return nil, fmt.Errorf("method must contain only uppercase letters")
		}
	}
	target := parts[1]
	versionPart := parts[2]
	//split versionPart into two parts and check the second part for 1.1
	versionSplit := strings.Split(versionPart, "/")
	if len(versionSplit) != 2 {
		return nil, fmt.Errorf("invalid HTTP version format")
	}
	if versionSplit[0] != "HTTP" {
		return nil, fmt.Errorf("invalid HTTP version prefix")
	}
	if versionSplit[1] != "1.1" {
		return nil, fmt.Errorf("version is not 1.1")
	}
	
	return &RequestLine{
		Method:			method,
		RequestTarget:	target,
		HttpVersion:	versionSplit[1],
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	rnSequence := []byte("\r\n")
	index := bytes.Index(data, rnSequence)
	if index == -1 {
		return nil, 0, nil
	}

	bytesBeforeRN := data[:index]
	line := string(bytesBeforeRN)

	rl, err := parseRequestLineString(line)
	if err != nil {
		return nil, 0, err
	}
	return rl, index + len(rnSequence), nil
}