package ui

import (
	"fmt"
	"time"
)

type spinner struct {
	frames    []string
	text      string
	currFrame int
	stopch    chan int
	timeStep  time.Duration
}

func NewSpinner() *spinner {
	return &spinner{
		frames:    []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		text:      "Waiting for absorber...",
		currFrame: 0,
		stopch:    make(chan int),
		timeStep:  100 * time.Millisecond,
	}
}

func (s *spinner) Start() {
	go func() {
		for {
			select {
			case <-s.stopch:
				return
			case <-time.After(s.timeStep):
				s.update()
			}
		}
	}()
}

func (s *spinner) Stop() {
	s.stopch <- 1
}

func (s *spinner) update() {
	if s.currFrame == len(s.frames) {
		s.currFrame = 0
	}
	fmt.Printf("\r\033[K%s %s", s.frames[s.currFrame], s.text)
	s.currFrame++
}
