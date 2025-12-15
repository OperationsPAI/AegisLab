# Color definitions
BLUE    := \033[1;34m
GREEN   := \033[1;32m
YELLOW  := \033[1;33m
RED     := \033[1;31m
CYAN    := \033[1;36m
GRAY    := \033[90m
RESET   := \033[0m

.PHONY: help install-openebs install-redis install-juicefs install-clickhouse install-rcabench install-trainticket install-all clean-all

help:  ## 📖 Display all available commands
	@printf "$(BLUE)╔══════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║               ByteDance Helm Install                         ║$(RESET)\n"
	@printf "$(BLUE)╚══════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"
	@printf "$(YELLOW)Usage:$(RESET) make $(CYAN)<target>$(RESET)\n"
	@printf "$(YELLOW)Examples:$(RESET)\n make help, make install-all, make clean-all \n"
	@printf "\n"
	@awk 'BEGIN { \
		FS = ":.*##"; \
		printf "$(YELLOW)Available commands:$(RESET)\n"; \
	} \
	/^[a-zA-Z_-]+:.*?##/ { \
		printf "  $(CYAN)%-25s$(RESET) $(GRAY)%s$(RESET)\n", $$1, $$2; \
	}' $(MAKEFILE_LIST)

install-clickhouse:  ## 📊 Deploy ClickHouse monitoring stack
	@echo "Deploying ClickHouse monitoring stack..."
	helm install clickstack clickstack/clickstack --namespace monitoring \
		--create-namespace \
		--values ./manifests/click-stack.yaml

install-otel-kube-stack:  ## 📈 Deploy OpenTelemetry Kube Stack for monitoring
	@echo "Deploying OpenTelemetry Kube Stack..."
	helm install opentelemetry-kube-stack open-telemetry/opentelemetry-kube-stack --namespace monitoring \
		--create-namespace \
		--values ./manifests/otel-kube-stack.yaml

install-chaos-mesh:
	@echo "Deploying Chaos Mesh..."
	helm install chaos-mesh chaos-mesh/chaos-mesh --namespace chaos-mesh \
		--create-namespace \
		--values ./manifests/chaos-mesh.yaml \
		--version 2.8.0

install-rcabench:  ## 🔧 Deploy RCABench application
	@echo "Deploying RCABench application..."
	helm install rcabench ../helm --namespace exp \
		--create-namespace \
		--set persistence.storageType=volcengine

install-trainticket:  ## 🚄 Deploy Train-Ticket workload
	@echo "Deploying Train-Ticket workload..."
	helm install ts0 train-ticket/trainticket --namespace ts0 \
		--create-namespace \
		--atomic --debug \
		--wait --timeout 10m0s \
		--values ./manifests/train-ticket.yaml

clean-all:  ## 🧹 Uninstall all deployed components
	@echo "Uninstalling all components..."
	-helm uninstall ts0 --namespace ts0
	-helm uninstall rcabench --namespace exp
	-helm uninstall opentelemetry-kube-stack --namespace monitoring
	-helm uninstall clickstack --namespace monitoring
	@echo "All components uninstalled!"
