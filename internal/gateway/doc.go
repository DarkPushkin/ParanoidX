// Package gateway implements a multi-platform messaging gateway that routes
// messages between SimpleX and external platforms: Telegram, Discord, Signal,
// WhatsApp, and Matrix. It defines a common Message/OutMessage format, a Platform
// interface for connectors, and a Router with command dispatch and fallback handlers.
// Each platform sub-package provides Start/Send/SendText implementations.
package gateway
