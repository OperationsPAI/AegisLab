import dataclasses
import json
import re
import subprocess
from typing import Any

from jinja2 import Template

from src.common.command import run_command
from src.common.common import ENV, INITIAL_DATA_PATH, console
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager
from src.util import parse_image_address

CONTEXT_MAPPING = {"dev": "kubernetes-admin@kubernetes", "test": "k3d-test-cluter"}


__all__ = ["Pedestal", "install_pedestals"]


@dataclasses.dataclass(frozen=True)
class HelmValue:
    key: str
    type: int  # 0: Fixed (use default_value), 1: Dynamic (use template_string)
    category: int
    default_value: str | None = None
    template_string: str | None = None
    required: bool = False


@dataclasses.dataclass(frozen=True)
class Pedestal:
    image_parts: dict[str, str | None]
    repo_name: str
    chart_name: str
    ns_prefix: str
    helm_values: list[HelmValue]


def _load_pedestals(name: str) -> Pedestal | None:
    with open(INITIAL_DATA_PATH) as f:
        data = json.load(f)

    for container in data["containers"]:
        if container["type"] != 2 or container["name"] != name:
            continue

        for version in container["versions"]:
            if "helm_config" in version:
                # Parse helm values from JSON
                helm_values = []
                for value_data in version["helm_config"].get("values", []):
                    helm_values.append(
                        HelmValue(
                            key=value_data["key"],
                            type=value_data["type"],
                            category=value_data["category"],
                            default_value=value_data.get("default_value"),
                            template_string=value_data.get("template_string"),
                            required=value_data.get("required", False),
                        )
                    )

                return Pedestal(
                    image_parts=parse_image_address(version["image_ref"]),
                    repo_name=version["helm_config"]["repo_name"],
                    chart_name=version["helm_config"]["chart_name"],
                    ns_prefix=version["helm_config"]["ns_prefix"],
                    helm_values=helm_values,
                )

    return None


def _get_pedestal_or_exit(name: str) -> Pedestal:
    """Retrieve a Pedestal by name or exit with an error message if not found."""
    pedestals = _load_pedestals(name)
    if not pedestals:
        raise Exception(f"Invalid pedestal container name '{name}'")
    return pedestals


def _convert_helm_values_to_set_list(
    values_dict: dict[str, Any],
    prefix: str = "",
    key_value_pairs: list[str] | None = None,
) -> list[str]:
    """
    Recursively converts a nested Helm values dictionary into an alternating
    list of ['--set', 'key=value', ...] suitable for subprocess calls.

    This function performs two main steps:
    1. Flattens the nested dictionary into a list of dot-separated 'key=value' strings.
    2. Transforms that list into the alternating '--set' structure required by Helm commands.

    Args:
        values_dict: The nested dictionary containing the Helm values.
                     (e.g., the content of the "values" field).
        prefix: Internal argument used during recursion to accumulate the dot-separated path.
        key_value_pairs: Internal argument used to accumulate the flat 'key=value' strings.

    Returns:
        A list of strings formatted as ['--set', 'key=value', ...].
    """
    if key_value_pairs is None:
        key_value_pairs = []

    for k, v in values_dict.items():
        current_key = f"{prefix}.{k}" if prefix else k

        if isinstance(v, dict):
            _convert_helm_values_to_set_list(v, current_key, key_value_pairs)
        else:
            if isinstance(v, bool):
                value_str = str(v).lower()
            elif v is None:
                value_str = ""
            else:
                value_str = str(v)

            key_value_pairs.append(f"{current_key}={value_str}")

    if prefix == "":
        alternating_list: list[str] = []
        for pair in key_value_pairs:
            alternating_list.append("--set")
            alternating_list.append(pair)
        return alternating_list

    return key_value_pairs


def _set_nested_dict_value(d: dict[str, Any], key: str, value: Any) -> None:
    """Set a value in a nested dictionary using dot notation.

    Args:
        d: The dictionary to modify
        key: Dot-separated key path (e.g., 'global.image.repository')
        value: The value to set
    """
    keys = key.split(".")
    current = d

    for k in keys[:-1]:
        if k not in current:
            current[k] = {}
        current = current[k]

    current[keys[-1]] = value


