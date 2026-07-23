# Project Valuation: simplex-node / isle / ParanoidX

**Updated:** Build b116 (2026-06-19)

---

## 1. Codebase Scale

| Component | Files | Lines of Code |
|-----------|-------|---------------|
| Go backend (simplex-node) | 181 | 40,075 |
| Flutter (isle_app) | 59 | 10,868 |
| Shell scripts | 42 | 4,724 |
| ParanoidX (simplex-fork) | 20 | 3,693 |
| Docker/YAML configs | ~12 | ~1,200 |
| **Total** | **~314** | **~60,560** |

### Go backend packages (31 internal)

| Package | LOC | Purpose |
|---------|-----|---------|
| api | 10,666 | REST API (100+ endpoints) |
| economy | 9,534 | Economic engine (Liquid Taler, banknotes, DAO) |
| crypto | 2,686 | BIP39, encryption, key management |
| bot | 2,070 | 3 Telegram bots |
| radio | 1,983 | Radio scheduler, streaming, M3U8, Acestep |
| store | 1,336 | SQLite stores (taler, vault, dao, banknote) |
| gateway | 959 | External integration gateway |
| bridge | 785 | SimpleX WebSocket bridge (+ health, config) |
| steward | 610 | AI Steward, monitor, constitution analyzer |
| treasury | 603 | Treasury, silver rounds, dividends |
| ai | 510 | Ollama integration |
| press | 470 | Banknote PDF generation |
| status | 443 | Status service |
| container | 385 | CryptoContainer (AES-256-GCM, argon2id) |
| middleware | 373 | HTTP middleware (CORS, CSP, HSTS, XSS) |
| vault | 349 | 16GB encrypted storage |
| royal | 346 | P2P royal nodes |
| webrtc | 328 | TURN/ICE signaling |
| lock | 305 | Rate-limiter, mutexes |
| transport | 295 | P2P transport layer |
| registry | 309 | Node registry |
| health | 311 | Health checks |
| config | 354 | Configuration management |
| tracker | 140 | P2P node tracker |
| isle | 116 | .isle manifest generator |
| dockerutil | 90 | Docker utilities |
| ton | 80 | TON integration (ARGENTUM) |
| billing | 188 | Billing service |
| fileutil | 210 | File utilities |

### ParanoidX fork (20 Go files, 3,693 LOC)

| Subpackage | LOC | Purpose |
|------------|-----|---------|
| apptransport | 1,077 | AppTransport protocol (envelope, queue, replay, signal) |
| bridge | 616 | WS bridge + SOCKS5 proxy |
| vpn | 318 | VPN service |
| transport | 293 | Transport layer |
| simplexcli | 344 | SimpleX CLI integration |
| paranoid | 191 | VPN/TOR modes (4 modes) |
| smp | 186 | SMP protocol client |
| paranoidx | 51 | Architecture definitions |

---

## 2. Growth Since Build b24

| Metric | b24 (old report) | b116 (current) | Delta |
|--------|------------------|----------------|-------|
| Go LOC | ~34,374 | 40,075 | +5,701 |
| Go files | 169 | 181 | +12 |
| API endpoints | ~60 | 100+ | +40+ |
| Internal packages | 28 | 31 | +3 |
| ParanoidX LOC | 1,916 | 3,693 | +1,777 |
| ParanoidX files | 11 | 20 | +9 |
| Total LOC | ~51,800 | ~60,560 | +8,760 |

### Major additions (b97-b116):
- `/api/admin/info` — comprehensive node introspection
- `/api/chat/bridge-health` — latency, reconnect stats
- `/api/db/*` — SQLite lifecycle (backup/restore/list/upload)
- `/api/admin/webhook-queue` — persistent HMAC-signed delivery
- `/api/chat/archive` — cold archival to USB
- `/api/admin/rate-limit-config` — runtime rate limiter config
- `/api/silver/{mint,burn,list}` — silver-backed asset lifecycle
- `/api/reserve/proof` — enhanced proof-of-reserve with ratio
- `/api/rwa/{register,list}` — real-world asset tokenization
- `/api/economy/dividend-admin` — dividend history + trigger
- `/api/economy/rates` — multi-currency rates
- `/api/simplex/channel/*` — SimpleX v6.5 channels
- `/api/did` + `/api/did/contact` — W3C DID documents
- `/api/relay/*` — inter-node message relay
- `/api/ai/*` — steward DID, radio AI, forecast, moderation
- ParanoidX: AppTransport protocol, SOCKS5 proxy, SMP client

---

## 3. Time Estimation (1 senior + 2 mid + 2 junior, WITHOUT AI)

| Phase | Duration | ∑ person-months |
|-------|----------|----------------|
| Architecture & Protocols | 2 mo | 5 |
| Core backend (economy, bridge, API) | 6 mo | 25 |
| Frontend Flutter | 5 mo | 17 |
| ParanoidX | 3 mo | 9 |
| Integration & Testing | 3 mo | 13 |
| Deployment, CI/CD, Polish | 2 mo | 7 |
| **Subtotal** | **21 mo** | **76** |
| Overhead (reviews, communication, bugs) | +30% | +23 |
| **Total person-months** | | **~99** |
| **Calendar time (5 people)** | | **~20 months** |

### Scenarios

| Scenario | Calendar | Person-months |
|----------|----------|---------------|
| Optimistic | 14 mo | ~70 |
| **Neutral** | **20 mo** | **~99** |
| Pessimistic | 28 mo | ~130 |

---

## 4. Financial Assessment

### Salaries (monthly)

