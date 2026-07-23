// Package container provides container management and lifecycle operations
package container

import (
	"os"
	"path/filepath"
	"testing"
)

func testPath(t *testing.T) string {
	return filepath.Join(t.TempDir(), "test.sxc")
}


// TestNew handles the TestNew HTTP request.
func TestNew(t *testing.T) {
	c := New("/nonexistent/container.sxc")
	if c == nil {
		t.Fatal("expected non-nil")
	}
	if len(c.entries) != 0 {
		t.Fatal("expected empty entries")
	}
}


// TestInitAndOpen handles the TestInitAndOpen HTTP request.
func TestInitAndOpen(t *testing.T) {
	path := testPath(t)
	c := New(path)
	seed := "test seed phrase for container init test"

	if err := c.Init(seed); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if !c.IsOpen() {
		t.Fatal("expected open after Init")
	}
	if !c.HasContainer() {
		t.Fatal("expected container file to exist")
	}
	c.Close()

	c2 := New(path)
	if err := c2.Open(seed); err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if !c2.IsOpen() {
		t.Fatal("expected open after Load")
	}
	c2.Close()
}


// TestInitFileExists handles the TestInitFileExists HTTP request.
func TestInitFileExists(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	c.Close()

	c2 := New(path)
	if err := c2.Init("seed2"); err == nil {
		t.Fatal("expected error when container file exists")
	}
}


// TestStoreAndLoad handles the TestStoreAndLoad HTTP request.
func TestStoreAndLoad(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	if err := c.Store("key1", []byte("hello world")); err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	data, err := c.Load("key1")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", string(data))
	}
}


// TestStoreAndLoadMultiple handles the TestStoreAndLoadMultiple HTTP request.
func TestStoreAndLoadMultiple(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	entries := map[string]string{
		"a": "alpha",
		"b": "beta",
		"c": "gamma",
	}
	for k, v := range entries {
		if err := c.Store(k, []byte(v)); err != nil {
			t.Fatalf("Store %s: %v", k, err)
		}
	}

	list := c.List()
	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}

	for k, expected := range entries {
		data, err := c.Load(k)
		if err != nil {
			t.Fatalf("Load %s: %v", k, err)
		}
		if string(data) != expected {
			t.Fatalf("entry %s: expected '%s', got '%s'", k, expected, string(data))
		}
	}
}


// TestDelete handles the TestDelete HTTP request.
func TestDelete(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	c.Store("k1", []byte("v1"))
	c.Store("k2", []byte("v2"))

	if err := c.Delete("k1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, err := c.Load("k1"); err == nil {
		t.Fatal("expected error loading deleted key")
	}

	list := c.List()
	if len(list) != 1 || list[0] != "k2" {
		t.Fatalf("expected [k2], got %v", list)
	}
}


// TestLoadNonExistent handles the TestLoadNonExistent HTTP request.
func TestLoadNonExistent(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	if _, err := c.Load("nonexistent"); err == nil {
		t.Fatal("expected error for non-existent entry")
	}
}


// TestStoreBeforeOpen handles the TestStoreBeforeOpen HTTP request.
func TestStoreBeforeOpen(t *testing.T) {
	c := New("/tmp/nonexistent.sxc")
	if err := c.Store("x", []byte("y")); err == nil {
		t.Fatal("expected error storing before open")
	}
}


// TestLoadBeforeOpen handles the TestLoadBeforeOpen HTTP request.
func TestLoadBeforeOpen(t *testing.T) {
	c := New("/tmp/nonexistent.sxc")
	if _, err := c.Load("x"); err == nil {
		t.Fatal("expected error loading before open")
	}
}


// TestStoreJSONAndLoadJSON handles the TestStoreJSONAndLoadJSON HTTP request.
func TestStoreJSONAndLoadJSON(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	type testObj struct {
		Name  string
		Value int
	}
	obj := testObj{Name: "test", Value: 42}
	if err := c.StoreJSON("obj", obj); err != nil {
		t.Fatalf("StoreJSON failed: %v", err)
	}

	var loaded testObj
	if err := c.LoadJSON("obj", &loaded); err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if loaded.Name != "test" || loaded.Value != 42 {
		t.Fatalf("expected {test 42}, got {%s %d}", loaded.Name, loaded.Value)
	}
}


