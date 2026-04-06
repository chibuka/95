package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chibuka/95-cli/ui/banner"
)

// Action describes what the user selected before the dashboard exited.
type Action struct {
	Command string
	UUID    string // non-empty for "test" and "run"
	Quit    bool
}

type dashState int

const (
	stateMenu   dashState = iota
	stateInput            // UUID prompt for "test" / "run"
	stateResult           // show result/error from previous command
)

type entry struct {
	name      string
	desc      string
	needsUUID bool
}

var allEntries = []entry{
	{"init", "initialize project run command", false},
	{"test", "run tests locally", true},
	{"run", "submit your solution", true},
	{"logout", "sign out", false},
}

var unauthEntries = []entry{
	{"login", "authenticate with GitHub", false},
}

var (
	orange = lipgloss.Color("208")
	muted  = lipgloss.Color("240")
	white  = lipgloss.Color("255")
	dim    = lipgloss.Color("244")

	activeName   = lipgloss.NewStyle().Foreground(orange).Bold(true)
	inactiveName = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	descStyle    = lipgloss.NewStyle().Foreground(dim)
	mutedStyle   = lipgloss.NewStyle().Foreground(muted)
	whiteStyle   = lipgloss.NewStyle().Foreground(white)
	hintStyle    = lipgloss.NewStyle().Foreground(muted)
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
	errStyle     = lipgloss.NewStyle().Foreground(orange)
)

// Model is the interactive dashboard bubbletea model.
type Model struct {
	state     dashState
	cursor    int
	input     textinput.Model
	username  string
	dir       string
	action    *Action
	resultMsg string
}

// New creates a fresh dashboard model with user context.
// lastErr is an optional error message from the previous command; if non-empty
// the dashboard opens in stateResult so the user can read it and press esc to retry.
func New(username, dir, lastErr string) Model {
	ti := textinput.New()
	ti.Placeholder = "paste stage UUID here..."
	ti.CharLimit = 64
	ti.Width = 44
	state := stateMenu
	if lastErr != "" {
		state = stateResult
	}
	return Model{username: username, dir: dir, input: ti, state: state, resultMsg: lastErr}
}

func (m Model) visibleEntries() []entry {
	if m.username == "not authenticated" {
		return unauthEntries
	}
	return allEntries
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateResult:
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		switch key.String() {
		case "ctrl+c":
			m.action = &Action{Quit: true}
			return m, tea.Quit
		default:
			m.state = stateMenu
			m.resultMsg = ""
		}
		return m, nil

	case stateMenu:
		key, ok := msg.(tea.KeyMsg)
		if !ok {
			return m, nil
		}
		visible := m.visibleEntries()
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(visible)-1 {
				m.cursor++
			}
		case "enter":
			e := visible[m.cursor]
			if e.needsUUID {
				m.state = stateInput
				m.input.Focus()
				return m, textinput.Blink
			}
			m.action = &Action{Command: e.name}
			return m, tea.Quit
		case "ctrl+c", "q":
			m.action = &Action{Quit: true}
			return m, tea.Quit
		}

	case stateInput:
		key, ok := msg.(tea.KeyMsg)
		if ok {
			switch key.String() {
			case "enter":
				uuid := strings.TrimSpace(m.input.Value())
				if uuid != "" {
					m.action = &Action{Command: m.visibleEntries()[m.cursor].name, UUID: uuid}
					return m, tea.Quit
				}
			case "esc":
				m.state = stateMenu
				m.input.Blur()
				m.input.SetValue("")
				return m, nil
			case "ctrl+c":
				m.action = &Action{Quit: true}
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	divider := dividerStyle.Render(strings.Repeat("─", 48))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(banner.Render())
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s                              %s\n",
		mutedStyle.Render("95™"),
		mutedStyle.Render("v"+banner.Version)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  " + m.dir))
	b.WriteString("\n\n")
	b.WriteString("  " + whiteStyle.Render("◇ "+m.username))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render("  Commands"))
	b.WriteString("\n\n")

	for i, e := range m.visibleEntries() {
		var bullet, name, desc string
		if i == m.cursor {
			bullet = activeName.Render("●")
			name = activeName.Render(fmt.Sprintf("%-8s", e.name))
			desc = descStyle.Render(" — " + e.desc)
		} else {
			bullet = inactiveName.Render("○")
			name = inactiveName.Render(fmt.Sprintf("%-8s", e.name))
			desc = descStyle.Render(" — " + e.desc)
		}
		b.WriteString(fmt.Sprintf("  %s  %s%s\n", bullet, name, desc))
	}

	b.WriteString("\n")
	b.WriteString("  " + divider + "\n")

	switch m.state {
	case stateResult:
		b.WriteString("  " + errStyle.Render("✗  "+m.resultMsg) + "\n\n")
		b.WriteString(hintStyle.Render("  any key to retry   ctrl+c quit"))
	case stateMenu:
		b.WriteString(hintStyle.Render("  ↑↓ navigate   enter select   q quit"))
	case stateInput:
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			mutedStyle.Render("UUID:"),
			m.input.View()))
		b.WriteString(hintStyle.Render("  enter confirm   esc back"))
	}

	b.WriteString("\n")
	return b.String()
}

// Result returns the selected action, or nil if the program exited abnormally.
func (m Model) Result() *Action {
	return m.action
}
