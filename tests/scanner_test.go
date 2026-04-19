package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/novelli-mo/cura/scanner"
)

func TestScanRepo_CollectsExtensions(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "utils.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	ctx, err := scanner.ScanRepo(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ctx.Extensions[".go"] != 2 {
		t.Errorf("expected 2 .go files, got %d", ctx.Extensions[".go"])
	}
	if ctx.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", ctx.TotalFiles)
	}
}

func TestScanRepo_ReadsReadme(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("This is a Go CLI tool"), 0644)

	ctx, err := scanner.ScanRepo(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(ctx.DocContent, "Go CLI tool") {
		t.Error("expected README content in DocContent")
	}
}
