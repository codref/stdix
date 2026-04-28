---
description: "Create a new stdix standard YAML file for the registry. Use when you want to add a standard, write a rule file, or define coding standards for a language or topic."
name: "New stdix Standard"
argument-hint: "Describe the standard (e.g. 'Python FastAPI REST API conventions')"
agent: "agent"
tools: ["create_file", "read_file", "run_in_terminal"]
---

Create a new stdix standard YAML file, commit it to the registry repo, and rebuild the local database.

## Input

The user will provide:
- A short description of the standard (language, framework, topic)
- Optionally: an example file, existing code, or a list of conventions to encode

If no input is attached, ask the user for a description before proceeding.

## Registry location

Determine where to write the file:

1. Check if `../stdix-registry/standards/` exists (the cloned remote registry).
2. If it exists, write there → this is the **remote registry** path.
3. If it does not exist, fall back to `testdata/stdix-registry/standards/` → the **local testdata** path.

If using the remote registry path, you will commit and push at the end.

### File path convention

```
<registry-root>/standards/<language>/<topic>.yaml
```

Use `shared/` as the language folder when the standard is language-agnostic.

## Schema

```yaml
id: <language>.<topic>          # dot-namespaced, lowercase, unique
title: <Human Readable Title>
version: 1.0.0
language: <python|go|node|...>  # omit for language-agnostic standards
tags:
  - <tag1>
  - <tag2>
applies_when:
  - <phrase describing when this standard applies>
rules:
  - <Actionable rule sentence. One behaviour per rule. Use imperative mood.>
outputs:
  agents: true
  claude: true
  copilot: true
  cursor: true
```

**Required:** `id`, `title`, `version` (semver `x.y.z`), `rules`
**Optional:** `language`, `tags`, `applies_when`, `outputs`

## Rules — writing guidelines

- One concrete behaviour per rule (not a category or heading).
- Imperative mood: "Use X", "Avoid Y", "Always Z".
- Specific enough to be actionable by an AI coding agent.
- 6–12 rules is a good range; avoid padding.
- If the user provided example code or conventions, extract rules directly from them.
- Quote rule strings that contain `:` to keep YAML valid.

## Example output

[testdata/stdix-registry/standards/python/cli.yaml](../../testdata/stdix-registry/standards/python/cli.yaml)

## Steps

1. Determine `id`, `language`, and file path from the description.
2. If the user attached example files or code, read them and derive rules from the patterns you observe.
3. Write the YAML file locally to a temporary path: `/tmp/<id>.yaml`
   (or `testdata/stdix-registry/standards/<language>/<topic>.yaml` if the user wants to keep it locally).
4. Run `make db` to validate the YAML and rebuild the local `registry.db`:
   ```sh
   make db
   ```
   If validation fails, fix the YAML and retry before proceeding.
5. Push the standard to the remote registry repository:
   ```sh
   ./bin/stdix push /tmp/<id>.yaml
   ```
   Requirements:
   - `STDIX_REGISTRY_TOKEN` env var must be set to a GitHub token with `contents:write` on the registry repo.
   - `registry.repo` must be set in `.stdix.yaml` (e.g. `owner/stdix-registry`), or use `--repo owner/stdix-registry`.
   - If the token is not set, ask the user to set it: `export STDIX_REGISTRY_TOKEN=<token>`
6. Confirm the push succeeded and tell the user:
   - The registry CI will rebuild `registry.db` automatically.
   - Once the release is available, run `stdix sync` to update the local cache.
7. Show the user the generated YAML and the `stdix push` output.
