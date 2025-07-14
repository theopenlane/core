#!/bin/bash
echo "=== Simple Test Script ==="
echo "Current directory: $(pwd)"
echo "Listing .buildkite directory:"
ls -la .buildkite/
echo "Script exists check:"
if [[ -f ".buildkite/draft-pr-automation.sh" ]]; then
    echo "✅ draft-pr-automation.sh found"
    echo "File permissions: $(ls -la .buildkite/draft-pr-automation.sh)"
else
    echo "❌ draft-pr-automation.sh NOT found"
fi
echo "=== End Test ==="