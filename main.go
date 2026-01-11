package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"prbox/github"

	tea "charm.land/bubbletea/v2"
)

type screen struct {
	width  int
	height int
}

type NotificationThread = github.TestingViewerUserNotificationThreadsNotificationThreadConnectionEdgesNotificationThreadEdge

type model struct {
	githubCliFound       bool
	githubCliPath        string
	darkMode             bool
	screen               screen
	activeItem           int
	loadingNotifications bool
	notifications        []NotificationThread
	notificationsErr     error
}

type githubCliPath (string)
type githubCliPathError struct{ err error }

type notificationsLoaded struct{ notifications []NotificationThread }
type notificationsError struct{ err error }

func FetchNotifications(ghPath string) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient(ghPath)
		resp, err := github.Testing(context.Background(), client)
		if err != nil {
			return notificationsError{err}
		}

		return notificationsLoaded{resp.Viewer.NotificationThreads.Edges}
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

func OpenInBrowser(notification NotificationThread) tea.Cmd {
	return func() tea.Msg {
		url := notification.Node.Url
		if notification.Node.OldestUnreadItemAnchor != "" {
			url += "#" + notification.Node.OldestUnreadItemAnchor
		}
		exec.Command("open", url).Start()
		return nil
	}
}

func initModel() model {
	return model{
		screen: screen{width: 50, height: 50},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		CheckIfGithubCliInstalled(),
	)
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
		m.notifications = msg.notifications
	case notificationsError:
		m.loadingNotifications = false
		m.notificationsErr = msg.err
	case tea.BackgroundColorMsg:
		m.darkMode = msg.IsDark()
	case tea.KeyMsg:
		// Cool what was the actual key pressed?
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.activeItem > 0 {
				m.activeItem--
			}
		case "down", "j":
			if m.activeItem < len(m.notifications)-1 {
				m.activeItem++
			}
		case "G":
			m.activeItem = len(m.notifications) - 1
		case "enter":
			if len(m.notifications) > 0 && m.activeItem < len(m.notifications) {
				return m, OpenInBrowser(m.notifications[m.activeItem])
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	view := tea.NewView(App(&m))
	view.AltScreen = true
	return view
}

func main() {
	p := tea.NewProgram(initModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's ben an error: %v", err)
		os.Exit(1)
	}
}
