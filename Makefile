# Developer Makefile for base_lara_go_project

.PHONY: help install certs npm-install sentry-secret env-inject clean up down reset install_dev install_staging install_prod quickstart health-check switch_domain

help:
	@echo "Available commands:"
	@echo "  make install        # Full project setup (certs, npm, secrets, envs)"
	@echo "  make certs          # Generate and trust local SSL certs"
	@echo "  make npm-install    # Run containerized npm install for frontend"
	@echo "  make sentry-secret  # Generate Sentry secret key and inject into envs"
	@echo "  make env-inject     # Copy env templates, inject secrets, etc."
	@echo "  make health-check   # Wait for Sentry to be ready"
	@echo "  make switch_domain  # Change the app domain and regenerate .env files"
	@echo "  make clean          # Full Docker clean slate (like dockerkill)"
	@echo "  make up             # Bring up the stack"
	@echo "  make down           # Bring down the stack"
	@echo "  make reset          # Clean + up"
	@echo "  make install_dev    # Install for development environment"
	@echo "  make install_staging # Install for staging environment"
	@echo "  make install_prod   # Install for production environment"
	@echo "  make quickstart     # Start stack with existing envs and node_modules (no full setup)"

install: certs npm-install sentry-secret env-inject

certs:
	bash setup/certs.sh

npm-install:
	bash setup/npm-install.sh

sentry-secret:
	bash setup/sentry-secret.sh

env-inject:
	bash setup/env-inject.sh

health-check:
	bash setup/health-check.sh

clean:
	bash setup/clean.sh

up:
	docker-compose up -d

down:
	docker-compose down

reset: clean up

install_dev:
	bash setup/install.sh dev

install_staging:
	bash setup/install.sh staging

install_prod:
	bash setup/install.sh prod

quickstart:
	@if [ ! -d frontend/node_modules ]; then \
		echo "node_modules missing, running npm install..."; \
		bash setup/npm-install.sh; \
	fi
	@if [ ! -f api/.env ]; then \
		echo "api/.env missing, please run make install_dev or similar first!"; \
		exit 1; \
	fi
	docker-compose up -d
	@echo "\n⏳ Stack is starting up..."
	@echo "   You can check if Sentry is ready with: make health-check"

switch_domain:
	bash setup/switch_domain.sh 