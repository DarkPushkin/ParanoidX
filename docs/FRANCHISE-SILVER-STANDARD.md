# ПРОМПТ:
# "Теперь нам надо изобрести дополнительные шаги например франшиза серебряного стандарта для создания центробанков других государств с национальным монетным двором, с небольшим роялти в пользу Банка Острова в качестве оплаты за создание и поддержание инфраструктуры нод, техподдержку, развитие."

# =============================================================================
# FRANCHISE SILVER STANDARD — ПОЛНАЯ АРХИТЕКТУРА
# =============================================================================

## 1. КОНТЕКСТ: ЧТО УЖЕ ЕСТЬ В КОДЕ

В `cmd/simplex-node/main.go` (строки 2181–2220) уже заложен скелет иерархии:

  - `royal.enabled` — маркерный файл: если есть, узел — королевский
  - `POST /royal/control` — приём команд Короля (сейчас заглушка)
  - `POST /royal/sync` — синхронизация sub-node (заглушка)
  - TODO в строке 2185: «Phase3: real sign (ed25519/hmac), SMP integration for cmds, multi-sub map»
  - TODO в строке 2186: «extract economy handlers to internal/economy»

Уже существует:
  - `internal/economy/ledger.go` — Ledger с Ed25519-аккаунтами, переводами, mint
  - `internal/economy/registry.go` — BanknoteV2, pre-mint, аудиторы, долг первого инвестора, treasury split
  - `cmd/banknote-press/main.go` — отдельный бинарник монетного двора (localhost:9192)
  - `internal/press/templates.go` — шаблоны банкнот, рендер PDF, Ed25519-подпись
  - Понятие Royal Node vs Sub-node уже введено, но не реализовано

Автор уже думал в эту сторону. Задача — формализовать и реализовать.

---

## 2. ЧЕТЫРЁХУРОВНЕВАЯ АРХИТЕКТУРА

  TIER 0:  ROYAL NODE — БАНК ОСТРОВА (Saint Mary Liberty Island)
  │
  ├── Master Silver Vault — физическое серебро в Island хранилище
  ├── Master Ledger — definitive ledger, единственный источник истины
  ├── Settlement Engine — межфранчайзинговые расчёты
  ├── Franchise Manager — CRUD лицензий, мониторинг
  ├── Root Mint — эмиссия Genesis Banknotes (только Остров)
  ├── Hierarchical Proof of Reserve — публичный API
  │
  ├── TIER 1:  ЦЕНТРОБАНКИ-ФРАНЧАЙЗИ (Franchise Nodes)
  │   │
  │   ├── Franchise «Republic of Liberia»
  │   │   ├── Sub-node (simplex-node fork, branded)
  │   │   ├── National Mint (banknote-press с нац. шаблонами)
  │   │   ├── Local Ledger (копия для быстрых операций)
  │   │   ├── National Silver Account (earmarked на Island)
  │   │   └── National Banknotes (серия LR-SILVER-*)
  │   │
  │   ├── Franchise «Commonwealth of ...»
  │   └── Franchise «Principality of ...»
  │
  ├── TIER 2:  КОММЕРЧЕСКИЕ БАНКИ / МЕРЧАНТЫ
  │   ├── Merchant nodes (упрощённый simplex-node)
  │   ├── POS-терминалы (QR-оплата)
  │   ├── Платёжные шлюзы
  │   └── Обменники (fiat ↔ silver)
  │
  └── TIER 3:  ПОЛЬЗОВАТЕЛИ
      ├── Mobile wallets (Ed25519)
      ├── Physical banknote holders
      └── SimpleX Chat users

---

## 3. СТРУКТУРА ФРАНЧАЙЗ-ЛИЦЕНЗИИ

Файл `franchise_licenses.json` на Royal Node:

```json
{
  "version": 1,
  "licenses": [
    {
      "franchise_id": "LIBERIA-001",
      "nation": "Republic of Liberia",
      "nation_code": "LR",
      "currency_name": "Liberian Silver Dollar",
      "currency_code": "LSD",
      "tier": "full_reserve",
      "reserve_ratio": 1.0,
      "silver_balance_ng": 31103480000000000,
      "national_mint_pubkey": "abc123...",
      "node_pubkey": "def456...",
      "settlement_pubkey": "ghi789...",
      "royalty_rate_bp": 50,
      "mint_cap_daily_ng": 15551740000000000,
      "mint_authority": true,
      "serial_prefix": "LR",
      "logo_url": "https://...",
      "activated_at": "2026-06-01T00:00:00Z",
      "expires_at": "2027-06-01T00:00:00Z",
      "status": "active"
    }
  ]
}
```

