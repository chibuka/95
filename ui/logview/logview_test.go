package logview

import (
	"strings"
	"testing"
)

func TestParseLogsSplitsFailureTranscript(t *testing.T) {
	raw := strings.Join([]string{
		"Running tests for fixed capacity",
		"Running LRU cache case: do not grow past fixed capacity",
		"$ ./your_program.sh <<EOF",
		"CAPACITY 2",
		"SET 1 10",
		"EOF",
		"Your program output:",
		"  OK",
		"  10",
		`Expected stdout ["OK\n10\n" "OK\n-1\n"], got "OK\n10\n"`,
		"Test failed",
	}, "\n")

	got := parseLogs(raw)

	if len(got.notes) != 3 {
		t.Fatalf("notes len = %d, want 3: %#v", len(got.notes), got.notes)
	}
	if got.command[len(got.command)-1] != "EOF" {
		t.Fatalf("command = %#v, want heredoc through EOF", got.command)
	}
	if strings.Join(got.actualOutput, "\n") != "OK\n10" {
		t.Fatalf("actualOutput = %#v", got.actualOutput)
	}
	if len(got.expectedVariants) != 2 {
		t.Fatalf("expected variants = %#v, want 2", got.expectedVariants)
	}
	if got.expectedVariants[1] != "OK\n-1\n" {
		t.Fatalf("second variant = %q", got.expectedVariants[1])
	}
	if !strings.HasPrefix(got.assertion, "Expected stdout") {
		t.Fatalf("assertion = %q", got.assertion)
	}
}

func TestParseLogsKeepsSingleLineCommandFocused(t *testing.T) {
	raw := strings.Join([]string{
		"$ ./your_program.sh test.db .tables",
		"Your program output:",
		"  apples",
		`Expected stdout "oranges\n", got "apples\n"`,
	}, "\n")

	got := parseLogs(raw)

	if len(got.command) != 1 {
		t.Fatalf("command = %#v, want one command line", got.command)
	}
	if len(got.actualOutput) != 1 || got.actualOutput[0] != "apples" {
		t.Fatalf("actualOutput = %#v", got.actualOutput)
	}
}

func TestParseLogsShowsEmptyActualOutput(t *testing.T) {
	got := parseLogs(`Expected stdout "ready\n", got ""`)

	if len(got.actualOutput) != 1 || got.actualOutput[0] != "" {
		t.Fatalf("actualOutput = %#v, want visible empty output", got.actualOutput)
	}
}

func TestParseLogsPreservesBlankLinesInActualOutput(t *testing.T) {
	raw := strings.Join([]string{
		"Your program output:",
		"  first",
		"",
		"  third",
		`Expected stdout "first\n\nthird\n", got "first\nthird\n"`,
	}, "\n")

	got := parseLogs(raw)

	if strings.Join(got.actualOutput, "\n") != "first\n\nthird" {
		t.Fatalf("actualOutput = %#v, want blank line preserved", got.actualOutput)
	}
}

func TestParseLogsTreatsLowercaseExpectedStdoutAsAssertionBoundary(t *testing.T) {
	raw := strings.Join([]string{
		"Running tests for single document single chunk",
		"$ ./your_program.sh <<EOF",
		`{"documents":[{"id":"doc-1","text":"Databases store data."}]}`,
		"EOF",
		"Your program output:",
		"  Entire Input:",
		`  {"documents":[{"id":"doc-1","text":"Databases store data."}]}`,
		"expected stdout to be valid JSON: invalid character 'E' looking for beginning of value",
	}, "\n")

	got := parseLogs(raw)

	if strings.Join(got.actualOutput, "\n") != "Entire Input:\n{\"documents\":[{\"id\":\"doc-1\",\"text\":\"Databases store data.\"}]}" {
		t.Fatalf("actualOutput = %#v", got.actualOutput)
	}
	if got.assertion != "expected stdout to be valid JSON: invalid character 'E' looking for beginning of value" {
		t.Fatalf("assertion = %q", got.assertion)
	}
}

