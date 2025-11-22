package utilities

import (
	"github.com/charmbracelet/lipgloss"
)

func GetEditingStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
}

func GetGreenStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
}

func GetYellowStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
}

func GetRedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
}

func GetNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
}