.PHONY: help build run clean install test lint fmt

# Binary name
BINARY_NAME=kefbar

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Main package path
MAIN_PACKAGE=./cmd/kefbar

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run the application
	./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run without building (faster iteration)
	$(GOCMD) run $(MAIN_PACKAGE)

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

install: build ## Install to /usr/local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

uninstall: ## Uninstall from /usr/local/bin
	rm -f /usr/local/bin/$(BINARY_NAME)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@if ! command -v $(GOLINT) &> /dev/null; then \
		echo ""; \
		echo "❌ golangci-lint is not installed!"; \
		echo ""; \
		echo "To install, run one of the following:"; \
		echo ""; \
		echo "  # Using Homebrew (macOS):"; \
		echo "  brew install golangci-lint"; \
		echo ""; \
		echo "  # Using Go:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo ""; \
		echo "  # Using curl:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		echo ""; \
		exit 1; \
	fi
	$(GOLINT) run ./...

fmt: ## Format code
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# macOS app bundle (optional)
APP_NAME=KEF Bar.app
BUNDLE_ID=com.kefbar.app

app: build ## Create macOS app bundle
	@mkdir -p "$(BUILD_DIR)/$(APP_NAME)/Contents/MacOS"
	@mkdir -p "$(BUILD_DIR)/$(APP_NAME)/Contents/Resources"
	@cp $(BUILD_DIR)/$(BINARY_NAME) "$(BUILD_DIR)/$(APP_NAME)/Contents/MacOS/"
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<plist version="1.0"><dict>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<key>CFBundleExecutable</key><string>$(BINARY_NAME)</string>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<key>CFBundleIdentifier</key><string>$(BUNDLE_ID)</string>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<key>CFBundleName</key><string>KEF Bar</string>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<key>CFBundlePackageType</key><string>APPL</string>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '<key>LSUIElement</key><true/>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo '</dict></plist>' >> "$(BUILD_DIR)/$(APP_NAME)/Contents/Info.plist"
	@echo "Created: $(BUILD_DIR)/$(APP_NAME)"
