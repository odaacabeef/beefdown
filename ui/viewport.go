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

type coordinates struct {
	x, y int
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
	lines := strings.Split(body, "\n")

	// vertical space remaining for playables
	bodyHeight := v.height - lipgloss.Height(header)

	if bodyHeight >= len(lines) {
		return header + body
	}

	// the height necessary for the selected playable to be completely in view
	selectedHeight := lipgloss.Height(strings.Join(groups[0:selected.y+1], ""))

	// the first line of the selected playable
	selectedStart := selectedHeight - lipgloss.Height(groups[selected.y])

	// last viewable line
	lastLine := bodyHeight + v.startAt

	switch {
	case selectedHeight >= lastLine:
		// downward navigation
		v.startAt = selectedHeight - bodyHeight
	case v.startAt > selectedStart:
		// upward navigation
		v.startAt = selectedStart
	case lastLine > len(lines):
		// lines have been added (sequence changed & reloaded, new errors or warnings)
		v.startAt = len(lines) - bodyHeight
	}

	lastLine = bodyHeight + v.startAt

	if lipgloss.Height(body) > bodyHeight {
		body = strings.Join(lines[v.startAt:lastLine], "\n")
	}

	return header + body
}
