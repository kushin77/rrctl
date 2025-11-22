# Changelog

All notable changes to rrctl will be documented in this file.

## [1.1.0] - 2025-11-22

### Added

- **JSON output format** for `repo-autofix` command via `--json` flag
- Structured JSON response schema: `{success, dry_run, files_modified, patch_file, message}`
- Machine-readable output for CI/CD pipeline integration
- Suppresses verbose console output when JSON mode is enabled

### Changed

- Improved error handling in `repo-autofix` command
- Enhanced output formatting for better readability in JSON mode

## [1.0.0] - 2025-11-21

### Added

- **repo-defrag** command: Comprehensive CI/CD workflow analyzer
  - Scans `.github/workflows` for staleness, unpinned actions, missing concurrency
  - Detects deprecated runners (ubuntu-22.04, macos-12) and suggests upgrades
  - Flags workflows with excessive triggers/schedules for consolidation
  - Optional GitHub API integration for:
    - Workflow failure rates over recent runs
    - Stale pull requests (> N days without updates)
    - Stale repository environments (no recent deployments)
  - Multi-format output:
    - JSON report for automation/integration
    - Markdown report for documentation
    - Cleanup Plan with actionable recommendations and code snippets
  - Tolerant YAML parsing with regex-based fallback for malformed files
  - Detects concurrency at both workflow and job levels
  - Skips local/container actions when checking for unpinned refs

- **repo-autofix** command: Automated workflow maintenance
  - Adds concurrency blocks with cancel-in-progress to prevent duplicate runs
  - Pins common unpinned actions to latest stable versions:
    - actions/checkout@v4
    - actions/setup-go@v5
    - actions/setup-node@v4
    - actions/setup-python@v5
    - docker/* actions to v3/v5
  - Dry-run mode (default) for safe preview
  - Generates unified diff patches for review/apply
  - Batch processes all workflows in a repository

### Changed

- Enhanced security-scan command with better pattern matching
- Improved file permission checks to reduce false positives

### Fixed

- README markdown lint compliance (heading structure, fenced blocks)
- YAML multi-document parsing for complex workflow files

## [v1.0.0] - 2024

### Added

- Initial open-source release
- Basic security scanning capabilities
- Version and completion commands
- MIT License
