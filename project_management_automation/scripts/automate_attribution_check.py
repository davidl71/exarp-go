#!/usr/bin/env python3
"""
Automated Attribution Compliance Check Script

Intelligently checks attribution compliance across the codebase using
IntelligentAutomationBase with comprehensive scanning.
"""

import argparse
import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Optional

from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

logger = logging.getLogger(__name__)


class AttributionComplianceChecker(IntelligentAutomationBase):
    """Intelligent attribution compliance checker using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        super().__init__(config, "Attribution Compliance Check", project_root)
        
        self.attributions_file = self.project_root / "ATTRIBUTIONS.md"
        self.results = {
            "attribution_score": 100.0,
            "issues": [],
            "warnings": [],
            "compliant_files": [],
            "missing_attribution": [],
            "patterns_found": [],
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is attribution compliance?"""
        return "What is attribution compliance? Attribution Compliance = File Headers × Central Documentation × Dependency Licensing × External Service Documentation × Concept Attribution"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we check attribution compliance?"""
        return "How do we systematically verify attribution compliance across all code files, documentation, and dependencies?"

    def _execute_analysis(self) -> dict:
        """Execute attribution compliance analysis."""
        logger.info("Executing attribution compliance check...")
        
        # Check ATTRIBUTIONS.md
        self._check_attributions_file()
        
        # Check Git-inspired files
        self._check_git_inspired_files()
        
        # Check wisdom source files
        self._check_wisdom_files()
        
        # Check dependencies
        self._check_dependencies()
        
        # Check README
        self._check_readme()
        
        # Scan for potential third-party references
        self._scan_for_third_party_references()
        
        # Calculate final score
        self.results["attribution_score"] = max(0.0, min(100.0, self.results["attribution_score"]))
        
        # Determine status
        if self.results["attribution_score"] >= 90:
            self.results["status"] = "compliant"
        elif self.results["attribution_score"] >= 70:
            self.results["status"] = "mostly_compliant"
        else:
            self.results["status"] = "needs_attention"
        
        return self.results

    def _check_attributions_file(self) -> None:
        """Check ATTRIBUTIONS.md exists and is complete."""
        if not self.attributions_file.exists():
            self.results["issues"].append({
                "type": "missing_file",
                "severity": "high",
                "file": "ATTRIBUTIONS.md",
                "message": "ATTRIBUTIONS.md file not found in project root"
            })
            self.results["attribution_score"] -= 20.0
            return
        
        content = self.attributions_file.read_text()
        
        # Check for key sections
        required_sections = ["GitTask", "External Services", "Wisdom Sources"]
        found_sections = []
        
        for section in required_sections:
            if section in content:
                found_sections.append(section)
                self.results["compliant_files"].append(f"ATTRIBUTIONS.md: {section} documented")
            else:
                self.results["warnings"].append({
                    "type": "missing_section",
                    "file": "ATTRIBUTIONS.md",
                    "section": section,
                    "message": f"ATTRIBUTIONS.md missing '{section}' section"
                })
                self.results["attribution_score"] -= 3.0

    def _check_git_inspired_files(self) -> None:
        """Check Git-inspired files have attribution headers."""
        git_files = [
            "project_management_automation/utils/commit_tracking.py",
            "project_management_automation/utils/branch_utils.py",
            "project_management_automation/tools/task_diff.py",
            "project_management_automation/tools/git_graph.py",
            "project_management_automation/tools/branch_merge.py",
            "project_management_automation/tools/git_inspired_tools.py",
        ]
        
        for file_path in git_files:
            full_path = self.project_root / file_path
            if not full_path.exists():
                continue
            
            content = full_path.read_text()
            
            # Check for attribution patterns
            has_attribution = (
                "GitTask" in content or
                "ATTRIBUTIONS.md" in content or
                "inspired by" in content.lower() or
                re.search(r'attribution|credit', content, re.IGNORECASE)
            )
            
            if has_attribution:
                self.results["compliant_files"].append(f"{file_path}: Attribution header present")
            else:
                self.results["missing_attribution"].append({
                    "file": file_path,
                    "type": "missing_header_attribution",
                    "severity": "medium",
                    "message": "Git-inspired file missing attribution header"
                })
                self.results["attribution_score"] -= 5.0

    def _check_wisdom_files(self) -> None:
        """Check wisdom source files have attribution."""
        wisdom_files = [
            "project_management_automation/tools/wisdom/sefaria.py",
            "project_management_automation/tools/wisdom/pistis_sophia.py",
            "project_management_automation/tools/wisdom/sources.py",
        ]
        
        for file_path in wisdom_files:
            full_path = self.project_root / file_path
            if not full_path.exists():
                continue
            
            content = full_path.read_text()
            
            # Check for attribution patterns
            has_attribution = (
                "sefaria.org" in content.lower() or
                "pistis sophia" in content.lower() or
                "sacred-texts.com" in content.lower() or
                "attribution" in content.lower() or
                "credit" in content.lower()
            )
            
            if has_attribution:
                self.results["compliant_files"].append(f"{file_path}: Attribution present")
            else:
                self.results["warnings"].append({
                    "type": "attribution_check",
                    "file": file_path,
                    "message": "Wisdom file may need attribution verification"
                })

    def _check_dependencies(self) -> None:
        """Check dependencies have license information."""
        pyproject_file = self.project_root / "pyproject.toml"
        if pyproject_file.exists():
            content = pyproject_file.read_text()
            if "license" in content.lower():
                self.results["compliant_files"].append("pyproject.toml: License field present")
            else:
                self.results["warnings"].append({
                    "type": "license_field",
                    "file": "pyproject.toml",
                    "message": "pyproject.toml may be missing license field"
                })

    def _check_readme(self) -> None:
        """Check README has attribution section."""
        readme_file = self.project_root / "README.md"
        if readme_file.exists():
            content = readme_file.read_text()
            if "attribution" in content.lower() or "ATTRIBUTIONS.md" in content:
                self.results["compliant_files"].append("README.md: Attribution section present")
            else:
                self.results["warnings"].append({
                    "type": "readme_attribution",
                    "file": "README.md",
                    "message": "README.md may be missing attribution section"
                })

    def _scan_for_third_party_references(self) -> None:
        """Scan for potential third-party references that might need attribution."""
        # Look for common patterns that suggest third-party code/concepts
        patterns = [
            r'inspired by|based on|adapted from|credit to|attribution',
            r'github\.com/[^/\s]+/[^/\s]+',  # GitHub repo references
            r'https?://[^\s]+',  # URLs that might be sources
        ]
        
        # Scan Python files in project_management_automation
        code_path = self.project_root / "project_management_automation"
        if code_path.exists():
            for py_file in code_path.rglob("*.py"):
                try:
                    content = py_file.read_text()
                    # Skip test files for this scan
                    if "test" in str(py_file):
                        continue
                    
                    for pattern in patterns:
                        matches = re.finditer(pattern, content, re.IGNORECASE)
                        for match in matches:
                            # Check if it's in a comment/docstring
                            context = content[max(0, match.start()-50):match.end()+50]
                            if any(marker in context for marker in ['"""', "'''", '#']):
                                self.results["patterns_found"].append({
                                    "file": str(py_file.relative_to(self.project_root)),
                                    "pattern": pattern,
                                    "match": match.group()[:50]
                                })
                                break
                except Exception as e:
                    logger.debug(f"Error scanning {py_file}: {e}")

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from attribution compliance analysis."""
        insights = []
        
        score = analysis_results.get('attribution_score', 0)
        status = analysis_results.get('status', 'unknown')
        
        insights.append(f"**Attribution Compliance Score: {score:.1f}/100**")
        insights.append(f"**Status**: {status.replace('_', ' ').title()}")
        
        issues_count = len(analysis_results.get('issues', []))
        missing_count = len(analysis_results.get('missing_attribution', []))
        warnings_count = len(analysis_results.get('warnings', []))
        compliant_count = len(analysis_results.get('compliant_files', []))
        
        if issues_count > 0:
            insights.append(f"⚠️ {issues_count} compliance issues found")
        
        if missing_count > 0:
            insights.append(f"⚠️ {missing_count} files missing attribution headers")
        
        if warnings_count > 0:
            insights.append(f"ℹ️ {warnings_count} warnings to review")
        
        if compliant_count > 0:
            insights.append(f"✅ {compliant_count} files are compliant")
        
        if score >= 90:
            insights.append("✅ Attribution compliance is excellent!")
        elif score >= 70:
            insights.append("⚠️ Attribution compliance needs improvement")
        else:
            insights.append("❌ Attribution compliance requires immediate attention")
        
        return "\n".join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate compliance report."""
        lines = []
        lines.append("# Attribution Compliance Report")
        lines.append("")
        lines.append(f"**Generated**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append("**Generated By**: Intelligent Attribution Compliance Checker")
        lines.append("")
        lines.append("---")
        lines.append("")
        
        # Add insights section
        lines.append("## Insights")
        lines.append("")
        lines.append(insights)
        lines.append("")
        lines.append("---")
        lines.append("")
        
        score = analysis_results.get('attribution_score', 0)
        analysis_results.get('status', 'unknown')
        
        lines.append("## Summary")
        lines.append("")
        lines.append(f"- **Compliant Files**: {len(analysis_results.get('compliant_files', []))}")
        lines.append(f"- **Issues Found**: {len(analysis_results.get('issues', []))}")
        lines.append(f"- **Warnings**: {len(analysis_results.get('warnings', []))}")
        lines.append(f"- **Missing Attribution**: {len(analysis_results.get('missing_attribution', []))}")
        lines.append("")
        
        if analysis_results.get("compliant_files"):
            lines.append("## ✅ Compliant Files")
            lines.append("")
            for file in analysis_results["compliant_files"]:
                lines.append(f"- ✅ {file}")
            lines.append("")
        
        if analysis_results.get("issues"):
            lines.append("## ❌ Issues")
            lines.append("")
            for issue in analysis_results["issues"]:
                severity = issue.get("severity", "unknown").upper()
                lines.append(f"### {severity}: {issue.get('type', 'unknown')}")
                lines.append("")
                lines.append(f"- **File**: `{issue.get('file', 'N/A')}`")
                lines.append(f"- **Message**: {issue.get('message', 'N/A')}")
                lines.append("")
        
        if analysis_results.get("missing_attribution"):
            lines.append("## ⚠️ Missing Attribution")
            lines.append("")
            for item in analysis_results["missing_attribution"]:
                lines.append(f"- **{item['file']}**: {item['message']}")
            lines.append("")
        
        if analysis_results.get("warnings"):
            lines.append("## ⚠️ Warnings")
            lines.append("")
            for warning in analysis_results["warnings"]:
                lines.append(f"- **{warning.get('file', 'N/A')}**: {warning.get('message', 'N/A')}")
            lines.append("")
        
        lines.append("## Recommendations")
        lines.append("")
        if score < 90:
            lines.append("1. Review and fix all high-severity issues")
            lines.append("2. Add missing attribution headers to identified files")
            lines.append("3. Verify ATTRIBUTIONS.md is complete and up to date")
            lines.append("4. Ensure README.md includes attribution section")
        else:
            lines.append("✅ Attribution compliance is excellent!")
            lines.append("Continue maintaining proper attribution as new features are added.")
        lines.append("")
        
        lines.append("---")
        lines.append("")
        lines.append("*This report was generated using intelligent automation with Tractatus Thinking, Sequential Thinking, and NetworkX analysis.*")
        
        return "\n".join(lines)

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify followup tasks for attribution issues."""
        tasks = []
        
        high_severity = [i for i in analysis_results.get("issues", []) if i.get("severity") == "high"]
        missing_attribution = analysis_results.get("missing_attribution", [])
        
        if high_severity:
            tasks.append({
                "name": "Fix high-severity attribution compliance issues",
                "long_description": f"Address {len(high_severity)} high-severity attribution compliance issues:\n\n" + 
                    "\n".join([f"- {i.get('file', 'N/A')}: {i.get('message', 'N/A')}" for i in high_severity]),
                "status": "todo",
                "priority": "high",
                "tags": ["attribution", "compliance", "legal"],
            })
        
        if missing_attribution:
            tasks.append({
                "name": "Add missing attribution headers",
                "long_description": f"Add attribution headers to {len(missing_attribution)} files:\n\n" +
                    "\n".join([f"- {item['file']}" for item in missing_attribution]),
                "status": "todo",
                "priority": "medium",
                "tags": ["attribution", "compliance"],
            })
        
        return tasks


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Check attribution compliance")
    parser.add_argument("--output-path", help="Path for report output")
    parser.add_argument("--create-tasks", action="store_true", default=True, help="Create Todo2 tasks")
    parser.add_argument("--no-create-tasks", dest="create_tasks", action="store_false", help="Don't create tasks")
    
    args = parser.parse_args()
    
    config = {
        "output_path": args.output_path or "docs/ATTRIBUTION_COMPLIANCE_REPORT.md",
        "create_tasks": args.create_tasks,
    }
    
    checker = AttributionComplianceChecker(config)
    checker.run()
    
    print(f"\nAttribution Compliance Score: {checker.results['attribution_score']:.1f}/100")
    print(f"Status: {checker.results['status'].replace('_', ' ').title()}")
    print(f"Report: {config['output_path']}")


if __name__ == "__main__":
    main()
