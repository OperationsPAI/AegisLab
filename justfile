# =============================================================================
# RCABench Justfile
# =============================================================================
# Use 'just --list' to view all available commands

set dotenv-load := true
set dotenv-filename := ".env"

# Configuration
env_mode     := env("ENV_MODE", "dev")
ns           := "exp"

root := justfile_directory()
command_dir  := root + "/scripts/command"
src_dir      := root + "/src"

command      := "UV_WITH_GROUPS=dev uv run --project " + command_dir + " python " + command_dir + "/main.py"

# Colors
blue   := '\033[1;34m'
green  := '\033[1;32m'
yellow := '\033[1;33m'
red    := '\033[1;31m'
cyan   := '\033[1;36m'
gray   := '\033[90m'
reset  := '\033[0m'

# =============================================================================
# Default Recipe
# =============================================================================

# 📖 Display all available commands
default:
    @just --list

# =============================================================================
# Command Tool Management
# =============================================================================

# 🔧 Run command tool (usage: just run-command "your args")
run-command *ARGS:
    {{command}} {{ARGS}}

# =============================================================================
# Environment Check and Setup
# =============================================================================

# 🔍 Check development environment dependencies
check-prerequisites:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}🔍 Checking development environment dependencies...{{reset}}\n"
    for cmd in devbox docker helm kubectx; do
        if command -v "$cmd" &>/dev/null; then
            printf "{{gray}}  - $cmd installed{{reset}}\n"
        else
            printf "{{red}}❌ $cmd not installed{{reset}}\n" && exit 1
        fi
    done
    devbox install >/dev/null 2>&1 || { printf "{{red}}❌ devbox dependencies installation failed{{reset}}\n"; exit 1; }
    printf "{{green}}✅ All dependency checks passed{{reset}}\n\n"

# 🔗 Start port forwarding to access application
forward-ports env="prod":
    just run-command port start -e {{env}} -n {{ns}}

# 🛠️ Setup development environment
setup-dev-env: check-prerequisites
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}🛠️  Setting up development environment...{{reset}}\n"
    if command -v uv &>/dev/null; then
        printf "{{green}}✅ 'uv' found in PATH{{reset}}\n"
    else
        printf "{{yellow}}Warning: 'uv' not found. Installing via script...{{reset}}\n"
        curl -LsSf https://astral.sh/uv/install.sh | sh
        printf "{{green}}✅ 'uv' installed!{{reset}}\n"
    fi
    printf "{{gray}}Applying Kubernetes manifests for local development...{{reset}}\n"
    kubectl apply -f manifests/local-dev/exp-dev-setup.yaml
    printf "{{gray}}Installing Git hooks with Lefthook...{{reset}}\n"
    if [ -f "lefthook.yml" ]; then
        devbox run install-hooks
        printf "{{green}}✅ Lefthook hooks installed successfully!{{reset}}\n"
    else
        printf "{{red}}❌ lefthook.yml not found{{reset}}\n" && exit 1
    fi
    printf "{{green}}✅ Development environment setup completed!{{reset}}\n"

# 🧪 Setup test environment
setup-test-env: check-prerequisites
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}🧪 Setting up test environment...{{reset}}\n"
    bash {{root}}/manifests/test/start.sh
    printf "{{green}}✅ Test environment setup completed!{{reset}}\n"

# =============================================================================
# Pedestal Function
# =============================================================================

# 🔍 Install pedestals in namespaces (usage: just install-pedestals <name> <count>)
install-pedestals pedestal_name pedestal_count:
    just run-command pedestal install -e {{env_mode}} -n {{pedestal_name}} -c {{pedestal_count}} -f

# =============================================================================
# Build and Deployment
# =============================================================================

# Deploy OpenEBS
install-openebs:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}Deploying OpenEBS...{{reset}}\n"
    helm upgrade -i openebs openebs/openebs --namespace openebs \
        --create-namespace \
        --values ./manifests/staging/openebs.yaml \
        --atomic --timeout 10m
    printf "{{green}}✅ OpenEBS installed successfully{{reset}}\n\n"

# 🔧 Deploy RCABench application in prod environment
install-rcabench:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}🔧 Deploying RCABench application...{{reset}}\n"
    helm upgrade -i rcabench ./helm --namespace {{ns}} \
        --create-namespace \
        --values ./manifests/prod/rcabench.yaml \
        --set-file initialDataFiles.data_yaml=data/initial_data/prod/data.yaml \
        --set-file initialDataFiles.otel_demo_yaml=data/initial_data/prod/otel-demo.yaml \
        --set-file initialDataFiles.ts_yaml=data/initial_data/prod/ts.yaml \
        --atomic --timeout 10m
    printf "{{green}}✅ RCABench installed successfully{{reset}}\n\n"
    printf "{{blue}}🔗 Starting automatic port forwarding...{{reset}}\n"
    just forward-ports

# 🛠️ Setup local development environment with basic services
local-deploy:
    just run-command rcabench local-deploy -f
    just init-etcd

# 🚀 Build and deploy application (using skaffold)
run: check-prerequisites
    ENV_MODE=staging devbox run skaffold run

