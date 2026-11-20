package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Title string
	Width int
	Right bool
}

// Table renders fixed terminal-cell columns with a one-cell gutter.
func Table(styles Styles, columns []Column, rows [][]string) []string {
	headers := make([]string, len(columns))
	for i, column := range columns {
		headers[i] = column.Title
	}
	lines := []string{styles.Label.Render(tableRow(columns, headers))}
	for _, row := range rows {
		lines = append(lines, styles.Value.Render(tableRow(columns, row)))
	}
	return lines
}

func tableRow(columns []Column, values []string) string {
	cells := make([]string, len(columns))
	for i, column := range columns {
		text := TruncateCells(values[i], column.Width)
		padding := strings.Repeat(" ", max(0, column.Width-lipgloss.Width(text)))
		if column.Right {
			cells[i] = padding + text
		} else {
			cells[i] = text + padding
		}
	}
	return strings.Join(cells, " ")
}
