import copy
import json
import shutil
import sys
from pathlib import Path
from typing import Any

from src.common.command import run_command
from src.common.common import console
from src.swagger.common import SWAGGER_ROOT, RunMode
from src.util import get_longest_common_substring

OPENAPI2_DIR = SWAGGER_ROOT / "openapi2"
OPENAPI3_DIR = SWAGGER_ROOT / "openapi3"
CONVERTED_DIR = SWAGGER_ROOT / "converted"

__all__ = ["init"]


def audience_flag_enabled(x_api_type: Any, audience: str) -> bool:
    """Return whether an x-api-type audience flag is enabled."""
    if not isinstance(x_api_type, dict):
        return False

    value = x_api_type.get(audience)
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.strip().lower() == "true"
    return False


def normalize_openapi_ref(ref: str) -> str:
    """Convert Swagger 2 refs to OpenAPI 3 component refs."""
    return (
        ref.replace("#/definitions/", "#/components/schemas/")
        .replace("#/parameters/", "#/components/parameters/")
        .replace("#/responses/", "#/components/responses/")
    )


def convert_schema_object(schema: Any) -> Any:
    """Recursively convert a Swagger 2 schema object into OpenAPI 3 format."""
    if isinstance(schema, dict):
        if schema.get("type") == "file":
            converted_file_schema = dict(schema)
            converted_file_schema["type"] = "string"
            converted_file_schema["format"] = "binary"
            return converted_file_schema

        converted: dict[str, Any] = {}
        for key, value in schema.items():
            if key == "$ref" and isinstance(value, str):
                converted[key] = normalize_openapi_ref(value)
                continue

            if key in {
                "schema",
                "items",
                "additionalProperties",
                "not",
                "propertyNames",
                "contains",
            }:
                converted[key] = convert_schema_object(value)
                continue

            if key in {"allOf", "anyOf", "oneOf"} and isinstance(value, list):
                converted[key] = [convert_schema_object(item) for item in value]
                continue

            if key == "properties" and isinstance(value, dict):
                converted[key] = {
                    name: convert_schema_object(prop) for name, prop in value.items()
                }
                continue

            converted[key] = convert_schema_object(value)

        return converted

    if isinstance(schema, list):
        return [convert_schema_object(item) for item in schema]

    return schema


def convert_swagger_parameter(parameter: dict[str, Any]) -> dict[str, Any]:
    """Convert a non-body Swagger 2 parameter to OpenAPI 3."""
    converted = copy.deepcopy(parameter)
    schema: dict[str, Any] = {}

    for field in (
        "type",
        "format",
        "items",
        "enum",
        "default",
        "minimum",
        "maximum",
        "exclusiveMinimum",
        "exclusiveMaximum",
        "minLength",
        "maxLength",
        "pattern",
        "multipleOf",
        "minItems",
        "maxItems",
        "uniqueItems",
    ):
        if field in converted:
            schema[field] = convert_schema_object(converted.pop(field))

    if "collectionFormat" in converted:
        collection_format = converted.pop("collectionFormat")
        if collection_format == "multi":
            converted["style"] = "form"
            converted["explode"] = True
        elif collection_format == "csv":
            converted["style"] = "form"
            converted["explode"] = False

    if schema:
        converted["schema"] = schema

    return converted


def make_request_body_content(
    schema: dict[str, Any], media_types: list[str]
) -> dict[str, Any]:
    """Build an OpenAPI 3 requestBody content map."""
    return {media_type: {"schema": schema} for media_type in media_types}


