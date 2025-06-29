#!/bin/bash

# Worker Configuration Generator
# This script generates Docker Compose services and environment files based on user input

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

# Function to prompt for user input
prompt_input() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    
    if [ -n "$default" ]; then
        read -p "$prompt [$default]: " input
        eval "$var_name=\${input:-$default}"
    else
        read -p "$prompt: " input
        eval "$var_name=\"$input\""
    fi
}

# Function to prompt for yes/no
prompt_yes_no() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"
    
    while true; do
        if [ "$default" = "y" ]; then
            read -p "$prompt [Y/n]: " yn
        else
            read -p "$prompt [y/N]: " yn
        fi
        
        case $yn in
            [Yy]* ) eval "$var_name=true"; break;;
            [Nn]* ) eval "$var_name=false"; break;;
            "" ) 
                if [ "$default" = "y" ]; then
                    eval "$var_name=true"
                else
                    eval "$var_name=false"
                fi
                break;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

# Function to generate worker environment file
generate_worker_env() {
    local worker_name="$1"
    local worker_queues="$2"
    local app_domain="$3"
    local env_file="$PROJECT_ROOT/api/env/.env.${worker_name}"
    
    print_status "Generating environment file for $worker_name worker..."
    
    cat > "$env_file" << EOF
# Worker-specific environment for $worker_name
WORKER_NAME=$worker_name
WORKER_QUEUES=$worker_queues

# Queue Configuration
QUEUE_CONNECTION=sqs
SQS_ACCESS_KEY=local
SQS_SECRET_KEY=local
SQS_REGION=us-east-1
SQS_ENDPOINT=http://sqs.$app_domain:9324

# Worker Configuration
WORKER_MAX_JOBS=1000
WORKER_MEMORY_LIMIT=128
WORKER_TIMEOUT=60
WORKER_SLEEP=3
WORKER_TRIES=3

# Database Configuration
DB_CONNECTION=mysql
DB_HOST=db
DB_PORT=3306
DB_USER=api_user
DB_PASSWORD=b4s3L4r4G0212!
DB_NAME=dev_base_lara_go
DB_CHARSET=utf8mb4
DB_COLLATION=utf8mb4_unicode_ci
DB_PREFIX=
DB_STRICT=true
DB_ENGINE=InnoDB

# Cache Configuration
CACHE_STORE=local
CACHE_PREFIX=base_lara_go_cache_
CACHE_TTL=3600

# Logging Configuration
LOG_CHANNEL=stack
LOG_LEVEL=debug
LOG_PATH=storage/logs/laravel.log

# API Configuration
API_SECRET=yoursecretstring
TOKEN_HOUR_LIFESPAN=1
EOF

    print_status "Generated $env_file"
}

