package main

import (
	"HTTPFTCP/internal/server"
	"HTTPFTCP/internal/request"
	"HTTPFTCP/internal/response"
	"log"
	"os"
	"os/signal"
	"syscall"
	"io"

)

const port = 42069

func main() {
	server, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func myHandler(w io.Writer, req *request.Request) *server.HandlerError {
	path := req.RequestLine.RequestTarget

	if path == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Message: 	"Your problem is not my problem\n",
		}
	} else if path == "/myproblem" {
		return &server.HandlerError{
			StatusCode: response.InternalSrvErr,
			Message:	"Woopsie, my bad\n",
		}
	}

	w.Write([]byte("All good, frfr\n"))
	return nil
}