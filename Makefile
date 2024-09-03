# Makefile

.PHONY: test cover clean

# Run all unit tests
test:
	go test ./... -v -count=1

# Generate and display test coverage
cover:
	sudo go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Clean up generated files
clean:
	rm -f coverage.out coverage.html

