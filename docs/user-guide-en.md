# simplex-node / Isle / ParanoidX — User Guide (EN)

**Build:** b116 (Go), b81 (Flutter) | **Date:** 2026-06-19

---

## 1. Overview

simplex-node is a sovereign digital network node combining:
- SimpleX-based private messaging (bridge to simplex-chat CLI)
- Silver-backed digital economy (Liquid Taler)
- Decentralized governance (DAO)
- Radio streaming
- AI steward
- CryptoContainer vault
- ParanoidX fork (embedded VPN/TOR, multi-network client)
- All behind Tor onion services

### System Architecture

```
simplex-node (Go :8080)
  ├── /api/* — REST API (100+ endpoints)
  ├── bridge — WebSocket ↔ simplex-chat CLI (:17225)
  ├── economy — Liquid Taler, banknotes, DAO, treasury
  ├── radio — Scheduler, streaming, M3U8
  ├── steward — AI agent (Ollama)
  ├── container — CryptoContainer (AES-256-GCM + argon2id)
  └── docker stack — Tor, SMP, XFTP, Coturn
        └── 🧅 Tor onion services (5 hidden services)

ParanoidX fork (/home/tomas/simplex-fork)
  ├── paranoid — Embedded VPN/TOR (4 modes)
  ├── apptransport — AppTransport protocol
  ├── bridge — SOCKS5 proxy + WS bridge
  ├── vpn — VPN service
  └── simplexcli — SimpleX CLI client
```

---

## 2. Chat & Messaging

### 2.1 Contacts
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/address/create` | GET | Create local chat address |
| `/api/chat/contacts` | GET | List all contacts |
| `/api/chat/contact?id=@N` | GET | Single contact details |
| `/api/chat/contact/info` | GET | Contact info (count + last_message) |
| `/api/chat/contact/alias` | POST | Rename contact (`id` + `alias`) |
| `/api/chat/qr` | GET | QR code for a contact |
| `/api/chat/connect` | POST | Connect via invitation string |

### 2.2 Messages
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/history` | GET | Full history (optional `chat_id`) |
| `/api/chat/stream` | GET | SSE real-time stream |
| `/api/chat/send` | POST | Send message (`chat_id` + `text`) |
| `/api/chat/edit` | POST | Edit message (`id` + `new_text`) |
| `/api/chat/delete` | POST | Delete single message by `id` |
| `/api/chat/clear` | GET | Clear all history |
| `/api/chat/clear-old` | POST | Delete msgs older than N days |
| `/api/chat/search` | GET | Search messages (`?q=`) |
| `/api/chat/search/advanced` | POST | Advanced search (date/sender/text) |
| `/api/chat/stats` | GET | Stats (total, today, per_chat) |
| `/api/chat/last-message` | GET | Last message per chat |
| `/api/chat/broadcast` | POST | Send to all contacts |

### 2.3 Message Features
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/pin` | POST | Toggle pin message |
| `/api/chat/react` | POST | Toggle emoji reaction |
| `/api/chat/typing` | POST | Send typing indicator |
| `/api/chat/schedule` | POST | Schedule message delivery |
| `/api/chat/forward` | POST | Forward message to another contact |
| `/api/chat/batch-forward` | POST | Batch forward multiple messages |
| `/api/chat/drafts` | GET/POST | Draft message management |

### 2.4 Automation
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/auto-reply` | GET/POST/PUT/DELETE | Auto-reply rules (keywords + regex) |
| `/api/chat/groups` | GET/POST/PUT/DELETE | Contact groups |
| `/api/chat/labels` | GET/POST/DELETE | Message labels |
| `/api/chat/content-filter` | POST | Filter/block messages |

### 2.5 Data Management
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/backup` | GET/POST | Download/upload chat backup |
| `/api/chat/export` | GET | Export chat (`?format=json|html`) |
| `/api/chat/archive` | POST | Archive old messages to USB |
| `/api/chat/status` | GET | Bridge health + message count |
| `/api/chat/bridge-health` | GET | Latency, reconnect stats, uptime |
| `/api/chat/bridge-config` | GET | Bridge configuration inspection |

### 2.6 Templates
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/templates` | GET/POST/PUT/DELETE | Message templates (with category) |

