# ParanoidX — Evolution Plan

**Project:** ParanoidX (Go Server + Infrastructure)
**Current Cycle:** 16 (v px-node-C41-C60)
**Next Cycle Target:** 17
**Status:** Production — running on Dell Latitude 3150 (dev), ready for Beelink SER9 migration

---

## Evolution Philosophy

> **Code that writes code. Infrastructure that deploys infrastructure. Economy that backs itself with silver.**

Each cycle = 1 evolutionary step. 8-step SOP (PRODUCTION-CYCLE.md). Approval gates via Telegram buttons.

---

## Completed Cycles (History)

| Cycle | Codename | Key Achievement |
|-------|----------|-----------------|
| 1-10 | Genesis | simplex-node born, economy core, chat bridge, radio |
| 11 | Torification | Full Tor routing, .onion addresses |
| 12 | ParanoidX | Multi-layer routing (VLESS/VMess/Tor), Docker orchestration |
| 13 | Silver Standard | 70% physical silver backing, dividend engine |
| 14 | Governance | Constitution, proposals, weighted voting |
| 15 | AI Integration | Ollama (gemma4/qwen2.5), Steward bot (@AskSteward_bot) |
| 16 | Restructure | Monorepo → ParanoidX / The-Isle / Royal-Isle / shared-libs |

---

## Current Cycle: 17 — "Beelink Migration"

**Target Hardware:** Beelink SER9 (24GB RAM / 500GB SSD)
**Migration Date:** 2026-08-01 to 2026-08-03

### Objectives

| # | Task | Status |
|---|------|--------|
| 1 | Bootstrap script: install Go, Flutter, Docker, Ollama, Hermes | 🔲 |
| 2 | Clone all repos from GitHub (the-grimoire, ParanoidX, The-Isle, Royal-Isle, shared-libs) | 🔲 |
| 3 | Run `bootstrap.sh` — idempotent setup | 🔲 |
| 4 | Build ParanoidX binary | 🔲 |
| 5 | Configure systemd services (ParanoidX-*) | 🔲 |
| 6 | Start Docker stack (5 services) | 🔲 |
| 7 | Verify health endpoint + bridge | 🔲 |
| 8 | Migrate data from Dell (`~/.local/share/simplex-node/`) | 🔲 |
| 9 | Update DNS / Tor hidden service keys | 🔲 |
| 10 | Telegram approval: "Migration Complete ✅" | 🔲 |

---

## Upcoming Cycles (18-30)

### Cycle 18: "Observability"
- Structured logging aggregation (Loki/Grafana or simple JSON tail)
- Metrics endpoint expansion (`/api/admin/metrics` → Prometheus format)
- Health check matrix (bridge, docker, tor, ollama, disk, mem)
- Alerting via Telegram (Inquisitor bot)

### Cycle 19: "Resilience"
- Automated backup to USB (encrypted, verified)
- Disaster recovery test (restore from backup on clean VM)
- Multi-instance HA (active/passive with shared data dir)
- Chaos engineering: kill random container, verify recovery

### Cycle 20: "Economy Hardening"
- Oracle redundancy (multiple price sources, median)
- Deflation engine automation
- Dividend distribution scheduler (cron + verification)
- Audit trail: immutable log of all treasury ops

### Cycle 21: "Chat Federation"
- SimpleX group management via API
- Message retention policies
- Bridge to Matrix / Nostr (experimental)
- Encrypted backups of chat history

### Cycle 22: "Radio 2.0"
- AI content pipeline: prompt → script → TTS → mix → schedule
- Listener analytics (anonymous, local)
- Offline sync for Isle app
- Emergency broadcast override

### Cycle 23: "ParanoidX VPN"
- Visual route map (world map + latency)
- One-tap route switching (VLESS ↔ VMess ↔ Tor)
- Kill switch integration (systemd + nftables)
- Split tunneling rules (Island traffic vs clearnet)

### Cycle 24: "DC Cloud"
- Distributed compute nodes (Beelink + Lenovo + MacBook)
- Task queue (FFmpeg, Whisper, Ollama inference)
- Result aggregation + verification
- Payment in NG for compute

### Cycle 25: "Isle App Completion"
- sqlite3 build fix (pre-built binaries)
- Identity service (BIP39 + Ed25519)
- Full wallet flow
- Market + escrow
- Chat integration

### Cycle 26: "Royal Polish"
- Multi-profile (Inquisitor/Auditor/Steward)
- Real-time WebSocket metrics
- AI Office prompt library
- Animated heraldry onboarding

### Cycle 27: "Sovereign Deploy"
- One-command deploy to any Linux box
- ARM64 build (Raspberry Pi 5, Orange Pi)
- Tailscale mesh for admin access
- Air-gapped signing device support

### Cycle 28: "Testnet Launch"
- Public testnet (invite-only)
- Faucet (test NG)
- Stress test (1000 simulated users)
- Bug bounty (NG rewards)

### Cycle 29: "Mainnet Preparation"
- Security audit (code + infra)
- Legal framework (Saint Mary Liberty Island charter)
- Silver vault audit (physical)
- Documentation freeze

### Cycle 30: "Genesis Block"
- Mainnet launch
- First 100 TLR minted
- Inquisitor address: 0x001
- **The Island lives.**

---

## Resource Requirements (Beelink SER9)

| Resource | Allocation | Notes |
|----------|------------|-------|
| **RAM** | 24 GB | Ollama (8-16GB) + Go + Docker + Flutter build |
| **SSD** | 500 GB | 100GB data, 100GB models, 100GB Docker, 50GB build, 150GB buffer |
| **CPU** | 8C/16T (Ryzen 9) | Parallel builds, concurrent Ollama |
| **Network** | 1 Gbps + Tor | Clearnet for builds, Tor for production |

---

## Model Zoo (Ollama on Beelink)

| Model | Size (Q4) | Use Case | Priority |
|-------|-----------|----------|----------|
| `gemma2:27b` | ~15 GB | Steward (primary) | ✅ Must |
| `qwen2.5:7b` | ~4.5 GB | Steward (fallback) | ✅ Must |
| `phi3.5:3.8b` | ~2.3 GB | Lightweight tasks | 🔲 Nice |
| `llama3.2:3b` | ~2 GB | Quick replies | 🔲 Nice |
| `nomic-embed-text` | ~0.5 GB | Embeddings/RAG | 🔲 Must |
| `bge-m3` | ~1.5 GB | Multilingual embed | 🔲 Nice |

---

## Skill Exports (Hermes)

All Hermes skills → `the-grimoire/skills-export/`
- Auto-export on each cycle completion
- Versioned with cycle number
- Bootstrap script installs on new machine

---

## Approval Gates

Each cycle requires **Inquisitor approval** via Telegram buttons:

```
[✅ Approve Cycle N]  [❌ Reject]  [🔄 Request Changes]
```

Auto-deploy on ✅. Rollback script on ❌.

---

*"Eight steps. Infinite cycles. One Island."*

**Next Review:** Cycle 17 completion — Beelink SER9 online.