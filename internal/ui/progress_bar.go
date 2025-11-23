package ui

import (
	"fmt"
	"strings"
)

type progressBar struct {
	toDo        int64
	progress    int64
	displayToDo string

	width     int
	fillChar  string
	emptyChar string
}

func NewProgressBar(workAmount int64) *progressBar {
	return &progressBar{
		toDo:        workAmount,
		progress:    0,
		displayToDo: FormatSize(workAmount),

		width:     24,
		fillChar:  "▰",
		emptyChar: " ",
	}
}

func (p *progressBar) Update(v int64) {
	p.progress = p.progress + v
	if p.progress > p.toDo {
		return
	}

	progressFraction := float64(p.progress) / float64(p.toDo)
	progressRounded := int(progressFraction * float64(p.width))
	toDoRounded := p.width - progressRounded
	progressDisplay, _ := convertSize(p.progress)

	fmt.Printf("\r\033K%s%s  %.0f%% (%.2f/%s) ",
		strings.Repeat(p.fillChar, progressRounded),
		strings.Repeat(p.emptyChar, toDoRounded),
		progressFraction*100,
		progressDisplay,
		p.displayToDo,
	)
}
