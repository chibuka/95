package config

import (
	"os"
	"testing"
)

func TestSaveProjectConfigPersistsBuildAndRunCommands(t *testing.T) {
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	const (
		buildCommand = "go build -o app"
		runCommand   = "./app"
		language     = "go"
	)
	if err := SaveProjectConfig(buildCommand, runCommand, language); err != nil {
		t.Fatalf("save project config: %v", err)
	}

	cfg, err := LoadProjectConfig()
	if err != nil {
		t.Fatalf("load project config: %v", err)
	}
	if cfg.BuildCommand != buildCommand {
		t.Fatalf("expected build %q, got %q", buildCommand, cfg.BuildCommand)
	}
	if cfg.RunCommand != runCommand {
		t.Fatalf("expected run %q, got %q", runCommand, cfg.RunCommand)
	}
	if cfg.Language != language {
		t.Fatalf("expected language %q, got %q", language, cfg.Language)
	}
}
