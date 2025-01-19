package ui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {

	var header string
	var groups []string

	st := style(lipgloss.NewStyle())

	header = st.sequence().Render(fmt.Sprintf("%s; bpm: %f; loop: %v; sync: %s", m.sequence.Path, m.sequence.BPM, m.sequence.Loop, m.sequence.Sync))

	t := ""
	if m.playStart != nil {
		t = fmt.Sprintf(" (%s)", time.Now().Sub(*m.playStart).Round(time.Second))
	}
	header += st.state().Render(fmt.Sprintf("state: %s%s", m.device.State(), t))

	if len(m.errs) > 0 {
		var errstr []string
		for _, err := range m.errs {
			errstr = append(errstr, err.Error())
		}
		errstr = append(errstr, fmt.Sprintf("%d errors:", len(m.errs)))
		slices.Reverse(errstr)
		header += st.errors().Render(strings.Join(errstr, "\n"))
	}
	w := m.sequence.Warnings()
	if len(w) > 0 {
		header += st.warnings().Render(strings.Join(w, "\n"))
	}

	header = st.header(m.viewport.width).Render(header)

	for gIdx, g := range m.groupNames {
		var sb strings.Builder
		for _, char := range g {
			sb.WriteRune(char)
			sb.WriteString("\n")
		}
		var playables []string
		for pIdx, p := range m.groups[g] {
			steps := p.Steps()
			lines := strings.Split(steps, "\n")
			chunkSize := 16
			if len(lines) > chunkSize {
				var chunks []string
				for chunkSize < len(lines) {
					lines, chunks = lines[chunkSize:], append(chunks, strings.Join(lines[0:chunkSize:chunkSize], "\n"))
					chunks = append(chunks, "  ")
				}
				steps = lipgloss.JoinHorizontal(lipgloss.Top, append(chunks, strings.Join(lines, "\n"))...)
			}
			playables = append(playables, st.playable(
				pIdx == m.selected.x && gIdx == m.selected.y,
				m.playing != nil && pIdx == m.playing.x && gIdx == m.playing.y,
			).Render(p.Title()+steps))
		}
		groups = append(groups, lipgloss.JoinHorizontal(lipgloss.Top, append([]string{st.groupName().Render(sb.String())}, playables...)...))
	}

	return m.viewport.view(header, groups, m.selected)
}
