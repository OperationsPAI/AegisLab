# =============================================================================
# MySQL Backup Tool Makefile
# =============================================================================
# This Makefile provides shortcuts for MySQL backup and restore operations
# Use 'make help' to see all available commands

# =============================================================================
# Configuration Variables
# =============================================================================

# Database Configuration
LOCAL_HOST := 127.0.0.1
LOCAL_PORT := 3306
LOCAL_USER := root
LOCAL_PASSWORD := yourpassword
LOCAL_DB := rcabench

REMOTE_HOST := 10.10.10.220
REMOTE_PORT := 32206
REMOTE_USER := root
REMOTE_PASSWORD := yourpassword
REMOTE_DB := rcabench

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
# Default target
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# Help information
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
# Information
# =============================================================================

status: ## 📊 Show backup status and disk usage
	@printf "$(BLUE)Backup Directory Status:$(RESET)\n"
	@ls -la ./temp/backup_mysql/ 2>/dev/null || printf "$(YELLOW)No backup directory found$(RESET)\n"
	@printf ""
	@printf "$(BLUE)Disk Usage:$(RESET)\n"
	@du -sh ./temp/backup_mysql/ 2>/dev/null || printf "$(YELLOW)No backup files found$(RESET)\n"

config: ## ⚙️ Show current configuration
	@printf "$(BLUE)Current Local MySQL Configuration:$(RESET)\n"
	@printf "$(CYAN)Local Host:     $(Local_HOST) $(RESET)\n"
	@printf "$(CYAN)Local Port:     $(Local_PORT) $(RESET)\n"
	@printf "$(CYAN)Local User:     $(Local_USER) $(RESET)\n"
	@printf "$(CYAN)Local Database: $(Local_DB) $(RESET)\n"
	@printf "$(CYAN)Local Password: $(shell printf $(Local_PASSWORD) | sed 's/./*/g') $(RESET)\n"

	@printf "$(BLUE)Current MySQL Configuration:$(RESET)\n"
	@printf "$(CYAN)Remote Host:     $(Remote_HOST) $(RESET)\n"
	@printf "$(CYAN)Remote Port:     $(Remote_PORT) $(RESET)\n"
	@printf "$(CYAN)Remote User:     $(Remote_USER) $(RESET)\n"
	@printf "$(CYAN)Remote Database: $(Remote_DB) $(RESET)\n"
	@printf "$(CYAN)Remote Password: $(shell printf $(Remote_PASSWORD) | sed 's/./*/g') $(RESET)\n"

# =============================================================================
# Setup and Installation
# =============================================================================

install: ## 🔧 Install MySQL client tools
	@printf "$(BLUE)Installing MySQL client tools...$(RESET)\n"
	uv run python cli.py install-tools
	@printf "$(GREEN)✅ Installation completed$(RESET)\n"

check-local: ## 🔍 Check MySQL local server version and connectivity
	@printf "$(BLUE)Checking MySQL server connection...$(RESET)\n"
	uv run python cli.py check-version \
		--host $(LOCAL_HOST) \
		--port $(LOCAL_PORT) \
		--user $(LOCAL_USER) \
		--password $(LOCAL_PASSWORD) \
		--database $(LOCAL_DB)

check-remote: ## 🔍 Check MySQL remote server version and connectivity
	@printf "$(BLUE)Checking MySQL server connection...$(RESET)\n"
	uv run python cli.py check-version \
		--host $(REMOTE_HOST) \
		--port $(REMOTE_PORT) \
		--user $(REMOTE_USER) \
		--password $(REMOTE_PASSWORD) \
		--database $(REMOTE_DB)

# =============================================================================
# Management
# =============================================================================

list: ## 📄 List all available backup files
	@printf "$(BLUE)Listing available backup files...$(RESET)\n"
	uv run python cli.py list

clean: ## 🧹 Clean old backup files (older than 7 days)
	@printf "$(BLUE)Cleaning old backup files...$(RESET)\n"
	@find ./temp/backup_mysql -name "*_mysql_backup_*" -type f -mtime +7 -delete 2>/dev/null || true
	@printf "$(GREEN)✅ Cleanup completed$(RESET)\n"

# =============================================================================
# Backup Operations
# =============================================================================

backup: ## 💾 Create database backup from remote with compression
	@printf "$(BLUE)Starting database backup...$(RESET)\n"
	@printf "$(BLUE)Host: $(REMOTE_HOST):$(REMOTE_PORT) | Database: $(REMOTE_DB)$(RESET)\n"
	uv run python cli.py backup \
		--host $(REMOTE_HOST) \
		--port $(REMOTE_PORT) \
		--user $(REMOTE_USER) \
		--password $(REMOTE_PASSWORD) \
		--database $(REMOTE_DB) \
		--compress \
		--single-transaction \
		--routines \
		--triggers
	@printf "$(GREEN)✅ Backup completed successfully$(RESET)\n"

# =============================================================================
# Restore Operations
# =============================================================================

restore-local: ## 📥 Restore latest backup to local MySQL from remote
	@printf "$(BLUE)Starting database restore to local MySQL...$(RESET)\n"
	uv run python cli.py restore \
		--host $(LOCAL_HOST) \
		--port $(LOCAL_PORT) \
		--user $(LOCAL_USER) \
		--password $(LOCAL_PASSWORD) \
		--database $(LOCAL_DB) \
		--force
	@printf "$(GREEN)✅ Local restore completed$(RESET)\n"

restore-remote: ## 📥 Restore latest backup to remote MySQL from local
	@printf "$(BLUE)Starting database restore to remote MySQL...$(RESET)\n"
	uv run python cli.py restore \
		--host $(REMOTE_HOST) \
		--port $(REMOTE_PORT) \
		--user $(REMOTE_USER) \
		--password $(REMOTE_PASSWORD) \
		--database $(REMOTE_DB) \
		--force
	@printf "$(GREEN)✅ Remote restore completed$(RESET)\n"

# =============================================================================
# Quick Actions
# =============================================================================

migrate: ## 🔀 Database migration (backup from remote → restore to local)
	@printf "$(BLUE)Starting database migration...$(RESET)\n"
	@printf "$(BLUE)Step 1: Creating backup from remote server...$(RESET)\n"
	@$(MAKE) backup
	@printf "$(BLUE)Step 2: Restoring to local database...$(RESET)\n"
	@$(MAKE) restore-local
	@printf "$(GREEN)✅ Database migration completed$(RESET)\n"