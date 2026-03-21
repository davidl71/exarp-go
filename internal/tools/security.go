// security.go — MCP "security" tool: vulnerability scanning and security checks.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

var runDependabotAlertsCommand = func(ctx context.Context, repo, jqQuery string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", fmt.Sprintf("repos/%s/dependabot/alerts", repo), "--jq", jqQuery)
	return cmd.CombinedOutput()
}

// handleSecurityScan handles the scan action for security tool (Go, Python, Rust, Node).
// Runs language-specific dependency scanners for each detected ecosystem and aggregates results.
func handleSecurityScan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	var allVulns []Vulnerability
	var ecosystems []string
	var ranAny bool

	// Go: govulncheck
	if DetectGoProject(projectRoot) {
		vulns, err := scanGoDependencies(ctx, projectRoot)
		if err == nil {
			ranAny = true
			for _, v := range vulns {
				v.Ecosystem = "go"
				allVulns = append(allVulns, v)
			}
			ecosystems = append(ecosystems, "go")
		}
	}

	// Python: pip-audit or safety
	if DetectPythonProject(projectRoot) {
		_, pythonRoot := DetectPythonProjectRoot(projectRoot)
		scanDir := filepath.Join(projectRoot, pythonRoot)
		vulns, err := scanPythonDependencies(ctx, scanDir)
		if err == nil {
			ranAny = true
			for _, v := range vulns {
				v.Ecosystem = "python"
				allVulns = append(allVulns, v)
			}
			ecosystems = append(ecosystems, "python")
		}
	}

	// Rust: cargo audit
	if DetectRustProject(projectRoot) {
		_, rustRoot := DetectRustProjectRoot(projectRoot)
		scanDir := filepath.Join(projectRoot, rustRoot)
		vulns, err := scanRustDependencies(ctx, scanDir)
		if err == nil {
			ranAny = true
			for _, v := range vulns {
				v.Ecosystem = "rust"
				allVulns = append(allVulns, v)
			}
			ecosystems = append(ecosystems, "rust")
		}
	}

	// Node/TypeScript: npm audit
	if DetectTypeScriptProject(projectRoot) {
		_, tsRoot := DetectTypeScriptProjectRoot(projectRoot)
		scanDir := filepath.Join(projectRoot, tsRoot)
		vulns, err := scanNpmDependencies(ctx, scanDir)
		if err == nil {
			ranAny = true
			for _, v := range vulns {
				v.Ecosystem = "npm"
				allVulns = append(allVulns, v)
			}
			ecosystems = append(ecosystems, "npm")
		}
	}

	if !ranAny {
		msg := "No supported dependency scanner ran for this project."
		if len(ecosystems) == 0 {
			msg = "No Go, Python, Rust, or Node project detected in project root. Use semgrep or CodeQL for static analysis."
		}
		result := formatSecurityScanResults(allVulns, "multilang")
		result = msg + "\n\n" + result
		return []framework.TextContent{{Type: "text", Text: result}}, nil
	}

	ecosystemLabel := strings.Join(ecosystems, ", ")
	result := formatSecurityScanResults(allVulns, ecosystemLabel)
	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// getGitHubRepoFromRemote returns the "owner/repo" from the current git remote URL.
// Falls back to empty string if no remote or not a GitHub URL.
func getGitHubRepoFromRemote(ctx context.Context, projectRoot string) string {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	url := strings.TrimSpace(string(output))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		return url
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
		url = strings.TrimSuffix(url, ".git")
		return url
	}

	return ""
}

