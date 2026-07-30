# ParanoidX

**Sovereign Go HTTP Server + Docker Infrastructure for Saint Mary Liberty Island**

> *"Code lives. Evolution is infinite. Silver is our conscience."*

---

## Overview

ParanoidX (ex. simplex-node) is the core backbone of **Saint Mary Liberty Island** — a silver-backed digital economy operating over SimpleX Chat and Tor. It provides 283+ API endpoints across wallet, market, treasury, governance, chat, radio, and ParanoidX routing layers.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ParanoidX Core (Go)                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────────┐  │
│  │ Economy  │ │  Chat    │ │  Radio   │ │  ParanoidX     │  │
│  │ Treasury │ │  Bridge  │ │ Scheduler│ │  Routing     │  │
│  └──────────┘ └──────────┘ └──────────┘ └────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐   ┌──────────────┐ ┌──────────┐
        │  Tor     │   │   Docker     │ │  SimpleX │
        │  (9050)  │   │  (5 svcs)    │ │  Chat    │
        └──────────┘   └──────────────┘ └──────────┘
```

## Key Features

| Layer | Capabilities |
|-------|--------------|
| **Wallet** | Create, balance, send, mint, redeem, banknotes, dividends |
| **Market** | Listings, escrow, orders, buy/sell, auctions |
| **Treasury** | Reserve state, oracle price, deflation engine, auto-mint, dividends |
| **Governance** | Constitution, proposals, voting, delegation |
| **Chat** | History, send/edit/delete, pin, react, broadcast, search |
| **Radio** | Schedule, AI content generation, streaming |
| **ParanoidX** | Multi-layer routing (VLESS/VMess/Tor), VPN management |
| **Admin** | Metrics, Docker control, backup, config management |

## Economic Parameters

```
NGPerTLR = 31,103,480,000 ng
SilverSpotUSDperOZ = 75.0
SilverBackingRatio = 0.70 (70% physical silver)
UtilityPremiumPct = 0.30 (30% digital convenience premium)

Issuance (70 oz → 100 TLR):
  Investor 70% | Treasury 4.2% | DividendPool 12.9%
  SilverBuy 3% | AuctionPool 6.6% | Buyback 3.3%
```

## Quick Start

```bash
# Build
go build ./cmd/ParanoidX/

# Run via systemd (production)
sudo systemctl start ParanoidX-dashboard.service

# Health check
curl http://127.0.0.1:8080/api/health
```

## Configuration

- **Config file:** `~/.local/share/simplex-node/simplex-node.json`
- **Chat history:** `~/.local/share/simplex-node/chat_history.json`
- **Radio data:** `~/.local/share/simplex-node/radio/`
- **Docker volumes:** `/home/tomas/ParanoidX/docker/`

## Docker Services (5)

| Service | Port | Description |
|---------|------|-------------|
| `ParanoidX-smp` | 17225 | SimpleX Chat bridge |
| `ParanoidX-xftp` | 17226 | File transfer |
| `ParanoidX-v2ray` | 1080/1081 | VLESS/VMess proxy |
| `ParanoidX-tor` | 9050 | Tor SOCKS5 |
| `ParanoidX-turn` | 3478 | TURN/STUN for WebRTC |

## Telegram Bots

- **@AskSteward_bot** — AI assistant (Ollama: gemma4:latest)
- **@Inquisitor_bot** — Production cycle reports
- **@Royal_bot** — Admin control panel
- **@Poet_bot** — Radio content generation

## Development

```bash
# Run tests
go test ./... -short -count=1 -timeout 30s

# Vet
go vet ./...

# Build for deployment
go build -o /home/tomas/bin/ParanoidX ./cmd/ParanoidX/
sudo systemctl restart ParanoidX-dashboard.service
```

## Project Structure

```
ParanoidX/
├── cmd/ParanoidX/           # Main entrypoint
├── internal/
│   ├── economy/             # Treasury, wallet, market
│   ├── chat/                # SimpleX bridge, history
│   ├── radio/               # Scheduler, AI generation
│   ├── paranoidx/           # VPN, routing, Docker
│   ├── ai/                  # Ollama integration
│   ├── api/                 # HTTP handlers (283+ endpoints)
│   ├── config/              # Configuration management
│   └── ...                  # 18 internal packages
├── docker/                  # Docker Compose + configs
├── scripts/                 # Deployment & utility scripts
├── MANIFEST.md              # Detailed project manifest
└── AGENTS.md                # Agent instructions
```

## Related Repositories

- **the-grimoire** — Evolution Protocol, docs, skills, bootstrap
- **The-Isle** — Civilian Flutter app
- **Royal-Isle** — Admin Flutter app
- **shared-libs** — Shared Dart packages (models, api_client, widgets)

---

*Saint Mary Liberty Island — ParanoidX Division*