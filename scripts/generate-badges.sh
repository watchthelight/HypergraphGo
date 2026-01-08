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

# Determine coverage color (using hex for better dark mode contrast)
get_coverage_color() {
    local pct=$1
    local pct_int=$(echo "$pct" | cut -d'.' -f1 | tr -d ' ')
    pct_int=${pct_int:-0}
    if [ "$pct_int" -ge 80 ]; then echo "2ea44f"      # GitHub green - good dark mode contrast
    elif [ "$pct_int" -ge 60 ]; then echo "3fb950"    # Lighter green
    elif [ "$pct_int" -ge 40 ]; then echo "d29922"    # GitHub yellow
    elif [ "$pct_int" -ge 20 ]; then echo "db6d28"    # GitHub orange
    else echo "f85149"                                 # GitHub red
    fi
}

COVERAGE_COLOR=$(get_coverage_color "$COVERAGE")

# Determine tests badge color (using hex for better dark mode contrast)
FAIL_COUNT=${FAIL_COUNT:-0}
PASS_COUNT=${PASS_COUNT:-0}
if [ "$FAIL_COUNT" -eq 0 ] && [ "$PASS_COUNT" -gt 0 ]; then
    TESTS_COLOR="2ea44f"  # GitHub green
    TESTS_MSG="$PASS_COUNT passing"
elif [ "$FAIL_COUNT" -gt 0 ]; then
    TESTS_COLOR="f85149"  # GitHub red
    TESTS_MSG="$FAIL_COUNT failing"
else
    TESTS_COLOR="d29922"  # GitHub yellow
    TESTS_MSG="no tests"
fi

# Count test functions
TEST_FUNC_COUNT=$(grep -r "^func Test" "$ROOT_DIR" --include="*_test.go" 2>/dev/null | wc -l | tr -d ' ')
echo "Test functions: $TEST_FUNC_COUNT"

# ============================================================
# Additional metrics for "Because I love stats" section
# ============================================================

echo ""
echo "Calculating additional metrics..."

