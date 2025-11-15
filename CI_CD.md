# CI/CD Pipeline Documentation

This document describes the continuous integration and delivery pipeline for the GXF Discord Bot.

## Overview

The CI/CD pipeline is implemented using GitHub Actions and includes:

1. **Linting** - Code quality and style checks
2. **Testing** - Unit tests with race detection and coverage
3. **SAST** - Static Application Security Testing
4. **Build** - Binary compilation and artifact creation
5. **Docker** - Container image build and publish to GHCR

## Pipeline Stages

### 1. Lint

**Runs on:** Every push and pull request

**Tools:**
- **golangci-lint** - Comprehensive Go linter with 25+ sub-linters

**Linters enabled:**
- `errcheck` - Unchecked error detection
- `gosimple` - Code simplification suggestions
- `govet` - Suspicious constructs
- `staticcheck` - Static analysis
- `gosec` - Security-focused checks
- `gocritic` - Opinionated Go linter
- `gocyclo` - Cyclomatic complexity (max 15)
- `dupl` - Code duplication detection
- `revive` - Fast, extensible linter
- And 15+ more (see `.golangci.yml`)

**Configuration:** `.golangci.yml`

**Local execution:**
```bash
make lint           # Run linter
make lint-fix       # Auto-fix issues
make install-lint   # Install golangci-lint
```

### 2. Test

**Runs on:** Every push and pull request

**Features:**
- Unit tests for all packages
- Race condition detection (`-race`)
- Code coverage reporting
- Codecov integration (optional)

**Coverage threshold:** No minimum (informational only)

**Local execution:**
```bash
make test              # Run tests
make test-race         # With race detector
make test-coverage     # Generate coverage report
make test-bench        # Run benchmarks
```

**Test files:**
- `internal/config/manager_test.go` - Config manager tests
- `internal/config/utils_test.go` - Utility function tests
- `internal/handlers/ratelimiter_test.go` - Rate limiter tests

### 3. SAST (Security Scanning)

**Runs on:** Every push and pull request

**Tools:**

#### Gosec
- **Purpose:** Go security vulnerability scanner
- **Checks:** SQL injection, command injection, file traversal, crypto issues
- **Output:** SARIF format uploaded to GitHub Security tab

#### Trivy
- **Purpose:** Vulnerability scanner for dependencies and code
- **Checks:** Known CVEs in dependencies, misconfigurations
- **Severity:** CRITICAL and HIGH only
- **Output:** SARIF format uploaded to GitHub Security tab

**Local execution:**
```bash
make security          # Run gosec locally
```

**Configuration:**
- Gosec excludes: G104 (handled by errcheck), G304 (needed for config loading)

### 4. Build

**Runs on:** After successful lint, test, and SAST
**Trigger:** Push and pull requests

**Process:**
1. Compile Go binary with optimizations
2. Strip debug symbols (`-ldflags="-w -s"`)
3. Static linking (`CGO_ENABLED=0`)
4. Upload as GitHub artifact (7-day retention)

**Platforms:** `linux/amd64`

**Local execution:**
```bash
make build             # Build binary
```

### 5. Docker Build & Push

**Runs on:** Push to `main` or `develop` branches only
**Registry:** GitHub Container Registry (ghcr.io)

**Process:**
1. Multi-stage Docker build
2. Build for multiple platforms (`linux/amd64`, `linux/arm64`)
3. Push to GHCR with multiple tags
4. Scan image with Trivy
5. Upload scan results to Security tab

**Image tags:**
- `latest` - Latest commit on main branch
- `main` - Latest commit on main branch
- `develop` - Latest commit on develop branch
- `main-<sha>` - Specific commit on main
- `develop-<sha>` - Specific commit on develop

**Image location:**
```
ghcr.io/<owner>/gxf-discord-bot:latest
ghcr.io/<owner>/gxf-discord-bot:main
ghcr.io/<owner>/gxf-discord-bot:develop
ghcr.io/<owner>/gxf-discord-bot:main-abc1234
```

**Local execution:**
```bash
make docker-build      # Build image
make docker-run        # Run container
make docker-stop       # Stop container
```

### 6. Config Validation

**Runs on:** After successful build
**Purpose:** Ensure generated config is valid

**Process:**
1. Download built binary artifact
2. Generate sample config
3. Validate config against CUE schema

**Local execution:**
```bash
make generate          # Generate config
make validate          # Validate config
```

## Workflow Diagram

```
┌─────────────────────┐
│   Push / PR Event   │
└──────────┬──────────┘
           │
           ├─────────────┬─────────────┬─────────────┐
           │             │             │             │
           ▼             ▼             ▼             ▼
      ┌────────┐   ┌────────┐   ┌────────┐   ┌────────┐
      │  Lint  │   │  Test  │   │  SAST  │   │ Build  │
      └───┬────┘   └───┬────┘   └───┬────┘   └───┬────┘
          │            │            │            │
          └────────────┴────────────┴────────────┘
                       │
                       ▼
              ┌────────────────┐
              │ All Jobs Pass? │
              └────────┬───────┘
                       │
          ┌────────────┴────────────┐
          │                         │
          ▼ (main/develop)          ▼ (other)
    ┌──────────┐             ┌──────────┐
    │  Docker  │             │ Complete │
    │ Build &  │             └──────────┘
    │  Push    │
    └────┬─────┘
         │
         ▼
    ┌──────────┐
    │  Trivy   │
    │  Image   │
    │  Scan    │
    └──────────┘
```

