// Package container implements AES-256-GCM CryptoContainer with argon2id key derivation.
//
// Security model:
//
//	User mnemonic (BIP39) → argon2id (salt + time + mem) → AES-256-GCM key
//
// The container encrypts individual blobs (files) on disk with unique nonces.
// Features include:
//   - Encrypt/decrypt with authenticated encryption
//   - Paranoid wipe (overwrite with random data + zeroes + delete)
//   - Auto-delete scheduler (1 minute to 24 hours)
//   - Emergency panic wipe of all container data
//   - Each operation is logged for audit trail
package container
