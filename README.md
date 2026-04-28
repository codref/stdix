# stdix

**stdix** connects your project to a curated standards registry and materialises
the relevant rules into AI-agent instruction files:

- `AGENTS.md`
- `CLAUDE.md`
- `.github/copilot-instructions.md`
- `.cursor/rules/stdix.mdc`

---

## How it works

```
registry repo (YAML) ──► stdix-build build ──► registry.db artifact
                                                       │
                                           stdix sync (download)
                                                       │
                                           stdix match "my task"
                                                       │
                                           stdix apply <id>
                                                       │
                                    AGENTS.md / CLAUDE.md / copilot / cursor
```

1. A **registry repo** holds versioned YAML standard files.  
   `stdix-build` validates them on every PR and produces a `registry.db` JSON artifact on merge.

2. Developers run **`stdix`** locally (or in project CI) to download `registry.db`,
   search it, and inject matching rules into their agent files.

Applied rules are written inside managed blocks that are updated in place on subsequent runs:

```
<!-- stdix:start -->
...rules injected here...
<!-- stdix:end -->
```

---

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/stdix/stdix/main/install.sh | sh
```

Installs `stdix` to `$HOME/.local/bin`. To install globally:

```sh
sudo INSTALL_DIR=/usr/local/bin sh <(curl -fsSL https://raw.githubusercontent.com/stdix/stdix/main/install.sh)
```

The script detects your OS and architecture, downloads the matching binary from
the latest GitHub Release, verifies the SHA-256 checksum, and prints the next steps.

### Build from source

```sh
git clone https://github.com/stdix/stdix
cd stdix
make build          # produces bin/stdix and bin/stdix-build
```

Both binaries compile with `CGO_ENABLED=0` — no runtime dependencies.

---

## Quick start (project side)

```sh
# 1. Initialise stdix in your project
stdix init --registry-url https://github.com/codref/stdix-registry/releases/latest/download/registry.db

# 2. Download the latest registry.db
stdix sync

# 3. Find standards relevant to your work
stdix match "build a Python CLI app" --lang python

# 4. Apply a standard
stdix apply python.cli

# 5. Refresh all applied standards after a sync
stdix deploy

# 6. Verify your setup
stdix doctor
```

---

## Project configuration — `.stdix.yaml`

`stdix init` creates this file in your project root.

```yaml
registry:
  source: remote          # remote | local
  url: https://github.com/codref/stdix-registry/releases/latest/download/registry.db
  checksum: ""            # optional SHA-256 hex; verified on sync
  db: ""                  # override db path (used when source: local)
  repo: codref/stdix-registry  # GitHub repo for 'stdix push'

standards: []             # IDs applied to this project (managed by stdix apply)

outputs:
  agents: true            # AGENTS.md
  claude: true            # CLAUDE.md
  copilot: true           # .github/copilot-instructions.md
  cursor: true            # .cursor/rules/stdix.mdc
```

Use `source: local` and `db: ./testdata/registry.db` when developing against a
local registry.

---

## stdix commands

| Command | Description |
|---|---|
| `stdix init` | Create `.stdix.yaml` in the current directory |
| `stdix sync` | Download `registry.db` to `~/.cache/stdix/registry.db` |
| `stdix list` | List all standards in the local `registry.db` |
| `stdix match <query>` | Rank standards by BM25 relevance to a query |
| `stdix apply <id>` | Inject a standard's rules into agent files |
| `stdix deploy` | Re-apply all standards in `.stdix.yaml` to agent files |
| `stdix push <yaml>` | Push a standard YAML to the registry repo via GitHub API |
| `stdix doctor` | Check config, db, and registry reachability |

**`stdix match` flags**

```
--lang string    language bonus (e.g. python, go, node)
-n, --limit int  max results (default 10)
```

**`stdix sync` flags**

```
--registry-url string   URL of registry.db (http, https, or file://)
```

**`stdix push` flags**

```
--repo string      registry repo in owner/repo format (overrides registry.repo in .stdix.yaml)
--branch string    target branch (default "main")
--message string   commit message (default: "add <id> standard")
```

Requires `STDIX_REGISTRY_TOKEN` env var — a GitHub token with `contents:write` on the registry repo.

---

## Registry authoring

A registry is a directory with the following layout:

```
registry/
  standards/
    python/
      cli.yaml
      testing.yaml
    go/
      cli.yaml
    shared/
      logging.yaml
```

Each YAML file is a single standard:

```yaml
id: python.cli
title: Python CLI Standard
version: 1.0.0
language: python
tags:
  - cli
  - terminal
applies_when:
  - building a CLI application
rules:
  - Use argparse or click for argument parsing.
  - Always provide --help.
  - Exit with non-zero status on error.
outputs:
  agents: true
  claude: true
  copilot: true
  cursor: true
```

**Required fields:** `id`, `title`, `version` (semver `x.y.z`), `rules`  
**Optional:** `language`, `tags`, `applies_when`, `outputs`

### Contributing a new standard

The recommended flow uses the `/new-standard` Copilot prompt
(`.github/prompts/new-standard.prompt.md`) — the agent generates the YAML,
validates it, and pushes it to the registry repo in one step.

Manually:

```sh
# 1. Write your YAML
vim /tmp/python.fastapi.yaml

# 2. Validate it locally (optional)
./bin/stdix-build validate --registry /tmp

# 3. Push to the registry repo
export STDIX_REGISTRY_TOKEN=<your-github-token>
stdix push /tmp/python.fastapi.yaml --repo codref/stdix-registry

# 4. Wait for registry CI to rebuild registry.db, then sync
stdix sync
stdix list
```

The registry path is derived automatically from the `id` field:
`python.fastapi` → `standards/python/fastapi.yaml`

---

## stdix-build commands (registry CI)

### Validate — runs on every PR

```sh
stdix-build validate --registry ./standards
```

Checks required fields, semver format, and duplicate IDs. Exits non-zero on any error.

### Build — runs on merge

```sh
stdix-build build \
  --registry ./standards \
  --out registry.db \
  --version 1.2.0
```

Produces `registry.db` — a JSON file containing all standards pre-indexed for BM25 search.
Distribute as a release artifact or via a static URL.

---

## Search ranking

`stdix match` uses BM25 with weighted fields:

| Field | Weight |
|---|---|
| `language` | ×5 |
| `tags` | ×3 |
| `applies_when` | ×3 |
| `title` | ×2 |
| `rules` | ×1 |

An additional language bonus of **+50** is applied when a standard's `language`
matches the `--lang` flag.

---

## Development

```sh
make build    # compile both binaries → bin/
make test     # run all unit tests
make db       # validate registry + rebuild testdata/registry.db
make clean    # remove bin/
```
