import copy
import json
import shutil
import sys
from pathlib import Path
from typing import Any

from python_on_whales import docker

from src.common.command import run_command
from src.common.common import console
from src.formatter import PythonFormatter
from src.util import get_longest_common_substring

# Project root: /home/nn/workspace/AegisLab
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent.parent
SWAGGER_ROOT = PROJECT_ROOT / "src" / "docs"
OPENAPI2_DIR = SWAGGER_ROOT / "openapi2"
OPENAPI3_DIR = SWAGGER_ROOT / "openapi3"
CONVERTED_DIR = SWAGGER_ROOT / "converted"

PYTHON_SDK_DIR = PROJECT_ROOT / "sdk" / "python"
PYTHON_SDK_GEN_DIR = PROJECT_ROOT / "sdk" / "python-gen"
PYTHON_GENERATOR_CONFIG_DIR = PROJECT_ROOT / ".openapi-generator" / "python"

GENERATOR_IMAGE = "docker.io/opspai/openapi-generator-cli:1.0.0"


class PostProcesser:
    """Process Swagger JSON to add SSE extensions and update model names."""

    SSE_MIME_TYPE = "text/event-stream"
    SSE_EXTENSION = "x-is-streaming-api"

    # Parameter to schema mapping for converting inline enums to $ref
    # Key format: "path|parameter_name" to handle same parameter names in different paths
    # Value: schema name to reference
    PARAMETER_SCHEMA_MAPPING = {
        # Container APIs
        "containers|type": "ContainerType",
        # Task APIs
        "tasks|state": "TaskState",
        "tasks|type": "TaskType",
        # Execution APIs
        "executions|state": "ExecutionState",
        # Injection APIs
        "injections|state": "DatapackState",
        # Label APIs
        "labels|category": "LabelCategory",
        # Resource APIs
        "resources|type": "ResourceType",
        "resources|category": "ResourceCategory",
        # Generic status filters (fallback for paths not specifically mapped)
        "*|size": "PageSize",
        "*|status": "StatusType",
    }

    def __init__(self, file_path: Path) -> None:
        self.file_path = file_path
        self.data: dict[str, Any] = {}
        self._read_json()

    def _read_json(self) -> None:
        """Read JSON data from the specified file path."""
        if not self.file_path.exists():
            console.print(f"[bold red]{self.file_path} not found[/bold red]")
            sys.exit(1)

        with open(self.file_path, encoding="utf-8") as f:
            data = json.load(f)

        if data is None or not isinstance(data, dict):
            console.print("[bold red]Unexpected JSON structure[/bold red]")
            sys.exit(1)

        self.data = data

    def add_sse_extensions(self) -> None:
        """
        Add SSE extensions to Swagger JSON for APIs that produce 'text/event-stream'.
        """
        if "paths" not in self.data:
            console.print(
                "[bold yellow]'paths' field not found in JSON data[/bold yellow]"
            )
            return

        paths = self.data["paths"]

        count = 0
        for path, operations in paths.items():
            for method, spec in operations.items():
                produces = spec.get("produces")
                if produces and self.SSE_MIME_TYPE in produces:
                    if self.SSE_EXTENSION not in spec:
                        spec[self.SSE_EXTENSION] = True
                        count += 1
                        console.print(
                            f"[gray]   -> Added extension to {method.upper()} {path}[/gray]"
                        )

        console.print(
            f"[bold green]✅ SSE extensions added successfully ({count} apis added)[/bold green]"
        )

    def update_model_name(self) -> None:
        """
        Clean up model names in OpenAPI schemas:
        1. Remove 'consts.' prefix from all constant types
        2. Remove 'dto.' prefix from all DTO types
        3. Replace 'handler.' prefix with 'Chaos' for handler types
        4. Update all $ref references accordingly

        Supports both OpenAPI 2.0 (definitions) and OpenAPI 3.0 (components.schemas)
        """
        schemas = None
        schema_path = ""

        schemas = self.data["components"]["schemas"]
        schema_path = "#/components/schemas/"

        name_mapping = {}

        for old_name in list(schemas.keys()):
            new_name = old_name

            # Remove 'consts.' prefix
            if new_name.startswith("consts."):
                new_name = new_name.replace("consts.", "", 1)

            # Remove 'dto.' prefix
            if new_name.startswith("dto."):
                new_name = new_name.replace("dto.", "", 1)

            # Replace 'handler.' prefix with 'Chaos'
            if new_name.startswith("handler."):
                new_name = new_name.replace("handler.", "Chaos", 1)

            # Also handle nested patterns like 'dto.GenericResponse-dto_XXX'
            # Convert to 'GenericResponse-XXX'
            new_name = new_name.replace("dto_", "")

            # Normalize GenericResponse-ListResp-XXX patterns:
            #   GenericResponse-ListResp-AuditLogResp -> GenericResponseListAuditLogResp
            #   GenericResponse-ListResp-ContainerResp -> GenericResponseListContainerResp
            if "GenericResponse-ListResp-" in new_name:
                new_name = new_name.replace(
                    "GenericResponse-ListResp-", "GenericResponseList"
                )

            # Normalize standalone list response types:
            #   ListResp-AuditLogResp -> ListAuditLogResp
            #   ListResp-ContainerResp -> ListContainerResp
            elif new_name.startswith("ListResp-") and len(new_name) > len("ListResp-"):
                new_name = "List" + new_name[len("ListResp-") :]

            if old_name != new_name:
                name_mapping[old_name] = new_name
                console.print(f"[gray]   {old_name} -> {new_name}[/gray]")

        # Step 1: Rename keys in schemas
        new_schemas = {}
        for old_name, schema_def in schemas.items():
            new_name = name_mapping.get(old_name, old_name)
            new_schemas[new_name] = schema_def

        for key, value in new_schemas.items():
            if "enum" not in value:
                continue

            varnames = value.get("x-enum-varnames", [])
            if varnames:
                lcs_varnames = get_longest_common_substring(key, strs=varnames)
                if key == "StatusType":
                    lcs_varnames = "Common"

                if len(lcs_varnames) > 1:
                    value["x-enum-varnames"] = [
                        s.replace(lcs_varnames, "") for s in varnames
                    ]

            comments = value.get("x-enum-comments", {})
            comment_keys: list[str] = []
            if comments:
                comment_keys = list(comments.keys())
                lcs_comments = get_longest_common_substring(key, strs=comment_keys)
                if len(lcs_varnames) > 1:
                    value["x-enum-comments"] = dict(
                        [(k.replace(lcs_comments, ""), v) for k, v in comments.items()]
                    )

        # Update the schemas in the original structure
        self.data["components"]["schemas"] = new_schemas

        # Step 2: Update all $ref references throughout the entire JSON
        def update_refs(obj: dict[str, Any] | list[dict[str, Any]]) -> None:
            """Recursively update all $ref values in the JSON object"""
            if isinstance(obj, dict):
                for key, value in list(obj.items()):
                    if key == "$ref" and isinstance(value, str):
                        # Extract the schema name from the reference
                        if value.startswith(schema_path):
                            old_schema_name = value.replace(schema_path, "")
                            new_schema_name = name_mapping.get(
                                old_schema_name, old_schema_name
                            )
                            obj[key] = f"{schema_path}{new_schema_name}"
                    else:
                        update_refs(value)
            elif isinstance(obj, list):
                for item in obj:
                    update_refs(item)

        # Update refs in paths
        update_refs(self.data.get("paths", {}))

        # Update refs in schemas themselves
        if "definitions" in self.data:
            update_refs(self.data["definitions"])
        elif "components" in self.data and "schemas" in self.data["components"]:
            update_refs(self.data["components"]["schemas"])

    def convert_inline_enums_to_refs(self) -> None:
        """
        Convert inline enum definitions in parameters to $ref references.

        Example transformation:

        Before:
          parameters:
            - name: type
              in: query
              schema:
                type: integer
                enum: [0, 1, 2]

        After:
          parameters:
            - name: type
              in: query
              schema:
                $ref: '#/components/schemas/ContainerType'
        """
        if "paths" not in self.data:
            return

        schema_path = "#/components/schemas/"
        converted_count = 0

        def process_parameters(
            params: list[dict[str, Any]], path: str, method: str
        ) -> None:
            """Process parameters and convert inline enums to refs."""
            nonlocal converted_count

            for param in params:
                if not isinstance(param, dict):
                    continue

                param_name = param.get("name")
                schema = param.get("schema")

                # Skip if no schema or already a $ref
                if not schema or "$ref" in schema:
                    continue

                # Check if it's an inline enum definition
                if "enum" not in schema or schema.get("type") not in [
                    "integer",
                    "string",
                ]:
                    continue

                # Try path-specific mapping first
                resource = "*"
                for prefix in ["/api/v2/", "/system/"]:
                    if path.startswith(prefix):
                        resource = path.removeprefix(prefix)
                        console.print(
                            f"[gray]Match Found: Removed prefix '{prefix}'[/gray]"
                        )

                mapping_key = f"{resource}|{param_name}"
                target_schema = self.PARAMETER_SCHEMA_MAPPING.get(mapping_key)

                if not target_schema:
                    wildcard_key = f"*|{param_name}"
                    target_schema = self.PARAMETER_SCHEMA_MAPPING.get(wildcard_key)

                if target_schema:
                    # Replace inline enum with $ref
                    param["schema"] = {"$ref": f"{schema_path}{target_schema}"}
                    converted_count += 1
                    console.print(
                        f"[gray]   -> Converted {method.upper()} {path} parameter '{param_name}' to use schema '{target_schema}'[/gray]"
                    )

        # Process all paths and their operations
        for path, operations in self.data["paths"].items():
            if not isinstance(operations, dict):
                continue

            for method, spec in operations.items():
                if not isinstance(spec, dict):
                    continue

                # Process parameters at operation level
                if "parameters" in spec and isinstance(spec["parameters"], list):
                    process_parameters(spec["parameters"], path, method)

        if converted_count > 0:
            console.print(
                f"[bold green]✅ Converted {converted_count} inline enum parameters to schema references[/bold green]"
            )

    def output(
        self,
        output_file: Path,
        category: str | None = None,
    ) -> None:
        output_data = self.data
        if category == "sdk":
            output_data = self._filter_sdk_apis()
            if output_data is None:
                console.print("[bold red]Processing function returned None[/bold red]")
                sys.exit(1)

        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(output_data, f, indent=2)

    def _filter_sdk_apis(self) -> dict[str, Any] | None:
        """
        Filter Swagger JSON to only keep APIs marked with x-api-type: {"sdk": "true"}.
        Remove all other APIs and their unused model definitions.
        """
        new_data = copy.deepcopy(self.data)

        # Step 1: Filter paths - keep only APIs with x-api-type.sdk = "true"
        original_paths = new_data["paths"]
        filtered_paths = {}
        removed_count = 0
        kept_count = 0

        for path, operations in original_paths.items():
            filtered_operations = {}
            for method, spec in operations.items():
                x_api_type = spec.get("x-api-type", {})
                # Check if sdk is explicitly "true" (string)
                if x_api_type.get("sdk") == "true":
                    filtered_operations[method] = spec
                    kept_count += 1
                    console.print(f"[gray]   ✓ Kept: {method.upper()} {path}[/gray]")
                else:
                    removed_count += 1
                    console.print(f"[gray]   ✗ Removed: {method.upper()} {path}[/gray]")

            # Only add path if it has at least one operation
            if filtered_operations:
                filtered_paths[path] = filtered_operations

        new_data["paths"] = filtered_paths

        # Step 2: Determine schema location and collect model references
        used_models = set()

        schemas = new_data["components"]["schemas"]
        schema_path = "#/components/schemas/"

        def collect_refs(obj: dict[str, Any] | list[dict[str, Any]]) -> None:
            """Recursively collect all $ref model names"""
            if isinstance(obj, dict):
                for key, value in obj.items():
                    if key == "$ref" and isinstance(value, str):
                        if value.startswith(schema_path):
                            model_name = value.replace(schema_path, "")
                            used_models.add(model_name)
                    else:
                        collect_refs(value)
            elif isinstance(obj, list):
                for item in obj:
                    collect_refs(item)

        # Collect refs from filtered paths
        collect_refs(filtered_paths)

        # Step 3: Recursively collect nested model dependencies
        if schemas is not None:
            # Keep adding models until no new models are found
            prev_size = 0
            while len(used_models) != prev_size:
                prev_size = len(used_models)
                for model_name in list(used_models):
                    if model_name in schemas:
                        collect_refs(schemas[model_name])

        # Step 4: Filter schemas - keep only used models
        if schemas is not None:
            original_count = len(schemas)
            filtered_schemas = {
                name: schema_def
                for name, schema_def in schemas.items()
                if name in used_models
            }

            # Update schemas in the original structure
            if "definitions" in new_data:
                new_data["definitions"] = filtered_schemas
            elif "components" in new_data and "schemas" in new_data["components"]:
                new_data["components"]["schemas"] = filtered_schemas

            removed_models = original_count - len(filtered_schemas)
            console.print(
                f"[gray]\n   Models: {len(filtered_schemas)} kept, {removed_models} removed[/gray]"
            )

        return new_data


