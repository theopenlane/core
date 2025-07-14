#!/bin/bash
set -euo pipefail

echo "=== Testing Docker Script Execution ==="
echo "This script tests how the build-image container handles script execution"

# Test 1: Check if the script file exists in the container
echo ""
echo "Test 1: Checking if draft-pr-automation.sh exists in container..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  ls -la .buildkite/draft-pr-automation.sh

# Test 2: Check file permissions
echo ""
echo "Test 2: Checking file permissions..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  stat .buildkite/draft-pr-automation.sh

# Test 3: Try to make it executable and check
echo ""
echo "Test 3: Making executable and checking..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  sh -c "chmod +x .buildkite/draft-pr-automation.sh && ls -la .buildkite/draft-pr-automation.sh"

# Test 4: Test the exact command format from Buildkite
echo ""
echo "Test 4: Testing exact command format from Buildkite..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  /bin/sh -e -c "chmod +x .buildkite/draft-pr-automation.sh && ./.buildkite/draft-pr-automation.sh --help"

# Test 5: Check what shell is available
echo ""
echo "Test 5: Checking available shells..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  sh -c "which sh; which bash; ls -la /bin/*sh*"

# Test 6: Check if script has proper shebang and is readable
echo ""
echo "Test 6: Checking script content..."
docker run --rm -v "$(pwd)":/workdir --workdir /workdir \
  ghcr.io/theopenlane/build-image:latest \
  head -5 .buildkite/draft-pr-automation.sh

echo ""
echo "=== Test Complete ==="