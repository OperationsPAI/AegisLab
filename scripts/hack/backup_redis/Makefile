# 声明所有非文件目标
.PHONY: help restore-loacl

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

##@ Restore
restore-local: ## Restore local Redis from remote
	uv run python cli.py restore-batch --file_path keys.txt