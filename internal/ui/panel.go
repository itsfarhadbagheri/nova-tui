package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const minPanelWidth = 16

func PanelBodyWidth(width int) int {
	if width < minPanelWidth {
		return max(0, width)
	}
	return width - 4
}

// Panel renders a titled frame with optional extra height for alignment.
func Panel(styles Styles, width, height int, title string, body []string) string {
	if width <= 0 {
		return ""
	}
	if width < minPanelWidth {
		lines := []string{styles.Accent.Render(title)}
		lines = append(lines, body...)
		return Constrain(strings.Join(lines, "\n"), width, len(lines))
	}
	title = TruncateCells(title, width-6)
	border := styles.Border.Render
	lines := []string{border("╭─ ") + styles.Accent.Render(title) + border(" "+strings.Repeat("─", width-5-lipgloss.Width(title))+"╮")}
	for i := range max(len(body), height-2) {
		line := ""
		if i < len(body) {
			line = TruncateCells(body[i], PanelBodyWidth(width))
		}
		lines = append(lines, border("│")+" "+line+strings.Repeat(" ", width-3-lipgloss.Width(line))+border("│"))
	}
	lines = append(lines, border("╰"+strings.Repeat("─", width-2)+"╯"))
	return strings.Join(lines, "\n")
}
