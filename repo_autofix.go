package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	autofixPath          string
	autofixWorkflowsPath string
	autofixDryRun        bool
	autofixPatchOut      string
	autofixJSON          bool
)

var repoAutofixCmd = &cobra.Command{
	Use:   "repo-autofix",
	Short: "Auto-fix workflows: add concurrency, pin common actions",
	Long: `Apply safe automated fixes to .github/workflows:
- Add concurrency block with cancel-in-progress
- Pin common unpinned actions to latest stable versions
- Output unified diff patch for review/apply`,
	RunE: runRepoAutofix,
}

func init() {
	rootCmd.AddCommand(repoAutofixCmd)

	repoAutofixCmd.Flags().StringVarP(&autofixPath, "path", "p", ".", "Root path of the repository")
	repoAutofixCmd.Flags().StringVar(&autofixWorkflowsPath, "workflows", ".github/workflows", "Relative path to workflows directory")
	repoAutofixCmd.Flags().BoolVar(&autofixDryRun, "dry-run", true, "Dry run mode (default true); set false to write changes")
	repoAutofixCmd.Flags().StringVar(&autofixPatchOut, "patch", "", "Write unified diff patch to file (optional)")
	repoAutofixCmd.Flags().BoolVar(&autofixJSON, "json", false, "Output results in JSON format")
}

func runRepoAutofix(cmd *cobra.Command, args []string) error {
	root := autofixPath
	wfPath := filepath.Join(root, autofixWorkflowsPath)

	entries, err := os.ReadDir(wfPath)
	if err != nil {
		return fmt.Errorf("read workflows dir: %w", err)
	}

	var allPatches []string
	fixCount := 0

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		full := filepath.Join(wfPath, name)
		original, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", full, err)
			continue
		}

		fixed, changed := applyAutoFixes(string(original), name)
		if !changed {
			continue
		}

		fixCount++
		if autofixDryRun {
			if !autofixJSON {
				fmt.Printf("[DRY RUN] Would fix: %s\n", name)
			}
		} else {
			if err := os.WriteFile(full, []byte(fixed), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", full, err)
				continue
			}
			if !autofixJSON {
				fmt.Printf("Fixed: %s\n", name)
			}
		}

		// Generate unified diff for patch
		if autofixPatchOut != "" {
			patch := generateUnifiedDiff(name, string(original), fixed)
			allPatches = append(allPatches, patch)
		}
	}

	if autofixPatchOut != "" && len(allPatches) > 0 {
		combined := strings.Join(allPatches, "\n")
		if err := os.WriteFile(autofixPatchOut, []byte(combined), 0o644); err != nil {
			return fmt.Errorf("write patch: %w", err)
		}
		if !autofixJSON {
			fmt.Printf("Wrote patch to %s\n", autofixPatchOut)
		}
	}

	if autofixJSON {
		return outputAutofixJSON(fixCount, autofixDryRun, allPatches)
	}

	if autofixDryRun {
		fmt.Printf("\nDry run complete. %d files would be modified.\n", fixCount)
		fmt.Println("Run with --dry-run=false to apply changes.")
	} else {
		fmt.Printf("\nApplied fixes to %d files.\n", fixCount)
	}

	return nil
}

type autofixResult struct {
	Success       bool     `json:"success"`
	DryRun        bool     `json:"dry_run"`
	FilesModified int      `json:"files_modified"`
	PatchFile     string   `json:"patch_file,omitempty"`
	Message       string   `json:"message"`
}

func outputAutofixJSON(fixCount int, dryRun bool, patches []string) error {
	result := autofixResult{
		Success:       true,
		DryRun:        dryRun,
		FilesModified: fixCount,
		PatchFile:     autofixPatchOut,
	}

	if dryRun {
		result.Message = fmt.Sprintf("Dry run complete. %d files would be modified.", fixCount)
	} else {
		result.Message = fmt.Sprintf("Applied fixes to %d files.", fixCount)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// applyAutoFixes attempts to add concurrency and pin common actions
func applyAutoFixes(content, filename string) (string, bool) {
	changed := false
	result := content

	// Parse YAML
	var doc map[string]any
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		// Fallback to text mode
		result, changed = applyTextFixes(content)
		return result, changed
	}

	// Add concurrency if missing
	if !hasConcurrency(doc) {
		result, changed = addConcurrencyBlock(result)
	}

	// Pin common actions
	pinned, wasChanged := pinCommonActions(result)
	if wasChanged {
		result = pinned
		changed = true
	}

	return result, changed
}

// addConcurrencyBlock inserts concurrency after 'name:' or before 'on:'
func addConcurrencyBlock(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	insertIdx := -1

	// Find insertion point after 'name:' or before 'on:'
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "name:") {
			insertIdx = i + 1
			break
		}
	}
	if insertIdx == -1 {
		for i, line := range lines {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "on:") {
				insertIdx = i
				break
			}
		}
	}

	if insertIdx == -1 {
		return content, false
	}

	concurrencyBlock := []string{
		"",
		"concurrency:",
		"  group: ${{ github.workflow }}-${{ github.ref }}",
		"  cancel-in-progress: true",
	}

	result := append(lines[:insertIdx], append(concurrencyBlock, lines[insertIdx:]...)...)
	return strings.Join(result, "\n"), true
}

// pinCommonActions pins unpinned actions to known stable versions
func pinCommonActions(content string) (string, bool) {
	pins := map[string]string{
		"actions/checkout":      "v4",
		"actions/setup-go":      "v5",
		"actions/setup-node":    "v4",
		"actions/setup-python":  "v5",
		"actions/cache":         "v4",
		"actions/upload-artifact": "v4",
		"actions/download-artifact": "v4",
		"docker/setup-buildx-action": "v3",
		"docker/login-action": "v3",
		"docker/build-push-action": "v5",
	}

	result := content
	changed := false

	for action, version := range pins {
		// Match uses: action@main or uses: action (no @)
		reUnpinned := regexp.MustCompile(`(\s+uses:\s+` + regexp.QuoteMeta(action) + `)(@(main|master|HEAD|latest))?(\s|$)`)
		if reUnpinned.MatchString(result) {
			result = reUnpinned.ReplaceAllString(result, "${1}@"+version+"${4}")
			changed = true
		}
	}

	return result, changed
}

// applyTextFixes for when YAML parsing fails
func applyTextFixes(content string) (string, bool) {
	result := content
	changed := false

	// Add concurrency if missing
	if !detectConcurrencyFallback(content) {
		result, changed = addConcurrencyBlock(result)
	}

	// Pin actions
	pinned, wasChanged := pinCommonActions(result)
	if wasChanged {
		result = pinned
		changed = true
	}

	return result, changed
}

// generateUnifiedDiff creates a unified diff format patch
func generateUnifiedDiff(filename, original, fixed string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "--- a/%s\n", filename)
	fmt.Fprintf(&buf, "+++ b/%s\n", filename)

	origLines := strings.Split(original, "\n")
	fixedLines := strings.Split(fixed, "\n")

	// Simple line-by-line diff
	maxLen := len(origLines)
	if len(fixedLines) > maxLen {
		maxLen = len(fixedLines)
	}

	fmt.Fprintf(&buf, "@@ -1,%d +1,%d @@\n", len(origLines), len(fixedLines))

	i := 0
	for i < maxLen {
		if i < len(origLines) && i < len(fixedLines) {
			if origLines[i] != fixedLines[i] {
				fmt.Fprintf(&buf, "-%s\n", origLines[i])
				fmt.Fprintf(&buf, "+%s\n", fixedLines[i])
			} else {
				fmt.Fprintf(&buf, " %s\n", origLines[i])
			}
		} else if i < len(origLines) {
			fmt.Fprintf(&buf, "-%s\n", origLines[i])
		} else if i < len(fixedLines) {
			fmt.Fprintf(&buf, "+%s\n", fixedLines[i])
		}
		i++
	}

	return buf.String()
}
