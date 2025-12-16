package response

import (
	"fmt"
	"io"
	"strconv"
	"HTTPFTCP/internal/headers"
)

type StatusCode int

const (
	Success 		StatusCode = 200
	BadRequest		StatusCode = 400
	InternalSrvErr	StatusCode = 500
)

func reasonPhraseFromStatus(code StatusCode) string {
	switch code {
	case Success:
		return "OK"
	case BadRequest:
		return "Bad Request"
	case InternalSrvErr:
		return "Internal Server Error"
	default:
		return ""
	}
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := reasonPhraseFromStatus(statusCode)
	line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	_, err := w.Write([]byte(line))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for k, v := range h {
		line := fmt.Sprintf("%s: %s\r\n", k, v)
		if _, err := w.Write([]byte(line)); err != nil {
			return err
		}
	}
	//insert blank line after all headers
	_, err := w.Write([]byte("\r\n"))
	return err
}