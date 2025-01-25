package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type viewport struct {
	width,
	height,
	yStart int
	xStart []int
}

type coordinates struct {
	x, y int
}

func (v *viewport) dim(w, h int) {
	v.width = w
	v.height = h
}

func (v *viewport) view(header string, groupNames []string, groupX []int, groupPlayables [][]string, selected coordinates) string {

	if v.height == 0 {
		return ""
	}

	return v.cropY(header, v.cropX(groupNames, groupX, groupPlayables), selected)
}

func (v *viewport) cropX(groupNames []string, groupX []int, groupPlayables [][]string) (x []string) {

	joinHorizontal := func(str ...string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, str...)
	}

	for i, playables := range groupPlayables {

		aside := groupNames[i]

		row := joinHorizontal(playables...)

		// horizontal width available for group playables
		rowWidth := v.width - lipgloss.Width(aside)

		if rowWidth >= lipgloss.Width(row) {
			x = append(x, joinHorizontal(aside, row))
			continue
		}

		// the width necessary for this groups (last) selected playable to be completely in view
		xSelectedWidth := lipgloss.Width(joinHorizontal(playables[0 : groupX[i]+1]...))

		// the index of the first charater of the (last) selected playable
		xSelectedStart := xSelectedWidth - lipgloss.Width(playables[groupX[i]])

		if i == len(v.xStart) {
			v.xStart = append(v.xStart, 0)
		}

		// last x index
		xLast := rowWidth + v.xStart[i]

		switch {
		case xSelectedWidth >= xLast:
			v.xStart[i] = xSelectedWidth - rowWidth
		case v.xStart[i] > xSelectedStart:
			v.xStart[i] = xSelectedStart
		}

		xLast = rowWidth + v.xStart[i]

		var linesCropped []string
		for _, line := range strings.Split(row, "\n") {
			linesCropped = append(linesCropped, line[v.xStart[i]:])
		}

		x = append(x, joinHorizontal(aside, strings.Join(linesCropped, "\n")))
	}

	return x
}

func (v *viewport) cropY(header string, groups []string, selected coordinates) string {

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
	lastLine := bodyHeight + v.yStart

	switch {
	case selectedHeight >= lastLine:
		// downward navigation
		v.yStart = selectedHeight - bodyHeight
	case v.yStart > selectedStart:
		// upward navigation
		v.yStart = selectedStart
	case lastLine > len(lines):
		// lines have been added (sequence changed & reloaded, new errors or warnings)
		v.yStart = len(lines) - bodyHeight
	}

	lastLine = bodyHeight + v.yStart

	if lipgloss.Height(body) > bodyHeight {
		body = strings.Join(lines[v.yStart:lastLine], "\n")
	}

	return header + body
}
