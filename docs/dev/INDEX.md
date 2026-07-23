# Developer Documentation

## Project Structure

```
simplex-node/
├── cmd/
│   ├── simplex-node/main.go      # HTTP-сервер + маршрутизация (1964 строки)
│   ├── banknote-press/main.go    # Генерация PDF банкнот (250 строк)
│   └── simplex-bridge/main.go    # SimpleX CLI bridge (20 строк)
├── internal/
│   ├── ai/         # Ollama AI интеграция (512 строк)
│   ├── api/        # REST-обработчики (4067 строк, 19 файлов)
│   ├── billing/    # Биллинг (188 строк)
│   ├── bot/        # Telegram-боты (1877 строк, 4 бота)
│   ├── bridge/     # SimpleX CLI bridge (423 строки)
│   ├── channels/   # (пусто)
│   ├── config/     # Конфигурация (298 строк)
│   ├── dockerutil/ # Docker utilities (90 строк)
│   ├── economy/    # Экономический движок (9441 строка, 37 файлов)
│   ├── fileutil/   # Атомарные файловые операции (210 строк)
│   ├── gateway/    # Мульти-канальный шлюз (291 строка)
│   ├── health/     # Мониторинг здоровья (311 строк)
│   ├── lock/       # Блокировка дашборда (305 строк)
│   ├── middleware/ # HTTP middleware (249 строк)
│   ├── press/      # Генерация PDF (470 строк)
│   ├── radio/      # (пусто)
│   ├── royal/      # Royal Sub-протокол (336 строк)
│   ├── status/     # Статус узла (411 строк)
│   ├── steward/    # AI-стюард (610 строк)
│   ├── ton/        # ARGENTUM Jetton (80 строк)
│   ├── treasury/   # Silver rounds (603 строки)
│   ├── vault/      # E2EE хранилище (280 строк)
│   └── webrtc/     # WebRTC сигналинг (112 строк)
├── docker/         # Docker compose + Tor + конфиги
├── docs/           # Документация
├── scripts/        # Скрипты развёртывания и CI
├── apps/           # Flutter-приложения
├── testdata/       # Тестовые данные
├── systemd/        # systemd-юниты
└── docker/         # Docker-инфраструктура
```

## API Routes (122 total)

### Status & Health
- `GET /api/status` — полный статус узла
- `GET /api/health` — отчёт о здоровье
- `GET /api/addresses` — onion-адреса
- `GET /api/disk-check` — использование диска
- `GET /api/metrics` — Prometheus метрики

### Economy
- `GET /api/economy/state` — состояние экономики
- `GET /api/economy/wheel` — колесо фортуны
- `GET /api/economy/auto-mint` — авто-майнинг
- `GET /api/economy/crafting` — крафтинг
- `GET /api/economy/reinvest` — реинвестирование
- `GET /api/economy/holdings` — балансы
- `GET /api/economy/pre-mint` — пред-майнинг
- `GET /api/economy/onboarding` — онбординг (+ welcome, starter, guide)

### Treasury
- `GET /api/treasury/state` — состояние казначейства
- `GET /api/treasury/proof-of-reserve` — Proof of Reserve
- `GET /api/treasury/register-banknote` — регистрация банкноты
- `GET /api/treasury/claim-dividends` — получение дивидендов
- `GET /api/treasury/simulate-deposit` — симуляция USDT-депозита
- `GET /api/treasury/init-silver-round` — инициализация серебряного раунда
- `GET /api/treasury/usdt-deposits` — USDT депозиты
- `GET /api/treasury/auto-round` — авто-раунд

### POS
- `GET /api/pos?action=create` — создание инвойса
- `GET /api/pos?action=pay` — оплата инвойса
- `GET /api/pos?action=stats` — статистика
- `GET /api/pos?action=merchant-revenue` — выручка мерчанта
- `GET /api/pos?action=voucher-create` — создание ваучера
- `GET /api/pos?action=voucher-redeem` — погашение ваучера
- `GET /api/pos?action=voucher-list` — список ваучеров
- `GET /api/pos/qr?id=X` — QR-код инвойса

### Docs
- `GET /api/docs/list` — список документов
- `GET /api/docs/download?name=X` — скачивание
- `GET /api/docs/view?name=X` — просмотр

### Vault
- `GET /api/vault/list` — список файлов
- `GET /api/vault/upload` — загрузка
- `GET /api/vault/download?name=X` — скачивание
- `GET /api/vault/delete` — удаление
- `GET /api/vault/save-note` — сохранение заметки

### Aукцион
- `GET /api/auction/list` — список аукционов
- `GET /api/auction/active` — активные аукционы
- `GET /api/auction/bid` — ставка
- `GET /api/auction/my` — мои ставки

### Royal Sub
- `POST /api/royal/relay` — relay-сообщение
- `GET /api/royal/sync` — синхронизация

