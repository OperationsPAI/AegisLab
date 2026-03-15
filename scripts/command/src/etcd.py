"""
etcd.py - Business logic for etcd configuration management

This module contains the core business logic for managing etcd configurations,
including loading, initializing, listing, and clearing consumer configurations.
Uses etcd gRPC-JSON Gateway (HTTP API) for direct communication.
"""

import base64
from typing import Any

import requests
import yaml
from rich.table import Table

from src.common.common import ENV, INITIAL_DATA_PATH, console, settings

__all__ = [
    "init_etcd_configs",
    "list_etcd_configs",
    "clear_etcd_configs",
]


class EtcdHTTPClient:
    """HTTP client for etcd gRPC-JSON Gateway API."""

    def __init__(
        self,
        host: str,
        port: int,
        username: str | None = None,
        password: str | None = None,
    ):
        """Initialize etcd HTTP client.

        Args:
            host: etcd server host
            port: etcd server port
            username: Optional username for authentication
            password: Optional password for authentication
        """
        self.base_url = f"http://{host}:{port}/v3"
        self.auth = (username, password) if username and password else None
        self.session = requests.Session()
        if self.auth:
            self.session.auth = self.auth

    def _encode_key(self, key: str) -> str:
        """Encode key to base64 as required by etcd API."""
        return base64.b64encode(key.encode()).decode()

    def _decode_key(self, encoded: str) -> str:
        """Decode base64 encoded key."""
        return base64.b64decode(encoded).decode()

    def _encode_value(self, value: str) -> str:
        """Encode value to base64 as required by etcd API."""
        return base64.b64encode(value.encode()).decode()

    def _decode_value(self, encoded: str) -> str:
        """Decode base64 encoded value."""
        return base64.b64decode(encoded).decode()

    def get(self, key: str) -> tuple[str | None, dict | None]:
        """Get value for a key.

        Args:
            key: The key to retrieve

        Returns:
            Tuple of (value, metadata) or (None, None) if not found
        """
        url = f"{self.base_url}/kv/range"
        payload = {"key": self._encode_key(key)}

        response = self.session.post(url, json=payload)
        response.raise_for_status()

        data = response.json()
        kvs = data.get("kvs", [])

        if not kvs:
            return None, None

        kv = kvs[0]
        value = self._decode_value(kv["value"])
        return value, kv

    def put(self, key: str, value: str) -> None:
        """Put a key-value pair.

        Args:
            key: The key to set
            value: The value to set
        """
        url = f"{self.base_url}/kv/put"
        payload = {"key": self._encode_key(key), "value": self._encode_value(value)}

        response = self.session.post(url, json=payload)
        response.raise_for_status()

    def get_prefix(self, prefix: str) -> list[tuple[bytes, Any]]:
        """Get all key-value pairs with a prefix.

        Args:
            prefix: The key prefix to search for

        Returns:
            List of (value, metadata) tuples where metadata has a 'key' attribute
        """
        url = f"{self.base_url}/kv/range"
        # range_end is the key following the prefix
        range_end = prefix[:-1] + chr(ord(prefix[-1]) + 1) if prefix else "\0"
        payload = {
            "key": self._encode_key(prefix),
            "range_end": self._encode_key(range_end),
        }

        response = self.session.post(url, json=payload)
        response.raise_for_status()

        data = response.json()
        kvs = data.get("kvs", [])

        results = []
        for kv in kvs:
            value = self._decode_value(kv["value"])
            decoded_key = self._decode_key(kv["key"])

            # Create a simple metadata object with key attribute
            class Metadata:
                def __init__(self, key: bytes):
                    self.key = key

            metadata = Metadata(decoded_key.encode())
            results.append((value.encode(), metadata))

        return results

    def delete_prefix(self, prefix: str) -> int:
        """Delete all keys with a prefix.

        Args:
            prefix: The key prefix to delete

        Returns:
            Number of keys deleted
        """
        url = f"{self.base_url}/kv/deleterange"
        # range_end is the key following the prefix
        range_end = prefix[:-1] + chr(ord(prefix[-1]) + 1) if prefix else "\0"
        payload = {
            "key": self._encode_key(prefix),
            "range_end": self._encode_key(range_end),
        }

        response = self.session.post(url, json=payload)
        response.raise_for_status()

        data = response.json()
        return int(data.get("deleted", 0))


