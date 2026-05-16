package presets

import "testing"

func TestUpgradeLegacyBuildRunUpgradesKnownGoCommands(t *testing.T) {
	tests := []struct {
		name      string
		run       string
		wantBuild string
	}{
		{name: "current package", run: "go run .", wantBuild: GoBuildCurrentPackage},
		{name: "main file", run: "go run main.go", wantBuild: GoBuildMainFile},
		{name: "old compiled option", run: "go build -o ./app && ./app", wantBuild: GoBuildCurrentPackage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build, run, ok := UpgradeLegacyBuildRun("go", "", tt.run)
			if !ok {
				t.Fatal("expected legacy go config to upgrade")
			}
			if build != tt.wantBuild {
				t.Fatalf("expected build %q, got %q", tt.wantBuild, build)
			}
			if run != GoRunBinary {
				t.Fatalf("expected run %q, got %q", GoRunBinary, run)
			}
		})
	}
}

func TestUpgradeLegacyBuildRunUpgradesRustCargoRun(t *testing.T) {
	build, run, ok := UpgradeLegacyBuildRun("rust", "", "cargo run")
	if !ok {
		t.Fatal("expected legacy rust config to upgrade")
	}
	if build != RustBuildRelease {
		t.Fatalf("expected build %q, got %q", RustBuildRelease, build)
	}
	if run != RustRunRelease {
		t.Fatalf("expected run %q, got %q", RustRunRelease, run)
	}
}

func TestUpgradeLegacyBuildRunLeavesCurrentConfigsAlone(t *testing.T) {
	build, run, ok := UpgradeLegacyBuildRun("go", GoBuildCurrentPackage, GoRunBinary)
	if ok {
		t.Fatal("did not expect current config to upgrade")
	}
	if build != GoBuildCurrentPackage {
		t.Fatalf("expected build %q, got %q", GoBuildCurrentPackage, build)
	}
	if run != GoRunBinary {
		t.Fatalf("expected run %q, got %q", GoRunBinary, run)
	}
}
