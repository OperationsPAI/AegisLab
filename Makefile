# Global Variables
DEFAULT_REPO ?= 10.10.10.240/library
NS          ?= exp
CHAOS_TYPES ?= dnschaos httpchaos jvmchaos networkchaos podchaos stresschaos timechaos
TS_NS       ?= ts
PORT        ?= 30080
CONTROLLER_DIR = experiments_controller
SDK_DIR = sdk/python-gen

# 声明所有非文件目标
.PHONY: help build run debug swagger import clean-finalizer delete-chaos k8s-resources ports \
        install-hooks git-sync upgrade-dep deploy-ts swag-init generate-sdk release

# 默认目标
.DEFAULT_GOAL := help

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
		printf "  \033[36m%-20s\033[0m \033[90m%s\033[0m\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)

##@ Building

run: ## Build and deploy using skaffold
	skaffold run --default-repo=$(DEFAULT_REPO)

remote-debug: ## Run application in debug mode with skaffold
	skaffold debug --default-repo=$(DEFAULT_REPO)

##@ Development

local-debug: ## Start local debug environment (databases + controller)
	docker compose down && \
	docker compose up redis postgres jaeger buildkitd -d && \
	kubectl delete jobs --all -n $(NS) && \
	cd $(CONTROLLER_DIR) && go run main.go both --port 8082

import: ## Import the latest version of chaos-experiment library
	cd $(CONTROLLER_DIR) && \
	go get github.com/LGU-SE-Internal/chaos-experiment@injectionv2 && \
	go mod tidy


##@ Chaos Management
clean-finalizer: ## Clean finalizer for specified chaos types in namespace $(NS)
	@for type in $(CHAOS_TYPES); do \
		kubectl get $$type -n $(NS) -o jsonpath='{range .items[*]}{.metadata.namespace}{":"}{.metadata.name}{"\n"}{end}' | \
		while IFS=: read -r ns name; do \
			[ -n "$$name" ] && kubectl patch $$type "$$name" -n "$$ns" --type=merge -p '{"metadata":{"finalizers":[]}}'; \
		done; \
	done

delete-chaos: ## Delete specified chaos types in namespace $(NS)
	@for type in $(CHAOS_TYPES); do \
		kubectl delete $$type --all -n $(NS); \
	done

##@ Kubernetes

k8s-resources: ## Display all jobs and pods
	@echo "Jobs in namespace $(NS):"
	@kubectl get jobs -n $(NS)
	@echo "\nPods in namespace $(NS):"
	@kubectl get pods -n $(NS)

ports: ## Port-forward service
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &

##@ Git Management

install-hooks: ## Install pre-commit hooks
	chmod +x scripts/hooks/pre-commit
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

git-sync: ## Synchronize Git submodules
	git submodule update --init --recursive --remote

upgrade-dep: git-sync ## Upgrade Git submodules to latest main branch
	@git submodule foreach 'branch=$$(git config -f $$toplevel/.gitmodules submodule.$$name.branch || echo main); \
		echo "Updating $$name to branch: $$branch"; \
		git checkout $$branch && git pull origin $$branch'

##@ SDK Generation
swagger: swag-init generate-sdk 

swag-init: ## Initialize Swagger documentation
	swag init -d ./$(CONTROLLER_DIR) --parseDependency --parseDepth 1 --output ./$(CONTROLLER_DIR)/docs

generate-sdk: swag-init ## Generate Python SDK from Swagger documentation
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		openapitools/openapi-generator-cli:latest generate \
		-i /local/$(CONTROLLER_DIR)/docs/swagger.json \
		-g python \
		-o /local/$(SDK_DIR) \
		-c /local/.openapi-generator/config.properties \
		--additional-properties=packageName=openapi,projectName=rcabench
	@echo "📦 Post-processing generated SDK..."
	./scripts/fix-generated-sdk.sh
	./scripts/mv-generated-sdk.sh

##@ Release Management
release: ## Release a new version (usage: make release VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide a version number: make release VERSION=1.0.1"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION)

release-dry-run: ## Dry run release process (usage: make release-dry-run VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Please provide a version number: make release-dry-run VERSION=1.0.1"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION) --dry-run
