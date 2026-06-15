# VibeLog 🧠

> **AI Coding Session Memory** — Never lose context again when vibe coding with Claude, Cursor, or any AI agent.

[![Go Report Card](https://goreportcard.com/badge/github.com/vibelog/vibelog)](https://goreportcard.com/report/github.com/vibelog/vibelog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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

## Features

- 🔌 **MCP Server** — Plug into Claude Code, Cursor, or any MCP-compatible agent
- 🔍 **Full-Text Search** — SQLite FTS5 for instant history search
- 🎯 **Decision Tracking** — Tag and retrieve important technical decisions
- 🖥️ **Beautiful TUI** — Browse sessions with Charmbracelet Bubble Tea
- 📝 **Markdown Export** — Export any session to a shareable document
- 🪝 **Git Hooks** — Auto-capture commits and branch switches
- 👁️ **File Watcher** — Real-time tracking of file changes during sessions
- 🏠 **Local-First** — SQLite database, no cloud, no telemetry
- 📦 **Single Binary** — One static binary, zero runtime dependencies

## Quick Start

### Install

```bash
curl -fsSL https://vibelog.dev/install | sh
```

Or download from [Releases](https://github.com/vibelog/vibelog/releases).

### Initialize a Repository

```bash
cd my-project
vibelog init
```

This installs git hooks and creates a `.vibelog/` directory.

### Start a Session

```bash
# Start a new session (or let git hooks auto-create one)
export SESSION_ID=$(uuidgen)
echo $SESSION_ID > .vibelog/session.id

# Capture prompts and responses
vibelog capture --session=$SESSION_ID --type=prompt --content="How do I refactor this auth module?"
claude --session=$SESSION_ID  # Your agent interaction
vibelog capture --session=$SESSION_ID --type=response --content="Here's the refactored code..."
```

### Connect to Your Agent (MCP)

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

Now your agent can ask:
- *"Why did we choose PostgreSQL over SQLite?"* → `get_decisions`
- *"What did we do to the auth module last week?"* → `search_sessions`
- *"Summarize Tuesday's session."* → `summarize_session`

### Browse History

```bash
vibelog tui        # Interactive browser
vibelog list       # Plain text list
vibelog search "UUID"   # Full-text search
```

### Export and Share

```bash
vibelog export <session-id>   # Creates vibelog-<id>.md
```

## Architecture

```
vibelog/
├── cmd/vibelog/          # CLI entrypoint
├── internal/
│   ├── capture/          # Git hooks + file watcher
│   ├── mcp/              # MCP server (stdio JSON-RPC)
│   ├── store/            # SQLite + FTS5
│   ├── tui/              # Bubble Tea interface
│   └── export/           # Markdown generation
└── pkg/types/            # Shared structs
```

## Commands

| Command | Description |
|---------|-------------|
| `vibelog init [path]` | Install git hooks in repo |
| `vibelog capture` | Manually capture an event |
| `vibelog mcp --repo .` | Start MCP server |
| `vibelog list` | List all sessions |
| `vibelog tui` | Interactive browser |
| `vibelog export <id>` | Export to markdown |
| `vibelog search <query>` | Full-text search |
| `vibelog decision` | Record a tagged decision |

## Tech Stack

- **Go 1.22+** — Single static binary
- **modernc.org/sqlite** — Pure Go SQLite (no CGO)
- **FTS5** — Full-text search engine
- **Bubble Tea** — Terminal UI framework
- **MCP** — Model Context Protocol for agent integration

## Building from Source

```bash
git clone https://github.com/vibelog/vibelog.git
cd vibelog
go build -tags=sqlite_fts5 -o vibelog ./cmd/vibelog
```

## Roadmap

- [ ] Web dashboard for remote browsing
- [ ] Cross-machine sync via CRDT
- [ ] AI-generated session summaries
- [ ] VS Code extension
- [ ] Decision confidence scoring
- [ ] Team shared sessions

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT © VibeLog Contributors
