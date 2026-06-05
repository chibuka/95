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
	successColor = lipgloss.Color("10")
	warnColor    = lipgloss.Color("11")
	failColor    = lipgloss.Color("208")
	mutedColor   = lipgloss.Color("240")
	textColor    = lipgloss.Color("252")

	passStyle    = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	failStyle    = lipgloss.NewStyle().Foreground(failColor).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(warnColor)
	mutedStyle   = lipgloss.NewStyle().Foreground(mutedColor)
	textStyle    = lipgloss.NewStyle().Foreground(textColor)
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Bold(true)
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	outputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

type parsedLogs struct {
	notes            []string
	command          []string
	actualOutput     []string
	expectedVariants []string
	assertion        string
	comparison       valueComparison
	hasComparison    bool
	caseSummary      string
}

type valueComparison struct {
	Subject  string
	Expected string
	Actual   string
}

var (
	quotedStringPattern = regexp.MustCompile(`"(?:\\.|[^"\\])*"`)
	caseSummaryPattern  = regexp.MustCompile(`^Cases:\s+(\d+)/(\d+)\s+passed$`)
)

const chunkingCasePrefix = "Running chunking case: "

var chunkingCasePassedCount = map[string]int{
	"single document with one sentence":                               0,
	"single document preserves document id":                           1,
	"single document preserves exact text":                            2,
	"single document with multiple sentences":                         0,
	"single document with two sentences":                              1,
	"single document with four sentences":                             2,
	"multiple documents keep their source references":                 0,
	"multiple documents reset chunk indexes":                          1,
	"three documents keep source references":                          2,
	"ignore empty chunks":                                             0,
	"only empty documents produce no chunks":                          1,
	"repeated punctuation does not create empty chunks":               2,
	"preserve original document and chunk order":                      0,
	"preserve document order before chunk index order":                1,
	"preserve duplicate chunk text order":                             2,
	"split punctuation and whitespace variants":                       0,
	"split newline and tab separated sentences":                       1,
	"preserve punctuation terminators":                                2,
	"copy document metadata into every chunk":                         0,
	"metadata does not leak into documents without metadata":          1,
	"metadata values are preserved exactly":                           2,
	"add stable chunk ids and source text indexes":                    0,
	"indexes skip whitespace after sentence boundaries":               1,
	"indexes include trailing chunk without punctuation":              2,
	"enforce chunk_size while preserving metadata and indexes":        0,
	"allow chunks exactly at chunk_size":                              1,
	"small chunk_size keeps short sentences separate":                 2,
	"add chunk_overlap between adjacent chunks":                       0,
	"chunk_overlap uses exact character count":                        1,
	"single chunk does not fabricate overlap":                         2,
	"split by configured separator":                                   0,
	"split by single character separator":                             1,
	"missing configured separator keeps whole document":               2,
	"recursively fall back to smaller separators":                     0,
	"recursively fall back to spaces":                                 1,
	"recursive separators keep fitting chunks together":               2,
	"fallback to character splitting when separators do not fit":      0,
	"character fallback handles exact multiples":                      1,
	"character fallback only splits oversized pieces":                 2,
	"do not overlap across document boundaries":                       0,
	"overlap restarts after a single chunk document":                  1,
	"overlap restarts for each later document":                        2,
	"add markdown header metadata to chunks":                          0,
	"nested markdown headers are inherited and cleared":               1,
	"content before first markdown header has only document metadata": 2,
}

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

	var parsed parsedLogs
	if !result.Passed && strings.TrimSpace(result.Logs) != "" {
		parsed = parseLogs(result.Logs)
	}

	header := fmt.Sprintf("  %s  Floor %d", status, result.Number)
	if parsed.caseSummary != "" {
		header += " · " + parsed.caseSummary
	}
	if result.Passed || strings.TrimSpace(result.Logs) == "" {
		return header
	}

	body := renderFailureBody(parsed)
	if body == "" {
		body = renderRawBlock("Details", strings.Split(strings.TrimRight(result.Logs, "\n"), "\n"), mutedStyle)
	}

	return header + "\n" + indentBlock(body, "    ")
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
	if logs.hasComparison {
		sections = append(sections, renderRawBlock("Expected", comparisonLines(logs.comparison.Subject, logs.comparison.Expected), outputStyle))
		sections = append(sections, renderRawBlock("Actual", comparisonLines(logs.comparison.Subject, logs.comparison.Actual), outputStyle))
	}
	if strings.TrimSpace(logs.assertion) != "" && !logs.hasComparison && (len(logs.expectedVariants) == 0 || !isExpectedStdoutComparison(logs.assertion)) {
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
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("│ "))
		b.WriteString(style.Render(line))
	}
	return b.String()
}

