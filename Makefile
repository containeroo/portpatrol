# Detect platform for sed compatibility
SED := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo gsed; else echo sed; fi)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Documentation Tools
LORE ?= $(LOCALBIN)/lore

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.1.2
# renovate: datasource=github-releases depName=gi8lino/lore
LORE_VERSION ?= v0.3.0

## Site Configuration
SITE_CONFIG ?= docs/lore-site.toml
SITE_OUTPUT ?= docs/site
SITE_PORT ?= 8081
SITE_LOGO ?= docs/assets/logo.svg
SITE_FAVICON_SVG ?= docs/content/assets/favicon.svg
SITE_FAVICON_ICO ?= docs/content/favicon.ico
MAGICK ?= magick

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: download
download: ## Download go packages
	go mod download

.PHONY: update-packages
update-packages: ## Update all Go packages to their latest versions
	go get -u ./...
	go mod tidy

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run unit tests.
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...

.PHONY: cover
cover: ## Display test coverage
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Clean up generated files
	rm -f coverage.out coverage.html

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes.
	$(GOLANGCI_LINT) run --fix

.PHONY: site-favicon
site-favicon: ## Generate the documentation SVG favicon and multi-size ICO (requires ImageMagick).
	@command -v "$(MAGICK)" >/dev/null 2>&1 || { echo "ImageMagick is required; install it and set MAGICK if needed." >&2; exit 1; }
	@test -f "$(SITE_LOGO)" || { echo "SVG logo not found: $(SITE_LOGO)" >&2; exit 1; }
	mkdir -p "$(dir $(SITE_FAVICON_SVG))" "$(dir $(SITE_FAVICON_ICO))"
	"$(MAGICK)" -background none -density 384 "$(SITE_LOGO)" \
		-resize 256x256 -gravity center -extent 256x256 \
		-define icon:auto-resize=256,128,64,48,32,16 "$(SITE_FAVICON_ICO)"
	cp "$(SITE_LOGO)" "$(SITE_FAVICON_SVG)"

.PHONY: site
site: lore ## Build the static documentation site with Lore.
	"$(LORE)" build --config "$(SITE_CONFIG)"

.PHONY: site-serve
site-serve: lore ## Build and serve the documentation site locally.
	"$(LORE)" build --config "$(SITE_CONFIG)" \
		--site-url "http://127.0.0.1:$(SITE_PORT)/"
	@echo "Serving Lore documentation at http://127.0.0.1:$(SITE_PORT)"
	python3 -m http.server $(SITE_PORT) --bind 127.0.0.1 --directory "$(SITE_OUTPUT)"

##@ Tagging

# Find the latest tag (with prefix filter if defined, default to 0.0.0 if none found)
# Lazy evaluation ensures fresh values on every run
VERSION_PREFIX ?= v
LATEST_TAG = $(shell git tag --list "$(VERSION_PREFIX)*" --sort=-v:refname | head -n 1)
VERSION = $(shell [ -n "$(LATEST_TAG)" ] && echo $(LATEST_TAG) | sed "s/^$(VERSION_PREFIX)//" || echo "0.0.0")

patch: ## Create a new patch release (x.y.Z+1)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

minor: ## Create a new minor release (x.Y+1.0)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.0", $$1, $$2+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

major: ## Create a new major release (X+1.0.0)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.0.0", $$1+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

tag: ## Show latest tag
	@echo "Latest version: $(LATEST_TAG)"

push: ## Push tags to remote
	git push --tags


##@ Dependencies

.PHONY: lore
lore: ## Download the pinned Lore binary into bin if necessary.
ifeq ($(LORE),$(LOCALBIN)/lore)
	@set -eu; \
	case "$$(uname -s)" in Darwin) os=darwin ;; Linux) os=linux ;; *) echo "Unsupported Lore operating system" >&2; exit 1 ;; esac; \
	case "$$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) echo "Unsupported Lore architecture" >&2; exit 1 ;; esac; \
	version="$(LORE_VERSION)"; \
	binary="lore-$${version}-$${os}-$${arch}"; \
	mkdir -p "$(LOCALBIN)"; \
	if [ ! -x "$(LOCALBIN)/$$binary" ]; then \
		tmp=$$(mktemp -d "$(LOCALBIN)/.lore-download.XXXXXX"); \
		trap 'rm -rf "$$tmp"' EXIT HUP INT TERM; \
		archive="lore_$${version#v}_$${os}_$${arch}.tar.gz"; \
		release="https://github.com/gi8lino/lore/releases/download/$$version"; \
		echo "Downloading Lore $$version ($$os/$$arch)"; \
		curl --fail --silent --show-error --location --retry 3 -o "$$tmp/$$archive" "$$release/$$archive"; \
		curl --fail --silent --show-error --location --retry 3 -o "$$tmp/checksums.txt" "$$release/lore_$${version#v}_checksums.txt"; \
		expected=$$(awk -v archive="$$archive" '$$2 == archive { print $$1; found = 1 } END { if (!found) exit 1 }' "$$tmp/checksums.txt"); \
		if command -v sha256sum >/dev/null 2>&1; then \
			actual=$$(sha256sum "$$tmp/$$archive" | awk '{print $$1}'); \
		else \
			actual=$$(shasum -a 256 "$$tmp/$$archive" | awk '{print $$1}'); \
		fi; \
		[ "$$actual" = "$$expected" ] || { echo "Lore checksum mismatch" >&2; exit 1; }; \
		tar -xzf "$$tmp/$$archive" -C "$$tmp" lore; \
		chmod +x "$$tmp/lore"; \
		mv "$$tmp/lore" "$(LOCALBIN)/$$binary"; \
	fi; \
	ln -sfn "$$binary" "$(LORE)"
else
	@command -v "$(LORE)" >/dev/null 2>&1 || { echo "Lore binary not found: $(LORE)" >&2; exit 1; }
endif

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef
