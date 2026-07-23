// Package eth provides an Ethereum bridge transfer registry for cross-chain token swaps.
// It manages BridgeTransfer records with direction (lock/burn/unlock/mint) and status
// (pending, verified, complete, failed) tracking. The BridgeRegistry supports Create,
// Confirm (with transaction hash), Complete (with proof transaction), and Fail operations
// for orchestrating two-step verified bridge transfers between chains.
package eth
