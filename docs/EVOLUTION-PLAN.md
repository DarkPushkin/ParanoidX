# Comprehensive Evolution Plan: simplex-node (A1 → B2)

**Date:** 2026-06-22
**Status:** Active development
**Version:** A3.2 (Cycle 39)

---

## Table of Contents

1. Overview
2. Track 1: Foundation & Technical Debt
3. Track 2: Feature Evolution (A2 → A5)
4. Track 3: Testing & Quality
5. Track 4: Load Modeling & Performance
6. Track 5: Client Applications — Flutter Mobile + Desktop
7. Track 6: Crypto Bridges & ICO (Phase B1)
8. Track 7: Multi-Platform Messaging Bots (Phase B2)
9. Master Timeline
10. Key Architectural Decisions

---

## 1. Overview

Seven parallel tracks covering the full evolution of simplex-node from a single-binary alpha to a multi-chain crypto sovereign network with AI-governed economy, multi-platform messaging bots, ICO funding, and client applications for Royal admin and citizens.

**Current baseline:** 283 API endpoints, 498 test functions across 48 test files (20/33 packages tested, 17 pass / 3 fail), 6 Docker containers (5 healthy, v2ray unhealthy), 4 Telegram bots + Go-based AskSteward bot, Flutter project scaffold in apps/ with shared packages, node-monitor Python GTK4 daemon, multi-chain bridge basics (TRON+TREASURY+TON), Steward AI Core with Ollama/Gemma 4 integration, Gateway unification module, POS with QR codes, Banknote PDF with Ed25519 signing, Onboarding funnel, ParanoidX chain orchestrator (V2Ray+Tor+WireGuard+VPN profiles), 33 internal Go packages with 283 HTTP handlers.

### Current Test Failures (Need Fixing)

| Package | Issue | Priority |
|---------|-------|----------|
| `internal/vault` | OOM crash on `TestUploadExceedsQuota` (16GB alloc), 4 test failures | **CRITICAL** |
| `internal/radio` | `TestPlaylistBuilderEmptyStation` — empty station should have placeholder | HIGH |
| `internal/ai` | `TestEconomySummary`, `TestModerationCheck` — Ollama unreachable from test env | HIGH |
| `internal/api` | `TestRadioHandlerPlaylistValid`, `TestRadioHandlerDefaultAction` — radio handler failures | HIGH |

---

## 2. Track 1: Foundation & Technical Debt

### 2.1 Critical Bugs (Fix Immediately, Week 1)

| ID | Issue | Location | Fix |
|----|-------|----------|-----|
| B1 | **Data race: `knownRoles` map** | `main.go` (set_role_chat + WS goroutine) | Replace `map` with `sync.Map` or add `sync.RWMutex` |
| B2 | **Data race: `islandWS` conn unlock-before-use** | `main.go` bridge goroutine | Move `Unlock()` after `WriteJSON` completes |
| B3 | **Connection leak: `r.Body` never closed** | 10+ HTTP handlers | Add `defer r.Body.Close()` after every `Decode` |
| B4 | **Nil dereference: `f.WriteString` on nil `*os.File`** | 6+ file write sites | Check `os.OpenFile` error before calling `WriteString` |
| B5 | **JSON WAL not implemented**: crash between marshal and write loses data | All JSON persistence | Write to `.tmp`, `Sync()`, then `Rename()` — atomic write pattern |

### 2.2 Security (Week 1-2)

| ID | Issue | Fix |
|----|-------|-----|
| S1 | Plaintext PIN in `lock.json` | `bcrypt` hash + constant-time comparison |
| S2 | No rate limiting on `/api/unlock` | Token bucket rate limiter (5/60s per IP) |
| S3 | No request body size limits | `http.MaxBytesReader` on all POST endpoints |
| S4 | Shell injection in Python heredoc in `royal-telegram-command-listener.sh` | Parameterize, never inline |

### 2.3 Infrastructure (Week 1-2)

| ID | Issue | Fix |
|----|-------|-----|
| I1 | No graceful shutdown | `signal.Notify(SIGTERM, SIGINT)` — drain connections, flush state, close bridge |
| I2 | 16+ hardcoded `/home/tomas/...` paths | `internal/config/config.go` — `simplex-node.json` with `-config` flag |
| I3 | Rotate goroutine leak | `sync.Mutex` + singleflight pattern |
| I4 | No structured logging | Replace `fmt.Printf` with `log/slog` (Go 1.21+) — JSON output to file |
| I5 | No Docker healthchecks | Add `healthcheck:` to all 4 containers in `docker-compose.yml` |

