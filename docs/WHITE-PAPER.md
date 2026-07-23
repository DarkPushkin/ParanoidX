# Saint Mary Liberty Island — White Paper

## The World's First Private Sovereign Digital Nation with Silver-Backed Economy

**Version:** 1.0 | **Date:** June 2026 | **Codename:** "The Isle" / "Остров"
**Contact:** admin@stmaria.org | **Protocol:** SimpleX + Tor

---

## Executive Summary

Saint Mary Liberty Island is building the infrastructure for a **sovereign digital nation** — a private, encrypted communication and economic network where people can communicate, transact, and govern themselves without surveillance, censorship, or intermediary risk. At its core is **simplex-node**, a single Go binary that powers the entire stack: private messaging (SimpleX/Tor), a silver-backed digital currency (Liquid Taler), AI-governed economic policy (Steward), merchant payment tools (POS), and multi-chain crypto bridges (TRON/TON).

We are not a blockchain startup. We are a **full-stack digital sovereignty platform** — the operating system for a nation without territory.

---

## 1. The Problem

### 1.1 Privacy Is Dead
Every message, every transaction, every financial move is tracked, logged, and monetized. End-to-end encryption is under attack globally. The five eyes, nine eyes, fourteen eyes — the surveillance apparatus grows daily.

### 1.2 Money Is Not Sound
- Fiat currencies lose 2-7% purchasing power annually
- Central bank digital currencies (CBDCs) threaten programmable money control
- Cryptocurrencies are volatile, complex, and surveillable
- Physical silver is impractical for digital payments

### 1.3 Digital Nations Don't Exist
No existing project combines **private messaging + silver-backed stablecoin + self-custody + merchant tools + AI governance** into one integrated platform. Users must cobble together 5-10 different services, each with its own security model, terms of service, and surveillance risk.

---

## 2. The Solution: The Isle

### 2.1 Architecture Overview

```
                  ┌─────────────────────────────────────┐
                  │          The Isle (simplex-node)     │
                  │  ┌─────────┐  ┌──────────┐          │
                  │  │ Gateway │  │  Core    │          │
                  │  │ Telegram│  │ Economy  │          │
                  │  │ Signal  │  │ Treasury │          │
                  │  │ WhatsApp│  │ Vault    │          │
                  │  │ Matrix  │  │ POS      │          │
                  │  │ Discord │  │ Steward  │          │
                  │  └─────────┘  └──────────┘          │
                  │       │             │               │
                  │  ┌────▼─────────────▼───────┐       │
                  │  │    Blockchain Bridges     │       │
                  │  │  TRON (USDT) | TON (Jetton)│      │
                  │  │  Bitcoin | Ethereum | Solana │     │
                  │  └────────────────────────────┘       │
                  └───────────────────────────────────────┘
```

### 2.2 Key Components

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Messaging** | SimpleX SMP + XFTP + Tor | Private E2EE messaging, file transfer |
| **Voice/Video** | WebRTC + Coturn TURN | Encrypted calls via hidden relay |
| **Currency** | Liquid Taler (ng) | Silver-backed digital unit |
| **Storage** | Vault (2GB E2EE) | Self-custodial file store |
| **Governance** | AI Steward | Constitution-based economic policy |
| **Payments** | POS Terminal | Merchant invoices, vouchers, QR codes |
| **Crypto Bridges** | TRON/TON → ng | On-ramp from external chains |
| **AI Assistant** | AskSteward (Ollama) | Self-hosted LLM (gemma4) |
| **Bot Fleet** | Telegram (Go) | Interactive admin & user interfaces |

### 2.3 Competitive Advantage

No existing project occupies this unique intersection:

| Need | Solutions | Our Advantage |
|------|-----------|---------------|
| Private messaging | Signal, SimpleX, Telegram | No phone/email required |
| Silver-backed currency | Kinesis, Swiss America | Self-sovereign, not custodial |
| AI governance | — | Constitutional AI Steward |
| Merchant tools | Square, Stripe | 1% fees, privacy-first |
| Multi-chain bridge | Various DEXs | Integrated into sovereign platform |

**Competitive Gap:** We combine ALL of these into ONE sovereign node that YOU control.

---

## 3. The Liquid Taler Economy

### 3.1 Monetary Policy

