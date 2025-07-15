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

##@ Backup
backup:
	uv run python cli.py pg-backup

##@ Restore
restore-local: ## Restore local PostgreSQL database from backup
	$(MAKE) backup
	uv run python cli.py pg-restore --clean --host 127.0.0.1 --port 5432

restore-remote: ## Restore remote PostgreSQL database from backup
	uv run python cli.py pg-restore --clean