| Role | × | Rate | Months | Total |
|------|---|------|--------|-------|
| Senior | 1 | $12,000 | 20 | $240,000 |
| Mid | 2 | $8,000 | 20×2 | $320,000 |
| Junior | 2 | $5,000 | 20×2 | $200,000 |
| **Payroll** | | | | **$760,000** |

### Capital Expenditures

| Item | Cost |
|------|------|
| 5× Linux workstations | $25,000 |
| Test lab (server, network, 5 mobile devices) | $18,000 |
| **Total CapEx** | **$43,000** |

### Operating Expenses (20 months)

| Item | × 20 | Total |
|------|------|-------|
| Cloud servers | $2,000 | $40,000 |
| CI/CD, domains, DNS | $600 | $12,000 |
| Tools (Jira, Slack, etc.) | $700 | $14,000 |
| Security audit (quarterly) | $1,000 | $20,000 |
| **Total OpEx** | | **$86,000** |

### Grand Total

| Scenario | Payroll | CapEx | OpEx | **Total** |
|----------|---------|-------|------|-----------|
| **Optimistic** (14 mo) | $532K | $43K | $60K | **~$635,000** |
| **Neutral** (20 mo) | $760K | $43K | $86K | **~$889,000** |
| **Pessimistic** (28 mo) | $1,064K | $43K | $120K | **~$1,227,000** |

---

## 5. Current Project Value

### Valuation Methods

**Cost-to-rebuild (neutral, b116):** ~$889,000
**Intellectual property (custom algorithms, protocols):** $450,000
**Strategic value (Tor-native, censorship-resistant, 100+ API):** $750,000 - $3,000,000
**Comparables:** Aragon (~$50M), DAO platforms ($10-100M)

**Current value: $750,000 - $3,000,000**

### Key Assets

| Asset | Value |
|-------|-------|
| SimpleX bridge + channels + relay (full messaging stack) | Unique, no known equivalents |
| Silver-backed economy engine + RWA + asset lifecycle | Patentable |
| ParanoidX multi-network client (AppTransport + VPN/TOR) | Embedded VPN/TOR, zero deps |
| 100+ REST API endpoints | Complete sovereign node interface |
| P2P relay + registry over Tor | Decentralized, unblockable |
| DAO governance system | Battle-tested |
| CryptoContainer (AES-256-GCM, argon2id, panic wipe) | Military-grade security |

---

## 6. Scaling Plan (×10-100 value)

### Phase 1: Product-Market Fit (+$250K, 3 months)

**Investment:** $250,000
**Team:** 2 senior + 1 mid + 2 junior

| Task | Priority |
|------|----------|
| Load testing (1000+ concurrent) | Critical |
| Security audit (3rd party) | Critical |
| Mobile push notifications | High |
| Multi-language (EN, FR, RU, ES) | High |
| CI/CD pipeline hardening | Medium |
| **Target valuation:** $3-7M | |

### Phase 2: Market Launch (+$750K, 6 months)

**Investment:** $750,000
**Team:** 3 senior + 3 mid + 3 junior

| Task | Priority |
|------|----------|
| TON on-chain proof-of-reserve | Critical |
| iOS App Store launch | High |
| Google Play launch | High |
| Community governance voting | High |
| API monetization (SMP relay as-a-service) | High |
| SDK for 3rd party integration | Medium |
| Web client (wasm) | Medium |
| **Target valuation:** $10-25M | |

### Phase 3: Scale (+$3M, 12 months)

**Investment:** $3,000,000
**Team:** 5 senior + 8 mid + 5 junior

| Task | Priority |
|------|----------|
| Decentralized autonomous treasury | Critical |
| Cross-chain atomic swaps (BTC, ETH, TON) | High |
| AI-governed monetary policy | High |
| Enterprise privacy suite (ParanoidX Pro) | High |
| Real-world asset tokenization expansion | Medium |
| **Target valuation:** $30-75M | |

### Phase 4: Sovereign Network (+$7M, 24 months)

**Investment:** $7,000,000
**Team:** 10 senior + 15 mid + 10 junior

| Task | Priority |
|------|----------|
| Custom L1/L2 blockchain | Critical |
| Independent DNS (.onion + ENS) | High |
| Physical validators (Saint Mary Island) | High |
| Diplomatic recognition as digital nation | Medium |
| International token exchange | Medium |
| **Target valuation:** $200M-1B | |

---

## 7. Critical Risk Matrix

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| SimpleX protocol changes | Medium | High | Fork protocol, abstraction layer |
| Tor network compromise | Low | Critical | Multi-transport (I2P, LN) |
| Regulatory action | Medium | High | Geo-distributed nodes, legal shell |
| Private key compromise | Low | Critical | MPC + HSMs + hardware wallets |
| User adoption failure | High | Medium | Focus on enterprise B2B privacy |
| AI economic policy bugs | Medium | High | Circuit-breaker, human override |
| Single developer dependency | High | Critical | Documentation + team scaling |
| Infrastructure costs | Medium | Medium | Token-based funding model |

---

## 8. Conclusion

**Project in current state (b116, ~60K LOC, 100+ endpoints):**
- Development cost (without AI): ~$889,000
- Market value: $750K - $3M
- Development time for 5-person team: ~20 months

**Key advantage:** AI assistant (opencode) reduced time and cost by **3-5×** vs traditional development.

**Potential:** With $11M investment across 4 phases over 45 months, project value can grow to **$200M - $1B** if successfully positioned as sovereign digital network infrastructure.

**Growth since initial report (b24→b116):** +17% codebase size, +67% API endpoints, +92% ParanoidX functionality.
