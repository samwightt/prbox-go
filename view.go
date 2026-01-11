package main

import (
	"fmt"
	"image/color"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

var style = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder())

var boxStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)
var errorTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red).Border(lipgloss.DoubleBorder())

func fullScreenBox(s *screen) lipgloss.Style {
	return boxStyle.Width(s.width).Height(s.height)
}

func showErrorMessage(m *model) string {
	s := errorTitle.Render("It looks like you don't have the GitHub CLI installed.")

	return fullScreenBox(&m.screen).Render(s)
}

var headerStyle = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Height(1)

func layout(screen *screen, content string) string {

	header := headerStyle.Width(screen.width).Border(lipgloss.NormalBorder(), false, false, true, false).Render("Header")
	footer := headerStyle.Width(screen.width).Border(lipgloss.NormalBorder(), true, false, false, false).Render("Footer")

	headerHeight := lipgloss.Height(header)
	footerWidth, footerHeight := lipgloss.Size(footer)
	contentHeight := screen.height - headerHeight - footerHeight
	contentArea := lipgloss.NewStyle().Height(contentHeight).MaxHeight(contentHeight).Width(footerWidth).Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, header, contentArea, footer)
}

var activeStyle = lipgloss.NewStyle().Bold(true)

func activeColor(darkMode bool) color.Color {
	if darkMode {
		return lipgloss.Color("#39BAE6")
		// return lipgloss.Color("#EAF4D3")
	} else {
		return lipgloss.Blue
		// return lipgloss.Color("#3A86FF")
	}
}

func inactiveColor(darkMode bool) color.Color {
	if darkMode {
		return lipgloss.Color("#BFBDB6")
	} else {
		return lipgloss.BrightBlue
		// return lipgloss.Color("#D8CBC7")
	}
}

func notification(n *NotificationThread, index int, m *model) string {
	if m.activeItem == index {
		style := activeStyle.Foreground(lipgloss.NoColor{}).Background(activeColor(m.darkMode))
		return style.Render("> " + n.Node.Title)
	} else {
		style := lipgloss.NewStyle().Foreground(inactiveColor(m.darkMode))
		return style.Render("  " + n.Node.Title)
	}
}

func content(m *model) string {
	if !m.githubCliFound {
		return showErrorMessage(m)
	} else if m.loadingNotifications {
		return "Loading notifications..."
	} else if m.notificationsErr != nil {
		return fmt.Sprintf("Error loading notifications: %v", m.notificationsErr)
	} else if len(m.notifications) == 0 {
		return "No notifications"
	} else {
		var sb strings.Builder
		for i := range m.notifications {
			sb.WriteString(notification(&m.notifications[i], i, m) + "\n")
		}
		return sb.String()
	}
}

func App(m *model) string {
	return layout(&m.screen, content(m))
	// return fullScreenBox(&m.screen).Render(s)
}
