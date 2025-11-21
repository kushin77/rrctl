# rrctl Examples

Practical examples for using rrctl in real-world scenarios.

## Repository Analysis

### Quick Health Check

Scan your repo for workflow issues:

```bash
rrctl repo-defrag --path . --json report.json --md report.md --plan plan.md
```

### Deep Dive with GitHub API

Include PR and environment analysis:

```bash
export GITHUB_TOKEN="ghp_your_token_here"
rrctl repo-defrag \
  --path . \
  --github-owner your-org \
  --github-repo your-repo \
  --days-stale 90 \
  --json report.json \
  --md report.md \
  --plan cleanup-plan.md
```

### Customer Audit

Run analysis for a customer repository:

```bash
# Clone customer repo
git clone https://github.com/customer/repo
cd repo

# Generate comprehensive report
rrctl repo-defrag \
  --path . \
  --github-owner customer \
  --github-repo repo \
  --github-token $CUSTOMER_GITHUB_TOKEN \
  --json audit-report.json \
  --md audit-report.md \
  --plan recommendations.md

# Email reports to customer
mail -s "CI/CD Health Audit" customer@example.com < audit-report.md
```

## Automated Fixes

### Preview Changes (Safe)

Review what would be fixed without modifying files:

```bash
rrctl repo-autofix --path . --dry-run --patch preview.patch
less preview.patch
```

### Apply Fixes to Development Branch

Create a feature branch with automated improvements:

```bash
# Create feature branch
git checkout -b ci/workflow-improvements

# Apply fixes
rrctl repo-autofix --path . --dry-run=false

# Review changes
git diff

# Commit and push
git add .github/workflows/
git commit -m "ci: add concurrency and pin actions

- Add concurrency blocks to prevent duplicate runs
- Pin actions to stable versions for reproducibility
- Generated via rrctl repo-autofix"

git push origin ci/workflow-improvements

# Create PR
gh pr create --title "CI Improvements" --body "Automated workflow enhancements"
```

### Batch Process Multiple Repos

Process all repos in an organization:

```bash
#!/bin/bash
REPOS=(
  "org/repo1"
  "org/repo2"
  "org/repo3"
)

for repo in "${REPOS[@]}"; do
  echo "Processing $repo..."
  git clone "https://github.com/$repo" "/tmp/$repo"
  
  rrctl repo-defrag \
    --path "/tmp/$repo" \
    --json "/tmp/reports/${repo//\//_}.json"
  
  rrctl repo-autofix \
    --path "/tmp/$repo" \
    --dry-run \
    --patch "/tmp/patches/${repo//\//_}.patch"
done
```

## CI/CD Integration

### GitHub Actions Workflow

Add as a CI check:

```yaml
name: Workflow Health Check

on:
  pull_request:
    paths:
      - '.github/workflows/**'
  schedule:
    - cron: '0 0 * * 0' # Weekly on Sunday

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install rrctl
        run: |
          curl -fsSL https://raw.githubusercontent.com/kushin77/elevatedIQ/main/rrctl-opensource/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Analyze workflows
        run: |
          rrctl repo-defrag \
            --path . \
            --github-owner ${{ github.repository_owner }} \
            --github-repo ${{ github.event.repository.name }} \
            --github-token ${{ secrets.GITHUB_TOKEN }} \
            --json workflow-report.json \
            --md workflow-report.md \
            --plan cleanup-plan.md
      
      - name: Upload reports
        uses: actions/upload-artifact@v4
        with:
          name: workflow-reports
          path: |
            workflow-report.json
            workflow-report.md
            cleanup-plan.md
      
      - name: Comment on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('workflow-report.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## Workflow Analysis\n\n${report}`
            });
```

### GitLab CI Pipeline

```yaml
workflow-health:
  stage: test
  image: golang:1.21
  script:
    - curl -fsSL https://raw.githubusercontent.com/kushin77/elevatedIQ/main/rrctl-opensource/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
    - rrctl repo-defrag --path . --json report.json --md report.md
  artifacts:
    paths:
      - report.json
      - report.md
    expire_in: 30 days
```

## Security Scanning

### Basic Security Scan

```bash
rrctl security-scan --path . --secrets --deps --perms
```

### Comprehensive Security Audit

```bash
# Run all security checks
rrctl security-scan \
  --path . \
  --secrets \
  --deps \
  --perms \
  > security-report.txt

# Review findings
cat security-report.txt
```

## Tips and Tricks

### Filter Reports by Severity

Extract only high-priority issues from JSON report:

```bash
rrctl repo-defrag --path . --json report.json

# Extract workflows with unpinned actions
jq '.workflows[] | select(.usesUnpinnedAction == true) | .file' report.json

# Count workflows missing concurrency
jq '[.workflows[] | select(.hasConcurrency == false)] | length' report.json
```

### Generate Summary for Executives

Create a high-level summary:

```bash
cat <<EOF > executive-summary.md
# CI/CD Health Summary

$(rrctl repo-defrag --path . --json report.json >/dev/null 2>&1)

## Key Metrics
- Total Workflows: $(jq '.summary.workflowCount' report.json)
- Needs Attention: $(jq '.summary.workflowsWithUnpinned + .summary.workflowsWithoutConcurrency' report.json)
- Stale Workflows: $(jq '.summary.workflowsStale' report.json)

## Recommended Actions
1. Pin unpinned actions for reproducibility
2. Add concurrency to reduce resource waste
3. Review and archive stale workflows

Detailed report: report.md
EOF
```

### Combine with Other Tools

Chain with other CI/CD tools:

```bash
# Run rrctl analysis
rrctl repo-defrag --path . --json report.json

# Check for critical issues
critical=$(jq '.summary.workflowsWithUnpinned' report.json)

# Fail if critical issues found
if [ "$critical" -gt 10 ]; then
  echo "Too many unpinned actions: $critical"
  exit 1
fi

# Generate fixes
rrctl repo-autofix --path . --patch fixes.patch

# Apply via git
git apply fixes.patch
```

## Customer Use Cases

### Onboarding New Customer

```bash
# 1. Initial assessment
rrctl repo-defrag \
  --path customer-repo \
  --json baseline.json \
  --md baseline-report.md

# 2. Generate improvement plan
rrctl repo-defrag \
  --path customer-repo \
  --plan improvements.md

# 3. Present to customer with automated fixes ready
rrctl repo-autofix \
  --path customer-repo \
  --dry-run \
  --patch proposed-fixes.patch
```

### Quarterly Review

```bash
# Run quarterly analysis
date=$(date +%Y-Q$(( ($(date +%-m)-1)/3+1 )))
rrctl repo-defrag \
  --path . \
  --json "reports/${date}-report.json" \
  --md "reports/${date}-report.md"

# Compare with previous quarter
diff reports/2024-Q3-report.json reports/2024-Q4-report.json
```
