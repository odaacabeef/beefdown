package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trotttrotttrott/seq/device"
	"github.com/trotttrotttrott/seq/sequence"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	device   *device.Device
	sequence *sequence.Sequence

	groupNames []string
	groups     map[string][]sequence.Playable
	groupX     map[string]int

	selected coordinates
	playing  *coordinates

	stop context.CancelFunc

	errs []error
}

type coordinates struct {
	x, y int
}

func Start(sequencePath string) error {

	m, err := initialModel(sequencePath)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModel(sequencePath string) (*model, error) {
	d, err := device.New()
	if err != nil {
		return nil, err
	}

	m := model{
		device: d,
	}

	err = m.loadSequence(sequencePath)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *model) loadSequence(sequencePath string) error {
	s, err := sequence.New(sequencePath)
	if err != nil {
		return err
	}
	m.sequence = s
	m.groups = map[string][]sequence.Playable{}
	m.groupNames = []string{}
	m.groupX = map[string]int{}
groupPlayables:
	for _, p := range m.sequence.Playable {
		m.groups[p.Group()] = append(m.groups[p.Group()], p)
		for _, name := range m.groupNames {
			if name == p.Group() {
				continue groupPlayables
			}
		}
		m.groupNames = append(m.groupNames, p.Group())
		m.groupX[p.Group()] = 0
	}
	return nil
}

type deviceTick <-chan time.Time

func listenForDeviceTick(c <-chan time.Time) tea.Cmd {
	return func() tea.Msg {
		return deviceTick(c)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case deviceTick:
		switch {
		case m.device.Stopped():
			m.playing = nil
		case m.device.Playing():
			return m, listenForDeviceTick(m.device.TickerC())
		}

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			if m.device.Playing() {
				m.stop()
			}
			return m, tea.Quit

		case "R":
			if m.device.Playing() {
				m.stop()
			}
			err := m.loadSequence(m.sequence.Path)
			if err != nil {
				m.errs = append(m.errs, err)
			}

		case "h", "left":
			if m.selected.x > 0 {
				m.selected.x--
				m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			}

		case "l", "right":
			if m.selected.x < len(m.groups[m.groupNames[m.selected.y]])-1 {
				m.selected.x++
				m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			}

		case "k", "up":
			if m.selected.y > 0 {
				m.selected.y--
				m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			}

		case "j", "down":
			if m.selected.y < len(m.groupNames)-1 {
				m.selected.y++
				m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			}

		case "0":
			m.selected.x = 0
			m.groupX[m.groupNames[m.selected.y]] = m.selected.x

		case "$":
			m.selected.x = len(m.groups[m.groupNames[m.selected.y]]) - 1
			m.groupX[m.groupNames[m.selected.y]] = m.selected.x

		case "g":
			m.selected.y = 0
			m.selected.x = m.groupX[m.groupNames[m.selected.y]]

		case "G":
			m.selected.y = len(m.groupNames) - 1
			m.selected.x = m.groupX[m.groupNames[m.selected.y]]

		case " ":
			switch {
			case m.device.Stopped():
				for _, p := range m.sequence.Playable {
					p.ClearStep()
				}
				selected := 0
				for i, g := range m.groupNames {
					switch {
					case i < m.selected.y:
						selected += len(m.groups[g])
					case i == m.selected.y:
						selected += m.selected.x
					default:
						continue
					}
				}
				p := m.sequence.Playable[selected]
				m.playing = &m.selected
				ctx, stop := context.WithCancel(context.Background())
				m.stop = stop
				m.device.Play(ctx, p, m.sequence.BPM, m.sequence.Loop)
				return m, listenForDeviceTick(m.device.TickerC())
			case m.device.Playing():
				m.stop()
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	st := style(lipgloss.NewStyle())

	s := ""

	s += st.metadata().Render(fmt.Sprintf("%s; bpm: %f; loop: %v", m.sequence.Path, m.sequence.BPM, m.sequence.Loop))

	s += st.state().Render(fmt.Sprintf("state: %s", m.device.State()))

	if len(m.errs) > 0 {
		s += st.errors().Render(fmt.Sprintf("%v", m.errs))
	}

	for gIdx, g := range m.groupNames {
		var sb strings.Builder
		for _, char := range g {
			sb.WriteRune(char)
			sb.WriteString("\n")
		}
		var playables []string
		for pIdx, p := range m.groups[g] {
			playables = append(playables, st.playable(
				pIdx == m.selected.x && gIdx == m.selected.y,
				m.playing != nil && pIdx == m.playing.x && gIdx == m.playing.y,
			).Render(p.String()))
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, append([]string{st.groupName().Render(sb.String())}, playables...)...)
	}

	return s
}
