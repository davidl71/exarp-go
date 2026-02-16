package projectroot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFind(t *testing.T) {
	tmpDir := t.TempDir()

	todo2Dir := filepath.Join(tmpDir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("create .todo2: %v", err)
	}

	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	root, err := Find()
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}

	rootNorm, _ := filepath.EvalSymlinks(root)
	tmpNorm, _ := filepath.EvalSymlinks(tmpDir)

	if rootNorm != tmpNorm {
		t.Errorf("Find() = %s, want %s", root, tmpDir)
	}
}

func TestFindFrom(t *testing.T) {
	tmpDir := t.TempDir()

	exarpDir := filepath.Join(tmpDir, ".exarp")
	if err := os.MkdirAll(exarpDir, 0755); err != nil {
		t.Fatalf("create .exarp: %v", err)
	}

	subFile := filepath.Join(tmpDir, "sub", "file.txt")
	if err := os.MkdirAll(filepath.Dir(subFile), 0755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	if err := os.WriteFile(subFile, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	root, err := FindFrom(subFile)
	if err != nil {
		t.Fatalf("FindFrom(%s) error = %v", subFile, err)
	}

	rootNorm, _ := filepath.EvalSymlinks(root)
	tmpNorm, _ := filepath.EvalSymlinks(tmpDir)

	if rootNorm != tmpNorm {
		t.Errorf("FindFrom(%s) = %s, want %s", subFile, root, tmpDir)
	}
}

func TestFindGoMod(t *testing.T) {
	tmpDir := t.TempDir()

	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	subdir := filepath.Join(projectRoot, "subdir", "nested")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	root, err := FindGoMod(subdir)
	if err != nil {
		t.Fatalf("FindGoMod(%s) error = %v", subdir, err)
	}

	rootNorm, _ := filepath.EvalSymlinks(root)
	projNorm, _ := filepath.EvalSymlinks(projectRoot)

	if rootNorm != projNorm {
		t.Errorf("FindGoMod(%s) = %s, want %s", subdir, root, projectRoot)
	}
}

func TestFindFromWithMarkers(t *testing.T) {
	tmpDir := t.TempDir()

	customMarker := filepath.Join(tmpDir, "my.marker")
	if err := os.WriteFile(customMarker, []byte("x"), 0644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	subdir := filepath.Join(tmpDir, "a", "b")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	root, err := FindFromWithMarkers(subdir, []string{"my.marker"})
	if err != nil {
		t.Fatalf("FindFromWithMarkers error = %v", err)
	}

	rootNorm, _ := filepath.EvalSymlinks(root)
	tmpNorm, _ := filepath.EvalSymlinks(tmpDir)

	if rootNorm != tmpNorm {
		t.Errorf("FindFromWithMarkers = %s, want %s", root, tmpDir)
	}
}
