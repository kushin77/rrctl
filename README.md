# rrctl - The DevOps Swiss Army Knife

Open-source mirror staging (ignored)

This directory is used to stage content for the rrctl open-source mirror. All files are ignored by git except this README.

Use external tooling/pipelines to publish or sync this content to the designated public repository.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/kushin77/elevatedIQ)](https://goreportcard.com/report/github.com/kushin77/elevatedIQ)

`rrctl` is a high-performance CLI tool that consolidates DevOps, security, and AI automation into a single, unified interface. Built for modern development teams that need to ship fast without compromising on security or quality.

## 🚀 What is rrctl?

rrctl (RoundRobin Control) is the command-line companion to the ElevatedIQ.ai platform. It brings together:

- **Security Automation**: SAST/DAST/SCA scanning with AI-powered threat detection
- **DevOps Orchestration**: CI/CD, deployment, and infrastructure automation
- **AI-Powered Analysis**: Root cause analysis, code enhancement, and intelligent assistance
- **Monitoring & Observability**: Real-time metrics, alerting, and performance tracking

## ✨ Key Features

### 🧹 Repo Defragmentation (CI hygiene)
Analyze GitHub Actions workflows, pull requests, and repository environments for staleness, failures, and consolidation opportunities. Works offline for local workflow scans; optionally enriches with GitHub API.

```bash
# Local scan of .github/workflows (no API required)
rrctl repo-defrag --path /path/to/repo \
  --json report.json \
  --md report.md \
  --plan cleanup-plan.md

# With GitHub API enrichment (PRs, environments, failure rates)
rrctl repo-defrag \
  --path /path/to/repo \
  --github-owner your-org \
  --github-repo your-repo \
  --github-token $GITHUB_TOKEN \
  --days-stale 60 \
  --json report.json \
  --md report.md \
  --plan cleanup-plan.md

# Auto-fix workflows (add concurrency, pin actions)
# Dry run (default) generates patch without modifying files
rrctl repo-autofix --path /path/to/repo \
  --dry-run \
  --patch autofix.patch

# Apply fixes directly
rrctl repo-autofix --path /path/to/repo \
  --dry-run=false
```

What it checks:
- .github/workflows: last modified time (stale > N days), schedules, triggers, runners, concurrency settings
- Pinning of actions: flags uses: owner/repo@main/latest or missing @ref
- Deprecated runner hints (e.g., ubuntu-22.04 → ubuntu-24.04)
- Duplicate/overlapping triggers suggesting consolidation
- Optional GitHub API:
  - Workflow failure rates (over last N runs)
  - Stale open PRs (> N days without update)
  - Stale repository environments (no recent deployments)

Auto-fix capabilities:
- Add concurrency blocks to prevent duplicate workflow runs
- Pin common actions (checkout, setup-go, setup-node, etc.) to stable versions
- Generate unified diff patches for review before applying

Outputs both JSON and Markdown reports with actionable recommendations plus optional Cleanup Plan and patch files.

### 🔒 Security Suite

```bash
# Comprehensive security scanning
rrctl security scan --comprehensive

# AI-powered threat analysis
rrctl ai security-analyst analyze --threat-level high

# Automated remediation
rrctl auto-remediation --issue CVE-2023-1234 --dry-run
```

### 🤖 AI Integration

```bash
# Root cause analysis with Ollama
rrctl rca oracle "build failing with dependency errors"

# Code enhancement suggestions
rrctl ai enhancer scan --language go

# IT Oracle chatbot
rrctl ai oracle chat --interactive
```

### 🚀 DevOps Automation

```bash
# Round-robin deployment
rrctl roundrobin --strategy blue-green

# Container management
rrctl container build --immutable
rrctl container deploy --environment staging

# Git workflow automation
rrctl git roundrobin --branches feature/*
```

### 📊 Monitoring & Analytics

```bash
# Performance monitoring
rrctl monitor performance --service api-gateway

# Security posture assessment
rrctl monitor security --continuous

# Metrics collection
rrctl monitor collect --exporter prometheus
```

## 🛠️ Installation

### From Source 

```bash
git clone https://github.com/kushin77/elevatedIQ.git
cd elevatedIQ/services/roundrobin-core/cmd/rrctl
go build -o rrctl .
sudo mv rrctl /usr/local/bin/
```

### Pre-built Binaries
Download from [GitHub Releases](https://github.com/kushin77/elevatedIQ/releases)

### Docker

```bash
docker run -it elevatediq/rrctl:latest rrctl --help
```

## 📖 Usage

### Quick Start

```bash
# One-line install
curl -fsSL https://raw.githubusercontent.com/kushin77/elevatedIQ/main/rrctl-opensource/install.sh | bash

# Or download binary directly
# Linux: rrctl-linux-amd64
# macOS Intel: rrctl-darwin-amd64
# macOS ARM: rrctl-darwin-arm64
# Windows: rrctl-windows-amd64.exe

# Verify installation
rrctl version

# View available commands
rrctl --help
```

### Getting Started

```bash
# Initialize rrctl
rrctl version

# View available commands
rrctl --help

# Enable verbose logging
rrctl --verbose security scan
```

### Common Workflows

#### Security-First Development

```bash
# 1. Scan code for vulnerabilities
rrctl security scan --sast --dast

# 2. AI analysis of findings
rrctl ai security-analyst analyze

# 3. Automated fixes
rrctl auto-remediation --apply

# 4. Commit with confidence
rrctl git commit --security-verified
```

#### AI-Enhanced Development

```bash
# 1. Analyze codebase for improvements
rrctl ai enhancer scan

# 2. Get root cause analysis
rrctl rca oracle "performance issue in API"

# 3. Generate documentation
rrctl ai document generate --format md
```

#### DevOps Orchestration

```bash
# 1. Build and test
rrctl build --parallel
rrctl test e2e

# 2. Deploy with monitoring
rrctl deploy --environment production --monitor

# 3. Rollback if needed
rrctl deploy rollback --to v1.2.3
```

## 🔧 Configuration

Create `~/.roundrobin.yaml`:

```yaml
observability:
  logging:
    level: info
    format: json
  metrics:
    enabled: true
    port: 9091

ai:
  ollama:
    base_url: http://localhost:11434
    model: it-oracle
  gemini:
    api_key: ${GEMINI_API_KEY}

security:
  scanners:
    semgrep:
      enabled: true
      rules: custom-rules.yaml
    zap:
      enabled: true
      config: zap-baseline.yaml

devops:
  docker:
    registry: gcr.io/your-project
  kubernetes:
    namespace: default
    context: your-cluster
```

## 🤖 AI Agents

rrctl integrates with multiple AI agents:

| Agent | Port | Purpose |
|-------|------|---------|
| **IT Oracle** | 8081 | Root cause analysis, technical guidance |
| **Security Analyst** | 8083 | Threat detection, compliance |
| **DevOps Engineer** | 8082 | Infrastructure automation |
| **Marketing Manager** | 8091 | Content generation, campaigns |
| **QA Troubleshooter** | 8089 | Test analysis, debugging |

## 📊 Monitoring & Metrics

rrctl exposes Prometheus metrics on port 9091:

```bash
# View metrics
curl http://localhost:9091/metrics

# Key metrics
rrctl_commands_total{command="security_scan"} 42
rrctl_scan_duration_seconds{scanner="sast"} 15.2
rrctl_ai_requests_total{agent="oracle"} 128
```

## 🔐 Security

- **Zero-trust architecture**: All operations are auditable
- **Secret management**: Integration with GCP Secret Manager
- **Compliance**: SOC2, PCI-DSS, HIPAA support
- **Vulnerability scanning**: Continuous security assessment

## 🧪 Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
rrctl test integration

# Run security tests
rrctl test security

# Run performance benchmarks
rrctl benchmark --duration 60s
```

 
## 📚 Documentation

- [Examples and Use Cases](./EXAMPLES.md) - Real-world scenarios and recipes
- [Changelog](./CHANGELOG.md) - Version history and updates
- [Complete Command Reference](./docs/commands.md)
- [API Documentation](./docs/api.md)
- [Security Guide](./docs/security.md)
- [AI Integration](./docs/ai.md)
- [Contributing Guidelines](./CONTRIBUTING.md)

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](./CONTRIBUTING.md) for details.

### Development Setup
```bash
# Clone and setup
git clone https://github.com/kushin77/elevatedIQ.git
cd elevatedIQ

# Install dependencies
go mod download

# Build rrctl
cd services/roundrobin-core/cmd/rrctl
go build -o rrctl .

# Run tests
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ by the ElevatedIQ.ai team
- Inspired by modern DevOps practices and AI-first development
- "Hack Me If You Can" - We're building unhackable systems

## 📞 Support

- **Documentation**: [docs.elevatediq.ai](https://docs.elevatediq.ai)
- **Issues**: [GitHub Issues](https://github.com/kushin77/elevatedIQ/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kushin77/elevatedIQ/discussions)
- **Security**: [security@elevatediq.ai](mailto:security@elevatediq.ai)

---

**Ready to supercharge your DevOps workflow?** 🚀

[Get Started](./docs/getting-started.md) | [Security Overview](./docs/security.md) | [AI Features](./docs/ai.md)
 