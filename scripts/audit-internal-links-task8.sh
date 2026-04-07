#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_REL="${1:-docs/task8-internal-link-audit.md}"
REPORT_ABS="$ROOT_DIR/$REPORT_REL"

TMP_ROWS="$(mktemp)"
trap 'rm -f "$TMP_ROWS"' EXIT

record_link() {
    local src_rel="$1"
    local line_no="$2"
    local kind="$3"
    local target_raw="$4"

    local target="${target_raw%%#*}"
    target="${target%%\?*}"

    if [[ -z "$target" ]]; then
        return 0
    fi
    if [[ "$target" =~ ^(http://|https://|mailto:|tel:|javascript:|data:|#) ]]; then
        return 0
    fi

    local src_abs="$ROOT_DIR/$src_rel"
    local src_dir
    src_dir="$(dirname "$src_abs")"

    local abs
    if [[ "$target" == /* ]]; then
        abs="$ROOT_DIR$target"
    else
        abs="$src_dir/$target"
    fi

    local no_slash="${abs%/}"
    local status="broken"
    local resolved="-"
    local c
    for c in \
        "$abs" \
        "$no_slash" \
        "$no_slash.md" \
        "$no_slash/_index.md" \
        "$no_slash/index.md"; do
        if [[ -e "$c" ]]; then
            status="working"
            resolved="${c#"$ROOT_DIR"/}"
            break
        fi
    done

    printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
        "$src_rel" "$line_no" "$kind" "$target_raw" "$status" "$resolved" >> "$TMP_ROWS"
}

collect_links() {
    local src_rel="$1"
    local src_abs="$ROOT_DIR/$src_rel"

    while IFS= read -r m; do
        local line_no="${m%%:*}"
        local body="${m#*:}"
        local target
        target="$(printf '%s' "$body" | sed -E 's/.*\]\(([^)]+)\).*/\1/')"
        record_link "$src_rel" "$line_no" "markdown" "$target"
    done < <(grep -nE '\[[^]]+\]\(([^)]+)\)' "$src_abs" || true)

    while IFS= read -r m; do
        local line_no="${m%%:*}"
        local body="${m#*:}"
        local target
        target="$(printf '%s' "$body" | sed -nE 's/.*link="([^"]+)".*/\1/p')"
        [[ -n "$target" ]] && record_link "$src_rel" "$line_no" "hugo-card" "$target"
    done < <(grep -nE 'link="[^"]+"' "$src_abs" || true)
}

main() {
    mkdir -p "$(dirname "$REPORT_ABS")"

    local files=()
    [[ -f "$ROOT_DIR/README.md" ]] && files+=("README.md")
    if [[ -d "$ROOT_DIR/content" ]]; then
        while IFS= read -r f; do files+=("$f"); done < <(cd "$ROOT_DIR" && find content -type f -name '*.md' | sort)
    fi
    if [[ -d "$ROOT_DIR/docs" ]]; then
        while IFS= read -r f; do files+=("$f"); done < <(cd "$ROOT_DIR" && find docs -type f -name '*.md' | sort)
    fi
    [[ -f "$ROOT_DIR/sdk/python/README.md" ]] && files+=("sdk/python/README.md")

    local uniq
    uniq="$(printf '%s\n' "${files[@]}" | sed '/^$/d' | sort -u)"
    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        collect_links "$f"
    done <<< "$uniq"

    local total=0
    local working=0
    local broken=0
    if [[ -s "$TMP_ROWS" ]]; then
        total="$(wc -l < "$TMP_ROWS")"
        working="$(grep -c $'\tworking\t' "$TMP_ROWS" || true)"
        broken="$(grep -c $'\tbroken\t' "$TMP_ROWS" || true)"
    fi

    {
        echo "# Task8 Internal Link Audit"
        echo
        echo "Generated at: $(date -Iseconds)"
        echo
        echo "Scope: README.md, content/**/*.md, docs/**/*.md, sdk/python/README.md"
        echo
        echo "Summary: total=$total, working=$working, broken=$broken"
        echo
        echo "| File | Line | Kind | Target | Status | Resolved Path |"
        echo "| --- | ---: | --- | --- | --- | --- |"
        if [[ -s "$TMP_ROWS" ]]; then
            sort -t $'\t' -k1,1 -k2,2n -k3,3 "$TMP_ROWS" | while IFS=$'\t' read -r file line kind target status resolved; do
                echo "| $file | $line | $kind | $target | $status | $resolved |"
            done
        fi
    } > "$REPORT_ABS"

    echo "Report generated: $REPORT_REL"
    echo "Summary: total=$total, working=$working, broken=$broken"

    [[ "$broken" -eq 0 ]]
}

main "$@"#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_REPORT="docs/task8-internal-link-audit.md"
REPORT_REL="${1:-$DEFAULT_REPORT}"
REPORT_ABS="$ROOT_DIR/$REPORT_REL"

TMP_ROWS="$(mktemp)"
trap 'rm -f "$TMP_ROWS"' EXIT

record_link() {
    local src_rel="$1"
    local line_no="$2"
    local kind="$3"
    local target_raw="$4"

    local target="${target_raw%%#*}"
    target="${target%%\?*}"

    # Internal links only
    if [[ -z "$target" ]]; then
        return 0
    fi
    if [[ "$target" =~ ^(http://|https://|mailto:|tel:|javascript:|data:|#) ]]; then
        return 0
    fi

    local src_abs="$ROOT_DIR/$src_rel"
    local src_dir
    src_dir="$(dirname "$src_abs")"

    local abs
    if [[ "$target" == /* ]]; then
        abs="$ROOT_DIR$target"
    else
        abs="$src_dir/$target"
    fi

    local no_slash="${abs%/}"
    local status="broken"
    local resolved="-"
    local candidate
    for candidate in \
        "$abs" \
        "$no_slash" \
        "$no_slash.md" \
        "$no_slash/_index.md" \
        "$no_slash/index.md"; do
        if [[ -e "$candidate" ]]; then
            status="working"
            resolved="${candidate#"$ROOT_DIR"/}"
            break
        fi
    done

    printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
        "$src_rel" "$line_no" "$kind" "$target_raw" "$status" "$resolved" >> "$TMP_ROWS"
}

collect_links_from_file() {
    local src_rel="$1"
    local src_abs="$ROOT_DIR/$src_rel"

    # Markdown links: [text](target)
    while IFS= read -r m; do
        local line_no="${m%%:*}"
        local body="${m#*:}"
        local target
        target="$(printf '%s' "$body" | sed -E 's/.*\]\(([^)]+)\).*/\1/')"
        record_link "$src_rel" "$line_no" "markdown" "$target"
    done < <(grep -nE '\[[^]]+\]\(([^)]+)\)' "$src_abs" || true)

    # Hugo cards: link="target"
    while IFS= read -r m; do
        local line_no="${m%%:*}"
        local body="${m#*:}"
        local target
        target="$(printf '%s' "$body" | sed -nE 's/.*link="([^"]+)".*/\1/p')"
        if [[ -n "$target" ]]; then
            record_link "$src_rel" "$line_no" "hugo-card" "$target"
        fi
    done < <(grep -nE 'link="[^"]+"' "$src_abs" || true)
}

main() {
    mkdir -p "$(dirname "$REPORT_ABS")"

    local files=()
    if [[ -f "$ROOT_DIR/README.md" ]]; then
        files+=("README.md")
    fi
    if [[ -d "$ROOT_DIR/content" ]]; then
        while IFS= read -r f; do files+=("$f"); done < <(cd "$ROOT_DIR" && find content -type f -name '*.md' | sort)
    fi
    if [[ -d "$ROOT_DIR/docs" ]]; then
        while IFS= read -r f; do files+=("$f"); done < <(cd "$ROOT_DIR" && find docs -type f -name '*.md' | sort)
    fi
    if [[ -f "$ROOT_DIR/sdk/python/README.md" ]]; then
        files+=("sdk/python/README.md")
    fi

    # Unique file list while keeping deterministic order
    local uniq_files
    uniq_files="$(printf '%s\n' "${files[@]}" | sed '/^$/d' | sort -u)"

    while IFS= read -r f; do
        [[ -z "$f" ]] && continue
        collect_links_from_file "$f"
    done <<< "$uniq_files"

    local total=0
    local working=0
    local broken=0

    if [[ -s "$TMP_ROWS" ]]; then
        total="$(wc -l < "$TMP_ROWS")"
        working="$(grep -c $'\tworking\t' "$TMP_ROWS" || true)"
        broken="$(grep -c $'\tbroken\t' "$TMP_ROWS" || true)"
    fi

    {
        echo "# Task8 Internal Link Audit"
        echo
        echo "Generated at: $(date -Iseconds)"
        echo
        echo "Scope: README.md, content/**/*.md, docs/**/*.md, sdk/python/README.md"
        echo
        echo "Summary: total=$total, working=$working, broken=$broken"
        echo
        echo "| File | Line | Kind | Target | Status | Resolved Path |"
        echo "| --- | ---: | --- | --- | --- | --- |"

        if [[ -s "$TMP_ROWS" ]]; then
            sort -t $'\t' -k1,1 -k2,2n -k3,3 "$TMP_ROWS" | while IFS=$'\t' read -r file line kind target status resolved; do
                echo "| $file | $line | $kind | $target | $status | $resolved |"
            done
        fi
    } > "$REPORT_ABS"

    echo "Report generated: $REPORT_REL"
    echo "Summary: total=$total, working=$working, broken=$broken"

    if [[ "$broken" -gt 0 ]]; then
        exit 1
    fi
}

main "$@"