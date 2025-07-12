#!/bin/bash

# MySQL Benchmark Runner
# This script runs repository benchmarks against MySQL for real-world performance numbers

echo "🚀 Running MySQL Repository Benchmarks"
echo "======================================"

# Check if MySQL is running
if ! nc -z localhost 3306 2>/dev/null; then
    echo "❌ MySQL is not running on localhost:3306"
    echo ""
    echo "To start MySQL, you can:"
    echo "1. Use your Docker Compose: docker-compose up -d db"
    echo "2. Or start MySQL manually: docker run --rm -d --name mysql-benchmark -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=benchmark_db mysql:8"
    echo ""
    echo "Then set these environment variables:"
    echo "export MYSQL_HOST=localhost"
    echo "export MYSQL_PORT=3306"
    echo "export MYSQL_USER=root"
    echo "export MYSQL_PASSWORD=root"
    echo "export MYSQL_DATABASE=benchmark_db"
    echo ""
    exit 1
fi

# Set default MySQL connection if not already set
export MYSQL_HOST=${MYSQL_HOST:-localhost}
export MYSQL_PORT=${MYSQL_PORT:-3306}
export MYSQL_USER=${MYSQL_USER:-root}
export MYSQL_PASSWORD=${MYSQL_PASSWORD:-root}
export MYSQL_DATABASE=${MYSQL_DATABASE:-benchmark_db}

echo "📊 MySQL Connection:"
echo "   Host: $MYSQL_HOST:$MYSQL_PORT"
echo "   User: $MYSQL_USER"
echo "   Database: $MYSQL_DATABASE"
echo ""

# Run MySQL benchmarks
echo "🏃 Running benchmarks..."
go test -tags=mysql -bench=BenchmarkRepository_MySQL -benchmem -benchtime=5s ./app/core/go_core

echo ""
echo "✅ MySQL benchmarks completed!"
echo ""
echo "💡 Compare with SQLite benchmarks:"
echo "   go test -bench=BenchmarkRepository -benchmem -benchtime=5s ./app/core/go_core" 