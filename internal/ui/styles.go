package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Label  lipgloss.Style
	Value  lipgloss.Style
	Accent lipgloss.Style
	Border lipgloss.Style
	Track  lipgloss.Style
}

func NewStyles(t Theme) Styles {
	return Styles{
		Label:  lipgloss.NewStyle().Foreground(t.Muted),
		Value:  lipgloss.NewStyle().Foreground(t.Text).Bold(true),
		Accent: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Border: lipgloss.NewStyle().Foreground(t.Border),
		Track:  lipgloss.NewStyle().Foreground(t.Track),
	}
}