# Function to generate Docker Compose worker service
generate_worker_service() {
    local worker_name="$1"
    local worker_queues="$2"
    local app_domain="$3"
    local is_first_worker="$4"
    local compose_file="$PROJECT_ROOT/docker-compose.yaml"
    
    print_status "Generating Docker Compose service for $worker_name worker..."
    
    if [ "$is_first_worker" = "true" ]; then
        # Replace the default worker service with the first custom worker
        print_status "Replacing default worker service with $worker_name..."
        
        awk -v worker_name="$worker_name" -v app_domain="$app_domain" '
        BEGIN { in_worker=0 }
        /^[[:space:]]*worker:[[:space:]]*$/ {
            in_worker=1
            print "  " worker_name ":"
            print "    build:"
            print "      context: ./api"
            print "      dockerfile: ../docker/worker/Dockerfile"
            print "    volumes:"
            print "      - ./api:/usr/src/app"
            print "    env_file:"
            print "      - ./api/env/.env." worker_name
            print "    depends_on:"
            print "      db:"
            print "        condition: service_healthy"
            print "      redis:"
            print "        condition: service_started"
            print "      elasticmq:"
            print "        condition: service_started"
            print "      dnsmasq:"
            print "        condition: service_started"
            print "    environment:"
            print "      - VIRTUAL_HOST=" worker_name "." app_domain
            print "      - VIRTUAL_PORT=8081"
            print "      - HTTPS_METHOD=static"
            print "    networks:"
            print "      default:"
            print "        aliases:"
            print "          - " worker_name "." app_domain
            print "    restart: unless-stopped"
            next
        }
        # End of worker block, resume printing
        in_worker && /^[[:space:]]{2}[a-zA-Z0-9_-]+:/ {
            in_worker=0
            print
            next
        }
        # If inside worker block, skip lines
        in_worker { next }
        # Otherwise, print everything
        { print }
        ' "$compose_file" > "$compose_file.tmp" && mv "$compose_file.tmp" "$compose_file"
        
        if [ $? -eq 0 ]; then
            print_status "Replaced default worker with $worker_name in $compose_file"
        else
            print_error "Could not find worker service in $compose_file"
            exit 1
        fi
    else
        # Add additional worker services to the existing file
        print_status "Adding $worker_name worker service to existing docker-compose.yaml..."
        
        cat >> "$compose_file" << EOF

  $worker_name:
    build:
      context: ./api
      dockerfile: ../docker/worker/Dockerfile
    volumes:
      - ./api:/usr/src/app
    env_file:
      - ./api/env/.env.${worker_name}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
      elasticmq:
        condition: service_started
      dnsmasq:
        condition: service_started
    environment:
      - VIRTUAL_HOST=${worker_name}.${app_domain}
      - VIRTUAL_PORT=8081
      - HTTPS_METHOD=static
    networks:
      default:
        aliases:
          - ${worker_name}.${app_domain}
    restart: unless-stopped
EOF
        
        print_status "Added $worker_name worker service to $compose_file"
    fi
}

# Function to prompt user to update Go config with worker configuration
prompt_go_config_update() {
    local workers_config="$1"
    
    print_status "IMPORTANT: You need to manually update your Go queue configuration"
    echo ""
    echo "Please update your config/queue.go file with the following worker configuration:"
    echo ""
    echo "// Add this to your QueueConfig() function:"
    echo ""
    
    # Display the worker configuration in Go format
    echo "$workers_config" | while IFS='|' read -r worker_name worker_queues; do
        if [ -n "$worker_name" ] && [ "$worker_name" != "default" ]; then
            echo "    // Worker: $worker_name"
            echo "    // Queues: $worker_queues"
            echo "    // Add this to your workers map:"
            echo "    \"$worker_name\": {"
            echo "        Queues: []string{\"$(echo $worker_queues | tr ',' '", "')\"},"
            echo "        MaxJobs: 1000,"
            echo "        MemoryLimit: 128,"
            echo "        Timeout: 60,"
            echo "        Sleep: 3,"
            echo "        Tries: 3,"
            echo "    },"
            echo ""
        fi
    done
    
    echo "Example complete configuration:"
    echo "func QueueConfig() map[string]interface{} {"
    echo "    return map[string]interface{}{"
    echo "        \"default_connection\": \"sqs\","
    echo "        \"connections\": map[string]interface{}{"
    echo "            \"sync\": map[string]interface{}{"
    echo "                \"driver\": \"sync\","
    echo "                \"queues\": []string{\"default\"},"
    echo "            },"
    echo "            \"sqs\": map[string]interface{}{"
    echo "                \"driver\": \"sqs\","
    echo "                \"key\": config.Get(\"queue.sqs.key\"),"
    echo "                \"secret\": config.Get(\"queue.sqs.secret\"),"
    echo "                \"region\": config.Get(\"queue.sqs.region\"),"
    echo "                \"endpoint\": config.Get(\"queue.sqs.endpoint\"),"
    echo "                \"queues\": []string{\"mail\", \"jobs\", \"events\", \"default\"},"
    echo "            },"
    echo "        },"
    echo "        \"workers\": map[string]interface{}{"
    echo "            \"default\": map[string]interface{}{"
    echo "                \"queues\": []string{\"mail\", \"jobs\", \"events\", \"default\"},"
    echo "                \"max_jobs\": 1000,"
    echo "                \"memory_limit\": 128,"
    echo "                \"timeout\": 60,"
    echo "                \"sleep\": 3,"
    echo "                \"tries\": 3,"
    echo "            },"
    echo "            // Add your custom workers here"
    echo "        },"
    echo "        \"api_queues\": map[string]interface{}{"
    echo "            \"mail\": \"mail\","
    echo "            \"jobs\": \"jobs\","
    echo "            \"events\": \"events\","
    echo "            \"default\": \"default\","
    echo "        },"
    echo "    }"
    echo "}"
    echo ""
    print_status "Please update your config/queue.go file with the above configuration"
}

