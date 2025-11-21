package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// RepoDefragReport is the top-level report structure
type RepoDefragReport struct {
	GeneratedAt   time.Time           `json:"generatedAt"`
	RootPath      string              `json:"rootPath"`
	WorkflowsPath string              `json:"workflowsPath"`
	StaleDays     int                 `json:"staleDays"`
	Workflows     []WorkflowReport    `json:"workflows"`
	GitHub        *GitHubReport       `json:"github,omitempty"`
	Summary       RepoDefragSummaries `json:"summary"`
}

// RepoDefragSummaries aggregates quick stats
type RepoDefragSummaries struct {
	WorkflowCount               int `json:"workflowCount"`
	WorkflowsStale              int `json:"workflowsStale"`
	WorkflowsWithUnpinned       int `json:"workflowsWithUnpinned"`
	WorkflowsWithoutConcurrency int `json:"workflowsWithoutConcurrency"`
}

type WorkflowReport struct {
	File               string     `json:"file"`
	Name               string     `json:"name"`
	Triggers           []string   `json:"triggers"`
	Schedules          []string   `json:"schedules"`
	Runners            []string   `json:"runners"`
	HasConcurrency     bool       `json:"hasConcurrency"`
	UsesUnpinnedAction bool       `json:"usesUnpinnedAction"`
	UnpinnedDetails    []string   `json:"unpinnedDetails"`
	DeprecatedHints    []string   `json:"deprecatedHints"`
	LastModified       *time.Time `json:"lastModified,omitempty"`
	Recommendations    []string   `json:"recommendations"`
}

type GitHubReport struct {
	Owner           string             `json:"owner"`
	Repo            string             `json:"repo"`
	WorkflowFailure []WorkflowFailure  `json:"workflowFailureRates,omitempty"`
	PRs             []PRReport         `json:"pullRequests,omitempty"`
	Environments    []EnvironmentProbe `json:"environments,omitempty"`
}

type WorkflowFailure struct {
	Name          string  `json:"name"`
	WorkflowID    int64   `json:"workflowId"`
	SampledRuns   int     `json:"sampledRuns"`
	FailureRate   float64 `json:"failureRate"`
	RecentFailure bool    `json:"recentFailure"`
}

type PRReport struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Draft     bool      `json:"draft"`
	UpdatedAt time.Time `json:"updatedAt"`
	Stale     bool      `json:"stale"`
	HeadSHA   string    `json:"headSha"`
}

type EnvironmentProbe struct {
	Name         string     `json:"name"`
	LastDeployed *time.Time `json:"lastDeployed,omitempty"`
	IsStale      bool       `json:"isStale"`
}

var (
	defragPath          string
	defragWorkflowsPath string
	defragDaysStale     int
	ghOwner             string
	ghRepo              string
	ghToken             string
	ghSampleRuns        int
	jsonOut             string
	mdOut               string
	planOut             string
)

var repoDefragCmd = &cobra.Command{
	Use:   "repo-defrag",
	Short: "Analyze .github/workflows, PRs, and environments for staleness and cleanup",
	Long: `Repo defragmentation scanner.
- Scans .github/workflows locally for staleness, unpinned actions, deprecated runners, and consolidation hints
- Optionally queries GitHub API for workflow run failure rates, stale PRs, and stale environments (requires --github flags)
- Produces JSON and Markdown reports`,
	RunE: runRepoDefrag,
}

