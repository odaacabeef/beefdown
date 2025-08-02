package ui

import (
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

	playStart *time.Time
	playMu    sync.RWMutex // Mutex for protecting playStart

	playCh  chan struct{}
	stopCh  chan struct{}
	clockCh chan struct{}

	errs  []error
	errMu sync.RWMutex // Mutex for protecting errs
}

func (m *model) loadSequence(sequencePath string) error {
	// Store current selection state before reload
	oldSelected := m.selected
	oldGroupX := make(map[string]int)
	for k, v := range m.groupX {
		oldGroupX[k] = v
	}
	oldGroupNames := make([]string, len(m.groupNames))
	copy(oldGroupNames, m.groupNames)

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

	// Restore and validate selection state
	m.restoreSelection(oldSelected, oldGroupX, oldGroupNames)

	// Reset viewport state to ensure current selection is visible
	m.viewport.xStart = []int{}
	m.viewport.yStart = 0

	if m.device != nil {
		_, playables := m.getCurrentGroup()
		m.device.UpdateCurrentPlayable(playables[m.selected.x])
		m.device.SetPlaybackConfig(m.sequence.BPM, m.sequence.Loop, m.sequence.Sync)
	}

	return nil
}

// restoreSelection attempts to restore the previous selection state after a reload,
// adjusting coordinates to be within valid bounds if necessary
func (m *model) restoreSelection(oldSelected coordinates, oldGroupX map[string]int, oldGroupNames []string) {
	// Try to find the same group name in the new group list
	targetGroupIndex := -1
	if oldSelected.y >= 0 && oldSelected.y < len(oldGroupNames) {
		oldGroupName := oldGroupNames[oldSelected.y]
		for i, newGroupName := range m.groupNames {
			if newGroupName == oldGroupName {
				targetGroupIndex = i
				break
			}
		}
	}

	// If we found the same group, try to restore the selection within that group
	if targetGroupIndex >= 0 {
		groupName := m.groupNames[targetGroupIndex]
		playables := m.groups[groupName]

		// Restore the groupX value for this group
		if oldX, exists := oldGroupX[groupName]; exists && oldX >= 0 && oldX < len(playables) {
			m.groupX[groupName] = oldX
			m.selected.x = oldX
		} else {
			// Fallback to a valid index
			if len(playables) > 0 {
				m.groupX[groupName] = 0
				m.selected.x = 0
			} else {
				m.groupX[groupName] = 0
				m.selected.x = 0
			}
		}
		m.selected.y = targetGroupIndex
	} else {
		// If we couldn't find the same group, select the first available group
		if len(m.groupNames) > 0 {
			m.selected.y = 0
			firstGroup := m.groupNames[0]
			firstGroupPlayables := m.groups[firstGroup]
			if len(firstGroupPlayables) > 0 {
				m.selected.x = 0
				m.groupX[firstGroup] = 0
			} else {
				m.selected.x = 0
				m.groupX[firstGroup] = 0
			}
		} else {
			// No groups available, reset to safe defaults
			m.selected.x = 0
			m.selected.y = 0
		}
	}

	// Final validation to ensure coordinates are within bounds
	if m.selected.y >= len(m.groupNames) {
		m.selected.y = len(m.groupNames) - 1
		if m.selected.y < 0 {
			m.selected.y = 0
		}
	}

	if m.selected.y >= 0 && m.selected.y < len(m.groupNames) {
		groupName := m.groupNames[m.selected.y]
		playables := m.groups[groupName]
		if m.selected.x >= len(playables) {
			m.selected.x = len(playables) - 1
			if m.selected.x < 0 {
				m.selected.x = 0
			}
		}
		m.groupX[groupName] = m.selected.x
	}
}

// validateSelection ensures that the current selection coordinates are within valid bounds
func (m *model) validateSelection() {
	// Ensure y coordinate is within bounds
	if m.selected.y < 0 {
		m.selected.y = 0
	}
	if m.selected.y >= len(m.groupNames) {
		if len(m.groupNames) > 0 {
			m.selected.y = len(m.groupNames) - 1
		} else {
			m.selected.y = 0
		}
	}

	// Ensure x coordinate is within bounds for the current group
	if m.selected.y >= 0 && m.selected.y < len(m.groupNames) {
		groupName := m.groupNames[m.selected.y]
		playables := m.groups[groupName]

		if m.selected.x < 0 {
			m.selected.x = 0
		}
		if m.selected.x >= len(playables) {
			if len(playables) > 0 {
				m.selected.x = len(playables) - 1
			} else {
				m.selected.x = 0
			}
		}

		// Update groupX to match the validated selection
		m.groupX[groupName] = m.selected.x
	}
}

