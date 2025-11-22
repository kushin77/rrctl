package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "v1.1.0"
	commit  = "open-source"
	date    = "2024"
	builtBy = "community"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "rrctl",
	Short:   "RoundRobin Control - Open Source DevOps CLI",
	Long:    `rrctl is an open-source CLI tool for DevOps automation and security scanning.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rrctl %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s (%s)\n", date, builtBy)
		fmt.Printf("Open Source Edition\n")
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|powershell|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for rrctl.

To load completions:

Bash:
  $ source <(rrctl completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ rrctl completion bash > /etc/bash_completion.d/rrctl
  # macOS:
  $ rrctl completion bash > /usr/local/etc/bash_completion.d/rrctl

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ rrctl completion zsh > "${fpath[1]}/_rrctl"

  # You will need to start a new shell for this setup to take effect.

PowerShell:
  PS> rrctl completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> rrctl completion powershell > rrctl.ps1
  # and source this file from your PowerShell profile.

Fish:
  $ rrctl completion fish | source

  # To load completions for each session, execute once:
  $ rrctl completion fish > ~/.config/fish/completions/rrctl.fish
`,
	ValidArgs: []string{"bash", "zsh", "powershell", "fish"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
	// repo-defrag command is registered in repo_defrag.go's init()
	// repo-autofix command is registered in repo_autofix.go's init()
}
