# p4u-skill

A cross-platform Perforce CLI enhancement tool, reimplementing [p4u](../p4u/) in Go.

## Features

- **Cross-platform**: Windows, macOS (Intel & Apple Silicon), Linux
- **Single binary**: no external dependencies (no bash, gnu-parallel, tput, etc.)
- **Automation-friendly**: `--non-interactive` and `--json` flags for CI/AI use
- **Concurrent**: parallel changelist fetching without gnu-parallel
- **ColorUI**: auto-detects terminal capability, respects `NO_COLOR`

## Commands

| Command | Description |
|---------|-------------|
| `p4u show` | Show client status: pending & shelved changelists |
| `p4u show-cl <CL>` | Pretty-print a single changelist |
| `p4u switch [CL...]` | Shelve current work, unshelve target changelist(s) |
| `p4u delete-cl [CL...]` | Delete changelist(s) completely |
| `p4u reshelve <CL>` | Re-shelve a changelist |
| `p4u unshelve <CL>` | Unshelve to itself (not default) |
| `p4u revert-all` | Revert all opened files |
| `p4u annotate <file> <line>` | Show changelist for specific file line |
| `p4u delete-client` | Delete a P4 client completely |
| `p4u untracked [dir...]` | Find untracked files |

## Installation

### From source

```bash
# Build for current platform
cd p4u-skill
make build

# Install to $GOPATH/bin
make install

# Build all platforms
make build-all
```

### Pre-built binaries (latest release)

```bash
# macOS / Linux — auto-detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]') && ARCH=$(uname -m)
[[ "$ARCH" == "x86_64" ]] && ARCH=amd64 || ARCH=arm64
curl -fsSL "https://github.com/m9rco/p4u-skill/releases/latest/download/p4u-${OS}-${ARCH}" \
  -o /tmp/p4u && chmod +x /tmp/p4u && sudo mv /tmp/p4u /usr/local/bin/p4u
```

Windows: download `p4u-windows-amd64.exe` from the [releases page](https://github.com/m9rco/p4u-skill/releases/latest).

## Usage

```bash
# Show current client status
p4u show

# Show all changelists (verbose)
p4u show -v

# Show only shelved changelists as JSON
p4u show -s --json

# Switch context to changelist 12345
p4u switch 12345

# Delete a changelist
p4u delete-cl 12345

# Find who changed line 42 in a file
p4u annotate //depot/main/foo.cpp 42

# Find untracked files
p4u untracked
```

## Global Flags

```
--non-interactive   Disable interactive prompts (for automation)
--json              Output in JSON format
-n, --no-color      Disable color output
-o, --force-color   Force color output (when piping)
```

## Architecture

```
p4u-skill/
├── main.go               # Entry point
├── cmd/                  # Cobra command definitions
│   ├── root.go           # Root command, global flags
│   ├── show.go           # p4u show
│   ├── show_changelist.go # p4u show-cl
│   ├── switch.go         # p4u switch
│   ├── delete_changelist.go # p4u delete-cl
│   ├── reshelve.go       # p4u reshelve
│   ├── unshelve.go       # p4u unshelve
│   ├── revert_all.go     # p4u revert-all
│   ├── annotate.go       # p4u annotate
│   ├── delete_client.go  # p4u delete-client
│   └── untracked.go      # p4u untracked
└── internal/
    ├── p4/               # P4 CLI wrapper
    │   ├── client.go     # Executor interface & CLIExecutor
    │   ├── info.go       # p4 info parsing
    │   ├── changes.go    # changelist queries
    │   ├── describe.go   # changelist detail parsing
    │   ├── opened.go     # opened files & client ops
    │   └── ops.go        # shelve/unshelve/revert/sync
    ├── ui/               # Terminal UI
    │   ├── color.go      # Cross-platform colors
    │   ├── progress.go   # Progress indicator
    │   ├── picker.go     # Interactive changelist picker
    │   └── prompt.go     # Y/n prompts & list selection
    └── output/           # Output formatting
        └── formatter.go  # Text & JSON printer
```

## Development

```bash
# Run tests
make test

# Lint
make lint

# Build & test
make build && ./p4u --help
```

## Skill Integration

This repo ships a skill definition compatible with both
[Claude Code](https://claude.ai/code) and [CodeBuddy Code](https://cnb.cool/codebuddy/codebuddy-code).

| Client | Skill path |
|--------|-----------|
| Claude Code (Anthropic) | `.claude/skills/p4u/SKILL.md` |
| CodeBuddy Code | `.codebuddy/skills/p4u/SKILL.md` |

Both files are identical in content.

### Project-level install

Copy the relevant directory into your project root so the AI client picks
it up automatically when you work in that project:

```bash
# Claude Code
cp -r .claude  /path/to/your/project/

# CodeBuddy Code
cp -r .codebuddy  /path/to/your/project/
```

### User-level install (available in all projects)

```bash
# Claude Code
mkdir -p ~/.claude/skills
cp -r .claude/skills/p4u  ~/.claude/skills/

# CodeBuddy Code
mkdir -p ~/.codebuddy/skills
cp -r .codebuddy/skills/p4u  ~/.codebuddy/skills/
```

Once installed, the skill auto-detects whether `p4u` is on your `$PATH`
and provides a one-liner to download the correct binary from GitHub Releases
if it is missing.