def convert_swagger_operation(
    operation: dict[str, Any],
    global_consumes: list[str],
    global_produces: list[str],
) -> dict[str, Any]:
    """Convert a Swagger 2 operation to OpenAPI 3."""
    converted = copy.deepcopy(operation)
    consumes = (
        converted.pop("consumes", None) or global_consumes or ["application/json"]
    )
    produces = (
        converted.pop("produces", None) or global_produces or ["application/json"]
    )

    parameters = converted.pop("parameters", [])
    request_body: dict[str, Any] | None = None
    form_properties: dict[str, Any] = {}
    form_required: list[str] = []
    converted_parameters: list[dict[str, Any]] = []

    for parameter in parameters:
        if not isinstance(parameter, dict):
            continue

        if "$ref" in parameter:
            parameter_ref = dict(parameter)
            parameter_ref["$ref"] = normalize_openapi_ref(parameter_ref["$ref"])
            converted_parameters.append(parameter_ref)
            continue

        location = parameter.get("in")
        if location == "body":
            request_schema = convert_schema_object(parameter.get("schema", {}))
            request_body = {
                "required": parameter.get("required", False),
                "content": make_request_body_content(request_schema, consumes),
            }
            if parameter.get("description"):
                request_body["description"] = parameter["description"]
            continue

        if location == "formData":
            property_schema: dict[str, Any] = {}
            parameter_type = parameter.get("type")
            if parameter_type == "file":
                property_schema = {"type": "string", "format": "binary"}
            else:
                property_schema = {
                    "type": parameter_type,
                }
                if "format" in parameter:
                    property_schema["format"] = parameter["format"]
                if "enum" in parameter:
                    property_schema["enum"] = parameter["enum"]
                if "items" in parameter:
                    property_schema["items"] = convert_schema_object(parameter["items"])
                if "default" in parameter:
                    property_schema["default"] = parameter["default"]

            if parameter.get("description"):
                property_schema["description"] = parameter["description"]

            form_properties[parameter["name"]] = property_schema
            if parameter.get("required"):
                form_required.append(parameter["name"])
            continue

        converted_parameters.append(convert_swagger_parameter(parameter))

    if form_properties:
        request_body = {
            "required": bool(form_required),
            "content": {
                "multipart/form-data": {
                    "schema": {
                        "type": "object",
                        "properties": form_properties,
                    }
                }
            },
        }
        if form_required:
            request_body["content"]["multipart/form-data"]["schema"]["required"] = (
                form_required
            )

    if converted_parameters:
        converted["parameters"] = converted_parameters
    if request_body is not None:
        converted["requestBody"] = request_body

    converted_responses: dict[str, Any] = {}
    for code, response in converted.get("responses", {}).items():
        if not isinstance(response, dict):
            converted_responses[code] = response
            continue

        response_copy = copy.deepcopy(response)
        response_schema = response_copy.pop("schema", None)
        if response_schema is not None:
            response_copy["content"] = {
                media_type: {"schema": convert_schema_object(response_schema)}
                for media_type in produces
            }
        converted_responses[code] = convert_schema_object(response_copy)

    converted["responses"] = converted_responses
    return convert_schema_object(converted)


def convert_swagger2_to_openapi3(swagger_data: dict[str, Any]) -> dict[str, Any]:
    """Convert the generated Swagger 2 document into a full OpenAPI 3 document."""
    openapi_data = copy.deepcopy(swagger_data)
    openapi_data["openapi"] = "3.0.3"
    openapi_data.pop("swagger", None)

    global_consumes = openapi_data.pop("consumes", None) or []
    global_produces = openapi_data.pop("produces", None) or []

    components: dict[str, Any] = {}
    definitions = openapi_data.pop("definitions", {})
    if definitions:
        components["schemas"] = {
            name: convert_schema_object(schema) for name, schema in definitions.items()
        }

    parameters = openapi_data.pop("parameters", {})
    if parameters:
        components["parameters"] = {
            name: convert_swagger_parameter(parameter)
            for name, parameter in parameters.items()
        }

    responses = openapi_data.pop("responses", {})
    if responses:
        components["responses"] = {
            name: convert_schema_object(response)
            for name, response in responses.items()
        }

    security_definitions = openapi_data.pop("securityDefinitions", {})
    if security_definitions:
        components["securitySchemes"] = {
            name: convert_schema_object(scheme)
            for name, scheme in security_definitions.items()
        }

    if components:
        openapi_data["components"] = components

    host = openapi_data.pop("host", "")
    base_path = openapi_data.pop("basePath", "") or ""
    schemes = openapi_data.pop("schemes", None) or []
    if host:
        if host.startswith("http://") or host.startswith("https://"):
            server_url = host.rstrip("/")
        else:
            scheme = schemes[0] if schemes else "http"
            server_url = f"{scheme}://{host}".rstrip("/")
        if base_path:
            server_url = f"{server_url}{base_path}"
        openapi_data["servers"] = [{"url": server_url}]

    converted_paths: dict[str, Any] = {}
    for path, path_item in openapi_data.get("paths", {}).items():
        if not isinstance(path_item, dict):
            converted_paths[path] = path_item
            continue

        converted_path_item: dict[str, Any] = {}
        path_level_parameters = path_item.get("parameters", [])
        if path_level_parameters:
            converted_path_item["parameters"] = [
                convert_swagger_parameter(parameter)
                if isinstance(parameter, dict) and "$ref" not in parameter
                else {"$ref": normalize_openapi_ref(parameter["$ref"])}
                for parameter in path_level_parameters
            ]

        for method, operation in path_item.items():
            if method == "parameters":
                continue
            if not isinstance(operation, dict):
                converted_path_item[method] = operation
                continue
            converted_path_item[method] = convert_swagger_operation(
                operation, global_consumes, global_produces
            )

        converted_paths[path] = converted_path_item

    openapi_data["paths"] = converted_paths
    return convert_schema_object(openapi_data)


