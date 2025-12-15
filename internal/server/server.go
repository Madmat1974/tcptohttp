package server

import (
	"net"
	"fmt"
	"sync/atomic"
	"log"
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

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hello World!"

	_, err := conn.Write([]byte(response))
	if err != nil {
		log.Println("error writing response:", err)
	}
}