## Plan: stdix Go CLI with Pre-built Registry Database

**TL;DR** — Pure-Go CLI with Cobra. The workflow is split in two: a **CI builder** (`stdix-build`) runs in the registry repo and emits a pre-built `registry.db` (JSON) artifact; the **client** (`stdix sync`) downloads that artifact and queries it locally with BM25 keyword ranking. No CGo, no native libs, no model downloads, no remote APIs.

---

### Architecture: CI Builder + Client Split

```
Registry repo (CI)                    Client machine
─────────────────────────────         ──────────────────────────────────
standards/python/cli.yaml  ─┐
standards/go/cli.yaml       ├─ stdix-build ──► registry.db (JSON artifact)
standards/shared/logging.yaml─┘                      │
                                                      │ stdix sync (download)
                                                      ▼
                                           ~/.cache/stdix/registry.db
                                                      │
                                           stdix match "build a Python CLI"
                                           (BM25 over local registry.db)
```

**Why this works:**
- The CI builder can use any tooling; CGo is irrelevant there.
- The client only needs to embed the user's query text against pre-indexed terms — pure-Go BM25, zero native deps.
- `registry.db` is a single versioned artifact (GitHub release, raw URL, or committed to the registry repo); the client pins a version via config.

---

### Phase 1 — Project Bootstrap

1. Initialize Go module `github.com/yourorg/stdix`
2. Add dependencies: `cobra`, `gopkg.in/yaml.v3`, inline BM25/TF-IDF (pure Go — no CGo)
3. Directory layout:

```
stdix/
  cmd/stdix/main.go          ← client CLI
  cmd/stdix-build/main.go    ← CI builder (separate binary)
  internal/config/
  internal/registry/         ← YAML parsing (shared by builder and client)
  internal/search/           ← BM25 scorer (client-side query matching)
  internal/db/               ← registry.db read/write (JSON)
  testdata/stdix-registry/
```

### Phase 2 — Fake Registry Root *(parallel with Phase 1)*

4. Create `testdata/stdix-registry/` with `index.yaml` and three sample standards matching the schema in [docs/plan.md](docs/plan.md): `standards/python/cli.yaml`, `standards/go/cli.yaml`, `standards/shared/logging.yaml`
5. This directory is the repo root stand-in; `stdix-build` and the config `registry.path` both point here

### Phase 3 — Config + `stdix init` *(depends on 1)*

6. `internal/config/config.go` — Go struct for `.stdix.yaml` (registry db URL/path, outputs flags, applied standards list)
7. `stdix init` — writes `.stdix.yaml` to CWD; prompts before overwrite. Dev default: `registry.source: local`, `registry.db: ./testdata/registry.db`

### Phase 4 — Registry Parsing *(depends on 2)*

8. `internal/registry/standard.go` — Go struct: `id`, `title`, `version`, `language`, `tags`, `applies_when`, `rules`
9. YAML loader: walk `standards/**/*.yaml`, parse, validate required fields, detect duplicate IDs
10. Unit tests with valid + invalid YAML fixtures

### Phase 5 — Registry DB Format *(depends on 4)*

11. `internal/db/db.go` — defines `DB` struct: metadata (registry version, built-at timestamp) + slice of `IndexedStandard` (all parsed fields + pre-tokenized term lists for BM25)
12. `db.Write(path, db)` — serializes to JSON; `db.Read(path)` — deserializes
13. The DB file is the sole artifact passed from CI to client; it is human-readable JSON

### Phase 6 — CI Builder `stdix-build` *(depends on 4, 5)*

14. `cmd/stdix-build/main.go` — reads registry path from flag, parses all YAMLs, builds `DB`, writes `registry.db`
15. Intended to run in GitHub Actions (or any CI) inside the registry repo; output artifact is committed or uploaded as a release asset
16. Print build summary: "Built registry.db: N standards, version X"

### Phase 7 — BM25 Search *(depends on 5)*

17. `internal/search/bm25.go` — pure-Go BM25 scorer; tokenizes query and scores each standard using title, tags, `applies_when`, and language fields (weights match the scoring model in [docs/plan.md](docs/plan.md))
18. `internal/search/scorer.go` — `Score(query string, standards []IndexedStandard) []Result`; applies language bonus when `--lang` flag matches
19. Unit tests: assert `python.cli` scores highest for "build a Python CLI app"

### Phase 8 — `stdix sync` *(depends on 3, 5)*

20. `stdix sync` — downloads `registry.db` from the configured URL (or copies from local path); stores at `~/.cache/stdix/registry.db`; verifies SHA-256 checksum if provided in config
21. Flags: `--registry-url`, `--pin <ref>`
22. No parsing, no indexing, no embedding — just fetch + checksum

### Phase 9 — `stdix match` *(depends on 7, 8)*

23. `stdix match "<query>"` — loads `~/.cache/stdix/registry.db`, runs BM25 scorer, prints ranked table: ID / title / score / first matched `applies_when` phrase
24. Flags: `--limit N`, `--lang go|python|...`
25. Integration test: build testdata registry, assert `stdix match "build a Python CLI"` returns `python.cli` first

### Phase 10 — Supporting Commands *(depends on 8)*

26. `stdix list` — loads local `registry.db`, prints standards table (ID, title, language, version)
27. `stdix doctor` — checks: `.stdix.yaml` exists, local `registry.db` present and readable, registry URL reachable (non-fatal if offline)

---

**Relevant files**

- [docs/plan.md](docs/plan.md) — Phase 2 YAML schema, Phase 4 keyword scoring model
- `testdata/stdix-registry/` — fake registry root (to be created)
- `testdata/registry.db` — pre-built artifact from `stdix-build` for integration tests
- `~/.cache/stdix/registry.db` — runtime local database

**Verification**

1. `go test ./internal/registry/...` — YAML parsing + validation errors
2. `go test ./internal/search/...` — BM25 scorer ranks correctly; language bonus applied
3. `go test ./internal/db/...` — round-trip write/read of registry.db
4. `stdix-build --registry ./testdata/stdix-registry --out ./testdata/registry.db` prints "Built 3 standards"
5. `stdix init` in a temp dir produces a valid `.stdix.yaml`
6. `stdix sync --registry-url file://./testdata/registry.db` stores db locally
7. `stdix match "build a Python CLI app"` returns `python.cli` ranked first
8. `stdix doctor` exits 0 on valid setup; reports missing db without crashing
9. `go build ./...` — no CGo, confirmed with `CGO_ENABLED=0 go build ./...`

**Decisions**

- **No CGo, no native libs** — `hugot`/onnxruntime and `chromem-go` are dropped entirely; pure-Go BM25 is sufficient for the keyword scoring model in `plan.md`
- **CI builder + client split** — `stdix-build` runs in the registry repo CI; the client never parses raw YAMLs; this decouples registry tooling from client portability requirements
- **`registry.db` as JSON** — human-readable, diffable, no binary format dependencies; can be committed to the registry repo or published as a release asset
- **BM25 scoring** — aligns with the weighted model already defined in `plan.md` (language match, tag match, `applies_when` match, title match); no semantic embeddings needed for MVP
- **Checksum verification** on sync — `registry.db` URL + SHA-256 stored in `.stdix.yaml`; client rejects tampered files
- Fake registry lives at `testdata/stdix-registry/`; `testdata/registry.db` is its pre-built artifact — integration tests use both