### 2.4 Code Architecture (Week 2-3)

Split monolithic `main.go` (3755 LOC) into domain packages as documented in `Architecture.md`:

```
cmd/simplex-node/main.go     → thin router (~200 LOC)
internal/
  api/
    router.go                → HTTP mux setup
    middleware.go            → auth, rate limiting, CORS
  treasury/
    treasury.go              → silver round, USDT monitoring, state
    handlers.go              → treasury API endpoints
  radio/
    radio.go                 → announcement management, audio serving
  channels/
    channels.go              → channel CRUD, posts, access control
  services/
    registry.go              → service registration, catalog, billing
```

---

## 3. Track 2: Feature Evolution

### 3.1 Phase A2: Core Completion (Week 3-5)

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| A2.1 | **Real TRON USDT monitoring** | Background goroutine polling TronGrid every 60s, auto-log to `treasury_usdt.log`, live TX tracking | 2d |
| A2.2 | **Auto silver rounds by threshold** | Ticker checks `totalUSDT >= AUTOROUND_THRESHOLD` (default $1000), dynamic threshold scaling (x10 after 10 rounds) | 2d |
| A2.3 | **Dividend claiming API** | `POST /api/treasury/claim-dividend?serial=&pubkey=` — transfers accrued ng from banknote holder to Liquid Taler ledger balance | 2d |
| A2.4 | **Proof of Reserve API** | `GET /api/reserve/por` — returns `total_silver_ng`, `total_issued_ng`, reserve ratio, Merkle-style hash chain | 1d |
| A2.5 | **Royal → Sub node SMP control** | Ed25519-signed command protocol, sub-node registration (`POST /royal/register`), sync heartbeat | 4d |
| A2.6 | **Marketplace escrow** | Buy/sell with ng held in escrow, release on confirmation, dispute timeout | 2d |
| A2.7 | **Create missing packages** | `internal/treasury/`, `internal/radio/`, `internal/channels/` — extract from main.go | 2d |
| A2.8 | **banknote-press real PDF** | Replace text-only stub with template rendering + Ed25519 double-signature (royal_sig + issuer_sig) | 2d |
| A2.9 | **Dynamic ICE credentials** | HMAC-SHA1 time-based TURN credentials, rotated every 12 hours | 1d |

### 3.2 Phase A3: Flywheel Economy (Week 5-9)

Based on `docs/FLYWHEEL-ECONOMY.md`:

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| A3.1 | **Onboarding funnel** | Welcome banknote (1 free Common), Starter Pack (5 TLR = 3C+1R+1E), guided tour, first-time flow | 3d |
| A3.2 | **Subscription tiers** | Colonist (free/5 TLR), Citizen (10 TLR/mo — vault 2GB, 0% P2P fee, voting), Aristocrat (100 TLR/mo — x4 dividends, auto-bid, early access) | 4d |
| A3.3 | **Golden Wheel** | Daily free spin: 90% 1000ng, 8% Common, 1.5% Rare, 0.4% Epic, 0.09% Legendary, 0.01% Genesis | 2d |
| A3.4 | **Auto-mint by treasury tier** | T1 (10T ng → 100 banknotes), T2 (50T → 250), T3 (200T → 500), T4 (1000T → 1000) — issued as limited sets | 3d |
| A3.5 | **Banknote sets** | Set definitions (GENESIS-001, COLONY-001), completion tracking per pubkey, bonus ng on full set | 3d |
| A3.6 | **Crafting 5→1** | 5 Common → 1 Rare, 5 Rare → 1 Epic, 5 Epic → 1 Legendary, 5 Legendary → 1 Genesis — source banknotes burned | 3d |
| A3.7 | **Leaderboard** | Top 10 by balance/rarity/activity, weekly ng rewards, public API | 2d |
| A3.8 | **Auto-reinvest dividends** | Opt-out auto-purchase of 1 Common when dividends ≥ threshold | 2d |
| A3.9 | **Deflation mechanisms** | Treasury burn (20-40% at fat tiers), 30% subscription burn, 5% auction fee burn | 2d |
| A3.10 | **Silver spot oracle** | `api.metals.live` polling, dynamic `$TLR = spot_price * 1oz / 31_103_480_000 ng`, auto-trigger on +5% daily move | 1d |

