package main

import (
	"strings"
	"time"
	util "github.com/unquenchedservant/ChillClock/utilities"
)

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.mode == viewConfig {
		return m.renderConfigView()
	}

	return m.renderClockView()
}

func (m model) renderClockView() string {
	now := time.Now()
	timeStr := now.Format("15:04:05")
	dateStr := now.Format("2006-01-02")

	clockLines := util.RenderLargeText(timeStr)

	var output strings.Builder

	totalLines := 1 + len(clockLines) + 3
	topPadding := (m.height - totalLines) / 2

	for i := 0; i < topPadding; i++ {
		output.WriteString("\n")
	}

	output.WriteString(util.CenterText(util.GetYellowStyle().Render(dateStr), m.width))
	output.WriteString("\n\n")

	for _, line := range clockLines {
		styledLine := util.GetGreenStyle().Render(line)
		output.WriteString(util.CenterText(styledLine, m.width))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	timerText, timerStyle := m.getTimerDisplay()
	output.WriteString(util.CenterText(timerStyle.Render(timerText), m.width))

	return output.String()
}