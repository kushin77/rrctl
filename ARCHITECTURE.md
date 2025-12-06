# ARCHITECTURE.md - rrctl Enterprise Architecture

## Overview

rrctl is a modular DevOps automation CLI built with Go, designed for enterprise-grade security scanning, CI/CD integration, and AI-powered root cause analysis (RCA).

**Key Design Principle:** "Supply Chain Security First"

All binaries are signed, all dependencies are vendored, all builds are reproducible.

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     rrctl CLI (main.go)                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Commands   │  │   Commands   │  │   Commands   │          │
│  │ (scan, rca)  │  │ (git, docker)│  │ (cicd, k8s)  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                 │                    │                │
│         └─────────────────┴────────────────────┘                │
│                           │                                      │
├─────────────────────────────────────────────────────────────────┤
│                     Core Packages (pkg/)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Telemetry   │  │   Security   │  │  Ollama AI   │          │
│  │ (OpenTel)    │  │  (gosec)     │  │  (RCA)       │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Remediation │  │  Compliance  │  │   Storage    │          │
│  │  (Repair)    │  │  (Audit)     │  │  (Config)    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                    External Integrations                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Ollama     │  │   Git        │  │   Docker     │          │
│  │   (Local)    │  │   (GitHub)   │  │   (Registry) │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐                            │
│  │  Kubernetes  │  │  CI/CD       │                            │
│  │  (kubectl)   │  │  (GitHub)    │                            │
│  └──────────────┘  └──────────────┘                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Key Design Decisions

### 1. **Modular Command Structure**

Each major command family (scan, rca, git, etc.) is isolated in separate files:

```go
// rrctl/scan.go       - Security scanning commands
// rrctl/rca.go        - Root cause analysis commands
// rrctl/git_cmd.go    - Git operations
// rrctl/docker_cmd.go - Container operations
```

**Why:** 
- Easy to add new commands without polluting main.go
- Commands can be independently unit tested
- Clear separation of concerns
- Easier refactoring as requirements change

### 2. **Package Organization**

```
pkg/
├── telemetry/       # OpenTelemetry distributed tracing
├── security/        # Security scanning engines
├── remediation/     # Auto-remediation logic
├── compliance/      # Audit and compliance tracking
├── storage/         # Config and state management
└── ai/              # AI integration (Ollama, RCA)
```

**Why:**
- Reusable across commands
- Clear dependency tree
- Easy to mock for testing
- Vendorable (no external dependencies in main)

### 3. **Telemetry-First**

Every command is instrumented with:
- **Distributed tracing** (OpenTelemetry/Jaeger)
- **Structured logging** (logrus with JSON)
- **Metrics** (Prometheus-compatible)

**Why:**
- Enterprises need observability to operate rrctl in production
- Tracing helps debug failures in CI/CD pipelines
- Audit logging for compliance (SOC2, HIPAA, PCI-DSS)

### 4. **Supply Chain Security**

All binaries are:
- **Built reproducibly** (same source → same binary)
- **Signed with Sigstore** (cosign verification)
- **Verified with attestations** (build metadata)
- **Dependencies vendored** (no network calls at runtime)

**Why:**
- Prevents binary tampering attacks
- Meets SLSA build level 3 requirements
- Customers can verify authenticity

---

## Dependency Management

### Vendoring Strategy

```bash
# All dependencies are in vendor/ directory
# CI/CD builds with: go build -mod=vendor

# This prevents:
# - Network failures during build
# - Dependency hijacking
# - Supply chain compromises
```

### Dependency Scanning

```bash
# Automated with:
# 1. Nancy (CVE scanning)
# 2. Dependabot (pull requests for updates)
# 3. Renovate (automatic updates + auto-merge safe ones)
# 4. Trivy (image scanning in Docker builds)
```

---

## Security Architecture

### Input Validation

Every command validates inputs:
```go
if len(targetFile) == 0 {
    return fmt.Errorf("target file is required")
}
if !strings.HasSuffix(targetFile, ".log") {
    return fmt.Errorf("target file must be a .log file")
}
```

### No Hardcoded Secrets

All sensitive data comes from:
- Environment variables (CI/CD secrets)
- ~/.rrctl/config.yaml (user-controlled)
- OS keyring integration (future)

### Audit Logging

Every action is logged:
```json
{
  "timestamp": "2025-12-05T21:50:00Z",
  "user": "alice@company.com",
  "command": "rca analyze",
  "target": "error.log",
  "result": "success",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "duration_ms": 1234
}
```

---

## Performance Characteristics

### RCA Analysis

- **Input:** Error log (1MB)
- **Latency:** <50ms (log parsing) + <5s (Ollama inference)
- **Memory:** ~50MB peak
- **Output:** JSON report

### Security Scanning

- **Input:** Source directory (100 files, 50K LOC)
- **Latency:** <5s (SAST) + <10s (dependency scan)
- **Memory:** ~100MB peak
- **Output:** SARIF vulnerability report

### Git Analysis

- **Input:** Git repository (10K commits)
- **Latency:** <2s (commit analysis)
- **Memory:** ~30MB peak
- **Output:** Risk report with commit correlation

---

## Testing Strategy

### Unit Tests

```bash
# Run with race condition detection
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Coverage threshold: 80% minimum
# Mutation testing: (planned for v1.1)
```

### Integration Tests

```bash
# Test with real Ollama, Docker, Git:
make test-integration

# Runs against:
# - Local Ollama instance
# - Real Docker daemon
# - Test Git repositories
```

### E2E Tests

```bash
# Test complete workflows:
make test-e2e

# Verifies:
# - RCA analysis end-to-end
# - Security scanning pipeline
# - CI/CD integration
# - Multi-command workflows
```

---

## Deployment Modes

### 1. **Local CLI**

```bash
rrctl install /usr/local/bin
rrctl scan security --target .
```

### 2. **Docker**

```bash
docker run -v /path/to/code:/workspace \
  elevatediq/rrctl:latest \
  scan security --target /workspace
```

### 3. **CI/CD Integration**

```yaml
# GitHub Actions
- uses: elevatediq/rrctl@v1
  with:
    command: scan security
    target: ./src
```

### 4. **Kubernetes**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: rrctl-scanner
spec:
  containers:
  - name: scanner
    image: elevatediq/rrctl:latest
    args: ["scan", "security", "--target", "/app"]
```

---

## Failure Modes & Recovery

### Command Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// If scan takes >30s, cancel and return error
```

### Ollama Connection Failure

```go
if err := ollama.Health(); err != nil {
    logrus.Warn("Ollama unavailable, falling back to pattern matching")
    // Use built-in patterns instead of AI
}
```

### Disk Space

```go
if !hasDiskSpace(1*GB) {
    return fmt.Errorf("insufficient disk space for report")
}
```

---

## Future Architecture Enhancements

- [ ] **Plugin System** (v1.1) - Custom command registration
- [ ] **gRPC API** (v1.1) - Service mode with API
- [ ] **Multi-tenancy** (v1.2) - Serve multiple teams
- [ ] **Caching Layer** (v1.2) - Redis for RCA results
- [ ] **Event Streaming** (v1.2) - Kafka for audit logs

---

**Document:** ARCHITECTURE.md  
**Version:** 1.0.0  
**Last Updated:** December 5, 2025  
**Audience:** Enterprise architects, DevOps engineers, security teams