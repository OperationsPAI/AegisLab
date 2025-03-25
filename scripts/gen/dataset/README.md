# Gen Dataset Script User Guide

## Project Structure
```
.
├── config/                 # Experiment static configuration
│   ├── dev.json            # Development environment configuration
│   └── prod.json           # Production environment configuration
├── docker-compose.yaml     # Multi-service orchestration
├── Dockerfile              # Base image build file
├── envs/                   # Microservice environment variables
│   └── ts.env              # Template for train-ticket system experiments
├── main.py                 # Main experiment platform program
├── output/                 # Experiment results
│   └── [timestamp]/        # Results for each experiment
├── README.md               # This document
```

## Adding New Microservice Experiments

### Step 1: Create Environment File
Create `<microservice-name>.env` under `envs/` directory:

```bash
# Example for payment system
cp envs/ts.env envs/payment.env
```

### Step2: Edit Configuration

```ini
# envs/payment.env
COMMAND=cd /app/payment && make deploy  # Build command
NAMESPACE=payment                        # Kubernetes namespace
SERVICES=payment-service,refund-service  # Target microservices
```

### Step 3: Update Docker Compose
Add new service block in `docker-compose.yaml`:

```yaml
services:
  x-experiment-template-ts: &exp-template-ts
    <<: *common-base
    volumes:
        - ~/.kube/config:/root/.kube/config
        - ~/workspace/payment:/app/payment
        - ./config:/app/config
        - ./output:/app/output
    env_file: 
        - ./envs/payment.env
  gen-dataset-payment-dev:
    <<: *common-base           # Inherit base configuration
    container_name: gen-dataset-payment-dev
    env_file:
      - ./envs/payment.env      # Load payment config
    environment:
      - ENV_MODE=dev         # Override environment mode
```

## Starting Experiments

```bash
# Start single service
docker-compose up gen-dataset-payment-dev -d

# Start all services
docker-compose up
```

## Important Notes
1. ​File Format Requirements:
    - Use KEY=VALUE format for environment files
    - Avoid spaces in values (use quotes if needed)
2. ​Path Conventions:
    - All paths are relative to docker-compose.yaml
    - Ensure mounted directories exist
3. ​Variable Conflicts:
    - Priority: environment > env_file > Dockerfile ENV