<div align="center">

# 🧠 VibeLog

**AI Coding Session Memory — Never lose context again when vibe coding with Claude, Cursor, or any AI agent.**

[![Go Report Card](https://goreportcard.com/badge/github.com/ShivamMathtech/vibelog)](https://goreportcard.com/report/github.com/ShivamMathtech/vibelog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/ShivamMathtech/vibelog)](https://github.com/ShivamMathtech/vibelog/releases)
[![Stars](https://img.shields.io/github/stars/ShivamMathtech/vibelog?style=social)](https://github.com/ShivamMathtech/vibelog)

</div>

---

## The Problem

When you "vibe code" with AI agents, you lose everything when the session ends:

- The context window resets
- The agent forgets why you chose that database schema
- 3 days later, you have no idea what prompts actually worked
- Every project becomes a black hole of lost context

## The Solution

**VibeLog** is a single-binary, zero-dependency tool that:

1. **Captures** every AI-human coding session (prompts, responses, file changes, decisions)
2. **Remembers** via an MCP server — any agent can query "what did we decide about the auth flow last Tuesday?"
3. **Surfaces** a beautiful TUI to browse, search, and export your coding history
4. **Syncs** context across branches and machines (local-first)

---

## ✨ Features

| Feature                  | Description                                                |
| ------------------------ | ---------------------------------------------------------- |
| 🔌 **MCP Server**        | Plug into Claude Code, Cursor, or any MCP-compatible agent |
| 🔍 **Full-Text Search**  | SQLite FTS5 for instant history search                     |
| 🎯 **Decision Tracking** | Tag and retrieve important technical decisions             |
| 🖥️ **Beautiful TUI**     | Browse sessions with Charmbracelet Bubble Tea              |
| 📝 **Markdown Export**   | Export any session to a shareable document                 |
| 🪝 **Git Hooks**         | Auto-capture commits and branch switches                   |
| 👁️ **File Watcher**      | Real-time tracking of file changes during sessions         |
| 🏠 **Local-First**       | SQLite database, no cloud, no telemetry                    |
| 📦 **Single Binary**     | One static binary, zero runtime dependencies               |

---

## 🚀 Quick Start

### Install

#### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/ShivamMathtech/vibelog/main/install.sh | sh
```

#### Windows (PowerShell)

```powershell
# Download latest release manually from:
# https://github.com/ShivamMathtech/vibelog/releases/latest
# Or use scoop/chocolatey (coming soon)
```

#### Build from Source

```bash
git clone https://github.com/ShivamMathtech/vibelog.git
cd vibelog
go mod tidy
CGO_ENABLED=0 go build -tags=sqlite_fts5 -ldflags="-s -w" -o vibelog ./cmd/vibelog
```

---

## 📋 Usage

### 1. Initialize a Repository

```bash
cd my-project
vibelog init
```

This installs git hooks and creates a `.vibelog/` directory.

### 2. Start a Session

```bash
# Start a new session (or let git hooks auto-create one)
export SESSION_ID=$(uuidgen)
echo $SESSION_ID > .vibelog/session.id
```

**Windows PowerShell:**

```powershell
$session = [guid]::NewGuid().ToString()
Set-Content -Path ".vibelog\session.id" -Value $session
```

### 3. Capture Events

```bash
# Capture a prompt
vibelog capture --session=$SESSION_ID --type=prompt --content="How do I refactor this auth module?"

# Capture the agent's response
vibelog capture --session=$SESSION_ID --type=response --content="Use JWT tokens with refresh strategy..."

# Capture a decision
vibelog capture --session=$SESSION_ID --type=decision --content="Switch from SQLite to PostgreSQL"
```

### 4. Connect to Your Agent (MCP)

Add to `.vibelog/mcp.json` (or your agent's MCP config):

```json
{
  "mcpServers": {
    "vibelog": {
      "command": "vibelog",
      "args": ["mcp", "--repo", "."]
    }
  }
}
```

**Now your agent can ask:**

- _"Why did we choose PostgreSQL over SQLite?"_ → `get_decisions`
- _"What did we do to the auth module last week?"_ → `search_sessions`
- _"Summarize Tuesday's session."_ → `summarize_session`

### 5. Browse & Search History

```bash
vibelog tui              # Interactive browser (arrow keys + Enter)
vibelog list             # Plain text list
vibelog search "UUID"     # Full-text search across all sessions
vibelog export <id>       # Export session to markdown
```

---

## 🏗️ Architecture

```
vibelog/
├── cmd/vibelog/          # CLI entrypoint
├── internal/
│   ├── capture/          # Git hooks + file watcher (fsnotify)
│   ├── mcp/              # MCP server (stdio JSON-RPC 2.0)
│   ├── store/            # SQLite + FTS5 full-text search
│   ├── tui/              # Bubble Tea interactive interface
│   └── export/           # Markdown generation
└── pkg/types/            # Shared data structures
```

---

## 🛠️ Commands

| Command                       | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| `vibelog init [path]`         | Install git hooks in current or specified repo               |
| `vibelog capture`             | Manually capture an event (prompt, response, decision, etc.) |
| `vibelog mcp --repo .`        | Start MCP server for agent integration                       |
| `vibelog list`                | List all sessions (plain text)                               |
| `vibelog tui`                 | Launch interactive TUI browser                               |
| `vibelog export <session-id>` | Export session to `vibelog-<id>.md`                          |
| `vibelog search <query>`      | Full-text search across all history                          |
| `vibelog decision`            | Record a tagged decision                                     |

---

## 💻 Tech Stack

- **Go 1.22+** — Single static binary, cross-platform
- **modernc.org/sqlite** — Pure Go SQLite (no CGO required)
- **FTS5** — Full-text search engine built into SQLite
- **Bubble Tea** — Terminal UI framework (Charmbracelet)
- **MCP** — Model Context Protocol for AI agent integration

---

## 📦 Releases

Download pre-built binaries for your platform:

| Platform              | Download                                                                                 |
| --------------------- | ---------------------------------------------------------------------------------------- |
| macOS (Intel)         | [vibelog_darwin_amd64.tar.gz](https://github.com/ShivamMathtech/vibelog/releases/latest) |
| macOS (Apple Silicon) | [vibelog_darwin_arm64.tar.gz](https://github.com/ShivamMathtech/vibelog/releases/latest) |
| Linux (x86_64)        | [vibelog_linux_amd64.tar.gz](https://github.com/ShivamMathtech/vibelog/releases/latest)  |
| Linux (ARM64)         | [vibelog_linux_arm64.tar.gz](https://github.com/ShivamMathtech/vibelog/releases/latest)  |
| Windows (x86_64)      | [vibelog_windows_amd64.zip](https://github.com/ShivamMathtech/vibelog/releases/latest)   |

---

## 🗺️ Roadmap

- [ ] Web dashboard for remote browsing
- [ ] Cross-machine sync via CRDT
- [ ] AI-generated session summaries
- [ ] VS Code extension
- [ ] Decision confidence scoring
- [ ] Team shared sessions
- [ ] Scoop / Chocolatey / Homebrew packages

---

## 🤝 Contributing

We welcome contributions! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the vibe coding community.**

[⭐ Star this repo](https://github.com/ShivamMathtech/vibelog) if you find it useful!

</div>
