#!/bin/bash
# check-imports.sh - Verify import boundary rules
#
# This script checks that kernel and internal packages follow import rules:
# - Kernel packages don't import parser, tactics, or cmd packages
# - Internal packages don't import tactics or cmd packages
# - Tactics doesn't import kernel internals in wrong direction
#
# Run: ./scripts/check-imports.sh
# Exit code: 0 if OK, 1 if violations found

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

VIOLATIONS=0

check_no_import() {
    local pkg=$1
    local forbidden=$2
    local desc=$3

    # Use go list to get imports
    imports=$(go list -f '{{join .Imports "\n"}}' "./$pkg/..." 2>/dev/null || echo "")

    if echo "$imports" | grep -q "$forbidden"; then
        echo -e "${RED}VIOLATION:${NC} $pkg imports $forbidden ($desc)"
        VIOLATIONS=$((VIOLATIONS + 1))
        return 1
    fi
    return 0
}

echo "Checking import boundaries..."
echo ""

# Kernel packages should not import parser
echo "Checking kernel/check..."
check_no_import "kernel/check" "internal/parser" "kernel must not depend on parser" || true

echo "Checking kernel/ctx..."
check_no_import "kernel/ctx" "internal/parser" "kernel must not depend on parser" || true

echo "Checking kernel/subst..."
check_no_import "kernel/subst" "internal/parser" "kernel must not depend on parser" || true

# Kernel packages should not import tactics
echo "Checking kernel packages for tactics imports..."
check_no_import "kernel" "tactics" "kernel must not depend on tactics" || true

# Internal packages should not import cmd
echo "Checking internal packages for cmd imports..."
check_no_import "internal" "/cmd/" "internal must not depend on cmd" || true

# Internal packages should not import tactics
echo "Checking internal packages for tactics imports..."
check_no_import "internal" "tactics" "internal must not depend on tactics" || true

# Kernel ctx should only import ast
echo "Checking kernel/ctx imports..."
ctx_imports=$(go list -f '{{join .Imports "\n"}}' ./kernel/ctx/... 2>/dev/null | grep "HypergraphGo" || echo "")
for imp in $ctx_imports; do
    if [[ "$imp" != *"internal/ast"* ]]; then
        echo -e "${RED}VIOLATION:${NC} kernel/ctx imports $imp (should only import internal/ast)"
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

# Kernel subst should only import ast
echo "Checking kernel/subst imports..."
subst_imports=$(go list -f '{{join .Imports "\n"}}' ./kernel/subst/... 2>/dev/null | grep "HypergraphGo" || echo "")
for imp in $subst_imports; do
    if [[ "$imp" != *"internal/ast"* ]]; then
        echo -e "${RED}VIOLATION:${NC} kernel/subst imports $imp (should only import internal/ast)"
        VIOLATIONS=$((VIOLATIONS + 1))
    fi
done

echo ""

if [ $VIOLATIONS -eq 0 ]; then
    echo -e "${GREEN}All import boundary checks passed!${NC}"
    exit 0
else
    echo -e "${RED}Found $VIOLATIONS import boundary violation(s)${NC}"
    exit 1
fi