- **1 Liquid Taler (TLR)** = 31,103,480,000 ng (nanograms) = 1 troy oz of silver at spot
- **Silver Backing Ratio:** 70% physical silver in reserve
- **Utility Premium:** 30% network value (speed, privacy, convenience)
- **Total Fee Cap:** 4.20% (treasury takes 2.28%, max)
- **Treasury Commission:** Exactly 228 BPS (2.28%) on every transaction — no exceptions

### 3.2 Issuance Model (70 oz → 100 TLR)

When a silver round is initiated:

| Allocation | % | Recipient |
|-----------|---|-----------|
| Investor | 70.0% | The party who paid USDT |
| Treasury | 2.4% | Node operational budget |
| Dividend Pool | 14.1% | Banknote holders (by rarity weight) |
| Silver Buy | 3.0% | Future silver purchases |
| Auction Pool | 4.5% | Auction house liquidity |
| Buyback Pool | 6.0% | Banknote buyback reserve |

### 3.3 Deflation Mechanisms

- **Treasury burn:** 20-40% at fat treasury tiers
- **Subscription burn:** 30% of subscription revenue
- **Auction fee burn:** 5% of auction fees
- **Advertising burn:** 20% of ad tag purchases

These deflationary forces create a built-in appreciation pressure on the ng.

---

## 4. ICO: Genesis Token Sale

### 4.1 The ARGENTUM Jetton (TON)

**ARGENTUM** is a TON blockchain Jetton pegged 1:1 with Liquid Taler (ng):

- Symbol: **ARGENTUM**
- Decimals: 9 (same as ng)
- Backing: 70% physical silver
- Swap fee: 0.5% (50 BPS)
- Min swap: 1,000,000 ng (~$0.0024)

The Jetton master contract will be deployed on TON mainnet, enabling:
- Trade on TON DEXs (STON.fi, DeDust)
- Hold in any TON wallet (Tonkeeper, Wallet)
- Bridge back to ng on The Isle at any time (1:1)

### 4.2 Sale Tiers

| Tier | Min Investment | Bonus | Vesting | Allocation |
|------|---------------|-------|---------|------------|
| **Genesis Angel** | 100,000 USDT | 30% | 6-month cliff, 12-month linear | First come, first served |
| **Major Investor** | 10,000 USDT | 20% | 3-month cliff, 9-month linear | By whitelist |
| **Minor Investor** | 1,000 USDT | 10% | 1-month cliff, 6-month linear | Public |
| **Citizen** | 100 USDT | 5% | No cliff, 3-month linear | Public |

### 4.3 Fund Allocation

| Use | % of Raise |
|-----|-----------|
| Silver Reserve Purchase | 50% |
| TON Jetton Liquidity Pools | 15% |
| Development Fund (2 years) | 15% |
| Marketing & Ecosystem | 10% |
| Legal & Compliance | 5% |
| Reserve (operations) | 5% |

### 4.4 Token Distribution (Max Supply: 1,000,000 TLR)

| Category | % of Supply | TLR | Vesting |
|----------|------------|-----|---------|
| ICO Sale | 40% | 400,000 | Per tier schedule |
| Treasury Reserve | 20% | 200,000 | DAO-controlled |
| Silver Reserve | 15% | 150,000 | Matching physical silver |
| Team & Advisors | 10% | 100,000 | 2-year linear vesting |
| Ecosystem Fund | 10% | 100,000 | DAO grants |
| Community Airdrops | 5% | 50,000 | Post-ICO distribution |

---

## 5. Multi-Chain Crypto Bridges

### 5.1 Architecture

```
 External Chain        Bridge Layer           The Isle
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ TRON     │──────▶│ USDT Monitor  │──────▶│ Treasury │
 │ USDT     │       │ (TronGrid)   │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ TON      │◀─────▶│ ARGENTUM     │◀─────▶│ ng       │
 │ Jetton   │       │ Swap Engine  │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Bitcoin  │──────▶│ BTC Bridge    │──────▶│ Treasury │
 │ (future) │       │ (Atomic Swap)│       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Ethereum │──────▶│ ETH Bridge    │──────▶│ Treasury │
 │ (future) │       │ (LayerZero)  │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Solana   │──────▶│ SOL Bridge    │──────▶│ Treasury │
 │ (future) │       │ (Wormhole)   │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
```

### 5.2 Current Bridges (Live)