func init() {
	rootCmd.AddCommand(repoDefragCmd)

	repoDefragCmd.Flags().StringVarP(&defragPath, "path", "p", ".", "Root path of the repository")
	repoDefragCmd.Flags().StringVar(&defragWorkflowsPath, "workflows", ".github/workflows", "Relative path to workflows directory")
	repoDefragCmd.Flags().IntVar(&defragDaysStale, "days-stale", 60, "Days without change considered stale for workflows/PRs/environments")

	repoDefragCmd.Flags().StringVar(&ghOwner, "github-owner", "", "GitHub owner/org (optional)")
	repoDefragCmd.Flags().StringVar(&ghRepo, "github-repo", "", "GitHub repository name (optional)")
	repoDefragCmd.Flags().StringVar(&ghToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub token for API access (env GITHUB_TOKEN supported)")
	repoDefragCmd.Flags().IntVar(&ghSampleRuns, "github-runs", 20, "Number of recent workflow runs to sample for failure rate")

	repoDefragCmd.Flags().StringVar(&jsonOut, "json", "", "Write JSON report to path (optional)")
	repoDefragCmd.Flags().StringVar(&mdOut, "md", "", "Write Markdown report to path (optional)")
	repoDefragCmd.Flags().StringVar(&planOut, "plan", "", "Write Cleanup Plan (Markdown) to path (optional)")
}

func runRepoDefrag(cmd *cobra.Command, args []string) error {
	root := defragPath
	wfPath := filepath.Join(root, defragWorkflowsPath)

	wfReports, err := scanWorkflows(wfPath, defragDaysStale)
	if err != nil {
		return err
	}

	report := RepoDefragReport{
		GeneratedAt:   time.Now().UTC(),
		RootPath:      root,
		WorkflowsPath: wfPath,
		StaleDays:     defragDaysStale,
		Workflows:     wfReports,
	}

	// Summary
	for _, w := range wfReports {
		report.Summary.WorkflowCount++
		if w.LastModified != nil && time.Since(*w.LastModified) > (time.Duration(defragDaysStale)*24*time.Hour) {
			report.Summary.WorkflowsStale++
		}
		if w.UsesUnpinnedAction {
			report.Summary.WorkflowsWithUnpinned++
		}
		if !w.HasConcurrency {
			report.Summary.WorkflowsWithoutConcurrency++
		}
	}

	// Optional GitHub API enrichments
	if ghOwner != "" && ghRepo != "" && ghToken != "" {
		gh, err := enrichFromGitHub(ghOwner, ghRepo, ghToken, ghSampleRuns, defragDaysStale)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: GitHub enrichment failed: %v\n", err)
		} else {
			report.GitHub = gh
		}
	}

	// Output
	if jsonOut != "" {
		if err := writeJSON(jsonOut, report); err != nil {
			return err
		}
		fmt.Printf("Wrote JSON report to %s\n", jsonOut)
	}
	if mdOut != "" {
		if err := writeMarkdown(mdOut, report); err != nil {
			return err
		}
		fmt.Printf("Wrote Markdown report to %s\n", mdOut)
	}

	if planOut != "" {
		if err := writeCleanupPlan(planOut, report); err != nil {
			return err
		}
		fmt.Printf("Wrote Cleanup Plan to %s\n", planOut)
	}

	// Also print a concise summary to stdout
	fmt.Printf("Workflows: %d, Stale: %d, Unpinned: %d, NoConcurrency: %d\n",
		report.Summary.WorkflowCount,
		report.Summary.WorkflowsStale,
		report.Summary.WorkflowsWithUnpinned,
		report.Summary.WorkflowsWithoutConcurrency,
	)
	if report.GitHub != nil {
		fmt.Printf("GitHub PRs: %d, Environments: %d, Workflows with failure stats: %d\n",
			len(report.GitHub.PRs), len(report.GitHub.Environments), len(report.GitHub.WorkflowFailure),
		)
	}

	return nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func writeMarkdown(path string, r RepoDefragReport) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Repo Defragmentation Report\n\nGenerated: %s UTC\n\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "- Workflows scanned: %d\n- Stale workflows (> %d days): %d\n- Workflows with unpinned actions: %d\n- Workflows without concurrency: %d\n\n",
		r.Summary.WorkflowCount, r.StaleDays, r.Summary.WorkflowsStale, r.Summary.WorkflowsWithUnpinned, r.Summary.WorkflowsWithoutConcurrency,
	)

	fmt.Fprintf(&buf, "## Workflows\n\n")
	for _, w := range r.Workflows {
		lm := "n/a"
		if w.LastModified != nil {
			lm = w.LastModified.Format("2006-01-02")
		}
		fmt.Fprintf(&buf, "### %s\n\n", w.File)
		fmt.Fprintf(&buf, "- Name: %s\n- Triggers: %s\n- Schedules: %s\n- Runners: %s\n- Last Modified: %s\n- Concurrency: %v\n- Unpinned Actions: %v\n",
			valueOr(w.Name, "(none)"), strings.Join(w.Triggers, ", "), strings.Join(w.Schedules, ", "), strings.Join(w.Runners, ", "), lm, w.HasConcurrency, w.UsesUnpinnedAction,
		)
		if len(w.UnpinnedDetails) > 0 {
			fmt.Fprintf(&buf, "  - Unpinned: %s\n", strings.Join(w.UnpinnedDetails, "; "))
		}
		if len(w.DeprecatedHints) > 0 {
			fmt.Fprintf(&buf, "  - Deprecated: %s\n", strings.Join(w.DeprecatedHints, "; "))
		}
		if len(w.Recommendations) > 0 {
			fmt.Fprintf(&buf, "  - Recommendations: %s\n", strings.Join(w.Recommendations, "; "))
		}
		fmt.Fprintln(&buf)
	}

	if r.GitHub != nil {
		fmt.Fprintf(&buf, "## GitHub Insights (%s/%s)\n\n", r.GitHub.Owner, r.GitHub.Repo)
		if len(r.GitHub.WorkflowFailure) > 0 {
			fmt.Fprintf(&buf, "### Workflow Failure Rates\n\n")
			for _, wf := range r.GitHub.WorkflowFailure {
				fmt.Fprintf(&buf, "- %s: failure rate %.0f%% over %d runs\n", wf.Name, wf.FailureRate*100, wf.SampledRuns)
			}
			fmt.Fprintln(&buf)
		}
		if len(r.GitHub.PRs) > 0 {
			fmt.Fprintf(&buf, "### Stale Pull Requests (> %d days)\n\n", r.StaleDays)
			for _, pr := range r.GitHub.PRs {
				if pr.Stale {
					fmt.Fprintf(&buf, "- #%d %s by %s (updated %s)\n", pr.Number, pr.Title, pr.Author, pr.UpdatedAt.Format("2006-01-02"))
				}
			}
			fmt.Fprintln(&buf)
		}
		if len(r.GitHub.Environments) > 0 {
			fmt.Fprintf(&buf, "### Environments\n\n")
			for _, env := range r.GitHub.Environments {
				lm := "never"
				if env.LastDeployed != nil {
					lm = env.LastDeployed.Format("2006-01-02")
				}
				stale := ""
				if env.IsStale {
					stale = " (stale)"
				}
				fmt.Fprintf(&buf, "- %s: last deployment %s%s\n", env.Name, lm, stale)
			}
			fmt.Fprintln(&buf)
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func valueOr(s, alt string) string {
	if strings.TrimSpace(s) == "" {
		return alt
	}
	return s
}

// scanWorkflows walks a workflows directory for YAML files and analyzes them
func scanWorkflows(dir string, daysStale int) ([]WorkflowReport, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read workflows dir: %w", err)
	}
	var out []WorkflowReport
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		full := filepath.Join(dir, name)
		wr, err := analyzeWorkflowFile(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to analyze %s: %v\n", full, err)
			continue
		}
		// Last modified from git
		if ts, err := gitLastModified(full); err == nil {
			wr.LastModified = &ts
		}
		// Recommendations
		wr.Recommendations = recommendForWorkflow(wr, daysStale)
		out = append(out, wr)
	}
	// sort by file for stable output
	sort.Slice(out, func(i, j int) bool { return out[i].File < out[j].File })
	return out, nil
}

