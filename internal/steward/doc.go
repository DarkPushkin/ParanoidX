// Package steward implements AI governance and autonomous monitoring.
//
// Components:
//   - Steward:      Main AI agent service (Ollama-powered Q&A with constitution context)
//   - Analyzer:     Treasury health analysis engine with scoring and recommendations
//   - Monitor:      Real-time system health monitoring with alert dispatch
//   - Constitution: Machine-readable governance rules and constraints
//
// The Steward serves as an autonomous governance agent that:
//   - Answers questions about node operations via natural language
//   - Analyzes treasury health and provides recommendations
//   - Monitors system health and dispatches alerts
//   - Operates with a written constitution defining its authority
package steward