| Chain | Asset | Mechanism | Status |
|-------|-------|-----------|--------|
| **TRON** | USDT TRC20 | TronGrid polling, auto-log deposits | ✅ Live (Cycle 30) |
| **TON** | ARGENTUM Jetton | Swap stub, contract pending | 🔄 Pre-launch |

### 5.3 Planned Bridges (Phase B1 — Post-ICO)

| Chain | Asset | Mechanism | Effort |
|-------|-------|-----------|--------|
| **Bitcoin** | BTC | Atomic swaps + Lightning Network | 4 weeks |
| **Ethereum** | ETH, USDC, USDT | LayerZero OFT / Chainlink CCIP | 6 weeks |
| **Solana** | SOL, USDC | Wormhole bridge | 4 weeks |
| **Polygon** | MATIC, USDC | Polygon PoS bridge | 3 weeks |
| **Arbitrum** | ETH, USDC | Arbitrum bridge | 3 weeks |
| **Base** | ETH, USDC | Base bridge (Coinbase) | 2 weeks |
| **Binance Smart Chain** | BNB, USDT | BSC bridge | 3 weeks |

### 5.4 Board of Bridges (DAO Governance)

Each bridge will be governed by a **Bridge Board** — a multi-sig committee elected by ARGENTUM holders:

- **Bridge validators**: 5-7 members, elected quarterly
- **Signing threshold**: 4/7 multi-sig
- **Bridge fees**: 0.1% per cross-chain transaction (goes to dividend pool)
- **Emergency pause**: 3/7 can pause any bridge
- **Audit cycle**: Monthly PoR verification by elected auditors

The Bridge Board is the first step toward full DAO governance of The Isle.

---

## 6. Multi-Platform Messaging Bots

### 6.1 Architecture

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

### 6.2 Implementation Plan (Phase B2 — 12 weeks)

| Platform | Library/API | Auth Method | Effort |
|----------|------------|-------------|--------|
| **Telegram** | Bot API (existing) | Bot token | ✅ Done |
| **WhatsApp** | WhatsApp Business API | Phone + API key | 2 weeks |
| **Signal** | signal-cli REST API | Registered number | 2 weeks |
| **Matrix** | Matrix Client-Server API | Access token | 3 weeks |
| **Discord** | Discord Bot API | Bot token | 2 weeks |
| **SimpleX** | CLI WS bridge (pending) | SMP queue | 3 weeks |

### 6.3 Unified Command Set

All platforms share the same command handlers:

| Command | Description | Access |
|---------|-------------|--------|
| `/wallet [pubkey]` | Liquid Taler balance | Public |
| `/economy` | Treasury, reserve, banknotes | Public |
| `/send [pubkey] [amount]` | Transfer ng | Authenticated |
| `/invoice [amount] [desc]` | Create POS invoice | Merchant |
| `/pay [invoice_id]` | Pay invoice | Authenticated |
| `/radio` | List audio announcements | Public |
| `/vault` | File storage | Authenticated |
| `/market` | Marketplace listings | Public |
| `/auction` | Active auctions | Public |
| `/king [command]` | Admin controls | Royal key only |

---

## 7. Governance: The AI Steward → DAO Transition

### Phase 1: AI Steward (Current — Cycle 38)

- **Constitution**: 16 hardcoded rules with min/max/target bounds
- **Monitor**: 60-second metric collection (15+ metrics)
- **Analyzer**: 3-level deviation detection (minor/major/critical)
- **Decision Engine**: Auto-adjust (minor), notify admin (major), require consensus (critical)
- **Dynamic Params**: 9 adjustable economy parameters persisted on disk

### Phase 2: AI + Human Board (Post-ICO)

- Elected Bridge Board (5-7 members)
- Auditor council (elected by banknote holders)
- AI makes proposals, human board confirms critical decisions
- Monthly governance votes

### Phase 3: Full DAO (Target: 2027)

- Weighted voting by ARGENTUM holdings
- On-chain proposal system
- Treasury management DAO
- Constitutional amendments via referendum

---

## 8. Roadmap

### Phase A (Now — Complete)
- ✅ 70% Silver Backing Model
- ✅ Real TRON USDT monitor
- ✅ AI Steward Core (constitution, monitor, analyzer)
- ✅ Dynamic economy parameters
- ✅ POS Terminal with QR invoices
- ✅ Royal→Sub node protocol
- ✅ Banknote PDF with Ed25519 double-signature
- ✅ Onboarding funnel + subscription tiers
- ✅ Gateway unification module
- ✅ 80+ API endpoints
- ✅ 18/18 Go packages test green
- ✅ 4 Telegram bots