func analyzeWorkflowFile(path string) (WorkflowReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return WorkflowReport{}, err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	var selected map[string]any
	for {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			// YAML parsing failed; use tolerant fallback via regex-based extraction
			return analyzeWorkflowTextFallback(path)
		}
		// Choose the first doc that looks like a workflow (has 'on' at top level or 'jobs')
		if m != nil && (m["on"] != nil || m["jobs"] != nil) {
			selected = m
			break
		}
		if selected == nil && m != nil {
			// fallback to first mapping document
			selected = m
		}
	}
	if selected == nil {
		// nothing decoded; fallback to text scan
		return analyzeWorkflowTextFallback(path)
	}
	wr := WorkflowReport{File: path}
	if n, _ := selected["name"].(string); n != "" {
		wr.Name = n
	}

	// triggers
	wr.Triggers = extractTriggers(selected["on"])
	// schedules
	wr.Schedules = extractSchedules(selected["on"])
	// runners
	wr.Runners = extractRunners(selected)
	// concurrency (workflow or job level)
	wr.HasConcurrency = hasConcurrency(selected)
	// actions pinning
	wr.UsesUnpinnedAction, wr.UnpinnedDetails = detectUnpinnedActions(selected)
	// deprecated hints
	wr.DeprecatedHints = detectDeprecated(wr)
	return wr, nil
}

