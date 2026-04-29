BINARY_DIR := bin
STDIX       := $(BINARY_DIR)/stdix
STDIX_BUILD := $(BINARY_DIR)/stdix-build

GO      := go
GOFLAGS := CGO_ENABLED=0
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/codref/stdix/internal/version.Version=$(VERSION)"

REGISTRY_DIR    := testdata/stdix-registry
REGISTRY_DB     := testdata/registry.db
REGISTRY_VER    := 1.0.0

# Remote registry — override on the command line or in your shell:
#   make registry-clone REGISTRY_REMOTE=git@github.com:org/stdix-registry.git
REGISTRY_REMOTE := git@github.com:codref/stdix-registry.git
REGISTRY_LOCAL  := ../stdix-registry

.PHONY: all build test db clean release registry-clone registry-pull registry-push registry-db

all: build

build: $(STDIX) $(STDIX_BUILD)

$(STDIX): $(shell find cmd/stdix internal -name '*.go')
	@mkdir -p $(BINARY_DIR)
	$(GOFLAGS) $(GO) build $(LDFLAGS) -o $@ ./cmd/stdix

$(STDIX_BUILD): $(shell find cmd/stdix-build internal -name '*.go')
	@mkdir -p $(BINARY_DIR)
	$(GOFLAGS) $(GO) build $(LDFLAGS) -o $@ ./cmd/stdix-build

test:
	$(GOFLAGS) $(GO) test ./...

db: $(STDIX_BUILD)
	$(STDIX_BUILD) validate --registry $(REGISTRY_DIR)
	$(STDIX_BUILD) build --registry $(REGISTRY_DIR) --out $(REGISTRY_DB) --version $(REGISTRY_VER)

# ── Registry helpers ────────────────────────────────────────────────────────

# Clone the remote registry repo once
registry-clone:
	@if [ -d $(REGISTRY_LOCAL) ]; then \
		echo "$(REGISTRY_LOCAL) already exists — run 'make registry-pull' to update."; \
	else \
		git clone $(REGISTRY_REMOTE) $(REGISTRY_LOCAL); \
	fi

# Pull latest changes from the remote registry
registry-pull:
	git -C $(REGISTRY_LOCAL) pull --ff-only

# Rebuild registry.db from the cloned remote registry
registry-db: $(STDIX_BUILD)
	$(STDIX_BUILD) validate --registry $(REGISTRY_LOCAL)
	$(STDIX_BUILD) build --registry $(REGISTRY_LOCAL) --out $(REGISTRY_DB) --version $(REGISTRY_VER)

# Commit and push the current state of the local registry clone
# Usage: make registry-push MSG="add python.fastapi standard"
registry-push:
	git -C $(REGISTRY_LOCAL) add -A
	git -C $(REGISTRY_LOCAL) commit -m "$(MSG)"
	git -C $(REGISTRY_LOCAL) push

# ── Release ─────────────────────────────────────────────────────────────────

# Tag and push a new version.
# Usage: make release V=v0.2.0
release:
	@[ -n "$(V)" ] || (echo "usage: make release V=v<major>.<minor>.<patch>"; exit 1)
	@git diff --quiet && git diff --cached --quiet || (echo "error: working tree is dirty — commit or stash first"; exit 1)
	git tag $(V)
	git push origin $(V)
	@echo "Tagged and pushed $(V)"

clean:
	rm -rf $(BINARY_DIR)