### Phase B1 (Post-ICO — Q3 2026)
- 🔄 ARGENTUM TON Jetton deployment + liquidity pools
- 🔄 ICO smart contracts + vesting
- 🔄 Bitcoin atomic swap bridge
- 🔄 Ethereum LayerZero bridge
- 🔄 Solana Wormhole bridge
- 🔄 Bridge Board election
- 🔄 Proof of Reserve dashboards (public)

### Phase B2 (Q4 2026)
- 🔄 WhatsApp Business API integration
- 🔄 Signal messenger bridge
- 🔄 Matrix federation
- 🔄 Discord bot marketplace
- 🔄 Flutter mobile app (iOS + Android)
- 🔄 WebRTC native mobile support

### Phase C (2027)
- 🔄 Full DAO governance
- 🔄 PostgreSQL migration (from JSON files)
- 🔄 100,000 user scalability
- 🔄 Multi-franchise network
- 🔄 Real-world asset tokenization platform
- 🔄 Physical silver redemption

---

## 9. Token Utility

| Utility | ARGENTUM (TON) | Liquid Taler (ng) |
|---------|---------------|--------------------|
| Silver-backed store of value | ✅ 1:1 with ng | ✅ 70% physical backed |
| Transaction fees | ❌ | ✅ 2.28% treasury fee |
| Governance votes | ✅ Weighted by holdings | ❌ |
| Cross-chain value | ✅ Trade on TON DEXs | ❌ Internal only |
| Dividend accrual | ✅ (via ng backing) | ✅ (on banknotes) |
| POS payments | ❌ | ✅ Merchant payments |
| Staking/DeFi | ✅ (TON ecosystem) | ❌ |

---

## 10. Risk Factors

| Risk | Mitigation |
|------|-----------|
| Silver price volatility | 70% backing ratio buffers 30% price drop |
| TON chain disruption | Multi-chain strategy, not TON-dependent |
| AI governance failure | Human override + constitutional bounds |
| Regulatory uncertainty | Decentralized, non-custodial, no KYC |
| USDT de-pegging | Diversifying to USDC, DAI, native crypto |
| SimpleX API changes | Pinned CLI version in Docker |
| Disk failure | Daily backups to USB + cloud |

---

## 11. Team

The Isle is built by a distributed team of privacy engineers, monetary economists, and AI researchers operating under the **Saint Mary Liberty Island** sovereign project. Key roles:

- **King Tomas** — Founder, system architect
- **AI Steward** — Constitutional AI governance
- **AskSteward (AI)** — Public-facing AI assistant
- **Inquisitor (AI)** — Production cycle QA automation

Development is funded through subscription revenue, treasury commissions, and the Genesis ICO.

---

## 12. Legal Structure

Saint Mary Liberty Island operates as a **digital sovereign entity** under the stmaria.org / markbank.org umbrella. The project is not a company, not a foundation — it is a **network of sovereign nodes** governed by software.

- **No KYC/AML**: The protocol is self-sovereign
- **No custodial risk**: Users control their keys
- **No securities**: Liquid Taler is a utility token backed by physical silver
- **Legal opinion**: Retained counsel for ICO structuring (Gibraltar/Dubai/Switzerland)

---

## 13. Call to Action

### For Minor Investors ($100+)
Participate in the Citizen tier of the Genesis ICO. Receive ARGENTUM Jettons on TON, vesting over 3 months, with full access to The Isle ecosystem.

- **Website:** https://stmaria.org
- **Node:** simplex-node at localhost:8080 (onion accessible)
- **Telegram:** @AskSteward_bot | @torquemada878_bot

### For Major Investors ($10,000+)
Whitelist for Major Investor tier. Direct communication with King Tomas. Bridge Board seat eligibility. First access to franchise node licenses.

### For Institutional Investors ($100,000+)
Genesis Angel tier. Custom vesting schedule. Dedicated bridge validator node. Preferred allocation in future franchise rounds.

---

*"Private messages. Private money. Silver backing."*

**Saint Mary Liberty Island**
stmaria.org — markbank.org
June 2026
