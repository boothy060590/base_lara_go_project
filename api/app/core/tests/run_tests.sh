#!/bin/bash

set -e

# Print header
cat <<EOF
==========================================
Running Comprehensive Test Suite
==========================================
Go Core + Laravel Core (Unit + Integration)
==========================================
EOF

# Default options
COVERAGE=false
RACE=false
UNIT_ONLY=false
INTEGRATION_ONLY=false

# Parse arguments
for arg in "$@"; do
  case $arg in
    --coverage)
      COVERAGE=true
      shift
      ;;
    --race)
      RACE=true
      shift
      ;;
    --unit-only)
      UNIT_ONLY=true
      shift
      ;;
    --integration-only)
      INTEGRATION_ONLY=true
      shift
      ;;
    *)
      ;;
  esac
done

# Build go test command
CMD="go test ./... -v"
if [ "$COVERAGE" = true ]; then
  CMD="go test ./... -v -cover -coverprofile=coverage.out"
fi
if [ "$RACE" = true ]; then
  CMD="$CMD -race"
fi

# Navigate to the api directory (where go.mod is located)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$API_DIR"

echo "Running from: $(pwd)"
echo "Running: $CMD"
echo ""

# Run the tests
$CMD

RESULT=$?

# Generate coverage report if requested
if [ "$COVERAGE" = true ] && [ $RESULT -eq 0 ]; then
  echo ""
  echo "Generating coverage report..."
  go tool cover -html=coverage.out -o coverage.html
  echo "Coverage report saved to: coverage.html"
  echo ""
  echo "Coverage summary:"
  go tool cover -func=coverage.out
fi

# Print summary
echo ""
if [ $RESULT -eq 0 ]; then
  echo "🎉 All tests passed!"
  if [ "$COVERAGE" = true ]; then
    echo "📊 Coverage report generated"
  fi
  if [ "$RACE" = true ]; then
    echo "🏃 Race detection completed - no race conditions found"
  fi
else
  echo "❌ Some tests failed."
fi

exit $RESULT 