### 3.3 Phase A4: Franchise Silver Standard (Week 9-13)

Based on `docs/FRANCHISE-SILVER-STANDARD.md`:

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| A4.1 | **License CRUD** | `internal/franchise/license.go` — `FranchiseLicense` struct, JSON persistence, CRUD API, Ed25519 key pair generation | 3d |
| A4.2 | **Earmarked accounts** | Extend `internal/economy/ledger.go` — `EarmarkEntry` with `OwnerNg`/`BackingNg`, invariant: `PhysicalSilver >= sum(EarmarkEntry.BackingNg)` | 2d |
| A4.3 | **Mint authorization** | `POST /api/franchise/mint` — Royal validates reserve, daily cap, signature → returns signed `mint_authorization_token` | 3d |
| A4.4 | **National templates** | `press.MultiTemplateManager` — SVG-based templates per nation, Royal approval workflow, serial prefix system | 3d |
| A4.5 | **Double signing** | Each franchise banknote carries `royal_sig` (Ed25519) + `issuer_sig` (franchise mint key) — verified independently | 2d |
| A4.6 | **Settlement layer** | Cross-franchise transfers: `POST /api/settlement/transfer` — debits from_franchise earmark, credits to_franchise earmark, 0.5% royalty to Island | 4d |
| A4.7 | **Hierarchical PoR** | Merkle tree per franchise, aggregated root on Royal, `GET /api/reserve/por` returns full tree, any node can verify inclusion | 3d |
| A4.8 | **Franchise node SDK** | Docker Compose template, `setup.sh` script, sync protocol (`/api/sync/push`, `/api/sync/pull`), monitoring | 3d |

### 3.4 Phase A5: Platform + AI Steward (Week 13-17)

Based on `docs/ISLAND-EVOLUTION.md`:

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| A5.1 | **Service registry** | `internal/services/registry.go` — register Island services with deposit, category, pricing, discovery API | 3d |
| A5.2 | **Steward core** | `internal/steward/` — `StewardConstitution` (hardcoded rules), `Monitor` (60s ticker, collects 15+ metrics), `Analyzer` (deviation detection), `DecisionEngine` (3-level: minor auto / major notify / critical require consensus) | 8d |
| A5.3 | **Steward actions** | Auto-adjust: mint rate, burn rate, auction fee (0-10%), subscription prices. All actions logged to `steward_actions.log` | 4d |
| A5.4 | **AI arbitration** | Dispute resolution: parties submit evidence to vault, AI analyzes tx history/logs, ruling + appeal to 3 auditors | 4d |
| A5.5 | **Service marketplace** | Browse, ratings, billing (2% tx tax, 5% revenue share, 30% of tax burned), category filters | 3d |
| A5.6 | **Mobile app (Flutter)** | SimpleX SMP library integration, Onion REST client via Tor, 5 tabs: Chat / Wallet / Market / Vault / Radio | 14d |
| A5.7 | **Gateway unification** | `internal/gateway/` — `Message` struct, platform router (Telegram / SimpleX / WhatsApp), shared command handlers | 5d |
| A5.8 | **Go Telegram bot** | Replace 3 Python polling bots with single native Go binary using `telegram-bot-api/v5`, webhook mode | 3d |

---

## 4. Track 3: Testing & Quality

### 4.1 Level 1: Unit Tests (Go, Start Week 1, Continuous)

| Package | Tests Needed | Priority |
|---------|-------------|----------|
| `economy/ledger.go` | `calcNewSilverNg(usdt)`, `proRata(holders, pool)`, `Transfer`, `Mint`, `History` — table-driven with edge cases (zero, overflow, negative) | **HIGH** |
| `economy/registry.go` | BanknoteV2 serialization, PreMint promotion, CalculateTreasurySplit (thin/normal/fat/very fat), auditor bubble sort ordering, first-investor debt repayment | **HIGH** |
| `economy/auction.go` | Bid validation (≥ min, < current), auto-extend (last 5min), escrow hold/release, 5% fee calculation, background closer (30s ticker) | **HIGH** |
| `economy/buyback.go` | Pricing tiers (Common -2.28%, Rare/Epic/Legendary par, Golden/Genesis → auction), burn registry, re-mint | MEDIUM |
| `economy/pack.go` | 5-banknote generation, ≥1 Rare guarantee, serial uniqueness across packs, sealed/open state transition | MEDIUM |
| `economy/p2p.go` | Offer CRUD (create/accept/reject/cancel), holder validation, auditor-gate enforcement | MEDIUM |
| `press/templates.go` | Template listing, manifest validation, serial prefix validation, signature verification | MEDIUM |
| `lock/lock.go` | bcrypt hash, constant-time compare, rate-limited unlock (5 fails → 5s delay), change-code flow | **HIGH** |
| `health/health.go` | Check aggregation, Docker status parsing, disk threshold alerts, Tor onion reachability | LOW |