// handleSecurityAlerts handles the alerts action for security tool.
func handleSecurityAlerts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Default to current git repo, not hardcoded exarp-go
	repo := ""
	if projectRoot, err := FindProjectRoot(); err == nil {
		repo = getGitHubRepoFromRemote(ctx, projectRoot)
	}
	// Override with explicit repo parameter if provided
	if r, ok := params["repo"].(string); ok && r != "" {
		repo = r
	}

	state := "open"
	if s, ok := params["state"].(string); ok && s != "" {
		state = s
	}

	alerts, err := fetchDependabotAlerts(ctx, repo, state)
	if err != nil {
		return nil, fmt.Errorf("security alerts: %w", err)
	}

	result := formatDependabotAlerts(alerts)

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSecurityReport handles the report action for security tool.
func handleSecurityReport(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get scan results
	scanParams := map[string]interface{}{"action": "scan"}
	scanResult, scanErr := handleSecurityScan(ctx, scanParams)

	// Get alerts
	alertsParams := map[string]interface{}{
		"action": params["action"],
		"repo":   params["repo"],
		"state":  params["state"],
	}
	alertsResult, alertsErr := handleSecurityAlerts(ctx, alertsParams)

	// Combine results
	var report strings.Builder

	report.WriteString("======================================================================\n")
	report.WriteString("  SECURITY REPORT\n")
	report.WriteString("======================================================================\n\n")

	if scanErr == nil && scanResult != nil {
		report.WriteString("## Dependency Scan Results\n\n")
		report.WriteString(scanResult[0].Text)
		report.WriteString("\n")
	}

	if alertsErr == nil && alertsResult != nil {
		report.WriteString("## Dependabot Alerts\n\n")
		report.WriteString(alertsResult[0].Text)
		report.WriteString("\n")
	}

	if scanErr != nil && alertsErr != nil {
		return nil, fmt.Errorf("security report: scan and alerts both failed (scan: %w; alerts: %w)", scanErr, alertsErr)
	}

	return []framework.TextContent{
		{Type: "text", Text: report.String()},
	}, nil
}

// scanGoDependencies scans Go dependencies for vulnerabilities.
func scanGoDependencies(ctx context.Context, projectRoot string) ([]Vulnerability, error) {
	vulns := []Vulnerability{}

	// Use govulncheck if available (Go 1.18+)
	cmd := exec.CommandContext(ctx, "go", "version")
	if err := cmd.Run(); err == nil {
		// Try govulncheck
		cmd = exec.CommandContext(ctx, "govulncheck", "./...")
		cmd.Dir = projectRoot

		output, err := cmd.CombinedOutput()
		if err == nil {
			// Parse govulncheck output
			parsed := parseGovulncheckOutput(string(output))
			vulns = append(vulns, parsed...)
		}
	}

	// Also check go.mod for known vulnerable packages
	// This is a simplified check - in production, would query vulnerability DB
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Could parse go.mod and check against vulnerability database
		// For now, just return any govulncheck findings
	}

	return vulns, nil
}

// scanPythonDependencies runs pip-audit (or safety) in dir and returns vulnerabilities.
func scanPythonDependencies(ctx context.Context, dir string) ([]Vulnerability, error) {
	// Prefer pip-audit (recommended by PyPA)
	cmd := exec.CommandContext(ctx, "pip-audit", "--format", "json")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// pip-audit exits 1 when vulns found; still parse output
		if len(output) == 0 {
			return nil, fmt.Errorf("pip-audit failed: %w", err)
		}
	}
	return parsePipAuditOutput(string(output))
}

