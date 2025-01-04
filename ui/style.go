package ui

import "github.com/charmbracelet/lipgloss"

type style lipgloss.Style

func (s style) state() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(1)
}

func (s style) sequence() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(1)
}

func (s style) errors() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Margin(0, 1)
}

func (s style) playable(selected, playing bool) lipgloss.Style {
	base := lipgloss.NewStyle().
		Padding(0, 1).
		Margin(1)

	switch {
	case playing:
		return base.Border(lipgloss.DoubleBorder())
	case selected:
		return base.Border(lipgloss.NormalBorder())
	}
	return base.Border(lipgloss.HiddenBorder())
}

func (s style) groupName() lipgloss.Style {
	return lipgloss.NewStyle().
		Padding(1).
		Margin(1)
}
