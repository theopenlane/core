#!/bin/bash

# Script to test 'get' command for all available CLI commands
# This script runs './openlane-cli <command> get' for every command that supports it

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log file for results
LOG_FILE="testget_results.log"
ERROR_LOG="testget_errors.log"

# Initialize log files
echo "Test run started at $(date)" > "$LOG_FILE"
echo "Error log started at $(date)" > "$ERROR_LOG"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

# Function to test a command
test_command() {
    local cmd=$1
    print_status $BLUE "Testing: ./openlane-cli $cmd get"

    # Run the command and capture output
    if output=$(./openlane-cli "$cmd" get 2>&1); then
        print_status $GREEN "✓ SUCCESS: ./openlane-cli $cmd get"
        echo "Output preview: $(echo "$output" | head -n 3 | tr '\n' ' ')" >> "$LOG_FILE"
        echo "" >> "$LOG_FILE"
        return 0  # Success
    else
        exit_code=$?
        print_status $RED "✗ FAILED: ./openlane-cli $cmd get (exit code: $exit_code)"
        echo "FAILED: ./openlane-cli $cmd get (exit code: $exit_code)" >> "$ERROR_LOG"
        echo "Error output: $output" >> "$ERROR_LOG"
        echo "---" >> "$ERROR_LOG"
        echo "" >> "$LOG_FILE"
        return 1  # Failure
    fi
}

# Function to parse commands from ./openlane-cli --help output
get_available_commands() {
    local help_output
    local raw_commands
    local filtered_commands=()

    # Get help output and extract command names (no print_status here to avoid output pollution)
    if ! help_output=$(./openlane-cli --help 2>&1); then
        echo "ERROR: Failed to get help output" >&2
        exit 1
    fi

    # Extract commands from "Available Commands:" section
    # This parses lines that start with 2+ spaces followed by a command name
    raw_commands=$(echo "$help_output" | sed -n '/^Available Commands:/,/^$/p' | \
                   grep -E '^\s{2,}[a-z]' | \
                   awk '{print $1}' | \
                   grep -v '^$')

    # Define commands to exclude (utility commands that don't typically have 'get')
    local exclude_commands=(
        "completion"
        "help"
        "login"
        "reconcile"
        "register"
        "reset"
        "search"
        "switch"
        "version"
    )

    # Filter out excluded commands
    while IFS= read -r cmd; do
        local should_exclude=false
        for exclude_cmd in "${exclude_commands[@]}"; do
            if [[ "$cmd" == "$exclude_cmd" ]]; then
                should_exclude=true
                break
            fi
        done

        if [[ "$should_exclude" == false && -n "$cmd" ]]; then
            filtered_commands+=("$cmd")
        fi
    done <<< "$raw_commands"

    # Return the filtered commands array (only output the commands, no status messages)
    printf '%s\n' "${filtered_commands[@]}"
}

# Function to check if a command supports 'get'
command_supports_get() {
    local cmd=$1
    local cmd_help

    # Get command help and check if 'get' subcommand exists
    if cmd_help=$(./openlane-cli "$cmd" --help 2>&1); then
        if echo "$cmd_help" | grep -q "^\s*get\s"; then
            return 0  # Command supports get
        fi
    fi
    return 1  # Command doesn't support get
}

# Get available commands dynamically
print_status $YELLOW "Discovering available commands..."
print_status $BLUE "Getting available commands from './openlane-cli --help'..."
all_commands=()
while IFS= read -r cmd; do
    all_commands+=("$cmd")
done < <(get_available_commands)

if [[ ${#all_commands[@]} -eq 0 ]]; then
    print_status $RED "No commands found! Check if './openlane-cli' is available and working."
    exit 1
fi

print_status $GREEN "Found ${#all_commands[@]} potential commands"

# Filter commands that actually support 'get'
commands=()
print_status $BLUE "Checking which commands support 'get'..."

for cmd in "${all_commands[@]}"; do
    if command_supports_get "$cmd"; then
        commands+=("$cmd")
        print_status $GREEN "  ✓ $cmd supports get"
    else
        print_status $YELLOW "  - $cmd does not support get (skipping)"
    fi
done

if [[ ${#commands[@]} -eq 0 ]]; then
    print_status $RED "No commands support 'get' operation!"
    exit 1
fi

print_status $YELLOW "Starting get command tests for ${#commands[@]} commands that support 'get'..."
echo "Total commands to test: ${#commands[@]}" >> "$LOG_FILE"
echo "Commands to test: ${commands[*]}" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Counter for statistics
success_count=0
failure_count=0

# Test each command
for cmd in "${commands[@]}"; do
    test_command "$cmd"
    if [ $? -eq 0 ]; then
        ((success_count++))
    else
        ((failure_count++))
    fi
    # Small delay to avoid overwhelming the API
    sleep 0.5
done

# Print summary
echo ""
print_status $YELLOW "=== TEST SUMMARY ==="
print_status $GREEN "Successful commands: $success_count"
print_status $RED "Failed commands: $failure_count"
print_status $BLUE "Total commands tested: ${#commands[@]}"

echo "" >> "$LOG_FILE"
echo "=== FINAL SUMMARY ===" >> "$LOG_FILE"
echo "Successful: $success_count" >> "$LOG_FILE"
echo "Failed: $failure_count" >> "$LOG_FILE"
echo "Total: ${#commands[@]}" >> "$LOG_FILE"
echo "Test run completed at $(date)" >> "$LOG_FILE"

if [ $failure_count -eq 0 ]; then
    print_status $GREEN "🎉 All tests passed!"
    exit 0
else
    print_status $YELLOW "⚠️  Some tests failed. Check $ERROR_LOG for details."
    exit 1
fi