# Main function
main() {
    print_header "Worker Configuration Generator"
    
    # Get app domain
    prompt_input "Enter your app domain" "baselaragoproject.test" "APP_DOMAIN"
    
    # Get number of workers
    prompt_input "How many worker instances do you want to create?" "1" "NUM_WORKERS"
    
    # Validate number of workers
    if ! [[ "$NUM_WORKERS" =~ ^[0-9]+$ ]] || [ "$NUM_WORKERS" -lt 1 ]; then
        print_error "Number of workers must be a positive integer"
        exit 1
    fi
    
    # Check if docker-compose.yaml exists
    if [ ! -f "$PROJECT_ROOT/docker-compose.yaml" ]; then
        print_error "docker-compose.yaml not found. Please run the install script first."
        exit 1
    fi
    
    # Initialize variables
    local workers_config=""
    local worker_envs=""
    local worker_names=()
    
    # Configure each worker
    for ((i=1; i<=NUM_WORKERS; i++)); do
        print_header "Configuring Worker $i"
        
        # Get worker name
        prompt_input "Enter name for worker $i" "worker_$i" "WORKER_NAME"
        
        # Get queues for this worker
        print_status "Available queues: mail, jobs, events, default"
        prompt_input "Enter queues for $WORKER_NAME (comma-separated)" "default" "WORKER_QUEUES"
        
        # Clean up queue names
        WORKER_QUEUES=$(echo "$WORKER_QUEUES" | tr -d ' ')
        
        # Store worker name for final output
        worker_names+=("$WORKER_NAME")
        
        # Generate worker environment file
        generate_worker_env "$WORKER_NAME" "$WORKER_QUEUES" "$APP_DOMAIN"
        
        # Generate Docker Compose service (first worker replaces default, others are added)
        if [ $i -eq 1 ]; then
            generate_worker_service "$WORKER_NAME" "$WORKER_QUEUES" "$APP_DOMAIN" "true"
        else
            generate_worker_service "$WORKER_NAME" "$WORKER_QUEUES" "$APP_DOMAIN" "false"
        fi
        
        # Add to workers config for queue.json
        workers_config="${workers_config}${WORKER_NAME}|${WORKER_QUEUES}"$'\n'
        
        # Store worker info for API env
        worker_envs="${worker_envs}${WORKER_NAME}|${WORKER_QUEUES}"$'\n'
    done
    
    # Update queue.json
    prompt_go_config_update "$workers_config"
    
    print_header "Configuration Complete"
    print_status "Generated files:"
    print_status "  - Modified docker-compose.yaml"
    
    for worker_name in "${worker_names[@]}"; do
        print_status "  - api/env/.env.${worker_name}"
    done
    
    print_status ""
    print_status "IMPORTANT: You must manually update your config/queue.go file with the worker configuration shown above."
    print_status ""
    print_status "To start your workers, run:"
    print_status "  docker-compose up -d"
    print_status ""
    print_status "Your workers will be available at:"
    for worker_name in "${worker_names[@]}"; do
        print_status "  - https://${worker_name}.${APP_DOMAIN}"
    done
}

# Run main function
main "$@" 