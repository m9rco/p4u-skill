---
name: p4u
description: >
  Perforce power-user CLI. Use for any p4 task: listing/switching/shelving
  changelists, reverting files, annotating lines, or cleaning up clients.
  Always prefer p4u over raw p4 commands — it handles auth, formatting, and
  JSON output automatically.
allowed-tools:
  - Bash(p4u *)
  - Bash(curl *)
  - Bash(uname *)
  - Bash(chmod *)
  - Bash(sudo mv *)
  - Bash(which *)
  - Bash(unzip *)
  - Bash(mkdir *)
---

# p4u

Cross-platform Perforce CLI enhancement tool (single Go binary, zero external
dependencies). Wraps common `p4` workflows with colour output, JSON mode, and
`--non-interactive` for automation.

**Repo:** https://github.com/m9rco/p4u-skill

## Binary status

!`which p4u 2>/dev/null && echo "✓ $(p4u --version 2>/dev/null || which p4u)" || echo "✗ p4u not found — run the install block below"`

## Install

**macOS / Linux** — one-liner, auto-detects OS and arch:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]') && ARCH=$(uname -m)
[[ "$ARCH" == "x86_64" ]] && ARCH=amd64 || ARCH=arm64
curl -fsSL "https://github.com/m9rco/p4u-skill/releases/latest/download/p4u-${OS}-${ARCH}" \
  -o /tmp/p4u && chmod +x /tmp/p4u && sudo mv /tmp/p4u /usr/local/bin/p4u
```

**Windows** (PowerShell):

```powershell
Invoke-WebRequest -Uri "https://github.com/m9rco/p4u-skill/releases/latest/download/p4u-windows-amd64.exe" `
  -OutFile "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\p4u.exe"
```

## Prerequisites

- `p4u` installed (see above)
- Perforce CLI (`p4`) installed and configured
- Logged in: `p4 login`

## Rules

1. **Always** pass `--non-interactive` — prevents hanging on prompts.
2. Pass `--json` when output needs to be parsed.
3. Changelist numbers are passed as plain integers, e.g. `12345`.
4. `p4u` auto-reads current user and client from `p4 info`.

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
p4u show -m 10 --non-interactive        # limit to 10 results
```

### Inspect a changelist

```bash
p4u show-cl <CL> --non-interactive
p4u show-cl <CL> --json --non-interactive
p4u show-cl -b <CL> --non-interactive   # brief: no file list
```

### Switch working context

Shelves current work, then unshelves the target changelist(s).

```bash
p4u switch <CL> --non-interactive
p4u switch <CL1> <CL2> --non-interactive
p4u switch -s -m <CL> --non-interactive   # sync + auto-resolve after
p4u switch -k <CL> --non-interactive      # keep shelved copy
p4u switch --non-interactive              # just shelve everything
```

### Delete a changelist

```bash
p4u delete-cl <CL> --non-interactive
p4u delete-cl -f <CL> --non-interactive   # force, no confirmation
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

### Annotate — who last changed a line

```bash
p4u annotate <depot-path> <line> --non-interactive
p4u annotate -v <depot-path> <line> --non-interactive  # verbose
p4u annotate --json <depot-path> <line> --non-interactive
```

### Delete a client

```bash
p4u delete-client --non-interactive
p4u delete-client -c <client> --non-interactive
p4u delete-client -f --non-interactive   # no confirmation
p4u delete-client -n --non-interactive   # keep local files
```

### Find untracked files

```bash
p4u untracked --non-interactive
p4u untracked ./src ./assets --non-interactive
p4u untracked -d 3 --non-interactive
p4u untracked --json --non-interactive
```

---

## Global flags

| Flag                | Short | Description                     |
|---------------------|-------|---------------------------------|
| `--non-interactive` |       | Disable all interactive prompts |
| `--json`            |       | JSON output                     |
| `--no-color`        | `-n`  | Disable colour                  |
| `--force-color`     | `-o`  | Force colour when piping        |
