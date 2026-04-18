# RCABench Python SDK

The SDK exposes two handwritten entry clients on top of the generated OpenAPI package:

- `RCABenchClient`: public/business API client authenticated by `Key ID` + `Key Secret`
- `RCABenchRuntimeClient`: runtime-only client authenticated by service token

Generated OpenAPI code lives under `src/rcabench/openapi`. Handwritten auth/session logic lives under `src/rcabench/client`.

## Installation

```bash
pip install rcabench
```

For local development:

```bash
cd sdk/python
pip install -e .
```

## Authentication Model

Secrets are never passed directly in code. The SDK reads credentials from environment variables only.

### Public Client

Required environment variables:

```bash
export RCABENCH_BASE_URL="http://localhost:8082"
export RCABENCH_KEY_ID="pk_xxx"
export RCABENCH_KEY_SECRET="sk_xxx"
```

`RCABenchClient` exchanges the key pair for a bearer token through the API-key token endpoint, then reuses the authenticated OpenAPI client.

### Runtime Client

Required environment variables:

```bash
export RCABENCH_BASE_URL="http://localhost:8082"
export RCABENCH_SERVICE_TOKEN="runtime_token_xxx"
```

`RCABenchRuntimeClient` is intended for managed runtime/wrapper usage. It injects the service token into the generated OpenAPI client directly.

## Usage

### Public API Client

```python
from rcabench import RCABenchClient
from rcabench.openapi.api.datasets_api import DatasetsApi

client = RCABenchClient()
api = DatasetsApi(client.get_client())

datasets = api.list_sdk_dataset_samples(page=1, size=10)
print(datasets)
```

You may still override `base_url` in code when needed:

```python
client = RCABenchClient(base_url="http://localhost:8082")
```

### Runtime API Client

```python
from rcabench import RCABenchRuntimeClient

runtime_client = RCABenchRuntimeClient()
api_client = runtime_client.get_client()

print(api_client.configuration.host)
```

Runtime-tagged API classes are generated from the current OpenAPI audience split. Use `runtime_client.get_client()` with the generated API module that corresponds to those runtime-only routes.

## Development

Run type checking only on handwritten SDK code:

```bash
cd sdk/python
uv run --with pyright pyright src/rcabench/client
```

The generated package under `src/rcabench/openapi` is excluded from Pyright.