**Target coverage:** 80%+ economy package, 60%+ overall. Enforced in CI.

### 4.2 Level 2: Integration Tests (Week 2-4)

| Fixture | Contents | Tests |
|---------|----------|-------|
| `testdata/fixtures/minimal` (exists) | 2 banknotes (1+10 TLR), 0 reserve, royal.enabled | Round simulation, dividend math, RWA register, radio anns ✅ |
| `testdata/fixtures/multi-holder` | 5 holders (1, 5, 10, 42, 100 TLR), 50kg reserve | Pro-rata distribution across unequal denominations, accrued monotonicity |
| `testdata/fixtures/fat-treasury` | 200kg reserve, 10 holders, 8 past rounds | Treasury split (thin→fat→very fat threshold), burn calculation |
| `testdata/fixtures/auction-house` | 10 active auctions, 3 escrowed banknotes, bid history | Bid validation, auto-extend, closing, fee collection |
| `testdata/fixtures/pack-shop` | Pre-mint of 100 banknotes, 20 packs available | Buy pack, open pack, verify rarity distribution |
| `testdata/fixtures/edge-cases` | Empty reserve, 0 holders, overflow ng values | Error handling, no crash on empty state |
| `testdata/fixtures/franchise` | 2 franchise licenses, earmarked accounts, mint tokens | License CRUD, mint authorization, settlement |

**Expand `scripts/test-royal.sh`** to:
- Test all 66 endpoints (currently 6)
- Run against each fixture as a test matrix
- Assert exact ng math (not just "contains")
- Test error paths (invalid params, auth failures, empty data)
- Test concurrent operations (race detection mode)
- Test data persistence (kill + restart + verify state)

### 4.3 Level 3: Load Testing (Week 6-8)

Tool: **k6** (JavaScript-based, high performance)

| Scenario | Script | Target |
|----------|--------|--------|
| Burst read: 100 concurrent `/api/economy/state` | 100 VUs, 30s ramp, 60s steady | <500ms p99, 0% errors, no OOM |
| Auction race: 50 concurrent bids on same lot | 50 VUs, single lot, last 60s of auction | Correct serialization, no double-spend, 1 winner |
| Silver round: 100 sequential rounds, growing holders | 10→100→1000 holders, measure round time per holder | <5s for 1k holders, O(n) scaling |
| Vault upload: 10 concurrent 10MB files | 10 VUs, 10MB random data, measure throughput | >20 MB/s aggregate, quota intact |
| Wallet transfer: 200 TPS | 200 VUs, sequential transfers between random accounts | <1s p99, 0 lost writes, ng conservation |
| Mixed workload: 80% reads / 20% writes | Realistic user behavior model | Stable latency under sustained load |

### 4.4 Level 4: Quality Gates (CI)

```yaml
# .github/workflows/ci.yml
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - run: golangci-lint run --timeout 5m
  race:
    steps:
      - run: go test -race -count=1 ./...
  unit:
    steps:
      - run: go test -coverprofile=coverage.out -covermode=atomic ./...
      - run: go tool cover -func=coverage.out | grep total
      - run: test $(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%') -ge 60
  integration:
    steps:
      - run: for f in testdata/fixtures/*/; do ./scripts/test-royal.sh "$f"; done
  build:
    steps:
      - run: go build -o /dev/null ./cmd/...
  bench:
    steps:
      - run: go test -bench=. -benchmem -count=5 ./... > new.txt
      - run: benchstat old.txt new.txt  # compare to main branch
```

**Quality thresholds:**

| Gate | Tool | Threshold |
|------|------|-----------|
| Race detector | `go test -race` | 0 races |
| Linting | `golangci-lint` (govet, errcheck, ineffassign, staticcheck) | 0 issues |
| Coverage | `go test -cover` | ≥60% overall, ≥80% economy |
| Fuzzing | `go test -fuzz` (registry, auction, ledger) | 10 min no panic |
| Benchmark regression | `benchstat` | No >5% regression |
| Build | `go build` | All targets compile |

