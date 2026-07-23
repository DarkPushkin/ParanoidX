// Package store implements SQLite-based persistence for all simplex-node data.
//
// Database tables:
//   - accounts:    User accounts and authentication
//   - wallets:     Wallet addresses and balances
//   - taler:       Digital currency (Taler) transactions
//   - vault:       Encrypted file vault metadata
//   - dao:         DAO governance records
//   - stations:    Radio station configuration
//   - crypto:      Cryptographic key storage
//   - pin:         PIN hash and lock state
//   - migrations:  Schema migration tracking
//
// Additional persistence:
//   - chatpersist.go: Chat history stored as JSON file (chat_history.json)
//   - persist.go:     Generic JSON file persistence utility
//
// All SQLite operations use modernc.org/sqlite (pure Go, no CGO needed).
package store
