package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/unquenchedservant/ChillClock/config"
	util "github.com/unquenchedservant/ChillClock/utilities"
)

type serverStateMsg struct {
	Text         string `json:"text"`
	Class        string `json:"class"`
	Phase        int    `json:"phase"`
	Running      bool   `json:"running"`
	RemainingInt int64  `json:"remainingInt"`
	Elapsed      int64  `json:"elapsed"`
	ActiveTimer  int    `json:"activeTimer"`
	Temp         int    `json:"temp"`
}

func pollServerCmd(serverUrl string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)
		req, _ := http.NewRequest(http.MethodGet, serverUrl+"/status", nil)
		req.Header.Set("X-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			return serverStateMsg{}
		}
		defer resp.Body.Close()
		var msg serverStateMsg
		if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
			return serverStateMsg{}
		}
		return msg
	}
}

func (m model) handleServerState(msg serverStateMsg) (tea.Model, tea.Cmd) {
	oldPhase := m.currentPhase
	m.timerRunning = msg.Running
	m.timerElapsed = time.Duration(msg.Elapsed) * time.Millisecond
	m.currentPhase = timerPhase(msg.Phase)
	m.timer = msg.ActiveTimer
	m.temp = msg.Temp

	if oldPhase != m.currentPhase && m.currentPhase != phaseNotStarted && m.timerRunning {
		return m, tea.Batch(pollServerCmd(m.serverURL, m.apiKey), dingCmd(m.currentPhase, m.temp))
	}
	return m, pollServerCmd(m.serverURL, m.apiKey)
}

func toggleServerCmd(serverURL string, timer int, apiKey string) tea.Cmd {
	return func() tea.Msg {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/toggle?timer=%d", serverURL, timer), nil)
		req.Header.Set("X-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return serverStateMsg{}
		}
		defer resp.Body.Close()
		var msg serverStateMsg
		json.NewDecoder(resp.Body).Decode(&msg)
		return msg
	}
}

func switchTimerCmd(serverURL string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/switch", serverURL), nil)
		req.Header.Set("X-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return serverStateMsg{}
		}

		defer resp.Body.Close()
		var msg serverStateMsg
		json.NewDecoder(resp.Body).Decode(&msg)
		return msg
	}
}
func (m model) getTimerDisplay() (string, lipgloss.Style) {
	var duration int
	switch m.timer {
	case TIMER_1:
		duration = m.serverConfig.Timer.Phase1Duration_Timer1 + m.serverConfig.Timer.Phase2Duration_Timer1 + m.serverConfig.Timer.Phase3Duration_Timer1
	case TIMER_2:
		duration = m.serverConfig.Timer.Phase1Duration_Timer2 + m.serverConfig.Timer.Phase2Duration_Timer2 + m.serverConfig.Timer.Phase3Duration_Timer2
	}
	if (!m.timerRunning && m.currentPhase == phaseNotStarted) || m.currentPhase == phaseCompleted {
		currentDefault := ""
		if m.timerDefault == TIMER_1 {
			duration := m.serverConfig.Timer.Phase1Duration_Timer1 + m.serverConfig.Timer.Phase2Duration_Timer1 + m.serverConfig.Timer.Phase3Duration_Timer1
			currentDefault = fmt.Sprintf("Timer 1 (%dm)", duration)
		}
		if m.timerDefault == TIMER_2 {
			duration := m.serverConfig.Timer.Phase1Duration_Timer2 + m.serverConfig.Timer.Phase2Duration_Timer2 + m.serverConfig.Timer.Phase3Duration_Timer2
			currentDefault = fmt.Sprintf("Timer 2 (%dm)", duration)
		}
		line1 := util.CenterText("Press Enter or Space to start default timer, '?' for config", m.width)
		line2 := util.CenterText("'1|2' to start respective timer", m.width)
		line3 := util.CenterText("(d)efault timer: "+currentDefault, m.width)
		return line1 + "\n" + line2 + "\n" + line3, util.GetNormalStyle()
	}

	elapsed := m.timerElapsed
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	timerText := fmt.Sprintf("Timer: %d:%02d (%dm)", minutes, seconds, duration)
	var temp = m.temp
	var style lipgloss.Style
	switch m.currentPhase {
	case phase1:
		style = util.GetGreenStyle()
	case phase2:
		style = util.GetYellowStyle()
	case phase3:
		style = util.GetRedStyle()
	default:
		style = util.GetNormalStyle()
	}
	timerText += fmt.Sprintf(" Temp: %d°", temp)
	line := util.CenterText(timerText, m.width)
	return line, style
}

func fetchConfigCmd(serverURL string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/config", serverURL), nil)
		req.Header.Set("X-API-KEY", apiKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return configLoadedMsg{}
		}
		defer resp.Body.Close()
		var cfg config.Config
		json.NewDecoder(resp.Body).Decode(&cfg)
		return configLoadedMsg{cfg: cfg}
	}
}

func patchConfigCmd(serverURL string, key string, val int, apiKey string) tea.Cmd {
	return func() tea.Msg {
		body, _ := json.Marshal(map[string]int{key: val})
		req, _ := http.NewRequest(http.MethodPatch, serverURL+"/config", bytes.NewReader(body))
		req.Header.Set("X-API-KEY", apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		var cfg config.Config
		json.NewDecoder(resp.Body).Decode(&cfg)
		return configLoadedMsg{cfg: cfg}
	}
}
func fieldToJSONKey(field configField) string {
	switch field {
	case fieldPhase1DurationT1:
		return "phase1_timer1_duration_minutes"
	case fieldPhase2DurationT1:
		return "phase2_timer1_duration_minutes"
	case fieldPhase3DurationT1:
		return "phase3_timer1_duration_minutes"
	case fieldPhase1TempT1:
		return "phase1_timer1_temp"
	case fieldPhase2TempT1:
		return "phase2_timer1_temp"
	case fieldPhase3TempT1:
		return "phase3_timer1_temp"
	case fieldPhase1DurationT2:
		return "phase1_timer2_duration_minutes"
	case fieldPhase2DurationT2:
		return "phase2_timer2_duration_minutes"
	case fieldPhase3DurationT2:
		return "phase3_timer2_duration_minutes"
	case fieldPhase1TempT2:
		return "phase1_timer2_temp"
	case fieldPhase2TempT2:
		return "phase2_timer2_temp"
	case fieldPhase3TempT2:
		return "phase3_timer2_temp"
	}
	return ""
}
