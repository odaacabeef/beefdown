package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type viewport struct {
	width  int
	height int

	startAt int
}

func (v *viewport) dim(w, h int) {
	v.width = w
	v.height = h
}

func (v *viewport) view(header string, groups []string, selected coordinates) string {

	if v.height == 0 {
		return ""
	}

	body := strings.Join(groups, "")

	// vertical space remaining for playables
	bodyHeight := v.height - lipgloss.Height(header)

	// the height necessary for the selected playable to be completely in view
	selectedHeight := lipgloss.Height(strings.Join(groups[0:selected.y+1], ""))

	// the first line of the selected playable
	selectedStart := selectedHeight - lipgloss.Height(groups[selected.y])

	// last viewable line
	lastLine := bodyHeight + v.startAt

	switch {
	case selectedHeight > lastLine:
		v.startAt = selectedHeight - bodyHeight
	case v.startAt > selectedStart:
		v.startAt = selectedStart
	}

	if lipgloss.Height(body) > bodyHeight {
		body = strings.Join(strings.Split(body, "\n")[v.startAt:bodyHeight+v.startAt], "\n")
	}

	return header + body
}