def _process_single_helm_value(
    helm_value: HelmValue,
    context: dict[str, Any],
    index: int,
) -> tuple[str, Any] | None:
    """Process a single helm value configuration and return (key, value) or None.

    This function mirrors the Go processParameterConfig logic.

    Args:
        helm_value: HelmValue configuration object
        context: Context dictionary containing values like Registry, Namespace, Tag
        index: Index for dynamic template rendering (e.g., for port numbers)

    Returns:
        Tuple of (key, rendered_value) if successful, None if value should be skipped

    Raises:
        ValueError: If required value is missing or invalid
    """
    # Type 0: Fixed - use default_value
    if helm_value.type == 0:
        final_value = helm_value.default_value

        # Check if required and missing
        if final_value is None:
            if helm_value.required:
                raise ValueError(
                    f"Required fixed parameter '{helm_value.key}' is missing a value and has no default"
                )
            # Optional parameter with no value - skip it
            return None

        # Convert string value to appropriate type (bool, int, etc.)
        try:
            # Try to parse as JSON to get proper types
            if final_value.lower() in ("true", "false"):
                final_value = final_value.lower() == "true"
            elif final_value.isdigit():
                final_value = int(final_value)
        except (AttributeError, ValueError):
            # Keep as string if conversion fails
            pass

        return (helm_value.key, final_value)

    # Type 1: Dynamic - use template_string
    elif helm_value.type == 1:
        if helm_value.template_string is None or helm_value.template_string == "":
            raise ValueError(
                f"Dynamic parameter '{helm_value.key}' is missing a template string"
            )

        template_str = helm_value.template_string

        # Extract template variables (e.g., {{ .Registry }} -> ["Registry"])
        # This mirrors Go's ExtractTemplateVars
        var_pattern = re.compile(r"\{\{\s*\.(\w+)\s*\}\}")
        template_vars = var_pattern.findall(template_str)

        # If no variables, return the template string as-is
        if not template_vars:
            # Check if it's a Python format string (e.g., "31%03d")
            if "%" in template_str:
                try:
                    rendered_value = template_str % index
                except (ValueError, TypeError) as e:
                    raise ValueError(
                        f"Failed to format template for '{helm_value.key}': {e}"
                    )
            else:
                rendered_value = template_str

            if helm_value.required and not rendered_value:
                raise ValueError(
                    f"Required dynamic parameter '{helm_value.key}' has no value"
                )

            return (helm_value.key, rendered_value) if rendered_value else None

        # Render template with context
        # Convert Go-style {{ .Registry }} to Jinja2 {{ Registry }}
        jinja_template = re.sub(r"\{\{\s*\.(\w+)\s*\}\}", r"{{ \1 }}", template_str)

        # Add index to context
        render_context = {**context, "Index": index}

        try:
            template = Template(jinja_template)
            rendered_value = template.render(render_context)
        except Exception as e:
            raise ValueError(
                f"Failed to render dynamic parameter '{helm_value.key}': {e}"
            )

        # Validate required parameters
        if helm_value.required and not rendered_value:
            raise ValueError(
                f"Required dynamic parameter '{helm_value.key}' rendered to an empty string"
            )

        # Skip if rendered to empty (for optional parameters)
        if not rendered_value:
            return None

        return (helm_value.key, rendered_value)

    else:
        raise ValueError(
            f"Unknown parameter type '{helm_value.type}' for key '{helm_value.key}'"
        )


def _render_helm_values(
    helm_values: list[HelmValue], context: dict[str, Any], index: int
) -> dict[str, Any]:
    """Render helm values based on their type.

    Args:
        helm_values: List of HelmValue objects to render
        context: Context dictionary containing values like Registry, Namespace, Tag
        index: Index for dynamic template rendering (e.g., for port numbers)

    Returns:
        A nested dictionary of rendered helm values

    Raises:
        ValueError: If any required parameter fails validation
    """
    result: dict[str, Any] = {}

    for helm_value in helm_values:
        try:
            processed = _process_single_helm_value(helm_value, context, index)

            if processed is None:
                # Optional parameter with no value - skip it
                continue

            key, value = processed

            # Set the value in the nested dictionary using dot notation
            _set_nested_dict_value(result, key, value)

        except ValueError as e:
            # Log error and re-raise for required parameters
            console.print(f"[red]Error processing parameter: {e}[/red]")
            raise

    return result


@with_k8s_manager
def install_pedestals(
    env: ENV, name: str, count: int, k8s_manager: KubernetesManager
) -> None:
    if count <= 0:
        console.print("[bold red]PEDESTAL_COUNT must be a positive number[/bold red]")
        raise SystemExit(1)

    pedestal = _get_pedestal_or_exit(name)
    ns_prefix = pedestal.ns_prefix
    console.print(
        f"[bold blue]Checking Helm releases in namespaces {ns_prefix}0 to {ns_prefix}{count - 1}...[/bold blue]"
    )

    for i in range(count):
        ns = f"{ns_prefix}{i}"
        console.print(f"[bold blue]Checking namespace: {ns}[/bold blue]")

        ns_ok = k8s_manager.check_and_create_namespace(ns)
        if not ns_ok:
            console.print(f"[bold yellow]Namespace {ns} does not exist[/bold yellow]")
            continue

        has_release = (
            run_command(
                ["helm", "list", "-n", ns, "-q"],
                check=False,
                capture_output=True,
                stderr=subprocess.DEVNULL,
            ).stdout
            or ""
        )
        if ns in has_release:
            console.print(
                f"[bold yellow]Helm release '{ns}' found in namespace {ns}[/bold yellow]"
            )
            continue

        console.print(
            f"[bold green]Helm release '{ns}' not found in namespace {ns}[/bold green]"
        )
        _install_release(pedestal, namespace=ns, index=i)

    console.print("[bold green]🎉 Check and installation completed![/bold green]")


def _install_release(pedestal: Pedestal, namespace: str, index: int) -> None:
    """Install a Helm release with rendered values.

    Args:
        ctx: Kubernetes context
        pedestal: Pedestal configuration
        ns: Namespace to install into
        index: Index for dynamic value rendering (e.g., for port numbers)
    """
    chart_ref = f"{pedestal.repo_name}/{pedestal.chart_name}"

    base_cmd: tuple[str, ...] = (
        "helm",
        "install",
        namespace,
        chart_ref,
        "--create-namespace",
        "-n",
        namespace,
    )

    values: list[str] = []
    if pedestal.helm_values:
        # Prepare context for rendering
        render_context = {
            "Registry": pedestal.image_parts.get("registry", ""),
            "Namespace": pedestal.image_parts.get("namespace", ""),
            "Repository": pedestal.image_parts.get("repository", ""),
            "Tag": pedestal.image_parts.get("tag", ""),
        }

        # Render helm values
        rendered_values = _render_helm_values(
            pedestal.helm_values, render_context, index
        )

        # Convert to --set format
        values.extend(_convert_helm_values_to_set_list(rendered_values))

    run_command([*base_cmd, *values])
