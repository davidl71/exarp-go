#!/usr/bin/env python3
"""
Security Features Detection for Exarp Scorecard
This file helps the Python scorecard checker detect Go security implementations
"""

import os
import json
from pathlib import Path

def detect_security_features(project_root):
    """Detect security features implemented in Go"""
    features = {
        "path_boundaries": False,
        "rate_limiting": False,
        "access_control": False
    }
    
    # Check for path boundary enforcement
    path_file = Path(project_root) / "internal" / "security" / "path.go"
    if path_file.exists():
        content = path_file.read_text()
        if "func ValidatePath" in content and "ValidatePathWithinRoot" in content:
            features["path_boundaries"] = True
    
    # Check for rate limiting
    ratelimit_file = Path(project_root) / "internal" / "security" / "ratelimit.go"
    if ratelimit_file.exists():
        content = ratelimit_file.read_text()
        if "type RateLimiter" in content and "func Allow" in content:
            features["rate_limiting"] = True
    
    # Check for access control
    access_file = Path(project_root) / "internal" / "security" / "access.go"
    if access_file.exists():
        content = access_file.read_text()
        if "type AccessControl" in content and "func CheckToolAccess" in content:
            features["access_control"] = True
    
    return features

def detect_git_hooks(project_root):
    """Detect git hooks"""
    hooks = {
        "pre_commit_hook": False,
        "pre_push_hook": False,
        "post_commit_hook": False,
        "post_merge_hook": False
    }
    
    git_hooks_dir = Path(project_root) / ".git" / "hooks"
    
    if (git_hooks_dir / "pre-commit").exists():
        hook = git_hooks_dir / "pre-commit"
        if hook.is_file() and os.access(hook, os.X_OK):
            hooks["pre_commit_hook"] = True
    
    if (git_hooks_dir / "pre-push").exists():
        hook = git_hooks_dir / "pre-push"
        if hook.is_file() and os.access(hook, os.X_OK):
            hooks["pre_push_hook"] = True
    
    if (git_hooks_dir / "post-commit").exists():
        hook = git_hooks_dir / "post-commit"
        if hook.is_file() and os.access(hook, os.X_OK):
            hooks["post_commit_hook"] = True
    
    if (git_hooks_dir / "post-merge").exists():
        hook = git_hooks_dir / "post-merge"
        if hook.is_file() and os.access(hook, os.X_OK):
            hooks["post_merge_hook"] = True
    
    return hooks

if __name__ == "__main__":
    project_root = os.getenv("PROJECT_ROOT", ".")
    security = detect_security_features(project_root)
    hooks = detect_git_hooks(project_root)
    
    result = {
        "security_features": security,
        "git_hooks": hooks
    }
    
    print(json.dumps(result, indent=2))

