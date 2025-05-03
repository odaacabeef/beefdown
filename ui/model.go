package ui

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/odaacabeef/beefdown/device"
	"github.com/odaacabeef/beefdown/sequence"
)

type model struct {
	device   *device.Device
	sequence *sequence.Sequence

	groupNames []string
	groups     map[string][]sequence.Playable
	groupX     map[string]int

	selected coordinates
	playing  *coordinates
	mu       sync.RWMutex // Mutex for protecting shared state

	viewport *viewport

	stop context.CancelFunc

	playStart *time.Time
	playMu    sync.RWMutex // Mutex for protecting playStart

	errs  []error
	errMu sync.RWMutex // Mutex for protecting errs
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

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		listenForDevicePlay(m.device.PlayCh()),
		listenForDeviceClock(m.device.ClockCh()),
		listenForDeviceErrors(m.device.ErrorsCh()),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case devicePlay:
		now := time.Now()
		m.playMu.Lock()
		m.playStart = &now
		m.playMu.Unlock()
		return m, listenForDeviceStop(m.device.StopCh())

	case deviceStop:
		m.mu.Lock()
		m.playing = nil
		m.mu.Unlock()
		m.playMu.Lock()
		m.playStart = nil
		m.playMu.Unlock()
		return m, listenForDevicePlay(m.device.PlayCh())

	case deviceClock:
		return m, listenForDeviceClock(m.device.ClockCh())

	case deviceError:
		m.errMu.Lock()
		m.errs = append(m.errs, msg)
		m.errMu.Unlock()
		return m, listenForDeviceErrors(m.device.ErrorsCh())

	case tea.WindowSizeMsg:
		m.viewport.dim(msg.Width, msg.Height)

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
				m.errMu.Lock()
				m.errs = append(m.errs, err)
				m.errMu.Unlock()
			}

		case "h", "left":
			m.mu.Lock()
			if m.selected.x > 0 {
				m.selected.x--
				m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			}
			m.mu.Unlock()

		case "l", "right":
			m.mu.Lock()
			if m.selected.x < len(m.groups[m.groupNames[m.selected.y]])-1 {
				m.selected.x++
				m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			}
			m.mu.Unlock()

		case "k", "up":
			m.mu.Lock()
			if m.selected.y > 0 {
				m.selected.y--
				m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			}
			m.mu.Unlock()

		case "j", "down":
			m.mu.Lock()
			if m.selected.y < len(m.groupNames)-1 {
				m.selected.y++
				m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			}
			m.mu.Unlock()

		case "0":
			m.mu.Lock()
			m.selected.x = 0
			m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			m.mu.Unlock()

		case "$":
			m.mu.Lock()
			m.selected.x = len(m.groups[m.groupNames[m.selected.y]]) - 1
			m.groupX[m.groupNames[m.selected.y]] = m.selected.x
			m.mu.Unlock()

		case "g":
			m.mu.Lock()
			m.selected.y = 0
			m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			m.mu.Unlock()

		case "G":
			m.mu.Lock()
			m.selected.y = len(m.groupNames) - 1
			m.selected.x = m.groupX[m.groupNames[m.selected.y]]
			m.mu.Unlock()

		case " ":
			switch {
			case m.device.Stopped():
				for _, p := range m.sequence.Playable {
					p.ClearStep()
				}
				m.mu.Lock()
				p := m.groups[m.groupNames[m.selected.y]][m.selected.x]
				playing := m.selected
				m.playing = &playing
				m.mu.Unlock()
				ctx, stop := context.WithCancel(context.Background())
				m.stop = stop
				m.device.Play(ctx, p, m.sequence.BPM, m.sequence.Loop, m.sequence.Sync)
			case m.device.Playing():
				m.stop()
			}
		}
	}

	return m, nil
}

func (m *model) View() string {

	var header string
	var groupNames []string
	var groupX []int
	var groupPlayables [][]string

	st := style(lipgloss.NewStyle())

	header = st.sequence().Render(fmt.Sprintf("%s; bpm: %f; loop: %v; sync: %s", m.sequence.Path, m.sequence.BPM, m.sequence.Loop, m.sequence.Sync))

	t := "-"
	m.playMu.RLock()
	if m.playStart != nil {
		t = fmt.Sprintf("%s", time.Now().Sub(*m.playStart).Round(time.Second))
	}
	m.playMu.RUnlock()
	header += st.state().Render(fmt.Sprintf("state: %s; goroutines: %d; time: %s", m.device.State(), runtime.NumGoroutine(), t))

	m.errMu.RLock()
	if len(m.errs) > 0 {
		var errstr []string
		for _, err := range m.errs {
			errstr = append(errstr, err.Error())
		}
		errstr = append(errstr, fmt.Sprintf("%d errors:", len(m.errs)))
		slices.Reverse(errstr)
		header += st.errors().Render(strings.Join(errstr, "\n"))
	}
	m.errMu.RUnlock()
	w := m.sequence.Warnings()
	if len(w) > 0 {
		header += st.warnings().Render(strings.Join(w, "\n"))
	}

	header = st.header(m.viewport.width).Render(header)

	for gIdx, groupName := range m.groupNames {
		var playables []string
		for pIdx, p := range m.groups[groupName] {
			steps := p.Steps()
			lines := strings.Split(steps, "\n")
			// limit playables to 16 vertical steps
			// wrap them horizontally
			chunkSize := 16
			if len(lines) > chunkSize {
				var chunks []string
				for chunkSize < len(lines) {
					lines, chunks = lines[chunkSize:], append(chunks, strings.Join(lines[0:chunkSize:chunkSize], "\n"))
					chunks = append(chunks, "  ")
				}
				steps = lipgloss.JoinHorizontal(lipgloss.Top, append(chunks, strings.Join(lines, "\n"))...)
			}
			selected := pIdx == m.selected.x && gIdx == m.selected.y
			playing := m.playing != nil && pIdx == m.playing.x && gIdx == m.playing.y
			playables = append(playables, st.playable(selected, playing).Render(p.Title()+steps))
		}
		// group name displayed vertically
		groupNames = append(groupNames, st.groupName().Render(strings.Join(strings.Split(groupName, ""), "\n")))
		groupX = append(groupX, m.groupX[groupName])
		groupPlayables = append(groupPlayables, playables)
	}

	return m.viewport.view(header, groupNames, groupX, groupPlayables, m.selected)
}
