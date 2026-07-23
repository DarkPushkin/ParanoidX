// Package ai provides AI client integration for simplex-node.
//
// This package implements an Ollama HTTP client used by the Steward
// service for AI-powered chat responses, content generation, and
// autonomous decision-making.
//
// Key types:
//   - OllamaClient: Connects to Ollama server for /api/generate
//   - ChatClient: Higher-level chat interface with context management
//
// The AI client is used by:
//   - steward (AI governance and monitoring)
//   - radio (AI content generation)
//   - admin (moderation stats, AI analytics)
package ai
