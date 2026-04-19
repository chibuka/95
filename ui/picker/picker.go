package picker

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option represents a predefined build+run pair for a language.
//
// Why two fields instead of one shell string?
// Some testers (git, http-server) change CWD before invoking your_program.sh.
// The server runs Build inside the submission dir in a subshell, then exec's Run
// from the tester's CWD. Splitting them lets compiled languages still find their
// sources, and lets the run step stay writable to the tester's working directory
// (needed for things like `git init` writing .git there).
//
// Conventions used in the presets below:
//   - Compiled outputs go to /tmp/app (absolute path, always writable in our
//     Alpine runner, cleared between submissions by Docker --rm).
//   - Interpreted runs reference "$_dir/<file>" so they don't break when the
//     tester chdirs. $_dir is defined in the your_program.sh header on the server.
//   - `make` presets assume the user's Makefile emits ./app at CWD — this is
//     documented in the Description. If the user needs something else, they
//     need a custom command (not yet supported).
type Option struct {
	Label       string
	Build       string
	Run         string
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
			{"go run .", "go build -o /tmp/app .", "/tmp/app", "build current package, run binary"},
			{"go run main.go", "go build -o /tmp/app main.go", "/tmp/app", "build main.go, run binary"},
		},
	},
	{
		name: "python", display: "Python",
		options: []Option{
			{"python3 main.py", "", `python3 "$_dir/main.py"`, "run with python3"},
			{"python main.py", "", `python "$_dir/main.py"`, "run with python"},
		},
	},
	{
		name: "rust", display: "Rust",
		options: []Option{
			// --manifest-path keeps cargo rooted at the submission dir even after the
			// tester chdirs; --release matches the cost model of other compiled presets.
			{"cargo build --release", `cargo build --release --manifest-path "$_dir/Cargo.toml"`, `"$_dir/target/release/app"`, "cargo release build (crate must be named 'app')"},
			{"rustc main.rs", `rustc "$_dir/main.rs" -o /tmp/app`, "/tmp/app", "single-file rustc build"},
		},
	},
	{
		name: "javascript", display: "JavaScript",
		options: []Option{
			{"node index.js", "", `node "$_dir/index.js"`, "run with node"},
			{"node main.js", "", `node "$_dir/main.js"`, "run with node"},
		},
	},
	{
		name: "typescript", display: "TypeScript",
		options: []Option{
			{"ts-node main.ts", "", `ts-node "$_dir/main.ts"`, "run with ts-node"},
			{"npx ts-node main.ts", "", `npx ts-node "$_dir/main.ts"`, "run with npx ts-node"},
		},
	},
	{
		name: "c", display: "C",
		options: []Option{
			{"gcc main.c", "gcc main.c -o /tmp/app", "/tmp/app", "compile main.c with gcc"},
			// `make` assumes your Makefile produces ./app — the convention is documented
			// here so users can't miss it. $_dir/app is the absolute location after build.
			// MakefileDocsURL below gets printed alongside this option in the picker view
			// and in `95 init` output so users can self-serve the setup rules.
			{"make", "make", `"$_dir/app"`, "run your Makefile (must produce ./app)"},
		},
	},
	{
		name: "cpp", display: "C++",
		options: []Option{
			{"g++ main.cpp", "g++ main.cpp -o /tmp/app", "/tmp/app", "compile main.cpp with g++"},
			{"make", "make", `"$_dir/app"`, "run your Makefile (must produce ./app)"},
		},
	},
	{
		name: "ruby", display: "Ruby",
		options: []Option{
			{"ruby main.rb", "", `ruby "$_dir/main.rb"`, "run with ruby"},
		},
	},
	{
		name: "elixir", display: "Elixir",
		options: []Option{
			{"elixir main.exs", "", `elixir "$_dir/main.exs"`, "run script with elixir"},
		},
	},
	{
		name: "haskell", display: "Haskell",
		options: []Option{
			{"runhaskell main.hs", "", `runhaskell "$_dir/main.hs"`, "interpret with runhaskell"},
			{"ghc main.hs", `ghc "$_dir/main.hs" -o /tmp/app`, "/tmp/app", "compile with ghc"},
		},
	},
	{
		name: "zig", display: "Zig",
		options: []Option{
			{"zig build-exe", `zig build-exe "$_dir/main.zig" -O ReleaseSafe --cache-dir /tmp/zig-cache -femit-bin=/tmp/app`, "/tmp/app", "build single file, run binary"},
		},
	},
	{
		name: "swift", display: "Swift",
		options: []Option{
			{"swiftc main.swift", `swiftc "$_dir/main.swift" -o /tmp/app`, "/tmp/app", "compile main.swift, run binary"},
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

// MakefileDocsURL is the deep-link anchor on the docs page that explains the
// Makefile convention (must produce ./app, etc). Exported so cmd/init.go can
// print it after a successful init as well, keeping the message consistent.
const MakefileDocsURL = "https://95ninefive.dev/docs#make"

// IsMakePreset is a cheap heuristic — the two make presets share the literal
// build command "make", and no other preset does. Kept as a helper so callers
// don't hardcode the string in multiple places.
func (o Option) IsMakePreset() bool { return o.Build == "make" }

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
		if len(filtered) > 0 {
			languages = filtered
		}
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
		b.WriteString(infoStyle.Render("  The server uses this same command to test your solution."))
		b.WriteString("\n")
		// Makefile setup rules (must produce ./app, tabs not spaces, ...) trip
		// users up often enough that it's worth surfacing the docs URL the moment
		// the cursor lands on a make preset — not only after they hit enter.
		if lang.options[m.cmdCursor].IsMakePreset() {
			b.WriteString(infoStyle.Render("  Makefile setup guide: " + MakefileDocsURL))
			b.WriteString("\n")
		}
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
