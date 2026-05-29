package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/unquenchedservant/ChillClock/config"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == viewConfig {
			return m.handleConfigInput(msg)
		}
		return m.handleClockInput(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case serverStateMsg:
		return m.handleServerState(msg)
	case dingMsg:
		return m, nil
	case configLoadedMsg:
		m.serverConfig = msg.cfg
		return m, nil
	}
	return m, nil
}

func (m model) handleClockInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "?":
		if !m.timerRunning {
			m.mode = viewConfig
			m.selectedField = fieldPhase1DurationT1
			m.editingField = false
			m.inputBuffer = ""
			return m, fetchConfigCmd(m.serverURL)
		}
	case "r":
		return m.handleTimerSwitch()
	case "esc":
		if m.timerRunning {
			return m.handleTimerToggle(TIMER_1)
		}
	case "d":
		if !m.timerRunning {
			switch m.timerDefault {
			case TIMER_1:
				m.timerDefault = TIMER_2
			case TIMER_2:
				m.timerDefault = TIMER_1
			}
			m.config.DefaultTimer = m.timerDefault
			config.SaveTuiConfig(m.config)
		}
	case "1":
		return m.handleTimerToggle(TIMER_1)
	case "2":
		return m.handleTimerToggle(TIMER_2)
	case "enter", "space":
		return m.handleTimerToggle(m.timerDefault)
	}
	return m, nil
}

func (m model) handleTimerSwitch() (tea.Model, tea.Cmd) {
	return m, switchTimerCmd(m.serverURL)
}

func (m model) handleTimerToggle(timer int) (tea.Model, tea.Cmd) {
	return m, toggleServerCmd(m.serverURL, timer)
}
