package security

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidatePath(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	os.MkdirAll(filepath.Join(projectRoot, "subdir"), 0755)

	tests := []struct {
		name        string
		path        string
		projectRoot string
		wantErr     bool
		description string
	}{
		{
			name:        "valid relative path",
			path:        "subdir",
			projectRoot: projectRoot,
			wantErr:     false,
			description: "Valid path within project root",
		},
		{
			name:        "valid absolute path within root",
			path:        filepath.Join(projectRoot, "subdir"),
			projectRoot: projectRoot,
			wantErr:     false,
			description: "Valid absolute path within root",
		},
		{
			name:        "directory traversal attempt",
			path:        "../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject directory traversal",
		},
		{
			name:        "directory traversal prevention (../../../etc/passwd)",
			path:        "../../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject directory traversal with multiple .. segments",
		},
		{
			name:        "absolute path outside root",
			path:        "/etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject absolute path outside root",
		},
		{
			name:        "absolute path outside root (portable)",
			path:        filepath.Join(tmpDir, "other"),
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject absolute path outside root (sibling dir, OS-portable)",
		},
		{
			name:        "path with .. in middle",
			path:        "subdir/../..",
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject path that escapes root",
		},
		{
			name:        "empty path",
			path:        "",
			projectRoot: projectRoot,
			wantErr:     false,
			description: "Empty path should resolve to project root",
		},
		{
			name:        "current directory",
			path:        ".",
			projectRoot: projectRoot,
			wantErr:     false,
			description: "Current directory should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := ValidatePath(tt.path, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the path is actually within project root
				relPath, relErr := filepath.Rel(tt.projectRoot, absPath)
				if relErr != nil {
					t.Errorf("ValidatePath() returned path outside root: %v", relErr)
				}

				if filepath.IsAbs(relPath) || filepath.HasPrefix(relPath, "..") {
					t.Errorf("ValidatePath() returned path outside root: %s", relPath)
				}
			}
		})
	}
}

// TestValidationValidPath ensures validation accepts a valid path within project root (T-1768320773738).
func TestValidationValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	os.MkdirAll(filepath.Join(projectRoot, "subdir"), 0755)

	absPath, err := ValidatePath("subdir", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(valid path) error = %v", err)
	}

	if absPath == "" {
		t.Error("ValidatePath(valid path) returned empty path")
	}
}

func TestValidatePathExists(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	testFile := filepath.Join(projectRoot, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	tests := []struct {
		name        string
		path        string
		projectRoot string
		wantErr     bool
	}{
		{
			name:        "existing file",
			path:        "test.txt",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "non-existing file",
			path:        "nonexistent.txt",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "directory traversal",
			path:        "../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "directory traversal prevention (../../../etc/passwd)",
			path:        "../../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePathExists(tt.path, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module test"), 0644)

	subdir := filepath.Join(projectRoot, "subdir", "nested")
	os.MkdirAll(subdir, 0755)

	root, err := GetProjectRoot(subdir)
	if err != nil {
		t.Errorf("GetProjectRoot() error = %v", err)
	}

	if root != projectRoot {
		t.Errorf("GetProjectRoot() = %v, want %v", root, projectRoot)
	}
}

// symlinkSupported reports whether os.Symlink can be used (Unix; on Windows may need privileges).
func symlinkSupported(t *testing.T) bool {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows (os.Symlink may require privileges)")
		return false
	}

	return true
}

// TestValidatePathSymlink_WithinProject tests Case 1: symlink within project → target within project.
// ValidatePath and ValidatePathExists should accept (path string within root; target exists).
func TestValidatePathSymlink_WithinProject(t *testing.T) {
	symlinkSupported(t)
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	os.MkdirAll(filepath.Join(projectRoot, "subdir"), 0755)
	realFile := filepath.Join(projectRoot, "real_file.txt")
	os.WriteFile(realFile, []byte("data"), 0644)

	linkPath := filepath.Join(projectRoot, "subdir", "link_in")
	if err := os.Symlink(realFile, linkPath); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	// ValidatePath: accept
	_, err := ValidatePath("subdir/link_in", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(symlink within→within) error = %v, want nil", err)
	}
	// ValidatePathExists: accept (target exists)
	_, err = ValidatePathExists("subdir/link_in", projectRoot)
	if err != nil {
		t.Errorf("ValidatePathExists(symlink within→within) error = %v, want nil", err)
	}
}