class SDKPostProcesser:
    """Process Swagger JSON to add SSE extensions and update model names."""

    SSE_FLAG = "x-request-type"
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

    # Models that should always be kept in SDK even if not directly referenced
    # These are typically used in SSE events or other indirect references
    ALWAYS_KEEP_MODELS = {
        "TraceStreamEvent",
        "GroupStreamEvent",
        "DatapackInfo",
        "DatapackResult",
        "ExecutionInfo",
        "ExecutionResult",
        "InfoPayloadTemplate",
        "JobMessage",
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

    def update_version(self, version: str) -> None:
        """Update the version field in the Swagger JSON."""
        if "info" in self.data:
            self.data["info"]["version"] = version

    def add_sse_extensions(self) -> None:
        """
        Add SSE extensions to Swagger JSON for APIs that produce 'text/event-stream'.
        """
        if "paths" not in self.data:
            console.print(
                "[bold yellow]'paths' field not found in JSON data[/bold yellow]"
            )
            return

        paths: dict[str, Any] = self.data["paths"]

        count = 0
        for path, operations in paths.items():
            for method, spec in operations.items():
                if self.SSE_FLAG in spec:
                    item: dict[str, str] = spec[self.SSE_FLAG]
                    if "stream" in item and item["stream"] == "true":
                        spec[self.SSE_EXTENSION] = True
                        count += 1
                        console.print(
                            f"[gray]    -> Added extension to {method.upper()} {path}[/gray]"
                        )

                        for code, response in spec["responses"].items():
                            if code == 200:
                                continue
                            if "content" not in response:
                                continue

                            if self.SSE_MIME_TYPE in response["content"]:
                                response["content"]["application/json"] = response[
                                    "content"
                                ].pop(self.SSE_MIME_TYPE)

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

            # Fix duplicate 'Chaos' prefix (e.g., ChaosChaosType -> ChaosType)
            if new_name.startswith("ChaosChaos"):
                new_name = new_name.replace("ChaosChaos", "Chaos", 1)

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

    def deduplicate_enum_values(self) -> None:
        """
        Remove duplicate values from enum arrays in schema definitions.
        This fixes the issue where swag generates duplicate enum values
        when parsing external package type aliases.
        """
        schemas = None
        if "definitions" in self.data:
            schemas = self.data["definitions"]
        elif "components" in self.data and "schemas" in self.data["components"]:
            schemas = self.data["components"]["schemas"]

        if not schemas:
            return

        count = 0
        for schema_name, schema_def in schemas.items():
            if not isinstance(schema_def, dict):
                continue

            # Check if this schema has an enum field
            if "enum" in schema_def and isinstance(schema_def["enum"], list):
                original_enum = schema_def["enum"]
                original_len = len(original_enum)

                # Remove duplicates while preserving order
                seen = set()
                deduped_enum = []
                for value in original_enum:
                    if value not in seen:
                        seen.add(value)
                        deduped_enum.append(value)

                # Update only if duplicates were found
                if len(deduped_enum) < original_len:
                    schema_def["enum"] = deduped_enum
                    count += 1
                    console.print(
                        f"[gray]   {schema_name}: removed {original_len - len(deduped_enum)} duplicate enum values[/gray]"
                    )

                    # Also deduplicate x-enum-varnames if present
                    if "x-enum-varnames" in schema_def and isinstance(
                        schema_def["x-enum-varnames"], list
                    ):
                        varnames = schema_def["x-enum-varnames"]
                        # Take only the first half if length matches original enum
                        if len(varnames) == original_len:
                            schema_def["x-enum-varnames"] = varnames[
                                : len(deduped_enum)
                            ]

        if count > 0:
            console.print(
                f"[bold green]✅ Deduplicated enums in {count} schemas[/bold green]"
            )

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

    def output(self, output_file: Path, category: RunMode) -> None:
        output_data = self.data
        if category != RunMode.CLIENT:
            output_data = self._filter_apis_by_audience(category)
            if output_data is None:
                console.print("[bold red]Processing function returned None[/bold red]")
                sys.exit(1)

        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(output_data, f, indent=2)

    def _filter_apis_by_audience(self, category: RunMode) -> dict[str, Any] | None:
        """
        Filter Swagger JSON according to the x-api-type audience flags.
        """
        audience_keys_by_mode = {
            RunMode.SDK: {"sdk"},
            RunMode.PORTAL: {"portal"},
            RunMode.ADMIN: {"admin"},
        }
        audience_keys = audience_keys_by_mode.get(category)
        if not audience_keys:
            return copy.deepcopy(self.data)

        new_data = copy.deepcopy(self.data)

        # Step 1: Filter paths - keep only operations tagged for the target audience.
        original_paths = new_data["paths"]
        filtered_paths = {}
        removed_count = 0
        kept_count = 0

        for path, operations in original_paths.items():
            filtered_operations = {}
            for method, spec in operations.items():
                x_api_type = spec.get("x-api-type", {})
                if any(audience_flag_enabled(x_api_type, key) for key in audience_keys):
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

        # Add models that should always be kept
        used_models.update(self.ALWAYS_KEEP_MODELS)
        console.print(
            f"[gray]   → Force-keeping {len(self.ALWAYS_KEEP_MODELS)} models: {', '.join(sorted(self.ALWAYS_KEEP_MODELS))}[/gray]"
        )

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

        console.print(
            f"[gray]\n   {category.value} operations: {kept_count} kept, {removed_count} removed[/gray]"
        )

        return new_data


def init(version: str) -> None:
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

    # 2. Convert Swagger 2.0 into a full OpenAPI 3 document locally.
    swagger2_file = OPENAPI2_DIR / "swagger.json"
    if not swagger2_file.exists():
        console.print(f"[bold red]{swagger2_file} not found[/bold red]")
        sys.exit(1)

    if OPENAPI3_DIR.exists():
        shutil.rmtree(OPENAPI3_DIR)
    OPENAPI3_DIR.mkdir(parents=True)

    with open(swagger2_file, encoding="utf-8") as f:
        swagger2_data = json.load(f)

    openapi3_data = convert_swagger2_to_openapi3(swagger2_data)
    with open(OPENAPI3_DIR / "openapi.json", "w", encoding="utf-8") as f:
        json.dump(openapi3_data, f, indent=2)

    # 3. Post-process Swagger JSON
    console.print("[bold blue]📦 Post-processing swagger initiaization...[/bold blue]")

    if not CONVERTED_DIR.exists():
        CONVERTED_DIR.mkdir(parents=True)
    else:
        legacy_typescript_file = CONVERTED_DIR / "typescript.json"
        legacy_typescript_file.unlink(missing_ok=True)

    post_input_file = OPENAPI3_DIR / "openapi.json"
    client_file = CONVERTED_DIR / "client.json"
    sdk_file = CONVERTED_DIR / "sdk.json"
    portal_file = CONVERTED_DIR / "portal.json"
    admin_file = CONVERTED_DIR / "admin.json"

    shutil.copyfile(post_input_file, dst=client_file)
    shutil.copyfile(post_input_file, dst=sdk_file)
    shutil.copyfile(post_input_file, dst=portal_file)
    shutil.copyfile(post_input_file, dst=admin_file)

    processor = SDKPostProcesser(post_input_file)
    processor.update_version(version)
    processor.add_sse_extensions()
    processor.update_model_name()
    processor.deduplicate_enum_values()  # Remove duplicate enum values
    processor.convert_inline_enums_to_refs()

    processor.output(client_file, RunMode.CLIENT)
    processor.output(sdk_file, RunMode.SDK)
    processor.output(portal_file, RunMode.PORTAL)
    processor.output(admin_file, RunMode.ADMIN)

    console.print(
        "[bold green]✅ Swagger documentation generation completed successfully![/bold green]"
    )
