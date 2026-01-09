package security

import (
	"os"
	"path/filepath"
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
			name:        "absolute path outside root",
			path:        "/etc/passwd",
			projectRoot: projectRoot,
			wantErr:     true,
			description: "Should reject absolute path outside root",
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
