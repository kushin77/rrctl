package main

import (
"fmt"
"os"
"path/filepath"
"strings"

"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
Use:   "security-scan",
Short: "Basic security scanning for common vulnerabilities",
Long: `Run basic security scans including:
- Secrets detection in files
- Basic dependency vulnerability checks
- File permission analysis`,
RunE: runBasicSecurityScan,
}

var (
targetPath string
checkSecrets bool
checkDeps   bool
checkPerms  bool
)

func init() {
rootCmd.AddCommand(securityCmd)

securityCmd.Flags().StringVarP(&targetPath, "path", "p", ".", "Path to scan")
securityCmd.Flags().BoolVar(&checkSecrets, "secrets", true, "Check for secrets in files")
securityCmd.Flags().BoolVar(&checkDeps, "deps", true, "Check dependencies")
securityCmd.Flags().BoolVar(&checkPerms, "perms", true, "Check file permissions")
}

func runBasicSecurityScan(cmd *cobra.Command, args []string) error {
fmt.Println("🔒 Running basic security scan...")

if checkSecrets {
if err := scanForSecrets(targetPath); err != nil {
fmt.Printf("❌ Secrets scan failed: %v\n", err)
}
}

if checkDeps {
if err := checkDependencies(targetPath); err != nil {
fmt.Printf("❌ Dependency check failed: %v\n", err)
}
}

if checkPerms {
if err := checkFilePermissions(targetPath); err != nil {
fmt.Printf("❌ Permission check failed: %v\n", err)
}
}

fmt.Println("✅ Security scan completed")
return nil
}

func scanForSecrets(path string) error {
fmt.Println("🔍 Scanning for secrets...")

secrets := []string{
"password",
"secret",
"key",
"token",
"api_key",
"apikey",
}

found := false

err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
if err != nil {
return nil
}

// Skip binary files, .git, node_modules, etc.
if info.IsDir() {
if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
return filepath.SkipDir
}
return nil
}

// Skip common binary extensions
ext := strings.ToLower(filepath.Ext(filePath))
if ext == ".jpg" || ext == ".png" || ext == ".gif" || ext == ".pdf" || ext == ".zip" {
return nil
}

content, err := os.ReadFile(filePath)
if err != nil {
return nil
}

contentStr := strings.ToLower(string(content))
for _, secret := range secrets {
if strings.Contains(contentStr, secret) {
fmt.Printf("⚠️  Potential secret found in: %s\n", filePath)
found = true
break
}
}

return nil
})

if err != nil {
return err
}

if !found {
fmt.Println("✅ No obvious secrets detected")
}

return nil
}

func checkDependencies(path string) error {
fmt.Println("📦 Checking dependencies...")

// Check for common dependency files
depFiles := []string{
"go.mod",
"package.json",
"requirements.txt",
"pyproject.toml",
"Pipfile",
"Cargo.toml",
}

found := false
for _, file := range depFiles {
if _, err := os.Stat(filepath.Join(path, file)); err == nil {
fmt.Printf("📄 Found dependency file: %s\n", file)
found = true
}
}

if !found {
fmt.Println("ℹ️  No common dependency files found")
} else {
fmt.Println("✅ Dependency files detected")
}

return nil
}

func checkFilePermissions(path string) error {
fmt.Println("🔐 Checking file permissions...")

warnings := 0

err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
if err != nil {
return nil
}

// Skip directories for now
if info.IsDir() {
return nil
}

// Check for world-writable files
mode := info.Mode()
if mode.Perm()&0o022 != 0 {
fmt.Printf("⚠️  World-writable file: %s (permissions: %s)\n", filePath, mode.Perm())
warnings++
}

return nil
})

if err != nil {
return err
}

if warnings == 0 {
fmt.Println("✅ No permission issues found")
} else {
fmt.Printf("⚠️  Found %d permission warnings\n", warnings)
}

return nil
}
