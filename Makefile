# Global Variables
DEFAULT_REPO ?= 10.10.10.240/library
NS          ?= experiment

.PHONY: build run debug swagger build-sdk-python gen-dataset-dev gen-dataset-prod import jobs pods ports help

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

build: ## Build and deploy using skaffold
	skaffold run --default-repo=$(DEFAULT_REPO)

run: ## Run application in debug mode with skaffold
	skaffold debug --default-repo=$(DEFAULT_REPO)

##@ Development

debug: ## Start local debug environment (databases + controller)
	docker compose down && \
	docker compose up redis mariadb -d && \
	kubectl delete jobs --all -n $(NS) && \
	cd experiments_controller && go run main.go both --port 8082

import: ## import the latest version of github.com/CUHK-SE-Group/chaos-experiment
	cd experiments_controller && \
	go get github.com/CUHK-SE-Group/chaos-experiment@injectionv2 && \
	go mod tidy

swagger: ## Generate Swagger API documentation
	swag init \
		-d ./experiments_controller \
		--parseDependency \
		--parseDepth 1

##@ Kubernetes

jobs: ## List all jobs
	kubectl get jobs -n $(NS)

pods: ## List all pods
	kubectl get pods -n $(NS)

ports: ## Port-forward service
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &