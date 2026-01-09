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
	fieldPhase1DurationT1 configField = iota
	fieldPhase2DurationT1
	fieldPhase3DurationT1
	fieldPhase1TempT1
	fieldPhase2TempT1
	fieldPhase3TempT1
	fieldPhase1DurationT2
	fieldPhase2DurationT2
	fieldPhase3DurationT2
	fieldPhase1TempT2
	fieldPhase2TempT2
	fieldPhase3TempT2
	fieldMax
)

type tickMsg time.Time
type dingMsg struct{}
type fileClickMsg struct{}
type fileClickMsg2 struct{}

type TimerOutput struct {
	Text  string `json:"text"`
	Class string `json:"class"`
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second / 10, func(t time.Time) tea.Msg {
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