package main

import (
	"fmt"
	"time"
)

type spinner struct {
	frames    []string
	currFrame int
	stopch    chan int
	timeStep  time.Duration
}

func NewSpinner() *spinner {
	return &spinner{
		frames:    []string{"Waiting.  ", "Waiting.. ", "Waiting..."},
		currFrame: 0,
		stopch:    make(chan int),
		timeStep:  200 * time.Millisecond,
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
	fmt.Print("\r\033[K", s.frames[s.currFrame])
	s.currFrame++
}