var (
	reName       = regexp.MustCompile(`(?m)^\s*name:\s*(.+?)\s*$`)
	reCron       = regexp.MustCompile(`(?m)cron:\s*['"]?([^'"\n]+)['"]?`)
	reRunsOnLine = regexp.MustCompile(`(?m)^\s*runs-on:\s*(.+)$`)
	reRunsOnItem = regexp.MustCompile(`(?m)^\s*-\s*([\w\-\.]+)\s*$`)
	reUses       = regexp.MustCompile(`(?m)^\s*uses:\s*([^@\s]+)(?:@([^\s]+))?\s*$`)
)

func analyzeWorkflowTextFallback(path string) (WorkflowReport, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return WorkflowReport{}, err
	}
	s := string(b)
	wr := WorkflowReport{File: path}
	if m := reName.FindStringSubmatch(s); len(m) == 2 {
		wr.Name = strings.TrimSpace(m[1])
	}
	// heuristic triggers
	known := []string{"push", "pull_request", "pull_request_target", "workflow_dispatch", "schedule", "release", "workflow_call"}
	triggerSet := map[string]struct{}{}
	for _, k := range known {
		if regexp.MustCompile("(?m)^\\s*"+regexp.QuoteMeta(k)+"\\s*:").FindStringIndex(s) != nil {
			triggerSet[k] = struct{}{}
		}
	}
	for k := range triggerSet {
		wr.Triggers = append(wr.Triggers, k)
	}
	sort.Strings(wr.Triggers)
	// schedules
	for _, m := range reCron.FindAllStringSubmatch(s, -1) {
		if len(m) == 2 {
			wr.Schedules = append(wr.Schedules, strings.TrimSpace(m[1]))
		}
	}
	// runners (line form)
	set := map[string]struct{}{}
	for _, m := range reRunsOnLine.FindAllStringSubmatch(s, -1) {
		raw := strings.TrimSpace(m[1])
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
			raw = strings.Trim(raw, "[]")
			for _, tok := range strings.Split(raw, ",") {
				rr := strings.TrimSpace(strings.Trim(tok, "'\""))
				if rr != "" {
					set[rr] = struct{}{}
				}
			}
		} else {
			rr := strings.TrimSpace(strings.Trim(raw, "'\""))
			if rr != "" {
				set[rr] = struct{}{}
			}
		}
	}
	// list item form
	for _, m := range reRunsOnItem.FindAllStringSubmatch(s, -1) {
		rr := strings.TrimSpace(m[1])
		if rr != "" {
			set[rr] = struct{}{}
		}
	}
	for k := range set {
		wr.Runners = append(wr.Runners, k)
	}
	sort.Strings(wr.Runners)
	// concurrency presence
	wr.HasConcurrency = detectConcurrencyFallback(s)
	// unpinned uses
	var unp []string
	for _, m := range reUses.FindAllStringSubmatch(s, -1) {
		ref := strings.TrimSpace(m[2])
		usesVal := strings.TrimSpace(m[1])
		if isLocalAction(usesVal) {
			continue
		}
		if ref == "" {
			unp = append(unp, fmt.Sprintf("uses:%s@%s", usesVal, ref))
			continue
		}
		full := fmt.Sprintf("%s@%s", usesVal, ref)
		if unpinnedRe.MatchString(full) {
			unp = append(unp, fmt.Sprintf("uses:%s", full))
		}
	}
	if len(unp) > 0 {
		wr.UsesUnpinnedAction = true
		wr.UnpinnedDetails = unp
	}
	// hints
	wr.DeprecatedHints = detectDeprecated(wr)
	return wr, nil
}

