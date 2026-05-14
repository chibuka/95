package logview

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type FloorResult struct {
	Number int
	Passed bool
	Logs   string
}

var (
	panelContentWidth = 72

	successColor = lipgloss.Color("10")
	warnColor    = lipgloss.Color("11")
	failColor    = lipgloss.Color("208")
	mutedColor   = lipgloss.Color("240")
	textColor    = lipgloss.Color("252")
	borderColor  = lipgloss.Color("238")

	passStyle    = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	failStyle    = lipgloss.NewStyle().Foreground(failColor).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(warnColor)
	mutedStyle   = lipgloss.NewStyle().Foreground(mutedColor)
	textStyle    = lipgloss.NewStyle().Foreground(textColor)
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	outputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	panelStyle   = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Width(panelContentWidth).
			Padding(0, 1).
			MarginLeft(2)
)

type parsedLogs struct {
	notes            []string
	command          []string
	actualOutput     []string
	expectedVariants []string
	assertion        string
}

var quotedStringPattern = regexp.MustCompile(`"(?:\\.|[^"\\])*"`)

func RenderProgress(message string) string {
	return fmt.Sprintf("  %s  %s", warnStyle.Render("•"), mutedStyle.Render(message))
}

func RenderSummary(passed bool, successHint string, recorded bool) string {
	var b strings.Builder
	if passed {
		b.WriteString("  ")
		b.WriteString(passStyle.Render("✓ All floors passed"))
		if !recorded {
			b.WriteString("\n\n")
			b.WriteString("  ")
			b.WriteString(warnStyle.Render("! Progress could not be saved. Please submit again."))
		}
		if strings.TrimSpace(successHint) != "" {
			b.WriteString("\n")
			b.WriteString("  ")
			b.WriteString(mutedStyle.Render(successHint))
		}
		return b.String()
	}

	b.WriteString("  ")
	b.WriteString(failStyle.Render("✗ Some floors failed"))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(mutedStyle.Render("Read the first failed floor from top to bottom: context, command, actual output, expected output."))
	return b.String()
}

func RenderFloor(result FloorResult) string {
	status := passStyle.Render("✓")
	if !result.Passed {
		status = failStyle.Render("✗")
	}

	header := fmt.Sprintf("  %s  Floor %d", status, result.Number)
	if result.Passed || strings.TrimSpace(result.Logs) == "" {
		return header
	}

	parsed := parseLogs(result.Logs)
	body := renderFailureBody(parsed)
	if body == "" {
		body = renderRawBlock("Details", strings.Split(strings.TrimRight(result.Logs, "\n"), "\n"), mutedStyle)
	}

	return header + "\n" + panelStyle.Render(body)
}

func renderFailureBody(logs parsedLogs) string {
	var sections []string
	if len(logs.notes) > 0 {
		sections = append(sections, renderRawBlock("What ran", logs.notes, textStyle))
	}
	if len(logs.command) > 0 {
		sections = append(sections, renderRawBlock("Command", logs.command, commandStyle))
	}
	if len(logs.actualOutput) > 0 {
		sections = append(sections, renderRawBlock("Your stdout", logs.actualOutput, outputStyle))
	}
	if len(logs.expectedVariants) > 0 {
		sections = append(sections, renderExpectedVariants(logs.expectedVariants))
	}
	if strings.TrimSpace(logs.assertion) != "" && len(logs.expectedVariants) == 0 {
		sections = append(sections, renderRawBlock("Assertion", []string{logs.assertion}, failStyle))
	}
	return strings.Join(sections, "\n\n")
}

func renderRawBlock(title string, lines []string, style lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(labelStyle.Render(title))
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render("│ "))
			b.WriteString(mutedStyle.Render("∅"))
			continue
		}
		for _, wrapped := range wrapLine(line, panelContentWidth-2) {
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render("│ "))
			b.WriteString(style.Render(wrapped))
		}
	}
	return b.String()
}

func renderExpectedVariants(variants []string) string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Expected stdout"))
	for i, variant := range variants {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("│ option %d", i+1)))
		lines := strings.Split(strings.TrimSuffix(variant, "\n"), "\n")
		if len(lines) == 1 && lines[0] == "" {
			b.WriteString(" ")
			b.WriteString(mutedStyle.Render("∅"))
			continue
		}
		for _, line := range lines {
			for _, wrapped := range wrapLine(line, panelContentWidth-4) {
				b.WriteString("\n")
				b.WriteString(mutedStyle.Render("│   "))
				b.WriteString(outputStyle.Render(wrapped))
			}
		}
	}
	return b.String()
}

func wrapLine(line string, width int) []string {
	if width <= 0 || len([]rune(line)) <= width {
		return []string{line}
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return []string{line}
	}

	var lines []string
	current := ""
	for _, word := range words {
		if len([]rune(word)) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, splitLongWord(word, width)...)
			continue
		}
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) <= width {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitLongWord(word string, width int) []string {
	runes := []rune(word)
	var parts []string
	for len(runes) > width {
		parts = append(parts, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}
	return parts
}

func parseLogs(raw string) parsedLogs {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	var parsed parsedLogs

	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "$ "):
			parsed.command = append(parsed.command, trimmed)
			if strings.Contains(trimmed, "<<EOF") {
				for i+1 < len(lines) {
					next := strings.TrimRight(lines[i+1], "\r")
					nextTrimmed := strings.TrimSpace(next)
					i++
					parsed.command = append(parsed.command, nextTrimmed)
					if nextTrimmed == "EOF" {
						break
					}
				}
			}

		case trimmed == "Your program output:":
			for i+1 < len(lines) {
				next := strings.TrimRight(lines[i+1], "\r")
				nextTrimmed := strings.TrimSpace(next)
				if isBoundaryLine(nextTrimmed) {
					break
				}
				i++
				parsed.actualOutput = append(parsed.actualOutput, trimTesterIndent(next))
			}

		case strings.HasPrefix(trimmed, "Expected stdout"):
			expected, actual, hasActual, ok := parseExpectedStdout(trimmed)
			if ok {
				parsed.expectedVariants = expected
				if len(parsed.actualOutput) == 0 && hasActual {
					parsed.actualOutput = strings.Split(strings.TrimSuffix(actual, "\n"), "\n")
				}
			}
			parsed.assertion = trimmed

		default:
			parsed.notes = append(parsed.notes, trimmed)
		}
	}

	return parsed
}

func trimTesterIndent(line string) string {
	if strings.HasPrefix(line, "     ") {
		return strings.TrimPrefix(line, "     ")
	}
	if strings.HasPrefix(line, "  ") {
		return strings.TrimPrefix(line, "  ")
	}
	return line
}

func isBoundaryLine(line string) bool {
	return strings.HasPrefix(line, "Expected stdout") ||
		strings.HasPrefix(line, "Test failed") ||
		strings.HasPrefix(line, "$ ")
}

func parseExpectedStdout(line string) ([]string, string, bool, bool) {
	matches := quotedStringPattern.FindAllString(line, -1)
	if len(matches) == 0 {
		return nil, "", false, false
	}

	values := make([]string, 0, len(matches))
	for _, match := range matches {
		value, err := strconv.Unquote(match)
		if err != nil {
			return nil, "", false, false
		}
		values = append(values, value)
	}

	gotIndex := strings.LastIndex(line, ", got ")
	if gotIndex == -1 || len(values) == 1 {
		return values, "", false, true
	}
	return values[:len(values)-1], values[len(values)-1], true, true
}
