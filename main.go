package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mati-33/beam/internal/absorber"
	"github.com/mati-33/beam/internal/emitter"
	p "github.com/mati-33/beam/internal/protocol"
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

	e, err := emitter.New()
	if err != nil {
		return fmt.Errorf("failed to initialize emiter: %v", err)
	}
	defer e.Close()

	spinner := ui.NewSpinner()
	spinner.Start()

	err = e.AcceptAbsorber()
	if err != nil {
		return fmt.Errorf("failed to accept absorber: %v", err)
	}

	for {
		beamCodeMsg, err := e.Receive()
		if err != nil {
			return fmt.Errorf("failed to get beam code from absorber: %v", err)
		}
		if beamCodeMsg.Type != p.BC {
			return fmt.Errorf("excpected BC message but got: %s", string(beamCodeMsg.Type))
		}

		if !bytes.Equal(beamCodeMsg.Payload, []byte(beamCode)) {
			if err := e.Send(*p.NewNO()); err != nil {
				return fmt.Errorf("failed to reply to absorber: %v", err)
			}
			continue
		}

		if err := e.Send(*p.NewOK()); err != nil {
			return fmt.Errorf("failed to reply to absorber: %v", err)
		}
		break
	}

	if err := e.Send(*p.NewFI([]byte("some file info"))); err != nil {
		return fmt.Errorf("failed to send file info: %v", err)
	}
	oknoMsg, err := e.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive OK/NO msg after sending FI: %v", err)
	}
	if oknoMsg.Type != p.OK {
		return errors.New("absorber rejected file transfer")
	}

	spinner.Stop()
	fmt.Println("\nemitting file to absorber...")

	cpBuff := make([]byte, 8)
	for {
		_, readErr := file.Read(cpBuff)
		fcMsg := p.NewFC(cpBuff)

		if errors.Is(readErr, io.EOF) {
			err := e.Send(*p.NewFC([]byte{}))
			if err != nil {
				return fmt.Errorf("failed to send FC EOF message: %v", err)
			}
			if _, err := e.Receive(); err != nil {
				return fmt.Errorf("failed to receive message: %v", err)
			}
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		if err := e.Send(*fcMsg); err != nil {
			return fmt.Errorf("failed to send FC message: %v", err)
		}

		oknoMsg, err := e.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive OK/NO msg after sending FC: %v", err)
		}
		if oknoMsg.Type != p.OK {
			return errors.New("absorber canceled file transfer")
		}

		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("file emitted!")
	return nil
}

func handleAbsorb() error {
	if len(os.Args) < 3 {
		return errors.New("'absorb' command expects beam code argument")
	}
	beamCode := os.Args[2]

	a, err := absorber.New()
	if err != nil {
		return fmt.Errorf("failed to initialize absorber: %v", err)
	}
	defer a.Close()

	if err := a.Send(*p.NewBC([]byte(beamCode))); err != nil {
		return fmt.Errorf("failed to send BC message: %v", err)
	}

	oknoMsg, err := a.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive OK/NO beam code confirmation: %v", err)
	}
	if oknoMsg.Type != p.OK {
		return errors.New("invalid beam code")
	}

	fiMsg, err := a.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive FI message: %v", err)
	}
	if fiMsg.Type != p.FI {
		return fmt.Errorf("expected FI, got: %v", string(fiMsg.Type))
	}

	fmt.Println()
	fmt.Println("file info:", string(fiMsg.Payload))
	fmt.Println("accept? y/n: ")
	fmt.Println()
	// Todo: confirmation here

	if err := a.Send(*p.NewOK()); err != nil {
		return fmt.Errorf("failed to send OK message: %v", err)
	}

	file, err := os.Create("copied")
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	fmt.Println("absorbing file...")
	for {
		fcMsg, err := a.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive FC message: %v", err)
		}

		if fcMsg.Type != p.FC {
			return fmt.Errorf("expected FC, got: %s", string(fcMsg.Type))
		}
		if err := a.Send(*p.NewOK()); err != nil {
			return fmt.Errorf("failed to send OK message: %v", err)
		}

		if len(fcMsg.Payload) == 0 {
			if err := a.Send(*p.NewOK()); err != nil {
				return fmt.Errorf("failed to send OK message: %v", err)
			}
			break
		}

		_, err = file.Write(fcMsg.Payload)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}

	}

	fmt.Println("file absorbed!")
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
