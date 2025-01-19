package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case devicePlay:
		now := time.Now()
		m.playStart = &now
		return m, listenForDeviceStop(m.device.StopCh())

	case deviceStop:
		m.playing = nil
		m.playStart = nil
		return m, listenForDevicePlay(m.device.PlayCh())

	case deviceClock:
		if m.device.Playing() {
			return m, listenForDeviceClock(m.device.ClockCh())
		}

	case deviceError:
		m.errs = append(m.errs, msg)
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
				p := m.groups[m.groupNames[m.selected.y]][m.selected.x]
				m.playing = &m.selected
				ctx, stop := context.WithCancel(context.Background())
				m.stop = stop
				m.device.Play(ctx, p, m.sequence.BPM, m.sequence.Loop, m.sequence.Sync)
				return m, listenForDeviceClock(m.device.ClockCh())

			case m.device.Playing():
				m.stop()
			}
		}
	}

	return m, nil
}
