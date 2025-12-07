#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo ""
echo -e "${BLUE}${BOLD}🧪 Running Unit Tests...${NC}"
echo ""

# Run tests and capture output
cd "$PROJECT_DIR"
OUTPUT=$(go test ./tests/... -v -count=1 2>&1)
EXIT_CODE=$?

# Count results
PASSED=$(echo "$OUTPUT" | grep -c "^--- PASS")
FAILED=$(echo "$OUTPUT" | grep -c "^--- FAIL")
SKIPPED=$(echo "$OUTPUT" | grep -c "^--- SKIP")
TOTAL=$((PASSED + FAILED + SKIPPED))

# Get time
TIME=$(echo "$OUTPUT" | grep "^ok" | awk '{print $3}' | head -1)

# Print test output (optional - uncomment to see full output)
# echo "$OUTPUT"

echo ""
echo -e "${BOLD}════════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}                    📊 TEST SUMMARY                         ${NC}"
echo -e "${BOLD}════════════════════════════════════════════════════════════${NC}"
echo ""

# Results table
printf "  %-20s %s\n" "Total Tests:" "$TOTAL"
printf "  ${GREEN}%-20s %s${NC}\n" "✅ Passed:" "$PASSED"
printf "  ${RED}%-20s %s${NC}\n" "❌ Failed:" "$FAILED"
printf "  ${YELLOW}%-20s %s${NC}\n" "⏭️  Skipped:" "$SKIPPED"
printf "  %-20s %s\n" "⏱️  Duration:" "$TIME"

echo ""
echo -e "${BOLD}════════════════════════════════════════════════════════════${NC}"

# Show failed tests if any
if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "${RED}${BOLD}❌ FAILED TESTS:${NC}"
    echo "$OUTPUT" | grep "^--- FAIL" | sed 's/--- FAIL: /  • /' | sed 's/ (.*//'
    echo ""
fi

# Show passed tests
if [ $PASSED -gt 0 ]; then
    echo ""
    echo -e "${GREEN}${BOLD}✅ PASSED TESTS:${NC}"
    echo "$OUTPUT" | grep "^--- PASS" | sed 's/--- PASS: /  ✓ /' | sed 's/ (.*//'
    echo ""
fi

echo -e "${BOLD}════════════════════════════════════════════════════════════${NC}"

# Final status
if [ $EXIT_CODE -eq 0 ]; then
    echo ""
    echo -e "${GREEN}${BOLD}🎉 ALL TESTS PASSED!${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}${BOLD}💥 SOME TESTS FAILED!${NC}"
    echo ""
fi

exit $EXIT_CODE

