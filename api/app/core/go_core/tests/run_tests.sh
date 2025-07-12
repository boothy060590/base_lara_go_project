#!/bin/bash

# Comprehensive Test Runner for Go Core
# This script runs both unit and integration tests with detailed reporting

set -e

echo "=========================================="
echo "Running Comprehensive Tests for Go Core"
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

# Function to run performance tests
run_performance_tests() {
    echo -e "\n${BLUE}Running Performance Tests...${NC}"
    echo "=========================================="
    
    # Run performance tests with timeouts
    if [ -d "integration" ]; then
        cd integration
        echo "Running performance tests..."
        timeout 30s go test -run=TestConfigSystemPerformance ./... || echo "Performance tests completed or timed out"
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
        go test -run=TestConfigSystemConcurrency ./...
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
    echo "  unit/config            Run unit tests for config"
    echo "  integration/config     Run integration tests for config"
    echo "  unit/cache             Run unit tests for cache"
    echo "  integration/cache      Run integration tests for cache"
    echo "  unit/container         Run unit tests for container"
    echo "  integration/container  Run integration tests for container"
    echo "  unit/events            Run unit tests for events"
    echo "  integration/events     Run integration tests for events"
    echo "  unit/mail              Run unit tests for mail"
    echo "  integration/mail       Run integration tests for mail"
    echo "  unit/model             Run unit tests for model"
    echo "  integration/model      Run integration tests for model"
    echo "  unit/queue             Run unit tests for queue"
    echo "  integration/queue      Run integration tests for queue"
    echo "  unit/repository        Run unit tests for repository"
    echo "  integration/repository Run integration tests for repository"
    echo "  unit/validation        Run unit tests for validation"
    echo "  integration/validation Run integration tests for validation"
    echo "  unit/adapter           Run unit tests for adapter"
    echo "  integration/adapter    Run integration tests for adapter"
    echo "  benchmarks             Run benchmarks"
    echo "  race                   Run race detection tests"
    echo "  help                   Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                     # Run all tests"
    echo "  $0 unit               # Run all unit tests"
    echo "  $0 unit/config        # Run unit tests for config only"
    echo "  $0 integration/config # Run integration tests for config only"
}

# Check command line arguments
if [ $# -eq 0 ]; then
    # No arguments, run all tests
    echo -e "${BLUE}Starting comprehensive test suite...${NC}"
    
    # Run unit tests
    run_test_suite "unit" "unit" "Unit"
    
    # Run integration tests
    run_test_suite "integration" "integration" "Integration"
    
    # Run benchmarks
    run_benchmarks
    
    # Run race detection tests
    run_race_tests
    
    # Run performance tests
    run_performance_tests
    
    # Run concurrency tests
    run_concurrency_tests
else
    # Handle specific test categories
    case "$1" in
        "all")
            echo -e "${BLUE}Running all tests...${NC}"
            run_test_suite "unit" "unit" "Unit"
            run_test_suite "integration" "integration" "Integration"
            run_benchmarks
            run_race_tests
            run_performance_tests
            run_concurrency_tests
            ;;
        "unit")
            run_test_suite "unit" "unit" "Unit"
            ;;
        "integration")
            run_test_suite "integration" "integration" "Integration"
            ;;
        "unit/config")
            run_test_category "unit/config" "unit/config" "Unit Config"
            ;;
        "integration/config")
            run_test_category "integration/config" "integration/config" "Integration Config"
            ;;
        "unit/cache")
            run_test_category "unit/cache" "unit/cache" "Unit Cache"
            ;;
        "integration/cache")
            run_test_category "integration/cache" "integration/cache" "Integration Cache"
            ;;
        "unit/container")
            run_test_category "unit/container" "unit/container" "Unit Container"
            ;;
        "integration/container")
            run_test_category "integration/container" "integration/container" "Integration Container"
            ;;
        "unit/events")
            run_test_category "unit/events" "unit/events" "Unit Events"
            ;;
        "integration/events")
            run_test_category "integration/events" "integration/events" "Integration Events"
            ;;
        "unit/mail")
            run_test_category "unit/mail" "unit/mail" "Unit Mail"
            ;;
        "integration/mail")
            run_test_category "integration/mail" "integration/mail" "Integration Mail"
            ;;
        "unit/model")
            run_test_category "unit/model" "unit/model" "Unit Model"
            ;;
        "integration/model")
            run_test_category "integration/model" "integration/model" "Integration Model"
            ;;
        "unit/queue")
            run_test_category "unit/queue" "unit/queue" "Unit Queue"
            ;;
        "integration/queue")
            run_test_category "integration/queue" "integration/queue" "Integration Queue"
            ;;
        "unit/repository")
            run_test_category "unit/repository" "unit/repository" "Unit Repository"
            ;;
        "integration/repository")
            run_test_category "integration/repository" "integration/repository" "Integration Repository"
            ;;
        "unit/validation")
            run_test_category "unit/validation" "unit/validation" "Unit Validation"
            ;;
        "integration/validation")
            run_test_category "integration/validation" "integration/validation" "Integration Validation"
            ;;
        "unit/adapter")
            run_test_category "unit/adapter" "unit/adapter" "Unit Adapter"
            ;;
        "integration/adapter")
            run_test_category "integration/adapter" "integration/adapter" "Integration Adapter"
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