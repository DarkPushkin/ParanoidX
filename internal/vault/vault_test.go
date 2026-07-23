// Package vault provides encryption and secure storage
package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// TestNew handles the TestNew HTTP request.
func TestNew(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if s.Path == "" {
		t.Fatal("expected non-empty path")
	}
	if _, err := os.Stat(s.Path); err != nil {
		t.Fatalf("expected vault dir to exist: %v", err)
	}
}


// TestFileCount handles the TestFileCount HTTP request.
func TestFileCount(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	if s.FileCount() != 0 {
		t.Fatalf("expected 0 files, got %d", s.FileCount())
	}

	os.WriteFile(filepath.Join(s.Path, "test.txt"), []byte("hello"), 0600)
	if s.FileCount() != 1 {
		t.Fatalf("expected 1 file, got %d", s.FileCount())
	}
}


// TestUploadAndList handles the TestUploadAndList HTTP request.
func TestUploadAndList(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	name, err := s.Upload("hello.txt", strings.NewReader("world"), 5)
	if err != nil {
		t.Fatal(err)
	}
	if name != "hello.txt" {
		t.Fatalf("expected hello.txt, got %s", name)
	}

	files := s.List()
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].Name != "hello.txt" {
		t.Fatalf("expected hello.txt, got %s", files[0].Name)
	}
	if files[0].Size != 5 {
		t.Fatalf("expected size 5, got %d", files[0].Size)
	}
}


// TestDownload handles the TestDownload HTTP request.
func TestDownload(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	os.WriteFile(filepath.Join(s.Path, "dl.txt"), []byte("download"), 0600)
	path := s.Download("dl.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "download" {
		t.Fatalf("expected 'download', got '%s'", string(b))
	}
}


// TestDelete handles the TestDelete HTTP request.
func TestDelete(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	os.WriteFile(filepath.Join(s.Path, "del.txt"), []byte("delete"), 0600)
	if err := s.Delete("del.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Path, "del.txt")); !os.IsNotExist(err) {
		t.Fatal("expected file to be deleted")
	}
}


// TestUploadSanitizesPath handles the TestUploadSanitizesPath HTTP request.
func TestUploadSanitizesPath(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	name, err := s.Upload("../outside.txt", strings.NewReader("safe"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if name != "outside.txt" {
		t.Fatalf("expected outside.txt, got %s", name)
	}
}


// TestSizeMB handles the TestSizeMB HTTP request.
func TestSizeMB(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	os.WriteFile(filepath.Join(s.Path, "a.dat"), make([]byte, 1024*1024), 0600)
	os.WriteFile(filepath.Join(s.Path, "b.dat"), make([]byte, 512*1024), 0600)

	sz := s.SizeMB()
	if sz < 1.4 || sz > 1.6 {
		t.Fatalf("expected ~1.5 MB, got %f", sz)
	}
}


// TestUploadExceedsQuota handles the TestUploadExceedsQuota HTTP request.
func TestUploadExceedsQuota(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	// Fill vault close to quota (sparse file, no RAM allocation)
	fillSize := int64(QuotaMB) * 1024 * 1024
	f, _ := os.Create(filepath.Join(s.Path, "big.dat"))
	f.Truncate(fillSize)
	f.Close()

	_, err := s.Upload("overflow.txt", strings.NewReader("x"), 1)
	if err == nil {
		t.Fatal("expected error when exceeding quota")
	}
}


// TestSaveNote handles the TestSaveNote HTTP request.
func TestSaveNote(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	name, err := s.SaveNote("mynote.txt", "note content")
	if err != nil {
		t.Fatal(err)
	}
	if name != "mynote.txt" {
		t.Fatalf("expected mynote.txt, got %s", name)
	}

	b, err := os.ReadFile(filepath.Join(s.Path, name))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "note content" {
		t.Fatalf("expected 'note content', got '%s'", string(b))
	}
}


// TestSaveNoteAutoName handles the TestSaveNoteAutoName HTTP request.
func TestSaveNoteAutoName(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	name, err := s.SaveNote("", "auto named")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(name, "note-") {
		t.Fatalf("expected note- prefix, got %s", name)
	}
	if !strings.HasSuffix(name, ".txt") {
		t.Fatalf("expected .txt suffix, got %s", name)
	}
}


// TestEmptyList handles the TestEmptyList HTTP request.
func TestEmptyList(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	files := s.List()
	if len(files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(files))
	}
}
