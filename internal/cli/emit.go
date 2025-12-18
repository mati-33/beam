package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mati-33/beam/internal/bc"
	"github.com/mati-33/beam/internal/emitter"
	p "github.com/mati-33/beam/internal/protocol"
	"github.com/mati-33/beam/internal/ui"
)

func Emit() error {
	if len(os.Args) < 3 {
		return errors.New("'emit' command expects filename argument")
	}

	path := os.Args[2]
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", path, err)
	}

	filename := filepath.Base(path)

	stats, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to fetch %s file information: %v", filename, err)
	}

	beamCode := bc.BeamCodeHex()
	hostBitsHex, err := bc.HostBitsHex()
	if err != nil {
		return fmt.Errorf("failed to generate beam code: %v", err)
	}

	fmt.Printf("Emiting '%s' (%s)\n", filename, ui.FormatSize((stats.Size())))
	fmt.Printf("beam code is: %s%s\n", hostBitsHex, beamCode)
	fmt.Println()

	e, err := emitter.New()
	if err != nil {
		return fmt.Errorf("failed to initialize emiter: %v", err)
	}
	defer e.Close()

	spinner := ui.NewSpinner()
	spinner.Start()

outer:
	for {
		err = e.AcceptAbsorber()
		if err != nil {
			return fmt.Errorf("failed to accept absorber: %v", err)
		}

		for {
			beamCodeMsg, err := e.Receive()
			if err != nil {
				e.ClearAbsorber()
				break
			}
			if beamCodeMsg.Type != p.BC {
				e.ClearAbsorber()
				break
			}

			decodedBeamCode := p.DecodeBCPayload(beamCodeMsg.Payload)

			if decodedBeamCode != beamCode {
				if err := e.Send(*p.NewNO()); err != nil {
					return fmt.Errorf("failed to reply to absorber: %v", err)
				}
				e.ClearAbsorber()
				break
			}

			if err := e.Send(*p.NewOK()); err != nil {
				return fmt.Errorf("failed to reply to absorber: %v", err)
			}
			break outer
		}
	}

	if err := e.Send(*p.NewFI(filename, stats.Size())); err != nil {
		return fmt.Errorf("failed to send file info: %v", err)
	}
	oknoMsg, err := e.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive OK/NO msg after sending FI: %v", err)
	}
	if oknoMsg.Type != p.OK {
		fmt.Println("\nabsorber rejected file transfer")
		return nil
	}

	spinner.Stop()
	fmt.Println()

	pb := ui.NewProgressBar(stats.Size())
	cpBuff := make([]byte, 64*1000)
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
			return fmt.Errorf("failed to receive OK/NO message after sending FC: %v", err)
		}
		if oknoMsg.Type != p.OK {
			return errors.New("file chunk rejected by absorber")
		}

		pb.Update(int64(n))
	}

	fmt.Println("\nfile emitted!")
	return nil
}
