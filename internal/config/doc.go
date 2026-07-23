// Package config provides configuration loading and management for the simplex-node server.
// It defines the Config struct with fields for HTTP listen address, data directory,
// Telegram bot tokens, Ollama AI settings, gateway credentials (WhatsApp, Signal, Matrix, Discord),
// billing prices, and Tron treasury addresses. Supports JSON-based config files merged
// over sensible defaults via DefaultConfig() and Load().
package config
