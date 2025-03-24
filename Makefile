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

build-sdk-python-docker:
	docker build -t 10.10.10.240/library/sdk_python:latest -f ./sdk/python/Dockerfile ./sdk/python && \
	docker push 10.10.10.240/library/sdk_python:latest

##@ Data Management

GEN_DATASET_DIR := ./scripts/gen/dataset
GEN_DATASET_IMAGE := 10.10.10.240/library/gen_dataset:latest

build-gen-dataset-docker:
	docker build -t $(GEN_DATASET_IMAGE) -f "$(GEN_DATASET_DIR)/Dockerfile" $(GEN_DATASET_DIR) && \
	docker push $(GEN_DATASET_IMAGE)

gen-dataset-dev: ## Generate test datasets (development)
	cd $(GEN_DATASET_DIR) && docker compose down && docker compose up gen-dataset-ts-dev -d

gen-dataset-prod: ## Generate production datasets (persistent)
	cd $(GEN_DATASET_DIR) && docker compose down && docker compose up gen-dataset-ts-prod -d

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