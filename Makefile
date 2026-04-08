GOLANGCI_LINT_VERSION := $(shell tr -d '\r\n' < .golangci-lint-version)

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
