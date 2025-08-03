# =============================================================================
# RCABench Makefile
# =============================================================================
# 这个Makefile提供了RCABench项目的所有构建、部署和开发工具
# 使用 'make help' 查看所有可用命令

# =============================================================================
# 配置变量
# =============================================================================

# 基础配置
DEFAULT_REPO ?= 10.10.10.240/library
NS          ?= exp
NS_PREFIX	?= ts
PORT        ?= 30080

# 目录配置
SRC_DIR = src
SDK_DIR = sdk/python-gen

# Chaos类型配置
CHAOS_TYPES ?= dnschaos httpchaos jvmchaos networkchaos podchaos stresschaos timechaos

# 颜色定义
BLUE    := \033[1;34m
GREEN   := \033[1;32m
YELLOW  := \033[1;33m
RED     := \033[1;31m
CYAN    := \033[1;36m
GRAY    := \033[90m
RESET   := \033[0m

# =============================================================================
# 声明所有非文件目标
# =============================================================================
.PHONY: help build run debug swagger import clean-finalizers delete-all-chaos k8s-resources ports \
        install-hooks deploy-ts swag-init generate-sdk release \
        check-prerequisites setup-dev-env clean-all status logs

# =============================================================================
# 默认目标
# =============================================================================
.DEFAULT_GOAL := help

# =============================================================================
# 帮助信息
# =============================================================================
help:  ## 📖 显示所有可用命令
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(BLUE)║                    RCABench 项目管理工具                     ║$(RESET)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(YELLOW)使用方法:$(RESET) make $(CYAN)<目标名称>$(RESET)"
	@echo "$(YELLOW)示例:$(RESET) make run, make help, make clean-all"
	@echo ""
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "$(YELLOW)可用命令:$(RESET)\n"; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  $(CYAN)%-25s$(RESET) $(GRAY)%s$(RESET)\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)快速开始:$(RESET)"
	@echo "  $(CYAN)make check-prerequisites$(RESET)  - 检查环境依赖"
	@echo "  $(CYAN)make run$(RESET)                 - 构建并部署应用"
	@echo "  $(CYAN)make status$(RESET)              - 查看应用状态"
	@echo "  $(CYAN)make logs$(RESET)                - 查看应用日志"

# =============================================================================
# 环境检查和设置
# =============================================================================

check-prerequisites: ## 🔍 检查开发环境依赖
	@echo "$(BLUE)🔍 检查开发环境依赖...$(RESET)"
	@echo "$(GRAY)检查 kubectl...$(RESET)"
	@command -v kubectl >/dev/null 2>&1 || { echo "$(RED)❌ kubectl 未安装$(RESET)"; exit 1; }
	@echo "$(GREEN)✅ kubectl 已安装$(RESET)"
	@echo "$(GRAY)检查 skaffold...$(RESET)"
	@command -v skaffold >/dev/null 2>&1 || { echo "$(RED)❌ skaffold 未安装$(RESET)"; exit 1; }
	@echo "$(GREEN)✅ skaffold 已安装$(RESET)"
	@echo "$(GRAY)检查 docker...$(RESET)"
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)❌ docker 未安装$(RESET)"; exit 1; }
	@echo "$(GREEN)✅ docker 已安装$(RESET)"
	@echo "$(GRAY)检查 helm...$(RESET)"
	@command -v helm >/dev/null 2>&1 || { echo "$(RED)❌ helm 未安装$(RESET)"; exit 1; }
	@echo "$(GREEN)✅ helm 已安装$(RESET)"
	@echo "$(GREEN)🎉 所有依赖检查通过！$(RESET)"

setup-dev-env: check-prerequisites ## 🛠️ 设置开发环境
	@echo "$(BLUE)🛠️ 设置开发环境...$(RESET)"
	@echo "$(GRAY)安装 Git hooks...$(RESET)"
	@$(MAKE) install-hooks
	@echo "$(GREEN)✅ 开发环境设置完成！$(RESET)"

# =============================================================================
# 构建和部署
# =============================================================================

run: check-prerequisites ## 🚀 构建并部署应用 (使用 skaffold)
	@echo "$(BLUE)🔄 开始部署流程...$(RESET)"
	# @if $(MAKE) check-db 2>/dev/null; then \
	# 	echo "$(YELLOW)📄 备份现有数据库...$(RESET)"; \
	# 	$(MAKE) -C scripts/hack/backup_mysql backup; \
	# else \
	# 	echo "$(YELLOW)⚠️  MySQL 未运行，跳过备份$(RESET)"; \
	# fi
	@echo "$(GRAY)使用 skaffold 部署...$(RESET)"
	skaffold run --default-repo=$(DEFAULT_REPO)
	@echo "$(BLUE)⏳ 等待部署稳定...$(RESET)"
	$(MAKE) wait-for-deployment
	@echo "$(GREEN)🎉 部署完成！$(RESET)"

wait-for-deployment: ## ⏳ 等待部署就绪
	@echo "$(BLUE)⏳ 等待所有部署就绪...$(RESET)"
	kubectl wait --for=condition=available --timeout=300s deployment --all -n $(NS)
	@echo "$(GREEN)✅ 所有部署已就绪$(RESET)"

build: ## 🔨 仅构建应用 (不部署)
	@echo "$(BLUE)🔨 构建应用...$(RESET)"
	skaffold build --default-repo=$(DEFAULT_REPO)
	@echo "$(GREEN)✅ 构建完成$(RESET)"

# =============================================================================
# 数据库管理
# =============================================================================

## 检查 数据库 状态
check-db: 
	@echo "$(BLUE)🔍 检查 MySQL 状态...$(RESET)"
	@if kubectl get pods -n $(NS) -l app=rcabench-mysql --field-selector=status.phase=Running | grep -q rcabench-mysql; then \
		echo "$(GREEN)✅ MySQL 正在运行$(RESET)"; \
	else \
		echo "$(RED)❌ MySQL 在命名空间 $(NS) 中未运行$(RESET)"; \
		echo "$(GRAY)可用 Pods:$(RESET)"; \
		kubectl get pods -n $(NS) -l app=rcabench-mysql || echo "$(GRAY)未找到 MySQL pods$(RESET)"; \
		exit 1; \
	fi

reset-db: ## 🗑️ 重置 MySQL 数据库 (⚠️ 将删除所有数据)
	@echo "$(RED)⚠️  警告：这将删除所有数据库数据！$(RESET)"
	@read -p "确认继续？(y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@if $(MAKE) check-db 2>/dev/null; then \
		echo "$(YELLOW)📄 备份现有数据库...$(RESET)"; \
		$(MAKE) -C scripts/hack/backup_psql backup; \
	else \
		echo "$(YELLOW)⚠️  MySQL 未运行，跳过备份$(RESET)"; \
	fi
	@echo "$(BLUE)🗑️  重置命名空间 $(NS) 中的 MySQL 数据库...$(RESET)"
	helm uninstall rcabench -n $(NS) || true
	@echo "$(BLUE)⏳ 等待 Pods 终止...$(RESET)"
	@while kubectl get pods -n $(NS) -l app=rcabench-mysql 2>/dev/null | grep -q .; do \
		echo "$(GRAY)  仍在等待 Pods 终止...$(RESET)"; \
		sleep 2; \
	done
	@echo "$(GREEN)✅ 所有 Pods 已终止$(RESET)"
	kubectl delete pvc rcabench-mysql-data -n $(NS) || true
	@echo "$(BLUE)⏳ 等待 PVC 删除...$(RESET)"
	@while kubectl get pvc -n $(NS) | grep -q rcabench-mysql-data; do \
		echo "$(GRAY)  仍在等待 PVC 删除...$(RESET)"; \
		sleep 2; \
	done
	@echo "$(GREEN)✅ PVC 删除成功$(RESET)"
	@echo "$(GREEN)✅ 数据库重置完成。重新部署中...$(RESET)"
	$(MAKE) run
	@echo "$(GREEN)🚀 应用重新部署成功。$(RESET)"
	$(MAKE) -C scripts/hack/backup_mysql migrate
	@echo "$(GREEN)📦 从备份恢复数据库。$(RESET)"

# =============================================================================
# 开发工具
# =============================================================================

local-debug: ## 🐛 启动本地调试环境
	@echo "$(BLUE)🚀 启动基础服务...$(RESET)"
	docker compose down && \
	docker compose up redis mysql jaeger buildkitd -d
	@echo "$(BLUE)🧹 清理 Kubernetes Jobs...$(RESET)"
	kubectl delete jobs --all -n $(NS)
	@echo "$(BLUE)📦 从正式环境备份 Redis...$(RESET)"
	$(MAKE) -C scripts/hack/backup_redis restore-local
	@echo "$(BLUE)🗄️ 从正式环境备份数据库...$(RESET)"
	$(MAKE) -C scripts/hack/backup_psql restore-local
	@echo "$(GREEN)✅ 环境准备完成！$(RESET)"
	@read -p "是否现在启动本地应用 (y/N)" start_app; \
	if [ "$$start_app" = "n" ] || [ "$$start_app" = "N" ]; then \
		echo "$(YELLOW)⏸️  本地应用未启动，你可以稍后手动启动: $(RESET)"; \
		echo "$(GRAY)cd $(SRC_DIR) && go run main.go both --port 8082$(RESET)"; \
	else \
		echo "$(BLUE)⌛️ 启动本地应用...$(RESET)"; \
		cd $(SRC_DIR) && go run main.go both --port 8082; \
	fi

import: ## 📦 导入最新版本的 chaos-experiment 库
	@echo "$(BLUE)📦 导入最新版本的 chaos-experiment 库...$(RESET)"
	cd $(SRC_DIR) && \
	go get -u github.com/LGU-SE-Internal/chaos-experiment@injectionv2 && \
	go mod tidy
	@echo "$(GREEN)✅ 依赖更新完成$(RESET)"

# =============================================================================
# Chaos 管理
# =============================================================================

define get_target_namespaces
    kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | tr ' ' '\n' | grep "^$(NS_PREFIX)[0-9]$$" | sort
endef

clean-finalizers: ## 🧹 清理所有 chaos 资源的finalizer
	@echo "$(BLUE)🧹 清理 chaos finalizers...$(RESET)"
	@echo "$(GRAY)动态获取以 $(NS_PREFIX) 为前缀的命名空间...$(RESET)"
	@namespaces=$$($(call get_target_namespaces)); \
	echo "$(CYAN)找到以下命名空间:$(RESET)"; \
	for ns in $$namespaces; do \
		echo "  - $$ns"; \
	done; \
	echo "$(GRAY)总计: $$(echo "$$namespaces" | wc -w) 个命名空间$(RESET)"; \
	echo ""; \
	for ns in $$namespaces; do \
		echo "$(BLUE)🔄 处理命名空间: $$ns$(RESET)"; \
		for type in $(CHAOS_TYPES); do \
			echo "$(GRAY)清理 $$type...$(RESET)"; \
			kubectl get $$type -n $$ns -o jsonpath='{range .items[*]}{.metadata.namespace}{":"}{.metadata.name}{"\n"}{end}' | \
			while IFS=: read -r ns name; do \
				[ -n "$$name" ] && kubectl patch $$type "$$name" -n "$$ns" --type=merge -p '{"metadata":{"finalizers":[]}}'; \
			done; \
		done; \
	done
	@echo "$(GREEN)✅ Finalizer 清理完成$(RESET)"

delete-all-chaos: ## 🗑️ 删除所有 chaos 资源
	@echo "$(BLUE)🗑️ 删除 chaos 资源...$(RESET)"
	@echo "$(GRAY)动态获取以 $(NS_PREFIX) 为前缀的命名空间...$(RESET)"
	@namespaces=$$($(call get_target_namespaces)); \
	echo "$(CYAN)找到以下命名空间:$(RESET)"; \
	for ns in $$namespaces; do \
		echo "  - $$ns"; \
	done; \
	echo "$(GRAY)总计: $$(echo "$$namespaces" | wc -w) 个命名空间$(RESET)"; \
	echo ""; \
	for ns in $$namespaces; do \
		echo "$(BLUE)🔄 处理命名空间: $$ns$(RESET)"; \
		for type in $(CHAOS_TYPES); do \
			echo "$(GRAY)删除 $$type...$(RESET)"; \
			kubectl delete $$type --all -n $$ns; \
		done; \
	done
	@echo "$(GREEN)✅ Chaos 资源删除完成$(RESET)"

# =============================================================================
# Kubernetes 管理
# =============================================================================

k8s-resources: ## 📊 显示所有 jobs 和 pods
	@echo "$(BLUE)📊 命名空间 $(NS) 中的资源:$(RESET)"
	@echo "$(YELLOW)Jobs:$(RESET)"
	@kubectl get jobs -n $(NS)
	@echo "$(YELLOW)Pods:$(RESET)"
	@kubectl get pods -n $(NS)

status: ## 📈 查看应用状态
	@echo "$(BLUE)📈 应用状态概览:$(RESET)"
	@echo "$(YELLOW)命名空间: $(NS)$(RESET)"
	@echo "$(GRAY)Deployments:$(RESET)"
	@kubectl get deployments -n $(NS)
	@echo "$(GRAY)Services:$(RESET)"
	@kubectl get services -n $(NS)
	@echo "$(GRAY)Pods 状态:$(RESET)"
	@kubectl get pods -n $(NS) -o wide

logs: ## 📋 查看应用日志
	@echo "$(BLUE)📋 应用日志:$(RESET)"
	@echo "$(YELLOW)选择要查看日志的 Pod:$(RESET)"
	@kubectl get pods -n $(NS) --no-headers -o custom-columns=":metadata.name" | head -10
	@echo "$(GRAY)使用 'kubectl logs <pod-name> -n $(NS)' 查看特定 Pod 的日志$(RESET)"

ports: ## 🔌 端口转发服务
	@echo "$(BLUE)🔌 启动端口转发...$(RESET)"
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &
	@echo "$(GREEN)✅ 端口转发已启动 (8081:8081)$(RESET)"
	@echo "$(GRAY)访问地址: http://localhost:8081$(RESET)"

# =============================================================================
# Git 管理
# =============================================================================

install-hooks: ## 🔧 安装 pre-commit hooks
	@echo "$(BLUE)🔧 安装 Git hooks...$(RESET)"
	chmod +x scripts/hooks/pre-commit
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)✅ Git hooks 安装完成$(RESET)"

# =============================================================================
# SDK 生成
# =============================================================================

swagger: swag-init generate-sdk ## 📚 生成完整的 Swagger 文档和 SDK

## 初始化 Swagger 文档
swag-init:
	@echo "$(BLUE)📝 初始化 Swagger 文档...$(RESET)"
	swag init -d ./$(SRC_DIR) --parseDependency --parseDepth 1 --output ./$(SRC_DIR)/docs
	@echo "$(GREEN)✅ Swagger 文档生成完成$(RESET)"

## 从 Swagger 文档生成 Python SDK
generate-sdk: swag-init
	@echo "$(BLUE)🐍 生成 Python SDK...$(RESET)"
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		openapitools/openapi-generator-cli:latest generate \
		-i /local/$(SRC_DIR)/docs/swagger.json \
		-g python \
		-o /local/$(SDK_DIR) \
		-c /local/.openapi-generator/config.properties \
		--additional-properties=packageName=openapi,projectName=rcabench
	@echo "$(BLUE)📦 后处理生成的 SDK...$(RESET)"
	./scripts/mv-generated-sdk.sh
	@echo "$(GREEN)✅ Python SDK 生成完成$(RESET)"

# =============================================================================
# 发布管理
# =============================================================================

release: ## 🏷️ 发布新版本 (用法: make release VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ 请提供版本号: make release VERSION=1.0.1$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)🏷️ 发布版本 $(VERSION)...$(RESET)"
	./scripts/release.sh $(VERSION)

release-dry-run: ## 🧪 发布流程试运行 (用法: make release-dry-run VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)❌ 请提供版本号: make release-dry-run VERSION=1.0.1$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)🧪 试运行发布流程 $(VERSION)...$(RESET)"
	./scripts/release.sh $(VERSION) --dry-run

upload: ## 📤 上传 SDK 包
	@echo "$(BLUE)📤 上传 SDK 包...$(RESET)"
	$(MAKE) -C sdk/python upload
	@echo "$(GREEN)✅ SDK 上传完成$(RESET)"

# =============================================================================
# 清理和维护
# =============================================================================

clean-all: ## 🧹 清理所有资源
	@echo "$(BLUE)🧹 清理所有资源...$(RESET)"
	@echo "$(YELLOW)⚠️  这将删除所有部署的资源！$(RESET)"
	@read -p "确认继续？(y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "$(GRAY)删除 Helm 发布...$(RESET)"
	helm uninstall rcabench -n $(NS) || true
	@echo "$(GRAY)删除命名空间...$(RESET)"
	kubectl delete namespace $(NS) || true
	@echo "$(GRAY)停止端口转发...$(RESET)"
	pkill -f "kubectl port-forward" || true
	@echo "$(GREEN)✅ 清理完成$(RESET)"

# =============================================================================
# 实用工具
# =============================================================================

restart: ## 🔄 重启应用
	@echo "$(BLUE)🔄 重启应用...$(RESET)"
	kubectl rollout restart deployment --all -n $(NS)
	@echo "$(GREEN)✅ 应用重启完成$(RESET)"

scale: ## 📏 扩展部署 (用法: make scale DEPLOYMENT=app REPLICAS=3)
	@if [ -z "$(DEPLOYMENT)" ] || [ -z "$(REPLICAS)" ]; then \
		echo "$(RED)❌ 请提供部署名称和副本数: make scale DEPLOYMENT=app REPLICAS=3$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)📏 扩展部署 $(DEPLOYMENT) 到 $(REPLICAS) 个副本...$(RESET)"
	kubectl scale deployment $(DEPLOYMENT) --replicas=$(REPLICAS) -n $(NS)
	@echo "$(GREEN)✅ 扩展完成$(RESET)"

# =============================================================================
# 信息显示
# =============================================================================

info: ## ℹ️ 显示项目信息
	@echo "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(BLUE)║                        RCABench 项目信息                     ║$(RESET)"
	@echo "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)"
	@echo "$(YELLOW)配置信息:$(RESET)"
	@echo "  $(CYAN)默认仓库:$(RESET) $(DEFAULT_REPO)"
	@echo "  $(CYAN)命名空间:$(RESET) $(NS)"
	@echo "  $(CYAN)端口:$(RESET) $(PORT)"
	@echo "  $(CYAN)控制器目录:$(RESET) $(SRC_DIR)"
	@echo "  $(CYAN)SDK 目录:$(RESET) $(SDK_DIR)"
	@echo ""
	@echo "$(YELLOW)Chaos 类型:$(RESET)"
	@for type in $(CHAOS_TYPES); do \
		echo "  - $$type"; \
	done