Поля лицензии:
  - `tier`: "full_reserve" (1:1) или "fractional" (напр. 0.5 — 50% резерва)
  - `reserve_ratio`: доля серебра, которую franchise обязан держать в Island Vault
  - `silver_balance_ng`: earmarked серебро на балансе franchise
  - `royalty_rate_bp`: базисные пункты роялти (50 = 0.5%)
  - `mint_cap_daily_ng`: дневной лимит эмиссии
  - `serial_prefix`: префикс для банкнот этого franchise
  - Три ключа: mint (для подписи банкнот), node (для API-аутентификации), settlement (для расчётов)

---

## 4. EARMARKED ACCOUNTS (целевые счета на Master Ledger)

В `internal/economy/ledger.go` добавляется новый тип счёта:

```go
type Account struct {
    BalanceNg   int64  `json:"balance_ng"`    // свободный баланс
    EarmarkedNg int64  `json:"earmarked_ng"`   // зарезервировано под franchise
    CreatedAt   string `json:"created_at"`
    LastTx      string `json:"last_tx"`
}

type EarmarkEntry struct {
    FranchiseID string `json:"franchise_id"`
    OwnerNg     int64  `json:"owner_ng"`     // юридически принадлежит franchise
    BackingNg   int64  `json:"backing_ng"`   // физически в Island vault
    UpdatedAt   string `json:"updated_at"`
}
```

Правило: `Physical Silver in Vault >= sum(EarmarkEntry.BackingNg) + RoyalNode.OwnBalance`.

Это гарантирует, что franchise не может эмитировать больше, чем обеспечено серебром.

---

## 5. ПРОЦЕСС ЭМИССИИ (National Mint)

### 5.1 Запрос на эмиссию

Franchise Node → POST `/api/franchise/mint` → Royal Node:

```json
{
  "franchise_id": "LIBERIA-001",
  "denomination_ng": 3110348000000,
  "rarity": "rare",
  "count": 100,
  "template_id": "liberia-silver-dollar-v1",
  "timestamp": "...",
  "signature": "<Ed25519(node_privkey, payload)>"
}
```

### 5.2 Валидация на Royal Node

1. Проверить Ed25519 подпись по `node_pubkey` из лицензии
2. Проверить статус лицензии (active, не expired)
3. Проверить резерв: `count * denomination_ng <= franchise.silver_balance_ng`
4. Проверить daily cap
5. Начислить royalty: `minting_fee_ng = count * denomination_ng * royalty_rate_bp / 10000`
6. Дебетнуть `franchise.silver_balance_ng` на сумму эмиссии
7. Кредитнуть royalty на счёт Острова
8. Вернуть `mint_authorization_token` (Ed25519-подписанный Королём)

### 5.3 Выпуск банкнот (локально на Franchise Node)

POST `/api/mint/issue` (локальный, franchise node):

1. Получить `mint_authorization_token` от Royal
2. Banknote-press рендерит PDF с национальным шаблоном
3. Двойная подпись:

```
Banknote Signature = {
  "royal_sig":  <Ed25519(royal_privkey, banknote_hash)>,
  "issuer_sig": <Ed25519(franchise_mint_privkey, banknote_hash)>
}
```

4. Регистрация в локальном банкнотном реестре franchise
5. Серийный номер: `LR-SILVER-RARE-2026-000001`

### 5.4 Структура банкноты franchise

```go
type FranchiseBanknote struct {
    Serial          string  `json:"serial"`           // LR-SILVER-RARE-2026-000001
    DenominationNg  int64   `json:"denomination_ng"`
    Rarity          string  `json:"rarity"`
    FranchiseID     string  `json:"franchise_id"`
    TemplateID      string  `json:"template_id"`
    RoyalSig        string  `json:"royal_sig"`        // подпись Короля
    IssuerSig       string  `json:"issuer_sig"`       // подпись центробанка
    Holder          string  `json:"holder"`
    Status          string  `json:"status"`            // active, burned, escrow
    MintedAt        string  `json:"minted_at"`
    PdfHash         string  `json:"pdf_hash,omitempty"`
}
```