## Required Secrets

### GitHub Secrets

| Secret | Required | Description |
|--------|----------|-------------|
| `GITHUB_TOKEN` | Yes (auto) | Provided by GitHub, used for GHCR push |
| `CODECOV_TOKEN` | No | Optional token for Codecov integration |

**Note:** `GITHUB_TOKEN` is automatically provided by GitHub Actions.

## Permissions

The workflow requires the following permissions:

```yaml
permissions:
  contents: read        # Read repository content
  packages: write       # Push to GHCR
  security-events: write # Upload SARIF results
  actions: read         # Read workflow status
```

## Branch Protection

Recommended branch protection rules for `main`:

- ✅ Require pull request reviews
- ✅ Require status checks to pass:
  - Lint
  - Test
  - SAST
  - Build
- ✅ Require branches to be up to date
- ✅ Require conversation resolution
- ❌ Allow force pushes (disabled)

## Local CI Execution

Run all CI checks locally before pushing:

```bash
make ci
```

This will:
1. Download dependencies
2. Run linter
3. Run tests with coverage
4. Run security scanner
5. Report results

## Continuous Deployment

### Manual Deployment

1. **Pull image from GHCR:**
   ```bash
   docker pull ghcr.io/<owner>/gxf-discord-bot:latest
   ```

2. **Run container:**
   ```bash
   docker run -d \
     -e DISCORD_BOT_TOKEN="your-token" \
     -v $(pwd)/config.yaml:/app/config/config.yaml \
     ghcr.io/<owner>/gxf-discord-bot:latest
   ```

### Kubernetes Deployment

Update image in your Kubernetes manifests:

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: bot
        image: ghcr.io/<owner>/gxf-discord-bot:main-abc1234
        imagePullPolicy: Always
```

Apply with:
```bash
kubectl apply -f k8s/deployment.yaml
```

### Image Pull from GHCR

For public repositories, no authentication needed.

For private repositories:

1. **Create GitHub Personal Access Token** with `read:packages` scope

2. **Create Kubernetes secret:**
   ```bash
   kubectl create secret docker-registry ghcr-secret \
     --docker-server=ghcr.io \
     --docker-username=<github-username> \
     --docker-password=<github-token> \
     --docker-email=<email>
   ```

3. **Reference in deployment:**
   ```yaml
   spec:
     imagePullSecrets:
     - name: ghcr-secret
   ```

## Monitoring & Alerts

### GitHub Actions

- View workflow runs: Repository → Actions tab
- Check logs for each step
- Download artifacts from workflow runs

### Security Alerts

- View SAST results: Repository → Security → Code scanning
- Alerts appear for high/critical vulnerabilities
- Review and dismiss false positives

### Coverage Reports

- View coverage trends in Codecov (if enabled)
- Download coverage artifacts from workflow runs
- Open `coverage.html` locally

## Troubleshooting

### Linting Failures

```bash
# Run locally to see issues
make lint

# Auto-fix simple issues
make lint-fix

# Check specific file
golangci-lint run path/to/file.go
```

### Test Failures

```bash
# Run specific package
go test -v ./internal/config

# Run specific test
go test -v -run TestName ./internal/config

# Verbose output with race detector
go test -v -race ./...
```

### SAST False Positives

Edit `.golangci.yml` to exclude specific issues:

```yaml
issues:
  exclude-rules:
    - linters:
        - gosec
      text: "G404: Use of weak random number generator"
```

### Docker Build Failures

```bash
# Test build locally
make docker-build

# Build without cache
docker build --no-cache -t gxf-discord-bot:test .

# Check multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 .
```

## Performance Metrics

### Build Times (typical)

- Lint: ~1 minute
- Test: ~2 minutes
- SAST: ~2 minutes
- Build: ~1 minute
- Docker: ~3-5 minutes

**Total pipeline time:** ~7-10 minutes

### Resource Usage

- Worker pool: 10 concurrent workers
- Max queue: 100 tasks
- Test parallelization: Enabled

## Best Practices

1. **Run `make ci` before pushing** to catch issues early
2. **Keep dependencies updated** - use Dependabot
3. **Write tests for new features** - maintain coverage
4. **Review security alerts** promptly
5. **Use semantic commit messages** for clear changelogs
6. **Tag releases** with semver (v1.0.0)
7. **Update CHANGELOG.md** for releases

## Future Improvements

- [ ] Add integration tests
- [ ] Implement E2E tests with mock Discord server
- [ ] Add performance benchmarking to CI
- [ ] Implement automatic dependency updates
- [ ] Add release automation with goreleaser
- [ ] Implement GitOps with ArgoCD/Flux
- [ ] Add Helm chart publishing
- [ ] Implement canary deployments
