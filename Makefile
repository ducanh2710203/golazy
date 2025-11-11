# Makefile for github.com/duhnnie/lazystruct

GO              := go
PKGS            := $(shell $(GO) list ./...)
COVER_PROFILE   := coverage.out
COVER_HTML      := coverage.html

.PHONY: all build test cover fmt lint tidy ci clean

# -------------------------------------------------------
# Main commands
# -------------------------------------------------------

all: build

build:
	@echo "🔨 Building..."
	@$(GO) build ./...

test:
	@echo "🧪 Running tests..."
	@$(GO) test ./... -v

cover:
	@echo "🧮 Running tests with coverage..."
	@$(GO) test ./... -coverprofile=$(COVER_PROFILE) -v
	@$(GO) tool cover -func=$(COVER_PROFILE)
	@echo "📊 Generate HTML report: $(COVER_HTML)"
	@$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

fmt:
	@echo "🎨 Formatting code..."
	@$(GO) fmt ./...

lint:
	@echo "🔍 Linting..."
	@$(GO) vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "⚠️ staticcheck not found (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)"; \
	fi

tidy:
	@echo "🧹 Tidying modules..."
	@$(GO) mod tidy

clean:
	@echo "🧼 Cleaning up..."
	@rm -f $(COVER_PROFILE) $(COVER_HTML)

ci: fmt lint test
	@echo "✅ All checks passed!"

# -------------------------------------------------------
# Release commands
# -------------------------------------------------------

VERSION ?=
BUMP ?= patch  # can be: patch, minor, or major

# Derive current version (most recent tag)
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Calculate next version
define bump_version
	OLD=$(1); \
	if [ "$(BUMP)" = "major" ]; then \
		NEW=$$(echo $$OLD | awk -F. '{printf "v%d.0.0", substr($$1,2)+1}'); \
	elif [ "$(BUMP)" = "minor" ]; then \
		NEW=$$(echo $$OLD | awk -F. '{printf "v%d.%d.0", substr($$1,2), $$2+1}'); \
	else \
		NEW=$$(echo $$OLD | awk -F. '{printf "v%d.%d.%d", substr($$1,2), $$2, $$3+1}'); \
	fi; \
	echo $$NEW
endef

next-version:
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next version: $$( $(call bump_version,$(CURRENT_VERSION)) )"

release:
	@echo "🚀 Releasing new $(BUMP) version..."
	@NEW_VERSION=$$( $(call bump_version,$(CURRENT_VERSION)) ); \
	echo "Tagging $$NEW_VERSION"; \
	git tag $$NEW_VERSION; \
	git push origin $$NEW_VERSION; \
	echo "✅ Released $$NEW_VERSION"

# | Command                   | Description                                  |
# | ------------------------- | -------------------------------------------- |
# | `make next-version`       | Shows what the next tag would be             |
# | `make release`            | Creates a **patch** release (e.g., `v1.0.1`) |
# | `make release BUMP=minor` | Creates a **minor** release (e.g., `v1.1.0`) |
# | `make release BUMP=major` | Creates a **major** release (e.g., `v2.0.0`) |