---

## 5. Track 4: Load Modeling & Performance

### 5.1 Current Architecture Bottlenecks

| Component | Constraint | Impact | Fix |
|-----------|------------|--------|-----|
| **JSON I/O** | `json.MarshalIndent` + `os.WriteFile` + `Sync()` per write | ~200 writes/sec per file | Write buffer + batch flush + `.tmp` atomic replace |
| **File handle leak** | No `defer f.Close()` after reads in some handlers | FD exhaustion under load | Audit all file ops, add `defer f.Close()` |
| **Mutex contention** | `BanknoteRegistry` uses single `sync.Mutex` (not RWMutex) | 1 writer blocks all readers | `sync.RWMutex` for registry + auction + ledger |
| **Bridge reconnect** | Infinite recursive `b.Run()` on disconnect | Stack growth, resource leak | Backoff (exponential, 1s→60s max), context timeout |
| **Single process** | All API + bridge + tickers in 1 binary | No horizontal scaling | Sub-node mesh for reads, write master |
| **No connection pool** | Each handler opens/closes files independently | High `open()` syscall count | File handle cache + reusable buffers |

### 5.2 Capacity Targets

| Metric | Target (1k users) | Target (10k users) | Strategy |
|--------|-------------------|--------------------|----------|
| API throughput | >100 req/s | >1,000 req/s | In-memory read cache, batch writes |
| Auction bids | >50/simultaneous | >500/simultaneous | Queue + batch process every 5s |
| Silver round processing | <5s (10k holders) | <30s (100k holders) | Pre-compute pro-rata with map-reduce |
| Vault throughput | >20 MB/s | >100 MB/s | Streaming writes, no buffering |
| Bridge messages | >50 msg/s | >500 msg/s | Goroutine-per-connection, read buffer pool |
| Data integrity | 0 lost writes | 0 lost writes | JSON WAL (append-only + periodic compaction) |

### 5.3 Scaling Roadmap

```
Phase A2 (Week 3):
  100 users → JSON files + sync.RWMutex + atomic writes
  Bottleneck: JSON serialization (ok for 100)

Phase A3 (Week 5):
  1,000 users → JSON + WAL + in-memory read cache (sync.Map for registry)
  Bottleneck: File I/O on vault + concurrent auction bids
  Mitigation: Batch bid processing, vault streaming

Phase A4 (Week 9):
  10,000 users → SQLite for ledger/registry, JSON for vault/metadata
  Bottleneck: SQLite write lock on concurrent transfers
  Mitigation: Write-ahead logging (WAL mode), read replicas

Phase A5 (Week 13):
  100,000 users → PostgreSQL for ledger, S3-compatible for vault, read replicas
  Bottleneck: Network I/O, cross-franchise settlement
  Mitigation: Shard by franchise, async settlement
```

### 5.4 Key Performance Indicators (KPI Dashboard)

| KPI | Collection | Alert |
|-----|-----------|-------|
| API latency (p50/p95/p99) | Prometheus histogram | p99 > 1s for 5 min |
| Error rate (4xx + 5xx) | Prometheus counter | >1% of requests |
| JSON write queue depth | Custom metric | >100 queued |
| Mutex wait time | `sync.Mutex` profiling | >50ms avg wait |
| File descriptor count | `/proc/self/fd` | >80% of ulimit |
| Goroutine count | `runtime.NumGoroutine()` | >5000 |
| Heap in use | `runtime.ReadMemStats` | >80% of max |
| Bridge reconnect count | Custom metric | >10/min |
| Auction closing latency | Timestamp diff | >30s behind schedule |

---

## 6. Track 5: Client Applications — Flutter Mobile + Desktop

**Goal:** Build two Flutter client applications — Royal Node Control (admin panel for the King family) and The Isle (citizen application for all island services).

**Tech stack:** Flutter (iOS + Android + macOS + Linux + Web), monorepo under `simplex-node/apps/`

### 6.1 Royal Node Control — For the King Family

Privileged admin dashboard to manage the simplex-node:

