---
description: "Create a new stdix standard YAML file and push it to the registry. Use when you want to add a standard, write a rule file, or define coding standards for a language or topic."
name: "New stdix Standard"
argument-hint: "Describe the standard (e.g. 'Python FastAPI REST API conventions')"
agent: "agent"
tools: ["create_file", "read_file", "run_in_terminal"]
---

Create a new stdix standard YAML file and push it to the registry.

## Input

The user will provide:
- A short description of the standard (language, framework, topic)
- Optionally: an example file, existing code, or a list of conventions to encode

If no input is attached, ask the user for a description before proceeding.

## File path convention

```
standards/<language>/<topic>.yaml
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

## Steps

1. Determine `id`, `language`, and file path from the description.
2. If the user attached example files or code, read them and derive rules from the patterns you observe.
3. Write the YAML to a temporary path: `/tmp/<id>.yaml`
4. Push the standard to the registry:
   ```sh
   stdix push /tmp/<id>.yaml
   ```
   Requirements:
   - `STDIX_REGISTRY_TOKEN` env var must be set to a GitHub token with `contents:write` on the registry repo.
   - `registry.repo` must be set in `.stdix.yaml`, or use `--repo owner/stdix-registry`.
   - If the token is not set, ask the user to set it: `export STDIX_REGISTRY_TOKEN=<token>`
5. Confirm the push succeeded and tell the user:
   - The registry CI will rebuild `registry.db` automatically.
   - Once the release is available, run `stdix sync` then `stdix apply <id>` to use it.
6. Show the user the generated YAML and the `stdix push` output.