def _get_etcd_client(env: ENV) -> EtcdHTTPClient:
    """Create and return an etcd HTTP client using settings configuration.

    Returns:
        EtcdHTTPClient: Configured etcd HTTP client instance

    Raises:
        Exception: If failed to create etcd client
    """

    settings.setenv(env.value)

    return EtcdHTTPClient(
        host=settings.etcd.host,
        port=settings.etcd.port,
        username=settings.etcd.username or None,
        password=settings.etcd.password or None,
    )


def _load_dynamic_configs() -> dict[str, list[dict[str, Any]]]:
    """Load dynamic_configs from data.yaml.

    Returns:
        List of consumer configuration dictionaries (scope = 1)

    Raises:
        FileNotFoundError: If data.yaml not found
        yaml.YAMLError: If data.yaml is invalid
        Exception: For other errors during loading
    """
    if not INITIAL_DATA_PATH.exists():
        console.print(
            f"[bold red]❌ Data file not found: {INITIAL_DATA_PATH}[/bold red]"
        )
        raise FileNotFoundError(f"Data file not found: {INITIAL_DATA_PATH}")

    with open(INITIAL_DATA_PATH, encoding="utf-8") as f:
        data = yaml.safe_load(f)

    configs = data.get("dynamic_configs", [])

    table = Table(
        header_style="bold cyan",
        title=f"Loaded from {INITIAL_DATA_PATH.name}",
    )
    table.add_column("Scope", style="cyan")
    table.add_column("Count", style="bold green", justify="right")

    scopes = ["producer", "consumer", "global"]
    scope_configs: dict[str, list[dict[str, Any]]] = {}
    for idx, scope in enumerate(scopes):
        scope_configs[scope] = [c for c in configs if c.get("scope") == idx]
        count = len(scope_configs[scope])
        table.add_row(scope.capitalize(), str(count))

    console.print(table)
    return scope_configs


