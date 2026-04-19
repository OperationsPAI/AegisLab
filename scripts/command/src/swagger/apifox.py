from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from enum import Enum
from pathlib import Path

import typer
from rich.panel import Panel

from src.common.common import console, settings

__all__ = [
    "ApifoxTarget",
    "upload_targets_to_apifox",
]


class ApifoxTarget(str, Enum):
    SDK = "sdk"
    PORTAL = "portal"
    ADMIN = "admin"
    ALL = "all"


TARGET_OPENAPI_FILES: dict[ApifoxTarget, str] = {
    ApifoxTarget.SDK: "sdk.json",
    ApifoxTarget.PORTAL: "portal.json",
    ApifoxTarget.ADMIN: "admin.json",
}


def upload_targets_to_apifox(
    converted_dir: Path,
    targets: list[ApifoxTarget] | None = None,
) -> None:
    """Upload one or more generated OpenAPI documents to Apifox."""
    normalized_targets = _normalize_targets(targets or [ApifoxTarget.ALL])
    _ensure_common_config()

    for target in normalized_targets:
        openapi_path = converted_dir / TARGET_OPENAPI_FILES[target]
        endpoint_folder_id = _required_env(
            f"APIFOX_{target.upper()}_ENDPOINT_FOLDER_ID"
        )
        schema_folder_id = _required_env(f"APIFOX_{target.upper()}_SCHEMA_FOLDER_ID")
        _upload_openapi(
            openapi_path=openapi_path,
            label=target.value,
            endpoint_folder_id=int(endpoint_folder_id),
            schema_folder_id=int(schema_folder_id),
        )


def _normalize_targets(targets: list[ApifoxTarget]) -> list[ApifoxTarget]:
    """Expand and de-duplicate Apifox upload targets."""
    normalized: list[ApifoxTarget] = []
    for target in targets:
        if target == ApifoxTarget.ALL:
            normalized.extend(
                [
                    ApifoxTarget.SDK,
                    ApifoxTarget.PORTAL,
                    ApifoxTarget.ADMIN,
                ]
            )
            continue
        normalized.append(target)

    deduped: list[ApifoxTarget] = []
    seen: set[ApifoxTarget] = set()
    for target in normalized:
        if target in seen:
            continue
        seen.add(target)
        deduped.append(target)
    return deduped


def _ensure_common_config() -> None:
    """Ensure project-level Apifox credentials exist before uploading."""
    missing = [
        name
        for name in ("APIFOX_PROJECT_ID", "APIFOX_ACCESS_TOKEN")
        if not os.getenv(name)
    ]
    if missing:
        console.print(
            "[bold red]Missing required Apifox config:[/bold red] " + ", ".join(missing)
        )
        raise typer.Exit(2)


def _required_env(name: str) -> str:
    """Return a required env var or exit with a clear message."""
    value = os.getenv(name)
    if value:
        return value
    console.print(f"[bold red]Missing required Apifox config:[/bold red] {name}")
    raise typer.Exit(2)


def _upload_openapi(
    *,
    openapi_path: Path,
    label: str,
    endpoint_folder_id: int,
    schema_folder_id: int,
) -> None:
    """Upload one OpenAPI document to Apifox."""
    if not openapi_path.exists():
        console.print(f"[bold red]OpenAPI file not found:[/bold red] {openapi_path}")
        raise typer.Exit(2)

    apifox_settings = settings.apifox
    project_id = _required_env("APIFOX_PROJECT_ID")
    access_token = _required_env("APIFOX_ACCESS_TOKEN")
    payload = {
        "input": openapi_path.read_text(encoding="utf-8"),
        "options": {
            "targetEndpointFolderId": endpoint_folder_id,
            "targetSchemaFolderId": schema_folder_id,
            "endpointOverwriteBehavior": "OVERWRITE_EXISTING",
            "schemaOverwriteBehavior": "OVERWRITE_EXISTING",
            "updateFolderOfChangedEndpoint": True,
            "prependBasePath": True,
        },
    }

    request = urllib.request.Request(
        (
            f"{str(apifox_settings.api_base_url).rstrip('/')}/projects/{project_id}/import-openapi"
            f"?locale={apifox_settings.locale}"
        ),
        data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        method="POST",
        headers={
            "X-Apifox-Api-Version": apifox_settings.api_version,
            "Authorization": f"Bearer {access_token}",
            "Content-Type": "application/json",
        },
    )

    console.print(
        Panel(
            f"file: {openapi_path}\n"
            f"endpoint folder: {endpoint_folder_id}\n"
            f"schema folder: {schema_folder_id}",
            title=f"Uploading {label} OpenAPI to Apifox",
        )
    )
    try:
        with urllib.request.urlopen(request) as response:
            body = response.read().decode("utf-8")
            console.print(
                f"[green]Apifox upload succeeded for {label} (HTTP {response.status}).[/green]"
            )
            _print_response(body)
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        console.print(
            f"[bold red]Apifox upload failed for {label} (HTTP {exc.code}).[/bold red]"
        )
        _print_response(body)
        raise typer.Exit(1) from exc


def _print_response(response_body: str) -> None:
    """Pretty-print Apifox responses when possible."""
    try:
        console.print_json(response_body)
    except Exception:
        console.print(response_body)
