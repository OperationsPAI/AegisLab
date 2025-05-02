SDK_PHTHON_IMAGE := 10.10.10.240/library/sdk_python:latest

.PHONY: help build build-docker

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
		printf "  \033[36m%-12s\033[0m \033[90m%s\033[0m\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)

##@ Building
build: ## Install Python SDK with hot reload
	uv pip install . --no-cache --force-reinstall

build-docker: ## Build SDK docker image 
	docker build -t $(SDK_PHTHON_IMAGE) -f Dockerfile . && \
	docker push $(SDK_PHTHON_IMAGE)

##@ Upload to pip
upload:
	rm -rf dist
	@if [ ! -f token ]; then echo "Error: token file not found"; exit 1; fi
	uv build && uv publish --token $(shell cat token)