| Feature | Description | Phase |
|---------|-------------|-------|
| System Health Dashboard | Container status, disk, Tor onions, uptime, alerts | MVP (C10-12) |
| Build/Deploy/Test/Backup | One-click build, test suite, backup to USB | MVP (C10-12) |
| Treasury Management | Approve silver rounds, adjust thresholds, view reserve | Core (C13-15) |
| Royal→Sub Relay | Register sub-nodes, relay commands, poll for results | Core (C13-15) |
| Audit Log | Service payments, reputation scores, node events | Core (C13-15) |
| Key Management | Rotate onion keys, set King pubkey, manage Ed25519 | Advanced (C16-18) |
| Multi-Node Map | View all registered sub-nodes and their status | Advanced (C16-18) |
| Service Registry | List/manage registered services across the mesh | Advanced (C16-18) |

### 6.2 The Isle — For Citizens

Full-featured citizen client for all island services:

| Feature | Description | Phase |
|---------|-------------|-------|
| Ed25519 Login | Keypair generation, QR scan, mnemonic wallet | MVP (C10-12) |
| Dashboard | Liquid Taler balance, banknotes, reputation tier | MVP (C10-12) |
| Wallet | Balance, transfer, transaction history | MVP (C10-12) |
| Contacts | SimpleX address book, E2EE messaging via bridge | Core (C13-15) |
| Cloud Storage | Vault file manager (list, upload, download, delete) | Core (C13-15) |
| Marketplace | Browse/sell/buy RWA with escrow protection | Core (C13-15) |
| Notifications | Poll-based real-time alerts from server | Core (C13-15) |
| Digital ID | Ed25519 pubkey QR, verification badge | Advanced (C16-18) |
| Banknote Viewer | PDF rendering with rarity badge and metadata | Advanced (C16-18) |
| Auction House | Browse lots, place bids, watch timer, history | Advanced (C16-18) |
| Radio Player | Stream audio announcements from the island | Advanced (C16-18) |

### 6.3 Shared Packages

| Package | Contents |
|---------|----------|
| `apps/shared/api_client/` | Typed HTTP client for all simplex-node REST endpoints |
| `apps/shared/models/` | Shared Dart data models (User, Banknote, Escrow, etc.) |
| `apps/shared/widgets/` | Reusable UI components (tiles, cards, charts) |

### 6.4 Platform Targets

| App | Primary | Secondary |
|-----|---------|-----------|
| Royal Node Control | macOS + Linux Desktop | iOS + Android |
| The Isle | iOS + Android | Web + Desktop |

---

## 7. Track 6: Crypto Bridges & ICO (Phase B1)

**Goal:** Deploy multi-chain crypto bridges (Bitcoin, Ethereum, Solana, Polygon) and launch Genesis ICO with tiered vesting, ARGENTUM TON Jetton mainnet deployment, and Bridge Board DAO.

### 7.1 Bridge Architecture

```
External Chain     Bridge Layer           The Isle
┌──────────┐    ┌──────────────┐       ┌──────────┐
│ TRON     │───▶│ USDT Monitor  │──────▶│ Treasury │
│ USDT     │    │ (TronGrid)   │       │ Ledger   │
└──────────┘    └──────────────┘       └──────────┘
┌──────────┐    ┌──────────────┐       ┌──────────┐
│ TON      │◀──▶│ ARGENTUM     │◀─────▶│ ng       │
│ Jetton   │    │ Swap Engine  │       │ Ledger   │
└──────────┘    └──────────────┘       └──────────┘
┌──────────┐    ┌──────────────┐       ┌──────────┐
│ Bitcoin  │───▶│ BTC Bridge    │──────▶│ Treasury │
│ (future) │    │ (Atomic Swap)│       │ Ledger   │
└──────────┘    └──────────────┘       └──────────┘
┌──────────┐    ┌──────────────┐       ┌──────────┐
│ Ethereum │───▶│ ETH Bridge    │──────▶│ Treasury │
│ (future) │    │ (LayerZero)  │       │ Ledger   │
└──────────┘    └──────────────┘       └──────────┘
┌──────────┐    ┌──────────────┐       ┌──────────┐
│ Solana   │───▶│ SOL Bridge    │──────▶│ Treasury │
│ (future) │    │ (Wormhole)   │       │ Ledger   │
└──────────┘    └──────────────┘       └──────────┘
```

### 7.2 ICO Sale Tiers

| Tier | Min Investment | Bonus | Vesting |
|------|---------------|-------|---------|
| Genesis Angel | 100,000 USDT | 30% | 6m cliff, 12m linear |
| Major Investor | 10,000 USDT | 20% | 3m cliff, 9m linear |
| Minor Investor | 1,000 USDT | 10% | 1m cliff, 6m linear |
| Citizen | 100 USDT | 5% | No cliff, 3m linear |

