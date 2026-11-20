package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Accent lipgloss.TerminalColor
	Text   lipgloss.TerminalColor
	Muted  lipgloss.TerminalColor
	Border lipgloss.TerminalColor
	Track  lipgloss.TerminalColor
}

func DefaultTheme() Theme {
	return Theme{
		Accent: lipgloss.AdaptiveColor{Light: "#006B61", Dark: "#70D6C5"},
		Text:   lipgloss.AdaptiveColor{Light: "#202725", Dark: "#E4ECE9"},
		Muted:  lipgloss.AdaptiveColor{Light: "#52615D", Dark: "#A2B4AF"},
		Border: lipgloss.AdaptiveColor{Light: "#788B85", Dark: "#667D75"},
		Track:  lipgloss.AdaptiveColor{Light: "#91A39D", Dark: "#4F655E"},
	}
}
