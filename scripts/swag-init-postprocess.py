import json
import sys
from typing import Any
from pathlib import Path
import shutil
import copy

from typing import Callable

DOCS_FOLDER = Path(__file__).parent.parent / "src" / "docs"

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


class Processer:
    """Process Swagger JSON to add SSE extensions and update model names."""

    def __init__(self, file_path: Path) -> None:
        self.file_path = file_path
        self.data: dict[str, Any] = {}
        self._read_json()

    def _read_json(self) -> None:
        """Read JSON data from the specified file path."""
        if not self.file_path.exists():
            print(f"Error: {self.file_path} not found")
            sys.exit(1)

        with open(self.file_path, encoding="utf-8") as f:
            data = json.load(f)

        if data is None or not isinstance(data, dict):
            print("Error: Unexpected JSON structure.")
            sys.exit(1)

        self.data = data

    def add_sse_extensions(self) -> None:
        """
        Add SSE extensions to Swagger JSON for APIs that produce 'text/event-stream'.
        """
        if "paths" not in self.data:
            print("Warning: 'paths' field not found in JSON data.")
            return

        paths = self.data["paths"]

        count = 0
        for path, operations in paths.items():
            for method, spec in operations.items():
                produces = spec.get("produces")
                if produces and SSE_MIME_TYPE in produces:
                    if SSE_EXTENSION not in spec:
                        spec[SSE_EXTENSION] = True
                        count += 1
                        print(f"   -> Added extension to {method.upper()} {path}")

        print(f"✅ SSE extensions added successfully ({count} apis added)")

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
                print(f"   {old_name} -> {new_name}")

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
                        print(f"Match Found: Removed prefix '{prefix}'")

                mapping_key = f"{resource}|{param_name}"
                target_schema = PARAMETER_SCHEMA_MAPPING.get(mapping_key)

                if not target_schema:
                    wildcard_key = f"*|{param_name}"
                    target_schema = PARAMETER_SCHEMA_MAPPING.get(wildcard_key)

                if target_schema:
                    # Replace inline enum with $ref
                    param["schema"] = {"$ref": f"{schema_path}{target_schema}"}
                    converted_count += 1
                    print(
                        f"   -> Converted {method.upper()} {path} parameter '{param_name}' to use schema '{target_schema}'"
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
            print(
                f"✅ Converted {converted_count} inline enum parameters to schema references"
            )

    def output(
        self,
        output_file: Path,
        function: Callable[[dict[str, Any]], dict[str, Any] | None] | None = None,
    ) -> None:
        data = self.data
        if function:
            data = function(data)
            if data is None:
                print("Error: Processing function returned None.")
                sys.exit(1)

        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)


def filter_sdk_apis(data: dict[str, Any]) -> dict[str, Any] | None:
    """
    Filter Swagger JSON to only keep APIs marked with x-api-type: {"sdk": "true"}.
    Remove all other APIs and their unused model definitions.
    """
    new_data = copy.deepcopy(data)

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
                print(f"   ✓ Kept: {method.upper()} {path}")
            else:
                removed_count += 1
                print(f"   ✗ Removed: {method.upper()} {path}")

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
        print(f"\n   Models: {len(filtered_schemas)} kept, {removed_models} removed")

    return new_data


def get_longest_common_substring(key: str, strs: list[str]) -> str:
    """Find the longest common substring among a list of strings, including the key."""
    if not strs:
        return ""

    newStrs = copy.deepcopy(strs)
    newStrs.insert(0, key)

    shortest_str = min(newStrs, key=len)
    n = len(shortest_str)

    for length in range(n, 0, -1):
        for i in range(n - length + 1):
            substring = shortest_str[i : i + length]
            if all(substring in s for s in newStrs):
                return substring

    return ""


def main() -> None:
    converted_folder = DOCS_FOLDER / "converted"
    if not converted_folder.exists():
        converted_folder.mkdir(parents=True)

    src_file = DOCS_FOLDER / "openapi3" / "openapi.json"
    api_file = converted_folder / "api.json"
    sdk_file = converted_folder / "sdk.json"

    shutil.copyfile(src_file, dst=api_file)
    shutil.copyfile(src_file, dst=sdk_file)

    processor = Processer(src_file)
    processor.add_sse_extensions()
    processor.update_model_name()

    processor.output(api_file)

    processor.convert_inline_enums_to_refs()
    processor.output(sdk_file, filter_sdk_apis)


if __name__ == "__main__":
    main()