func renderExpectedVariants(variants []string) string {
	var b strings.Builder
	b.WriteString(labelStyle.Render("Expected stdout"))
	for i, variant := range variants {
		if len(variants) > 1 {
			b.WriteString("\n")
			b.WriteString(mutedStyle.Render(fmt.Sprintf("│ option %d", i+1)))
		}
		lines := strings.Split(strings.TrimSuffix(variant, "\n"), "\n")
		if len(lines) == 1 && lines[0] == "" {
			if len(variants) > 1 {
				b.WriteString(" ")
			} else {
				b.WriteString("\n")
				b.WriteString(mutedStyle.Render("│ "))
			}
			b.WriteString(mutedStyle.Render("∅"))
			continue
		}
		for _, line := range lines {
			b.WriteString("\n")
			if len(variants) > 1 {
				b.WriteString(mutedStyle.Render("│   "))
			} else {
				b.WriteString(mutedStyle.Render("│ "))
			}
			b.WriteString(outputStyle.Render(line))
		}
	}
	return b.String()
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
				if isBoundaryLine(next) {
					break
				}
				i++
				parsed.actualOutput = append(parsed.actualOutput, trimTesterIndent(next))
			}

		case hasExpectedStdoutPrefix(trimmed):
			expected, actual, hasActual, ok := parseExpectedStdout(trimmed)
			if ok {
				parsed.expectedVariants = expected
				actualLines := strings.Split(strings.TrimSuffix(actual, "\n"), "\n")
				if hasActual && shouldUseParsedActual(parsed.actualOutput, actualLines) {
					parsed.actualOutput = actualLines
				}
			}
			parsed.assertion = trimmed

		case isValueComparisonLine(trimmed):
			comparison, ok := parseValueComparison(trimmed)
			if ok {
				parsed.comparison = comparison
				parsed.hasComparison = true
			}
			parsed.assertion = trimmed

		case isCaseSummaryLine(trimmed):
			parsed.caseSummary = formatCaseSummary(trimmed)

		case strings.HasPrefix(trimmed, chunkingCasePrefix):
			parsed.notes = append(parsed.notes, trimmed)
			if parsed.caseSummary == "" {
				parsed.caseSummary = inferChunkingCaseSummary(trimmed)
			}

		case shouldAttachToAssertion(parsed, trimmed):
			if isExpectedStdoutComparison(parsed.assertion) {
				parsed.assertion = trimmed
			} else {
				parsed.assertion += "\n" + trimmed
			}

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

func indentBlock(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func shouldUseParsedActual(existing []string, parsed []string) bool {
	if len(existing) == 0 {
		return true
	}
	if len(parsed) <= 1 {
		return false
	}

	first := strings.TrimSpace(parsed[0])
	return first == "{" || first == "["
}

func isCaseSummaryLine(line string) bool {
	return caseSummaryPattern.MatchString(line)
}

func formatCaseSummary(line string) string {
	matches := caseSummaryPattern.FindStringSubmatch(line)
	if len(matches) != 3 {
		return ""
	}

	return fmt.Sprintf("%s/%s cases passed", matches[1], matches[2])
}

func inferChunkingCaseSummary(line string) string {
	name := strings.TrimPrefix(line, chunkingCasePrefix)
	passed, ok := chunkingCasePassedCount[name]
	if !ok {
		return ""
	}

	return fmt.Sprintf("%d/3 cases passed", passed)
}

func isBoundaryLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return hasExpectedStdoutPrefix(trimmed) ||
		strings.HasPrefix(trimmed, "Test failed") ||
		strings.HasPrefix(trimmed, "$ ") ||
		(!isIndentedOutputLine(line) && isValueComparisonLine(trimmed))
}

func isIndentedOutputLine(line string) bool {
	return strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "     ")
}

func hasExpectedStdoutPrefix(line string) bool {
	return strings.HasPrefix(strings.ToLower(line), "expected stdout")
}

func isExpectedStdoutComparison(line string) bool {
	if !hasExpectedStdoutPrefix(line) {
		return false
	}

	_, _, _, ok := parseExpectedStdout(line)
	return ok
}

func shouldAttachToAssertion(parsed parsedLogs, line string) bool {
	return len(parsed.expectedVariants) > 0 &&
		strings.TrimSpace(parsed.assertion) != "" &&
		line != "Test failed"
}

func isValueComparisonLine(line string) bool {
	_, ok := parseValueComparison(line)
	return ok
}

func parseValueComparison(line string) (valueComparison, bool) {
	if hasExpectedStdoutPrefix(line) {
		return valueComparison{}, false
	}

	const expectedMarker = ": expected "
	const actualMarker = ", got "

	expectedIndex := strings.Index(line, expectedMarker)
	if expectedIndex <= 0 {
		return valueComparison{}, false
	}

	rest := line[expectedIndex+len(expectedMarker):]
	actualIndex := strings.LastIndex(rest, actualMarker)
	if actualIndex == -1 {
		return valueComparison{}, false
	}

	subject := strings.TrimSpace(line[:expectedIndex])
	expected := strings.TrimSpace(rest[:actualIndex])
	actual := strings.TrimSpace(rest[actualIndex+len(actualMarker):])
	if subject == "" || expected == "" {
		return valueComparison{}, false
	}

	return valueComparison{
		Subject:  subject,
		Expected: formatComparisonValue(expected),
		Actual:   formatComparisonValue(actual),
	}, true
}

func formatComparisonValue(value string) string {
	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}
	return value
}

func comparisonLines(subject, value string) []string {
	lines := strings.Split(strings.TrimSuffix(value, "\n"), "\n")
	if len(lines) == 1 {
		return []string{fmt.Sprintf("%s: %s", subject, lines[0])}
	}

	out := []string{subject + ":"}
	for _, line := range lines {
		out = append(out, "  "+line)
	}
	return out
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
