# Detect platform for sed compatibility
SED := $(shell if [ "$(shell uname)" = "Darwin" ]; then echo gsed; else echo sed; fi)

# VERSION defines the project version, extracted from cmd/portpatrol/main.go without leading 'v'.
VERSION := $(shell awk -F'"' '/const version/{gsub(/^v/, "", $$2); print $$2}' cmd/portpatrol/main.go)

.PHONY: test cover clean update patch minor major tag

# Run all unit tests
test:
	go test ./... -v -count=1

# Generate and display test coverage
cover:
	go test ./cmd/... ./internal/... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Clean up generated files
clean:
	rm -f coverage.out coverage.html

# Update dependencies
update:
	go get -u ./...
	go mod tidy

##@ Versioning

patch: ## Increment the patch version (x.y.Z -> x.y.(Z+1)).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1"."$$2"."$$3+1}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

minor: ## Increment the minor version (x.Y.z -> x.(Y+1).0).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1"."$$2+1".0"}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

major: ## Increment the major version (X.y.z -> (X+1).0.0).
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{print $$1+1".0.0"}') && \
	$(SED) -i -E "s/(const version string = \"v)[^\"]+/\1$${NEW_VERSION}/" cmd/portpatrol/main.go

tag: ## Tag the current commit with the current version if no tag exists and the repository is clean.
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Repository has uncommitted changes. Please commit or stash them before tagging."; \
		exit 1; \
	fi
	@if [ -z "$$(git tag --list v$(VERSION))" ]; then \
		echo "Tagging version v$(VERSION)"; \
		git tag "v$(VERSION)"; \
		git push origin "v$(VERSION)"; \
	else \
		echo "Tag v$(VERSION) already exists."; \
	fi

