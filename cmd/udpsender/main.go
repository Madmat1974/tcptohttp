package main

import (
	"fmt"
	"bufio"
	"os"
	"log"
	"net"
)

func main() {
	udpAddress, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf("Failed to resolve UDP address: %v\n", err)
		return
	}

	//Dial a connection to remoteaddress
	conn, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		fmt.Printf("Failed to dial UDP connection: %v\n", err)
		return
	}
	//close connection when exiting
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Error reading input: %v", err)
				break
			}

			_, err = conn.Write([]byte(line))
			if err != nil {
				log.Printf("Failed to write UDP connection: %v", err)
				break
			}
	}
}