---

## 3. Economy (Liquid Taler)

### 3.1 Silver-Backed Currency
The Liquid Taler (ng) is silver-backed digital currency.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/economy/wheel` | GET | Wheel of fortune mini-game |
| `/api/economy/auto-mint` | GET | Auto-mint banknotes |
| `/api/economy/crafting` | POST | Banknote crafting system |
| `/api/economy/reinvest` | POST | Auto-reinvest dividends |
| `/api/economy/onboarding` | GET | Economy onboarding flow |
| `/api/economy/oracle` | GET | Live silver spot price |
| `/api/economy/deflate` | POST | Deflation management |
| `/api/economy/tokenomics` | GET | Tokenomics dashboard |
| `/api/economy/dividend-admin` | GET/POST | Dividend history + manual trigger |
| `/api/economy/rates` | GET | Multi-currency rates (EUR, GBP, JPY, BTC, XAG...) |
| `/api/economy/invoice-webhook-test` | GET | Test invoice webhook |

### 3.2 Treasury & Reserve
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/treasury/state` | GET | Treasury state |
| `/api/treasury/proof-of-reserve` | GET | Proof of reserve summary |
| `/api/reserve/por` | GET | Legacy PoR endpoint |
| `/api/reserve/proof` | GET | Enhanced proof with backing ratio, audit trail |
| `/api/treasury/usdt-deposits` | GET | USDT deposit history |
| `/api/treasury/register-banknote` | POST | Register new banknote |
| `/api/treasury/claim-dividends` | POST | Claim banknote dividends |
| `/api/treasury/init-silver-round` | POST | Initiate silver round |
| `/api/treasury/auto-round` | GET | Auto silver round |

### 3.3 Silver Assets
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/silver/mint` | POST | Mint silver-backed asset |
| `/api/silver/burn` | POST | Burn (redeem) asset |
| `/api/silver/list` | GET | List all silver assets |

### 3.4 RWA (Real-World Assets)
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/rwa/register` | POST | Register tokenized real-world asset |
| `/api/rwa/list` | GET | List all registered RWAs |

### 3.5 Invoices
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/invoice/create` | POST | Create invoice (amount + description) |
| `/api/chat/invoice/list` | GET | List invoices |
| `/api/chat/invoice/pay` | POST | Pay invoice by id |
| `/api/chat/invoice/stats` | GET | Stats (total/pending/paid) |
| `/api/chat/invoice/export-csv` | GET | Export invoices as CSV |

---

## 4. AI & Steward

### 4.1 AI Steward (Ollama-based)
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/steward` | POST | Talk to AI Steward |
| `/api/ai/constitution` | GET | Constitutional analysis |
| `/api/ai/monitor` | GET | Steward monitor metrics |
| `/api/ai/steward-did` | GET | Steward DID document |
| `/api/ai/radio-content` | GET | AI-generated radio content |

### 4.2 Moderation
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/admin/moderation-stats` | GET | Moderation usage statistics |

---

## 5. Webhooks & Integrations

### 5.1 Webhook
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/webhook` | GET/POST/PUT/DELETE | Webhook config (retry, retry_delay) |
| `/api/admin/webhook-queue` | GET | Persistent webhook queue status |
| `/api/webhook/whatsapp` | POST | WhatsApp integration |
| `/api/webhook/signal` | POST | Signal integration |
| `/api/webhook/matrix` | POST | Matrix integration |
| `/api/webhook/discord` | POST | Discord integration |

---

## 6. SimpleX Channels (v6.5)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/simplex/channel/create` | POST | Create channel |
| `/api/simplex/channel/list` | GET | List channels |
| `/api/simplex/channel/join` | POST | Join channel by URI |
| `/api/simplex/channel/post` | POST | Post to channel |