---

## 6. МЕЖФРАНЧАЙЗИНГОВЫЕ РАСЧЁТЫ (Settlement Layer)

### 6.1 Прямой перевод между franchise

```
Пользователь A (Liberia) хочет отправить 10 LSD пользователю B (Micronesia):

1. Local Ledger Либерии: -10 LSD (hold)
2. Settlement Request: Royal Node /api/settlement/transfer
   {
     "from_franchise": "LIBERIA-001",
     "to_franchise": "MICRONESIA-001",
     "amount_ng": 311034800000,
     "from_account": "pubkey_A",
     "to_account": "pubkey_B",
     "signature_from": "<A's Ed25519>"
   }
3. Royal проверяет:
   - Баланс franchise LIBERIA достаточно?
   - Подпись A валидна?
4. Royal списывает с earmark LIBERIA, зачисляет на earmark MICRONESIA
5. С каждой такой транзакции: 0.5% (50bp) royalty Острову
6. Local Ledger Либерии: -10 LSD (confirmed)
7. Local Ledger Микронезии: +10 LSD (в эквиваленте)
8. Оба получают settlement receipt (подписанный Королём)
```

### 6.2 Курсовая сетка

Все franchise валюты привязаны к одному silver standard (1 unit = X ng серебра).
Поэтому курс между любыми двумя franchise всегда 1:1 по silver weight.

Но franchise может установить:
  - `tx_fee` — свою комиссию за входящие/исходящие переводы
  - `spread` — спред при обмене на фиат

Курс franchise → TLR: всегда `1 franchise_unit = denomination_ng TLR`.

### 6.3 Settlement Periods

  - Вариант A: **Real-time gross settlement** (каждая транзакция сразу)
  - Вариант B: **Batch settlement** (раз в N часов, как банки)
  - Вариант C: **Lazy settlement** (только при превышении лимита)

Для малых franchise — batch (меньше трафика на Royal). Для крупных — RTGS.

---

## 7. ИЕРАРХИЧЕСКИЙ PROOF OF RESERVE (PoR)

### 7.1 Структура Merkle Tree

```
Level 2 (Root): Island Root Hash
│
├── Level 1: Hash(LIBERIA PoR || MICRONESIA PoR || ...)
│   │
│   ├── Level 0: Hash(LR-banknote-001 || LR-banknote-002 || ...)
│   │
│   ├── Level 0: Hash(MC-banknote-001 || ...)
│   │
│   └── Level 0: Island Self PoR (Genesis banknotes + silver reserve)
│
└── Physical Silver In Vault (signed by independent auditor)
```

### 7.2 Аттестация

Каждый franchise обязан:
  1. Раз в сутки публиковать Merkle root своих банкнот
  2. Подписывать своей `settlement_pubkey`
  3. Royal Node проверяет, что `sum(denomination_ng) <= silver_balance_ng`
  4. Royal публикует сводный PoR на `/api/reserve/por`

Любой может проверить:
  - Взять любую банкноту → проверить её в Merkle tree franchise
  - Взять Merkle root franchise → проверить в Island Root
  - Сравнить Island Root с физическим серебром в vault (аудитор)

### 7.3 Открытый API

```go
GET /api/reserve/por
{
  "root_hash": "abc...",
  "timestamp": "...",
  "total_silver_ng": 311034800000000000,
  "total_physical_oz": 10000000,
  "total_usd_value": 716200000,
  "franchises": [
    {
      "franchise_id": "LIBERIA-001",
      "merkle_root": "def...",
      "total_issued_ng": 15551740000000000,
      "silver_balance_ng": 31103480000000000,
      "reserve_ratio": 2.0,
      "last_attested": "...",
      "signature": "<franchise sig>"
    }
  ],
  "royal_signature": "<royal sig>"
}
```

---

## 8. REVENUE MODEL (РОЯЛТИ БАНКУ ОСТРОВА)

### 8.1 Таблица роялти

| Поток | Ставка | Механизм |
|-------|--------|----------|
| **Setup Fee** | 10 kg Ag (фикс) | Единоразово при подписании |
| **Infrastructure** | 100 ng / месяц | Автоматически с earmark |
| **Minting Fee** | 1.0% от номинала | При каждой эмиссии банкнот |
| **Settlement Fee** | 0.5% от суммы | При кросс-франчайз переводе |
| **Storage Fee** | 0.25% годовых на earmark | Ежемесячно |
| **Audit Fee** | 1000 ng / год | Ежегодно |
| **Compliance Fee** | 500 ng / квартал | При наличии KYC/AML |

