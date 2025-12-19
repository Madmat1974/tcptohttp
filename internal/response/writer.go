package response

import (
	"io"
	"fmt"
	"HTTPFTCP/internal/headers"
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

type Writer struct {
    writerState writerState
    writer      io.Writer
}

func NewWriter(w io.Writer) *Writer {
    return &Writer{
        writerState: writerStateStatusLine,
        writer:      w,
    }
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
    if w.writerState != writerStateStatusLine {
        return fmt.Errorf("cannot write status line in state %d", w.writerState)
    }
    defer func() { w.writerState = writerStateHeaders }()
    _, err := w.writer.Write(getStatusLine(statusCode))
    return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
    if w.writerState != writerStateHeaders {
        return fmt.Errorf("cannot write headers in state %d", w.writerState)
    }
    defer func() { w.writerState = writerStateBody }()
    for k, v := range h {
        _, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
        if err != nil {
            return err
        }
    }
    _, err := w.writer.Write([]byte("\r\n"))
    return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
    if w.writerState != writerStateBody {
        return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
    }
    return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
    hexSize := fmt.Sprintf("%x", len(p))
    if _, err := w.writer.Write([]byte(hexSize + "\r\n")); err != nil {
        return 0, err
    }
    if _, err := w.writer.Write(p); err != nil {
        return 0, err
    }
    if _, err := w.writer.Write([]byte("\r\n")); err != nil {
        return 0, err
    }

    return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
    if w.writerState != writerStateBody {
        return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
    }
    n, err := w.writer.Write([]byte("0\r\n"))
    if err != nil {
        return n, err
    }
    // keep state as writerStateBody
    return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
    if w.writerState != writerStateBody {
        return fmt.Errorf("cannot write trailers in state %d", w.writerState)
    }

    for k, v := range h {
        _, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
        if err != nil {
            return err
        }
    }

    // final blank line after trailers
    _, err := w.writer.Write([]byte("\r\n"))
    return err
}
