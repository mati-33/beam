package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/mati-33/beam/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "expected 'emit' or 'absorb' commands\n")
		os.Exit(1)
	}

	var err error

	switch os.Args[1] {
	case "emit":
		err = handleEmit()
	case "absorb":
		err = handleAbsorb()
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nan error occured: %v\n", err)
		os.Exit(1)
	}
}

func handleEmit() error {
	if len(os.Args) < 3 {
		return errors.New("'emit' command expects filename argument")
	}

	filename := os.Args[2]
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("failed to open %s file: %v", filename, err)
	}

	stats, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to fetch %s file information: %v", filename, err)
	}

	fmt.Printf("Emiting '%s' (%s)\n", filename, formatFileSize(stats.Size()))
	beamCode := generateBeamCode()
	fmt.Println("beam code is:", beamCode)
	fmt.Println()

	l, err := net.Listen("tcp", "localhost:3000")
	defer l.Close()
	if err != nil {
		return fmt.Errorf("failed to start tcp server: %v", err)
	}

	spinner := ui.NewSpinner()
	spinner.Start()

	conn, err := l.Accept()
	defer conn.Close()
	if err != nil {
		return fmt.Errorf("failed to accept connection: %v", err)
	}

	buff := make([]byte, 8)
	for {
		n, err := conn.Read(buff)
		if n == 0 {
			return errors.New("absorber diconnected")
		}
		if string(buff[:n]) != beamCode {
			conn.Write([]byte("NO"))
			continue
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errors.New("absorber diconnected")
			}
			return fmt.Errorf("failed to verify beam code: %v", err)
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
		n, readErr := file.Read(cpBuff)
		_, writeErr := conn.Write(cpBuff[:n])
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return fmt.Errorf("failed to read file: %v", readErr)
		}
		if writeErr != nil {
			if errors.Is(writeErr, io.EOF) {
				return errors.New("absorber disconnected during file transfer")
			}
			return fmt.Errorf("failed to emit file: %v", writeErr)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("file emited!")
	return nil
}

func handleAbsorb() error {
	if len(os.Args) < 3 {
		return errors.New("'absorb' command expects beam code argument")
	}
	beamCode := os.Args[2]
	fmt.Println("handling absorb...")
	fmt.Println("beam code is:", beamCode)

	return nil
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