func extractTriggers(on any) []string {
	var out []string
	switch v := on.(type) {
	case string:
		out = append(out, v)
	case []any:
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
	case map[string]any:
		for k := range v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func extractSchedules(on any) []string {
	var out []string
	m, ok := on.(map[string]any)
	if !ok {
		return out
	}
	if s, ok := m["schedule"].([]any); ok {
		for _, it := range s {
			if mm, ok := it.(map[string]any); ok {
				if cron, ok := mm["cron"].(string); ok {
					out = append(out, cron)
				}
			}
		}
	}
	return out
}

func extractRunners(root map[string]any) []string {
	set := map[string]struct{}{}
	jobs, ok := root["jobs"].(map[string]any)
	if !ok {
		return nil
	}
	for _, jv := range jobs {
		jm, ok := jv.(map[string]any)
		if !ok {
			continue
		}
		if r, ok := jm["runs-on"]; ok {
			switch rv := r.(type) {
			case string:
				set[rv] = struct{}{}
			case []any:
				for _, it := range rv {
					if s, ok := it.(string); ok {
						set[s] = struct{}{}
					}
				}
			}
		}
	}
	var out []string
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func hasConcurrency(root map[string]any) bool {
	if root == nil {
		return false
	}
	if root["concurrency"] != nil {
		return true
	}
	jobs, ok := root["jobs"].(map[string]any)
	if !ok {
		return false
	}
	for _, jv := range jobs {
		jm, ok := jv.(map[string]any)
		if !ok {
			continue
		}
		if jm["concurrency"] != nil {
			return true
		}
	}
	return false
}

func detectConcurrencyFallback(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if strings.HasPrefix(trim, "concurrency") {
			// match `concurrency:` optionally followed by value
			colon := strings.Index(trim, ":")
			if colon != -1 {
				key := strings.TrimSpace(trim[:colon])
				if key == "concurrency" {
					return true
				}
			}
		}
	}
	return false
}

var unpinnedRe = regexp.MustCompile(`^[^@]+@(main|master|HEAD|latest)$`)

func isLocalAction(u string) bool {
	trim := strings.TrimSpace(u)
	return strings.HasPrefix(trim, "./") || strings.HasPrefix(trim, "../") || strings.HasPrefix(trim, "docker://")
}

func detectUnpinnedActions(root map[string]any) (bool, []string) {
	var details []string
	flag := false
	jobs, ok := root["jobs"].(map[string]any)
	if !ok {
		return false, nil
	}
	for jname, jv := range jobs {
		jm, ok := jv.(map[string]any)
		if !ok {
			continue
		}
		steps, ok := jm["steps"].([]any)
		if !ok {
			continue
		}
		for _, sv := range steps {
			sm, ok := sv.(map[string]any)
			if !ok {
				continue
			}
			if u, ok := sm["uses"].(string); ok {
				if isLocalAction(u) {
					continue
				}
				// Not pinned if pointing to a mutable branch or tag
				if !strings.Contains(u, "@") || unpinnedRe.MatchString(u) {
					flag = true
					details = append(details, fmt.Sprintf("job:%s uses:%s", jname, u))
				}
			}
		}
	}
	return flag, details
}

func detectDeprecated(w WorkflowReport) []string {
	var hints []string
	for _, r := range w.Runners {
		if r == "ubuntu-22.04" {
			hints = append(hints, "Consider ubuntu-24.04")
		}
		if r == "macos-12" {
			hints = append(hints, "macos-12 deprecated; use macos-13/14/15")
		}
		if r == "self-hosted" {
			hints = append(hints, "Ensure self-hosted runner labels are specific; add timeouts/concurrency")
		}
	}
	if len(w.Schedules) > 5 {
		hints = append(hints, "Too many schedules; consider consolidation")
	}
	if len(w.Triggers) > 5 {
		hints = append(hints, "Many triggers; check for overlap with other workflows")
	}
	return hints
}

func recommendForWorkflow(w WorkflowReport, daysStale int) []string {
	var rec []string
	if w.LastModified != nil && time.Since(*w.LastModified) > (time.Duration(daysStale)*24*time.Hour) {
		rec = append(rec, "Stale: review necessity or update tooling pins")
	}
	if w.UsesUnpinnedAction {
		rec = append(rec, "Pin actions to specific tags or SHAs")
	}
	if !w.HasConcurrency {
		rec = append(rec, "Add 'concurrency' to avoid duplicate runs on busy repos")
	}
	if len(w.Runners) == 0 {
		rec = append(rec, "Specify runs-on for each job explicitly")
	}
	return rec
}

func gitLastModified(path string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct", "--", path)
	cmd.Dir = filepath.Dir(path)
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}
	secStr := strings.TrimSpace(string(out))
	sec, err := parseInt64(secStr)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

func parseInt64(s string) (int64, error) {
	var x int64
	_, err := fmt.Sscanf(s, "%d", &x)
	return x, err
}

// GitHub API minimal client
func enrichFromGitHub(owner, repo, token string, sampleRuns, daysStale int) (*GitHubReport, error) {
	base := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	cli := &http.Client{Timeout: 15 * time.Second}
	auth := "token " + token

	// Workflows list
	type ghWorkflow struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	type wfResp struct {
		Workflows []ghWorkflow `json:"workflows"`
	}
	var wf wfResp
	if err := ghGet(cli, base+"/actions/workflows", auth, &wf); err != nil {
		return nil, err
	}

	var failures []WorkflowFailure
	for _, w := range wf.Workflows {
		// Runs for each workflow
		type run struct {
			Conclusion string `json:"conclusion"`
		}
		type runsResp struct {
			WorkflowRuns []run `json:"workflow_runs"`
		}
		var rr runsResp
		url := fmt.Sprintf("%s/actions/workflows/%d/runs?per_page=%d", base, w.ID, sampleRuns)
		if err := ghGet(cli, url, auth, &rr); err != nil {
			continue
		}
		total := len(rr.WorkflowRuns)
		if total == 0 {
			continue
		}
		fails := 0
		recentFail := false
		for i, r := range rr.WorkflowRuns {
			if r.Conclusion == "failure" || r.Conclusion == "timed_out" || r.Conclusion == "cancelled" {
				fails++
				if i == 0 {
					recentFail = true
				}
			}
		}
		failures = append(failures, WorkflowFailure{Name: w.Name, WorkflowID: w.ID, SampledRuns: total, FailureRate: float64(fails) / float64(total), RecentFailure: recentFail})
	}

	// PRs
	type ghPR struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		Draft     bool      `json:"draft"`
		UpdatedAt time.Time `json:"updated_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	var prs []ghPR
	if err := ghGet(cli, base+"/pulls?state=open&per_page=100", auth, &prs); err != nil {
		return nil, err
	}
	var prReports []PRReport
	for _, p := range prs {
		stale := time.Since(p.UpdatedAt) > (time.Duration(daysStale) * 24 * time.Hour)
		prReports = append(prReports, PRReport{Number: p.Number, Title: p.Title, Author: p.User.Login, Draft: p.Draft, UpdatedAt: p.UpdatedAt, Stale: stale, HeadSHA: p.Head.SHA})
	}

	// Environments
	var envs struct {
		Environments []struct {
			Name string `json:"name"`
		} `json:"environments"`
	}
	if err := ghGet(cli, base+"/environments", auth, &envs); err != nil {
		return nil, err
	}
	var envReports []EnvironmentProbe
	for _, e := range envs.Environments {
		// deployments (most recent)
		var deps []struct {
			UpdatedAt time.Time `json:"updated_at"`
		}
		if err := ghGet(cli, base+"/deployments?per_page=1&environment="+urlQueryEscape(e.Name), auth, &deps); err != nil {
			continue
		}
		var last *time.Time
		if len(deps) > 0 {
			last = &deps[0].UpdatedAt
		}
		stale := true
		if last != nil {
			stale = time.Since(*last) > (time.Duration(daysStale) * 24 * time.Hour)
		}
		envReports = append(envReports, EnvironmentProbe{Name: e.Name, LastDeployed: last, IsStale: stale})
	}

	return &GitHubReport{Owner: owner, Repo: repo, WorkflowFailure: failures, PRs: prReports, Environments: envReports}, nil
}

func ghGet(cli *http.Client, url, auth string, v any) error {
	req, _ := http.NewRequest("GET", url, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return errors.New("resource not found: " + url)
	}
	if res.StatusCode == 401 {
		return errors.New("unauthorized: bad token or permissions")
	}
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("github %d: %s", res.StatusCode, string(b))
	}
	return json.NewDecoder(res.Body).Decode(v)
}

func urlQueryEscape(s string) string {
	// minimal escape for spaces -> %20 and plus signs -> %2B
	r := strings.NewReplacer(" ", "%20", "+", "%2B")
	return r.Replace(s)
}

func writeCleanupPlan(path string, r RepoDefragReport) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# CI Cleanup Plan\n\nGenerated: %s UTC\n\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "- Workflows scanned: %d\n- Stale threshold: %d days\n\n", r.Summary.WorkflowCount, r.StaleDays)

	// High level actions
	fmt.Fprintf(&buf, "## High-level actions\n\n")
	if r.Summary.WorkflowsWithUnpinned > 0 {
		fmt.Fprintf(&buf, "- Pin actions to tags or SHAs (found %d workflows)\n", r.Summary.WorkflowsWithUnpinned)
	}
	if r.Summary.WorkflowsWithoutConcurrency > 0 {
		fmt.Fprintf(&buf, "- Add concurrency to prevent duplicate runs (missing in %d workflows)\n", r.Summary.WorkflowsWithoutConcurrency)
	}
	if r.Summary.WorkflowsStale > 0 {
		fmt.Fprintf(&buf, "- Review or remove stale workflows (found %d)\n", r.Summary.WorkflowsStale)
	}
	fmt.Fprintln(&buf)

	fmt.Fprintf(&buf, "## Recommended snippets\n\n")
	fmt.Fprintf(&buf, "### Concurrency example\n\n")
	fmt.Fprintf(&buf, "```yaml\nconcurrency:\n  group: ${{ github.workflow }}-${{ github.ref }}\n  cancel-in-progress: true\n```\n\n")
	fmt.Fprintf(&buf, "### Actions pinning example\n\n")
	fmt.Fprintf(&buf, "```yaml\n- uses: actions/checkout@v4\n- uses: actions/setup-go@v5\n  with:\n    go-version: '1.22'\n```\n\n")

	// Per-workflow recommendations
	fmt.Fprintf(&buf, "## Workflow-specific recommendations\n\n")
	for _, w := range r.Workflows {
		if len(w.Recommendations) == 0 && len(w.DeprecatedHints) == 0 && !w.UsesUnpinnedAction && w.HasConcurrency {
			continue
		}
		fmt.Fprintf(&buf, "### %s\n\n", filepath.Base(w.File))
		if w.Name != "" {
			fmt.Fprintf(&buf, "- Name: %s\n", w.Name)
		}
		if len(w.Recommendations) > 0 {
			fmt.Fprintf(&buf, "- Recommendations: %s\n", strings.Join(w.Recommendations, "; "))
		}
		if len(w.DeprecatedHints) > 0 {
			fmt.Fprintf(&buf, "- Hints: %s\n", strings.Join(w.DeprecatedHints, "; "))
		}
		if w.UsesUnpinnedAction {
			fmt.Fprintf(&buf, "- Unpinned steps: %s\n", strings.Join(w.UnpinnedDetails, "; "))
		}
		if !w.HasConcurrency {
			fmt.Fprintf(&buf, "- Add 'concurrency' block (see snippet above)\n")
		}
		fmt.Fprintln(&buf)
	}

	if r.GitHub != nil {
		fmt.Fprintf(&buf, "## Pull Requests (staleness > %d days)\n\n", r.StaleDays)
		for _, pr := range r.GitHub.PRs {
			if pr.Stale {
				fmt.Fprintf(&buf, "- #%d %s by %s (updated %s)\n", pr.Number, pr.Title, pr.Author, pr.UpdatedAt.Format("2006-01-02"))
			}
		}
		fmt.Fprintln(&buf)
		fmt.Fprintf(&buf, "## Environments\n\n")
		for _, env := range r.GitHub.Environments {
			lm := "never"
			if env.LastDeployed != nil {
				lm = env.LastDeployed.Format("2006-01-02")
			}
			stale := ""
			if env.IsStale {
				stale = " (stale)"
			}
			fmt.Fprintf(&buf, "- %s: last deployment %s%s\n", env.Name, lm, stale)
		}
		fmt.Fprintln(&buf)
	}

	return os.WriteFile(path, buf.Bytes(), 0o644)
}
