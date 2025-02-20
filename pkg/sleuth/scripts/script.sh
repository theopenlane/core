#!/bin/bash

required_tools=("gau" "katana" "waymore" "gf" "httpx" "nuclei")
for tool in "${required_tools[@]}"; do
  command -v "$tool" || { echo "❌ $tool missing!"; exit 1; }
done

# Input Validation and Target Parsing
if [ -z "$1" ]; then
  echo "❌ Error: No target provided!"
  exit 1
fi

TARGET="$1"
OUTPUT_DIR=""

# Function to process a single domain or subdomain
process_single_target() {
  TARGET_NAME=$(echo "$1" | sed 's|^https\?://||' | sed 's|/$||') # Remove http/https and trailing slash
  OUTPUT_DIR="results/$TARGET_NAME-$(date +%s)"
  mkdir -p "$OUTPUT_DIR"
  echo "$TARGET_NAME" > "$OUTPUT_DIR/target.txt"
}

# Function to process a list of domains/subdomains from a file
process_file_target() {
  if [ ! -f "$1" ]; then
    echo "❌ Error: File '$1' not found!"
    exit 1
  fi
  TARGET_NAME=$(basename "$1" .txt) # Extract filename without extension
  OUTPUT_DIR="results/$TARGET_NAME-$(date +%s)"
  mkdir -p "$OUTPUT_DIR"
  cp "$1" "$OUTPUT_DIR/target.txt"
}

# Determine the type of input (single target or file)
if [[ "$TARGET" == *.txt ]]; then
  process_file_target "$TARGET"
else
  process_single_target "$TARGET"
fi

# 🚀 Optimized Discovery
run_discovery() {
  echo "🔍 Starting URL Discovery..."

  # Read targets from target.txt
  while read -r line; do
    # gau
    gau --threads 20 --blacklist woff,css,png,svg,jpg,woff2,jpeg,gif,svg --subs --providers wayback,urlscan,otx "$line" 2>/dev/null | anew "$OUTPUT_DIR/gau.txt" > /dev/null

    # katana
    timeout 600 katana -u "https://$line" -d 3 -kf all -c 15 -silent -duc -nc -jc -fx -xhr -ef woff,css,png,svg,jpg,woff2,jpeg,gif,svg 2>/dev/null | anew "$OUTPUT_DIR/katana.txt" > /dev/null

    # waymore
    waymore -i "$line" -mode U -ft "font/woff,font/woff2,text/css,image/png,image/svg+xml,image/jpeg,image/gif" -oU "$OUTPUT_DIR/waymore.txt" 2>/dev/null > /dev/null
  done < "$OUTPUT_DIR/target.txt"

  echo "✅ URL Discovery Complete."
}

# 🛠️ URL Processing
process_urls() {
  echo "🔄 Starting URL Processing..."

  # merge and deduplicate
  cat "$OUTPUT_DIR"/{gau,katana,waymore}.txt | sort | uniq > "$OUTPUT_DIR/all_unique_urls.txt"

  # filter
  cat "$OUTPUT_DIR/all_unique_urls.txt" \
    | grep -E '^(https?|ftp|file)://' \
    | perl -MURI::Escape -ne 'chomp; print uri_unescape($_), "\n"' \
    | httpx -silent -timeout 8 -threads 100 -retries 2 > "$OUTPUT_DIR/valid_urls.txt"

  echo "✅ URL Processing Complete."
}

# 🔥 Vulnerability Detection
detect_vulns() {
  echo "⚠️ Starting Vulnerability Analysis..."

  # pattern class
  gf_patterns=("xss" "sqli" "lfi" "ssrf" "redirect")
  for pattern in "${gf_patterns[@]}"; do
    gf "$pattern" "$OUTPUT_DIR/valid_urls.txt" > "$OUTPUT_DIR/${pattern}_urls.txt" 2>/dev/null
  done

  # nuclei
  nuclei -update-templates > /dev/null 2>&1
  nuclei -l "$OUTPUT_DIR/valid_urls.txt" \
    -tags "sqli,xss,lfi,ssrf" \
    -severity critical,high,medium \
    -rate-limit 200 \
    -concurrency 100 \
    -o "$OUTPUT_DIR/nuclei_results.txt" > /dev/null 2>&1

  # sup boiiii
  if [ -s "$OUTPUT_DIR/nuclei_results.txt" ]; then
    vuln_count=$(wc -l < "$OUTPUT_DIR/nuclei_results.txt")

    vulnerabilities=$(awk -F' \[' '{print "- URL: " $1 "\n  Template: " $2}' "$OUTPUT_DIR/nuclei_results.txt" | sed 's/\]//g')
  fi

  echo "✅ Vulnerability Analysis Complete."
}

{
  run_discovery
  process_urls
  detect_vulns
} 2>&1 | tee "$OUTPUT_DIR/scan.log"
echo "✅ Final Output: $OUTPUT_DIR"