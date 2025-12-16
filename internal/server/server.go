package server

import (
	"net"
	"fmt"
	"sync/atomic"
	"log"
	"HTTPFTCP/internal/response"
)

type Server struct {
	listener	net.Listener
	closed		atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port) //convert str to int
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	fmt.Printf("TCP server listening on %s\n", addr)
	s := &Server{
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

	if err := response.WriteStatusLine(conn, response.Success); err != nil {
		log.Println("error writing status line:", err)
		return
	}

	headers := response.GetDefaultHeaders(0)

	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Println("error writing headers:", err)
		return
	}
}