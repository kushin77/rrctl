# Contributing to rrctl

Thank you for your interest in contributing to rrctl! We follow elite standards for code quality, security, and testing.

## Code of Conduct

Be respectful. Assume good intent. This project is for everyone.

---

## Getting Started

### 1. Fork & Clone
```bash
git clone https://github.com/YOUR-USERNAME/rrctl.git
cd rrctl
git remote add upstream https://github.com/elevatediq/rrctl.git
```

### 2. Create Feature Branch
```bash
git checkout -b feature/my-feature
```

### 3. Make Changes & Test
```bash
make dev-setup
make test
make lint
```

### 4. Commit with Conventional Commits
```bash
git commit -m "feat(rca): add support for custom models"
# Format: <type>(<scope>): <subject>
# Types: feat, fix, refactor, test, docs, chore, perf, security
```

### 5. Push & Create PR
```bash
git push origin feature/my-feature
# Create PR on GitHub
```

---

## Code Standards

### Go Code Quality
- **Format:** gofmt (enforced in CI)
- **Lint:** golangci-lint (must pass)
- **Tests:** 80% minimum coverage
- **Race detection:** `go test -race ./...`

### Security Requirements
- ✓ No hardcoded secrets
- ✓ Input validation on all user-provided data
- ✓ Output escaping for logs/reports
- ✓ No `unsafe` pointer operations
- ✓ Dependency audit: `nancy sleuth`

### Testing Requirements
- ✓ Unit tests for all new functions
- ✓ Integration tests for new commands
- ✓ Table-driven tests for multiple scenarios
- ✓ Mocking of external dependencies (Ollama, Docker, Git)

### Documentation Requirements
- ✓ Comments for exported functions
- ✓ Example usage in PR description
- ✓ Update CHANGELOG.md
- ✓ Update README.md if user-facing

---

## Pull Request Process

### Before Submitting
1. **Self-review your changes**
   - Does it solve the problem?
   - Are there edge cases?
   - Is error handling comprehensive?

2. **Run full test suite**
   ```bash
   make ci-build  # Equivalent to CI pipeline
   ```

3. **Check code quality**
   ```bash
   make lint
   make fmt
   go mod verify
   ```

4. **Write descriptive PR**
   - What problem does this solve?
   - How was it tested?
   - Any breaking changes?

### Review Process
- Code review required from @elevatediq/maintainers
- All CI checks must pass
- Coverage must not decrease
- No security issues identified

### Approval & Merge
- One approval required for fixes
- Two approvals required for features
- Maintainer will merge when ready

---

## Major Contribution Areas

### Security Improvements
- Vulnerability detection enhancements
- New scanning engines
- Compliance framework additions
- Container security features

**→ Start an issue before major work**

### Performance Optimization
- RCA latency improvements
- Memory optimization
- Caching strategies
- Parallel execution

**→ Include benchmarks in PR**

### Documentation
- User guides
- Runbooks
- Architecture diagrams
- Tutorials

**→ No review delays for docs!**

### Tooling & CI/CD
- GitHub Actions improvements
- Build automation
- Deployment optimization

**→ Test extensively before merging**

---

## Common Commands

```bash
# Run everything locally (like CI does)
make ci-build

# Run specific test subset
go test -run TestRCAAnalysis ./pkg/rca

# Generate coverage report
make test-coverage

# Check for security issues
go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

# Scan dependencies for CVEs
go run github.com/sonatype-nexus-community/nancy@latest sleuth

# Format all Go files
go fmt ./...

# Run linter with auto-fix
golangci-lint run --fix
```

---

## Commit Message Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code restructure (no behavior change)
- `perf` - Performance improvement
- `test` - Test changes
- `docs` - Documentation changes
- `chore` - Build, deps, tooling
- `security` - Security fix or improvement

### Examples
```
feat(rca): add support for custom Ollama models

This allows users to specify custom-trained models for RCA analysis.
Users can now pass --model my-custom-model to rca analyze.

Fixes #123
```

```
fix(security): prevent log injection in audit logs

Input is now properly escaped before being written to audit logs.
Adds test coverage for malicious input patterns.

Refs #456
```

---

## Testing Checklist

- [ ] Unit tests added for new code
- [ ] Integration tests for new features
- [ ] Existing tests still pass
- [ ] `go test -race` passes
- [ ] Coverage ≥ 80%
- [ ] No hardcoded test secrets
- [ ] Mocks used for external dependencies

---

## Security Review Checklist

- [ ] No hardcoded secrets, API keys, or credentials
- [ ] All user input validated
- [ ] All outputs properly escaped
- [ ] No use of `eval()`, `exec()`, or dynamic shell
- [ ] Proper error handling (no stack traces to user)
- [ ] Dependencies up-to-date (check go.mod)
- [ ] No external network calls in tests
- [ ] Secure random generation (crypto/rand)

---

## Performance Considerations

- RCA analysis should complete in <5s
- Security scans should scale to 100K LOC
- Memory usage should stay <200MB peak
- No unbounded loops or allocations
- Use goroutines wisely (connection pooling)

---

## Releasing

Only maintainers can release. Process:

```bash
git checkout main
git pull upstream main

# Bump version using semver
git tag v1.2.3

# Push tag
git push upstream v1.2.3

# GitHub Actions automatically:
# 1. Builds binaries
# 2. Generates SBOM
# 3. Signs with Sigstore
# 4. Creates release with binaries
```

---

## Questions?

- **Discussions**: [GitHub Discussions](https://github.com/elevatediq/rrctl/discussions)
- **Issues**: [GitHub Issues](https://github.com/elevatediq/rrctl/issues)
- **Email**: [contact@elevatediq.com](mailto:contact@elevatediq.com)

Thank you for making rrctl better! 🚀