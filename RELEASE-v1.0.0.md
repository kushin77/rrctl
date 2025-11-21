# rrctl v1.0.0 - Repo Defragmentation & Auto-Fix Tools

Complete workflow analysis and auto-remediation suite for GitHub repositories.

## 🎯 New Features

### `repo-defrag` Command
Comprehensive workflow analysis tool that scans `.github/workflows` for:
- **Staleness/Freshness**: Detects scheduled workflows and last run times
- **Failures**: Integrates with GitHub API to fetch workflow run history
- **Cleanup Opportunities**: Identifies missing concurrency blocks, unpinned actions
- **Pull Requests**: Lists open PRs for review
- **Environments**: Shows configured deployment environments

### `repo-autofix` Command
Automated workflow remediation with:
- **Concurrency Insertion**: Adds concurrency blocks to prevent duplicate runs
- **Action Pinning**: Pins common actions to stable versions (checkout@v4, setup-go@v5, etc.)
- **Dry-Run Mode**: Generate patches for review before applying
- **Safe Execution**: Creates backups and validates YAML integrity

## 📦 Installation

### One-Line Installer (Linux/macOS)
```bash
curl -fsSL https://raw.githubusercontent.com/kushin77/rrctl/master/install.sh | bash
```

### Manual Download
Download the appropriate binary for your platform from the Assets section below.

## 📚 Quick Start

### Analyze Your Repository
```bash
# Basic scan with Markdown output
rrctl repo-defrag --owner kushin77 --repo elevatedIQ --format markdown

# Full analysis with GitHub API enrichment
rrctl repo-defrag --owner kushin77 --repo elevatedIQ --token $GITHUB_TOKEN --format plan
```

### Auto-Fix Workflows
```bash
# Dry-run to preview changes
rrctl repo-autofix --dir .github/workflows --dry-run

# Apply fixes
rrctl repo-autofix --dir .github/workflows
```

## 🔍 Customer Use Cases

See [EXAMPLES.md](https://github.com/kushin77/rrctl/blob/master/EXAMPLES.md) for detailed scenarios:
- Repository health audits
- CI/CD cost optimization
- Security compliance scanning
- Executive summaries for stakeholders
- Automated cleanup pipelines

## 📊 Tested & Validated

Ran on elevatedIQ repository:
- ✅ 63 workflows analyzed
- ✅ 59 missing concurrency blocks detected
- ✅ 3 workflows with unpinned actions found
- ✅ Generated 386KB patch with fixes

## 🛠️ Technical Details

- **Language**: Go 1.24.10
- **CLI Framework**: Cobra v1.10.1
- **YAML Parsing**: yaml.v3 with regex fallback for malformed files
- **GitHub Integration**: REST API for workflow runs, PRs, environments
- **Cross-Platform**: Linux (amd64), macOS (amd64/arm64), Windows (amd64)

## 📄 Documentation

- [README.md](https://github.com/kushin77/rrctl/blob/master/README.md) - Main documentation
- [EXAMPLES.md](https://github.com/kushin77/rrctl/blob/master/EXAMPLES.md) - Real-world use cases
- [BINARIES.md](https://github.com/kushin77/rrctl/blob/master/BINARIES.md) - Installation guide
- [CHANGELOG.md](https://github.com/kushin77/rrctl/blob/master/CHANGELOG.md) - Version history

## 🔐 Checksums

SHA256 checksums for all binaries are available in [checksums.txt](https://github.com/kushin77/rrctl/blob/master/checksums.txt).

---

**Ready for Customer Use** - Fully tested, documented, and packaged for immediate deployment.