---

## 7. DID (Decentralized Identity)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/did` | GET | Node DID document |
| `/api/did/contact` | GET | Contact DID document |

---

## 8. Inter-Node Relay (P2P)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/relay/send` | POST | Send message to remote node |
| `/api/relay/receive` | GET | Receive messages from remote |
| `/api/relay/history` | GET | Relay message history |

---

## 9. Radio

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/radio` | GET | Web radio player |
| `/api/radio/ai-content` | GET | AI-generated radio scripts |
| Radio streaming | GET | M3U8 playlist, track streaming |

---

## 10. Node Info & Health

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/version` | GET | Build version |
| `/api/status` | GET | Server status |
| `/api/health` | GET | Health (uptime_hours, bridge, healthy, msg_count) |
| `/api/admin/info` | GET | Comprehensive node introspection (all subsystems) |
| `/api/admin/docker` | GET | Docker container health |
| `/api/admin/backup` | POST | Trigger backup |
| `/api/admin/metrics/system` | GET | System metrics (CPU, RAM, disk) |
| `/api/tracker/nodes` | GET | P2P tracker nodes |
| `/api/disk-check` | GET | Disk space check |

---

## 11. Database Administration

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/db/list` | GET | List SQLite database files |
| `/api/db/backup` | POST | Backup database to USB |
| `/api/db/backup/list` | GET | List database backups |
| `/api/db/restore` | POST | Restore database from backup |
| `/api/db/upload` | POST | Upload and restore database file |

---

## 12. Crypto & Security

### 12.1 Wallet
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/addresses` | GET | List wallet addresses |
| `/api/rotate` | GET | Rotate wallet keys |

