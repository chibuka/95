package picker

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option represents a predefined run command.
type Option struct {
	Label       string
	Command     string
	Description string
}

type language struct {
	name    string
	display string
	options []Option
}

var languages = []language{
	{
		name: "go", display: "Go",
		options: []Option{
			{"go run .", "go run .", "run current package"},
			{"go run main.go", "go run main.go", "explicit entry point"},
			{"go build -o ./app && ./app", "go build -o ./app && ./app", "compile, then run"},
		},
	},
	{
		name: "python", display: "Python",
		options: []Option{
			{"python3 main.py", "python3 main.py", "run with python3"},
			{"python main.py", "python main.py", "run with python"},
		},
	},
	{
		name: "rust", display: "Rust",
		options: []Option{
			// Command is stored for local reference; the runner always uses a
			// canonical Cargo release build and run on the server.
			{"Cargo project", "cargo run", "Cargo.toml + src/main.rs; platform runs release build & run"},
		},
	},
	{
		name: "java", display: "Java",
		options: []Option{
			{"javac Main.java && java Main", "javac Main.java && java Main", "compile, then run"},
			{"java Main.java", "java Main.java", "single-file source (Java 11+)"},
		},
	},
	{
		name: "javascript", display: "JavaScript",
		options: []Option{
			{"node index.js", "node index.js", "run with node"},
			{"node main.js", "node main.js", "explicit entry point"},
		},
	},
	{
		name: "typescript", display: "TypeScript",
		options: []Option{
			{"ts-node main.ts", "ts-node main.ts", "run with ts-node"},
			{"npx ts-node main.ts", "npx ts-node main.ts", "run with npx ts-node"},
		},
	},
	{
		name: "c", display: "C",
		options: []Option{
			{"make && ./app", "make && chmod +x app && ./app", "build with make, then run ./app"},
			{"gcc main.c -o app && ./app", "gcc main.c -o app && ./app", "compile with gcc, then run"},
			{"clang main.c -o app && ./app", "clang main.c -o app && ./app", "compile with clang, then run"},
		},
	},
	{
		name: "cpp", display: "C++",
		options: []Option{
			{"make && ./app", "make && chmod +x app && ./app", "build with make, then run ./app"},
			{"g++ main.cpp -o app && ./app", "g++ main.cpp -o app && ./app", "compile with g++, then run"},
			{"clang++ main.cpp -o app && ./app", "clang++ main.cpp -o app && ./app", "compile with clang++, then run"},
		},
	},
	{
		name: "ruby", display: "Ruby",
		options: []Option{
			{"ruby main.rb", "ruby main.rb", "run with ruby"},
		},
	},
	{
		name: "elixir", display: "Elixir",
		options: []Option{
			{"elixir main.exs", "elixir main.exs", "run script with elixir"},
			{"mix run", "mix run", "run with mix"},
		},
	},
	{
		name: "haskell", display: "Haskell",
		options: []Option{
			{"runhaskell main.hs", "runhaskell main.hs", "interpret with runhaskell"},
			{"ghc main.hs -o app && ./app", "ghc main.hs -o app && ./app", "compile with ghc, then run"},
		},
	},
	{
		name: "zig", display: "Zig",
		options: []Option{
			{"zig run main.zig", "zig run main.zig", "run with zig"},
			{"zig build run", "zig build run", "build and run"},
		},
	},
	{
		name: "kotlin", display: "Kotlin",
		options: []Option{
			{"kotlinc main.kt -include-runtime -d app.jar && java -jar app.jar", "kotlinc main.kt -include-runtime -d app.jar && java -jar app.jar", "compile, then run"},
			{"kotlin main.kt", "kotlin main.kt", "run as script"},
		},
	},
	{
		name: "swift", display: "Swift",
		options: []Option{
			{"swift main.swift", "swift main.swift", "run with swift"},
			{"swift run", "swift run", "run with swift package manager"},
		},
	},
}