// TestCloseZeroesKeys handles the TestCloseZeroesKeys HTTP request.
func TestCloseZeroesKeys(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	if err := c.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if c.IsOpen() {
		t.Fatal("expected closed")
	}
	// Keys should be zeroed
	for _, b := range c.encKey {
		if b != 0 {
			t.Fatal("encKey not zeroed")
		}
	}
}


// TestWipe handles the TestWipe HTTP request.
func TestWipe(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	c.Store("k", []byte("v"))

	if err := c.Wipe(); err != nil {
		t.Fatalf("Wipe failed: %v", err)
	}
	if c.IsOpen() {
		t.Fatal("expected closed after Wipe")
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("expected container file removed after Wipe")
	}
}


// TestWipeBeforeOpen handles the TestWipeBeforeOpen HTTP request.
func TestWipeBeforeOpen(t *testing.T) {
	path := testPath(t)
	c := New(path)
	if err := c.Wipe(); err == nil {
		t.Fatal("expected error Wiping non-existent container")
	}
}


// TestHasContainer handles the TestHasContainer HTTP request.
func TestHasContainer(t *testing.T) {
	path := testPath(t)
	c := New(path)
	if c.HasContainer() {
		t.Fatal("expected false before Init")
	}
	c.Init("seed")
	c.Close()
	if !c.HasContainer() {
		t.Fatal("expected true after Init")
	}
}


// TestDeriveKeys handles the TestDeriveKeys HTTP request.
func TestDeriveKeys(t *testing.T) {
	salt := make([]byte, saltSize)
	salt[0] = 1
	salt[1] = 2

	e1, h1 := deriveKeys("test", salt)
	e2, h2 := deriveKeys("test", salt)
	if len(e1) != 32 || len(h1) != 32 {
		t.Fatalf("expected 32-byte keys, got %d/%d", len(e1), len(h1))
	}
	for i := range e1 {
		if e1[i] != e2[i] {
			t.Fatal("deriveKeys not deterministic")
		}
	}
	for i := range h1 {
		if h1[i] != h2[i] {
			t.Fatal("deriveKeys not deterministic for hmac key")
		}
	}
}


// TestDeriveKeysDifferentSalt handles the TestDeriveKeysDifferentSalt HTTP request.
func TestDeriveKeysDifferentSalt(t *testing.T) {
	salt1 := make([]byte, saltSize)
	salt2 := make([]byte, saltSize)
	salt2[0] = 0xFF

	e1, _ := deriveKeys("seed", salt1)
	e2, _ := deriveKeys("seed", salt2)

	same := true
	for i := range e1 {
		if e1[i] != e2[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different keys for different salts")
	}
}


// TestInitThenOpenDifferentSeed handles the TestInitThenOpenDifferentSeed HTTP request.
func TestInitThenOpenDifferentSeed(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("correct seed")
	c.Close()

	c2 := New(path)
	if err := c2.Open("wrong seed"); err == nil {
		t.Fatal("expected error opening with wrong seed")
	}
}


// TestListEmpty handles the TestListEmpty HTTP request.
func TestListEmpty(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	list := c.List()
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %v", list)
	}
}


// TestPersistenceAcrossOpenClose handles the TestPersistenceAcrossOpenClose HTTP request.
func TestPersistenceAcrossOpenClose(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("persist-seed")
	c.Store("k1", []byte("value1"))
	c.Close()

	c2 := New(path)
	c2.Open("persist-seed")
	defer c2.Close()

	data, err := c2.Load("k1")
	if err != nil {
		t.Fatalf("Load after reopen: %v", err)
	}
	if string(data) != "value1" {
		t.Fatalf("expected 'value1', got '%s'", string(data))
	}

	list := c2.List()
	if len(list) != 1 || list[0] != "k1" {
		t.Fatalf("expected [k1], got %v", list)
	}
}


// TestStoreEmptyData handles the TestStoreEmptyData HTTP request.
func TestStoreEmptyData(t *testing.T) {
	path := testPath(t)
	c := New(path)
	c.Init("seed")
	defer c.Close()

	if err := c.Store("empty", []byte{}); err != nil {
		t.Fatalf("Store empty: %v", err)
	}
	data, err := c.Load("empty")
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if len(data) != 0 {
		t.Fatal("expected empty data")
	}
}