def init():
    """
    Initialize Swagger documentation by generating OpenAPI 2.0 and converting to OpenAPI 3.0.
    """
    console.print("[bold blue]📝 Initializing Swagger documentation...[/bold blue]")
    # 1. Swag Init
    src_dir = SWAGGER_ROOT.parent

    run_command(
        [
            "swag",
            "init",
            "-d",
            src_dir.as_posix(),
            "--parseDependency",
            "--parseDepth",
            "1",
            "--output",
            OPENAPI2_DIR.as_posix(),
        ]
    )

    # 2. Generate OpenAPI3 using OpenAPI Generator
    volume_path = Path("/local")
    relative_swagger = SWAGGER_ROOT.relative_to(PROJECT_ROOT)
    container_input_path = volume_path / relative_swagger / "openapi2" / "swagger.json"
    container_output_path = volume_path / relative_swagger / "openapi3"

    try:
        docker.run(
            GENERATOR_IMAGE,
            command=[
                "generate",
                "-i",
                container_input_path.as_posix(),
                "-g",
                "openapi",
                "-o",
                container_output_path.as_posix(),
            ],
            volumes=[(PROJECT_ROOT, volume_path)],
            remove=True,
        )
    except Exception as e:
        console.print(f"[bold_red]❌ Error during OpenAPI3 generation: {e}[/bold_red]")
        sys.exit(1)

    # 3. Post-process Swagger JSON
    console.print("[bold blue]📦 Post-processing swagger initiaization...[/bold blue]")

    if not CONVERTED_DIR.exists():
        CONVERTED_DIR.mkdir(parents=True)

    post_input_file = OPENAPI3_DIR / "openapi.json"
    api_file = CONVERTED_DIR / "api.json"
    sdk_file = CONVERTED_DIR / "sdk.json"

    shutil.copyfile(post_input_file, dst=api_file)
    shutil.copyfile(post_input_file, dst=sdk_file)

    processor = PostProcesser(post_input_file)
    processor.add_sse_extensions()
    processor.update_model_name()

    processor.output(api_file)

    processor.convert_inline_enums_to_refs()
    processor.output(sdk_file, "sdk")

    console.print(
        "[bold green]✅ Swagger documentation generation completed successfully![/bold green]"
    )


