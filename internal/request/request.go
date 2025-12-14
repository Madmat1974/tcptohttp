package request

import (
	"io"
	"strings"
	"fmt"
	"unicode"
	"bytes"
	"errors"
	"HTTPFTCP/internal/headers"
	"strconv"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers		headers.Headers
	state requestState
	Body 		[]byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
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
		r.state = requestStateParsingHeaders
		return n, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil	
	
	case requestStateParsingBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone //no Content-Length header found
			return len(data), nil
		}
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %v", err)
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLen {
			return 0, fmt.Errorf("body longer than Content-Length")
		}
		if len(r.Body) == contentLen {
			r.state = requestStateDone
		}
		return len(data), nil

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
		Headers: headers.NewHeaders(),
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