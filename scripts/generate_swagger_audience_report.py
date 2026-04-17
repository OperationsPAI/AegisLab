#!/usr/bin/env python3
from __future__ import annotations

import json
import re
from collections import Counter
from dataclasses import dataclass
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SOURCE_ROOT = REPO_ROOT / "src" / "module"
OUTPUT_FILE = REPO_ROOT / "docs" / "swagger-audience-marking-report.md"

ROUTER_RE = re.compile(r"@Router\s+(\S+)\s+\[(\w+)\]")
SUMMARY_RE = re.compile(r"@Summary\s+(.+)")
X_API_TYPE_RE = re.compile(r"@x-api-type\s+(.+)")
FUNC_RE = re.compile(r"^func\s+([A-Za-z0-9_]+)\s*\(")


@dataclass
class Operation:
    method: str
    path: str
    summary: str
    file_path: str
    router_line: int
    x_api_type_line: int | None
    function_name: str | None
    function_line: int | None
    audiences: list[str]
    raw_x_api_type: str | None
    status: str

    @property
    def source(self) -> str:
        parts = [f"`{self.file_path}:{self.router_line}`"]
        if self.x_api_type_line is not None:
            parts.append(f"`{self.file_path}:{self.x_api_type_line}`")
        if self.function_line is not None:
            parts.append(f"`{self.file_path}:{self.function_line}`")
        return " / ".join(parts)

    @property
    def audience_text(self) -> str:
        return ", ".join(self.audiences) if self.audiences else "-"


def parse_x_api_type(raw_value: str | None) -> list[str]:
    if raw_value is None:
        return []

    raw_value = raw_value.strip()
    try:
        parsed = json.loads(raw_value)
    except json.JSONDecodeError:
        return []

    if not isinstance(parsed, dict):
        return []

    audiences: list[str] = []
    for key in ("sdk", "portal", "admin"):
        value = parsed.get(key)
        if isinstance(value, bool) and value:
            audiences.append(key)
        elif isinstance(value, str) and value.strip().lower() == "true":
            audiences.append(key)
    return audiences


def iter_handler_files() -> list[Path]:
    files: list[Path] = []
    files.extend(
        path
        for path in SOURCE_ROOT.rglob("*.go")
        if path.is_file() and not path.name.endswith("_test.go")
    )
    return sorted(files)


def collect_operations() -> list[Operation]:
    operations: list[Operation] = []

    for file_path in iter_handler_files():
        rel_path = file_path.relative_to(REPO_ROOT).as_posix()
        lines = file_path.read_text(encoding="utf-8").splitlines()

        for index, line in enumerate(lines):
            router_match = ROUTER_RE.search(line)
            if not router_match:
                continue

            path = router_match.group(1)
            method = router_match.group(2).upper()

            summary = ""
            raw_x_api_type: str | None = None
            x_api_type_line: int | None = None

            # Search around the router annotation inside the current comment block.
            start = max(0, index - 40)
            end = min(len(lines), index + 12)
            for scan_index in range(index, end):
                x_match = X_API_TYPE_RE.search(lines[scan_index])
                if x_match:
                    raw_x_api_type = x_match.group(1).strip()
                    x_api_type_line = scan_index + 1
                    break

            for scan_index in range(index, start - 1, -1):
                summary_match = SUMMARY_RE.search(lines[scan_index])
                if summary_match:
                    summary = summary_match.group(1).strip()
                    break
                if (
                    scan_index != index
                    and lines[scan_index].strip()
                    and not lines[scan_index].lstrip().startswith("//")
                ):
                    break

            function_name: str | None = None
            function_line: int | None = None
            for scan_index in range(index + 1, min(len(lines), index + 20)):
                func_match = FUNC_RE.match(lines[scan_index].strip())
                if func_match:
                    function_name = func_match.group(1)
                    function_line = scan_index + 1
                    break

            audiences = parse_x_api_type(raw_x_api_type)
            if audiences:
                status = "marked"
            elif raw_x_api_type is None:
                status = "missing"
            else:
                status = "empty"

            operations.append(
                Operation(
                    method=method,
                    path=path,
                    summary=summary,
                    file_path=rel_path,
                    router_line=index + 1,
                    x_api_type_line=x_api_type_line,
                    function_name=function_name,
                    function_line=function_line,
                    audiences=audiences,
                    raw_x_api_type=raw_x_api_type,
                    status=status,
                )
            )

    operations.sort(
        key=lambda item: (item.path, item.method, item.file_path, item.router_line)
    )
    return operations