var (
	orange = lipgloss.Color("208")
	muted  = lipgloss.Color("240")
	dim    = lipgloss.Color("244")

	subtitleStyle = lipgloss.NewStyle().Foreground(muted)
	infoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	activeLabel = lipgloss.NewStyle().Foreground(orange).Bold(true)
	activeDesc  = lipgloss.NewStyle().Foreground(dim)

	inactiveLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	inactiveDesc  = lipgloss.NewStyle().Foreground(muted)

	hintStyle    = lipgloss.NewStyle().Foreground(muted)
	dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))
)

type pickerState int

const (
	stateLang pickerState = iota
	stateCmd
)

// Model is the bubbletea picker model.
type Model struct {
	state      pickerState
	langCursor int
	langIdx    int
	cmdCursor  int
	chosen     *Option
	quitting   bool
}

// New returns a fresh picker starting at language selection.
// If enabled is non-nil, only languages whose name appears in the map are shown.
func New(enabled map[string]string) Model {
	if enabled != nil {
		var filtered []language
		for _, lang := range languages {
			if ver, ok := enabled[lang.name]; ok {
				lang.display = fmt.Sprintf("%s (v%s)", lang.display, ver)
				filtered = append(filtered, lang)
			}
		}
		languages = filtered
	}
	return Model{}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.state {
	case stateLang:
		switch key.String() {
		case "up", "k":
			if m.langCursor > 0 {
				m.langCursor--
			}
		case "down", "j":
			if m.langCursor < len(languages)-1 {
				m.langCursor++
			}
		case "enter", " ":
			m.langIdx = m.langCursor
			m.state = stateCmd
			m.cmdCursor = 0
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case stateCmd:
		opts := languages[m.langIdx].options
		switch key.String() {
		case "up", "k":
			if m.cmdCursor > 0 {
				m.cmdCursor--
			}
		case "down", "j":
			if m.cmdCursor < len(opts)-1 {
				m.cmdCursor++
			}
		case "enter", " ":
			opt := opts[m.cmdCursor]
			m.chosen = &opt
			return m, tea.Quit
		case "esc":
			m.state = stateLang
			m.cmdCursor = 0
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
	b.WriteString("  " + divider + "\n")
	b.WriteString("\n")

	switch m.state {
	case stateLang:
		b.WriteString(subtitleStyle.Render("  Pick a language"))
		b.WriteString("\n\n")

		for i, lang := range languages {
			var line string
			if i == m.langCursor {
				bullet := activeLabel.Render("●")
				label := activeLabel.Render(fmt.Sprintf("%-12s", lang.name))
				desc := activeDesc.Render(" — " + lang.display)
				line = fmt.Sprintf("  %s  %s%s", bullet, label, desc)
			} else {
				bullet := inactiveLabel.Render("○")
				label := inactiveLabel.Render(fmt.Sprintf("%-12s", lang.name))
				desc := inactiveDesc.Render(" — " + lang.display)
				line = fmt.Sprintf("  %s  %s%s", bullet, label, desc)
			}
			b.WriteString(line + "\n")
		}

		b.WriteString("\n")
		b.WriteString("  " + divider + "\n")
		b.WriteString(hintStyle.Render("  ↑↓ navigate   enter select   q quit"))

	case stateCmd:
		lang := languages[m.langIdx]
		b.WriteString(subtitleStyle.Render("  " + lang.display + " — pick your run command"))
		b.WriteString("\n\n")

		for i, opt := range lang.options {
			var line string
			if i == m.cmdCursor {
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
		b.WriteString(infoStyle.Render("  Saved to project config; the server may use language-specific build & run rules."))
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("  ↑↓ navigate   enter select   esc back   q quit"))
	}

	b.WriteString("\n")
	return b.String()
}

// Chosen returns the selected option, or nil if cancelled.
func (m Model) Chosen() *Option {
	return m.chosen
}

// Language returns the name of the selected language.
func (m Model) Language() string {
	return languages[m.langIdx].name
}
