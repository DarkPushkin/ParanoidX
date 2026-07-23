// Package bot implements Telegram bot integration.
//
// Three bots:
//   1. Steward (asksteward_bot):     AI Q&A using Ollama Llama 3.2
//   2. DarkPushkin (darkpushkin_bot): System monitoring alerts and reporting
//   3. Torquemada (torquemada_bot):  Inquisitor reporting bot
//
// Each bot runs as a long-polling Telegram bot using the Bot API.
// They are embedded in the server process (not separate systemd services)
// to avoid port conflicts and simplify deployment.
//
// The notify.go module provides notification dispatch to all bots.
package bot