func parsePipAuditOutput(output string) ([]Vulnerability, error) {
	var vulns []Vulnerability
	var result struct {
		Dependencies []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Vulns   []struct {
				ID          string `json:"id"`
				FixVersions []struct {
					ID string `json:"id"`
				} `json:"fix_versions"`
				Description string `json:"description"`
			} `json:"vulns"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, err
	}
	for _, dep := range result.Dependencies {
		for _, v := range dep.Vulns {
			fixVer := ""
			if len(v.FixVersions) > 0 {
				fixVer = v.FixVersions[0].ID
			}
			vulns = append(vulns, Vulnerability{
				Package:     dep.Name,
				Version:     dep.Version,
				VulnID:      v.ID,
				Description: v.Description,
				FixVersion:  fixVer,
				Severity:    "unknown",
			})
		}
	}
	return vulns, nil
}

// scanRustDependencies runs cargo audit in dir and returns vulnerabilities.
func scanRustDependencies(ctx context.Context, dir string) ([]Vulnerability, error) {
	cmd := exec.CommandContext(ctx, "cargo", "audit", "--json")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return nil, fmt.Errorf("cargo audit failed: %w", err)
		}
	}
	return parseCargoAuditOutput(string(output))
}

func parseCargoAuditOutput(output string) ([]Vulnerability, error) {
	var vulns []Vulnerability
	// cargo-audit --json: top-level "vulnerabilities" object with "list" array
	var result struct {
		Vulnerabilities struct {
			List []struct {
				ID          string `json:"id"`
				Package     string `json:"package"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Versions    struct {
					Patched []string `json:"patched"`
				} `json:"versions"`
			} `json:"list"`
		} `json:"vulnerabilities"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, err
	}
	for _, v := range result.Vulnerabilities.List {
		fixVer := ""
		if len(v.Versions.Patched) > 0 {
			fixVer = v.Versions.Patched[0]
		}
		vulns = append(vulns, Vulnerability{
			Package:     v.Package,
			VulnID:      v.ID,
			Description: v.Title + ": " + v.Description,
			FixVersion:  fixVer,
			Severity:    "unknown",
		})
	}
	return vulns, nil
}

// scanNpmDependencies runs npm audit --json in dir and returns vulnerabilities.
func scanNpmDependencies(ctx context.Context, dir string) ([]Vulnerability, error) {
	cmd := exec.CommandContext(ctx, "npm", "audit", "--json")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		// npm audit exits non-zero when vulns exist; still parse
		if len(output) == 0 {
			return nil, fmt.Errorf("npm audit failed: %w", err)
		}
	}
	return parseNpmAuditOutput(string(output))
}

func parseNpmAuditOutput(output string) ([]Vulnerability, error) {
	var vulns []Vulnerability
	var result struct {
		Vulnerabilities map[string]struct {
			Severity     string        `json:"severity"`
			Via          []interface{} `json:"via"`
			Range        string        `json:"range"`
			FixAvailable interface{}   `json:"fixAvailable"`
		} `json:"vulnerabilities"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, err
	}
	for pkg, v := range result.Vulnerabilities {
		desc := pkg
		if len(v.Via) > 0 {
			if m, ok := v.Via[0].(map[string]interface{}); ok {
				if id, _ := m["id"].(string); id != "" {
					desc = id
				}
			}
		}
		fixVer := ""
		if s, ok := v.FixAvailable.(string); ok && s != "false" {
			fixVer = s
		}
		vulns = append(vulns, Vulnerability{
			Package:     pkg,
			VulnID:      desc,
			Severity:    v.Severity,
			FixVersion:  fixVer,
			Description: v.Range,
		})
	}
	return vulns, nil
}

