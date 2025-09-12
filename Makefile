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

BACKUP_DATA ?= $(shell [ -t 0 ] && echo "ask" || echo "no")
START_APP   ?= $(shell [ -t 0 ] && echo "ask" || echo "yes")

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
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║                    RCABench 项目管理工具                     ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"
	@printf "$(YELLOW)使用方法:$(RESET) make $(CYAN)<目标名称>$(RESET)\n"
	@printf "$(YELLOW)示例:$(RESET)\n make run, make help, make clean-all \n"
	@printf "\n"
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "$(YELLOW)可用命令:$(RESET)\n"; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  $(CYAN)%-25s$(RESET) $(GRAY)%s$(RESET)\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)
	@printf "\n"
	@printf "$(YELLOW)快速开始:$(RESET)\n"
	@printf "  $(CYAN)make check-prerequisites$(RESET) - 检查环境依赖\n"
	@printf "  $(CYAN)make run$(RESET)                 - 构建并部署应用\n"
	@printf "  $(CYAN)make status$(RESET)              - 查看应用状态\n"
	@printf "  $(CYAN)make logs$(RESET)                - 查看应用日志\n"

# =============================================================================
# 环境检查和设置
# =============================================================================

check-prerequisites: ## 🔍 检查开发环境依赖
	@printf "$(BLUE)🔍 检查开发环境依赖...$(RESET)\n"
	@printf "$(GRAY)检查 kubectl...$(RESET)\n"
	@command -v kubectl >/dev/null 2>&1 || { printf "$(RED)❌ kubectl 未安装$(RESET)"; exit 1; }
	@printf "$(GREEN)✅ kubectl 已安装$(RESET)\n"
	@printf "$(GRAY)检查 skaffold...$(RESET)\n"
	@command -v skaffold >/dev/null 2>&1 || { printf "$(RED)❌ skaffold 未安装$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ skaffold 已安装$(RESET)\n"
	@printf "$(GRAY)检查 docker...$(RESET)\n"
	@command -v docker >/dev/null 2>&1 || { printf "$(RED)❌ docker 未安装$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ docker 已安装$(RESET)\n"
	@printf "$(GRAY)检查 helm...$(RESET)\n"
	@command -v helm >/dev/null 2>&1 || { printf "$(RED)❌ helm 未安装$(RESET)\n"; exit 1; }
	@printf "$(GREEN)✅ helm 已安装$(RESET)\n"
	@printf "$(GREEN)🎉 所有依赖检查通过！$(RESET)\n"

setup-dev-env: check-prerequisites ## 🛠️  设置开发环境
	@printf "$(BLUE)🛠️ 设置开发环境...$(RESET)\n"
	@printf "$(GRAY)安装 Git hooks...$(RESET)\n"
	@$(MAKE) install-hooks
	@printf "$(GREEN)✅ 开发环境设置完成！$(RESET)\n"

# =============================================================================
# 构建和部署
# =============================================================================

run: check-prerequisites ## 🚀 构建并部署应用 (使用 skaffold)
	@printf "$(BLUE)🔄 开始部署流程...$(RESET)\n"
	@if $(MAKE) check-db 2>/dev/null; then \
		printf "$(YELLOW)📄 备份现有数据库...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_mysql backup; \
	else \
		printf "$(YELLOW)⚠️ 数据库未运行，跳过备份$(RESET)\n"; \
	fi
	@printf "$(GRAY)使用 skaffold 部署...$(RESET)\n"
	skaffold run --default-repo=$(DEFAULT_REPO)
	@printf "$(BLUE)⏳ 等待部署稳定...$(RESET)\n"
	$(MAKE) wait-for-deployment
	@printf "$(GREEN)🎉 部署完成！$(RESET)\n"

wait-for-deployment: ## ⏳ 等待部署就绪
	@printf "$(BLUE)⏳ 等待所有部署就绪...$(RESET)\n"
	kubectl wait --for=condition=available --timeout=300s deployment --all -n $(NS)
	@printf "$(GREEN)✅ 所有部署已就绪$(RESET)\n"

build: ## 🔨 仅构建应用 (不部署)
	@printf "$(BLUE)🔨 构建应用...$(RESET)\n"
	skaffold build --default-repo=$(DEFAULT_REPO)
	@printf "$(GREEN)✅ 构建完成$(RESET)\n"

# =============================================================================
# 数据库管理
# =============================================================================

## 检查数据库状态
check-db: 
	@printf "$(BLUE)🔍 检查数据库状态...$(RESET)\n"
	@if kubectl get pods -n $(NS) -l app=rcabench-mysql --field-selector=status.phase=Running | grep -q rcabench-mysql; then \
		printf "$(GREEN)✅ 数据库正在运行$(RESET)\n"; \
	else \
		printf "$(RED)❌ 数据库在命名空间 $(NS) 中未运行$(RESET)\n"; \
		printf "$(GRAY)可用 Pods:$(RESET)\n"; \
		kubectl get pods -n $(NS) -l app=rcabench-mysql || printf "$(GRAY)未找到数据库 pods$(RESET)\n"; \
		exit 1; \
	fi

reset-db: ## 🗑️  重置数据库 (⚠️ 将删除所有数据)
	@printf "$(RED)⚠️  警告：这将删除所有数据库数据！$(RESET)\n"
	@read -p "确认继续？(y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@if $(MAKE) check-db 2>/dev/null; then \
		printf "$(YELLOW)📄 备份现有数据库...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_psql backup; \
	else \
		printf "$(YELLOW)⚠️ 数据库未运行，跳过备份$(RESET)\n"; \
	fi
	@printf "$(BLUE)🗑️  重置命名空间 $(NS) 中的数据库...$(RESET)\n"
	helm uninstall rcabench -n $(NS) || true
	@printf "$(BLUE)⏳ 等待 Pods 终止...$(RESET)\n"
	@while kubectl get pods -n $(NS) -l app=rcabench-mysql 2>/dev/null | grep -q .; do \
		printf "$(GRAY)  仍在等待 Pods 终止...$(RESET)\n"; \
		sleep 2; \
	done
	@printf "$(GREEN)✅ 所有 Pods 已终止$(RESET)\n"
	kubectl delete pvc rcabench-mysql-data -n $(NS) || true
	@printf "$(BLUE)⏳ 等待 PVC 删除...$(RESET)\n"
	@while kubectl get pvc -n $(NS) | grep -q rcabench-mysql-data; do \
		printf "$(GRAY)  仍在等待 PVC 删除...$(RESET)\n"; \
		sleep 2; \
	done
	@printf "$(GREEN)✅ PVC 删除成功$(RESET)\n"
	@printf "$(GREEN)✅ 数据库重置完成。重新部署中...$(RESET)\n"
	$(MAKE) run
	@printf "$(GREEN)🚀 应用重新部署成功。$(RESET)\n"
	$(MAKE) -C scripts/hack/backup_mysql migrate
	@printf "$(GREEN)📦 从备份恢复数据库。$(RESET)\n"

# =============================================================================
# 开发工具
# =============================================================================

local-debug: ## 🐛 启动本地调试环境
	@printf "$(BLUE)🚀 启动基础服务...$(RESET)\n"
	@if ! docker compose down; then \
		printf "$(RED)❌ Docker Compose 停止失败$(RESET)\n"; \
		exit 1; \
	fi
	@if ! docker compose up redis mysql jaeger buildkitd -d; then \
		printf "$(RED)❌ Docker Compose 启动失败$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🧹 清理 Kubernetes Jobs...$(RESET)\n"
	@kubectl delete jobs --all -n $(NS) || printf "$(YELLOW)⚠️  清理 Jobs 失败或无 Jobs 需要清理$(RESET)\n"
	@set -e; \
	if [ "$(BACKUP_DATA)" = "ask" ]; then \
		read -p "是否备份数据 (y/n)? " use_backup; \
	elif [ "$(BACKUP_DATA)" = "yes" ]; then \
		use_backup="y"; \
	else \
		use_backup="n"; \
	fi; \
	if [ "$$use_backup" = "y" ] || [ "$$use_backup" = "Y" ]; then \
		printf "$(BLUE)📦 从正式环境备份 Redis...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_redis restore-local; \
		printf "$(BLUE)🗄️ 从正式环境备份数据库...$(RESET)\n"; \
		$(MAKE) -C scripts/hack/backup_mysql migrate; \
		printf "$(GREEN)✅ 环境准备完成！$(RESET)\n"; \
	fi; \
	if [ "$(START_APP)" = "ask" ]; then \
		read -p "是否现在启动本地应用 (y/n)? " start_app; \
	elif [ "$(START_APP)" = "yes" ]; then \
		start_app="y"; \
	else \
		start_app="n"; \
	fi; \
	if [ "$$start_app" = "n" ] || [ "$$start_app" = "N" ]; then \
		printf "$(YELLOW)⏸️  本地应用未启动，你可以稍后手动启动:$(RESET)\n"; \
		printf "$(GRAY)cd $(SRC_DIR) && go run main.go both --port 8082$(RESET)\n"; \
	else \
		printf "$(BLUE)⌛️ 启动本地应用...$(RESET)\n"; \
		cd $(SRC_DIR) && go run main.go both --port 8082; \
	fi

local-debug-auto: ## 🤖 启动本地调试环境 (自动模式，无交互)
	@$(MAKE) local-debug BACKUP_DATA=yes START_APP=yes

local-debug-minimal: ## 🚀 启动本地调试环境 (最小模式，无备份无自动启动)
	@$(MAKE) local-debug BACKUP_DATA=no START_APP=no

import: ## 📦 导入最新版本的 chaos-experiment 库
	@printf "$(BLUE)📦 导入最新版本的 chaos-experiment 库...$(RESET)\n"
	cd $(SRC_DIR) && \
	go get -u github.com/rcabench/chaos-experiment@injectionv2 && \
	go mod tidy
	@printf "$(GREEN)✅ 依赖更新完成$(RESET)\n"

# =============================================================================
# Chaos 管理
# =============================================================================

define get_target_namespaces
    kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null | tr ' ' '\n' | grep "^$(NS_PREFIX)[0-9]$$" | sort
endef

clean-finalizers: ## 🧹 清理所有 chaos 资源的finalizer
	@printf "$(BLUE)🧹 清理 chaos finalizers...$(RESET)\n"
	@printf "$(GRAY)动态获取以 $(NS_PREFIX) 为前缀的命名空间...$(RESET)\n"
	@namespaces=$$($(call get_target_namespaces)); \
	printf "$(CYAN)找到以下命名空间:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns"; \
	done; \
	printf "$(GRAY)总计: $$(printf "$$namespaces" | wc -w) 个命名空间$(RESET)\n"; \
	printf ""; \
	for ns in $$namespaces; do \
		printf "$(BLUE)🔄 处理命名空间: $$ns$(RESET)\n"; \
		for type in $(CHAOS_TYPES); do \
			printf "$(GRAY)清理 $$type...$(RESET)\n"; \
			kubectl get $$type -n $$ns -o jsonpath='{range .items[*]}{.metadata.namespace}{":"}{.metadata.name}{"\n"}{end}' | \
			while IFS=: read -r ns name; do \
				[ -n "$$name" ] && kubectl patch $$type "$$name" -n "$$ns" --type=merge -p '{"metadata":{"finalizers":[]}}'; \
			done; \
		done; \
	done
	@printf "$(GREEN)✅ Finalizer 清理完成$(RESET)\n"

delete-all-chaos: ## 🗑️  删除所有 chaos 资源
	@printf "$(BLUE)🗑️ 删除 chaos 资源...$(RESET)\n"
	@printf "$(GRAY)动态获取以 $(NS_PREFIX) 为前缀的命名空间...$(RESET)\n"
	@namespaces=$$($(call get_target_namespaces)); \
	printf "$(CYAN)找到以下命名空间:$(RESET)\n"; \
	for ns in $$namespaces; do \
		printf "  - $$ns"; \
	done; \
	printf "$(GRAY)总计: $$(printf "$$namespaces" | wc -w) 个命名空间$(RESET)\n"; \
	printf ""; \
	for ns in $$namespaces; do \
		printf "$(BLUE)🔄 处理命名空间: $$ns$(RESET)\n"; \
		for type in $(CHAOS_TYPES); do \
			printf "$(GRAY)删除 $$type...$(RESET)\n"; \
			kubectl delete $$type --all -n $$ns; \
		done; \
	done
	@printf "$(GREEN)✅ Chaos 资源删除完成$(RESET)\n"

# =============================================================================
# Kubernetes 管理
# =============================================================================

k8s-resources: ## 📊 显示所有 jobs 和 pods
	@printf "$(BLUE)📊 命名空间 $(NS) 中的资源:$(RESET)\n"
	@printf "$(YELLOW)Jobs:$(RESET)\n"
	@kubectl get jobs -n $(NS)
	@printf "$(YELLOW)Pods:$(RESET)\n"
	@kubectl get pods -n $(NS)

status: ## 📈 查看应用状态
	@printf "$(BLUE)📈 应用状态概览:$(RESET)\n"
	@printf "$(YELLOW)命名空间: $(NS)$(RESET)\n"
	@printf "$(GRAY)Deployments:$(RESET)\n"
	@kubectl get deployments -n $(NS)
	@printf "$(GRAY)Services:$(RESET)\n"
	@kubectl get services -n $(NS)
	@printf "$(GRAY)Pods 状态:$(RESET)\n"
	@kubectl get pods -n $(NS) -o wide

logs: ## 📋 查看应用日志
	@printf "$(BLUE)📋 应用日志:$(RESET)\n"
	@printf "$(YELLOW)选择要查看日志的 Pod:$(RESET)\n"
	@kubectl get pods -n $(NS) --no-headers -o custom-columns=":metadata.name" | head -10
	@printf "$(GRAY)使用 'kubectl logs <pod-name> -n $(NS)' 查看特定 Pod 的日志$(RESET)\n"

ports: ## 🔌 端口转发服务
	@printf "$(BLUE)🔌 启动端口转发...$(RESET)\n"
	kubectl port-forward svc/exp -n $(NS) --address 0.0.0.0 8081:8081 &
	@printf "$(GREEN)✅ 端口转发已启动 (8081:8081)$(RESET)\n"
	@printf "$(GRAY)访问地址: http://localhost:8081$(RESET)\n"

# =============================================================================
# Git 管理
# =============================================================================

install-hooks: ## 🔧 安装 pre-commit hooks
	@printf "$(BLUE)🔧 安装 Git hooks...$(RESET)\n"
	chmod +x scripts/hooks/pre-commit
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@printf "$(GREEN)✅ Git hooks 安装完成$(RESET)\n"

# =============================================================================
# SDK 生成
# =============================================================================

swagger: swag-init generate-sdk ## 📚 生成完整的 Swagger 文档和 SDK

## 初始化 Swagger 文档
swag-init:
	@printf "$(BLUE)📝 初始化 Swagger 文档...$(RESET)\n"
	swag init -d ./$(SRC_DIR) --parseDependency --parseDepth 1 --output ./$(SRC_DIR)/docs
	@printf "$(GREEN)✅ Swagger 文档生成完成$(RESET)\n"

## 从 Swagger 文档生成 Python SDK
generate-sdk: swag-init
	@printf "$(BLUE)🐍 生成 Python SDK...$(RESET)\n"
	docker run --rm -u $(shell id -u):$(shell id -g) -v $(shell pwd):/local \
		openapitools/openapi-generator-cli:latest generate \
		-i /local/$(SRC_DIR)/docs/swagger.json \
		-g python \
		-o /local/$(SDK_DIR) \
		-c /local/.openapi-generator/config.properties \
		--additional-properties=packageName=openapi,projectName=rcabench
	@printf "$(BLUE)📦 后处理生成的 SDK...$(RESET)\n"
	./scripts/mv-generated-sdk.sh
	@printf "$(GREEN)✅ Python SDK 生成完成$(RESET)\n"

# =============================================================================
# 发布管理
# =============================================================================

release: ## 🏷️  发布新版本 (用法: make release VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		printf "$(RED)❌ 请提供版本号: make release VERSION=1.0.1$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🏷️ 发布版本 $(VERSION)...$(RESET)\n"
	./scripts/release.sh $(VERSION)

release-dry-run: ## 🧪 发布流程试运行 (用法: make release-dry-run VERSION=1.0.1)
	@if [ -z "$(VERSION)" ]; then \
		printf "$(RED)❌ 请提供版本号: make release-dry-run VERSION=1.0.1$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)🧪 试运行发布流程 $(VERSION)...$(RESET)\n"
	./scripts/release.sh $(VERSION) --dry-run

upload: ## 📤 上传 SDK 包
	@printf "$(BLUE)📤 上传 SDK 包...$(RESET)\n"
	$(MAKE) -C sdk/python upload
	@printf "$(GREEN)✅ SDK 上传完成$(RESET)\n"

# =============================================================================
# 清理和维护
# =============================================================================

clean-all: ## 🧹 清理所有资源
	@printf "$(BLUE)🧹 清理所有资源...$(RESET)\n"
	@printf "$(YELLOW)⚠️  这将删除所有部署的资源！$(RESET)\n"
	@read -p "确认继续？(y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@printf "$(GRAY)删除 Helm 发布...$(RESET)\n"
	helm uninstall rcabench -n $(NS) || true
	@printf "$(GRAY)删除命名空间...$(RESET)\n"
	kubectl delete namespace $(NS) || true
	@printf "$(GRAY)停止端口转发...$(RESET)\n"
	pkill -f "kubectl port-forward" || true
	@printf "$(GREEN)✅ 清理完成$(RESET)\n"

# =============================================================================
# 实用工具
# =============================================================================

restart: ## 🔄 重启应用
	@printf "$(BLUE)🔄 重启应用...$(RESET)\n"
	kubectl rollout restart deployment --all -n $(NS)
	@printf "$(GREEN)✅ 应用重启完成$(RESET)\n"

scale: ## 📏 扩展部署 (用法: make scale DEPLOYMENT=app REPLICAS=3)
	@if [ -z "$(DEPLOYMENT)" ] || [ -z "$(REPLICAS)" ]; then \
		printf "$(RED)❌ 请提供部署名称和副本数: make scale DEPLOYMENT=app REPLICAS=3$(RESET)\n"; \
		exit 1; \
	fi
	@printf "$(BLUE)📏 扩展部署 $(DEPLOYMENT) 到 $(REPLICAS) 个副本...$(RESET)\n"
	kubectl scale deployment $(DEPLOYMENT) --replicas=$(REPLICAS) -n $(NS)
	@printf "$(GREEN)✅ 扩展完成$(RESET)\n"

# =============================================================================
# 信息显示
# =============================================================================

info: ## ℹ️  显示项目信息
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║                        RCABench 项目信息                     ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "$(YELLOW)配置信息:$(RESET)\n"
	@printf "  $(CYAN)默认仓库:$(RESET) $(DEFAULT_REPO)\n"
	@printf "  $(CYAN)命名空间:$(RESET) $(NS)\n"
	@printf "  $(CYAN)端口:$(RESET) $(PORT)\n"
	@printf "  $(CYAN)控制器目录:$(RESET) $(SRC_DIR)\n"
	@printf "  $(CYAN)SDK 目录:$(RESET) $(SDK_DIR)\n"
	@printf "\n"
	@printf "$(YELLOW)Chaos 类型:$(RESET)\n"
	@for type in $(CHAOS_TYPES); do \
		printf "  - $$type\n"; \
	done