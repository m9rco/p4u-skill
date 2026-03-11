---
name: p4u
description: Cross-platform Perforce CLI enhancement tool. Use to query and manipulate Perforce changelists, shelves, clients, and file annotations. Supports Windows/macOS/Linux.
allowed-tools:
  - Bash(p4u:*)
---

# p4u Skill

`p4u` is a cross-platform Perforce CLI enhancement tool built with Go.
It wraps common `p4` workflows with better UX, color output, and automation support.

## Prerequisites

- `p4u` binary must be installed and on `$PATH`
- Perforce CLI (`p4`) must be installed and configured
- User must be logged in (`p4 login`)

## Installing p4u

Build from source:
```bash
cd p4u-skill && make install
```

Or download a pre-built binary from the releases page and place it on `$PATH`.

## Commands

Always pass `--non-interactive` when using from automation (avoids prompts).
Use `--json` for structured output.

### Show client status

```bash
# Show all pending & shelved changelists
p4u show --non-interactive

# Show only shelved changelists
p4u show -s --non-interactive

# Show as JSON
p4u show --json --non-interactive

# Filter by user or client
p4u show -u username --non-interactive
p4u show -c clientname --non-interactive
```

### Show a specific changelist

```bash
p4u show-cl 12345 --non-interactive
p4u show-cl 12345 --json --non-interactive
p4u show-cl -b 12345  # brief (no file list)
```

### Switch working context

```bash
# Shelve all current work and unshelve changelist 12345
p4u switch 12345 --non-interactive

# Switch to multiple changelists
p4u switch 12345 12346 --non-interactive

# Switch with sync and auto-resolve
p4u switch -s -m 12345 --non-interactive

# Keep the shelved copy after unshelving
p4u switch -k 12345 --non-interactive
```

### Delete a changelist

```bash
p4u delete-cl 12345 --non-interactive
p4u delete-cl -f 12345  # force, no prompts
p4u delete-cl 12345 12346  # delete multiple
```

### Reshelve / Unshelve

```bash
p4u reshelve 12345 --non-interactive
p4u unshelve 12345 --non-interactive
```

### Revert all opened files

```bash
p4u revert-all --non-interactive
```

### Annotate a file line

```bash
# Show changelist that last modified line 42 of a file
p4u annotate //depot/main/src/foo.cpp 42 --non-interactive

# With verbose changelist info
p4u annotate -v //depot/main/src/foo.cpp 42 --non-interactive

# JSON output
p4u annotate --json //depot/main/src/foo.cpp 42 --non-interactive
```

### Delete a client

```bash
p4u delete-client --non-interactive          # current client
p4u delete-client -c myclient --non-interactive
p4u delete-client -f --non-interactive       # force, no prompts
p4u delete-client -n --non-interactive       # keep local files
```

### Find untracked files

```bash
p4u untracked --non-interactive
p4u untracked ./src ./assets --non-interactive
p4u untracked -d 3 --non-interactive  # max depth 3
p4u untracked --json --non-interactive
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--non-interactive` | | Disable prompts (required for automation) |
| `--json` | | Output as JSON |
| `--no-color` | `-n` | Disable color output |
| `--force-color` | `-o` | Force color even when piping |

## Usage Notes for AI Assistants

1. **Always** pass `--non-interactive` to avoid hanging on prompts.
2. Use `--json` when you need to parse the output programmatically.
3. Changelist numbers are strings (e.g., `"12345"`), not integers.
4. `p4u switch` without arguments just shelves everything - useful for "save my work" scenarios.
5. The tool reads `p4 info` to detect the current client/user automatically.
