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
DEFAULT_REPO 	:= 10.10.10.240/library
NS          	:= exp
PORT        	:= 30080
RELEASE_NAME    := rcabench

# Directory Configuration
HUSKY_DIR := .husky
SRC_DIR := src
COMMAND_DIR := scripts/command
COMMAND_BIN := command.bin

SDK_VERSION ?=0.0.0
GENERATOR_IMAGE ?= docker.io/opspai/openapi-generator-cli:1.0.0

PYTHON_SDK_DIR := sdk/python
PYTHON_SDK_GEN_DIR := sdk/python-gen
PYTHON_SDK_CONFIG := .openapi-generator/python/config.json
PYTHON_SDK_TEMPLATES := .openapi-generator/python/templates

# Chaos Types Configuration
CHAOS_TYPES := dnschaos httpchaos jvmchaos networkchaos podchaos stresschaos timechaos

# Color definitions
BLUE    := \033[1;34m
GREEN   := \033[1;32m
YELLOW  := \033[1;33m
RED     := \033[1;31m
CYAN    := \033[1;36m
GRAY    := \033[90m
RESET   := \033[0m

BACKUP_DATA ?= $(shell [ -t 0 ] && echo "ask" || echo "no")
START_APP   ?= $(shell [ -t 0 ] && echo "ask" || echo "yes")

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
	@printf "$(GREEN)🎉 All dependency checks passed!$(RESET)\n"

setup-dev-env: check-prerequisites ## 🛠️  Setup development environment
	@printf "$(BLUE)🛠️  Setting up development environment...$(RESET)\n"
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

build-make-command:
	@if [ -f "$(COMMAND_DIR)/$(COMMAND_BIN)" ]; then \
		printf "$(GREEN)✅ Make command tool binary found. Skipping installation.$(RESET)\n"; \
	else \
		printf "$(BLUE)📦 Command tool binary not found. Building now...$(RESET)\n"; \
		sudo apt install patchelf ccache; \
		cd $(COMMAND_DIR) && \
		uv venv --clear && \
		. .venv/bin/activate && \
		uv sync --quiet --extra nuitka-build && \
		uv run python -m nuitka --standalone --onefile --lto=yes \
			--output-dir=. \
			--output-filename=command.bin \
			main.py; \
		printf "$(GREEN)✅ Make command tool installation completed.$(RESET)\n"; \
	fi

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

# =============================================================================
# Pedestal Function
# =============================================================================

# Function to extract pedestal information
get_pedestal_chart_ref = $(shell echo "$(PEDESTAL_MAPPING)" | grep "^$(1)=" | cut -d'=' -f2 | cut -d':' -f1)
get_pedestal_image_tag = $(shell echo "$(PEDESTAL_MAPPING)" | grep "^$(1)=" | cut -d'=' -f2 | cut -d':' -f2)
get_pedestal_node_port = $(shell echo "$(PEDESTAL_MAPPING)" | grep "^$(1)=" | cut -d'=' -f2 | cut -d':' -f3)

# Function to validate pedestal (usage: $(call validate-pedestal-key,key))
validate_pedestal_key = $(shell echo "$(PEDESTAL_MAPPING)" | grep -q "^$(1)=" && echo "valid" || echo "invalid")

install-release: ## 🚀 Install Pedestal Release (usage: make install-release PEDESTAL_KEY=ts)
	@if [ -z "$(PEDESTAL_KEY)" ]; then \
		printf "$(RED)❌ Please provide pedestal key: make install-release PEDESTAL_KEY=ts$(RESET)\n"; \
		exit 1; \
	fi
	@if [ "$(call validate_pedestal_key,$(PEDESTAL_KEY))" = "invalid" ]; then \
		printf "$(RED)❌ Invalid pedestal key '$(PEDESTAL_KEY)'$(RESET)\n"; \
		printf "$(YELLOW)Available keys:$(RESET)\n"; \
		$(MAKE) show-pedestal-map; \
		exit 1; \
	fi
	@if [ "$(PEDESTAL_KEY)" = "ts" ] && [ -n "$(NS_COUNT)" ] && [ -n "$(NODE_PORT)" ]; then \
		pedestal_key="$(PEDESTAL_KEY)"; \
		ns="$${pedestal_key}$(NS_COUNT)"; \
		chart_ref="$(call get_pedestal_chart_ref,$(PEDESTAL_KEY))"; \
		image_tag="$(call get_pedestal_image_tag,$(PEDESTAL_KEY))"; \
		kube_context="$(shell kubectl config current-context)"; \
		printf "$(GRAY)Using Kubernetes context: $$kube_context$(RESET)\n"; \
		printf "$(BLUE)🚀 Installing $$chart_ref release in namespace $$ns on port $(NODE_PORT)...$(RESET)\n"; \
		if [ "$$kube_context" = "$(DEV_CONTEXT)" ]; then \
			helm install "$$ns" "$$chart_ref" -n "$$ns" --create-namespace \
				--set global.image.tag="$$image_tag" \
				--set global.security.allowInsecureImages=true \
				--set services.tsUiDashboard.nodePort="$(NODE_PORT)"; \
		elif [ "$$kube_context" = "$(PROD_CONTEXT)" ]; then \
			helm install "$$ns" "$$chart_ref" -n "$$ns" --create-namespace \
				--set global.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/opspai \
				--set global.image.tag="$$image_tag" \
				--set global.security.allowInsecureImages=true \
				--set mysql.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/library/mysql \
  				--set rabbitmq.image.registry=pair-diagnose-cn-guangzhou.cr.volces.com \
  				--set rabbitmq.image.repository=bitnamilegacy/rabbitmq \
  				--set loadgenerator.image.repository=pair-diagnose-cn-guangzhou.cr.volces.com/opspai/loadgenerator \
  				--set loadgenerator.initContainer.image=pair-diagnose-cn-guangzhou.cr.volces.com/nicolaka/netshoot:v0.14 \
				--set services.tsUiDashboard.nodePort="$(NODE_PORT)"; \
		else \
			printf "$(RED)❌ Unknown Kubernetes context '$$kube_context'. Please switch to a valid context.$(RESET)\n"; \
			printf "$(YELLOW)Available contexts:$(RESET)\n"; \
			kubectl config get-contexts -o name; \
			exit 1; \
		fi; \
	else \
		printf "$(RED)❌ Please provide NS_COUNT and NODE_PORT for pedestal key '$(PEDESTAL_KEY)': make install-release PEDESTAL_KEY=ts NS_COUNT=0 NODE_PORT=31000$(RESET)\n"; \
		exit 1; \
	fi

install-releases: ## 🔍 Install Helm releases in namespaces (usage: make install-releases PEDESTAL_KEY=ts PEDESTAL_COUNT=2)
	@if [ -z "$(PEDESTAL_KEY)" ] || [ -z "$(PEDESTAL_COUNT)" ]; then \
		printf "$(RED)❌ Please provide pedestal key and count: make install-releases PEDESTAL_KEY=ts PEDESTAL_COUNT=2$(RESET)\n"; \
		exit 1; \
	fi
	@if ! printf "$(PEDESTAL_COUNT)" | grep -Eq '^[0-9]+$$'; then \
		printf "$(RED)❌ PEDESTAL_COUNT must be a positive number$(RESET)\n"; \
		exit 1; \
	fi
	@if [ "$(call validate_pedestal_key,$(PEDESTAL_KEY))" = "invalid" ]; then \
		printf "$(RED)❌ Invalid pedestal key '$(PEDESTAL_KEY)'$(RESET)\n"; \
		printf "$(YELLOW)Available keys:$(RESET)\n"; \
		$(MAKE) show-pedestal-map; \
		exit 1; \
	fi
	@$(call get-pedestal-info PEDESTAL_KEY="$(PEDESTAL_KEY)")
	@printf "\n$(BLUE)🔍 Checking Helm releases in namespaces $(PEDESTAL_KEY)0 to $(PEDESTAL_KEY)$$(( $(PEDESTAL_COUNT) - 1 ))...$(RESET)\n"
	@base_port="$(call get_pedestal_node_port,$(PEDESTAL_KEY))"; \
	for i in $$(seq 0 $$(( $(PEDESTAL_COUNT) - 1 ))); do \
		ns="$(PEDESTAL_KEY)$$i"; \
		port="$$(($$base_port + i))"; \
		printf "$(BLUE)🔍 Checking namespace: $$ns$(RESET)\n"; \
		if ! kubectl get namespace "$$ns" >/dev/null 2>&1; then \
			printf "$(YELLOW)❌ Namespace $$ns does not exist$(RESET)\n"; \
		elif helm list -n "$$ns" 2>/dev/null | grep -q "$$ns"; then \
			printf "$(GREEN)✅ Helm release '$$ns' found in namespace $$ns$(RESET)\n"; \
		else \
			printf "$(YELLOW)⚠️ Helm release '$$ns' not found in namespace $$ns$(RESET)\n"; \
			if [ "$(PEDESTAL_KEY)" = "ts" ]; then \
				$(MAKE) install-release PEDESTAL_KEY="ts" NS_COUNT="$$i" NODE_PORT="$$port" ; \
			fi; \
		fi; \
	done
	@printf "$(GREEN)🎉 Check and installation completed!$(RESET)\n"

define get-pedestal-info
	@printf "$(BLUE)ℹ️ Pedestal Information for '$(PEDESTAL_KEY)':$(RESET)\n"
	@printf "$(YELLOW)Full Chart Reference:$(RESET) $(call get_pedestal_chart_ref,$(PEDESTAL_KEY))\n"
	@printf "$(YELLOW)Image Tag:$(RESET) $(call get_pedestal_image_tag,$(PEDESTAL_KEY))\n"
	@printf "$(YELLOW)Node Port:$(RESET) $(call get_pedestal_node_port,$(PEDESTAL_KEY))\n"
endef

show-pedestal-info: ## ℹ️  Show pedestal repository information (usage: make show-pedestal-info PEDESTAL_KEY=ts)
	@if [ -z "$(PEDESTAL_KEY)" ]; then \
		printf "$(RED)❌ Please provide pedestal key: make show-pedestal-info PEDESTAL_KEY=ts$(RESET)\n"; \
		exit 1; \
	fi
	@if [ "$(call validate_pedestal_key,$(PEDESTAL_KEY))" = "invalid" ]; then \
		printf "$(RED)❌ Invalid pedestal key '$(PEDESTAL_KEY)'$(RESET)\n"; \
		printf "$(YELLOW)Available keys:$(RESET)\n"; \
		$(MAKE) show-pedestal-map; \
		exit 1; \
	fi
	@$(call get-pedestal-info)

show-pedestal-map: ## 🗺️  Show available pedestal mappings
	@printf "$(BLUE)🗺️ Available Pedestal Mappings:$(RESET)\n\n"
	@printf "$(YELLOW)Format: KEY -> REPO_NAME/CHART_NAME$(RESET)\n\n"
	@echo "$(PEDESTAL_MAPPING)" | tr ' ' '\n' | while IFS='=' read -r key mapping; do \
		if [ -n "$$key" ] && [ -n "$$mapping" ]; then \
			chart_ref=$$(echo "$$mapping" | cut -d':' -f1); \
			printf "$(CYAN)%-0s$(RESET) -> $(GREEN)%-20s$(RESET)\n" "$$key" "$$chart_ref"; \
		fi; \
	done
	

# =============================================================================
# Secret Management
# =============================================================================

install-secrets: ## 🔑 Install all Secrets from Helm templates
	@printf "$(BLUE)🔑 Installing Secrets in namespace $(NS)...$(RESET)\n"
	@helm template $(RELEASE_NAME) ./helm -n $(NS) -s templates/secret.yaml | kubectl apply -f -
	@printf "$(GREEN)✅ Secrets installed$(RESET)\n"

check-secrets: ## 🔍 Check required Secrets exist
	@printf "$(BLUE)🔍 Checking required Secrets in namespace $(NS)...$(RESET)\n"
	@printf "$(GRAY)Extracting Secret names from Helm templates...$(RESET)\n"
	@all_ok=true; \
	expected_secrets=$$(helm template $(RELEASE_NAME) ./helm -n $(NS) -s templates/secret.yaml 2>/dev/null | \
		awk '/^kind: Secret/,/^---/ { if (/^metadata:/) { getline; if (/name:/) print $$2 } }' | \
		sort -u); \
	if [ -z "$$expected_secrets" ]; then \
		printf "$(YELLOW)⚠️  No Secrets defined in Helm templates$(RESET)\n"; \
		exit 0; \
	fi; \
	printf "$(CYAN)Expected Secrets:$(RESET)\n"; \
	echo "$$expected_secrets" | while read secret; do \
		printf "  - $$secret\n"; \
	done; \
	printf "\n"; \
	echo "$$expected_secrets" | while read secret; do \
		if kubectl get secret $$secret -n $(NS) >/dev/null 2>&1; then \
			printf "$(GREEN)✅ $$secret exists$(RESET)\n"; \
		else \
			printf "$(RED)❌ $$secret not found$(RESET)\n"; \
			all_ok=false; \
		fi; \
	done; \
	if [ "$$all_ok" = "false" ]; then \
		printf "$(YELLOW)💡 Run: make install-secrets$(RESET)\n"; \
		exit 1; \ 
	fi

# =============================================================================
# HostPath Management
# =============================================================================

install-hostpath: ## 🚀 Install HostPath DaemonSet
	@printf "$(BLUE)🚀 Installing HostPath DaemonSet in namespace $(NS)...$(RESET)\n"
	@helm template $(RELEASE_NAME) ./helm -n $(NS) -s templates/daemonset.yaml | kubectl apply -f -
	@printf "$(GREEN)✅ HostPath DaemonSet installation initiated$(RESET)\n"

check-hostpath-daemonset: ## 🔍 Check if HostPath DaemonSet is installed and ready
	@printf "$(BLUE)🔍 Checking HostPath DaemonSet status...$(RESET)\n"
	@if kubectl get daemonset $(RELEASE_NAME)-hostpath-init -n $(NS) >/dev/null 2>&1; then \
		printf "$(GREEN)✅ HostPath DaemonSet exists$(RESET)\n"; \
		desired=$$(kubectl get daemonset $(RELEASE_NAME)-hostpath-init -n $(NS) -o jsonpath='{.status.desiredNumberScheduled}'); \
		ready=$$(kubectl get daemonset $(RELEASE_NAME)-hostpath-init -n $(NS) -o jsonpath='{.status.numberReady}'); \
		printf "$(CYAN)📊 Status: $$ready/$$desired pods ready$(RESET)\n"; \
		if [ "$$ready" = "$$desired" ] && [ "$$ready" != "0" ]; then \
			printf "$(GREEN)✅ All HostPath init pods are ready$(RESET)\n"; \
			$(MAKE) check-hostpath-logs; \
		else \
			printf "$(YELLOW)⚠️  HostPath init pods not fully ready$(RESET)\n"; \
			exit 1; \
		fi; \
	else \
		printf "$(RED)❌ HostPath DaemonSet not found$(RESET)\n"; \
		printf "$(YELLOW)💡 It will be installed during 'make run'$(RESET)\n"; \
		exit 1; \
	fi

check-hostpath-logs: ## 🔍 Check HostPath initialization from pod logs
	@printf "$(BLUE)🔍 Checking HostPath initialization logs...$(RESET)\n"
	@pods=$$(kubectl get pods -l app=$(RELEASE_NAME)-hostpath-init -n $(NS) -o jsonpath='{.items[*].metadata.name}'); \
	all_ok=true; \
	for pod in $$pods; do \
		node=$$(kubectl get pod $$pod -n $(NS) -o jsonpath='{.spec.nodeName}'); \
		printf "$(CYAN)📍 Checking pod $$pod on node $$node$(RESET)\n"; \
		if kubectl logs $$pod -n $(NS) 2>/dev/null | grep -q "HostPath directories initialized successfully"; then \
			printf "$(GREEN)  ✅ Directories initialized$(RESET)\n"; \
		else \
			printf "$(RED)  ❌ Initialization failed or incomplete$(RESET)\n"; \
			all_ok=false; \
		fi; \
	done; \
	if [ "$$all_ok" = "false" ]; then \
		printf "$(RED)❌ Some nodes have incomplete HostPath initialization$(RESET)\n"; \
		exit 1; \
	fi	

# =============================================================================
# Build and Deployment
# =============================================================================

run: check-prerequisites ## 🚀 Build and deploy application (using skaffold)
	@printf "$(BLUE)🔄 Starting deployment process...$(RESET)\n"
	@printf "$(BLUE)📋 Step 1: Backup existing data...$(RESET)\n"
	@if $(MAKE) check-db 2>/dev/null; then \
		printf "$(YELLOW)📄 Backing up existing database...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_mysql backup; \
	else \
		printf "$(YELLOW)⚠️ Database not running, skipping backup$(RESET)\n"; \
	fi
	@printf "$(BLUE)📋 Step 2: Deploy with Skaffold...$(RESET)\n"
	skaffold run --default-repo=$(DEFAULT_REPO)
	@printf "$(BLUE)📋 Step 3: Wait for deployment...$(RESET)\n"
	$(MAKE) wait-for-deployment
	@printf "$(BLUE)📋 Step 4: Wait for HostPath initialization...$(RESET)\n"
	$(MAKE) wait-for-hostpath-init
	@printf "$(BLUE)📋 Step 5: Verifying Secrets...$(RESET)\n"
	@$(MAKE) check-secrets
	@printf "$(GREEN)🎉 Deployment completed!$(RESET)\n"
	@printf "$(CYAN)📊 Deployment Summary:$(RESET)\n"
	@printf "$(GRAY)  - Namespace: $(NS)$(RESET)\n"
	@printf "$(GRAY)  - Access URL: http://$$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}'):$(PORT)$(RESET)\n"

wait-for-deployment: ## ⏳ Wait for deployment to be ready
	@printf "$(BLUE)⏳ Waiting for all deployments to be ready...$(RESET)\n"
	kubectl wait --for=condition=available --timeout=300s deployment --all -n $(NS)
	@printf "$(GREEN)✅ All deployments are ready$(RESET)\n"

wait-for-hostpath-init: ## ⏳ Wait for HostPath initialization DaemonSet to complete
	@printf "$(BLUE)⏳ Waiting for HostPath initialization DaemonSet...$(RESET)\n"
	@if kubectl get daemonset $(RELEASE_NAME)-hostpath-init -n $(NS) >/dev/null 2>&1; then \
		printf "$(GRAY)DaemonSet found, checking status...$(RESET)\n"; \
		kubectl rollout status daemonset/$(RELEASE_NAME)-hostpath-init -n $(NS) --timeout=120s || \
			(printf "$(RED)❌ DaemonSet failed to initialize$(RESET)\n" && exit 1); \
		printf "$(GREEN)✅ HostPath DaemonSet is ready$(RESET)\n"; \
		printf "$(BLUE)🔍 Verifying directories on nodes...$(RESET)\n"; \
		sleep 5; \
		pods=$$(kubectl get pods -l app=$(RELEASE_NAME)-hostpath-init -n $(NS) -o jsonpath='{.items[*].metadata.name}'); \
		for pod in $$pods; do \
			node=$$(kubectl get pod $$pod -n $(NS) -o jsonpath='{.spec.nodeName}'); \
			printf "$(CYAN)📍 Checking pod $$pod on node $$node$(RESET)\n"; \
			if kubectl logs $$pod -n $(NS) | grep -q "HostPath directories initialized successfully"; then \
				printf "$(GREEN)  ✅ Directories initialized$(RESET)\n"; \
			else \
				printf "$(YELLOW)  ⚠️  Initialization in progress or failed$(RESET)\n"; \
			fi; \
		done; \
		printf "$(GREEN)✅ HostPath initialization completed$(RESET)\n"; \
	else \
		printf "$(YELLOW)⚠️  HostPath DaemonSet not found, skipping check$(RESET)\n"; \
	fi

build: ## 🔨 Build application only (no deployment)
	@printf "$(BLUE)🔨 Building application...$(RESET)\n"
	skaffold build --default-repo=$(DEFAULT_REPO)
	@printf "$(GREEN)✅ Build completed$(RESET)\n"

# =============================================================================
# Database Management
# =============================================================================

check-db: ## 🔍 Check database status
	@printf "$(BLUE)🔍 Checking database status...$(RESET)\n"
	@if kubectl get pods -n $(NS) -l app=rcabench-mysql --field-selector=status.phase=Running | grep -q rcabench-mysql; then \
		printf "$(GREEN)✅ Database is running$(RESET)\n"; \
	else \
		printf "$(RED)❌ Database not running in namespace $(NS)$(RESET)\n"; \
		printf "$(GRAY)Available Pods:$(RESET)\n"; \
		kubectl get pods -n $(NS) -l app=rcabench-mysql || printf "$(GRAY)No database pods found$(RESET)\n"; \
		exit 1; \
	fi

check-redis: ## 🔍 check Redis status
	@printf "$(BLUE)🔍 Checking Redis status...$(RESET)\n"
	@if kubectl get pods -n $(NS) -l app=rcabench-redis --field-selector=status.phase=Running | grep -q rcabench-redis; then \
		printf "$(GREEN)✅ Redis is running$(RESET)\n"; \
	else \
		printf "$(RED)❌ Redis not running in namespace $(NS)$(RESET)\n"; \
		printf "$(GRAY)Available Pods:$(RESET)\n"; \
		kubectl get pods -n $(NS) -l app=rcabench-redis || printf "$(GRAY)No Redis pods found$(RESET)\n"; \
		exit 1; \
	fi

# =============================================================================
# Development Tools
# =============================================================================

local-debug:  ## 🐛 Start local debugging environment
	@printf "$(BLUE)⌛️ Starting local application...$(RESET)\n"; \
	cd $(SRC_DIR) && go run main.go both --port 8082

local-deploy: ## 🛠️  Setup local development environment with basic services
	@printf "$(BLUE)🚀 Starting basic services...$(RESET)\n"
	@if ! docker compose down; then \
		printf "$(RED)❌ Docker Compose stop failed$(RESET)\n"; \
		exit 1; \
	fi
	@if ! docker compose up redis mysql jaeger buildkitd -d; then \
		printf "$(RED)❌ Docker Compose start failed$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🧹 Cleaning up Kubernetes Jobs...$(RESET)\n"
	@kubectl delete jobs --all -n $(NS) || printf "$(YELLOW)⚠️  Job cleanup failed or no Jobs to clean$(RESET)\n"
	@set -e; \
	if [ "$(BACKUP_DATA)" = "ask" ]; then \
		read -p "Backup data (y/n)? " use_backup; \
	elif [ "$(BACKUP_DATA)" = "yes" ]; then \
		use_backup="y"; \
	else \
		use_backup="n"; \
	fi; \
	if [ "$$use_backup" = "y" ]; then \
		db_status="down"; \
		redis_status="down"; \
		if $(MAKE) check-db 2>/dev/null; then \
		    db_status="up"; \
		fi; \
		if $(MAKE) check-redis 2>/dev/null; then \
		    redis_status="up"; \
		fi; \
		printf "$(GRAY)Database status: $$db_status$(RESET)\n"; \
		printf "$(GRAY)Redis status: $$redis_status$(RESET)\n"; \
		if [ "$$db_status" = "up" ]; then \
			printf "$(BLUE)🗄️ Backing up database from production environment...$(RESET)\n"; \
			$(MAKE) -C scripts/hack/backup_mysql migrate; \
		else \
		    printf "$(YELLOW)⚠️ Database not available, skipping database backup$(RESET)\n"; \
		fi; \
		if [ "$$redis_status" = "up" ]; then \
			printf "$(BLUE)📦 Backing up Redis from production environment...$(RESET)\n"; \
			$(MAKE) -C scripts/hack/backup_redis restore-local; \
		else \
			printf "$(YELLOW)⚠️ Redis not available, skipping Redis backup$(RESET)\n"; \
		fi; \
        printf "$(GREEN)✅ Environment preparation completed!$(RESET)\n"; \
	fi; \
	if [ "$(START_APP)" = "ask" ]; then \
		read -p "Start local application now (y/n)? " start_app; \
	elif [ "$(START_APP)" = "yes" ]; then \
		start_app="y"; \
	else \
		start_app="n"; \
	fi; \
	if [ "$$start_app" = "n" ]; then \
		printf "$(YELLOW)⏸️  Local application not started, you can start it manually later:$(RESET)\n"; \
		printf "$(GRAY)cd $(SRC_DIR) && go run main.go both --port 8082$(RESET)\n"; \
	else \
		$(MAKE) local-debug; \
	fi

update-dependencies: ## 📦 Update latest version of dependencies
	@printf "$(BLUE)📦 Updating latest version of chaos-experiment library...$(RESET)\n"
	cd $(SRC_DIR) && go get -u $(CHAOS_EXPERIMENT_REPO) && go mod tidy
	@printf "$(GREEN)✅ Dependencies update completed$(RESET)\n"

# =============================================================================
# Chaos Management
# =============================================================================

# Function to get target namespaces matching prefix pattern with optional count limit
# Usage: $(call get_target_namespaces,prefix) or $(call get_target_namespaces,prefix,count)
define get_target_namespaces
    kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | tr ' ' '\n' | grep "^$(1)[0-9]\+$$" | sort | $(if $(2),head -n $(2),cat)
endef

clean-finalizers: ## 🧹 Clean all chaos resource finalizers in namespaces with specific prefix
	@if [ -z "$(NS_PREFIX)" ]; then \
		printf "$(RED)❌ Please provide namespace prefix: make clean-finalizers NS_PREFIX=ts NS_COUNT=2$(RESET)\n"; \
		exit 1; \
	fi
	@if [ -z "$(NS_COUNT)" ]; then \
		printf "$(RED)❌ Please provide namespace count for namespace prefix '$(NS_PREFIX)': make clean-finalizers NS_PREFIX=$(NS_PREFIX) NS_COUNT=2$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🧹 Cleaning chaos finalizers...$(RESET)\n"
	@printf "$(GRAY)Dynamically getting namespaces with prefix $(NS_PREFIX)...$(RESET)\n"
	@namespaces="$$($(call get_target_namespaces,$(NS_PREFIX),$(NS_COUNT)))"; \
	printf "$(CYAN)Found the following namespaces:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns\n"; \
	done; \
	printf "$(GRAY)Total: $$(printf "$$namespaces" | wc -w) namespaces$(RESET)\n"; \
	printf ""; \
	for ns in $$namespaces; do \
		printf "$(BLUE)🔄 Processing namespace: $$ns$(RESET)\n"; \
		for type in $(CHAOS_TYPES); do \
			printf "$(GRAY)Cleaning $$type...$(RESET)\n"; \
			kubectl get $$type -n $$ns -o jsonpath='{range .items[*]}{.metadata.namespace}{":"}{.metadata.name}{"\n"}{end}' | \
			while IFS=: read -r ns name; do \
				[ -n "$$name" ] && kubectl patch $$type "$$name" -n "$$ns" --type=merge -p '{"metadata":{"finalizers":[]}}'; \
			done; \
		done; \
	done
	@printf "$(GREEN)✅ Finalizer cleanup completed$(RESET)\n"

delete-all-chaos: ## 🗑️  Delete all chaos resources in namespaces with specific prefix
	@if [ -z "$(NS_PREFIX)" ]; then \
		printf "$(RED)❌ Please provide namespace prefix: make delete-all-chaos NS_PREFIX=ts NS_COUNT=2$(RESET)\n"; \
		exit 1; \
	fi
	@if [ -z "$(NS_COUNT)" ]; then \
		printf "$(RED)❌ Please provide namespace count for namespace prefix '$(NS_PREFIX)': make delete-all-chaos NS_PREFIX=$(NS_PREFIX) NS_COUNT=2$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🗑️ Deleting all chaos resources...$(RESET)\n"
	@printf "$(GRAY)Dynamically getting namespaces with prefix $(NS_PREFIX)...$(RESET)\n"
	@namespaces="$$($(call get_target_namespaces,$(NS_PREFIX),$(NS_COUNT)))"; \
	printf "$(CYAN)Found the following namespaces:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns\n"; \
	done; \
	printf "$(GRAY)Total: $$(printf "$$namespaces" | wc -w) namespaces$(RESET)\n"; \
	printf ""; \
	for ns in $$namespaces; do \
		printf "$(BLUE)🔄 Processing namespace: $$ns$(RESET)\n"; \
		for type in $(CHAOS_TYPES); do \
			printf "$(GRAY)Deleting $$type...$(RESET)\n"; \
			kubectl delete $$type --all -n $$ns; \
		done; \
	done
	@printf "$(GREEN)✅ Chaos resources deletion completed$(RESET)\n"

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

format-staged-python: ## 🎨 Lint and format staged python files with ruff
	source ./scripts/command/.venv/bin/activate && uv run ./scripts/command/main.py format python

# =============================================================================
# SDK Generation
# =============================================================================

swag-init: ## 📝 Initialize Swagger documentation
	@printf "$(BLUE)📝 Initializing Swagger documentation...$(RESET)\n"
	swag init -d ./$(SRC_DIR) --parseDependency --parseDepth 1 --output ./$(SRC_DIR)/docs/openapi2
	@printf ""
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		$(GENERATOR_IMAGE) generate \
		-i /local/$(SRC_DIR)/docs/openapi2/swagger.json \
		-g openapi \
		-o /local/$(SRC_DIR)/docs/openapi3 
	@printf "$(BLUE)📦 Post-processing swagger initiaization...$(RESET)\n"
	python ./scripts/swag-init-postprocess.py
	@printf "$(GREEN)✅ Swagger documentation generation completed$(RESET)\n"

generate-python-sdk: swag-init ## ⚙️ Generate Python SDK from Swagger documentation
	@printf "$(BLUE)🐍 Generating Python SDK...$(RESET)\n"
	@printf "$(BLUE)Updating "$(PYTHON_SDK_CONFIG)" packageVersion to $(SDK_VERSION)...$(RESET)\n"
	@jq --arg ver "$(SDK_VERSION)" '.packageVersion = $$ver' \
        $(PYTHON_SDK_CONFIG) > temp.json && \
        mv temp.json $(PYTHON_SDK_CONFIG)
	@mkdir -p $(PYTHON_SDK_GEN_DIR); \
    find $(PYTHON_SDK_GEN_DIR) -mindepth 1 -delete; \
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		$(GENERATOR_IMAGE) generate \
		-i /local/$(SRC_DIR)/docs/converted/sdk.json \
		-g python \
		-o /local/$(PYTHON_SDK_GEN_DIR) \
		-c /local/$(PYTHON_SDK_CONFIG) \
		-t /local/$(PYTHON_SDK_TEMPLATES) \
		--git-host github.com \
		--git-repo-id AegisLab \
		--git-user-id OperationsPAI
	@printf "$(BLUE)📦 Post-processing generated SDK...$(RESET)\n"
	@$(MAKE) build-make-command
	./scripts/mv-generated-python-sdk.sh
	@printf "$(BLUE) 🐍 Formatting generated Python SDK using Venv in $(COMMAND_DIR)...$(RESET)\n"
	./$(COMMAND_DIR)/command.bin format python
	@printf "$(GREEN)✅ Python SDK generation completed$(RESET)\n"


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
	@printf "  $(CYAN)Default Repository:$(RESET) $(DEFAULT_REPO)\n"
	@printf "  $(CYAN)Namespace:$(RESET) $(NS)\n"
	@printf "  $(CYAN)Port:$(RESET) $(PORT)\n"
	@printf "  $(CYAN)Controller Directory:$(RESET) $(SRC_DIR)\n"
	@printf "  $(CYAN)Python SDK Directory:$(RESET) $(PYTHON_SDK_DIR)\n"
	@printf "\n"
	@printf "$(YELLOW)Chaos Types:$(RESET)\n"
	@for type in $(CHAOS_TYPES); do \
		printf "  - $$type\n"; \
	done