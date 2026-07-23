// Package registry manages node discovery, regional routing, and health tracking
// for the simplex-node relay mesh network. It maintains a Registry of NodeInfo entries
// with capabilities (relay-smp, relay-xftp, radio-seed, vault-peer, dc-seed), heartbeat
// monitoring, and HTTP handlers for announce, discover, heartbeat, list, and status.
// Includes latency-based node selection via PingLatency and SelectBestNodes.
package registry
