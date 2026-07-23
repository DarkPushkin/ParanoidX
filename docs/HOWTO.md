# The Isle — User Guide

**Saint Mary Liberty Island** — Private sovereign network with silver-backed digital currency.

## What Is The Isle?

The Isle is a private messaging + digital money platform built on the SimpleX protocol. It has three layers:

1. **Silver-backed digital currency** (Liquid Taler / TLR) — 70% backed by physical silver, 30% utility premium
2. **Private messaging** via SimpleX SMP + XFTP over Tor onion services
3. **Desktop client** (Linux) for wallet, vault, marketplace, POS terminal, and radio

## Quick Start

### 1. Download the Desktop App

```bash
# Download the latest release
wget https://example.com/the-isle-latest.tar.gz
tar -xzf the-isle-latest.tar.gz
cd the-isle
sudo ./install.sh
```

The app appears in your application menu as "The Isle".

### 2. Get a Public Key

You need an Ed25519 keypair to use the network. Generate one:

```bash
# From the simplex-node source directory
go run cmd/genkey/main.go
```

This prints your **public key** (share this to receive funds) and **private key** (never share this — keep it secret).

Your public key is your account address on the network.

### 3. Connect

Launch the app and click the gear icon (Settings) or press `Ctrl+,`:

- **Server URL**: `http://your-server:8080` (replace with your node address)
- **Public Key**: Paste your Ed25519 public key
- Click **Connect**

Or launch from terminal with CLI arguments:

```bash
the-isle --server http://192.168.1.100:8080 --pubkey ed25519_pubkey_here
```

## Features

### Dashboard
Shows your Liquid Taler balance and economy state (total supply, reserve, banknotes in circulation).

### Wallet
View your balance and transaction history. Send Liquid Taler to other users (2.28% treasury fee).

### Vault
Secure file storage — upload, download, and delete files (2 GB quota). Files can be encrypted with AES-256-GCM via the API.

### Marketplace
Browse and trade real-world assets (RWA), silver-backed banknotes, and digital goods. Escrow system for safe P2P trading.

### POS Terminal
For merchants: create invoices, accept payments, generate QR codes. Processing fee: 1% (100 bps). Invoices expire after 30 minutes. Voucher codes for offline-capable payments.

### Radio
Listen to 5 stations (English, Russian, Spanish + Torquemada Monitor + Steward AI). Make announcements as King, Torquemada, or Steward.

## Keyboard Shortcuts (Desktop)

| Shortcut | Action |
|----------|--------|
| `Ctrl+1` | Dashboard |
| `Ctrl+2` | Wallet |
| `Ctrl+3` | Vault |
| `Ctrl+4` | Market |
| `Ctrl+5` | POS Terminal |
| `Ctrl+6` | Radio |
| `Ctrl+R` | Reconnect to server |
| `Ctrl+Q` | Quit |
| `Ctrl+,` | Settings |

## Connection Checklist

- [ ] Server is running: `systemctl status simplex-node-dashboard`
- [ ] Server URL is reachable: `curl http://server:8080/api/status`
- [ ] Ed25519 public key is valid (32 bytes, hex-encoded)
- [ ] Firewall allows port 8080 (or your custom port)
- [ ] For remote access: Tor onion URL or Tailscale IP

## API Overview

The node exposes 80+ API endpoints. Key ones for users:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/status` | GET | Node health, uptime, versions |
| `/api/economy/state` | GET | Economy state (supply, reserve) |
| `/api/treasury/state` | GET | Treasury state, silver reserve |
| `/api/vault/list` | GET | List vault files |
| `/api/vault/upload` | POST | Upload file to vault |
| `/api/market/list` | GET | Marketplace listings |
| `/api/pos?action=create-invoice` | POST | Create POS invoice |
| `/api/wallet/send` | POST | Send Liquid Taler |
| `/api/wallet/history` | GET | Transaction history |

Full API at `http://server:8080/api/...` (requires local or onion access).

## Security Notes

- Your Ed25519 private key never leaves your device
- All connections go through Tor onion services when using SimpleX
- Vault files can be encrypted with AES-256-GCM
- Rate limiting on unlock attempts (1 req/min, burst 5)
- Treasury commission: 2.28% on all transactions (max total: 4.20%)
- Excess fees above 2.28% go to dividend pool

## Troubleshooting

**"Connection error"** — Server unreachable. Check:
- Is the server running? `systemctl status simplex-node-dashboard`
- Is the URL correct? Try `curl http://server:8080/api/status`
- Are you on the same network? For remote: use Tailscale or Tor onion

**"Insufficient balance"** — You need Liquid Taler (TLR) to transact.
- Ask the node admin to credit your account
- Or receive from another user

**"File too large"** — Vault uploads limited to 50 MB per file, 2 GB total.

**App won't start** — Missing dependencies:
```bash
sudo apt install libgtk-3-0 liblzma-dev libayatana-appindicator3-dev
```

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  The Isle    │────▶│  simplex-node │────▶│  Silver       │
│  Desktop App │     │  (Go server)  │     │  Reserve      │
│  (Flutter)   │◀────│   :8080       │     │  (physical)   │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                    ┌───────▼───────┐
                    │  SimpleX SMP  │
                    │  (messaging)  │
                    └───────────────┘
```

- **The Isle App**: Flutter desktop client (Linux)
- **simplex-node**: Go backend server with REST API
- **Silver Reserve**: Physical silver backing the Liquid Taler
- **SimpleX**: Private messaging protocol over Tor

## Getting Help

- **AskSteward bot** (Telegram): @AskSteward_bot — AI guide to the economy
- **Torquemada bot** (Telegram): Admin notifications and node control
- **Documentation**: `http://server:8080/docs` (if local/onion access)
- **White Paper**: Available via the docs API
