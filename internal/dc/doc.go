// Package dc implements DC CryptoCloud — a P2P torrent-like container distribution network.
//
// Architecture:
//   - cloud.go:    Core state, container metadata, swarm management
//   - manifest.go: .dc manifest format (SHA-256 piece hashes, 256KB pieces)
//   - seed.go:     Seeding containers to the network
//   - leech.go:    Fetching containers from the swarm
//   - swarm.go:    Replication manager with healing loop
//   - api.go:      HTTP handlers for DC endpoints
//   - transport.go: P2P wire protocol (TCP port 17001)
//
// DC CryptoCloud enables decentralized container distribution with:
//   - 256KB pieces with SHA-256 integrity verification
//   - Swarm tracking (seeders/leechers per infohash)
//   - Replication factor (default 3x) with 120s healing loop
//   - Integration with CryptoContainer via /api/dc/seed-container
//   - HTTP API: seed, announce, swarm, fetch, list, status, manifest, piece, unseed
package dc
