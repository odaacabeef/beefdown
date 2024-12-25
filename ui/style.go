package ui

import "github.com/charmbracelet/lipgloss"

type style lipgloss.Style

func (s style) state() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(1)
}

func (s style) errors() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1)
}

func (s style) block() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(1)
}
