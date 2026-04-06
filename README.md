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

### Manual Installation

1. Download the latest binary for your platform from [Releases](https://github.com/chibuka/95-cli/releases)
2. Extract and move to your PATH:
   ```bash
   # macOS/Linux
   sudo mv 95 /usr/local/bin/
   chmod +x /usr/local/bin/95
   ```

### Build from Source

Requirements: Go 1.23+

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
