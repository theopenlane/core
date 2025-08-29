#!/usr/bin/env bash
set -euo pipefail

load_template() {
    local template_file="$1"
    shift

    if [[ ! -f "$template_file" ]]; then
        echo "⚠️  Template file not found: $template_file" >&2
        return 1
    fi

    local content
    content=$(<"$template_file")   # faster than cat

    for arg in "$@"; do
        if [[ "$arg" == *"="* ]]; then
            key="${arg%%=*}"
            value="${arg#*=}"
            content="${content//\{\{${key}\}\}/$value}"
        fi
    done

    echo "$content"
}

# --- Test cases ---
mkdir -p tmp
cat > tmp/test.md <<'EOF'
Release: {{TAG}}

Summary:
{{SUMMARY}}

Commit: {{COMMIT}}
EOF

echo "----- Substitution test -----"
load_template tmp/test.md \
  "TAG=v1.2.3" \
  "SUMMARY=Line one
Line two
Line three" \
  "COMMIT=abc123"