# File counts
FILES_TOTAL=$(find "$ROOT_DIR" -type f ! -path '*/.git/*' ! -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
GO_FILES=$(find "$ROOT_DIR" -name '*.go' ! -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
SOURCE_FILES=$(find "$ROOT_DIR" -name '*.go' ! -name '*_test.go' ! -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
TEST_FILE_COUNT=$(find "$ROOT_DIR" -name '*_test.go' ! -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
FOLDERS=$(find "$ROOT_DIR" -type d ! -path '*/.git/*' ! -path '*/vendor/*' 2>/dev/null | wc -l | tr -d ' ')
echo "Files: $FILES_TOTAL total, $GO_FILES Go, $SOURCE_FILES source, $TEST_FILE_COUNT test files, $FOLDERS folders"

# Code metrics
FUNCTIONS=$(grep -r '^func ' "$ROOT_DIR" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
STRUCTS=$(grep -r '^type .* struct' "$ROOT_DIR" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
BENCHMARKS=$(grep -r '^func Benchmark' "$ROOT_DIR" --include='*_test.go' 2>/dev/null | wc -l | tr -d ' ')
echo "Code: $FUNCTIONS functions, $STRUCTS structs, $BENCHMARKS benchmarks"

# Dependencies (count require lines in go.mod, 0 if no deps)
DEPS=$(grep -E '^\s+\S+\s+v' "$ROOT_DIR/go.mod" 2>/dev/null | wc -l | tr -d ' \n')
DEPS=${DEPS:-0}
if [ "$DEPS" -eq 0 ] 2>/dev/null; then
    DEPS_COLOR="2ea44f"  # GitHub green - 0 deps is good!
else
    DEPS_COLOR="58a6ff"  # GitHub blue (informational)
fi
echo "Dependencies: $DEPS"

# Architecture counts
KERNEL_FILES=$(find "$ROOT_DIR/kernel" -name '*.go' ! -name '*_test.go' 2>/dev/null | wc -l | tr -d ' ')
INTERNAL_FILES=$(find "$ROOT_DIR/internal" -name '*.go' ! -name '*_test.go' 2>/dev/null | wc -l | tr -d ' ')
EXAMPLES=$(ls "$ROOT_DIR/examples/" 2>/dev/null | wc -l | tr -d ' ')
WORKFLOWS=$(ls "$ROOT_DIR/.github/workflows/" 2>/dev/null | wc -l | tr -d ' ')
CLI_COMMANDS=22  # hg CLI has 22 commands
echo "Architecture: $KERNEL_FILES kernel, $INTERNAL_FILES internal, $EXAMPLES examples, $WORKFLOWS workflows"

# Git/project vitals
COMMITS=$(git -C "$ROOT_DIR" rev-list --count HEAD 2>/dev/null || echo "0")
COMMITS=$(echo "$COMMITS" | tr -d ' ')

# Count releases (requires gh CLI, fallback to 0)
if command -v gh &> /dev/null; then
    RELEASES=$(gh release list --repo watchthelight/HypergraphGo --limit 100 2>/dev/null | wc -l | tr -d ' ')
else
    RELEASES="0"
fi
RELEASES=$(echo "$RELEASES" | tr -d ' ')

# Calculate project age in days
FIRST_COMMIT_TS=$(git -C "$ROOT_DIR" log --reverse --format='%ct' 2>/dev/null | head -1)
if [ -n "$FIRST_COMMIT_TS" ]; then
    NOW_TS=$(date +%s)
    AGE_DAYS=$(( (NOW_TS - FIRST_COMMIT_TS) / 86400 ))
else
    AGE_DAYS="0"
fi
echo "Vitals: $COMMITS commits, $RELEASES releases, $AGE_DAYS days old"

# TODOs in code (0 is good!)
TODOS=$(grep -ri 'TODO' "$ROOT_DIR" --include='*.go' 2>/dev/null | wc -l | tr -d ' ')
if [ "$TODOS" -eq 0 ]; then
    TODOS_COLOR="2ea44f"  # GitHub green - clean code!
else
    TODOS_COLOR="d29922"  # GitHub yellow
fi
echo "TODOs: $TODOS"

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
  "color": "58a6ff"
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

# ============================================================
# New badges for "Because I love stats" section
# ============================================================

# File structure badges
cat > "$BADGES_DIR/badge-files.json" << EOF
{
  "schemaVersion": 1,
  "label": "files",
  "message": "$FILES_TOTAL",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-go-files.json" << EOF
{
  "schemaVersion": 1,
  "label": "go files",
  "message": "$GO_FILES",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-source-files.json" << EOF
{
  "schemaVersion": 1,
  "label": "source",
  "message": "$SOURCE_FILES",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-test-files.json" << EOF
{
  "schemaVersion": 1,
  "label": "test files",
  "message": "$TEST_FILE_COUNT",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-folders.json" << EOF
{
  "schemaVersion": 1,
  "label": "folders",
  "message": "$FOLDERS",
  "color": "58a6ff"
}
EOF

# Code breakdown badges
cat > "$BADGES_DIR/badge-functions.json" << EOF
{
  "schemaVersion": 1,
  "label": "functions",
  "message": "$FUNCTIONS",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-structs.json" << EOF
{
  "schemaVersion": 1,
  "label": "structs",
  "message": "$STRUCTS",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-test-funcs.json" << EOF
{
  "schemaVersion": 1,
  "label": "test funcs",
  "message": "$TEST_FUNC_COUNT",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-benchmarks.json" << EOF
{
  "schemaVersion": 1,
  "label": "benchmarks",
  "message": "$BENCHMARKS",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-deps.json" << EOF
{
  "schemaVersion": 1,
  "label": "deps",
  "message": "$DEPS",
  "color": "$DEPS_COLOR"
}
EOF

# Architecture badges
cat > "$BADGES_DIR/badge-kernel.json" << EOF
{
  "schemaVersion": 1,
  "label": "kernel",
  "message": "$KERNEL_FILES files",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-internal.json" << EOF
{
  "schemaVersion": 1,
  "label": "internal",
  "message": "$INTERNAL_FILES files",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-examples.json" << EOF
{
  "schemaVersion": 1,
  "label": "examples",
  "message": "$EXAMPLES",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-workflows.json" << EOF
{
  "schemaVersion": 1,
  "label": "workflows",
  "message": "$WORKFLOWS",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-cli.json" << EOF
{
  "schemaVersion": 1,
  "label": "CLI commands",
  "message": "$CLI_COMMANDS",
  "color": "58a6ff"
}
EOF

# Vitals badges
cat > "$BADGES_DIR/badge-commits.json" << EOF
{
  "schemaVersion": 1,
  "label": "commits",
  "message": "$COMMITS",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-releases.json" << EOF
{
  "schemaVersion": 1,
  "label": "releases",
  "message": "$RELEASES",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-age.json" << EOF
{
  "schemaVersion": 1,
  "label": "age",
  "message": "${AGE_DAYS} days",
  "color": "58a6ff"
}
EOF

cat > "$BADGES_DIR/badge-todos.json" << EOF
{
  "schemaVersion": 1,
  "label": "TODOs",
  "message": "$TODOS",
  "color": "$TODOS_COLOR"
}
EOF

echo ""
echo "Badge files written to $BADGES_DIR:"
ls -la "$BADGES_DIR"
