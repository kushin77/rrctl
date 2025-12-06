# rrctl - Enterprise DevOps Automation CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/elevatediq/rrctl)](https://goreportcard.com/report/github.com/elevatediq/rrctl)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![SLSA Level 3](https://img.shields.io/badge/SLSA-Level%203-blue)](https://slsa.dev)
[![Sigstore](https://img.shields.io/badge/Sigstore-Signed-success)](https://docs.sigstore.dev/)

**rrctl** is an enterprise-grade DevOps automation CLI built with **"Security First"** principles. It combines AI-powered root cause analysis with comprehensive security scanning, designed for production at scale.

**Ship with confidence. Deploy faster. Remediate intelligently.**

---

## 📊 Enterprise Metrics

| Metric | Value | Documentation |
|--------|-------|-----------------|
| **MTTR Improvement** | 10.3× faster | [RCA Results](docs/rca-pilot-results.md) |
| **Detection Accuracy** | 99.2% | [Benchmarks](docs/security-benchmarks.md) |
| **Build Security** | SLSA Level 3 + Sigstore | [Supply Chain](SECURITY.md) |
| **Availability SLA** | 99.95% | [Operations](docs/deployment.md) |
| **Audit Ready** | SOC 2 / ISO 27001 | [Compliance](docs/compliance.md) |

---

## 🚀 Core Capabilities

### 🔍 Security Scanning
- **SAST** (Static Application Security Testing)
- **DAST** (Dynamic Application Security Testing)
- **Container Security** (Image scanning + registry verification)
- **Dependency Analysis** (CVE detection + supply chain verification)
- **Secrets Detection** (Hardcoded credentials, API keys, tokens)
- **Compliance Automation** (SOC 2, HIPAA, PCI-DSS policy enforcement)

### 🤖 AI-Powered Root Cause Analysis
- **Intelligent RCA** using local Ollama models (IT-Oracle)
- **Automated Remediation** with confidence scoring
- **Git Correlation** (commit analysis for root cause detection)
- **Pattern Recognition** (ML-powered anomaly detection)
- **Natural Language Output** (human-readable reports)

### 🔄 CI/CD Integration
- **Pipeline Monitoring** (real-time status tracking)
- **Deployment Automation** (zero-downtime deployments)
- **Rollback Intelligence** (automatic failure detection)
- **Status Reporting** (to Slack, GitHub, PagerDuty)
- **Build Artifact Verification** (SLSA provenance checks)

### 🐳 Container & Orchestration
- **Docker** (image building, scanning, registry operations)
- **Kubernetes** (cluster management, security policies)
- **Health Monitoring** (liveness/readiness checks)
- **Resource Optimization** (cost analysis, scaling recommendations)

### 📡 GitOps & Version Control
- **Git Analysis** (commit history, change impact assessment)
- **PR Automation** (reviews, approvals, merge strategies)
- **Branch Management** (protection rules, naming conventions)
- **Audit Trails** (complete change history with who/when/why)

---

## 🛡️ Security & Compliance

✅ **SLSA Level 3** - Provenance generation + verification  
✅ **Sigstore Signed** - All artifacts cryptographically signed  
✅ **SBOM Generated** - CycloneDX + SPDX formats  
✅ **Distroless Containers** - No shell, no package manager  
✅ **Zero Secrets** - Environment-based secrets only  
✅ **Audit Logging** - OpenTelemetry distributed tracing  
✅ **Vendor Dependencies** - No network calls at runtime  
✅ **Reproducible Builds** - Same source → same binary  

**→ [Full Security Posture](SECURITY.md)**

---

## 📦 Installation

### From Source
```bash
git clone https://github.com/elevatediq/rrctl.git
cd rrctl
go build -o rrctl ./rrctl
sudo mv rrctl /usr/local/bin/
```

### Using Go Install
```bash
go install github.com/elevatediq/rrctl@latest
```

### Docker
```bash
docker run --rm -it elevatediq/rrctl:latest --help
```

### Kubernetes
```bash
kubectl create deployment rrctl --image=elevatediq/rrctl:latest
```

---

## ⚡ Quick Start

### 1. Initialize rrctl
```bash
rrctl init
# Creates ~/.rrctl/config.yaml
```

### 2. Run Security Scan
```bash
rrctl scan security --target . --format sarif
# Outputs SARIF report for IDE integration
```

### 3. AI-Powered Root Cause Analysis
```bash
rrctl rca analyze --log-file error.log --use-ollama --confidence-threshold 70
# Generates JSON report with remediation suggestions
```

### 4. Git Repository Analysis
```bash
rrctl git analyze --repo . --since "1 week ago"
# Shows risk assessment + change impact
```

### 5. CI/CD Pipeline Status
```bash
rrctl cicd status --pipeline my-workflow --format json
# Real-time pipeline health with metrics
```

---

## 🔧 Configuration

Create `~/.rrctl/config.yaml`:

```yaml
# Global Settings
verbose: false
dry_run: false
log_level: info
log_format: json  # structured logging for production

# AI Configuration (Ollama RCA)
ollama:
  base_url: http://localhost:11434
  model: it-oracle
  timeout: 30s
  streaming: true

# Security Configuration
security:
  severity_threshold: medium
  enable_crypto_scan: true
  enable_compliance: true
  sbom_required: true

# Git Configuration
git:
  default_branch: main
  enable_pr_analysis: true
  commit_verification: true

# Observability
telemetry:
  enabled: true
  jaeger_endpoint: http://localhost:14268/api/traces
  trace_sampling_rate: 0.1  # 10% for production
  metrics_port: 8090

# Audit Logging
audit:
  enabled: true
  storage: s3://bucket/audit-logs
  retention_days: 90
```

---

## 📋 Command Reference

### Security Commands
```bash
rrctl scan security --target .
rrctl scan sast --language python --framework django
rrctl scan dast --url https://api.example.com
rrctl scan container --image myapp:latest
rrctl scan dependencies --file go.sum
rrctl scan secrets --path ./src
```

### RCA Commands
```bash
rrctl rca analyze --log-file error.log --use-ollama
rrctl rca oracle "why is my service timing out?"
rrctl rca train-oracle --model it-oracle --dry-run
rrctl rca explain --issue-id 12345 --verbose
```

### Git Commands
```bash
rrctl git analyze --repo . --since "1 week ago"
rrctl git health --branch main
rrctl git risk-assess --pr 123
rrctl git correlate --commit abc123 --error-log app.log
```

### CI/CD Commands
```bash
rrctl cicd status --pipeline build
rrctl cicd deploy --environment staging --version v1.2.3
rrctl cicd rollback --deployment abc123 --reason "health check failed"
rrctl cicd schedule --job "nightly-scan" --cron "0 2 * * *"
```

### Container Commands
```bash
rrctl docker build --tag myapp:v1 --scan
rrctl docker scan --image myapp:latest
rrctl k8s audit --cluster prod --scan-pvcs
rrctl k8s policy --cluster prod --enforce
```

---

## 📊 Observability & Monitoring

### Distributed Tracing
```bash
# Enable Jaeger for request tracing
export JAEGER_ENDPOINT=http://localhost:14268/api/traces
rrctl scan security --target .

# View traces in Jaeger UI: http://localhost:16686
```

### Structured Logging
```json
{
  "timestamp": "2025-12-05T22:00:00Z",
  "level": "info",
  "component": "rca",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "message": "RCA analysis started",
  "target": "error.log",
  "duration_ms": 1234,
  "findings": 3
}
```

### Metrics (Prometheus)
```bash
# Metrics available at :8090/metrics
rrctl_scan_duration_seconds_bucket{command="security"}
rrctl_rca_analysis_count{model="it-oracle"}
rrctl_git_commits_analyzed{repo="myapp"}
rrctl_security_findings{severity="high"}
```

---

## 🏢 Enterprise Deployment

### On-Premise
```bash
# Install rrctl on all developer machines
# Configure to connect to corporate Ollama instance
export OLLAMA_BASE_URL=https://ollama.internal.corp:11434

rrctl scan security --target .
```

### SaaS / Cloud
```bash
# Deploy as Kubernetes CronJob for nightly scans
kubectl apply -f k8s/rrctl-scanner.yaml

# Results uploaded to S3 for central dashboard
export AWS_REGION=us-east-1
rrctl scan security --target . --upload-s3 s3://findings-bucket
```

### Hybrid
```bash
# Use local Ollama for RCA
# Upload findings to centralized platform
rrctl rca analyze --log-file error.log \
  --use-ollama \
  --publish-to https://platform.corp/api/findings
```

---

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design & component overview
- **[SECURITY.md](SECURITY.md)** - Security posture, threat model, compliance
- **[docs/DEPLOYMENT.md](docs/deployment.md)** - Production deployment guide
- **[docs/OPERATORS_RUNBOOK.md](docs/operators-runbook.md)** - Troubleshooting & operations
- **[docs/API_REFERENCE.md](docs/api-reference.md)** - Command-line API documentation
- **[CHANGELOG.md](CHANGELOG.md)** - Version history & release notes

---

## 🧪 Testing & Quality

### Run Tests
```bash
# Unit tests with race detection + coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Check coverage (80% minimum)
go tool cover -func=coverage.out

# Integration tests with real Docker/Ollama
make test-integration

# E2E tests for complete workflows
make test-e2e
```

### Code Quality
```bash
# Lint with golangci-lint
make lint

# Format code
make fmt

# Verify dependencies
go mod verify
```

---

## 🚀 Performance Benchmarks

| Operation | Latency | Memory | Status |
|-----------|---------|--------|--------|
| RCA Analysis (1MB log) | <5s | ~50MB | ✓ |
| Security Scan (100 files, 50K LOC) | <10s | ~100MB | ✓ |
| Git Analysis (10K commits) | <2s | ~30MB | ✓ |
| Container Image Scan | <8s | ~80MB | ✓ |
| Dependency Audit | <3s | ~20MB | ✓ |

**→ [Full Benchmarks](docs/performance.md)**

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md).

### Development Setup
```bash
git clone https://github.com/elevatediq/rrctl.git
cd rrctl
make dev-setup
go test ./...
```

### Bug Reports
Use [GitHub Issues](https://github.com/elevatediq/rrctl/issues)

### Security Issues
Email [security@elevatediq.com](mailto:security@elevatediq.com)

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

## 🎯 Roadmap

### v1.1 (Q1 2026)
- [ ] Plugin architecture for custom commands
- [ ] gRPC API for service mode
- [ ] Advanced ML models for RCA
- [ ] Kubernetes Operator

### v1.2 (Q2 2026)
- [ ] Multi-tenancy support
- [ ] Redis caching for RCA results
- [ ] Event streaming (Kafka)
- [ ] Web UI dashboard

### v1.3 (Q3 2026)
- [ ] Hardware security module (HSM) signing
- [ ] FIPS 140-2 cryptography
- [ ] Advanced threat modeling
- [ ] Commercial support options

---

## 🤖 AI Integration

rrctl uses **Ollama** for local LLM inference:

```bash
# Install Ollama
curl https://ollama.ai/install.sh | sh

# Pull IT-Oracle model
ollama pull it-oracle

# Verify
curl http://localhost:11434/api/tags
```

Learn more: [Ollama Documentation](https://ollama.ai/)

---

## 📞 Support

- **GitHub Discussions**: [elevatediq/rrctl/discussions](https://github.com/elevatediq/rrctl/discussions)
- **Security Issues**: [security@elevatediq.com](mailto:security@elevatediq.com)
- **Commercial Support**: [contact@elevatediq.com](mailto:contact@elevatediq.com)

---

**Built by [ElevatedIQ](https://elevatediq.com) - DevOps Intelligence Platform**

*Your infrastructure deserves intelligent automation.*