func TestRenderFloorShowsStandaloneAssertionAfterExpectedStdoutComparison(t *testing.T) {
	rendered := RenderFloor(FloorResult{
		Number: 1,
		Passed: false,
		Logs: strings.Join([]string{
			"Running tests",
			`Expected stdout "{\n  \"chunks\": []\n}", got "Entire Input:\n{}"`,
			"expected stdout to be valid JSON: invalid character 'E' looking for beginning of value",
		}, "\n"),
	})

	for _, want := range []string{"Your stdout", "Expected stdout", "Assertion", "expected stdout to be valid JSON"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, rendered)
		}
	}
}

func TestParseLogsKeepsExpectedStdoutAssertionOutOfNotes(t *testing.T) {
	raw := strings.Join([]string{
		"Running tests for single document multiple chunks",
		"Running chunking case: single document with multiple sentences",
		`Expected stdout "{\n  \"chunks\": []\n}", got "{\"chunks\":[]}"`,
		`chunk[1].text: expected "Indexes make lookups faster.", got "Indexes\nmake lookups faster."`,
		"Test failed",
	}, "\n")

	got := parseLogs(raw)

	for _, note := range got.notes {
		if strings.HasPrefix(note, "chunk[1].text") {
			t.Fatalf("assertion leaked into notes: %#v", got.notes)
		}
	}
	if got.assertion != `chunk[1].text: expected "Indexes make lookups faster.", got "Indexes\nmake lookups faster."` {
		t.Fatalf("assertion = %q", got.assertion)
	}
}

func TestRenderFloorShowsQuotedChunkMismatchComparison(t *testing.T) {
	rendered := RenderFloor(FloorResult{
		Number: 1,
		Passed: false,
		Logs: strings.Join([]string{
			"Running tests for single document multiple chunks",
			"Running chunking case: single document with multiple sentences",
			"Cases: 0/3 passed",
			`Expected stdout "{\n  \"chunks\": []\n}", got "{\"chunks\":[]}"`,
			`chunk[1].text: expected "Indexes make lookups faster.", got "Indexes\nmake lookups faster."`,
			"Test failed",
		}, "\n"),
	})

	for _, want := range []string{"Floor 1 · 0/3 cases passed", "Expected stdout", "Expected", "Actual", "chunk[1].text"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Assertion") {
		t.Fatalf("comparison should render as expected/actual sections:\n%s", rendered)
	}
	if strings.Contains(rendered, "Cases: 0/3 passed") {
		t.Fatalf("case summary leaked into body:\n%s", rendered)
	}
}

func TestParseLogsKeepsStandaloneComparisonOutOfActualOutput(t *testing.T) {
	raw := strings.Join([]string{
		"Your program output:",
		`  {"chunks":[{"id":"doc-1:0","text":"Alpha one. Beta two."}]}`,
		`chunk[1].text: expected "two. Gamma three. Delta four.", got "Gamma three. Delta four."`,
		"Test failed",
	}, "\n")

	got := parseLogs(raw)
	actual := strings.Join(got.actualOutput, "\n")

	if strings.Contains(actual, "chunk[1].text") {
		t.Fatalf("comparison leaked into actual output:\n%s", actual)
	}
	if !got.hasComparison {
		t.Fatalf("expected comparison to be parsed: %#v", got)
	}
	if got.comparison.Subject != "chunk[1].text" {
		t.Fatalf("comparison subject = %q", got.comparison.Subject)
	}
	if got.comparison.Expected != "two. Gamma three. Delta four." {
		t.Fatalf("comparison expected = %q", got.comparison.Expected)
	}
	if got.comparison.Actual != "Gamma three. Delta four." {
		t.Fatalf("comparison actual = %q", got.comparison.Actual)
	}
}

