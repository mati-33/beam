package ui

import (
	"fmt"
	"strings"
)

const (
	bThreshold  = 100
	kbThreshold = 100_000
	mbThreshold = 100_000_000
)

func convertSize(size int64) (float64, string) {
	if size < bThreshold {
		return float64(size), "B"
	} else if size < kbThreshold {
		return float64(size) / 1_000., "KB"
	} else if size < mbThreshold {
		return float64(size) / 1_000_000., "MB"
	} else {
		return float64(size) / 1_000_000_000., "GB"
	}
}

func FormatSize(size int64) string {
	s, u := convertSize(size)
	return fmt.Sprintf("%s %s", strings.TrimRight(fmt.Sprintf("%.2f", s), ".0"), u)
}