// getCurrentGroup safely returns the current group name and playables, with bounds checking
func (m *model) getCurrentGroup() (string, []sequence.Playable) {
	if m.selected.y < 0 || m.selected.y >= len(m.groupNames) {
		if len(m.groupNames) > 0 {
			return m.groupNames[0], m.groups[m.groupNames[0]]
		}
		return "", nil
	}
	groupName := m.groupNames[m.selected.y]
	playables := m.groups[groupName]
	return groupName, playables
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		listenForDevicePlay(m.playCh),
		listenForDeviceClock(m.clockCh),
		listenForDeviceErrors(m.device.ErrorsCh()),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case devicePlay:
		m.mu.Lock()
		groupName, playables := m.getCurrentGroup()
		if groupName != "" && len(playables) > 0 && m.selected.x < len(playables) {
			now := time.Now()
			m.playMu.Lock()
			m.playStart = &now
			m.playMu.Unlock()
			playing := m.selected
			m.playing = &playing
		}
		m.mu.Unlock()
		return m, listenForDeviceStop(m.stopCh)

	case deviceStop:
		// Device stopped playing - update UI state
		for _, p := range m.sequence.Playable {
			p.ClearStep()
		}
		m.mu.Lock()
		m.playing = nil
		m.mu.Unlock()
		m.playMu.Lock()
		m.playStart = nil
		m.playMu.Unlock()
		return m, listenForDevicePlay(m.playCh)

	case deviceClock:
		return m, listenForDeviceClock(m.clockCh)

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
			if m.device.Playing() && m.device.CancelF != nil {
				m.device.CancelF()
			}
			return m, tea.Quit

		case "R":
			if m.device.Playing() {
				m.device.CancelF()
			}
			err := m.loadSequence(m.sequence.Path)

			m.errMu.Lock()
			m.errs = []error{} // clear errors
			if err != nil {
				m.errs = append(m.errs, err)
			}
			m.errMu.Unlock()

		case "h", "left":
			m.mu.Lock()
			groupName, playables := m.getCurrentGroup()
			if groupName != "" && len(playables) > 0 && m.selected.x > 0 {
				m.selected.x--
				m.groupX[groupName] = m.selected.x
				// Update the device's current playable
				if m.selected.x < len(playables) {
					m.device.UpdateCurrentPlayable(playables[m.selected.x])
				}
			}
			m.mu.Unlock()

		case "l", "right":
			m.mu.Lock()
			groupName, playables := m.getCurrentGroup()
			if groupName != "" && m.selected.x < len(playables)-1 {
				m.selected.x++
				m.groupX[groupName] = m.selected.x
				// Update the device's current playable
				if m.selected.x < len(playables) {
					m.device.UpdateCurrentPlayable(playables[m.selected.x])
				}
			}
			m.mu.Unlock()

		case "k", "up":
			m.mu.Lock()
			if m.selected.y > 0 {
				m.selected.y--
				groupName, playables := m.getCurrentGroup()
				if groupName != "" {
					// Ensure the restored x coordinate is within bounds
					if m.groupX[groupName] >= len(playables) {
						m.groupX[groupName] = 0
					}
					m.selected.x = m.groupX[groupName]
					// Update the device's current playable
					if m.selected.x < len(playables) {
						m.device.UpdateCurrentPlayable(playables[m.selected.x])
					}
				}
			}
			m.mu.Unlock()

		case "j", "down":
			m.mu.Lock()
			if m.selected.y < len(m.groupNames)-1 {
				m.selected.y++
				groupName, playables := m.getCurrentGroup()
				if groupName != "" {
					// Ensure the restored x coordinate is within bounds
					if m.groupX[groupName] >= len(playables) {
						m.groupX[groupName] = 0
					}
					m.selected.x = m.groupX[groupName]
					// Update the device's current playable
					if m.selected.x < len(playables) {
						m.device.UpdateCurrentPlayable(playables[m.selected.x])
					}
				}
			}
			m.mu.Unlock()

		case "0":
			m.mu.Lock()
			groupName, playables := m.getCurrentGroup()
			if groupName != "" && len(playables) > 0 {
				m.selected.x = 0
				m.groupX[groupName] = m.selected.x
				// Update the device's current playable
				if m.selected.x < len(playables) {
					m.device.UpdateCurrentPlayable(playables[m.selected.x])
				}
			}
			m.mu.Unlock()

		case "$":
			m.mu.Lock()
			groupName, playables := m.getCurrentGroup()
			if groupName != "" && len(playables) > 0 {
				m.selected.x = len(playables) - 1
				m.groupX[groupName] = m.selected.x
				// Update the device's current playable
				if m.selected.x < len(playables) {
					m.device.UpdateCurrentPlayable(playables[m.selected.x])
				}
			}
			m.mu.Unlock()

		case "g":
			m.mu.Lock()
			if len(m.groupNames) > 0 {
				m.selected.y = 0
				groupName, playables := m.getCurrentGroup()
				if groupName != "" {
					// Ensure the restored x coordinate is within bounds
					if m.groupX[groupName] >= len(playables) {
						m.groupX[groupName] = 0
					}
					m.selected.x = m.groupX[groupName]
					// Update the device's current playable
					if m.selected.x < len(playables) {
						m.device.UpdateCurrentPlayable(playables[m.selected.x])
					}
				}
			}
			m.mu.Unlock()

		case "G":
			m.mu.Lock()
			if len(m.groupNames) > 0 {
				m.selected.y = len(m.groupNames) - 1
				groupName, playables := m.getCurrentGroup()
				if groupName != "" {
					// Ensure the restored x coordinate is within bounds
					if m.groupX[groupName] >= len(playables) {
						m.groupX[groupName] = 0
					}
					m.selected.x = m.groupX[groupName]
					// Update the device's current playable
					if m.selected.x < len(playables) {
						m.device.UpdateCurrentPlayable(playables[m.selected.x])
					}
				}
			}
			m.mu.Unlock()

		case " ":
			if m.sequence.Sync == "follower" {
				break
			}
			// Toggle playback state
			if m.device.Stopped() {
				// Device is stopped, start playback
				m.device.PlaySub.Pub()
			} else {
				// Device is playing or in unknown state, stop it
				m.device.StopSub.Pub()
			}
		}
	}

	return m, nil
}

func (m *model) View() string {
	// Ensure selection coordinates are valid before rendering
	m.validateSelection()

	var groupNames []string
	var groupX []int
	var groupPlayables [][]string

	st := style(lipgloss.NewStyle())

	header := fmt.Sprintf("%s;", m.sequence.Path)
	if m.sequence.Sync != "follower" {
		header += fmt.Sprintf(" bpm: %f; loop: %v;", m.sequence.BPM, m.sequence.Loop)
	}
	header += fmt.Sprintf(" sync: %s", m.sequence.Sync)
	header = st.sequence().Render(header)

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
