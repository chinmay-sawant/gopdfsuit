# 🛠️ Makefile Reference

Complete guide to the Makefile targets for building, testing, and deploying GoPdfSuit.

---

## Table of Contents

- [Overview](#overview)
- [Variables](#variables)
- [Docker Targets](#docker-targets)
- [Build Targets](#build-targets)
- [Development Targets](#development-targets)
- [Quick Reference](#quick-reference)

---

## Overview

The Makefile provides convenient shortcuts for common development and deployment tasks. All targets are designed to work from the repository root directory.

---

## Variables

Customize behavior using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `VERSION` | `2.0.0` | Application version tag |
| `DOCKERUSERNAME` | `chinmaysawant` | Docker Hub username |

### Setting Variables

```bash
# Inline
VERSION=1.5.0 make docker

# Export for session
export VERSION=1.5.0
export DOCKERUSERNAME=myusername
make docker
```

---

## Docker Targets

### `make docker`

Build and run the Docker container locally.

```bash
make docker
```

**What it does:**
1. Builds Docker image from `dockerfolder/Dockerfile`
2. Tags image as `gopdfsuit:<VERSION>`
3. Runs container on port 8080

**Equivalent commands:**
```bash
docker build -f dockerfolder/Dockerfile --build-arg VERSION=2.0.0 -t gopdfsuit:2.0.0 .
docker run -d -p 8080:8080 gopdfsuit:2.0.0
```

---

### `make dockertag`

Tag and push the Docker image to Docker Hub.

```bash
make dockertag
```

**What it does:**
1. Tags image with version number
2. Tags image as `latest`
3. Prompts for Docker Hub login
4. Pushes both tags to Docker Hub

**Equivalent commands:**
```bash
docker tag gopdfsuit:2.0.0 chinmaysawant/gopdfsuit:2.0.0
docker tag gopdfsuit:2.0.0 chinmaysawant/gopdfsuit:latest
docker login
docker push chinmaysawant/gopdfsuit:2.0.0
docker push chinmaysawant/gopdfsuit:latest
```

---

### `make pull`

Pull and run the latest image from Docker Hub.

```bash
make pull
```

**What it does:**
1. Pulls image from Docker Hub
2. Runs container on port 8080

**Equivalent commands:**
```bash
docker pull chinmaysawant/gopdfsuit:2.0.0
docker run -d -p 8080:8080 chinmaysawant/gopdfsuit:2.0.0
```

---

## Build Targets

### `make build`

Compile the Go application.

```bash
make build
```

**What it does:**
1. Creates `bin/` directory
2. Compiles to `bin/app`

**Output:** `bin/app`

---

### `make run`

Build frontend and run the application locally.

```bash
make run
```

**What it does:**
1. Builds React frontend (`npm run build`)
2. Runs Go application

**Access:** `http://localhost:8080`

---

### `make clean`

Remove build artifacts.

```bash
make clean
```

**What it does:**
- Deletes `bin/` directory

---

## Development Targets

### `make test`

Run all Go tests.

```bash
make test
```

**Equivalent:**
```bash
go test ./...
```

---

### `make fmt`

Format Go source code.

```bash
make fmt
```

**Equivalent:**
```bash
go fmt ./...
```

---

### `make vet`

Run Go static analysis.

```bash
make vet
```

**Equivalent:**
```bash
go vet ./...
```

---

### `make mod`

Tidy Go module dependencies.

```bash
make mod
```

**Equivalent:**
```bash
go mod tidy
```

---

## Quick Reference

| Command | Description | Use Case |
|---------|-------------|----------|
| `make docker` | Build & run Docker container | Local Docker testing |
| `make dockertag` | Push to Docker Hub | Release deployment |
| `make pull` | Pull & run from Docker Hub | Production deployment |
| `make build` | Compile Go binary | Local builds |
| `make run` | Build frontend & run app | Development |
| `make test` | Run tests | CI/CD, pre-commit |
| `make clean` | Remove build artifacts | Cleanup |
| `make fmt` | Format code | Pre-commit |
| `make vet` | Static analysis | Code quality |
| `make mod` | Tidy dependencies | After adding packages |

---

## Common Workflows

### Development Cycle

```bash
make fmt          # Format code
make vet          # Check for issues
make test         # Run tests
make run          # Start development server
```

### Release Workflow

```bash
export VERSION=2.1.0
make docker       # Build and test locally
make dockertag    # Push to Docker Hub
```

### Fresh Start

```bash
make clean        # Remove old builds
make mod          # Update dependencies
make build        # Compile fresh binary
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find and kill existing container
docker ps
docker stop <container_id>

# Or use different port
docker run -d -p 9090:8080 gopdfsuit:2.0.0
```

### Docker Login Issues

```bash
# Manual login
docker login

# Then retry
make dockertag
```

### Build Failures

```bash
# Clean and rebuild
make clean
make mod
make build
```
