---
name: p4u
description: >
  Perforce CLI enhancement tool. Use when the user needs to inspect or
  manipulate Perforce changelists, shelved files, clients, or file history.
  Wraps p4 with better UX, JSON output, and automation flags.
allowed-tools:
  - Bash(p4u *)
  - Bash(curl *)
  - Bash(uname *)
  - Bash(chmod *)
  - Bash(sudo mv *)
  - Bash(which *)
---

# p4u

`p4u` is a cross-platform Perforce CLI enhancement tool (Go binary, no
external dependencies). It wraps common `p4` workflows with colour output,
JSON mode, and `--non-interactive` support for automation.

**GitHub:** https://github.com/m9rco/p4u-skill

## Binary status

!`which p4u 2>/dev/null && printf "✓ p4u found: %s\n" "$(which p4u)" || printf "✗ p4u not found on PATH — see 'Install' below\n"`

## Install (if not yet installed)

Auto-detects OS and architecture, downloads the latest release from GitHub:

```bash
set -e
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in x86_64) ARCH=amd64 ;; aarch64|arm64) ARCH=arm64 ;; esac
URL="https://github.com/m9rco/p4u-skill/releases/latest/download/p4u-${OS}-${ARCH}"
echo "Downloading p4u (latest) for ${OS}/${ARCH}..."
curl -fsSL "$URL" -o /tmp/p4u
chmod +x /tmp/p4u
sudo mv /tmp/p4u /usr/local/bin/p4u
echo "Done: $(p4u --help 2>&1 | head -1)"
```

> **Windows**: download `p4u-windows-amd64.exe` from
> https://github.com/m9rco/p4u-skill/releases/latest, rename to `p4u.exe`,
> and place it on `%PATH%`.

## Prerequisites

- `p4u` installed (see above)
- Perforce CLI (`p4`) installed and configured
- Logged in via `p4 login`

## Usage rules

1. **Always** append `--non-interactive` to every command — prevents hanging on prompts.
2. Append `--json` when the output needs to be parsed programmatically.
3. Changelist numbers are plain integers passed as strings (e.g. `12345`).
4. `p4u` auto-detects the current user and client from `p4 info`.

---

## Commands

### Show client status

```bash
p4u show --non-interactive              # pending + shelved changelists
p4u show -s --non-interactive           # shelved only
p4u show -p --non-interactive           # pending only
p4u show --json --non-interactive       # JSON output
p4u show -u <user> --non-interactive    # filter by user
p4u show -c <client> --non-interactive  # filter by client
p4u show -m 10 --non-interactive        # last 10 changelists
```

### Inspect a single changelist

```bash
p4u show-cl <CL> --non-interactive
p4u show-cl <CL> --json --non-interactive
p4u show-cl -b <CL> --non-interactive   # brief: no file list
```

### Switch working context

Shelves all current work then unshelves the target changelist(s).

```bash
p4u switch <CL> --non-interactive
p4u switch <CL1> <CL2> --non-interactive    # multiple changelists
p4u switch -s -m <CL> --non-interactive     # sync + auto-resolve after
p4u switch -k <CL> --non-interactive        # keep shelved copy
p4u switch --non-interactive                # just shelve everything
```

### Delete a changelist

```bash
p4u delete-cl <CL> --non-interactive
p4u delete-cl -f <CL> --non-interactive     # force, skip confirmation
p4u delete-cl <CL1> <CL2> --non-interactive
```

### Reshelve / unshelve

```bash
p4u reshelve <CL> --non-interactive
p4u unshelve <CL> --non-interactive
```

### Revert all opened files

```bash
p4u revert-all --non-interactive
```

### Annotate — who changed a line

```bash
p4u annotate <depot-path> <line> --non-interactive
p4u annotate -v <depot-path> <line> --non-interactive   # verbose CL info
p4u annotate --json <depot-path> <line> --non-interactive
```

Example:
```bash
p4u annotate //depot/main/src/foo.cpp 42 --non-interactive
```

### Delete a client

```bash
p4u delete-client --non-interactive            # current client
p4u delete-client -c <client> --non-interactive
p4u delete-client -f --non-interactive         # no confirmation
p4u delete-client -n --non-interactive         # keep local files
```

### Find untracked files

```bash
p4u untracked --non-interactive
p4u untracked ./src ./assets --non-interactive
p4u untracked -d 3 --non-interactive           # max depth 3
p4u untracked --json --non-interactive
```

---

## Global flags

| Flag               | Short | Description                          |
|--------------------|-------|--------------------------------------|
| `--non-interactive`|       | Disable all interactive prompts      |
| `--json`           |       | Emit JSON instead of coloured text   |
| `--no-color`       | `-n`  | Disable colour output                |
| `--force-color`    | `-o`  | Force colour even when piping        |
