package presets

import "strings"

const (
	GoBuildCurrentPackage = "mkdir -p .95bin && CGO_ENABLED=0 GOTOOLCHAIN=local go build -o .95bin/_95prog ."
	GoBuildMainFile       = "mkdir -p .95bin && CGO_ENABLED=0 GOTOOLCHAIN=local go build -o .95bin/_95prog ./main.go"
	GoRunBinary           = "/submission/.95bin/_95prog"

	RustBuildRelease = "/usr/local/cargo/bin/cargo build --release"
	RustRunRelease   = "/usr/local/cargo/bin/cargo run --release --offline --quiet --"
)

func UpgradeLegacyBuildRun(language, buildCommand, runCommand string) (string, string, bool) {
	lang := strings.ToLower(strings.TrimSpace(language))
	build := strings.TrimSpace(buildCommand)
	run := strings.TrimSpace(runCommand)
	if build != "" {
		return buildCommand, runCommand, false
	}

	switch lang {
	case "go":
		switch run {
		case "go run .":
			return GoBuildCurrentPackage, GoRunBinary, true
		case "go run main.go":
			return GoBuildMainFile, GoRunBinary, true
		case "go build -o ./app && ./app":
			return GoBuildCurrentPackage, GoRunBinary, true
		}
	case "rust":
		if run == "cargo run" {
			return RustBuildRelease, RustRunRelease, true
		}
	}

	return buildCommand, runCommand, false
}