### Lock
- `GET /api/lock-status` — статус блокировки
- `POST /api/lock` — заблокировать
- `POST /api/unlock` — разблокировать
- `POST /api/change-lock-code` — сменить код

### Other
- `GET /api/ice-config` — ICE/TURN конфиг
- `POST /api/call-signal` — WebRTC сигналинг
- `GET /api/subscription` — подписки
- `GET /api/rotate` — ротация onion-адресов
- `GET /api/p2p/explore` — P2P explore
- `GET /api/genesis/info` — Genesis ICO информация
- `POST /api/pack/buy` — покупка пака
- `POST /api/pack/open` — открытие пака
- `GET /api/pack/list` — список паков
- `GET /api/buyback/quote` — котировка buyback
- `POST /api/buyback/sell` — продажа через buyback

## Экономический движок

### Ключевые константы (tokenomics.go)
- `TreasuryCommissionBPS = 228` (2.28%)
- `MaxTotalFeeBPS = 420` (4.20%)
- `NGPerTLR = 31_103_480_000` (1 Troy oz в нанограммах)
- `SilverBackingRatio = 0.70` (70% обеспечение)
- `UtilityPremiumPct = 0.30` (30% утилитарная премия)

### DynamicParams (params.go)
9 параметров, управляемых Steward'ом:
- treasury_commission_bps, max_total_fee_bps, auction_listing_fee_bps
- auction_buyer_premium, auction_seller_fee_bps, utility_premium_pct
- silver_backing_ratio, pos_fee_bps, monthly_ops_ng

### Подписки
- Citizen: 2 000 000 000 ng/мес (~$4.82)
- Aristocrat: 20 000 000 000 ng/мес (~$48.20)
- Colonist: бесплатно (2.28% комиссия P2P)

## Telegram-боты

### AskSteward (@AskSteward_bot)
- AI-ассистент с интеграцией Ollama
- Меню документации (уайтпейперы, THEPLAN, эволюция)
- Отправка документов и фото через Bot API
- Токен: 8885061690:AAEkJ6Y...

### Torquemada (@torquemada878_bot)
- Мониторинг узла с live-данными
- Кнопки: статус, диск, здоровье, адреса, экономика, параметры, стюард, POS
- NodeAPI клиент для /api/status, /api/economy/state, /api/disk-check и т.д.
- Токен: 8825368561:AAF3HAM...

### DarkPushkin
- Push-уведомления
- Токен: 8471637894:AAFXJb_...

### Inquisitor
- Приём отчётов от CI/CD
- Поддержка sendMessage, sendDocument, sendPhoto
- Токен: 8933708843:AAGiAD...

## Docker-инфраструктура

4 контейнера (docker/docker-compose.yml):
- `smp-server` — SimpleX SMP message relay (simplexchat/smp-server)
- `xftp-server` — SimpleX XFTP file relay (simplexchat/xftp-server)
- `tor` — Alpine Tor с постоянными .onion ключами (кастомный Dockerfile)
- `coturn` — TURN сервер для WebRTC (coturn/coturn)

Hidden services (torrc):
- Dashboard → host.docker.internal:8080
- SMP → smp-server:5223
- XFTP → xftp-server:443
- ICE/TURN → coturn:3478, coturn:5349
- Auditor → host.docker.internal:8080

## Полные имена файлов (full file listing)

Source files with their full paths, sorted by size:

### cmd/
- `cmd/simplex-node/main.go` — 1964 строки, HTTP-сервер и маршрутизация
- `cmd/banknote-press/main.go` — 250 строк, генерация PDF
- `cmd/simplex-bridge/main.go` — 20 строк, CLI bridge

### internal/
- `internal/economy/registry.go` — 614 строк, реестр RWA/банкнот
- `internal/economy/franchise.go` — 602 строки, франчайзинг
- `internal/api/treasury.go` — 1089 строк, treasury handlers
- `internal/api/market.go` — 457 строк, market handlers
- `internal/api/docs.go` — 136 строк, документация
- `internal/bot/steward_bot.go` — 780 строк, AskSteward
- `internal/bot/torquemada_bot.go` — 510 строк, Torquemada
- `internal/bot/nodeapi.go` — 264 строки, NodeAPI клиент
- `internal/steward/steward.go` — 209 строк, Steward service
- `internal/treasury/round.go` — 266 строк, silver rounds
- `internal/gateway/gateway.go` — 82 строки, gateway interface

### Тесты
- 39 test files, 418 test functions, 5995 строк тестов
- 18/18 Go packages test green
- `internal/webrtc/` — 14 тестов (SignalState + ICEConfig)

## Сборка и деплой

```bash
go build ./cmd/simplex-node/
go vet ./...
go test ./... -short -count=1 -timeout 30s
```

Бинарник: 12MB статический ELF (без CGO, без external dependencies)
