# Setup Scripts & Automation Guide

This document explains the automation and scripting system for onboarding, environment management, and developer experience in this project.

---

## Overview

All setup, install, clean, and domain switching logic is managed by scripts in the `setup/` directory. These scripts ensure:
- Consistent onboarding for all developers
- No hardcoded or stale config/env files
- Easy switching between domains and environments
- Secure handling of secrets and SSL certs

---

## Key Scripts

### `setup/install.sh`
- **Main entry point for project setup.**
- Prompts for environment (`dev`, `staging`, `prod`) and app domain (e.g. `myproject.test`).
- Generates all config, env, and SSL cert files from `.template` files.
- Injects environment-specific values and secrets.
- Starts all containers and services automatically.
- Checks Docker Compose version and offers to update if needed.

### `setup/clean.sh`
- **Removes all generated files and Docker artifacts.**
- Stops and removes all containers, images, and volumes.
- Deletes all generated `.env`, Docker Compose, Nginx, and SSL cert files.
- Ensures a truly clean slate for onboarding or troubleshooting.

### `setup/switch_domain.sh`
- **Switches the app domain and/or environment.**
- Prompts for a new domain and environment.
- Regenerates all configs and envs from templates.
- Reminds the user to run `make clean` and `make install_dev` to apply changes.

### `setup/node.sh`
- **Ensures the correct Node.js version is installed and active via nvm.**
- Installs nvm if missing, and prompts the user if needed.

### `setup/certs.sh`
- **Generates and trusts local SSL certs.**
- Copies certs to the correct locations for Docker and frontend use.

### `setup/health-check.sh`
- **Waits for Sentry (and other services) to be ready.**
- Ignores self-signed cert warnings for local development.

---

## Template System

- All config and env files are generated from `.template` files (e.g. `.env.template`, `docker-compose.yaml.template`).
- Only template files are committed to git; generated files are always ignored.
- The install and switch scripts use `sed` to inject the correct domain, environment, and secrets into the generated files.

---

## Workflow

1. **Onboarding:**
   - Run `make clean && make install_dev`.
   - Enter your desired domain when prompted.
   - All configs, envs, and certs are generated and the stack is started.

2. **Switching Domains/Environments:**
   - Run `make switch_domain`.
   - Enter the new domain and environment.
   - Run `make clean && make install_dev` to apply changes.

3. **Cleaning:**
   - Run `make clean` to remove all generated files and Docker artifacts.

---

## Best Practices

- **Never edit generated files directly.** Always edit the `.template` files and rerun the install or switch scripts.
- **Commit only template files.** All generated files are ignored by git.
- **If you hit issues, always start with `make clean && make install_dev`.**

---

For more details, see the comments in each script in the `setup/` directory. 