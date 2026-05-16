package picker

import (
	"testing"

	"github.com/chibuka/95-cli/internal/presets"
)

func TestPickerGoAndRustOptionsExposeBuildRunPairs(t *testing.T) {
	goOptions := optionsForLanguage(t, "go")
	if len(goOptions) != 2 {
		t.Fatalf("expected two go options, got %d", len(goOptions))
	}
	assertOption(t, goOptions[0], "go run .", presets.GoBuildCurrentPackage, presets.GoRunBinary)
	assertOption(t, goOptions[1], "go run main.go", presets.GoBuildMainFile, presets.GoRunBinary)

	rustOptions := optionsForLanguage(t, "rust")
	if len(rustOptions) != 1 {
		t.Fatalf("expected one rust option, got %d", len(rustOptions))
	}
	assertOption(t, rustOptions[0], "Cargo project", presets.RustBuildRelease, presets.RustRunRelease)
}

func optionsForLanguage(t *testing.T, name string) []Option {
	t.Helper()

	for _, lang := range languages {
		if lang.name == name {
			return lang.options
		}
	}
	t.Fatalf("language %q not found", name)
	return nil
}

func assertOption(t *testing.T, opt Option, label, buildCommand, runCommand string) {
	t.Helper()

	if opt.Label != label {
		t.Fatalf("expected label %q, got %q", label, opt.Label)
	}
	if opt.BuildCommand != buildCommand {
		t.Fatalf("expected build %q, got %q", buildCommand, opt.BuildCommand)
	}
	if opt.RunCommand != runCommand {
		t.Fatalf("expected run %q, got %q", runCommand, opt.RunCommand)
	}
}
