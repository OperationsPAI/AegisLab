# =============================================================================
# RCABench Makefile
# ========================================r	@printf	@printf "$(BLUE)🗑️ Resetting database in namespace $(NS)...$(RESET)\n""$(RED)⚠️ Warning: This will delete all database data!$(RESET)\n"set-db: ## 🗑️ Reset database (⚠️ Will delete all data)====================================
# This Makefile provides all build, deployment, and development tools for the RCABench project
# Use 'make help' to view all available commands

# =============================================================================
# Configuration Variables
# =============================================================================

# Basic Configuration
DEFAULT_REPO ?= docker.io/opspai
NS          ?= exp
NS_PREFIX	?= ts
PORT        ?= 30080

# Directory Configuration
SRC_DIR = src
SDK_DIR = sdk/python-gen

# Chaos Types Configuration
CHAOS_TYPES ?= dnschaos httpchaos jvmchaos networkchaos podchaos stresschaos timechaos

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

# =============================================================================
# Declare all non-file targets
# =============================================================================
.PHONY: help build run debug swagger import clean-finalizers delete-all-chaos k8s-resources ports \
        install-hooks deploy-ts swag-init generate-sdk release \
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
	@printf "$(BLUE)║                    RCABench Project Management Tool         ║$(RESET)\n"
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
	@printf "$(GRAY)Checking kubectl...$(RESET)\n"
	@command -v kubectl >/dev/null 2>&1 || { printf "$(RED)❌ kubectl not installed$(RESET)"; exit 1; }
	@printf "$(GREEN)✅ kubectl installed$(RESET)\n"
	@printf "$(GRAY)Checking skaffold...$(RESET)\n"
	@command -v skaffold >/dev/null 2>&1 || { printf "$(RED)❌ skaffold not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ skaffold installed$(RESET)\n"
	@printf "$(GRAY)Checking docker...$(RESET)\n"
	@command -v docker >/dev/null 2>&1 || { printf "$(RED)❌ docker not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ docker installed$(RESET)\n"
	@printf "$(GRAY)Checking helm...$(RESET)\n"
	@command -v helm >/dev/null 2>&1 || { printf "$(RED)❌ helm not installed$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ helm installed$(RESET)\n"
	@printf "$(GREEN)🎉 All dependency checks passed!$(RESET)\n"

setup-dev-env: check-prerequisites ## 🛠️ Setup development environment
	@printf "$(BLUE)🛠️ Setting up development environment...$(RESET)\n"
	@printf "$(GRAY)Installing Git hooks...$(RESET)\n"
	@$(MAKE) install-hooks
	@printf "$(GREEN)✅ Development environment setup completed!$(RESET)\n"

# =============================================================================
# Build and Deployment
# =============================================================================

run: check-prerequisites ## 🚀 Build and deploy application (using skaffold)
	@printf "$(BLUE)🔄 Starting deployment process...$(RESET)\n"
	@if $(MAKE) check-db 2>/dev/null; then \
		printf "$(YELLOW)📄 Backing up existing database...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_mysql backup; \
	else \
		printf "$(YELLOW)⚠️ Database not running, skipping backup$(RESET)\n"; \
	fi
	@printf "$(GRAY)Deploying using skaffold...$(RESET)\n"
	skaffold run --default-repo=$(DEFAULT_REPO)
	@printf "$(BLUE)⏳ Waiting for deployment to stabilize...$(RESET)\n"
	$(MAKE) wait-for-deployment
	@printf "$(GREEN)🎉 Deployment completed!$(RESET)\n"

wait-for-deployment: ## ⏳ Wait for deployment to be ready
	@printf "$(BLUE)⏳ Waiting for all deployments to be ready...$(RESET)\n"
	kubectl wait --for=condition=available --timeout=300s deployment --all -n $(NS)
	@printf "$(GREEN)✅ All deployments are ready$(RESET)\n"

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

reset-db: ## 🗑️ Reset database (⚠️ Will delete all data)
	@printf "$(RED)⚠️ Warning: This will delete all database data!$(RESET)\n"
	@read -p "Confirm to continue? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@if $(MAKE) check-db 2>/dev/null; then \
		printf "$(YELLOW)📄 Backing up existing database...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_psql backup; \
	else \
		printf "$(YELLOW)⚠️ Database not running, skipping backup$(RESET)\n"; \
	fi
	@printf "$(BLUE)🗑️ Resetting database in namespace $(NS)...$(RESET)\n"
	helm uninstall rcabench -n $(NS) || true
	@printf "$(BLUE)⏳ Waiting for Pods to terminate...$(RESET)\n"
	@while kubectl get pods -n $(NS) -l app=rcabench-mysql 2>/dev/null | grep -q .; do \
		printf "$(GRAY)  Still waiting for Pods to terminate...$(RESET)\n"; \
		sleep 2; \
	done
	@printf "$(GREEN)✅ All Pods terminated$(RESET)\n"
	kubectl delete pvc rcabench-mysql-data -n $(NS) || true
	@printf "$(BLUE)⏳ Waiting for PVC deletion...$(RESET)\n"
	@while kubectl get pvc -n $(NS) | grep -q rcabench-mysql-data; do \
		printf "$(GRAY)  Still waiting for PVC deletion...$(RESET)\n"; \
		sleep 2; \
	done
	@printf "$(GREEN)✅ PVC deletion successful$(RESET)\n"
	@printf "$(GREEN)✅ Database reset completed. Redeploying...$(RESET)\n"
	$(MAKE) run
	@printf "$(GREEN)🚀 Application redeployed successfully.$(RESET)\n"
	$(MAKE) -C scripts/hack/backup_mysql migrate
	@printf "$(GREEN)📦 Restoring database from backup.$(RESET)\n"

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

local-debug: ## 🐛 Start local debugging environment
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
		printf "$(BLUE)⌛️ Starting local application...$(RESET)\n"; \
		cd $(SRC_DIR) && go run main.go both --port 8082; \
	fi

local-debug-auto: ## 🤖 Start local debugging environment (auto mode, no interaction)
	@$(MAKE) local-debug BACKUP_DATA=yes START_APP=yes

local-debug-minimal: ## 🚀 Start local debugging environment (minimal mode, no backup, no auto start)
	@$(MAKE) local-debug BACKUP_DATA=no START_APP=no

import: ## 📦 Import latest version of chaos-experiment library
	@printf "$(BLUE)📦 Importing latest version of chaos-experiment library...$(RESET)\n"
	cd $(SRC_DIR) && \
	go get -u github.com/LGU-SE-Internal/chaos-experiment@injectionv2 && \
	go mod tidy
	@printf "$(GREEN)✅ Dependency update completed$(RESET)\n"

# =============================================================================
# Chaos Management
# =============================================================================

define get_target_namespaces
    kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | tr ' ' '\n' | grep "^$(NS_PREFIX)[0-9]$$" | sort
endef

clean-finalizers: ## 🧹 Clean all chaos resource finalizers
	@printf "$(BLUE)🧹 Cleaning chaos finalizers...$(RESET)\n"
	@printf "$(GRAY)Dynamically getting namespaces with prefix $(NS_PREFIX)...$(RESET)\n"
	@namespaces=$$($(call get_target_namespaces)); \
	printf "$(CYAN)Found the following namespaces:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns"; \
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

delete-all-chaos: ## 🗑️ Delete all chaos resources
	@printf "$(BLUE)🗑️ Deleting chaos resources...$(RESET)\n"
	@printf "$(GRAY)Dynamically getting namespaces with prefix $(NS_PREFIX)...$(RESET)\n"
	@namespaces=$$($(call get_target_namespaces)); \
	printf "$(CYAN)Found the following namespaces:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns"; \
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
# Kubernetes Management
# =============================================================================

k8s-resources: ## 📊 Show all jobs and pods
	@printf "$(BLUE)📊 Resources in namespace $(NS):$(RESET)\n"
	@printf "$(YELLOW)Jobs:$(RESET)\n"
	@kubectl get jobs -n $(NS)
	@printf "$(YELLOW)Pods:$(RESET)\n"
	@kubectl get pods -n $(NS)

status: ## 📈 View application status
	@printf "$(BLUE)📈 Application status overview:$(RESET)\n"
	@printf "$(YELLOW)Namespace: $(NS)$(RESET)\n"
	@printf "$(GRAY)Deployments:$(RESET)\n"
	@kubectl get deployments -n $(NS)
	@printf "$(GRAY)Services:$(RESET)\n"
	@kubectl get services -n $(NS)
	@printf "$(GRAY)Pods status:$(RESET)\n"
	@kubectl get pods -n $(NS) -o wide

logs: ## 📋 View application logs
	@printf "$(BLUE)📋 Application logs:$(RESET)\n"
	@printf "$(YELLOW)Select the Pod to view logs:$(RESET)\n"
	@kubectl get pods -n $(NS) --no-headers -o custom-columns=":metadata.name" | head -10
	@printf "$(GRAY)Use 'kubectl logs <pod-name> -n $(NS)' to view logs of a specific Pod$(RESET)\n"

ports: ## 🔌 Port forward services
	@printf "$(BLUE)🔌 Starting port forwarding...$(RESET)\n"
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &
	@printf "$(GREEN)✅ Port forwarding started (8081:8081)$(RESET)\n"
	@printf "$(GRAY)Access URL: http://localhost:8081$(RESET)\n"

# =============================================================================
# Git Management
# =============================================================================

install-hooks: ## 🔧 Install pre-commit hooks
	@printf "$(BLUE)🔧 Installing Git hooks...$(RESET)\n"
	chmod +x scripts/hooks/pre-commit
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@printf "$(GREEN)✅ Git hooks installation completed$(RESET)\n"

# =============================================================================
# SDK Generation
# =============================================================================

swagger: swag-init generate-sdk ## 📚 Generate complete Swagger documentation and SDK

## Initialize Swagger documentation
swag-init:
	@printf "$(BLUE)📝 Initializing Swagger documentation...$(RESET)\n"
	swag init -d ./$(SRC_DIR) --parseDependency --parseDepth 1 --output ./$(SRC_DIR)/docs
	@printf "$(GREEN)✅ Swagger documentation generation completed$(RESET)\n"

## Generate Python SDK from Swagger documentation
generate-sdk: swag-init
	@printf "$(BLUE)🐍 Generating Python SDK...$(RESET)\n"
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		openapitools/openapi-generator-cli:latest generate \
		-i /local/$(SRC_DIR)/docs/swagger.json \
		-g python \
		-o /local/$(SDK_DIR) \
		-c /local/.openapi-generator/config.properties \
		--additional-properties=packageName=openapi,projectName=rcabench
	@printf "$(BLUE)📦 Post-processing generated SDK...$(RESET)\n"
	./scripts/mv-generated-sdk.sh
	@printf "$(GREEN)✅ Python SDK generation completed$(RESET)\n"

# =============================================================================
# Release Management
# =============================================================================

release: ## 🏷️ Release new version (usage: make release VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		printf "$(RED)❌ Please provide version number: make release VERSION=1.0.1$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🏷️ Releasing version $(VERSION)...$(RESET)\n"
	./scripts/release.sh $(VERSION)

release-dry-run: ## 🧪 Release process dry run (usage: make release-dry-run VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		printf "$(RED)❌ Please provide version number: make release-dry-run VERSION=1.0.1$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🧪 Dry run release process $(VERSION)...$(RESET)\n"
	./scripts/release.sh $(VERSION) --dry-run

upload: ## 📤 Upload SDK package
	@printf "$(BLUE)📤 Uploading SDK package...$(RESET)\n"
	$(MAKE) -C sdk/python upload
	@printf "$(GREEN)✅ SDK upload completed$(RESET)\n"

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
	scale: ## 📏 Scale deployment (usage: make scale DEPLOYMENT=app REPLICAS=3)
	@printf "$(BLUE)📏 Scaling deployment $(DEPLOYMENT) to $(REPLICAS) replicas...$(RESET)\n"
	kubectl scale deployment $(DEPLOYMENT) --replicas=$(REPLICAS) -n $(NS)
	@printf "$(GREEN)✅ Extension completed$(RESET)\n"

# =============================================================================
# Information Display
# =============================================================================

info: ## ℹ️ Display project information
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║                        RCABench Project Information                     ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "$(YELLOW)Configuration Information:$(RESET)\n"
	@printf "  $(CYAN)Default Repository:$(RESET) $(DEFAULT_REPO)\n"
	@printf "  $(CYAN)Namespace:$(RESET) $(NS)\n"
	@printf "  $(CYAN)Port:$(RESET) $(PORT)\n"
	@printf "  $(CYAN)Controller Directory:$(RESET) $(SRC_DIR)\n"
	@printf "  $(CYAN)SDK Directory:$(RESET) $(SDK_DIR)\n"
	@printf "\n"
	@printf "$(YELLOW)Chaos Types:$(RESET)\n"
	@for type in $(CHAOS_TYPES); do \
		printf "  - $$type\n"; \
	done