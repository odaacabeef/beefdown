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

	selected coordinates
	playing  *coordinates

	ctx    context.Context
	cancel context.CancelFunc

	errs []error
}

type coordinates struct {
	x, y int
}

func Start() error {

	m, err := initialModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func initialModel() (*model, error) {
	d, err := device.New()
	if err != nil {
		return nil, err
	}

	m := model{
		device: d,
	}

	err = m.loadSequence()
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *model) loadSequence() error {
	s, err := sequence.New("_test/test.md")
	if err != nil {
		return err
	}
	m.sequence = s
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
				m.cancel()
			}
			return m, tea.Quit

		case "R":
			if m.device.Playing() {
				m.cancel()
			}
			err := m.loadSequence()
			if err != nil {
				m.errs = append(m.errs, err)
			}

		case "h", "left":
			if m.selected.x > 0 {
				m.selected.x--
			}

		case "l", "right":
			if m.selected.x < len(m.sequence.Playable)-1 {
				m.selected.x++
			}

		case "0":
			m.selected.x = 0

		case "$":
			m.selected.x = len(m.sequence.Playable) - 1

		case " ":
			switch {
			case m.device.Stopped():
				for _, p := range m.sequence.Playable {
					p.ClearStep()
				}
				p := m.sequence.Playable[m.selected.x]
				m.playing = &m.selected
				m.ctx, m.cancel = context.WithCancel(context.Background())
				m.device.Play(m.ctx, m.sequence.BPM, p)
				return m, listenForDeviceTick(m.device.TickerC())
			case m.device.Playing():
				m.cancel()
			}
		}
	}

	return m, nil
}

func (m model) View() string {

	st := style(lipgloss.NewStyle())

	s := ""

	s += st.metadata().Render(fmt.Sprintf("%s; bpm: %f", m.sequence.Path, m.sequence.BPM))

	s += st.state().Render(fmt.Sprintf("state: %s", m.device.State()))

	if len(m.errs) > 0 {
		s += st.errors().Render(fmt.Sprintf("%v", m.errs))
	}

	var groupkeys []string
	groups := map[string][]string{}

groupPlayables:
	for i, p := range m.sequence.Playable {
		groups[p.Group()] = append(groups[p.Group()], st.playable(i == m.selected.x, (m.playing != nil && i == m.playing.x)).Render(p.String()))
		for _, k := range groupkeys {
			if k == p.Group() {
				continue groupPlayables
			}
		}
		groupkeys = append(groupkeys, p.Group())
	}

	for _, g := range groupkeys {
		var sb strings.Builder
		for _, char := range g {
			sb.WriteRune(char)
			sb.WriteString("\n")
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, append([]string{st.groupName().Render(sb.String())}, groups[g]...)...)
	}

	return s
}
