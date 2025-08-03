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
	@echo "$(BLUE)Backup Directory Status:$(RESET)"
	@ls -la ./temp/backup_mysql/ 2>/dev/null || echo "$(YELLOW)No backup directory found$(RESET)"
	@echo ""
	@echo "$(BLUE)Disk Usage:$(RESET)"
	@du -sh ./temp/backup_mysql/ 2>/dev/null || echo "$(YELLOW)No backup files found$(RESET)"

config: ## ⚙️ Show current configuration
	@echo "$(BLUE)Current Local MySQL Configuration:$(RESET)"
	@echo "$(CYAN)Local Host:$(RESET)     $(Local_HOST) $(RESET)"
	@echo "$(CYAN)Local Port:$(RESET)     $(Local_PORT)"
	@echo "$(CYAN)Local User:$(RESET)     $(Local_USER)"
	@echo "$(CYAN)Local Database:$(RESET) $(Local_DB)"
	@echo "$(CYAN)Local Password:$(RESET) $(shell echo $(Local_PASSWORD) | sed 's/./*/g')"

	@echo "$(BLUE)Current MySQL Configuration:$(RESET)"
	@echo "$(CYAN)Remote Host:$(RESET)     $(Remote_HOST)"
	@echo "$(CYAN)Remote Port:$(RESET)     $(Remote_PORT)"
	@echo "$(CYAN)Remote User:$(RESET)     $(Remote_USER)"
	@echo "$(CYAN)Remote Database:$(RESET) $(Remote_DB)"
	@echo "$(CYAN)Remote Password:$(RESET) $(shell echo $(Remote_PASSWORD) | sed 's/./*/g')"

# =============================================================================
# Setup and Installation
# =============================================================================

install: ## 🔧 Install MySQL client tools
	@echo "$(BLUE)Installing MySQL client tools...$(RESET)"
	uv run python cli.py install-tools
	@echo "$(GREEN)✅ Installation completed$(RESET)"

check-local: ## 🔍 Check MySQL local server version and connectivity
	@echo "$(BLUE)Checking MySQL server connection...$(RESET)"
	uv run python cli.py check-version \
		--host $(LOCAL_HOST) \
		--port $(LOCAL_PORT) \
		--user $(LOCAL_USER) \
		--password $(LOCAL_PASSWORD) \
		--database $(LOCAL_DB)

check-remote: ## 🔍 Check MySQL remote server version and connectivity
	@echo "$(BLUE)Checking MySQL server connection...$(RESET)"
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
	@echo "$(BLUE)Listing available backup files...$(RESET)"
	uv run python cli.py list

clean: ## 🧹 Clean old backup files (older than 7 days)
	@echo "$(BLUE)Cleaning old backup files...$(RESET)"
	@find ./temp/backup_mysql -name "*_mysql_backup_*" -type f -mtime +7 -delete 2>/dev/null || true
	@echo "$(GREEN)✅ Cleanup completed$(RESET)"

# =============================================================================
# Backup Operations
# =============================================================================

backup: ## 💾 Create database backup from remote with compression
	@echo "$(BLUE)Starting database backup...$(RESET)"
	@echo "$(BLUE)Host: $(REMOTE_HOST):$(REMOTE_PORT) | Database: $(REMOTE_DB)$(RESET)"
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
	@echo "$(GREEN)✅ Backup completed successfully$(RESET)"

# =============================================================================
# Restore Operations
# =============================================================================

restore-local: ## 📥 Restore latest backup to local MySQL from remote
	@echo "$(BLUE)Starting database restore to local MySQL...$(RESET)"
	uv run python cli.py restore \
		--host $(LOCAL_HOST) \
		--port $(LOCAL_PORT) \
		--user $(LOCAL_USER) \
		--password $(LOCAL_PASSWORD) \
		--database $(LOCAL_DB) \
		--force
	@echo "$(GREEN)✅ Local restore completed$(RESET)"

restore-remote: ## 📥 Restore latest backup to remote MySQL from local
	@echo "$(BLUE)Starting database restore to remote MySQL...$(RESET)"
	uv run python cli.py restore \
		--host $(REMOTE_HOST) \
		--port $(REMOTE_PORT) \
		--user $(REMOTE_USER) \
		--password $(REMOTE_PASSWORD) \
		--database $(REMOTE_DB) \
		--force
	@echo "$(GREEN)✅ Remote restore completed$(RESET)"

# =============================================================================
# Quick Actions
# =============================================================================

migrate: ## 🔀 Database migration (backup from remote → restore to local)
	@echo "$(BLUE)Starting database migration...$(RESET)"
	@echo "$(BLUE)Step 1: Creating backup from remote server...$(RESET)"
	@$(MAKE) backup
	@echo "$(BLUE)Step 2: Restoring to local database...$(RESET)"
	@$(MAKE) restore-local
	@echo "$(GREEN)✅ Database migration completed$(RESET)"