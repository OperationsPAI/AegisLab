# CI Workflow Documentation

This repository now includes a comprehensive CI workflow that performs the following checks:

## Workflow Overview

The CI workflow (`CI.yml`) runs on:
- Push to `main`, `master`, or `develop` branches
- Pull requests to `main`, `master`, or `develop` branches

## Jobs

### 1. Go Code Quality & Tests (`go-checks`)
- **Format Check**: Validates Go code formatting with `gofmt`
- **Linting**: Runs `go vet` to catch common errors
- **Build**: Compiles the Go application
- **Unit Tests**: Executes Go tests (excludes integration tests requiring external dependencies)
- **Dependencies**: Includes Go module caching for faster builds

### 2. Python SDK Quality & Tests (`python-checks`)
- **Package Build**: Verifies Python package can be built successfully
- **Unit Tests**: Runs pytest tests (handles missing SDK dependencies gracefully)
- **Dependencies**: Includes pip package caching for faster builds

### 3. API Documentation (`swagger-docs`)
- **Swagger Generation**: Generates API documentation using swag
- **Validation**: Ensures documentation files are properly created

## Features

- **Caching**: Both Go modules and Python packages are cached for performance
- **Error Handling**: Gracefully handles expected failures from integration tests
- **Multi-language Support**: Handles both Go backend and Python SDK
- **Dependency Management**: Properly manages dependencies for both ecosystems

## Requirements

The workflow uses:
- Go 1.23.2
- Python 3.12
- Ubuntu latest runners

## Status Checks

All three jobs must pass for the CI to be considered successful, ensuring code quality and preventing regressions.