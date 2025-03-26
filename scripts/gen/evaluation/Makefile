MODULE      := eval-tool
BINARY_NAME := eval-tool
OUTPUT_DIR  := bin
GOFLAGS     := -v
LDFLAGS     := -s -w

PLATFORMS := linux windows
OSARCH_linux   := GOOS=linux GOARCH=amd64
BINNAME_linux  := $(BINARY_NAME)
OSARCH_windows := GOOS=windows GOARCH=amd64
BINNAME_windows:= $(BINARY_NAME).exe

.PHONY: all clean $(PLATFORMS) test test-dev

all: help

define build-target
$(1):
	@echo "Building for $(1)..."
	@mkdir -p $(OUTPUT_DIR)
	$$(OSARCH_$(1)) go build $$(GOFLAGS) -ldflags="$$(LDFLAGS)" \
		-o $(OUTPUT_DIR)/$$(BINNAME_$(1)) ./cmd/eval-tool
endef

$(foreach platform,$(PLATFORMS),$(eval $(call build-target,$(platform))))

help:  ## Display targets with category headers
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "\n\033[1;34mUsage:\033[0m\n  make \033[36m<target>\033[0m\n\n\033[1;34mTargets:\033[0m\n"; \
	} \
	/^##@/ { \
		header = substr($$0, 5); \
		printf "\n\033[1;33m%s\033[0m\n", header; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  \033[36m%-12s\033[0m \033[90m%s\033[0m\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)

##@ Build
build: ## Build for all platforms
linux: ## Build Linux version
windows: ## Build Windows version

##@ Testing
test: ## Run production version
	./$(OUTPUT_DIR)/$(BINNAME_linux) \
		--input /mnt/nfs/rcabench_dataset/ts-ts-preserve-service-cpu-exhaustion-zcw5lh/result.csv \
		--output "$(shell pwd)/output/test.json" \
		--service service 

test-dev: ## Run development version
	@go run ./cmd/eval-tool \
		--input /mnt/nfs/rcabench_dataset/ts-ts-preserve-service-cpu-exhaustion-zcw5lh/result.csv \
		--output "$(shell pwd)/output/test.json" \
		--service service 

##@ Cleanup
clean: ## Clean build artifacts
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleanup completed"