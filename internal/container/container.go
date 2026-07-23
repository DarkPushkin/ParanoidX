// Package container provides container management and lifecycle operations
package container

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	magic          = "SIMPLEX-CONTAINER-V1"
	headerSize     = 64
	saltSize       = 16
	nonceSize      = 12
	keySize        = 32
	hmacKeySize    = 32
	blockHeaderLen = 8 // 4 nameLen + 4 dataLen
	hmacSize       = 32
	iterations     = 3
	memory         = 64 * 1024
	threads        = 4
)

// entry represents a single encrypted blob stored inside the container.
type entry struct {
	Name    string `json:"name"`
	Data    []byte `json:"data"`
	Created int64  `json:"created"`
}

// Container implements an AES-256-GCM encrypted storage envelope.
//
// Key derivation chain:
//
//	BIP39 mnemonic (user) → argon2id(salt, time=3, mem=64MB) → 32-byte key
//	key → first 32 bytes = AES-256-GCM encryption key (encKey)
//	key → next 32 bytes  = HMAC key (hmacKey) for integrity verification
//
// Each blob is encrypted with a unique 12-byte nonce (random).
// The container file has a header (magic + salt + nonce + key check)
// followed by individually encrypted entries.
type Container struct {
	mu       sync.RWMutex
	path     string       // Path to the container file on disk
	entries  map[string]*entry // In-memory entry cache
	key      []byte       // Derived key (encKey + hmacKey combined)
	hmacKey  []byte       // HMAC key for blob integrity verification
	encKey   []byte       // AES-256-GCM encryption key
	nonce    []byte       // Current write nonce
	salt     []byte       // Random salt for argon2id derivation
	opened   bool
	dirty    bool
}


// New handles the New HTTP request.
func New(path string) *Container {
	return &Container{
		path:    path,
		entries: make(map[string]*entry),
	}
}

func deriveKeys(seedPhrase string, salt []byte) (encKey, hmacKey []byte) {
	key := argon2.IDKey([]byte(seedPhrase), salt, iterations, memory, threads, keySize+hmacKeySize)
	return key[:keySize], key[keySize:]
}


// Init handles the Init HTTP request.
func (c *Container) Init(seedPhrase string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := os.Stat(c.path); err == nil {
		return fmt.Errorf("container already exists at %s", c.path)
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("salt: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("nonce: %w", err)
	}

	c.encKey, c.hmacKey = deriveKeys(seedPhrase, salt)
	c.salt = salt
	c.nonce = nonce
	c.opened = true
	c.dirty = false

	if err := c.flushLocked(); err != nil {
		c.opened = false
		return err
	}
	return nil
}


// Open handles the Open HTTP request.
func (c *Container) Open(seedPhrase string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Open(c.path)
	if err != nil {
		return fmt.Errorf("open container: %w", err)
	}
	defer f.Close()

	var header [headerSize]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	var magicBytes [len(magic)]byte
	copy(magicBytes[:], header[:len(magic)])
	if string(magicBytes[:]) != magic {
		return fmt.Errorf("invalid container magic")
	}

	salt := header[len(magic) : len(magic)+saltSize]
	nonce := header[len(magic)+saltSize : len(magic)+saltSize+nonceSize]

	encKey, hmacKey := deriveKeys(seedPhrase, salt)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return fmt.Errorf("aes: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gcm: %w", err)
	}

	remaining, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	entries := make(map[string]*entry)
	if len(remaining) > 0 {
		plain, err := gcm.Open(nil, nonce, remaining, nil)
		if err != nil {
			return fmt.Errorf("decrypt (wrong seed phrase?): %w", err)
		}
		if err := c.parseEntries(plain, entries, hmacKey); err != nil {
			return err
		}
	}

	c.entries = entries
	c.encKey = encKey
	c.hmacKey = hmacKey
	c.salt = salt
	c.nonce = nonce
	c.opened = true
	c.dirty = false
	return nil
}

