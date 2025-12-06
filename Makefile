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
COMMAND_BIN := $(COMMAND_DIR)/command.bin

NS_PREFIX ?= ts
NS_COUNT  ?= 2

BUILD_COMMAND_SCRIPT := ./scripts/build-command.sh

ifeq ($(ENV_MODE),prod)
	COMMAND := . $(COMMAND_BIN)
else
	DEFAULT_REPO := 10.10.10.240/library
	COMMAND := . $(COMMAND_DIR)/.venv/bin/activate && uv run python $(COMMAND_DIR)/main.py
endif

# Directory Configuration
HUSKY_DIR := .husky
SRC_DIR := src

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
CHAOS_EXPERIMENT_REPO := github.com/LGU-SE-Internal/chaos-experiment@injectionv2

# =============================================================================
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

build-make-command: ## 🔨 Build command tool
	@chmod +x $(BUILD_COMMAND_SCRIPT)
	@ENV_MODE=$(ENV_MODE) $(BUILD_COMMAND_SCRIPT)

run-command: build-make-command ## 🔧 Run command tool (usage: make run-command ARGS="your args")
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
	@command -v helm >/dev/null 2>&1 || { printf "$(RED)❌ helm not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ helm installed$(RESET)\n"
	@printf "$(GRAY)Checking kubectx...$(RESET)\n"
	@command -v kubectx >/dev/null 2>&1 || { printf "$(RED)❌ kubectx not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ kubectx installed$(RESET)\n"
	@printf "$(GREEN)🎉 All dependency checks passed!$(RESET)\n\n"

install-chaos-mesh: ## 📦 Install Chaos Mesh
	@printf "$(BLUE)📦 Installing Chaos Mesh...$(RESET)\n"
	helm repo add chaos-mesh https://charts.chaos-mesh.org
	helm install chaos-mesh chaos-mesh/chaos-mesh \
		--namespace=chaos-mesh \
		--create-namespace \
		--set chaosDaemon.runtime=containerd \
		--set chaosDaemon.socketPath=/run/k3s/containerd/containerd.sock \
		--version 2.7.2
	@printf "$(GREEN)✅ Chaos Mesh installation completed$(RESET)\n"

install-jfs-driver: ## 🚀 Install JuiceFS CSI Driver
	@printf "$(BLUE)🚀 Installing JuiceFS CSI Driver...$(RESET)\n"
	helm repo add juicefs https://juicedata.github.io/charts/
	helm repo update

	helm install juicefs-csi-driver juicefs/juicefs-csi-driver \
	--namespace kube-system \
	--set storageClasses[0].enabled=false
	@printf "$(GREEN)✅ JuiceFS CSI Driver installation completed$(RESET)\n"

install-secrets: ## 🔑 Install all Secrets from Helm templates
	@printf "$(BLUE)🔑 Installing Secrets in namespace $(NS)...$(RESET)\n"
	@helm template $(RELEASE_NAME) ./helm -n $(NS) -s templates/secret.yaml | kubectl apply -f -
	@printf "$(GREEN)✅ Secrets installed$(RESET)\n"

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
	@printf "$(GRAY)Installing Git hooks...$(RESET)\n"
	@printf "$(GRAY)Checking Husky Installation Status...$(RESET)\n"
	@if test -d $(HUSKY_DIR); then \
		printf "$(YELLOW)Warning: The $(HUSKY_DIR) directory already exists$(RESET)\n"; \
		printf "$(YELLOW)If you need to re-install, please remove the directory first$(RESET)\n"; \
	else \
		printf "$(BLUE)Directory $(HUSKY_DIR) not found. Running initialization...$(RESET)\n"; \
		devbox run install-hooks; \
		printf "$(GREEN)✅ Development environment setup completed!$(RESET)\n"; \
	fi

# =============================================================================
# Pedestal Function
# =============================================================================

install-pedestals: ## 🔍 Install pedestals in namespaces (usage: make install-releases PEDESTAL_NAME=ts PEDESTAL_COUNT=2)
	$(MAKE) run-command ARGS="pedestal install -e $(ENV_MODE) -n $(PEDESTAL_NAME) -c $(PEDESTAL_COUNT) -f"

# =============================================================================
# Build and Deployment
# =============================================================================

run: check-prerequisites ## 🚀 Build and deploy application (using skaffold)
	$(MAKE) run-command ARGS="rcabench run -e $(ENV_MODE)"

check-secrets: ## 🔍 Check required Secrets exist
	$(MAKE) run-command ARGS="rcabench check-secrets -e $(ENV_MODE)"

# =============================================================================
# Development Tools
# =============================================================================

local-debug:  ## 🐛 Start local debugging environment
	@printf "$(BLUE)⌛️ Starting local application...$(RESET)\n"; \
	cd $(SRC_DIR) && go run main.go both --port 8082

local-deploy: ## 🛠️  Setup local development environment with basic services
	@$(MAKE) run-command ARGS="rcabench local-deploy -e $(ENV_MODE)"

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
	@printf "$(BLUE)Running pre-commit checks...$(RESET)\n"
	@devbox run format-staged-go
	@if [ $$? -ne 0]; then \
		echo "❌ Go formatting failed. Please fix the issues before committing."; \
		exit 1; \
	fi
	@devbox run format-staged-python
	@if [ $$? -ne 0]; then \
		echo "❌ Python formatting failed. Please fix the issues before committing."; \
		exit 1; \
	fi
	@printf "$(GREEN)✅ Pre-commit checks passed!$(RESET)\n"

pre-push: ## 🚀 Run pre-push checks (validates tags and runs tests)
	@printf "$(BLUE)🚀 Running pre-push checks...$(RESET)\n"
	@printf "$(GRAY)Checking if pushing tags...$(RESET)\n"
	@while read local_ref local_sha remote_ref remote_sha; do \
		if echo "$$remote_ref" | grep -q "refs/tags/"; then \
			tag_name=$$(echo "$$remote_ref" | sed 's/refs\/tags\///'); \
			printf "$(CYAN)📌 Detected tag push: $$tag_name$(RESET)\n"; \
			if echo "$$tag_name" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
				printf "$(GREEN)✅ Valid semver tag: $$tag_name$(RESET)\n"; \
			else \
				printf "$(RED)❌ Invalid tag format: $$tag_name (expected: vX.Y.Z)$(RESET)\n"; \
				exit 1; \
			fi; \
		fi; \
	done
	@printf "$(GREEN)✅ Pre-push checks passed!$(RESET)\n"

format-staged-go: ## 🔍 Lint and format staged Go files with golangci-lint
	@printf "$(BLUE)🔍 Checking Uncommitted Go Issues...$(RESET)\n"
	@if [ -z "$$(git status --porcelain | grep '\.go$$')" ]; then \
		printf "$(YELLOW)No uncommitted Go file changes found to lint$(RESET)\n"; \
		exit 0; \
	fi
	@printf "$(CYAN)⚙️  Linting and formating new issues found in uncommitted changes...$(RESET)\n"
	@cd src && golangci-lint run \
		--issues-exit-code=1 \
		--path-prefix=src \
		--whole-files \
		--new-from-rev=HEAD~1

# =============================================================================
# SDK Generation
# =============================================================================

swag-init: ## 📝 Initialize Swagger documentation
	$(MAKE) run-command ARGS="swagger init -v $(SDK_VERSION)"

generate-python-sdk: swag-init ## ⚙️ Generate Python SDK from Swagger documentation
	$(MAKE) run-command ARGS="swagger generate-sdk -l python -v $(SDK_VERSION)"


# =============================================================================
# Cleanup and Maintenance
# =============================================================================

clean-all: ## 🧹 Clean all resources
	@printf "$(BLUE)🧹 Cleaning all resources...$(RESET)\n"
	@printf "$(YELLOW)⚠️ This will delete all deployed resources!$(RESET)\n"
	@read -p "Confirm to continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@printf "$(GRAY)Deleting Helm release...$(RESET)\n"
	helm uninstall rcabench -n $(NS) || true
	@printf "$(GRAY)Deleting namespace...$(RESET)\n"
	kubectl delete namespace $(NS) || true
	@printf "$(GRAY)Stopping port forwarding...$(RESET)\n"
	pkill -f "kubectl port-forward" || true
	@printf "$(GREEN)✅ Cleanup completed$(RESET)\n"

# =============================================================================
# Utilities
# =============================================================================

restart: ## 🔄 Restart application
	@printf "$(BLUE)🔄 Restarting application...$(RESET)\n"
	kubectl rollout restart deployment --all -n $(NS)
	@printf "$(GREEN)✅ Application restart completed$(RESET)\n"

scale: ## 📏 Scale deployment (usage: make scale DEPLOYMENT=app REPLICAS=3)
	@if [ -z "$(DEPLOYMENT)" ] || [ -z "$(REPLICAS)" ]; then \
		printf "$(RED)❌ Please provide deployment name and replica count: make scale DEPLOYMENT=app REPLICAS=3$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)📏 Scaling deployment $(DEPLOYMENT) to $(REPLICAS) replicas...$(RESET)\n"
	kubectl scale deployment $(DEPLOYMENT) --replicas=$(REPLICAS) -n $(NS)
	@printf "$(GREEN)✅ Extension completed$(RESET)\n"

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