### 7.3 Bridge Implementation

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| B1.1 | ARGENTUM Jetton mainnet | Deploy TON Jetton, add liquidity pools (STON.fi, DeDust) | 2w |
| B1.2 | Bitcoin atomic swap | HTLC-based BTC↔ng atomic swap, Lightning tips | 4w |
| B1.3 | Ethereum LayerZero bridge | OFT adapter for ETH/USDC/USDT, CCIP fallback | 6w |
| B1.4 | Solana Wormhole bridge | SOL/USDC via Wormhole VAAs | 4w |
| B1.5 | Polygon/Arbitrum/Base | PoS bridge + Arbitrum bridge + Base bridge | 3w |
| B1.6 | Bridge Board election | 5-7 multi-sig validators, quarterly elections | 2w |
| B1.7 | ICO smart contracts | Vesting contracts, ARGENTUM distribution | 3w |
| B1.8 | PoR dashboard | Public proof-of-reserve page for all bridges | 2w |

---

## 8. Track 7: Multi-Platform Messaging Bots (Phase B2)

**Goal:** Extend the gateway module (`internal/gateway/`) with real senders for WhatsApp Business API, Signal, Matrix, Discord, SimpleX WS — sharing a unified command set.

### 8.1 Architecture

```
                    Gateway (internal/gateway/)
                    ┌────────────────────────┐
                    │     MultiSender         │
                    │  Sender interface       │
                    └────────┬───────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
  ┌────────────┐    ┌────────────┐    ┌────────────┐
  │ Telegram   │    │ WhatsApp   │    │ Signal     │
  │ Sender     │    │ Sender     │    │ Sender     │
  └────────────┘    └────────────┘    └────────────┘
  ┌────────────┐    ┌────────────┐    ┌────────────┐
  │ Matrix     │    │ Discord    │    │ SimpleX    │
  │ Sender     │    │ Sender     │    │ Sender     │
  └────────────┘    └────────────┘    └────────────┘
```

### 8.2 Platform Integration

| ID | Feature | Description | Effort |
|----|---------|-------------|--------|
| B2.1 | WhatsApp Business API | Meta Cloud API, webhook receiver, template messages | 2w |
| B2.2 | Signal messenger | signal-cli REST API via subprocess, registered number | 2w |
| B2.3 | Matrix federation | Matrix Client-Server API, room management, E2EE | 3w |
| B2.4 | Discord Bot API | Gateway intents, slash commands, voice channels | 2w |
| B2.5 | SimpleX WS gateway | Bidirectional sender via SimpleX CLI WS bridge | 3w |
| B2.6 | Unified command router | Shared handler table, platform-agnostic auth | 2w |
| B2.7 | Platform user registry | Cross-platform identity linking (pubkey → all handles) | 2w |
| B2.8 | Admin cross-post | Send one command → all platforms simultaneously | 1w |

### 8.3 Unified Command Set

| Command | Description | Platform Status |
|---------|-------------|-----------------|
| `/wallet [pubkey]` | Balance | Telegram✅ WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/economy` | Economy state | Telegram✅ WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/send [pubkey] [amt]` | Transfer ng | Telegram🔜 WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/invoice [amt] [desc]` | POS invoice | Telegram🔜 WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/pay [id]` | Pay invoice | Telegram🔜 WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/radio` | Announcements | Telegram✅ WA🔜 Signal🔜 Matrix🔜 Discord🔜 |
| `/king [cmd]` | Admin (royal key) | Telegram✅ WA🔜 Signal🔜 Matrix🔜 Discord🔜 |

---

## 9. Master Timeline

```
Week 1-2:   TRACK 1 (Foundation)
            ├── B1-B5: All critical bugs fixed
            ├── S1-S4: Security hardening
            ├── I1-I5: Graceful shutdown, config, logging, healthchecks
            ├── A4.1: Package extraction (split main.go)
            └── Unit tests for economy/ + lock/ + press/

Week 3-5:   TRACK 2 (Phase A2)
            ├── A2.1: Real TRON monitoring
            ├── A2.2: Auto silver rounds
            ├── A2.3: Dividend claiming
            ├── A2.4: Proof of Reserve API
            ├── A2.5: Royal → Sub SMP control
            ├── A2.6: Marketplace escrow
            ├── A2.7: treasury/radio/channels packages
            ├── A2.8: Real banknote PDF
            ├── A2.9: Dynamic ICE credentials
            └── Integration tests: all fixtures