### 12.2 CryptoContainer
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/container/*` | GET/POST/PUT/DELETE | CryptoContainer management |
| `/api/panic` | POST | PANIC wipe (emergency) |

### 12.3 Vault (16GB Encrypted)
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/vault/list` | GET | List vault contents |
| `/api/vault/upload` | POST | Upload to vault |
| `/api/vault/download` | GET | Download from vault |
| `/api/vault/delete` | DELETE | Delete from vault |
| `/api/vault/save-note` | POST | Save encrypted note |

### 12.4 Lock
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/lock-status` | GET | Lock status |
| `/api/lock` | POST | Lock node |
| `/api/unlock` | POST | Unlock node |
| `/api/change-lock-code` | POST | Change lock code |

---

## 13. Diagnostics & Audit

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/admin/audit-log` | GET | Security audit log |
| `/api/admin/metrics` | GET | Performance metrics |
| `/api/admin/diagnostics` | GET | Full diagnostics (incl. goroutine dump) |
| `/api/admin/status-page` | GET | HTML status page |
| `/api/admin/search-index` | GET | Search index status |
| `/api/admin/rate-limit-status` | GET | Rate limiter status |
| `/api/admin/rate-limit-config` | GET/POST | Configure rate limits |
| `/api/chat/analytics` | GET | Chat analytics |
| `/api/inquisitor/report` | GET | Consolidated status report |

---

## 14. Telegram Bots

The node runs 3 Telegram bots (embedded, no systemd):
- **AskSteward** — `/steward` command → Ollama AI
- **DarkPushkin** — Economy & administrative commands
- **Torquemada** — Inquisitor, receives automated reports

All accessed via the gateway webhook endpoints.

---

## 15. Economy Tiers & Governance

### Banknotes (5 rarities)
| Tier | Rarity |
|------|--------|
| Common | Standard issue |
| Uncommon | Limited print |
| Rare | Scarce |
| Epic | Very limited |
| Genesis | Founders' edition |

### DAO
- Governance voting
- Mint authorization
- Service registry
- Settlements & royalties

---

## 16. ParanoidX Fork

Located at `/home/tomas/simplex-fork` (separate repository).

### Components
| Component | Path | LOC | Description |
|-----------|------|-----|-------------|
| AppTransport | `internal/apptransport/` | 1,077 | Envelope, queue, replay, signal, types |
| Bridge | `internal/bridge/` | 616 | WS bridge + SOCKS5 proxy |
| Paranoid | `internal/paranoid/` | 191 | VPN/TOR modes, embedded transport |
| ParanoidX | `internal/paranoidx/` | 51 | Architecture definitions |
| SimpleX CLI | `internal/simplexcli/` | 344 | CLI client + tests |
| SMP | `internal/smp/` | 186 | SMP protocol client |
| Transport | `internal/transport/` | 293 | Transport layer |
| VPN | `internal/vpn/` | 318 | VPN service |
| Probe | `cmd/paranoidx-probe/` | 102 | Node probe tool |
| Test | `cmd/paranoidx-test/` | 199 | Testing harness |
| Main fork | `cmd/simplex-fork/` | 316 | Entry point |

**Total:** 20 Go files, ~3,693 LOC

### VPN/TOR Modes (4)
1. **Full Tor** — All traffic via Tor
2. **Split Tunnel** — Chat via Tor, rest clearnet
3. **VPN Only** — Standard VPN without Tor
4. **Stealth** — Randomized routing for obfuscation

---

## 17. Deployment

### Launch Node
```bash
bash /home/tomas/simplex-node/scripts/launch-node.sh
```

### Post-Reboot Restore
```bash
bash /home/tomas/simplex-node/scripts/post-reboot-restore.sh
```

### Backup to USB
```bash
bash /home/tomas/simplex-node/scripts/backup-to-usb.sh
```

### Build Go
```bash
cd /home/tomas/simplex-node
VERSION="b116"
CGO_ENABLED=0 go build -ldflags="-X main.buildVersion=$VERSION" -o simplex-node ./cmd/simplex-node/
cp simplex-node /home/tomas/bin/simplex-node
```

### Build Flutter
```bash
cd apps/isle_app
flutter build linux
cp build/linux/x64/release/bundle/lib/libapp.so /home/tomas/.local/bin/the-isle/lib/
cp build/linux/x64/release/bundle/isle_app /home/tomas/.local/bin/the-isle/
```

---

## 18. API Usage Examples

### Send a message
```bash
curl -X POST http://localhost:8080/api/chat/send \
  -H "Content-Type: application/json" \
  -d '{"chat_id":"@123456","text":"Hello from API"}'
```

### Get chat history
```bash
curl "http://localhost:8080/api/chat/history?chat_id=@123456"
```

### Get silver price
```bash
curl http://localhost:8080/api/economy/oracle
```

### Check bridge health
```bash
curl http://localhost:8080/api/chat/bridge-health
```

### Trigger backup
```bash
curl -X POST http://localhost:8080/api/admin/backup
```

---

## 19. Data Locations

| Data | Path |
|------|------|
| Chat history | `~/.local/share/simplex-node/chat_history.json` |
| Invoices | `~/.local/share/simplex-node/invoices.json` |
| Economy state | SQLite in data dir |
| Vault | Encrypted 16GB SQLite |
| CryptoContainer | AES-256-GCM container |
| Radio tracks | `~/.local/share/simplex-node/radio/` |
| Docker stack | `~/simplex-node/docker/` |
| USB backup | `/mnt/simplex-backup/` |
| Binary | `~/bin/simplex-node` |
| Flutter app | `~/.local/bin/the-isle/` |
| ParanoidX fork | `/home/tomas/simplex-fork/` |

---

## 20. Quick Reference

| Context | Command |
|---------|---------|
| Server URL | `http://localhost:8080` |
| Health check | `curl http://localhost:8080/api/health` |
| Version | `curl http://localhost:8080/api/version` |
| Node info | `curl http://localhost:8080/api/admin/info` |
| Tor dashboard | Onion URL from `/api/dashboard-onion` |
| Inquisitor report | `curl http://localhost:8080/api/inquisitor/report` |
| Stop server | `pkill simplex-node` |
| Flutter | `bash scripts/run-isle-app.sh` |