def init_etcd_configs(
    env: ENV, force: bool = False, dry_run: bool = False, values_file: str | None = None
) -> dict[str, int]:
    """Initialize etcd with consumer configurations from data.yaml.

    Args:
        force: Force overwrite existing values in etcd
        dry_run: Show what would be done without making changes
        values_file: Optional YAML file path with custom values (format: {"config_key": "value"})

    Returns:
        Dict: {"success": int, "skipped": int, "error": int}

    Raises:
        Exception: If failed to create etcd client or other errors
    """
    scope_configs = _load_dynamic_configs()

    if not scope_configs:
        console.print(
            "[bold yellow]⚠️  No scope configurations found to initialize[/bold yellow]"
        )
        return {"success": 0, "skipped": 0, "error": 0}

    # Load custom values from JSON file if provided
    custom_values = {}
    if values_file:
        try:
            from pathlib import Path

            values_path = Path(values_file)
            if not values_path.exists():
                console.print(
                    f"[bold red]❌ Values file not found: {values_file}[/bold red]"
                )
                raise FileNotFoundError(f"Values file not found: {values_file}")

            with open(values_path, encoding="utf-8") as f:
                custom_values = yaml.safe_load(f)

            console.print(
                f"[bold green]✅ Loaded {len(custom_values)} custom values from {values_path.name}[/bold green]"
            )
        except yaml.YAMLError as e:
            console.print(f"[bold red]❌ Invalid YAML in values file: {e}[/bold red]")
            raise

    # Create etcd client
    etcd_client = _get_etcd_client(env)

    if force:
        console.print(
            "[bold yellow]⚠️  Force mode enabled - will overwrite existing values[/bold yellow]"
        )
    if dry_run:
        console.print(
            "[bold yellow]⚠️  Dry-run mode - no actual changes will be made[/bold yellow]"
        )

    # Initialize etcd with configurations
    success_count = 0
    skipped_count = 0
    error_count = 0

    for scope, configs in scope_configs.items():
        if len(configs) == 0:
            continue

        etcd_prefix = f"{settings.etcd.prefix}/{scope}"
        console.print(
            f"[bold blue]\nInitializing {scope} configurations under prefix: {etcd_prefix}[/bold blue]"
        )

        for config in configs:
            key = config.get("key", "")
            default_value = config.get("default_value", "")
            description = config.get("description", "")
            is_secret = config.get("is_secret", False)

            if not key:
                console.print("[red]⚠️  Skipped config with empty key[/red]")
                error_count += 1
                continue

            # Use custom value if provided, otherwise use default
            value_to_set = custom_values.get(key, default_value)
            is_custom = key in custom_values

            etcd_key = f"{etcd_prefix}/{key}"

            # Check if key already exists in etcd
            existing_value = None
            try:
                result, _ = etcd_client.get(etcd_key)
                if result is not None:
                    existing_value = result
            except Exception as e:
                console.print(
                    f"[red]⚠️  Failed to check existing value for {key}: {e}[/red]"
                )
                error_count += 1
                continue

            # Skip if exists and not force mode
            if existing_value is not None and not force:
                display_value = "***" if is_secret else existing_value
                console.print(
                    f"[yellow]⊝[/yellow] [dim]{key}[/dim] (already exists: {display_value})"
                )
                skipped_count += 1
                continue

            # Set value in etcd
            if not dry_run:
                try:
                    etcd_client.put(etcd_key, value_to_set)
                    display_value = "***" if is_secret else value_to_set
                    action = "Updated" if existing_value else "Set"
                    source = (
                        f" [magenta]({'custom' if is_custom else 'default'})[/magenta]"
                    )

                    console.print(
                        f"[green]✓[/green] [cyan]{key}[/cyan] = [bold]{display_value}[/bold]{source}"
                    )

                    if description:
                        console.print(f"  [dim]└─ {description}[/dim]")

                    success_count += 1
                except Exception as e:
                    console.print(f"[red]✗[/red] [red]Failed to set {key}: {e}[/red]")
                    error_count += 1
            else:
                display_value = "***" if is_secret else value_to_set
                action = "Would update" if existing_value else "Would set"
                source = " [magenta](custom)[/magenta]" if is_custom else ""
                console.print(
                    f"[blue]•[/blue] [dim][DRY-RUN][/dim] {action} [cyan]{key}[/cyan] = {display_value}{source}"
                )
                success_count += 1

    # Summary
    console.print()
    console.print("=" * 60)

    table = Table(show_header=False, box=None, padding=(0, 2))
    table.add_column(style="bold")
    table.add_column(style="cyan")

    if dry_run:
        table.add_row("Would initialize:", f"{success_count} configurations")
    else:
        table.add_row("Initialized:", f"{success_count} configurations")
        table.add_row("Skipped:", f"{skipped_count} configurations")

    if error_count > 0:
        table.add_row("Failed:", f"{error_count} configurations")

    console.print(table)

    if not dry_run and success_count > 0:
        console.print(
            "\n[bold green]✅ etcd initialization completed successfully![/bold green]"
        )
    elif skipped_count > 0 and success_count == 0:
        console.print(
            "[bold yellow]⚠️  All configurations already exist (use --force to overwrite)[/bold yellow]"
        )

    return {"success": success_count, "skipped": skipped_count, "error": error_count}


def list_etcd_configs(env: ENV) -> int:
    """List all consumer configurations in etcd.

    Returns:
        Number of configurations found

    Raises:
        Exception: If failed to create etcd client or list configurations
    """
    # Create etcd client
    etcd_client = _get_etcd_client(env)

    etcd_prefix = settings.etcd.prefix

    configs = etcd_client.get_prefix(etcd_prefix)
    count = 0

    table = Table(show_header=True, header_style="bold cyan")
    table.add_column("Key", style="cyan", no_wrap=True)
    table.add_column("Value", style="white")

    for value, metadata in configs:
        key = metadata.key.decode("utf-8")
        # Remove prefix for display
        display_key = key[len(etcd_prefix) :]
        display_value = (
            value.decode("utf-8") if isinstance(value, bytes) else str(value)
        )

        table.add_row(display_key, display_value)
        count += 1

    if count > 0:
        console.print(table)
        console.print()
        console.print(f"[bold green]✅ Found {count} configurations[/bold green]")
    else:
        console.print("[bold yellow]⚠️  No configurations found in etcd[/bold yellow]")

    return count


def clear_etcd_configs(env: ENV) -> int:
    """Clear all consumer configurations from etcd.

    Returns:
        Number of configurations deleted

    Raises:
        Exception: If failed to create etcd client or clear configurations
    """
    # Create etcd client
    etcd_client = _get_etcd_client(env)

    etcd_prefix = settings.etcd.prefix

    deleted = etcd_client.delete_prefix(etcd_prefix)
    console.print(
        f"[bold green]✅ Deleted {deleted} configurations from etcd[/bold green]"
    )

    return deleted