### 8.2 Пример расчёта для одного franchise

```
Эмиссия: 10,000 банкнот × 10 TLR = 100,000 TLR = 100,000 oz Ag
По споту $71.62/oz = $7,162,000

Minting Fee (1%):         1,000 oz Ag  = $71,620
Settlement (год ~5% vol):   500 oz Ag  = $35,810
Storage (0.25%):            250 oz Ag  = $17,905
Infrastructure:               1 oz Ag  = $71.62
Audit:                        0.03 oz  = $2.30

Annual Total ≈ 1,751 oz Ag ≈ $125,406 с одного franchise
```

При 50 franchise: **~$6.27M/год** пассивного дохода Острову
При 500 franchise: **~$62.7M/год**

### 8.3 Автоматический сбор роялти

```go
type RoyaltyEngine struct {
    mu sync.Mutex
}

func (re *RoyaltyEngine) ChargeMintingFee(franchiseID string, amountNg int64) error {
    // 1. Рассчитать fee = amountNg * license.RoyaltyRateBP / 10000
    // 2. Списать с earmark счёта franchise
    // 3. Зачислить на Royal Treasury Account
    // 4. Записать в royalty_ledger.json
}

func (re *RoyaltyEngine) ChargeSettlementFee(fromFranchise, toFranchise string, amountNg int64) {
    // Аналогично
}
```

---

## 9. НАЦИОНАЛЬНЫЕ ШАБЛОНЫ (Template System)

### 9.1 Что меняется в `internal/press/templates.go`

```go
type NationTemplate struct {
    ID           string `json:"id"`
    FranchiseID  string `json:"franchise_id"`
    Name         string `json:"name"`
    Language     string `json:"language"`
    DesignSVG    string `json:"design_svg"`     // или путь к файлу
    Watermark    string `json:"watermark"`       // SVG водяного знака
    SerialFont   string `json:"serial_font"`
    // Поля для размещения:
    // - герб страны
    // - портрет национального лидера
    // - номинал прописью на нац. языке
    // - текст обязательства центробанка
}

type MultiTemplateManager struct {
    Templates map[string]*NationTemplate  // template_id -> template
    KingKey   ed25519.PublicKey
    mu        sync.RWMutex
}
```

### 9.2 Жизненный цикл шаблона

1. Franchise предоставляет SVG-дизайн с размеченными полями
2. Royal Node утверждает (anti-counterfeit review)
3. Шаблон регистрируется в `MultiTemplateManager`
4. При эмиссии franchise указывает `template_id`
5. Mint рендерит банкноту по шаблону + накладывает double signature

---

## 10. SECURITY MODEL

```
Franchise запрос
    → Ed25519(franchise_node_privkey, payload)
    → Royal Node проверяет node_pubkey из лицензии
    → Проверяет статус лицензии и лимиты
    → Исполняет операцию
    → Подписывает ответ Ed25519(royal_privkey, receipt)
    → Franchise верифицирует подпись Короля
```

Каждый franchise имеет три ключевых пары:
  - **Node Keypair** — аутентификация API-запросов к Royal
  - **Mint Keypair** — подпись банкнот (issuer_sig)
  - **Settlement Keypair** — подтверждение межбанковских транзакций

Ключи Короля:
  - **Royal Signing Key** — подпись банкнот (royal_sig), подпись PoR
  - **Royal Admin Key** — управление лицензиями (может быть отдельный cold key)

---

## 11. ЧТО НУЖНО НАПИСАТЬ (CODE PLAN)

### 11.1 Новые пакеты

```
internal/
├── franchise/
│   ├── license.go         — License struct, CRUD, валидация
│   ├── settlement.go      — межфранчайзинговые переводы
│   ├── mint_authority.go  — авторизация эмиссии, mint_auth_token
│   ├── por.go             — иерархический Proof of Reserve
│   └── royalty.go         — сбор роялти, royalty_ledger
├── press/
│   └── templates.go       — (расширить) MultiTemplateManager
└── economy/
    └── ledger.go          — (расширить) EarmarkEntry, earmarked accounts
```

