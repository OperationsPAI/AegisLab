# =============================================================================
# RCABench Makefile
# =============================================================================
# This Makefile provides all build, deployment, and development tools for the RCABench project
# Use 'make help' to view all available commands

# =============================================================================
# Configuration Loading
# =============================================================================

# Load environment-specific configuration
-include .env

# Basic Configuration
# Defines the environment mode: 'prod' (default, compiled binary) or 'dev', 'test' (interpreted script).
ENV_MODE ?= dev

DEFAULT_REPO 	?= docker.io/opspai
NS          	:= exp
PORT        	:= 30080
RELEASE_NAME    := rcabench

# COMMAND Tool Configuration
COMMAND_DIR := scripts/command
COMMAND := UV_WITH_GROUPS=dev uv run --project $(COMMAND_DIR) python $(COMMAND_DIR)/main.py

# Directory Configuration
LEFTHOOK_CONFIG := lefthook.yml
SRC_DIR := src

CLIENT_VERSION ?= 0.0.0
SDK_VERSION ?=0.0.0

# Color definitions
BLUE    := \033[1;34m
GREEN   := \033[1;32m
YELLOW  := \033[1;33m
RED     := \033[1;31m
CYAN    := \033[1;36m
GRAY    := \033[90m
RESET   := \033[0m

# Dependency Repositories
CHAOS_EXPERIMENT_REPO := github.com/OperationsPAI/chaos-experiment@injectionv2-dev

# ===================================================·==========================
# Declare all non-file targets
# =============================================================================
.PHONY: help build run debug swagger import clean-finalizers delete-all-chaos k8s-resources ports \
        pre-commit deploy-ts swag-init generate-sdk release \
        check-prerequisites setup-dev-env clean-all status logs

# =============================================================================
# Default Goal
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# Help Information
# =============================================================================
help:  ## 📖 Display all available commands
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║               RCABench Project Management Tool               ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"
	@printf "$(YELLOW)Usage:$(RESET) make $(CYAN)<target>$(RESET)\n"
	@printf "$(YELLOW)Examples:$(RESET)\n make run, make help, make clean-all \n"
	@printf "\n"
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "$(YELLOW)Available commands:$(RESET)\n"; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  $(CYAN)%-25s$(RESET) $(GRAY)%s$(RESET)\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)
	@printf "\n"
	@printf "$(YELLOW)Quick start:$(RESET)\n"
	@printf "  $(CYAN)make check-prerequisites$(RESET) - Check environment dependencies\n"
	@printf "  $(CYAN)make run$(RESET)                 - Build and deploy application\n"
	@printf "  $(CYAN)make status$(RESET)              - View application status\n"
	@printf "  $(CYAN)make logs$(RESET)                - View application logs\n"

# =============================================================================
# Command Tool Management
# =============================================================================

run-command: ## 🔧 Run command tool (usage: make run-command ARGS="your args")
	@$(COMMAND) $(ARGS)

# =============================================================================
# Environment Check and Setup
# =============================================================================

check-prerequisites: ## 🔍 Check development environment dependencies
	@printf "$(BLUE)🔍 Checking development environment dependencies...$(RESET)\n"
	@printf "$(GRAY)Checking devbox...$(RESET)\n"
	@command -v devbox >/dev/null 2>&1 || { printf "$(RED)❌ devbox not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ devbox installed$(RESET)\n"
	@devbox install >/dev/null 2>&1 || { printf "$(RED)❌ devbox dependencies installation failed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ devbox dependencies installed$(RESET)\n"
	@printf "$(GRAY)Checking docker...$(RESET)\n"
	@command -v docker >/dev/null 2>&1 || { printf "$(RED)❌ docker not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ docker installed$(RESET)\n"
	@printf "$(GRAY)Checking helm...$(RESET)\n"
	@command -v helm >/dev/null 2>&1 || { printf "$(RED)❌ helm not insalled$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ helm installed$(RESET)\n"
	@printf "$(GRAY)Checking kubectx...$(RESET)\n"
	@command -v kubectx >/dev/null 2>&1 || { printf "$(RED)❌ kubectx not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ kubectx installed$(RESET)\n"
	@printf "$(GREEN)🎉 All dependency checks passed!$(RESET)\n\n"

forward-ports: ## 🔗 Start port forwarding to access application
	@$(MAKE) run-command ARGS="port start -e $(ENV_MODE) -n $(NS)"

setup-dev-env: check-prerequisites ## 🛠️  Setup development environment
	@printf "$(BLUE)🛠️  Setting up development environment...$(RESET)\n"
	@printf "$(GRAY)Checking for 'uv' installation...$(RESET)\n"
	@if command -v uv &> /dev/null; then \
		printf "$(GREEN)✅ 'uv' found in PATH$(RESET)\n"; \
	else \
		printf "$(YELLOW)Warning: 'uv' not found. Installing via script...$(RESET)\n"; \
		curl -LsSf https://astral.sh/uv/install.sh | sh; \
		printf "$(GREEN)✅ 'uv' installed!$(RESET)\n"; \
	fi
	@printf "$(GRAY)Applying Kubernetes manifests for local development...$(RESET)\n"
	kubectl apply -f manifests/local-dev/exp-dev-setup.yaml
	@printf "$(GRAY)Installing Git hooks with Lefthook...$(RESET)\n"
	@if test -f $(LEFTHOOK_CONFIG); then \
		devbox run install-hooks; \
		printf "$(GREEN)✅ Lefthook hooks installed successfully!$(RESET)\n"; \
	else \
		printf "$(RED)❌ lefthook.yml not found$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(GREEN)✅ Development environment setup completed!$(RESET)\n"

setup-test-env: check-prerequisites ## 🧪 Setup test environment
	@printf "$(BLUE)🧪 Setting up test environment...$(RESET)\n"
	@printf "$(GRAY)Executing test environment setup script...$(RESET)\n"
	bash ./manifests/test/start.sh
	@printf "$(GREEN)✅ Test environment setup completed!$(RESET)\n"

# =============================================================================
# Pedestal Function
# =============================================================================

install-pedestals: ## 🔍 Install pedestals in namespaces (usage: make install-releases PEDESTAL_NAME=ts PEDESTAL_COUNT=2)
	$(MAKE) run-command ARGS="pedestal install -e $(ENV_MODE) -n $(PEDESTAL_NAME) -c $(PEDESTAL_COUNT) -f"

# =============================================================================
# Build and Deployment
# =============================================================================

install-openebs:
	@printf "$(BLUE)Deploying OpenEBS...$(RESET)\n"
	helm upgrade -i openebs openebs/openebs --namespace openebs \
  		--create-namespace \
		--values ./manifests/staging/openebs.yaml \
		--atomic --timeout 10m
	@printf "$(GREEN)✅ OpenEBS installed successfully$(RESET)\n\n"

install-rcabench:  ## 🔧 Deploy RCABench application in prod environment
	@printf "$(BLUE)Deploying RCABench application...$(RESET)\n"
	helm upgrade -i rcabench ./helm --namespace exp \
		--create-namespace \
		--values ./manifests/prod/rcabench.yaml \
		--set-file initialDataFiles.data_yaml=data/initial_data/prod/data.yaml \
		--set-file initialDataFiles.otel_demo_yaml=data/initial_data/prod/otel-demo.yaml \
		--set-file initialDataFiles.ts_yaml=data/initial_data/prod/ts.yaml \
		--atomic --timeout 10m
	@printf "$(GREEN)✅ RCABench installed successfully$(RESET)\n\n"
	@printf "$(BLUE)🔗 Starting automatic port forwarding...$(RESET)\n"
	@$(MAKE) forward-ports ENV_MODE=prod

local-deploy: ## 🛠️  Setup local development environment with basic services
	$(MAKE) run-command ARGS="rcabench local-deploy -f"
	$(MAKE) init-etcd ENV_MODE=dev

run: check-prerequisites ## 🚀 Build and deploy application (using skaffold)
	ENV_MODE=staging devbox run skaffold run

init-etcd:
	$(MAKE) run-command ARGS="etcd init -e $(ENV_MODE) -f"

# =============================================================================
# Test
# =============================================================================

test:
	SDK_VERSION="$(SDK_VERSION)" ENV_MODE=test devbox run skaffold run

regression-test:
	chmod +x ./scripts/regression-test.sh && ./scripts/regression-test.sh

# =============================================================================
# Development Tools
# =============================================================================

local-debug:  ## 🐛 Start local debugging environment
	@printf "$(BLUE)⌛️ Starting local application...$(RESET)\n"; \
	cd $(SRC_DIR) && go run main.go both --port 8082

update-dependencies: ## 📦 Update latest version of dependencies
	@printf "$(BLUE)📦 Updating latest version of chaos-experiment library...$(RESET)\n"
	cd $(SRC_DIR) && go get -u $(CHAOS_EXPERIMENT_REPO) && go mod tidy
	@printf "$(GREEN)✅ Dependencies update completed$(RESET)\n"

# =============================================================================
# Chaos Management
# =============================================================================

clean-finalizers: ## 🧹 Clean all chaos resource finalizers in namespaces with specific prefix
	$(MAKE) run-command ARGS="chaos clean-finalizers -e $(ENV_MODE) -p $(NS_PREFIX) -c $(NS_COUNT)"

delete-chaos: ## 🗑️  Delete chaos resources in namespaces with specific prefix
	$(MAKE) run-command ARGS="chaos delete-resources -e $(ENV_MODE) -p $(NS_PREFIX) -c $(NS_COUNT)"

# =============================================================================
# Git Hooks
# =============================================================================

pre-commit:
	$(MAKE) run-command ARGS="git pre-commit"

# =============================================================================
# SDK Generation
# =============================================================================

swag-init: ## 📝 Initialize Swagger documentation
	$(MAKE) run-command ARGS="swagger init -v $(SDK_VERSION)"

generate-typescript-client: swag-init ## ⚙️ Generate TypeScript Client from Swagger documentation
	$(MAKE) run-command ARGS="swagger generate-client -l typescript -v $(CLIENT_VERSION)"

generate-python-sdk: swag-init ## ⚙️ Generate Python SDK from Swagger documentation
	$(MAKE) run-command ARGS="swagger generate-sdk -l python -v $(SDK_VERSION)"

# =============================================================================
# Utilities
# =============================================================================

changelog: ## 📝 Generate CHANGELOG.md (usage: make changelog)
	@printf "$(BLUE)📝 Generating CHANGELOG.md...$(RESET)\n"
	@eval "$$(devbox shellenv)" && git-cliff -o CHANGELOG.md
	@printf "$(GREEN)✅ CHANGELOG.md generated successfully$(RESET)\n"

changelog-preview: ## 👁️  Preview unreleased changes
	@printf "$(BLUE)👁️  Previewing unreleased changes...$(RESET)\n"
	@eval "$$(devbox shellenv)" && git-cliff --unreleased

changelog-latest: ## 📋 Show latest release changes
	@printf "$(BLUE)📋 Showing latest release changes...$(RESET)\n"
	@eval "$$(devbox shellenv)" && git-cliff --latest

# =============================================================================
# Information Display
# =============================================================================

info: ## ℹ️  Display project information
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║                 RCABench Project Information                 ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "$(YELLOW)Configuration Information:$(RESET)\n"
	@printf "  $(CYAN)Namespace:$(RESET) $(NS)\n"
	@printf "  $(CYAN)Port:$(RESET) $(PORT)\n"
	@printf "  $(CYAN)Controller Directory:$(RESET) $(SRC_DIR)\n"
	@printf "  $(CYAN)Python SDK Directory:$(RESET) sdk/python\n"
	@printf "\n"
	@printf "$(YELLOW)Chaos Types:$(RESET)\n"
	@for type in $(CHAOS_TYPES); do \
		printf "  - $$type\n"; \
	done