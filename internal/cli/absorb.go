package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mati-33/beam/internal/absorber"
	"github.com/mati-33/beam/internal/bc"
	p "github.com/mati-33/beam/internal/protocol"
	"github.com/mati-33/beam/internal/ui"
)

func Absorb() error {
	if len(os.Args) < 3 {
		return errors.New("'absorb' command expects beam code argument")
	}

	ipBeamCode := os.Args[2]
	if !bc.IsBeamCodeValid(ipBeamCode) {
		return errors.New("invalid beam code")
	}

	beamCode := ipBeamCode[len(ipBeamCode)-2:]
	ipCode := ipBeamCode[:len(ipBeamCode)-2]

	address, err := bc.AbsorberAddress(ipCode)
	if err != nil {
		return fmt.Errorf("failed to determine emitter IPv4 address: %v", err)
	}

	beamCodeBytes, err := p.EncodeBCPayload(beamCode)
	if err != nil {
		return fmt.Errorf("failed to encode beam code: %v", err)
	}

	a, err := absorber.New(address)
	if err != nil {
		return fmt.Errorf("failed to initialize absorber: %v", err)
	}
	defer func() { _ = a.Close() }()

	if err := a.Send(*p.NewBC(beamCodeBytes)); err != nil {
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

	fmt.Printf("%s - %s\n", fileInfo.Name, ui.FormatSize(fileInfo.Size))

	r := bufio.NewReader(os.Stdin)
	fmt.Print("absorb? (y/n): ")
	input, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %v", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "y" {
		if err := a.Send(*p.NewNO()); err != nil {
			return fmt.Errorf("failed to send NO message: %v", err)
		}
		fmt.Println("absorbtion rejected")
		return nil
	}

	if err := a.Send(*p.NewOK()); err != nil {
		return fmt.Errorf("failed to send OK message: %v", err)
	}

	file, err := os.Create(fileInfo.Name)
	absorbFailed := true
	defer func() {
		_ = file.Close()
		if absorbFailed == true {
			_ = os.Remove(fileInfo.Name)
		}
	}()
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}

	pb := ui.NewProgressBar(fileInfo.Size)

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
			absorbFailed = false
			break
		}

		n, err := file.Write(bytes.Trim(fcMsg.Payload, "\x00"))
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		pb.Update(int64(n))
	}

	fmt.Println("\nfile absorbed!")
	return nil
}
