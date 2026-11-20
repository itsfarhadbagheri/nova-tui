package ui

import (
	"math"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// RowBetween puts the entries at opposite ends, favoring the right entry when space is tight.
func RowBetween(width int, left, right string) string {
	right = TruncateCells(right, width)
	rightWidth := lipgloss.Width(right)
	gap := 0
	if left != "" && width > rightWidth {
		gap = 1
	}
	left = TruncateCells(left, width-rightWidth-gap)
	return left + strings.Repeat(" ", max(0, width-lipgloss.Width(left)-rightWidth)) + right
}

// SingleLine keeps labels read from the system on one terminal row.
func SingleLine(value string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value))
}

// TruncateCells clips plain or ANSI-styled text at terminal-cell boundaries.
func TruncateCells(value string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(value, width, "…")
}

// Constrain clips a rendered block to a terminal rectangle.
func Constrain(value string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(value, "\n")
	lines = lines[:min(len(lines), height)]
	for i := range lines {
		lines[i] = TruncateCells(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func PercentageBar(styles Styles, value float64, width int) string {
	if width <= 0 {
		return ""
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		value = 0
	}
	filled := int(min(100, max(0, value)) / 100 * float64(width))
	return styles.Accent.Render(strings.Repeat("━", filled)) +
		styles.Track.Render(strings.Repeat("─", width-filled))
}
