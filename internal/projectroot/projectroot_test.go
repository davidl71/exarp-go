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

	oldProjectRoot := os.Getenv("PROJECT_ROOT")
	os.Unsetenv("PROJECT_ROOT")
	defer os.Setenv("PROJECT_ROOT", oldProjectRoot)

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

func TestFindPrefersConsumerWhenEnvPointsAtExarpSource(t *testing.T) {
	consumer := t.TempDir()
	if err := os.MkdirAll(filepath.Join(consumer, ".todo2"), 0755); err != nil {
		t.Fatalf("consumer .todo2: %v", err)
	}
	exarpSrc := t.TempDir()
	if err := os.WriteFile(filepath.Join(exarpSrc, "go.mod"), []byte("module github.com/davidl71/exarp-go\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(exarpSrc, "cmd", "server"), 0755); err != nil {
		t.Fatalf("cmd/server: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(exarpSrc, ".todo2"), 0755); err != nil {
		t.Fatalf("exarp .todo2: %v", err)
	}

	sub := filepath.Join(consumer, "apps", "svc")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("subdir: %v", err)
	}

	origWd, _ := os.Getwd()
	if err := os.Chdir(sub); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origWd)

	oldEnv := os.Getenv("PROJECT_ROOT")
	defer os.Setenv("PROJECT_ROOT", oldEnv)
	os.Setenv("PROJECT_ROOT", exarpSrc)

	root, err := Find()
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	want, _ := filepath.EvalSymlinks(consumer)
	got, _ := filepath.EvalSymlinks(root)
	if got != want {
		t.Fatalf("Find() = %s, want consumer root %s", root, consumer)
	}
}

func TestFindStrictProjectRootSkipsConsumerPreference(t *testing.T) {
	consumer := t.TempDir()
	if err := os.MkdirAll(filepath.Join(consumer, ".todo2"), 0755); err != nil {
		t.Fatalf("consumer .todo2: %v", err)
	}
	exarpSrc := t.TempDir()
	if err := os.WriteFile(filepath.Join(exarpSrc, "go.mod"), []byte("module github.com/davidl71/exarp-go\n\ngo 1.25\n"), 0644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(exarpSrc, "cmd", "server"), 0755); err != nil {
		t.Fatalf("cmd/server: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(exarpSrc, ".todo2"), 0755); err != nil {
		t.Fatalf("exarp .todo2: %v", err)
	}
	sub := filepath.Join(consumer, "x")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("subdir: %v", err)
	}

	origWd, _ := os.Getwd()
	if err := os.Chdir(sub); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origWd)

	oldEnv := os.Getenv("PROJECT_ROOT")
	oldStrict := os.Getenv("EXARP_STRICT_PROJECT_ROOT")
	defer func() {
		os.Setenv("PROJECT_ROOT", oldEnv)
		os.Setenv("EXARP_STRICT_PROJECT_ROOT", oldStrict)
	}()
	os.Setenv("PROJECT_ROOT", exarpSrc)
	os.Setenv("EXARP_STRICT_PROJECT_ROOT", "1")

	root, err := Find()
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	want, _ := filepath.EvalSymlinks(exarpSrc)
	got, _ := filepath.EvalSymlinks(root)
	if got != want {
		t.Fatalf("Find() = %s, want PROJECT_ROOT %s", root, exarpSrc)
	}
}
