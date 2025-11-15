package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
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
}

func handleEmit() {
	if len(os.Args) < 3 {
		exit("'emit' command expects filename argument")
	}

	filename := os.Args[2]
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		exitf("failed to open '%s' file", filename)
	}

	stats, err := file.Stat()
	if err != nil {
		exitf("failed to fetch '%s' file statistics", filename)
	}
	fmt.Printf("Emiting '%s' (%s)\n", filename, formatFileSize(stats.Size()))
	beamCode := generateBeamCode()
	fmt.Println("beam code is:", beamCode)
	fmt.Println()

	l, err := net.Listen("tcp", "localhost:3000")
	defer l.Close()
	if err != nil {
		exit("failed to start tcp server, error:", err)
	}

	spinner := NewSpinner()
	spinner.Start()

	conn, err := l.Accept()
	defer conn.Close()
	if err != nil {
		exit("failed to accept connection, error:", err)
	}

	buff := make([]byte, 8)
	for {
		n, err := conn.Read(buff)
		if n == 0 {
			exit("absorber disconnected")
		}
		if string(buff[:n]) != beamCode {
			conn.Write([]byte("NO"))
			continue
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				exit("absorber disconnected")
			}
			exit("failed to verify beam code, error:", err)
		}
		if string(buff[:n]) == beamCode {
			conn.Write([]byte("OK"))
			break
		}
	}

	spinner.Stop()
	fmt.Println()
	fmt.Println("Emiting to", conn.RemoteAddr())

	cpBuff := make([]byte, 8)
	for {
		n, err := file.Read(cpBuff)
		_, err2 := conn.Write(cpBuff[:n])
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			exit("failed to read file, error:", err)
		}
		if err2 != nil {
			if errors.Is(err, io.EOF) {
				exit("client disconnected while transfering file")
			}
			exit("failed to emit file, error:", err2)
		}
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("file emited!")
}

func handleAbsorb() {
	if len(os.Args) < 3 {
		exit("'absorb' command expects beam code argument")
	}
	beamCode := os.Args[2]
	fmt.Println("handling absorb...")
	fmt.Println("beam code is:", beamCode)
}

func generateBeamCode() string {
	return "secret"
}

func formatFileSize(size int64) string {
	if size < 100 {
		return fmt.Sprintf("%d B", size)
	} else if size < 100_000 {
		return fmt.Sprintf("%.2f KB", float64(size)/1_000.)
	} else if size < 100_000_000 {
		return fmt.Sprintf("%.2f MB", float64(size)/1_000_000.)
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/1_000_000_000.)
	}
}

func exit(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func exitf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
