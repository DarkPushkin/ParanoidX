// Package isle provides the Isle manifest and build versioning system
package isle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// TestBuildManifest handles the TestBuildManifest HTTP request.
func TestBuildManifest(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.mp3")
	content := make([]byte, PieceSize*3+100) // 3 full pieces + partial
	for i := range content {
		content[i] = byte(i % 256)
	}
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	m, err := BuildManifest("track-001", filePath, "Test Track", "mp3")
	if err != nil {
		t.Fatalf("BuildManifest: %v", err)
	}

	if m.TrackID != "track-001" {
		t.Fatalf("expected track-001, got %s", m.TrackID)
	}
	if m.Title != "Test Track" {
		t.Fatalf("expected 'Test Track', got '%s'", m.Title)
	}
	if m.Kind != "mp3" {
		t.Fatalf("expected mp3, got %s", m.Kind)
	}
	if m.Size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), m.Size)
	}
	if m.PieceSize != PieceSize {
		t.Fatalf("expected piece size %d, got %d", PieceSize, m.PieceSize)
	}
	// 3 full pieces + 1 partial
	if len(m.Pieces) != 4 {
		t.Fatalf("expected 4 pieces, got %d", len(m.Pieces))
	}
	for i, p := range m.Pieces {
		if len(p) != 64 {
			t.Fatalf("piece %d: expected 64-char hex hash, got %d chars", i, len(p))
		}
	}
}


// TestBuildManifestEmptyFile handles the TestBuildManifestEmptyFile HTTP request.
func TestBuildManifestEmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.mp3")
	os.WriteFile(filePath, []byte{}, 0644)

	m, err := BuildManifest("empty", filePath, "Empty", "mp3")
	if err != nil {
		t.Fatalf("BuildManifest empty: %v", err)
	}
	if m.Size != 0 {
		t.Fatalf("expected size 0, got %d", m.Size)
	}
	if len(m.Pieces) != 0 {
		t.Fatalf("expected 0 pieces, got %d", len(m.Pieces))
	}
}


// TestBuildManifestNonexistentFile handles the TestBuildManifestNonexistentFile HTTP request.
func TestBuildManifestNonexistentFile(t *testing.T) {
	_, err := BuildManifest("x", "/nonexistent/file.mp3", "X", "mp3")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}


// TestSaveAndLoadManifest handles the TestSaveAndLoadManifest HTTP request.
func TestSaveAndLoadManifest(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.mp3")
	os.WriteFile(filePath, []byte("test audio content"), 0644)

	m, err := BuildManifest("save-test", filePath, "Save Test", "ogg")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	manifestPath := filepath.Join(dir, "save-test.isle")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Fatal("manifest file was not created")
	}

	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}

	if loaded.TrackID != "save-test" {
		t.Fatalf("expected track-id save-test, got %s", loaded.TrackID)
	}
	if loaded.Title != "Save Test" {
		t.Fatalf("expected 'Save Test', got '%s'", loaded.Title)
	}
	if loaded.Size != 18 {
		t.Fatalf("expected size 18, got %d", loaded.Size)
	}
	if len(loaded.Pieces) != 1 {
		t.Fatalf("expected 1 piece, got %d", len(loaded.Pieces))
	}
}


// TestLoadManifestNonexistent handles the TestLoadManifestNonexistent HTTP request.
func TestLoadManifestNonexistent(t *testing.T) {
	_, err := LoadManifest("/nonexistent/file.isle")
	if err == nil {
		t.Fatal("expected error for nonexistent manifest")
	}
}


// TestVerifyManifest handles the TestVerifyManifest HTTP request.
func TestVerifyManifest(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "verify.mp3")
	content := []byte(strings.Repeat("hello world ", 1000))
	os.WriteFile(filePath, content, 0644)

	m, err := BuildManifest("verify-test", filePath, "Verify Test", "mp3")
	if err != nil {
		t.Fatal(err)
	}

	if !m.Verify(filePath) {
		t.Fatal("expected verification to pass")
	}
}


// TestVerifyManifestCorrupted handles the TestVerifyManifestCorrupted HTTP request.
func TestVerifyManifestCorrupted(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "corrupt.mp3")
	os.WriteFile(filePath, []byte("original content"), 0644)

	m, err := BuildManifest("corrupt-test", filePath, "Corrupt", "mp3")
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the file
	os.WriteFile(filePath, []byte("corrupted content"), 0644)

	if m.Verify(filePath) {
		t.Fatal("expected verification to fail for corrupted file")
	}
}


// TestVerifyManifestNonexistent handles the TestVerifyManifestNonexistent HTTP request.
func TestVerifyManifestNonexistent(t *testing.T) {
	m := &Manifest{TrackID: "x", PieceSize: PieceSize}
	if m.Verify("/nonexistent/file.mp3") {
		t.Fatal("expected verification to fail for nonexistent file")
	}
}


// TestDuplicateBuildsProduceSameHashes handles the TestDuplicateBuildsProduceSameHashes HTTP request.
func TestDuplicateBuildsProduceSameHashes(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "dup.mp3")
	os.WriteFile(filePath, []byte("deterministic test data for hash comparison"), 0644)

	m1, _ := BuildManifest("d1", filePath, "Dup", "mp3")
	m2, _ := BuildManifest("d2", filePath, "Dup", "mp3")

	if len(m1.Pieces) != len(m2.Pieces) {
		t.Fatal("expected same number of pieces")
	}
	for i := range m1.Pieces {
		if m1.Pieces[i] != m2.Pieces[i] {
			t.Fatalf("piece %d hash mismatch", i)
		}
	}
}