### 11.2 Новые API endpoints (Royal Node)

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | `/api/franchise/register` | Создать лицензию |
| POST | `/api/franchise/revoke` | Отозвать лицензию |
| GET | `/api/franchise/list` | Список лицензий |
| GET | `/api/franchise/{id}/state` | Состояние franchise |
| POST | `/api/franchise/mint` | Разрешить эмиссию → mint_token |
| POST | `/api/settlement/transfer` | Кросс-франчайз перевод |
| GET | `/api/reserve/por` | Hierarchical PoR (публичный) |
| GET | `/api/royalty/balance` | Баланс роялти Острова |

### 11.3 Новые API endpoints (Franchise Node)

| Method | Endpoint | Описание |
|--------|----------|----------|
| POST | `/api/mint/render` | Рендер национальной банкноты |
| POST | `/api/mint/issue` | Выпустить банкноты (+ mint_token) |
| POST | `/api/sync/push` | Отправить snapshot Острову |
| POST | `/api/sync/pull` | Получить settlement |
| GET | `/api/mint/templates` | Список шаблонов franchise |

---

## 12. ДОРОЖНАЯ КАРТА РЕАЛИЗАЦИИ

```
Фаза 1 (2–3 недели): Ядро франчайзинга
├── internal/franchise/license.go — все структуры, CRUD, JSON persistence
├── internal/franchise/royalty.go — расчёт и сбор роялти
├── Расширение economy/ledger.go — earmarked accounts
├── API: register, revoke, list, state
├── API: franchise/mint (авторизация эмиссии)
├── Тесты: создание franchise, earmark, mint authorization

Фаза 2 (2 недели): Национальный Mint
├── press.MultiTemplateManager — поддержка множества шаблонов
├── API mint/render, mint/issue на franchise node
├── Двойная подпись (royal_sig + issuer_sig)
├── Национальные серийные префиксы
├── Тесты: рендер, issue, двойная подпись

Фаза 3 (2 недели): Settlement Layer
├── internal/franchise/settlement.go — кросс-франчайз переводы
├── API settlement/transfer
├── Settlement fee (роялти с транзакций)
├── API reserve/por
├── Hierarchical Merkle Tree PoR
├── Тесты: settlement между franchise, PoR верификация

Фаза 4 (1–2 недели): Franchise Node SDK
├── Шаблон franchise-node (fork simplex-node)
├── Docker Compose для franchise
├── setup.sh — скрипт развёртывания franchise
├── API sync/push, sync/pull
├── Документация: полный franchise manual
└── Демо: развернуть 2 franchise, провести транзакцию
```

---

## 13. КЛЮЧЕВОЕ КОНКУРЕНТНОЕ ПРЕИМУЩЕСТВО

На рынке сейчас:

| Продукт | Приватная связь | Физ. обеспечение | Нац. суверенитет | PoR | Готовая экономика |
|---------|----------------|------------------|------------------|-----|-------------------|
| Pax Gold | Нет | Gold | Нет | Да | Нет |
| Tether Gold | Нет | Gold | Нет | Нет | Нет |
| CBDC (цифровые валюты ЦБ) | Нет | Нет | Да (1 страна) | Нет | Нет |
| Monero | Да (только tx) | Нет | Нет | Нет | Нет |
| Status | Да | Нет | Нет | Нет | Частично |
| **simplex-node Franchise** | **Да** | **Silver** | **Да (любая страна)** | **Hierarchical** | **Полная** |

**Silver Standard Franchise** — единственное решение, объединяющее:
  - Физическое обеспечение (серебро в vault, не IOU)
  - Приватную децентрализованную связь (SimpleX)
  - Национальный суверенитет (каждый franchise — свой ЦБ со своим монетным двором)
  - Иерархический Proof of Reserve (полная прозрачность)
  - Готовую экономику (банкноты, аукционы, P2P, packs, реклама, терминалы)
  - Масштабируемость через franchise модель с роялти

В мире (июнь 2026), где серебро стоит $71.62/oz и растёт (+40% за год), физический дефицит нарастает, центробанки уходят от доллара, а приватность становится дефицитом — эта модель может стать инфраструктурой нового Bretton Woods, но на серебре, на приватных каналах и с национальным суверенитетом каждого участника.

---

*Документ создан: июнь 2026*
*Автор: AI-архитектор simplex-node*
