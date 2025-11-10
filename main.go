package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatalln("failed to start tcp server, error:", err)
	}
	fmt.Println("server started at tcp://localhost:3000...")

	conn, err := l.Accept()
	if err != nil {
		log.Fatalln("failed to accept connection, error:", err)
	}
	fmt.Println("accepted connection...")

	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		fmt.Println("received message:", string(buff[:n]))

		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("client disconnected")
				return
			}
			log.Fatalln("failed to read message, error:", err)
		}
		conn.Write([]byte("thank you for your message\n"))
	}
}
