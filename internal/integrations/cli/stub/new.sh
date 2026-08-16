#!/usr/bin/env bash
# scaffolds a new integration definition package from the templates in this directory
#
#   ./new.sh oci "Oracle Cloud Infrastructure"
#   ./new.sh oci "Oracle Cloud Infrastructure" Oracle identity
#
# existing files are never overwritten, so re-running only fills in what is missing

set -euo pipefail

name=${1:-}
display=${2:-$name}
family=${3:-$display}
category=${4:-security-posture}

if [ -z "$name" ]; then
	echo "usage: $(basename "$0") <name> [display] [family] [category]" >&2
	exit 1
fi

src=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
dest="$src/../../definitions/$name"

# a ULID is 26 characters, so pad the mnemonic out to that width and end on a 1
mnemonic=$(printf '%s' "$name" | tr '[:lower:]' '[:upper:]' | tr -cd 'A-Z0-9' | cut -c1-20)
defid=$(printf 'def_01K0%s%0*d' "$mnemonic" $((22 - ${#mnemonic})) 1)

mkdir -p "$dest"

for tmpl in "$src"/*.go.tmpl; do
	out="$dest/$(basename "$tmpl" .tmpl)"

	if [ -e "$out" ]; then
		echo "----> skipping existing: $(basename "$out")"
		continue
	fi

	echo "----> generating: $(basename "$out")"

	sed -e "s|{{ name }}|$name|g" \
		-e "s|{{ display }}|$display|g" \
		-e "s|{{ family }}|$family|g" \
		-e "s|{{ category }}|$category|g" \
		-e "s|{{ slug }}|$name|g" \
		-e "s|{{ defid }}|$defid|g" \
		"$tmpl" >"$out"
done

gofmt -w "$dest"

echo
echo "definition id: $defid"
echo "next: add the import and ${name}.Builder() to internal/integrations/definitions/catalog/catalog.go"
