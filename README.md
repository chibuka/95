# 95 CLI

**Build from scratch**

A command-line tool for [95ninefive.dev](https://95ninefive.dev). Your code runs locally, gets validated server-side, and tracks your progress as you level up your skills.

## Features

- **GitHub OAuth Authentication** - Secure login with your GitHub account
- **Local Execution** - Your code runs on your machine for fast feedback
- **Server-Side Validation** - Expected outputs never leave the server (prevents cheating)
- **Progress Tracking** - Automatically saves your progress as you complete stages
- **Cascading Tests** - Tests all prerequisite stages to ensure backward compatibility

## Installation

### Quick Install (Recommended)

**macOS/Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/chibuka/95-cli/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
go install github.com/chibuka/95-cli/cmd/95@latest
```

Notes (Windows only):
- macOS/Linux keep using the curl installer above, or `go build` from the repo root (output `95`).
- `go install github.com/chibuka/95-cli/cmd/95@latest` is required on Windows so the binary is `95.exe` (installing the module root would produce `95-cli.exe`).
- This requires Go (1.24+ recommended). Binaries land in `%USERPROFILE%\go\bin` by default.
- Go does **not** always update your PATH automatically. If `95` is not found, add `%USERPROFILE%\go\bin` to your user PATH.
- If you are using WSL and the browser does not open automatically, run `95 login --headless` and open the printed URL in your Windows browser.

### Manual Installation

1. Download the latest binary for your platform from [Releases](https://github.com/chibuka/95-cli/releases)
2. Extract and move to your PATH:
   ```bash
   # macOS/Linux
   sudo mv 95 /usr/local/bin/
   chmod +x /usr/local/bin/95
   ```

### Build from Source

Requirements: Go 1.24+

```bash
git clone https://github.com/chibuka/95-cli.git
cd 95-cli
go build -o 95
sudo mv 95 /usr/local/bin/
```

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a branch
3. Make your changes
5. Submit a pr, thank you :)

## License

MIT License - see [LICENSE](LICENSE) for details

## Links

- **Website:** https://95ninefive.dev
- **GitHub Repo:** https://github.com/chibuka/95-cli
- **Personal GitHub:** https://github.com/grainme

---

ありがとうございます 💯
