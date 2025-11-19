package utilities

import (
	"fmt"
	"os/exec"
	"runtime"
)

type TimerPhase int

const (
	phaseNotStarted TimerPhase = iota
	phase1
	phase2
	phase3
	phaseCompleted
)

func SendNotification(phase TimerPhase, temp int) {
	var title, body string

	switch phase {
	case phase1:
		title = "Phase 1"
	case phase2:
		title = "Phase 2"
	case phase3:
		title = "Phase 3"
	case phaseCompleted:
		title = "Timer Complete"
		body = "All phases finished!"
	default:
		return
	}

	if phase != phaseCompleted {
		body = fmt.Sprintf("%d°", temp)
	}

	switch runtime.GOOS {
	case "linux":
		// Use notify-send for desktop notifications
		exec.Command("notify-send", "-u", "normal", "-t", "5000", title, body).Run()
	case "darwin":
		// macOS - use osascript to display notification
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, body, title)
		exec.Command("osascript", "-e", script).Run()
	case "windows":
		// Windows - use PowerShell to show toast notification
		script := fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; $Template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $RawXml = [xml] $Template.GetXml(); ($RawXml.toast.visual.binding.text|where {$_.id -eq "1"}).AppendChild($RawXml.CreateTextNode("%s")) > $null; ($RawXml.toast.visual.binding.text|where {$_.id -eq "2"}).AppendChild($RawXml.CreateTextNode("%s")) > $null; $SerializedXml = New-Object Windows.Data.Xml.Dom.XmlDocument; $SerializedXml.LoadXml($RawXml.OuterXml); $Toast = [Windows.UI.Notifications.ToastNotification]::new($SerializedXml); $Toast.Tag = "ChillClock"; $Toast.Group = "ChillClock"; $Notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("ChillClock"); $Notifier.Show($Toast);`, title, body)
		exec.Command("powershell", "-Command", script).Run()
	}
}

func PlayBeep() {
	// Play a system beep sound
	switch runtime.GOOS {
	case "linux":
		// Try paplay (PulseAudio) first, fall back to speaker-test
		cmd := exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
		if err := cmd.Run(); err != nil {
			// Fallback to beep command or speaker-test
			exec.Command("speaker-test", "-t", "sine", "-f", "1000", "-l", "1").Run()
		}
	case "darwin":
		// macOS
		exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	case "windows":
		// Windows - use rundll32 to play system sound
		exec.Command("rundll32", "user32.dll,MessageBeep").Run()
	default:
		// Fallback: print bell character
		fmt.Print("\a")
	}
}
