package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

	fmt.Printf("Emiting '%s' (%s)\n", filename, ui.FormatSize((stats.Size())))
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

	if err := e.Send(*p.NewFI(filename, stats.Size())); err != nil {
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

	pb := ui.NewProgressBar(stats.Size())
	cpBuff := make([]byte, 8)
	for {
		n, readErr := file.Read(cpBuff)
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

		pb.Update(int64(n))
		time.Sleep(2 * time.Millisecond)
	}

	fmt.Println("\nfile emitted!")
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
	fileInfo := p.DecodeFIPayload(fiMsg.Payload)

	fmt.Println()
	fmt.Printf("%s - %s\n", fileInfo.Name, ui.FormatSize(fileInfo.Size))

	r := bufio.NewReader(os.Stdin)
	fmt.Print("absorb? (y/n): ")
	input, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %v", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "y" {
		fmt.Println("cancelled")
		return nil
	}

	if err := a.Send(*p.NewOK()); err != nil {
		return fmt.Errorf("failed to send OK message: %v", err)
	}

	file, err := os.Create("copied")
	defer file.Close()
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	pb := ui.NewProgressBar(fileInfo.Size)
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

		n, err := file.Write(fcMsg.Payload)
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		pb.Update(int64(n))

	}

	fmt.Println("\nfile absorbed!")
	return nil
}

func generateBeamCode() string {
	return "secret"
}
