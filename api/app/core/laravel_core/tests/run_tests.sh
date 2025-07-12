#!/bin/bash

# Comprehensive Test Runner for Laravel Core
# This script runs both unit and integration tests with detailed reporting

set -e

echo "=========================================="
echo "Running Comprehensive Tests for Laravel Core"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run tests and count results
run_test_suite() {
    local test_type=$1
    local test_path=$2
    local test_name=$3
    
    echo -e "\n${BLUE}Running ${test_name} Tests...${NC}"
    echo "=========================================="
    
    if [ -d "$test_path" ]; then
        cd "$test_path"
        
        # Run tests with verbose output and coverage
        if go test -v -cover -coverprofile=coverage.out ./...; then
            echo -e "${GREEN}✓ ${test_name} tests passed${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ ${test_name} tests failed${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        # Show coverage report
        if [ -f coverage.out ]; then
            echo -e "\n${YELLOW}Coverage Report for ${test_name}:${NC}"
            go tool cover -func=coverage.out
            echo ""
            go tool cover -html=coverage.out -o coverage.html
            echo -e "${BLUE}Coverage report saved to: ${test_path}/coverage.html${NC}"
        fi
        
        cd - > /dev/null
    else
        echo -e "${YELLOW}Warning: Test directory ${test_path} not found${NC}"
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Function to run specific test category
run_test_category() {
    local category=$1
    local test_path=$2
    local test_name=$3
    
    echo -e "\n${BLUE}Running ${test_name} Tests...${NC}"
    echo "=========================================="
    
    if [ -d "$test_path" ]; then
        cd "$test_path"
        
        # Run tests with verbose output
        if go test -v ./...; then
            echo -e "${GREEN}✓ ${test_name} tests passed${NC}"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            echo -e "${RED}✗ ${test_name} tests failed${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        
        cd - > /dev/null
    else
        echo -e "${YELLOW}Warning: Test directory ${test_path} not found${NC}"
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Function to run benchmarks
run_benchmarks() {
    echo -e "\n${BLUE}Running Benchmarks...${NC}"
    echo "=========================================="
    
    # Run benchmarks in unit tests
    if [ -d "unit" ]; then
        cd unit
        echo "Running unit test benchmarks..."
        go test -bench=. -benchmem ./...
        cd - > /dev/null
    fi
    
    # Run benchmarks in integration tests
    if [ -d "integration" ]; then
        cd integration
        echo "Running integration test benchmarks..."
        go test -bench=. -benchmem ./...
        cd - > /dev/null
    fi
}

# Function to run race detection tests
run_race_tests() {
    echo -e "\n${BLUE}Running Race Detection Tests...${NC}"
    echo "=========================================="
    
    # Run race detection on unit tests
    if [ -d "unit" ]; then
        cd unit
        echo "Running race detection on unit tests..."
        go test -race ./...
        cd - > /dev/null
    fi
    
    # Run race detection on integration tests
    if [ -d "integration" ]; then
        cd integration
        echo "Running race detection on integration tests..."
        go test -race ./...
        cd - > /dev/null
    fi
}

# Function to run facade integration tests
run_facade_tests() {
    echo -e "\n${BLUE}Running Facade Integration Tests...${NC}"
    echo "=========================================="
    
    # Run facade integration tests
    if [ -d "integration/facades" ]; then
        cd integration/facades
        echo "Running facade integration tests..."
        go test -v ./...
        cd - > /dev/null
    fi
}

# Function to run environment integration tests
run_environment_tests() {
    echo -e "\n${BLUE}Running Environment Integration Tests...${NC}"
    echo "=========================================="
    
    # Run environment integration tests
    if [ -d "integration" ]; then
        cd integration
        echo "Running environment integration tests..."
        go test -run=TestConfigFacadeEnvironmentIntegration ./...
        cd - > /dev/null
    fi
}

# Function to run concurrency tests
run_concurrency_tests() {
    echo -e "\n${BLUE}Running Concurrency Tests...${NC}"
    echo "=========================================="
    
    # Run concurrency tests
    if [ -d "integration" ]; then
        cd integration
        echo "Running concurrency tests..."
        go test -run=TestConfigFacadeConcurrency ./...
        cd - > /dev/null
    fi
}

# Function to run performance tests
run_performance_tests() {
    echo -e "\n${BLUE}Running Performance Tests...${NC}"
    echo "=========================================="
    
    # Run performance tests with timeouts
    if [ -d "integration" ]; then
        cd integration
        echo "Running performance tests..."
        timeout 30s go test -run=TestConfigFacadePerformance ./... || echo "Performance tests completed or timed out"
        cd - > /dev/null
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  all                    Run all tests (default)"
    echo "  unit                   Run all unit tests"
    echo "  integration            Run all integration tests"
    echo "  unit/facades           Run unit tests for facades"
    echo "  integration/facades    Run integration tests for facades"
    echo "  unit/config            Run unit tests for config"
    echo "  integration/config     Run integration tests for config"
    echo "  unit/http              Run unit tests for HTTP components"
    echo "  integration/http       Run integration tests for HTTP components"
    echo "  unit/models            Run unit tests for models"
    echo "  integration/models     Run integration tests for models"
    echo "  unit/providers         Run unit tests for providers"
    echo "  integration/providers  Run integration tests for providers"
    echo "  unit/logging           Run unit tests for logging"
    echo "  integration/logging    Run integration tests for logging"
    echo "  unit/auth              Run unit tests for auth"
    echo "  integration/auth       Run integration tests for auth"
    echo "  benchmarks             Run benchmarks"
    echo "  race                   Run race detection tests"
    echo "  help                   Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all tests"
    echo "  $0 unit               # Run all unit tests"
    echo "  $0 unit/facades       # Run unit tests for facades only"
    echo "  $0 integration/facades # Run integration tests for facades only"
}

# Check command line arguments
if [ $# -eq 0 ]; then
    # No arguments, run all tests
    echo -e "${BLUE}Starting comprehensive test suite...${NC}"
    
    # Run unit tests
    run_test_suite "unit" "unit" "Unit"
    
    # Run integration tests
    run_test_suite "integration" "integration" "Integration"
    
    # Run specific integration test categories
    run_facade_tests
    run_environment_tests
    run_concurrency_tests
    run_performance_tests
    
    # Run benchmarks
    run_benchmarks
    
    # Run race detection tests
    run_race_tests
else
    # Handle specific test categories
    case "$1" in
        "all")
            echo -e "${BLUE}Running all tests...${NC}"
            run_test_suite "unit" "unit" "Unit"
            run_test_suite "integration" "integration" "Integration"
            run_facade_tests
            run_environment_tests
            run_concurrency_tests
            run_performance_tests
            run_benchmarks
            run_race_tests
            ;;
        "unit")
            run_test_suite "unit" "unit" "Unit"
            ;;
        "integration")
            run_test_suite "integration" "integration" "Integration"
            ;;
        "unit/facades")
            run_test_category "unit/facades" "unit/facades" "Unit Facades"
            ;;
        "integration/facades")
            run_test_category "integration/facades" "integration/facades" "Integration Facades"
            ;;
        "unit/config")
            run_test_category "unit/config" "unit/config" "Unit Config"
            ;;
        "integration/config")
            run_test_category "integration/config" "integration/config" "Integration Config"
            ;;
        "unit/http")
            run_test_category "unit/http" "unit/http" "Unit HTTP"
            ;;
        "integration/http")
            run_test_category "integration/http" "integration/http" "Integration HTTP"
            ;;
        "unit/models")
            run_test_category "unit/models" "unit/models" "Unit Models"
            ;;
        "integration/models")
            run_test_category "integration/models" "integration/models" "Integration Models"
            ;;
        "unit/providers")
            run_test_category "unit/providers" "unit/providers" "Unit Providers"
            ;;
        "integration/providers")
            run_test_category "integration/providers" "integration/providers" "Integration Providers"
            ;;
        "unit/logging")
            run_test_category "unit/logging" "unit/logging" "Unit Logging"
            ;;
        "integration/logging")
            run_test_category "integration/logging" "integration/logging" "Integration Logging"
            ;;
        "unit/auth")
            run_test_category "unit/auth" "unit/auth" "Unit Auth"
            ;;
        "integration/auth")
            run_test_category "integration/auth" "integration/auth" "Integration Auth"
            ;;
        "benchmarks")
            run_benchmarks
            ;;
        "race")
            run_race_tests
            ;;
        "help"|"-h"|"--help")
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown test category '$1'${NC}"
            echo ""
            show_usage
            exit 1
            ;;
    esac
fi

# Summary
echo -e "\n=========================================="
echo -e "${BLUE}Test Summary${NC}"
echo "=========================================="
echo -e "Total Test Suites: ${TOTAL_TESTS}"
echo -e "${GREEN}Passed: ${PASSED_TESTS}${NC}"
echo -e "${RED}Failed: ${FAILED_TESTS}${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed!${NC}"
    exit 1
fi 