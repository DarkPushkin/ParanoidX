// Package dc implements P2P container distribution (DC CryptoCloud)
package dc

import (
	"os"
	"path/filepath"
	"testing"
)


// TestBuildManifest handles the TestBuildManifest HTTP request.
func TestBuildManifest(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.bin")
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	m, err := BuildManifest("test-container", filePath)
	if err != nil {
		t.Fatal(err)
	}
	if m.Infohash == "" {
		t.Fatal("expected non-empty infohash")
	}
	if m.PieceCount == 0 {
		t.Fatal("expected pieces")
	}
	if m.Size != 1024*1024 {
		t.Fatalf("expected size 1MB, got %d", m.Size)
	}
	if !m.VerifyAll(filePath) {
		t.Fatal("verify failed")
	}
}


// TestCloudSeedAndList handles the TestCloudSeedAndList HTTP request.
func TestCloudSeedAndList(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "dc", "containers"), 0755)
	os.MkdirAll(filepath.Join(dir, "dc", "manifests"), 0755)
	os.MkdirAll(filepath.Join(dir, "dc", "pieces"), 0755)

	cloud := NewCloud(dir)
	cloud.Start()
	defer cloud.Stop()

	filePath := filepath.Join(dir, "test.bin")
	data := make([]byte, 512*1024)
	os.WriteFile(filePath, data, 0644)

	m, err := cloud.SeedContainer(filePath, "test-id")
	if err != nil {
		t.Fatal(err)
	}
	if !cloud.IsSeeding(m.Infohash) {
		t.Fatal("expected seeding")
	}

	containers := cloud.ListContainers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	if containers[0].Infohash != m.Infohash {
		t.Fatalf("infohash mismatch: %s vs %s", containers[0].Infohash, m.Infohash)
	}
}


// TestCloudFetch handles the TestCloudFetch HTTP request.
func TestCloudFetch(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "dc", "containers"), 0755)
	os.MkdirAll(filepath.Join(dir, "dc", "manifests"), 0755)
	os.MkdirAll(filepath.Join(dir, "dc", "pieces"), 0755)

	cloud := NewCloud(dir)
	cloud.Start()
	defer cloud.Stop()

	filePath := filepath.Join(dir, "source.bin")
	data := make([]byte, 256*1024+100)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(filePath, data, 0644)

	m, err := cloud.SeedContainer(filePath, "fetch-test")
	if err != nil {
		t.Fatal(err)
	}

	outputPath := filepath.Join(dir, "fetched.bin")
	if err := cloud.FetchContainer(m.Infohash, outputPath); err != nil {
		t.Fatal(err)
	}

	fetched, _ := os.ReadFile(outputPath)
	if len(fetched) != len(data) {
		t.Fatalf("size mismatch: %d vs %d", len(fetched), len(data))
	}
	for i := range data {
		if fetched[i] != data[i] {
			t.Fatalf("data mismatch at byte %d", i)
		}
	}
}


// TestInfohash handles the TestInfohash HTTP request.
func TestInfohash(t *testing.T) {
	h1 := Infohash([]byte("hello"))
	h2 := Infohash([]byte("hello"))
	h3 := Infohash([]byte("world"))
	if h1 != h2 {
		t.Fatal("same input should produce same hash")
	}
	if h1 == h3 {
		t.Fatal("different input should produce different hash")
	}
	if len(h1) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(h1))
	}
}
