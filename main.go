package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
	promptStyle   = lipgloss.NewStyle().Bold(true).MarginTop(1)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	correctStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	wrongStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	explainStyle  = lipgloss.NewStyle().Faint(true).MarginTop(1)
)

// model is the Bubble Tea state for a training session.
type model struct {
	questions []Question
	index     int  // current question
	cursor    int  // highlighted choice
	chosen    int  // selected choice once answered, else -1
	score     int
	answered  int
	quitting  bool
}

func newModel(qs []Question) model {
	return model{questions: qs, chosen: -1}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.chosen == -1 && m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.chosen == -1 && m.cursor < len(m.current().Choices)-1 {
			m.cursor++
		}
	case "enter", " ":
		if m.chosen == -1 {
			m.answer()
		} else {
			m.next()
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return fmt.Sprintf("\nSession over — %d/%d correct. See you at the gym.\n", m.score, m.answered)
	}

	q := m.current()
	var b strings.Builder

	b.WriteString(titleStyle.Render("Brain Gym — system design trainer"))
	b.WriteString(fmt.Sprintf("   score %d/%d\n", m.score, m.answered))
	b.WriteString(promptStyle.Render(q.Prompt) + "\n\n")

	correct := q.correctIndex()
	for i, c := range q.Choices {
		line := fmt.Sprintf("%d) %s", i+1, c.Label)
		switch {
		case m.chosen != -1 && i == correct:
			b.WriteString(correctStyle.Render("✓ "+line) + "\n")
		case m.chosen != -1 && i == m.chosen:
			b.WriteString(wrongStyle.Render("✗ "+line) + "\n")
		case m.chosen == -1 && i == m.cursor:
			b.WriteString(cursorStyle.Render("> ") + selectedStyle.Render(line) + "\n")
		default:
			b.WriteString("  " + line + "\n")
		}
	}

	if m.chosen != -1 {
		b.WriteString(explainStyle.Render(q.Explanation) + "\n")
		b.WriteString("\n" + helpStyle.Render("enter: next • q: quit"))
	} else {
		b.WriteString("\n" + helpStyle.Render("↑/↓: move • enter: answer • q: quit"))
	}
	return b.String()
}

func (m model) current() Question { return m.questions[m.index] }

// answer locks in the highlighted choice and scores it.
func (m *model) answer() {
	m.chosen = m.cursor
	m.answered++
	if m.questions[m.index].Choices[m.chosen].Correct {
		m.score++
	}
}

// next advances to the following question, wrapping around the pool.
func (m *model) next() {
	m.index = (m.index + 1) % len(m.questions)
	m.cursor = 0
	m.chosen = -1
}

func main() {
	if _, err := tea.NewProgram(newModel(questions)).Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
