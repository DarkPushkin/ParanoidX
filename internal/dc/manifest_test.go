// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)


// TestManifestCreateAndVerify handles the TestManifestCreateAndVerify HTTP request.
func TestManifestCreateAndVerify(t *testing.T) {
	dir := t.TempDir()

	content := []byte("hello dc cloud manifest test")
	testFile := filepath.Join(dir, "test_container.bin")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	manifest, err := BuildManifest("container-1", testFile)
	if err != nil {
		t.Fatalf("BuildManifest error: %v", err)
	}

	if manifest.Infohash == "" {
		t.Fatal("expected non-empty infohash")
	}
	if manifest.Size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), manifest.Size)
	}
	if manifest.ContainerID != "container-1" {
		t.Fatalf("expected container-1, got %s", manifest.ContainerID)
	}

	// Verify pieces
	if len(manifest.Pieces) == 0 {
		t.Fatal("expected at least 1 piece")
	}
	if manifest.PieceCount != len(manifest.Pieces) {
		t.Fatalf("piece count mismatch: %d vs %d", manifest.PieceCount, len(manifest.Pieces))
	}

	// Verify first piece hash
	expectedHash := sha256.Sum256(content)
	expectedHex := fmt.Sprintf("%x", expectedHash)
	if manifest.Pieces[0] != expectedHex {
		t.Fatalf("piece 0 hash mismatch: expected %s, got %s", expectedHex, manifest.Pieces[0])
	}

	// Verify infohash computed correctly
	if !manifest.VerifyAll(testFile) {
		t.Fatal("VerifyAll should pass for original file")
	}

	// VerifyPiece for index 0
	if !manifest.VerifyPiece(0, content) {
		t.Fatal("VerifyPiece(0) should pass")
	}
}


// TestManifestVerifyCorrupted handles the TestManifestVerifyCorrupted HTTP request.
func TestManifestVerifyCorrupted(t *testing.T) {
	dir := t.TempDir()

	content := []byte("verify me please")
	testFile := filepath.Join(dir, "verify.bin")
	os.WriteFile(testFile, content, 0644)

	manifest, err := BuildManifest("verify-container", testFile)
	if err != nil {
		t.Fatalf("BuildManifest error: %v", err)
	}

	if !manifest.VerifyAll(testFile) {
		t.Fatal("VerifyAll should pass for original")
	}

	corrupted := []byte("corrupted data!!!!")
	corruptedFile := filepath.Join(dir, "corrupted.bin")
	os.WriteFile(corruptedFile, corrupted, 0644)

	if manifest.VerifyAll(corruptedFile) {
		t.Fatal("VerifyAll should fail for corrupted file")
	}
}


// TestManifestSaveLoad handles the TestManifestSaveLoad HTTP request.
func TestManifestSaveLoad(t *testing.T) {
	dir := t.TempDir()

	content := []byte("save and load test")
	testFile := filepath.Join(dir, "container.bin")
	os.WriteFile(testFile, content, 0644)

	manifest, err := BuildManifest("test-container", testFile)
	if err != nil {
		t.Fatalf("BuildManifest error: %v", err)
	}

	// Save using dir (creates infohash.dc in dir)
	if err := manifest.Save(dir); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Load manifest
	savedPath := filepath.Join(dir, manifest.Infohash+".dc")
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatalf("saved manifest not found: %s", savedPath)
	}

	loaded, err := LoadManifest(savedPath)
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}

	if loaded.Infohash != manifest.Infohash {
		t.Fatalf("infohash mismatch: %s vs %s", loaded.Infohash, manifest.Infohash)
	}
	if loaded.Size != manifest.Size {
		t.Fatalf("size mismatch: %d vs %d", loaded.Size, manifest.Size)
	}
	if loaded.PieceCount != manifest.PieceCount {
		t.Fatalf("piece count mismatch: %d vs %d", loaded.PieceCount, manifest.PieceCount)
	}
	if loaded.ContainerID != manifest.ContainerID {
		t.Fatalf("container id mismatch: %s vs %s", loaded.ContainerID, manifest.ContainerID)
	}
}


// TestManifestMultiPiece handles the TestManifestMultiPiece HTTP request.
func TestManifestMultiPiece(t *testing.T) {
	dir := t.TempDir()

	size := ContainerPieceSize + ContainerPieceSize/2
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i % 256)
	}
	testFile := filepath.Join(dir, "multi_piece.bin")
	os.WriteFile(testFile, content, 0644)

	manifest, err := BuildManifest("multi", testFile)
	if err != nil {
		t.Fatalf("BuildManifest error: %v", err)
	}

	expectedPieces := int(size / ContainerPieceSize)
	if size%ContainerPieceSize > 0 {
		expectedPieces++
	}
	if manifest.PieceCount != expectedPieces {
		t.Fatalf("expected %d pieces, got %d", expectedPieces, manifest.PieceCount)
	}

	if !manifest.VerifyAll(testFile) {
		t.Fatal("VerifyAll should pass")
	}

	// Verify each piece individually
	f, _ := os.Open(testFile)
	defer f.Close()
	buf := make([]byte, ContainerPieceSize)
	for i := 0; i < manifest.PieceCount; i++ {
		n, _ := f.Read(buf)
		if !manifest.VerifyPiece(i, buf[:n]) {
			t.Fatalf("piece %d verification failed", i)
		}
	}
}


// TestPieceCountForSize handles the TestPieceCountForSize HTTP request.
func TestPieceCountForSize(t *testing.T) {
	manifest := &Manifest{PieceSize: ContainerPieceSize}

	tests := []struct {
		size    int64
		expect  int
	}{
		{1, 1},
		{ContainerPieceSize, 1},
		{ContainerPieceSize + 1, 2},
		{ContainerPieceSize * 3, 3},
		{ContainerPieceSize*3 + 100, 4},
	}

	for _, tt := range tests {
		count := manifest.PieceCountForSize(tt.size)
		if count != tt.expect {
			t.Errorf("PieceCountForSize(%d) = %d, expected %d", tt.size, count, tt.expect)
		}
	}
}