Week 5-9:   TRACK 2 (Phase A3 — Flywheel)
            ├── A3.1-A3.10: Full flywheel economy
            └── Load testing with k6, fix bottlenecks

Week 9-13:  TRACK 2 (Phase A4 — Franchise)
            ├── A4.1-A4.8: Full franchise system
            └── Franchise node SDK + documentation

Week 13-17: TRACK 2 (Phase A5 — Platform + AI)
            ├── A5.1-A5.5: Service registry + Steward
            ├── A5.7: Gateway unification
            ├── A5.8: Go Telegram bot (replace Python)
            └── Full performance benchmark + report

Week 10-12: TRACK 5 (Client Foundation)
            ├── Flutter monorepo scaffold (apps/royal_app, apps/isle_app, apps/shared)
            ├── Shared API client for all endpoints
            ├── Royal App MVP: dashboard + build/deploy commands
            └── Isle App MVP: Ed25519 login + wallet dashboard

Week 13-15: TRACK 5 (Core Features)
            ├── Royal App: treasury, relay, audit log
            ├── Isle App: contacts, vault, marketplace, notifications
            └── Shared: design system, dark mode, i18n stubs

Week 16-18: TRACK 5 (Advanced Features)
            ├── Royal App: multi-node map, key rotation
            ├── Isle App: digital ID, banknote viewer, auction, radio
            └── Push notifications via SMP

Week 19-20: TRACK 5 (Polish & Release)
            ├── CI/CD for Flutter builds
            ├── App store submission prep + TestFlight
            └── Integration tests against staging node

Week 21-28: TRACK 6 (Phase B1 — Crypto Bridges & ICO)
            ├── B1.1: ARGENTUM Jetton mainnet deployment
            ├── B1.2: Bitcoin atomic swap bridge
            ├── B1.3: Ethereum LayerZero bridge
            ├── B1.4: Solana Wormhole bridge
            ├── B1.5: Polygon/Arbitrum/Base bridges
            ├── B1.6: Bridge Board election + DAO setup
            ├── B1.7: ICO smart contracts + vesting
            └── B1.8: Public PoR dashboard

Week 29-40: TRACK 7 (Phase B2 — Multi-Platform Bots)
            ├── B2.1: WhatsApp Business API integration
            ├── B2.2: Signal messenger bridge
            ├── B2.3: Matrix federation
            ├── B2.4: Discord Bot API
            ├── B2.5: SimpleX WS gateway sender
            ├── B2.6: Unified command router
            ├── B2.7: Cross-platform identity registry
            └── B2.8: Admin cross-post system
```

---

## 10. Key Architectural Decisions

### D1: Config-first approach
All hardcoded paths, intervals, thresholds, and parameters move to `simplex-node.json`. Binary reads `-config` flag, defaults to `$DATA_DIR/config.json`. Hot-reload via SIGHUP or `POST /api/reload`.

### D2: Structured logging with slog
Replace all `fmt.Printf` with `slog.Info/Debug/Warn/Error`. Log format: JSON with timestamp, level, message, request_id, handler. Benefits: searchable, structured, no interleaved output.

### D3: Atomic JSON writes
Every JSON write follows: `write .tmp → f.Sync() → rename .tmp → .json`. Prevents partial writes on crash. Add WAL for high-write paths (ledger transfers, auction bids).

### D4: sync.RWMutex everywhere
All shared state (banknote registry, auction list, P2P offers, channel list, ledger) uses `sync.RWMutex`. Readers don't block each other, only writers serialize.

### D5: API versioning
Start with `/api/v1/...` prefix. Legacy endpoints at `/api/...` get deprecation headers. Enables backward-compatible evolution.

### D6: SQLite migration path
JSON files for A1-A2 (100 users). SQLite at A3 (1,000 users) — migrate ledger + registry. PostgreSQL at A5 (100k users). All wrapped behind same `internal/economy/storage.go` interface.

### D7: Sub-agent testing
Use `spawn_subagent` with `docs/tester-prompt-royal.txt` for autonomous regression testing. Agent launches node, runs curls, asserts math, produces report.

---

*Plan updated: 2026-06-22*

*Full project report: PROJECT-REPORT.md (generated 2026-06-22)*
*Author: AI-architect simplex-node*
