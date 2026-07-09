package main

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pickerItem struct {
	name   string // display text, plain, already padded (includes "├ " prefix for children)
	status string // raw check status, styled at render time; "" for PR rows
	extra  string // plain dim suffix like "+3"; "" if none
	url    string
	child  bool // sub-stage row, hidden until the picker is expanded
}

var (
	pickerDimColor    = lipgloss.Color("8")
	pickerCursorColor = lipgloss.Color("6")
	// pickerSelBg highlights the selected row; adaptive so the bar stays
	// subtle on both light and dark terminal themes.
	pickerSelBg = lipgloss.AdaptiveColor{Light: "254", Dark: "237"}

	pickerDimStyle = lipgloss.NewStyle().Foreground(pickerDimColor)
	pickerErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	pickerFooter   = pickerDimStyle.Render("o: open • enter: open & exit • esc: quit")
	pickerExpand   = pickerDimStyle.Render(" • a: expand")
	pickerCollapse = pickerDimStyle.Render(" • a: collapse")
)

// openResultMsg reports the outcome of opening a URL in the background,
// keeping the exec out of the event loop (xdg-open can block for seconds).
type openResultMsg struct {
	idx int
	err error
}

type pickerModel struct {
	title   string
	items   []pickerItem
	cursor  int
	opened  []bool
	showAll bool
	err     error
}

// hidden reports whether the item at index i is filtered out of the view.
func (m pickerModel) hidden(i int) bool {
	return m.items[i].child && !m.showAll
}

// moveCursor returns the cursor shifted by delta to the next visible row,
// clamped at either end.
func (m pickerModel) moveCursor(delta int) int {
	for i := m.cursor + delta; i >= 0 && i < len(m.items); i += delta {
		if !m.hidden(i) {
			return i
		}
	}
	return m.cursor
}

func (m pickerModel) hasChildren() bool {
	return slices.ContainsFunc(m.items, func(item pickerItem) bool { return item.child })
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
			m.cursor = m.moveCursor(-1)
		case "down", "j":
			m.cursor = m.moveCursor(1)
		case "a":
			m.showAll = !m.showAll
			for m.cursor > 0 && m.hidden(m.cursor) {
				m.cursor--
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
		if m.hidden(i) {
			continue
		}
		// All segments compose onto base so the selection bar runs unbroken
		// under the cursor, name, badge, and suffixes.
		base := lipgloss.NewStyle()
		if i == m.cursor {
			base = base.Background(pickerSelBg)
			b.WriteString(base.Foreground(pickerCursorColor).Render("> "))
		} else {
			b.WriteString("  ")
		}
		name := base
		if item.child {
			name = name.Foreground(pickerDimColor)
		}
		b.WriteString(name.Render(item.name))
		if item.status != "" {
			b.WriteString(base.Render(" ") + statusBadge(item.status, base))
		}
		if item.extra != "" {
			b.WriteString(base.Foreground(pickerDimColor).Render(" " + item.extra))
		}
		if m.opened[i] {
			b.WriteString(base.Foreground(pickerDimColor).Render(" (opened)"))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(pickerErrStyle.Render("Error: "+m.err.Error()) + "\n")
	}
	b.WriteString(pickerFooter)
	if m.hasChildren() {
		if m.showAll {
			b.WriteString(pickerCollapse)
		} else {
			b.WriteString(pickerExpand)
		}
	}
	b.WriteString("\n")
	return b.String()
}

func runPicker(title string, items []pickerItem) error {
	if len(items) == 0 {
		return nil
	}
	if len(items) == 1 {
		fmt.Println("Opening:", strings.TrimRight(items[0].name, " "))
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
