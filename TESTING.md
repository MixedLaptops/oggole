# Testing Guide

This document describes how to run tests for the Oggole (WhoKnows) application.

## Test Types

We have implemented three types of tests as per course requirements:

1. **Unit Tests** - Test pure functions in isolation (Go)
2. **Integration Tests** - Test HTTP handlers with database (Go)
3. **End-to-End Tests** - Test full user journeys in browser (Playwright)

## Running Tests

### Unit Tests (Go)

Unit tests don't require a database and test pure functions:

```bash
cd src/backend
go test -v
```

### Integration Tests (Go)

Integration tests require a running PostgreSQL database:

```bash
# Set database URL
export DATABASE_URL="postgresql://user:password@localhost:5432/whoknows"

# Run all tests (unit + integration)
cd src/backend
go test -v
```

### End-to-End Tests (Playwright)

E2E tests require the application to be running:

```bash
# Terminal 1: Start the application
cd src
./main

# Terminal 2: Run Playwright tests
npm run test:e2e

# Run in headed mode (see browser)
npm run test:e2e:headed

# Run in UI mode (interactive)
npm run test:e2e:ui
```

## Test Coverage

### Unit Tests
- Token generation (cryptographic security)
- Client IP extraction (X-Forwarded-For handling)
- Cookie security settings
- Password hashing (bcrypt)
- Session cookie creation

### Integration Tests
- Search handler functionality
- Login authentication (invalid credentials)
- Logout session cleanup

### E2E Tests
- Homepage loads
- Login page accessibility
- Register page accessibility

## CI/CD Integration

Tests are split across two workflows following DevOps best practices:

### Continuous Integration (CI)
**Workflow**: `.github/workflows/continuous-integration.yml`
**Triggers**: Every push to any branch, PRs to main/development
**Purpose**: Fast feedback on code quality
**Tests**:
- Unit tests (Go)
- Linting (Super-Linter in `lint.yml`)

### Continuous Delivery (CD)
**Workflow**: `.github/workflows/continuous-delivery.yml`
**Triggers**: Only when code is merged to main
**Purpose**: System-level validation before deployment
**Tests**:
- Go integration tests (with PostgreSQL)
- End-to-end tests (Playwright - all browsers)
- Docker image build and push (deployment artifact)

This follows the course guide structure:
- **CI**: Unit tests + Linting (fast, runs on all branches)
- **CD**: Integration tests + E2E tests (slower, runs before deployment)