func (c *Container) parseEntries(data []byte, entries map[string]*entry, hmacKey []byte) error {
	buf := bytes.NewReader(data)
	for buf.Len() > 0 {
		var nameLen uint32
		if err := binary.Read(buf, binary.BigEndian, &nameLen); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("nameLen: %w", err)
		}
		if nameLen > 1024 {
			return fmt.Errorf("name too long: %d", nameLen)
		}
		name := make([]byte, nameLen)
		if _, err := io.ReadFull(buf, name); err != nil {
			return fmt.Errorf("name: %w", err)
		}
		var dataLen uint32
		if err := binary.Read(buf, binary.BigEndian, &dataLen); err != nil {
			return fmt.Errorf("dataLen: %w", err)
		}
		if dataLen > 100*1024*1024 {
			return fmt.Errorf("data too large: %d", dataLen)
		}
		data := make([]byte, dataLen)
		if _, err := io.ReadFull(buf, data); err != nil {
			return fmt.Errorf("data: %w", err)
		}
		var storedHMAC [hmacSize]byte
		if _, err := io.ReadFull(buf, storedHMAC[:]); err != nil {
			return fmt.Errorf("hmac: %w", err)
		}
		mac := hmac.New(sha256.New, hmacKey)
		mac.Write(name)
		binary.Write(mac, binary.BigEndian, dataLen)
		mac.Write(data)
		if !hmac.Equal(storedHMAC[:], mac.Sum(nil)) {
			return fmt.Errorf("hmac mismatch for %s", string(name))
		}
		entries[string(name)] = &entry{Name: string(name), Data: data, Created: time.Now().UnixNano()}
	}
	return nil
}

func (c *Container) flushLocked() error {
	if !c.opened {
		return fmt.Errorf("container not open")
	}

	var buf bytes.Buffer
	for _, e := range c.entries {
		nameBytes := []byte(e.Name)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(nameBytes))); err != nil {
			return err
		}
		buf.Write(nameBytes)
		if err := binary.Write(&buf, binary.BigEndian, uint32(len(e.Data))); err != nil {
			return err
		}
		buf.Write(e.Data)

		mac := hmac.New(sha256.New, c.hmacKey)
		mac.Write(nameBytes)
		binary.Write(mac, binary.BigEndian, uint32(len(e.Data)))
		mac.Write(e.Data)
		buf.Write(mac.Sum(nil))
	}

	block, err := aes.NewCipher(c.encKey)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	ciphertext := gcm.Seal(nil, c.nonce, buf.Bytes(), nil)

	var header [headerSize]byte
	copy(header[:], []byte(magic))
	copy(header[len(magic):], c.salt)
	copy(header[len(magic)+saltSize:], c.nonce)

	tmpPath := c.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(tmpPath, append(header[:], ciphertext...), 0600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		return err
	}
	c.dirty = false
	return nil
}


// Store handles the Store HTTP request.
func (c *Container) Store(name string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.opened {
		return fmt.Errorf("container not open")
	}
	c.entries[name] = &entry{Name: name, Data: data, Created: time.Now().UnixNano()}
	c.dirty = true
	return c.flushLocked()
}


// Load handles the Load HTTP request.
func (c *Container) Load(name string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.opened {
		return nil, fmt.Errorf("container not open")
	}
	e, ok := c.entries[name]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", name)
	}
	out := make([]byte, len(e.Data))
	copy(out, e.Data)
	return out, nil
}


// Delete handles the Delete HTTP request.
func (c *Container) Delete(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.opened {
		return fmt.Errorf("container not open")
	}
	delete(c.entries, name)
	c.dirty = true
	return c.flushLocked()
}


// List handles the List HTTP request.
func (c *Container) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.entries))
	for n := range c.entries {
		names = append(names, n)
	}
	return names
}


// Close handles the Close HTTP request.
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.opened {
		return nil
	}
	if c.dirty {
		if err := c.flushLocked(); err != nil {
			return err
		}
	}
	c.opened = false
	zero(c.encKey)
	zero(c.hmacKey)
	zero(c.salt)
	zero(c.nonce)
	c.entries = make(map[string]*entry)
	return nil
}


// Wipe handles the Wipe HTTP request.
func (c *Container) Wipe() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*entry)
	c.dirty = false
	if c.opened {
		zero(c.encKey)
		zero(c.hmacKey)
		zero(c.salt)
		zero(c.nonce)
		c.opened = false
	}

	f, err := os.OpenFile(c.path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	zeros := make([]byte, headerSize)
	if _, err := f.Write(zeros); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := os.Remove(c.path); err != nil {
		return err
	}
	return nil
}


// IsOpen handles the IsOpen HTTP request.
func (c *Container) IsOpen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.opened
}


// HasContainer handles the HasContainer HTTP request.
func (c *Container) HasContainer() bool {
	_, err := os.Stat(c.path)
	return err == nil
}


// StoreJSON handles the StoreJSON HTTP request.
func (c *Container) StoreJSON(name string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Store(name, data)
}


// LoadJSON handles the LoadJSON HTTP request.
func (c *Container) LoadJSON(name string, v any) error {
	data, err := c.Load(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
