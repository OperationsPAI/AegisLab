#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CONCEPTS_INDEX="content/concepts/_index.md"
ARCH_INDEX="content/concepts/architecture/_index.md"

PASS_COUNT=0
FAIL_COUNT=0

pass() {
    echo "PASS: $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo "FAIL: $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

resolve_and_check() {
    local src_rel="$1"
    local target_raw="$2"
    local label="$3"

    local src_abs="$ROOT_DIR/$src_rel"
    local src_dir
    src_dir="$(dirname "$src_abs")"

    local target="${target_raw%%#*}"
    target="${target%%\?*}"

    local abs
    if [[ "$target" == /* ]]; then
        abs="$ROOT_DIR${target}"
    else
        abs="$src_dir/$target"
    fi

    local no_slash="${abs%/}"
    local candidates=(
        "$abs"
        "$no_slash"
        "$no_slash.md"
        "$no_slash/_index.md"
        "$no_slash/index.md"
    )

    local hit=""
    for c in "${candidates[@]}"; do
        if [[ -e "$c" ]]; then
            hit="$c"
            break
        fi
    done

    if [[ -n "$hit" ]]; then
        local rel_hit="${hit#"$ROOT_DIR"/}"
        pass "$label -> $target_raw (resolved: $rel_hit)"
    else
        fail "$label -> $target_raw (unreachable from $src_rel)"
    fi
}

extract_data_flow_markdown_target() {
    local file_abs="$1"
    grep -oE '\[Data Flow\]\(([^)]+)\)' "$file_abs" | head -n1 | sed -E 's/.*\(([^)]+)\)/\1/'
}

extract_data_flow_card_target() {
    local file_abs="$1"
    grep 'title="Data Flow"' "$file_abs" | head -n1 | sed -nE 's/.*link="([^"]+)".*/\1/p'
}

main() {
    if [[ ! -f "$ROOT_DIR/$CONCEPTS_INDEX" ]]; then
        echo "ERROR: missing $CONCEPTS_INDEX"
        exit 2
    fi
    if [[ ! -f "$ROOT_DIR/$ARCH_INDEX" ]]; then
        echo "ERROR: missing $ARCH_INDEX"
        exit 2
    fi

    local concepts_abs="$ROOT_DIR/$CONCEPTS_INDEX"
    local arch_abs="$ROOT_DIR/$ARCH_INDEX"

    local card_target
    card_target="$(extract_data_flow_card_target "$concepts_abs")"
    if [[ -z "$card_target" ]]; then
        fail "Data Flow card link not found in $CONCEPTS_INDEX"
    else
        resolve_and_check "$CONCEPTS_INDEX" "$card_target" "Concepts card"
    fi

    local concepts_md_target
    concepts_md_target="$(extract_data_flow_markdown_target "$concepts_abs")"
    if [[ -z "$concepts_md_target" ]]; then
        fail "Data Flow markdown link not found in $CONCEPTS_INDEX"
    else
        resolve_and_check "$CONCEPTS_INDEX" "$concepts_md_target" "Concepts markdown link"
    fi

    local arch_md_target
    arch_md_target="$(extract_data_flow_markdown_target "$arch_abs")"
    if [[ -z "$arch_md_target" ]]; then
        fail "Data Flow markdown link not found in $ARCH_INDEX"
    else
        resolve_and_check "$ARCH_INDEX" "$arch_md_target" "Architecture markdown link"
    fi

    echo ""
    echo "Summary: PASS=$PASS_COUNT FAIL=$FAIL_COUNT"

    if [[ $FAIL_COUNT -gt 0 ]]; then
        exit 1
    fi
}

main "$@"