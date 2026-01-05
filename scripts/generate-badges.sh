#!/bin/bash
# Generate badge metrics for Shields.io dynamic badges
# Outputs JSON files to .github/badges/ for upload to GitHub Gist

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BADGES_DIR="$ROOT_DIR/.github/badges"

mkdir -p "$BADGES_DIR"

# Get version from VERSION file
VERSION=$(cat "$ROOT_DIR/VERSION" 2>/dev/null | tr -d '\n' || echo "0.0.0")
echo "Version: $VERSION"

# Count lines of Go code (excluding tests, vendor, generated)
LOC=$(find "$ROOT_DIR" -name '*.go' \
    ! -path '*/vendor/*' \
    ! -name '*_test.go' \
    ! -name '*.pb.go' \
    -exec cat {} \; 2>/dev/null | \
    grep -v '^\s*$' | \
    grep -v '^\s*//' | \
    wc -l | tr -d ' ')
echo "Lines of Code: $LOC"

# Format LOC with K suffix
if [ "$LOC" -ge 1000 ]; then
    LOC_FORMATTED="$(echo "scale=1; $LOC / 1000" | bc)k"
else
    LOC_FORMATTED="$LOC"
fi

# Run tests and get coverage
echo "Running tests..."
cd "$ROOT_DIR"
TEST_OUTPUT=$(go test -cover ./... 2>&1 || true)

# Count passing packages
PASS_COUNT=$(echo "$TEST_OUTPUT" | grep -c '^ok' 2>/dev/null || echo "0")
PASS_COUNT=$(echo "$PASS_COUNT" | tr -d '\n ')
FAIL_COUNT=$(echo "$TEST_OUTPUT" | grep -c '^FAIL' 2>/dev/null || echo "0")
FAIL_COUNT=$(echo "$FAIL_COUNT" | tr -d '\n ')
TOTAL_PACKAGES=$((PASS_COUNT + FAIL_COUNT))
echo "Tests: $PASS_COUNT/$TOTAL_PACKAGES packages pass"

# Calculate average coverage
COVERAGE_SUM=0
COVERAGE_COUNT=0
while IFS= read -r line; do
    if [[ "$line" =~ coverage:\ ([0-9.]+)% ]]; then
        COV="${BASH_REMATCH[1]}"
        COVERAGE_SUM=$(echo "$COVERAGE_SUM + $COV" | bc)
        COVERAGE_COUNT=$((COVERAGE_COUNT + 1))
    fi
done <<< "$TEST_OUTPUT"

if [ "$COVERAGE_COUNT" -gt 0 ]; then
    COVERAGE=$(echo "scale=1; $COVERAGE_SUM / $COVERAGE_COUNT" | bc)
else
    COVERAGE="0"
fi
echo "Coverage: ${COVERAGE}%"

# Determine coverage color
get_coverage_color() {
    local pct=$1
    local pct_int=$(echo "$pct" | cut -d'.' -f1 | tr -d ' ')
    pct_int=${pct_int:-0}
    if [ "$pct_int" -ge 80 ]; then echo "brightgreen"
    elif [ "$pct_int" -ge 60 ]; then echo "green"
    elif [ "$pct_int" -ge 40 ]; then echo "yellow"
    elif [ "$pct_int" -ge 20 ]; then echo "orange"
    else echo "red"
    fi
}

COVERAGE_COLOR=$(get_coverage_color "$COVERAGE")

# Determine tests badge color
FAIL_COUNT=${FAIL_COUNT:-0}
PASS_COUNT=${PASS_COUNT:-0}
if [ "$FAIL_COUNT" -eq 0 ] && [ "$PASS_COUNT" -gt 0 ]; then
    TESTS_COLOR="brightgreen"
    TESTS_MSG="$PASS_COUNT passing"
elif [ "$FAIL_COUNT" -gt 0 ]; then
    TESTS_COLOR="red"
    TESTS_MSG="$FAIL_COUNT failing"
else
    TESTS_COLOR="yellow"
    TESTS_MSG="no tests"
fi

# Count test functions
TEST_FUNC_COUNT=$(grep -r "^func Test" "$ROOT_DIR" --include="*_test.go" 2>/dev/null | wc -l | tr -d ' ')
echo "Test functions: $TEST_FUNC_COUNT"

# Generate badge JSON files (Shields.io endpoint format)
cat > "$BADGES_DIR/badge-version.json" << EOF
{
  "schemaVersion": 1,
  "label": "version",
  "message": "v$VERSION",
  "color": "blue"
}
EOF

cat > "$BADGES_DIR/badge-coverage.json" << EOF
{
  "schemaVersion": 1,
  "label": "coverage",
  "message": "${COVERAGE}%",
  "color": "$COVERAGE_COLOR"
}
EOF

cat > "$BADGES_DIR/badge-tests.json" << EOF
{
  "schemaVersion": 1,
  "label": "tests",
  "message": "$TEST_FUNC_COUNT tests",
  "color": "$TESTS_COLOR"
}
EOF

cat > "$BADGES_DIR/badge-loc.json" << EOF
{
  "schemaVersion": 1,
  "label": "lines of code",
  "message": "$LOC_FORMATTED",
  "color": "informational"
}
EOF

cat > "$BADGES_DIR/badge-packages.json" << EOF
{
  "schemaVersion": 1,
  "label": "packages",
  "message": "$PASS_COUNT passing",
  "color": "$TESTS_COLOR"
}
EOF

echo ""
echo "Badge files written to $BADGES_DIR:"
ls -la "$BADGES_DIR"
