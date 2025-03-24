# Global Variables
DEFAULT_REPO ?= 10.10.10.240/library
NS          ?= experiment

.PHONY: build run debug swagger build-sdk-python gen-dataset-dev gen-dataset-prod import jobs pods ports help

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

swagger: ## Generate Swagger API documentation
	swag init \
		-d ./experiments_controller \
		--parseDependency \
		--parseDepth 1

##@ SDK

build-sdk-python: ## Install Python SDK with hot reload
	cd sdk/python && uv pip install . --no-cache --force-reinstall

##@ Data Management

gen-dataset-dev: ## Generate test datasets (development)
	screen -dmS gen-dataset-dev bash -c 'python scripts/gen/dataset/main.py; exec bash'

gen-dataset-prod: ## Generate production datasets (persistent)
	screen -dmS gen-dataset-prod bash -c 'python scripts/gen/dataset/main.py; exec bash'

import: ## Import generated data to the system
	python scripts/cmd/main.py --algo -d1

##@ Kubernetes

jobs: ## List all experiment jobs
	kubectl get jobs -n $(NS)

pods: ## List all experiment pods
	kubectl get pods -n $(NS)

ports: ## Port-forward experiment service
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &

##@ Help

help:  ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)