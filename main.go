package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		exit("expected 'emit' or 'absorb' commands")
	}

	switch os.Args[1] {

	case "emit":
		handleEmit()
	case "absorb":
		handleAbsorb()
	default:
		exit("unknown command:", os.Args[1])
	}

	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		exit("failed to start tcp server, error:", err)
	}
	fmt.Println("server started at tcp://localhost:3000...")

	conn, err := l.Accept()
	if err != nil {
		exit("failed to accept connection, error:", err)
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
			exit("failed to read message, error:", err)
		}
		conn.Write([]byte("thank you for your message\n"))
	}
}

func handleEmit() {
	if len(os.Args) < 3 {
		exit("'emit' command expects filename argument")
	}
	filename := os.Args[2]
	fmt.Println("handling emit...")
	fmt.Println("filename is:", filename)

	// 1. Check if file exists
	// 2. Print file info
	// 3. Generate beam code
	// 4. Start tcp server
	// 5. Print "on other computer run beam absorb 1234"
	// 6. Wait for connection
	// 7. After opening connection send file info and wait for confirmation
	// 8. Start sending file (with progress bar)
}

func handleAbsorb() {
	if len(os.Args) < 3 {
		exit("'absorb' command expects beam code argument")
	}
	beamCode := os.Args[2]
	fmt.Println("handling absorb...")
	fmt.Println("beam code is:", beamCode)
}

func exit(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}
