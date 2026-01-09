package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"prbox/github"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss/v2"
)

type screen struct {
	width  int
	height int
}

type model struct {
	githubCliFound        bool
	githubCliPath         string
	screen                screen
	loadingNotifications  bool
	notifications         []string
	notificationsErr      error
}

type githubCliPath (string)
type githubCliPathError struct{ err error }

type notificationsLoaded struct{ titles []string }
type notificationsError struct{ err error }

func FetchNotifications(ghPath string) tea.Cmd {
	return func() tea.Msg {
		client := newCLIClient(ghPath)
		resp, err := github.Testing(context.Background(), client)
		if err != nil {
			return notificationsError{err}
		}

		var titles []string
		for _, edge := range resp.Viewer.NotificationThreads.Edges {
			titles = append(titles, edge.Node.Title)
		}
		return notificationsLoaded{titles}
	}
}

func CheckIfGithubCliInstalled() tea.Cmd {
	return func() tea.Msg {
		path, err := exec.LookPath("gh")
		if err != nil {
			return githubCliPathError{err}
		} else {
			return githubCliPath(path)
		}
	}
}

func initModel() model {
	return model{
		screen: screen{width: 50, height: 50},
	}
}

func (m model) Init() tea.Cmd {
	return CheckIfGithubCliInstalled()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screen.width = msg.Width
		m.screen.height = msg.Height
	case githubCliPath:
		m.githubCliPath = string(msg)
		m.githubCliFound = true
		m.loadingNotifications = true
		return m, FetchNotifications(m.githubCliPath)
	case githubCliPathError:
		m.githubCliFound = false
	case notificationsLoaded:
		m.loadingNotifications = false
		m.notifications = msg.titles
	case notificationsError:
		m.loadingNotifications = false
		m.notificationsErr = msg.err
	case tea.KeyMsg:
		// Cool what was the actual key pressed?
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
			// case "up", "k":
			// 	if m.cursor > 0 {
			// 		m.cursor--
			// 	}
			// case "down", "j":
			// 	if m.cursor < len(m.choices)-1 {
			// 		m.cursor++
			// 	}
			// case "enter", " ":
			// 	_, ok := m.selected[m.cursor]
			// 	if ok {
			// 		delete(m.selected, m.cursor)
			// 	} else {
			// 		m.selected[m.cursor] = struct{}{}
			// 	}
		}
	}

	return m, nil
}

var style = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder())

var boxStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center)
var errorTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Red).Border(lipgloss.DoubleBorder())

func FullScreenBox(s screen) lipgloss.Style {
	return boxStyle.Width(s.width).Height(s.height)
}

func ShowErrorMessage(m model) string {
	s := errorTitle.Render("It looks like you don't have the GitHub CLI installed.")

	return FullScreenBox(m.screen).Render(s)
}

func (m model) View() string {
	if !m.githubCliFound {
		return ShowErrorMessage(m)
	}

	var s string

	if m.loadingNotifications {
		s = "Loading notifications..."
	} else if m.notificationsErr != nil {
		s = fmt.Sprintf("Error loading notifications: %v", m.notificationsErr)
	} else if len(m.notifications) == 0 {
		s = "No notifications"
	} else {
		s = strings.Join(m.notifications, "\n")
	}

	return FullScreenBox(m.screen).Render(s)
}

func main() {
	p := tea.NewProgram(initModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's ben an error: %v", err)
		os.Exit(1)
	}
}
