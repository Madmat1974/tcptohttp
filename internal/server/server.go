package server

import (
	"net"
	"fmt"
	"sync/atomic"
	"log"
	"HTTPFTCP/internal/response"
	"io"
	"HTTPFTCP/internal/request"
	"bytes"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode 	response.StatusCode
	Message		string
}



type Server struct {
	handler		Handler
	listener	net.Listener
	closed		atomic.Bool
}


func WriteHandlerError(w io.Writer, he *HandlerError) {
    messageBytes := []byte(he.Message)

    response.WriteStatusLine(w, he.StatusCode)

    headers := response.GetDefaultHeaders(len(messageBytes))
    response.WriteHeaders(w, headers)

    w.Write(messageBytes)
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port) //convert str to int
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("TCP server listening on %s\n", addr)
	s := &Server{
		handler: handler,
		listener: ln,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
		log.Println("error accepting connection:", err)
		continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		he := &HandlerError{
			StatusCode: response.BadRequest,
			Message:	err.Error(),
		}
		WriteHandlerError(conn, he)
		return
	}
	buf := &bytes.Buffer{}
	hErr := s.handler(buf, req)
		if hErr != nil {
    	WriteHandlerError(conn, hErr)
    	return
	}

	body := buf.Bytes()
	response.WriteStatusLine(conn, response.Success)
	headers := response.GetDefaultHeaders(len(body))
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}