def markdown_table(rows: list[list[str]]) -> list[str]:
    if not rows:
        return ["_None_"]

    header = rows[0]
    lines = [
        "| " + " | ".join(header) + " |",
        "| " + " | ".join(["---"] * len(header)) + " |",
    ]
    for row in rows[1:]:
        lines.append("| " + " | ".join(row) + " |")
    return lines


def build_report(operations: list[Operation]) -> str:
    marked = [item for item in operations if item.status == "marked"]
    empty = [item for item in operations if item.status == "empty"]
    missing = [item for item in operations if item.status == "missing"]

    audience_counter: Counter[str] = Counter()
    for item in marked:
        audience_counter.update(item.audiences)

    lines: list[str] = []
    lines.append("# Swagger Audience Marking Report")
    lines.append("")
    lines.append("> Source of truth: Swagger annotations in `src/module/**/*.go`.")
    lines.append(
        "> Route position column uses the `@Router` line, then `@x-api-type`, then function line when available."
    )
    lines.append("")
    lines.append("## Summary")
    lines.append("")
    lines.append(f"- Total operations scanned: **{len(operations)}**")
    lines.append(f"- Marked operations: **{len(marked)}**")
    lines.append(f"- Empty `@x-api-type {{}}` operations: **{len(empty)}**")
    lines.append(f"- Missing `@x-api-type` operations: **{len(missing)}**")
    lines.append(
        "- Audience counts among marked operations: "
        f"`sdk={audience_counter.get('sdk', 0)}` "
        f"`portal={audience_counter.get('portal', 0)}` "
        f"`admin={audience_counter.get('admin', 0)}`"
    )
    lines.append("")

    lines.append("## Marked Operations")
    lines.append("")
    lines.extend(
        markdown_table(
            [
                ["Method", "Path", "Audience", "Summary", "Location"],
                *[
                    [
                        item.method,
                        f"`{item.path}`",
                        f"`{item.audience_text}`",
                        item.summary or "-",
                        item.source,
                    ]
                    for item in marked
                ],
            ]
        )
    )
    lines.append("")

    lines.append("## Empty `@x-api-type {}` Operations")
    lines.append("")
    lines.extend(
        markdown_table(
            [
                ["Method", "Path", "Summary", "Raw", "Location"],
                *[
                    [
                        item.method,
                        f"`{item.path}`",
                        item.summary or "-",
                        f"`{item.raw_x_api_type or ''}`",
                        item.source,
                    ]
                    for item in empty
                ],
            ]
        )
    )
    lines.append("")

    lines.append("## Missing `@x-api-type` Operations")
    lines.append("")
    lines.extend(
        markdown_table(
            [
                ["Method", "Path", "Summary", "Location"],
                *[
                    [
                        item.method,
                        f"`{item.path}`",
                        item.summary or "-",
                        item.source,
                    ]
                    for item in missing
                ],
            ]
        )
    )
    lines.append("")

    return "\n".join(lines) + "\n"


def main() -> None:
    operations = collect_operations()
    OUTPUT_FILE.write_text(build_report(operations), encoding="utf-8")
    print(
        f"generated {OUTPUT_FILE.relative_to(REPO_ROOT)} with {len(operations)} operations"
    )


if __name__ == "__main__":
    main()