def generate_python_sdk(version: str) -> None:
    """
    Generate Python SDK from Swagger JSON using OpenAPI Generator.

    Post-process the generated SDK to adjust package structure and formatting.
    """
    # 1. Update config.json with the specified version
    sdk_config = PYTHON_GENERATOR_CONFIG_DIR / "config.json"
    with open(sdk_config) as f:
        config_data = json.load(f)

    config_data["packageVersion"] = version

    with open(sdk_config, "w") as f:
        json.dump(config_data, f, indent=2)

    console.print(f"[bold green]✅ Updated packageVersion to {version}[/bold green]")

    # 2. Generate SDK using OpenAPI Generator
    console.print("[bold blue]🐍 Generating Python SDK...[/bold blue]")

    if PYTHON_SDK_GEN_DIR.exists():
        shutil.rmtree(PYTHON_SDK_GEN_DIR)

    PYTHON_SDK_GEN_DIR.mkdir(parents=True)

    volume_path = Path("/local")
    relative_swagger = SWAGGER_ROOT.relative_to(PROJECT_ROOT)
    relative_sdk_gen = PYTHON_SDK_GEN_DIR.relative_to(PROJECT_ROOT)
    relative_generator_config = PYTHON_GENERATOR_CONFIG_DIR.relative_to(PROJECT_ROOT)

    container_input_path = volume_path / relative_swagger / "converted" / "sdk.json"
    container_output_path = volume_path / relative_sdk_gen
    container_config_path = volume_path / relative_generator_config / "config.json"
    container_templates_path = volume_path / relative_generator_config / "templates"

    # Get current user UID and GID to avoid permission issues
    import os

    current_user = os.getuid()
    current_group = os.getgid()

    try:
        docker.run(
            GENERATOR_IMAGE,
            command=[
                "generate",
                "-i",
                container_input_path.as_posix(),
                "-g",
                "python",
                "-o",
                container_output_path.as_posix(),
                "-c",
                container_config_path.as_posix(),
                "-t",
                container_templates_path.as_posix(),
                "--git-host",
                "github.com",
                "--git-repo-id",
                "AegisLab",
                "--git-user-id",
                "OperationsPAI",
            ],
            volumes=[(PROJECT_ROOT, volume_path)],
            user=f"{current_user}:{current_group}",
            remove=True,
        )

    except Exception as e:
        console.print(
            f"[bold_red]❌ Error during python sdk generation: {e}[/bold_red]"
        )
        sys.exit(1)

    console.print(
        "[bold green]✅ Original python SDK generated successfully![/bold green]"
    )

    # 3. Post-process generated SDK (if any post-processing is needed)
    console.print("[bold blue]📦 Post-processing generated SDK...[/bold blue]")

    dst = PYTHON_SDK_DIR / "src" / "rcabench" / "openapi"
    python_sdk_docs = PYTHON_SDK_DIR / "docs"
    python_sdk_pyproject = PYTHON_SDK_DIR / "pyproject.toml"

    if dst.exists():
        shutil.rmtree(dst)
    if python_sdk_docs.exists():
        shutil.rmtree(python_sdk_docs)
    if python_sdk_pyproject.exists():
        python_sdk_pyproject.unlink(missing_ok=True)

    shutil.copytree(PYTHON_SDK_GEN_DIR / "openapi", dst)
    shutil.copytree(PYTHON_SDK_GEN_DIR / "docs", python_sdk_docs)
    shutil.copyfile(PYTHON_SDK_GEN_DIR / "pyproject.toml", python_sdk_pyproject)

    old_str = "openapi"
    new_str = "rcabench.openapi"

    for filepath in dst.rglob("*.py"):
        if filepath.is_file():
            content = filepath.read_text(encoding="utf-8")
            new_content = content.replace(old_str, new_str)

            if new_content != content:
                filepath.write_text(new_content, encoding="utf-8")

    for filepath in python_sdk_docs.rglob("*.md"):
        if filepath.is_file():
            content = filepath.read_text(encoding="utf-8")
            new_content = content.replace(old_str, new_str)

            if new_content != content:
                filepath.write_text(new_content, encoding="utf-8")

    py_typed_files = dst / "py.typed"
    if py_typed_files.exists():
        py_typed_files.unlink(missing_ok=True)

    console.print(
        "[bold green]✅ Python SDK post procession completed successfully![/bold green]"
    )

    # 4. Format the generated SDK code
    console.print("[bold blue]📦 Formatting post-processed Python SDK...[/bold blue]")
    formatter = PythonFormatter()
    files_to_format = PythonFormatter.get_sdk_files(PYTHON_SDK_DIR.as_posix())
    formatter.run(files_to_format)

    console.print(
        "[bold green]✅ Python SDK generation completed successfully![/bold green]"
    )
