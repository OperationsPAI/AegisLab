import re
from pathlib import Path
from typing import Any

import yaml
from jinja2 import Template
from pydantic import BaseModel
from sqlalchemy import text

from src.backup.mysql import MysqlClient, mysql_configs
from src.common.common import ENV, console
from src.common.helm_cli import HelmCLI, HelmRelease
from src.common.kubernetes_manager import KubernetesManager, with_k8s_manager

__all__ = ["Pedestal", "install_pedestals"]


def _extract_prefix_from_pattern(pattern: str) -> str:
    """Extract the namespace prefix from a regex pattern.

    Args:
        pattern: Regex pattern like "^ts\\d+$" or "ts"

    Returns:
        The extracted prefix (e.g., "ts")

    Examples:
        >>> _extract_prefix_from_pattern("^ts\\d+$")
        "ts"
        >>> _extract_prefix_from_pattern("^mysql\\d+$")
        "mysql"
        >>> _extract_prefix_from_pattern("ts")
        "ts"
    """
    import re

    # Remove common regex anchors and patterns
    # Pattern like "^ts\d+$" -> "ts"
    cleaned = pattern.strip()
    cleaned = re.sub(r"^\^", "", cleaned)  # Remove leading ^
    cleaned = re.sub(r"\$$", "", cleaned)  # Remove trailing $
    cleaned = re.sub(r"\\d\+", "", cleaned)  # Remove \d+
    cleaned = re.sub(r"\\d\*", "", cleaned)  # Remove \d*
    cleaned = re.sub(r"\[\d\-\]\+", "", cleaned)  # Remove [0-9]+
    cleaned = re.sub(r"\[\d\-\]\*", "", cleaned)  # Remove [0-9]*
    cleaned = re.sub(r"\[0-9\]\+", "", cleaned)  # Remove [0-9]+
    cleaned = re.sub(r"\[0-9\]\*", "", cleaned)  # Remove [0-9]*

    return cleaned


class HelmValue(BaseModel, frozen=True):
    """Represents a Helm value configuration."""

    key: str
    type: int  # 0: Fixed (use default_value), 1: Dynamic (use template_string)
    category: int
    default_value: str | None = None
    template_string: str | None = None
    required: bool = False
    overridable: bool = False