func TestRenderFloorSeparatesStandaloneComparison(t *testing.T) {
	rendered := RenderFloor(FloorResult{
		Number: 9,
		Passed: false,
		Logs: strings.Join([]string{
			"Running tests for chunk overlap",
			"Running chunking case: add chunk_overlap between adjacent chunks",
			"Your program output:",
			`  {"chunks":[{"id":"doc-1:0","text":"Alpha one. Beta two."}]}`,
			`chunk[1].text: expected "two. Gamma three. Delta four.", got "Gamma three. Delta four."`,
			"Test failed",
		}, "\n"),
	})

	for _, want := range []string{
		"Your stdout",
		"Expected",
		"Actual",
		"chunk[1].text: two. Gamma three. Delta four.",
		"chunk[1].text: Gamma three. Delta four.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, `chunk[1].text: expected "two. Gamma three. Delta four."`) {
		t.Fatalf("raw comparison leaked into rendered output:\n%s", rendered)
	}
}

func TestRenderFloorInfersChunkingCaseProgressWhenSummaryLineIsMissing(t *testing.T) {
	rendered := RenderFloor(FloorResult{
		Number: 2,
		Passed: false,
		Logs: strings.Join([]string{
			"Running tests for multiple documents",
			"Running chunking case: multiple documents keep their source references",
			`Expected stdout "{\n  \"chunks\": []\n}", got "{\"chunks\":[]}"`,
			`chunk[2].chunk_index: expected 0, got 2`,
			"Test failed",
		}, "\n"),
	})

	if !strings.Contains(rendered, "Floor 2 · 0/3 cases passed") {
		t.Fatalf("rendered output missing inferred case progress:\n%s", rendered)
	}
}

func TestRenderFloorShowsStructuredSections(t *testing.T) {
	rendered := RenderFloor(FloorResult{
		Number: 2,
		Passed: false,
		Logs: strings.Join([]string{
			"Running tests",
			"$ ./your_program.sh",
			"Your program output:",
			"  wrong",
			`Expected stdout "right\n", got "wrong\n"`,
		}, "\n"),
	})

	for _, want := range []string{"Floor 2", "What ran", "Command", "Your stdout", "Expected stdout"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Assertion") {
		t.Fatalf("rendered output should not duplicate parsed assertion:\n%s", rendered)
	}
}

func TestRenderFloorDoesNotArtificiallyWrapLongLines(t *testing.T) {
	longLine := "Running tests for a very long scenario name that should stay on one rendered line instead of being split by the CLI formatter"
	rendered := RenderFloor(FloorResult{
		Number: 2,
		Passed: false,
		Logs: strings.Join([]string{
			longLine,
			"$ ./your_program.sh <<EOF",
			"CAPACITY 2",
			"EOF",
			`Expected stdout ["OK\nOK\nOK\nOK\n10\n20\n-1\n" "OK\nOK\nOK\nOK\n-1\n20\n30\n"], got "OK\nOK\nOK\n10\n20\n-1\n"`,
		}, "\n"),
	})

	if !strings.Contains(rendered, longLine) {
		t.Fatalf("long line was artificially wrapped:\n%s", rendered)
	}
}

func TestParseLogsPrefersStructuredComparisonActual(t *testing.T) {
	raw := strings.Join([]string{
		"Your program output:",
		`  {"chunks":[{"document_id":"doc-1","chunk_index":1,"text":" Indexes make lookups faster."}]}`,
		`Expected stdout "{\n  \"chunks\": []\n}", got "{\n  \"chunks\": [\n    {\n      \"document_id\": \"doc-1\",\n      \"chunk_index\": 1,\n      \"text\": \" Indexes make lookups faster.\"\n    }\n  ]\n}"`,
		`chunk[1].text: expected "Indexes make lookups faster.", got " Indexes make lookups faster."`,
	}, "\n")

	got := parseLogs(raw)
	actual := strings.Join(got.actualOutput, "\n")

	for _, want := range []string{
		`"chunks": [`,
		`"text": " Indexes make lookups faster."`,
	} {
		if !strings.Contains(actual, want) {
			t.Fatalf("actual output missing %q:\n%s", want, actual)
		}
	}
}
