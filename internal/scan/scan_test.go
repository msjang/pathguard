package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanDetectsNFDOver(t *testing.T) {
	root := t.TempDir()
	// A short, safe name.
	write(t, filepath.Join(root, "정상.txt"))
	// 30 syllables with a final consonant → NFD 9 bytes each = 270B > 255.
	long := strings.Repeat("강", 30) + ".txt"
	write(t, filepath.Join(root, long))

	res, err := Scan(root, "/remote/prefix", Limits{NameMax: 255, PathMax: 4096, WarnRatio: 0.80}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total < 2 {
		t.Fatalf("total = %d, want >= 2", res.Total)
	}
	if len(res.NameOver) < 1 {
		t.Fatalf("NameOver = %d, want >= 1", len(res.NameOver))
	}
	if got := res.NameOver[0].NameNFD; got != 270+len(".txt") {
		t.Fatalf("worst NameNFD = %d, want %d", got, 270+len(".txt"))
	}
}

func TestScanExcludesDirs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, ".git", "inside.txt"))
	write(t, filepath.Join(root, "keep.txt"))

	lim := Limits{NameMax: 255, PathMax: 4096, WarnRatio: 0.80}

	with, _ := Scan(root, "/r", lim, map[string]bool{".git": true})
	if with.Total != 1 {
		t.Fatalf("with exclude: total = %d, want 1 (keep.txt only)", with.Total)
	}
	without, _ := Scan(root, "/r", lim, nil)
	if without.Total != 3 {
		t.Fatalf("without exclude: total = %d, want 3 (.git, inside.txt, keep.txt)", without.Total)
	}
}
