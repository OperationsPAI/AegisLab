#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PAGE_REL="content/concepts/fault-injection-lifecycle/_index.md"
CONCEPTS_INDEX_REL="content/concepts/_index.md"
ARCH_INDEX_REL="content/concepts/architecture/_index.md"

PASS_COUNT=0
FAIL_COUNT=0
SKIP_COUNT=0

pass() {
    echo "PASS: $1"
    PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
    echo "FAIL: $1"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

skip() {
    echo "SKIP: $1"
    SKIP_COUNT=$((SKIP_COUNT + 1))
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

extract_markdown_target() {
    local file_abs="$1"
    local text="$2"
    grep -oE "\[$text\]\(([^)]+)\)" "$file_abs" | head -n1 | sed -E 's/.*\(([^)]+)\)/\1/'
}

extract_card_target() {
    local file_abs="$1"
    local title="$2"
    grep "title=\"$title\"" "$file_abs" | head -n1 | sed -nE 's/.*link="([^"]+)".*/\1/p'
}

check_page_quality() {
    local page_abs="$1"

    if grep -q '^title: Fault Injection Lifecycle$' "$page_abs"; then
        pass "Front matter title is correct"
    else
        fail "Front matter title missing or incorrect"
    fi

    local word_count
    word_count=$(wc -w < "$page_abs")
    if [[ "$word_count" -ge 200 ]]; then
        pass "Page content length >= 200 words ($word_count)"
    else
        fail "Page content length < 200 words ($word_count)"
    fi

    local heading_count
    heading_count=$(grep -c '^## ' "$page_abs" || true)
    if [[ "$heading_count" -ge 3 ]]; then
        pass "Page has structured headings (## count: $heading_count)"
    else
        fail "Insufficient section headings (## count: $heading_count)"
    fi

    if grep -qi 'CRD' "$page_abs"; then
        pass "Mentions CRD success callback context"
    else
        fail "Missing CRD-related explanation"
    fi

    if grep -qi 'BuildDatapack' "$page_abs"; then
        pass "Mentions BuildDatapack child task"
    else
        fail "Missing BuildDatapack explanation"
    fi

    if grep -qi 'Hybrid' "$page_abs"; then
        pass "Mentions Hybrid batch completion gate"
    else
        fail "Missing Hybrid completion explanation"
    fi

    if grep -Eqi 'state update|state transition|task state|状态更新' "$page_abs"; then
        pass "Mentions task/injection state update"
    else
        fail "Missing state update explanation"
    fi
}

check_hugo_if_possible() {
    local cfg_found=0
    local cfg
    for cfg in hugo.toml hugo.yaml hugo.yml config.toml config.yaml config.yml; do
        if [[ -f "$ROOT_DIR/$cfg" ]]; then
            cfg_found=1
            break
        fi
    done

    if [[ "$cfg_found" -eq 0 ]]; then
        skip "No Hugo site config found in repo root; skip clean-build check"
        return
    fi

    if command -v hugo >/dev/null 2>&1; then
        if (cd "$ROOT_DIR" && hugo >/tmp/hugo-task8-check.log 2>&1); then
            pass "Hugo build succeeded"
        else
            fail "Hugo build failed (see /tmp/hugo-task8-check.log)"
        fi
    else
        skip "hugo command not found; skip clean-build check"
    fi
}

main() {
    local page_abs="$ROOT_DIR/$PAGE_REL"
    local concepts_abs="$ROOT_DIR/$CONCEPTS_INDEX_REL"
    local arch_abs="$ROOT_DIR/$ARCH_INDEX_REL"

    if [[ -f "$page_abs" ]]; then
        pass "Page exists: $PAGE_REL"
    else
        fail "Missing page: $PAGE_REL"
    fi

    if [[ -f "$concepts_abs" ]]; then
        pass "Concepts index exists"
    else
        fail "Missing concepts index: $CONCEPTS_INDEX_REL"
    fi

    if [[ -f "$arch_abs" ]]; then
        pass "Architecture index exists"
    else
        fail "Missing architecture index: $ARCH_INDEX_REL"
    fi

    if [[ -f "$page_abs" ]]; then
        check_page_quality "$page_abs"
    fi

    if [[ -f "$concepts_abs" ]]; then
        local card_target
        card_target="$(extract_card_target "$concepts_abs" "Fault Injection Lifecycle")"
        if [[ -n "$card_target" ]]; then
            resolve_and_check "$CONCEPTS_INDEX_REL" "$card_target" "Concepts card link"
        else
            fail "Fault Injection Lifecycle card link missing in $CONCEPTS_INDEX_REL"
        fi

        local md_target
        md_target="$(extract_markdown_target "$concepts_abs" "Fault Injection Lifecycle")"
        if [[ -n "$md_target" ]]; then
            resolve_and_check "$CONCEPTS_INDEX_REL" "$md_target" "Concepts markdown link"
        else
            fail "Fault Injection Lifecycle markdown link missing in $CONCEPTS_INDEX_REL"
        fi
    fi

    if [[ -f "$arch_abs" ]]; then
        local arch_target
        arch_target="$(extract_markdown_target "$arch_abs" "Fault Injection Lifecycle")"
        if [[ -n "$arch_target" ]]; then
            resolve_and_check "$ARCH_INDEX_REL" "$arch_target" "Architecture markdown link"
        else
            fail "Fault Injection Lifecycle markdown link missing in $ARCH_INDEX_REL"
        fi
    fi

    check_hugo_if_possible

    echo ""
    echo "Summary: PASS=$PASS_COUNT FAIL=$FAIL_COUNT SKIP=$SKIP_COUNT"

    if [[ "$FAIL_COUNT" -gt 0 ]]; then
        exit 1
    fi
}

main "$@"