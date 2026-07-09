package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pickerItem struct {
	label string
	url   string
}

var (
	pickerDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	pickerErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	pickerCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("> ")
	pickerOpened = pickerDimStyle.Render(" (opened)")
	pickerFooter = pickerDimStyle.Render("o: open • enter: open & exit • esc: quit")
)

// openResultMsg reports the outcome of opening a URL in the background,
// keeping the exec out of the event loop (xdg-open can block for seconds).
type openResultMsg struct {
	idx int
	err error
}

type pickerModel struct {
	title  string
	items  []pickerItem
	cursor int
	opened []bool
	err    error
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case openResultMsg:
		m.err = msg.err
		if msg.err == nil {
			m.opened[msg.idx] = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "o":
			i := m.cursor
			return m, func() tea.Msg {
				return openResultMsg{idx: i, err: openURL(m.items[i].url)}
			}
		case "enter":
			m.err = openURL(m.items[m.cursor].url)
			return m, tea.Quit
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m pickerModel) View() string {
	var b strings.Builder
	b.WriteString(m.title + "\n\n")
	for i, item := range m.items {
		if i == m.cursor {
			b.WriteString(pickerCursor)
		} else {
			b.WriteString("  ")
		}
		b.WriteString(item.label)
		if m.opened[i] {
			b.WriteString(pickerOpened)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(pickerErrStyle.Render("Error: "+m.err.Error()) + "\n")
	}
	b.WriteString(pickerFooter + "\n")
	return b.String()
}

func runPicker(title string, items []pickerItem) error {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		fmt.Println("Opening:", items[0].label)
		return openURL(items[0].url)
	}

	final, err := tea.NewProgram(pickerModel{
		title:  lipgloss.NewStyle().Bold(true).Render(title),
		items:  items,
		opened: make([]bool, len(items)),
	}).Run()
	if err != nil {
		return fmt.Errorf("running picker: %w", err)
	}
	return final.(pickerModel).err
}
