package main

import (
	"fmt"
	"net"
	"HTTPFTCP/internal/request"
)


func main() {
	address := ":42069" //port address to listen on
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Error setting up listener: %v\n", err)
	}
	defer listener.Close() //close the listener when program exits

	for {
		//infinite loop/setup to Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting a connection: %v", err)
		}
		fmt.Println("Connection has been accepted")
		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Error with request: %v\n", err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
		fmt.Printf("Connection %v has been closed", conn)
	}	
}

