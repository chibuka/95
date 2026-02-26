package picker

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/banner"
)

// Option represents a predefined run command.
type Option struct {
	Label       string
	Command     string
	Description string
}

// GoOptions are the predefined run commands for Go projects.
var GoOptions = []Option{
	{
		Label:       "go run .",
		Command:     "go run .",
		Description: "run current package",
	},
	{
		Label:       "go run main.go",
		Command:     "go run main.go",
		Description: "explicit entry point",
	},
	{
		Label:       "go build -o ./app && ./app",
		Command:     "go build -o ./app && ./app",
		Description: "compile, then run",
	},
}

var (
	orange = lipgloss.Color("208")
	muted  = lipgloss.Color("240")
	dim    = lipgloss.Color("244")

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	activeLabel = lipgloss.NewStyle().
			Foreground(orange).
			Bold(true)

	activeDesc = lipgloss.NewStyle().
			Foreground(dim)

	inactiveLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	inactiveDesc = lipgloss.NewStyle().
			Foreground(muted)

	hintStyle = lipgloss.NewStyle().
			Foreground(muted)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("235"))
)

// Model is the bubbletea picker model.
type Model struct {
	options  []Option
	cursor   int
	chosen   *Option
	quitting bool
}

// New returns a picker for the given options.
func New(options []Option) Model {
	return Model{options: options}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			opt := m.options[m.cursor]
			m.chosen = &opt
			return m, tea.Quit
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	divider := dividerStyle.Render(strings.Repeat("─", 48))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(banner.Render())
	b.WriteString("\n\n")
	b.WriteString("  " + divider + "\n")
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("  Go — pick your run command"))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		var line string
		if i == m.cursor {
			bullet := activeLabel.Render("●")
			label := activeLabel.Render(opt.Label)
			desc := activeDesc.Render(" — " + opt.Description)
			line = fmt.Sprintf("  %s  %s%s", bullet, label, desc)
		} else {
			bullet := inactiveLabel.Render("○")
			label := inactiveLabel.Render(opt.Label)
			desc := inactiveDesc.Render(" — " + opt.Description)
			line = fmt.Sprintf("  %s  %s%s", bullet, label, desc)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString("  " + divider + "\n")
	b.WriteString(hintStyle.Render("  ↑↓ navigate   enter select   q quit"))
	b.WriteString("\n")

	return b.String()
}

// Chosen returns the selected option, or nil if cancelled.
func (m Model) Chosen() *Option {
	return m.chosen
}
