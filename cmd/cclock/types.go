package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	util "github.com/unquenchedservant/ChillClock/utilities"
)

type timerPhase int

const (
	phaseNotStarted timerPhase = iota
	phase1
	phase2
	phase3
	phaseCompleted
)

type viewMode int

const (
	viewClock viewMode = iota
	viewConfig
)

type configField int

const (
	fieldPhase1Duration configField = iota
	fieldPhase2Duration
	fieldPhase3Duration
	fieldPhase1Temp
	fieldPhase2Temp
	fieldPhase3Temp
	fieldMax
)

type tickMsg time.Time
type dingMsg struct{}
type fileClickMsg struct{}

type TimerOutput struct {
	Text  string `json:"text"`
	Class string `json:"class"`
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func dingCmd(phase timerPhase, temp int) tea.Cmd {
	return func() tea.Msg {
		util.PlayBeep()
		util.SendNotification(util.TimerPhase(phase), temp)
		return dingMsg{}
	}
}