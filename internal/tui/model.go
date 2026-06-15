package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vibelog/vibelog/internal/export"
	"github.com/vibelog/vibelog/internal/store"
	"github.com/vibelog/vibelog/pkg/types"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0"))

	eventTypeStyle = lipgloss.NewStyle().
			Bold(true).
			Width(12)

	promptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6"))
	responseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	decisionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	fileStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
)

type sessionItem struct {
	session types.Session
}

func (i sessionItem) FilterValue() string { return i.session.RepoPath + " " + i.session.Branch }
func (i sessionItem) Title() string      { return fmt.Sprintf("%s (%s)", i.session.Branch, i.session.RepoPath) }
func (i sessionItem) Description() string {
	status := "active"
	if i.session.EndedAt != nil {
		status = "ended " + i.session.EndedAt.Format("15:04")
	}
	return fmt.Sprintf("%s • %s • %s", i.session.StartedAt.Format("Jan 02 15:04"), status, i.session.ID[:8])
}

type Model struct {
	store       *store.Store
	mode        string // list, detail, search
	sessions    list.Model
	events      []types.Event
	viewport    viewport.Model
	detailID    string
	searchQuery string
	width       int
	height      int
	err         error
}

func NewModel(s *store.Store) (*Model, error) {
	sessions, err := s.GetSessions()
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, len(sessions))
	for i, sess := range sessions {
		items[i] = sessionItem{session: sess}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "VibeLog Sessions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return &Model{
		store:    s,
		mode:     "list",
		sessions: l,
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.sessions.SetSize(msg.Width, msg.Height-2)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.mode == "detail" {
				m.mode = "list"
				return m, nil
			}
			return m, tea.Quit

		case "enter":
			if m.mode == "list" {
				if item, ok := m.sessions.SelectedItem().(sessionItem); ok {
					m.detailID = item.session.ID
					m.loadDetail(item.session.ID)
					m.mode = "detail"
				}
			}
			return m, nil

		case "e":
			if m.mode == "detail" && m.detailID != "" {
				m.exportSession()
			}
			return m, nil

		case "/":
			if m.mode == "list" {
				m.sessions.SetFilterState(list.Filtering)
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.mode == "list" {
		m.sessions, cmd = m.sessions.Update(msg)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.mode == "list" {
		return m.sessions.View()
	}

	if m.mode == "detail" {
		header := titleStyle.Render(" Session Detail ") + "\n"
		footer := statusStyle.Render("\n↑/↓ scroll • e: export to markdown • q: back") + "\n"
		return header + m.viewport.View() + footer
	}

	return "Loading..."
}

func (m *Model) loadDetail(sessionID string) {
	events, err := m.store.GetEvents(sessionID)
	if err != nil {
		m.err = err
		return
	}
	m.events = events

	var content strings.Builder
	session, _ := m.store.GetSession(sessionID)
	if session != nil {
		content.WriteString(fmt.Sprintf("ID: %s\n", session.ID))
		content.WriteString(fmt.Sprintf("Branch: %s\n", session.Branch))
		content.WriteString(fmt.Sprintf("Started: %s\n\n", session.StartedAt.Format(time.RFC1123)))
	}

	for _, ev := range events {
		style := eventTypeStyle
		switch ev.Type {
		case "prompt":
			style = style.Inherit(promptStyle)
		case "response":
			style = style.Inherit(responseStyle)
		case "decision":
			style = style.Inherit(decisionStyle)
		case "file_change":
			style = style.Inherit(fileStyle)
		case "error":
			style = style.Inherit(errorStyle)
		}

		content.WriteString(style.Render(ev.Type) + " ")
		content.WriteString(fmt.Sprintf("%s\n", ev.CreatedAt.Format("15:04:05")))
		content.WriteString(wrap(ev.Content, m.width-4) + "\n\n")
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoTop()
}

func (m *Model) exportSession() {
	if m.detailID == "" {
		return
	}
	session, err := m.store.GetSession(m.detailID)
	if err != nil {
		return
	}
	events, err := m.store.GetEvents(m.detailID)
	if err != nil {
		return
	}

	md := export.SessionToMarkdown(session, events)
	filename := fmt.Sprintf("vibelog-%s.md", m.detailID[:8])
	os.WriteFile(filename, []byte(md), 0644)
}

func wrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var lines []string
	var current string
	words := strings.Fields(s)
	for _, word := range words {
		if len(current)+len(word)+1 > width {
			lines = append(lines, current)
			current = word
		} else {
			if current != "" {
				current += " "
			}
			current += word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}
