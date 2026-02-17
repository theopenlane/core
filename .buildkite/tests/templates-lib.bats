#!/usr/bin/env bats

@test "load_template converts escaped newline sequences in substitutions" {
  template_file="$BATS_TEST_TMPDIR/template.md"
  cat > "$template_file" <<'EOF'
## Changes
{{CHANGE_SUMMARY}}
EOF

  source .buildkite/lib/templates.sh
  rendered=$(load_template "$template_file" "CHANGE_SUMMARY=\\n- first\\n- second")

  [[ "$rendered" == *$'\n- first\n- second'* ]]
  [[ "$rendered" != *"\\n- first\\n- second"* ]]
}