// TestValidatePathSymlink_OutsideProject tests Case 2: symlink within project → target outside root.
// Current behavior: ValidatePath accepts (path string is within root; no EvalSymlinks).
func TestValidatePathSymlink_OutsideProject(t *testing.T) {
	symlinkSupported(t)
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	outsideDir := filepath.Join(tmpDir, "outside")

	os.MkdirAll(projectRoot, 0755)
	os.MkdirAll(outsideDir, 0755)
	secret := filepath.Join(outsideDir, "secret")
	os.WriteFile(secret, []byte("secret"), 0644)

	evilLink := filepath.Join(projectRoot, "evil_link")
	if err := os.Symlink(secret, evilLink); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	// Current behavior: ValidatePath accepts (does not resolve symlink)
	_, err := ValidatePath("evil_link", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(symlink to outside) current behavior: error = %v, want nil (path string within root)", err)
	}
}

// TestValidatePathExistsSymlink_Broken tests Case 3: broken symlink.
// ValidatePath accepts (path string within root); ValidatePathExists rejects (path does not exist).
func TestValidatePathExistsSymlink_Broken(t *testing.T) {
	symlinkSupported(t)
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)

	brokenLink := filepath.Join(projectRoot, "broken_link")
	if err := os.Symlink("nonexistent_target", brokenLink); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	_, err := ValidatePath("broken_link", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(broken symlink) error = %v, want nil", err)
	}

	_, err = ValidatePathExists("broken_link", projectRoot)
	if err == nil {
		t.Error("ValidatePathExists(broken symlink) want error (path does not exist), got nil")
	}
}

// TestValidatePathSymlink_Nested tests Case 4: symlink in path component (a/link → b/file).
// ValidatePath and ValidatePathExists should accept when target is within root.
func TestValidatePathSymlink_Nested(t *testing.T) {
	symlinkSupported(t)
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	dirA := filepath.Join(projectRoot, "a")
	dirB := filepath.Join(projectRoot, "b")

	os.MkdirAll(dirA, 0755)
	os.MkdirAll(dirB, 0755)
	targetFile := filepath.Join(dirB, "file")
	os.WriteFile(targetFile, []byte("data"), 0644)

	linkPath := filepath.Join(dirA, "link")
	if err := os.Symlink(targetFile, linkPath); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	_, err := ValidatePath("a/link", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(nested symlink) error = %v, want nil", err)
	}

	_, err = ValidatePathExists("a/link", projectRoot)
	if err != nil {
		t.Errorf("ValidatePathExists(nested symlink) error = %v, want nil", err)
	}
}

// TestValidatePath_EdgeCases tests T-287: edge cases (redundant slashes, dot segments, special chars).
func TestValidatePath_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(projectRoot, 0755)
	os.MkdirAll(filepath.Join(projectRoot, "subdir"), 0755)

	tests := []struct {
		name        string
		path        string
		projectRoot string
		wantErr     bool
	}{
		{
			name:        "redundant slashes",
			path:        "subdir//file",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "dot segment in path",
			path:        "subdir/./file",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "trailing slash",
			path:        "subdir/",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "dot-dot resolved within root",
			path:        "subdir/../subdir",
			projectRoot: projectRoot,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := ValidatePath(tt.path, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				relPath, relErr := filepath.Rel(tt.projectRoot, absPath)
				if relErr != nil {
					t.Errorf("ValidatePath() returned path outside root: %v", relErr)
				}

				if filepath.IsAbs(relPath) || filepath.HasPrefix(relPath, "..") {
					t.Errorf("ValidatePath() returned path outside root: %s", relPath)
				}
			}
		})
	}
}
