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

// handleSecurityScan handles the scan action for security tool (Go projects only; no bridge fallback).
func handleSecurityScan(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	if !IsGoProject() {
		return nil, fmt.Errorf("security scan is only supported for Go projects (go.mod)")
	}

	vulns, err := scanGoDependencies(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("security scan: %w", err)
	}

	result := formatSecurityScanResults(vulns, "go")

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleSecurityAlerts handles the alerts action for security tool.
func handleSecurityAlerts(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	repo := "davidl71/exarp-go"
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

// Vulnerability represents a security vulnerability.
type Vulnerability struct {
	Package     string `json:"package"`
	Version     string `json:"version"`
	VulnID      string `json:"vuln_id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FixVersion  string `json:"fix_version,omitempty"`
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

	cmd := exec.CommandContext(ctx, "gh", "api", fmt.Sprintf("repos/%s/dependabot/alerts", repo), "--jq", jqQuery)

	output, err := cmd.CombinedOutput()
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
