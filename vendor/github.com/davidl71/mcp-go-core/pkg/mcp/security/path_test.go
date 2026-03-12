package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetProjectRoot(t *testing.T) {
	// Get the actual project root (where this test is running)
	expectedRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("Failed to get expected root: %v", err)
	}

	// Test from current directory
	root, err := GetProjectRoot(".")
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}

	// Normalize paths for comparison
	expectedRoot = filepath.Clean(expectedRoot)
	root = filepath.Clean(root)

	if root != expectedRoot {
		t.Errorf("GetProjectRoot() = %v, want %v", root, expectedRoot)
	}
}

func TestGetProjectRootFromSubdirectory(t *testing.T) {
	// Test from a subdirectory
	root, err := GetProjectRoot("pkg/mcp/security")
	if err != nil {
		t.Fatalf("GetProjectRoot failed: %v", err)
	}

	// Should still find the project root
	expectedRoot, _ := filepath.Abs("../../..")
	expectedRoot = filepath.Clean(expectedRoot)
	root = filepath.Clean(root)

	if root != expectedRoot {
		t.Errorf("GetProjectRoot() = %v, want %v", root, expectedRoot)
	}
}

func TestValidatePath(t *testing.T) {
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
			name:        "valid relative path",
			path:        "subdir",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "valid absolute path within root",
			path:        filepath.Join(projectRoot, "subdir"),
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "directory traversal attempt",
			path:        "../outside",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "directory traversal prevention (../../../etc/passwd)",
			path:        "../../../etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "absolute path outside root",
			path:        "/etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "absolute path outside root (portable)",
			path:        filepath.Join(tmpDir, "other"),
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "path with .. in middle",
			path:        "subdir/../..",
			projectRoot: projectRoot,
			wantErr:     true,
		},
		{
			name:        "empty path",
			path:        "",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "current directory",
			path:        ".",
			projectRoot: projectRoot,
			wantErr:     false,
		},
		{
			name:        "empty project root",
			path:        "subdir",
			projectRoot: "",
			wantErr:     true,
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
				if filepath.IsAbs(relPath) || strings.HasPrefix(relPath, "..") {
					t.Errorf("ValidatePath() returned path outside root: %s", relPath)
				}
			}
		})
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
		{"existing file", "test.txt", projectRoot, false},
		{"non-existing file", "nonexistent.txt", projectRoot, true},
		{"directory traversal", "../../etc/passwd", projectRoot, true},
		{"directory traversal (../../../etc/passwd)", "../../../etc/passwd", projectRoot, true},
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

// TestValidatePath_EdgeCases tests T-287: empty paths, special characters, redundant separators.
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
		{"redundant slashes", "subdir//file", projectRoot, false},
		{"dot segment in path", "subdir/./file", projectRoot, false},
		{"trailing slash", "subdir/", projectRoot, false},
		{"dot-dot resolved within root", "subdir/../subdir", projectRoot, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePath(tt.path, tt.projectRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// symlinkSupported skips on Windows where os.Symlink may require privileges.
func symlinkSupported(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows (os.Symlink may require privileges)")
	}
}

// TestValidatePathSymlink_WithinProject tests Case 1: symlink within project → target within project.
func TestValidatePathSymlink_WithinProject(t *testing.T) {
	symlinkSupported(t)
	tmpDir := t.TempDir()
	projectRoot := filepath.Join(tmpDir, "project")
	os.MkdirAll(filepath.Join(projectRoot, "subdir"), 0755)
	realFile := filepath.Join(projectRoot, "real_file.txt")
	os.WriteFile(realFile, []byte("data"), 0644)
	linkPath := filepath.Join(projectRoot, "subdir", "link_in")
	if err := os.Symlink(realFile, linkPath); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}
	_, err := ValidatePath("subdir/link_in", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(symlink within→within) error = %v, want nil", err)
	}
	_, err = ValidatePathExists("subdir/link_in", projectRoot)
	if err != nil {
		t.Errorf("ValidatePathExists(symlink within→within) error = %v, want nil", err)
	}
}

// TestValidatePathSymlink_OutsideProject tests Case 2: symlink within project → target outside root.
// Current behavior: ValidatePath accepts (path string within root; no EvalSymlinks).
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
	_, err := ValidatePath("evil_link", projectRoot)
	if err != nil {
		t.Errorf("ValidatePath(symlink to outside) current behavior: error = %v, want nil", err)
	}
}

// TestValidatePathExistsSymlink_Broken tests Case 3: broken symlink.
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
		t.Error("ValidatePathExists(broken symlink) want error, got nil")
	}
}

// TestValidatePathSymlink_Nested tests Case 4: symlink in path component (a/link → b/file).
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
