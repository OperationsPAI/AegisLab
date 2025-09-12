# =============================================================================
# Backup-redis Makefile
# =============================================================================
# This Makefile provides shortcuts for Backup-redis tool commands
# Use 'make help' to see all available commands

# =============================================================================
# Configuration Variables
# =============================================================================

# Basic Configuration
HASH_PATTERN := injection:algorithms
STREAM_PATTERN := trace:*:log

# Color Definitions
BLUE    := \033[1;34m
GREEN   := \033[1;32m
YELLOW  := \033[1;33m
RED     := \033[1;31m
CYAN    := \033[1;36m
GRAY    := \033[90m
RESET   := \033[0m

# =============================================================================
# Declare all non-file targets
# =============================================================================
.PHONY: help restore-loacl

# =============================================================================
# Default Target
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# Help Information
# =============================================================================
help: ## 📖 Show all available commands
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "$(YELLOW)Available Commands:$(RESET)\n"; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  $(CYAN)%-25s$(RESET) $(GRAY)%s$(RESET)\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)

# =============================================================================
# Data Restoration
# =============================================================================
restore-local: ## 📦 Restore data from remote to local Redis
	@printf "$(BLUE)Starting data restoration from remote to local Redis...$(RESET)\n"
	@printf "$(BLUE)Restoring hash data$(RESET)\n"
	uv run python cli.py restore-hashes --pattern "$(HASH_PATTERN)" --overwrite
	@printf "$(GREEN)✅ Hash data restoration completed$(RESET)\n"
	@printf "$(BLUE)Restoring stream data$(RESET)\n"
	@read -p "Choose matching method - Exact match (y) or Fuzzy match (N): " choice; \
	if [ "$$choice" = "y" ] || [ "$$choice" = "Y" ]; then \
		printf "$(BLUE)Using exact match mode...$(RESET)\n"; \
		uv run python cli.py restore-streams --exact_match --overwrite; \
	else \
		printf "$(BLUE)Using fuzzy match mode: $(STREAM_PATTERN)$(RESET)\n"; \
		uv run python cli.py restore-streams --fuzzy_match "$(STREAM_PATTERN)" --overwrite; \
	fi
	@printf "$(GREEN)✅ Stream data restoration completed$(RESET)\n"