// Vulnerability represents a security vulnerability.
type Vulnerability struct {
	Package     string `json:"package"`
	Version     string `json:"version"`
	VulnID      string `json:"vuln_id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FixVersion  string `json:"fix_version,omitempty"`
	Ecosystem   string `json:"ecosystem,omitempty"` // go, python, rust, npm
}

// parseGovulncheckOutput parses govulncheck output.
func parseGovulncheckOutput(output string) []Vulnerability {
	vulns := []Vulnerability{}
	// Simplified parsing - govulncheck output format may vary
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Vulnerability") || strings.Contains(line, "CVE") {
			// Basic parsing - would need more sophisticated parsing in production
			vulns = append(vulns, Vulnerability{
				Description: line,
				Severity:    "unknown",
			})
		}
	}

	return vulns
}

// formatSecurityScanResults formats scan results as text.
func formatSecurityScanResults(vulns []Vulnerability, ecosystem string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Security Scan Results (%s)\n", ecosystem))
	sb.WriteString(fmt.Sprintf("Total Vulnerabilities: %d\n\n", len(vulns)))

	if len(vulns) == 0 {
		sb.WriteString("✅ No vulnerabilities found\n")
		return sb.String()
	}

	for i, vuln := range vulns {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, vuln.Package))
		if vuln.Ecosystem != "" {
			sb.WriteString(fmt.Sprintf("   Ecosystem: %s\n", vuln.Ecosystem))
		}

		if vuln.Version != "" {
			sb.WriteString(fmt.Sprintf("   Version: %s\n", vuln.Version))
		}

		if vuln.VulnID != "" {
			sb.WriteString(fmt.Sprintf("   CVE: %s\n", vuln.VulnID))
		}

		if vuln.Severity != "" {
			sb.WriteString(fmt.Sprintf("   Severity: %s\n", vuln.Severity))
		}

		if vuln.Description != "" {
			sb.WriteString(fmt.Sprintf("   Description: %s\n", vuln.Description))
		}

		if vuln.FixVersion != "" {
			sb.WriteString(fmt.Sprintf("   Fix: Upgrade to %s\n", vuln.FixVersion))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// DependabotAlert represents a Dependabot alert.
type DependabotAlert struct {
	Package      string `json:"package"`
	Severity     string `json:"severity"`
	CVE          string `json:"cve"`
	State        string `json:"state"`
	Ecosystem    string `json:"ecosystem"`
	Description  string `json:"description"`
	FixAvailable bool   `json:"fix_available"`
	FixVersion   string `json:"fixed_version,omitempty"`
}

// fetchDependabotAlerts fetches Dependabot alerts using gh CLI.
func fetchDependabotAlerts(ctx context.Context, repo, state string) ([]DependabotAlert, error) {
	// Use gh CLI to fetch alerts (same approach as Python)
	jqQuery := `.[] | {package: .security_vulnerability.package.name, severity: .security_vulnerability.severity, cve: .security_advisory.cve_id, state: .state, ecosystem: .security_vulnerability.package.ecosystem, description: .security_advisory.summary, fix_available: .security_vulnerability.first_patched_version != null, fixed_version: .security_vulnerability.first_patched_version.identifier}`

	output, err := runDependabotAlertsCommand(ctx, repo, jqQuery)
	if err != nil {
		return nil, fmt.Errorf("gh CLI failed: %w, output: %s", err, output)
	}

	// Parse JSONL output
	alerts := []DependabotAlert{}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var alert DependabotAlert
		if err := json.Unmarshal([]byte(line), &alert); err == nil {
			// Filter by state if not "all"
			if state == "all" || alert.State == state {
				alerts = append(alerts, alert)
			}
		}
	}

	return alerts, nil
}

// formatDependabotAlerts formats Dependabot alerts as text.
func formatDependabotAlerts(alerts []DependabotAlert) string {
	var sb strings.Builder

	sb.WriteString("Dependabot Alerts\n")
	sb.WriteString(fmt.Sprintf("Total Alerts: %d\n\n", len(alerts)))

	if len(alerts) == 0 {
		sb.WriteString("✅ No open alerts\n")
		return sb.String()
	}

	// Count by severity
	bySeverity := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, alert := range alerts {
		sev := strings.ToLower(alert.Severity)
		if count, ok := bySeverity[sev]; ok {
			bySeverity[sev] = count + 1
		}
	}

	sb.WriteString("By Severity:\n")

	for sev, count := range bySeverity {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", sev, count))
		}
	}

	sb.WriteString("\n")

	// List alerts
	for i, alert := range alerts {
		if i >= 20 { // Limit to first 20
			sb.WriteString(fmt.Sprintf("\n... and %d more alerts\n", len(alerts)-20))
			break
		}

		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, alert.Package, alert.Ecosystem))

		if alert.CVE != "" {
			sb.WriteString(fmt.Sprintf("   CVE: %s\n", alert.CVE))
		}

		sb.WriteString(fmt.Sprintf("   Severity: %s\n", alert.Severity))

		if alert.FixAvailable && alert.FixVersion != "" {
			sb.WriteString(fmt.Sprintf("   Fix: Upgrade to %s\n", alert.FixVersion))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