class Pedestal(BaseModel, frozen=True):
    """Represents a Pedestal configuration."""

    image_parts: dict[str, str | None]
    chart_name: str
    repo_name: str
    repo_url: str
    ns_pattern: str
    helm_values: list[HelmValue]

    def render_helm_values(
        self,
        index: int = 0,
        overrides: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Render all helm values into a nested dictionary.

        This is the main rendering function that processes all parameter configurations
        according to their type (Fixed/Dynamic) and overridable settings.

        Args:
            index: Index for dynamic value rendering (e.g., for port numbers in multi-instance deployments)
            overrides: Optional dictionary of user-provided values to override default values
                      Only values with overridable=True can be overridden.
                      Format: {"global.image.tag": "v2.0.0", "services.tsUiDashboard.nodePort": "32000"}

        Returns:
            Nested dictionary of rendered helm values ready for Helm installation

        Raises:
            ValueError: If any required parameter fails validation

        Example:
            >>> pedestal = Pedestal(...)
            >>> values = pedestal.render_helm_values(
            ...     index=0,
            ...     overrides={"global.image.tag": "v2.0.0"}
            ... )
            >>> # Returns:
            >>> # {
            >>> #     "global": {
            >>> #         "image": {
            >>> #             "repository": "registry.example.com/namespace",
            >>> #             "tag": "v2.0.0"  # overridden
            >>> #         }
            >>> #     }
            >>> # }
        """
        if overrides is None:
            overrides = {}

        # Prepare rendering context from image parts
        render_context = {
            "Registry": self.image_parts.get("registry", ""),
            "Namespace": self.image_parts.get("namespace", ""),
            "Repository": self.image_parts.get("repository", ""),
            "Tag": self.image_parts.get("tag", ""),
            "Index": index,
        }

        result: dict[str, Any] = {}

        for helm_value in self.helm_values:
            try:
                # Step 1: Determine the final value based on type and overridable
                final_value = self._resolve_parameter_value(
                    helm_value, render_context, overrides
                )

                # Step 2: Skip if value is None (optional parameter with no value)
                if final_value is None:
                    continue

                # Step 3: Set the value in the nested dictionary
                self._set_nested_dict_value(result, helm_value.key, final_value)

            except ValueError as e:
                # Log error and re-raise for required parameters
                console.print(
                    f"[red]Error processing parameter '{helm_value.key}': {e}[/red]"
                )
                raise

        return result

    def _resolve_parameter_value(
        self,
        helm_value: HelmValue,
        context: dict[str, Any],
        overrides: dict[str, Any],
    ) -> Any:
        """Resolve the final value for a parameter based on type and override rules.

        Resolution priority:
        1. If type=0 (Fixed) and overridable=True: Use override if provided, else default_value
        2. If type=0 (Fixed) and overridable=False: Use default_value only
        3. If type=1 (Dynamic): Always render template_string (overrides not applicable)

        Args:
            helm_value: Parameter configuration
            context: Rendering context (Registry, Namespace, Tag, Index)
            overrides: User-provided override values

        Returns:
            Resolved value (may be string, int, bool, or None)

        Raises:
            ValueError: If required parameter has no value or validation fails
        """
        # Type 0: Fixed value
        if helm_value.type == 0:
            # Check if user provided an override
            if helm_value.overridable and helm_value.key in overrides:
                final_value = overrides[helm_value.key]
                console.print(
                    f"[yellow]Override: {helm_value.key} = {final_value}[/yellow]"
                )
            else:
                # Use default_value
                final_value = helm_value.default_value

            # Validate required parameters
            if final_value is None:
                if helm_value.required:
                    raise ValueError(
                        f"Required parameter '{helm_value.key}' has no value "
                        f"(default_value is None and no override provided)"
                    )
                # Optional parameter with no value - skip it
                return None

            # Convert string to appropriate type
            return self._convert_value_type(final_value)

        # Type 1: Dynamic value (template rendering)
        elif helm_value.type == 1:
            if helm_value.template_string is None or helm_value.template_string == "":
                raise ValueError(
                    f"Dynamic parameter '{helm_value.key}' is missing template_string"
                )

            # Render the template
            rendered_value = self._render_template(helm_value.template_string, context)

            # Validate required parameters
            if helm_value.required and not rendered_value:
                raise ValueError(
                    f"Required dynamic parameter '{helm_value.key}' rendered to empty string"
                )

            # Skip if empty (for optional parameters)
            if not rendered_value:
                return None

            return rendered_value

        else:
            raise ValueError(
                f"Unknown parameter type '{helm_value.type}' for key '{helm_value.key}'"
            )

    def _render_template(self, template_str: str, context: dict[str, Any]) -> str:
        """Render a template string with the given context.

        Supports two template formats:
        1. Go-style templates: {{ .Registry }}/{{ .Namespace }}
        2. Python format strings: 31%03d

        Args:
            template_str: Template string to render
            context: Context dictionary containing variables

        Returns:
            Rendered string

        Raises:
            ValueError: If template rendering fails
        """
        # Extract Go-style template variables (e.g., {{ .Registry }} -> ["Registry"])
        var_pattern = re.compile(r"\{\{\s*\.(\w+)\s*\}\}")
        template_vars = var_pattern.findall(template_str)

        # If no template variables found
        if not template_vars:
            # Check if it's a Python format string (e.g., "31%03d")
            if "%" in template_str:
                try:
                    return template_str % context["Index"]
                except (ValueError, TypeError, KeyError) as e:
                    raise ValueError(f"Failed to format template '{template_str}': {e}")
            else:
                # Plain string, return as-is
                return template_str

        # Convert Go-style {{ .Registry }} to Jinja2 {{ Registry }}
        jinja_template = re.sub(r"\{\{\s*\.(\w+)\s*\}\}", r"{{ \1 }}", template_str)

        try:
            template = Template(jinja_template)
            return template.render(context)
        except Exception as e:
            raise ValueError(f"Failed to render template '{template_str}': {e}")

    def _convert_value_type(self, value: Any) -> Any:
        """Convert string value to appropriate Python type.

        Args:
            value: Value to convert (typically a string)

        Returns:
            Converted value (bool or str)

        Note:
            Numeric strings are kept as strings to preserve formatting (e.g., "023" stays "023").
            Only boolean values are converted from strings.
        """
        if not isinstance(value, str):
            return value

        # Try to parse as boolean
        if value.lower() in ("true", "false"):
            return value.lower() == "true"

        # Keep all other values as strings (including numeric strings like "023")
        # This preserves leading zeros and allows Helm to handle type conversion
        return value

    def to_helm_release(
        self,
        env: ENV,
        namespace: str,
        index: int = 0,
        overrides: dict[str, Any] | None = None,
    ) -> HelmRelease:
        """Convert Pedestal to a HelmRelease with rendered values.

        Args:
            env: Environment for Kubernetes context
            namespace: Namespace to install into
            index: Index for dynamic value rendering (e.g., for port numbers)
            overrides: Optional dictionary of user-provided override values

        Returns:
            HelmRelease configured with rendered values
        """
        extra_args: list[str] = []

        if self.helm_values:
            # Use the centralized render_helm_values function
            rendered_values = self.render_helm_values(index=index, overrides=overrides)

            # Convert to --set format
            extra_args.extend(self._convert_helm_values_to_set_list(rendered_values))

        extra_args.extend(
            ["--kube-context", KubernetesManager.get_context_mapping()[env]]
        )

        return HelmRelease(
            name=namespace,
            chart=f"{self.repo_name}/{self.chart_name}",
            namespace=namespace,
            repo_name=self.repo_name,
            repo_url=self.repo_url,
            create_namespace=True,
            extra_args=extra_args,
        )

    def _convert_helm_values_to_set_list(
        self,
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
                self._convert_helm_values_to_set_list(v, current_key, key_value_pairs)
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

    def _set_nested_dict_value(self, d: dict[str, Any], key: str, value: Any) -> None:
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


def _load_pedestals(env: ENV, name: str) -> Pedestal | None:
    """Load Pedestal configuration from database.

    Args:
        env: Environment to load configuration from
        name: Container name (e.g., 'ts_cn')

    Returns:
        Pedestal object if found, None otherwise
    """
    mysql_config = mysql_configs[env]
    mysql_client = MysqlClient(mysql_config)
    session = mysql_client.get_session()

    try:
        # Query to get the latest container version with helm_config
        query = text("""
            SELECT 
                cv.registry,
                cv.namespace,
                cv.repository,
                cv.tag,
                hc.repo_url,
                hc.repo_name,
                hc.chart_name,
                hc.ns_pattern,
                hc.id as helm_config_id
            FROM containers c
            JOIN container_versions cv ON c.id = cv.container_id
            JOIN helm_configs hc ON cv.id = hc.container_version_id
            WHERE c.name = :name 
                AND c.type = 2 
                AND c.status >= 0 
                AND cv.status >= 0
            ORDER BY cv.name_major DESC, cv.name_minor DESC, cv.name_patch DESC
            LIMIT 1
        """)

        result = session.execute(query, {"name": name}).fetchone()

        if not result:
            return None

        # Parse image parts
        image_parts = {
            "registry": result[0],
            "namespace": result[1],
            "repository": result[2],
            "tag": result[3],
        }

        helm_config_id = result[8]

        # Query to get helm values (parameter configs)
        values_query = text("""
            SELECT 
                pc.config_key,
                pc.type,
                pc.category,
                pc.default_value,
                pc.template_string,
                pc.required,
                pc.overridable
            FROM helm_config_values hcv
            JOIN parameter_configs pc ON hcv.parameter_config_id = pc.id
            WHERE hcv.helm_config_id = :helm_config_id
            ORDER BY pc.id
        """)

        values_result = session.execute(
            values_query, {"helm_config_id": helm_config_id}
        ).fetchall()

        # Build helm_values list
        helm_values = []
        for row in values_result:
            helm_value = HelmValue(
                key=row[0],
                type=row[1],
                category=row[2],
                default_value=row[3],
                template_string=row[4],
                required=bool(row[5]),
                overridable=bool(row[6]),
            )
            helm_values.append(helm_value)

        return Pedestal.model_validate(
            {
                "image_parts": image_parts,
                "repo_name": result[5],
                "repo_url": result[4],
                "chart_name": result[6],
                "ns_pattern": result[7],
                "helm_values": helm_values,
            }
        )

    except Exception as e:
        console.print(f"[bold red]Error loading pedestal from database: {e}[/bold red]")
        return None
    finally:
        session.close()


def _get_pedestal_or_exit(env: ENV, name: str) -> Pedestal:
    """Retrieve a Pedestal by name or exit with an error message if not found.

    Args:
        name: Container name to retrieve
        env: Environment to load from

    Returns:
        Pedestal object

    Raises:
        SystemExit: If pedestal not found
    """
    pedestals = _load_pedestals(env, name)
    if not pedestals:
        console.print(f"[bold red]Pedestal '{name}' not found or invalid[/bold red]")
        raise SystemExit(1)
    return pedestals


@with_k8s_manager()
def install_pedestals(
    env: ENV,
    k8s_manager: KubernetesManager,
    name: str,
    count: int,
    values_file: Path | None = None,
    dry_run: bool = False,
    force: bool = False,
) -> None:
    if count <= 0:
        console.print("[bold red]PEDESTAL_COUNT must be a positive number[/bold red]")
        raise SystemExit(1)

    helm_cli = HelmCLI()

    pedestal = _get_pedestal_or_exit(env, name=name)

    # Extract prefix from pattern (e.g., "^ts\d+$" -> "ts")
    # The pattern should be like "^ts\d+$" which matches "ts0", "ts1", etc.
    ns_pattern = pedestal.ns_pattern
    ns_prefix = _extract_prefix_from_pattern(ns_pattern)

    console.print(
        f"[bold blue]Checking Helm releases in namespaces {ns_prefix}0 to {ns_prefix}{count - 1}...[/bold blue]"
    )

    all_finished: list[bool] = []
    for i in range(count):
        ns = f"{ns_prefix}{i}"
        console.print(f"[bold blue]Checking namespace: {ns}[/bold blue]")

        ns_ok = k8s_manager.check_and_create_namespace(ns)
        if not ns_ok:
            console.print(f"[bold yellow]Namespace {ns} does not exist[/bold yellow]")
            continue

        console.print()

        console.print(
            f"[bold blue]Checking Helm release '{ns}' in namespace {ns}[/bold blue]"
        )
        has_release = helm_cli.is_release_exist(ns, namespace=ns)
        if has_release:
            console.print(f"[gray]Helm release '{ns}' found in namespace {ns}[/gray]")
            if force:
                console.print()
                helm_cli.uninstall(
                    ns,
                    namespace=ns,
                    verbose=True,
                    wait=True,
                    extra_args=[
                        "--kube-context",
                        KubernetesManager.get_context_mapping()[env],
                    ],
                )
            else:
                continue
        else:
            console.print(
                f"[bold yellow]Helm release '{ns}' not found in namespace {ns}[/bold yellow]"
            )

        console.print()

        overrides: dict[str, Any] | None = None
        if values_file is not None:
            with open(values_file, encoding="utf-8") as f:
                overrides = yaml.safe_load(f)

        release = pedestal.to_helm_release(
            env, namespace=ns, index=i, overrides=overrides
        )
        helm_cli.install(
            release, verbose=True, wait=True, timeout="10m0s", dry_run=dry_run
        )
        all_finished.append(True)

        console.print(
            f"[bold green]Installed Helm release '{ns}' in namespace {ns}[/bold green]"
        )
        console.print()

    if all(all_finished):
        console.print("[bold green]🎉 Check and installation completed![/bold green]")
    else:
        console.print(
            "[bold yellow]⚠️ Some installations failed. Please check the logs above.[/bold yellow]"
        )