# Initialize etcd
init-etcd:
    just run-command etcd init -e {{env_mode}} -f

update-version version:
    just run-command rcabench update-version -v {{version}}

# =============================================================================
# Backup
# =============================================================================

# 💾 Backup MySQL database
backup-mysql:
    just run-command backup mysql backup

# =============================================================================
# Test
# =============================================================================

# Run tests
test version:
    SDK_VERSION={{version}} ENV_MODE=test devbox run skaffold run

# Run regression tests
regression-test:
    chmod +x ./scripts/regression-test.sh && ./scripts/regression-test.sh

# =============================================================================
# Development Tools
# =============================================================================

# 🐛 Start local debugging environment
local-debug:
    #!/usr/bin/env bash
    printf "{{blue}}⌛️ Starting local application...{{reset}}\n"
    cd {{src_dir}} && go run -tags=duckdb_arrow main.go both --port 8082

# 📦 Update latest version of dependencies
update-dependencies:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}📦 Updating latest version of chaos-experiment library...{{reset}}\n"
    cd {{src_dir}} && go get -u github.com/OperationsPAI/chaos-experiment@injectionv2-dev && go mod tidy
    printf "{{green}}✅ Dependencies update completed{{reset}}\n"

# =============================================================================
# Chaos Management
# =============================================================================

# 🧹 Clean all chaos resource finalizers (usage: just clean-finalizers <prefix> <count>)
clean-finalizers ns_prefix ns_count:
    just run-command chaos clean-finalizers -e {{env_mode}} -p {{ns_prefix}} -c {{ns_count}}

# 🗑️ Delete chaos resources (usage: just delete-chaos <prefix> <count>)
delete-chaos ns_prefix ns_count:
    just run-command chaos delete-resources -e {{env_mode}} -p {{ns_prefix}} -c {{ns_count}}

# =============================================================================
# Image Sync
# =============================================================================

# 🔄 Sync Docker images from DockerHub to prod repository
sync-images bv="latest" fv="latest":
    #!/usr/bin/env bash
    set -euo pipefail
    source {{root}}/.secret

    # Application images — use provided version tag
    app_images=(
        "opspai/rcabench:{{bv}}"
        "opspai/rcabench-frontend:{{fv}}"
    )

    sync_image() {
        local img="$1"
        local name="${img%%:*}"
        local tag="${img##*:}"
        local base="${name##*/}"
        local src="docker://docker.io/${img}"
        local dest="docker://${PROD_REPO}/${base}:${tag}"

        printf "{{gray}}   📤 ${src} → ${dest}{{reset}}\n"
        skopeo copy \
            --src-creds  "${DOCKERHUB_USERNAME}:${DOCKERHUB_PASSWORD}" \
            --dest-creds "${VOLCE_USERNAME}:${VOLCE_PASSWORD}" \
            --override-arch amd64 \
            "$src" "$dest"
    }

    printf "{{blue}}🔄 Syncing application images (backend_version:{{bv}}, frontend_version:{{fv}})...{{reset}}\n"
    for img in "${app_images[@]}"; do sync_image "$img"; done

    printf "{{green}}✅ Image sync completed successfully{{reset}}\n"

# =============================================================================
# SDK Generation
# =============================================================================

# 📝 Initialize Swagger documentation
swag-init version:
    just run-command swagger init -v {{version}}

# ⚙️ Generate TypeScript Client from Swagger documentation
generate-typescript-client version:
    just swag-init {{version}}
    just run-command swagger generate-client -l typescript -v {{version}}

# ⚙️ Generate Python SDK from Swagger documentation
generate-python-sdk version:
    just swag-init {{version}}
    just run-command swagger generate-sdk -l python -v {{version}}

# =============================================================================
# Utilities
# =============================================================================

# 📝 Generate CHANGELOG.md
changelog:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}📝 Generating CHANGELOG.md...{{reset}}\n"
    eval "$(devbox shellenv)" && git-cliff -o CHANGELOG.md
    printf "{{green}}✅ CHANGELOG.md generated successfully{{reset}}\n"

# 👁️ Preview unreleased changes
changelog-preview:
    #!/usr/bin/env bash
    eval "$(devbox shellenv)" && git-cliff --unreleased

# 📋 Show latest release changes
changelog-latest:
    #!/usr/bin/env bash
    eval "$(devbox shellenv)" && git-cliff --latest

release version:
    #!/usr/bin/env bash
    set -euo pipefail
    printf "{{blue}}🚀 Releasing version {{version}}...{{reset}}\n"
    just update-version {{version}}
    just changelog
    git add {{root}}/CHANGELOG.md {{root}}/helm/Chart.yaml {{root}}/helm/values.yaml \
        {{root}}/src/config.dev.toml {{root}}/src/main.go
    git commit -m "chore(release): version {{version}}" --no-verify
    git push -u origin main
    git tag -a "v{{version}}" -m "Release version {{version}}"
    git push origin "v{{version}}"
    printf "{{green}}✅ Version {{version}} released successfully{{reset}}\n"