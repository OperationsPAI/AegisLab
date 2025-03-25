GEN_DATASET_IMAGE := 10.10.10.240/library/gen_dataset:latest

.PHONY: build help ts-dev ts-prod

all: help

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

##@ Build
build: ## Build docker image
	docker build -t $(GEN_DATASET_IMAGE) -f Dockerfile . && \
	docker push $(GEN_DATASET_IMAGE)

##@ Dataset Generation
ts-dev: ## Generate train-ticket test datasets (development)
	docker compose down && docker compose up gen-dataset-ts-dev -d

ts-prod: ## Generate train-ticket production datasets (persistent)
	docker compose down && docker compose up gen-dataset-ts-prod -d
	