// Package btc implements a Bitcoin atomic swap registry for cross-chain exchanges.
// It provides the AtomicSwap type with HTLC (Hash Time-Locked Contract) semantics:
// Create, Claim (with SHA-256 secret verification), Refund (after timelock), Confirm,
// Cancel, and CleanExpired operations. The SwapRegistry manages in-memory swap state
// with status tracking (pending, confirmed, claimed, refunded, cancelled, expired).
package btc
