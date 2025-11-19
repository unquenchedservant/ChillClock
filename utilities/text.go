package utilities

import (
	"strings"
	"github.com/charmbracelet/lipgloss"
)

func RenderLargeText(text string) []string {
	var lines [5]strings.Builder

	for _, char := range text {
		digitLines, exists := Digits[char]
		if !exists {
			digitLines = Digits[' ']
		}

		for i, line := range digitLines {
			lines[i].WriteString(line)
		}
	}

	result := make([]string, 5)
	for i, line := range lines {
		result[i] = line.String()
	}

	return result
}

func CenterText(text string, width int) string {
	// Use lipgloss width calculation to handle ANSI codes
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}

	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}