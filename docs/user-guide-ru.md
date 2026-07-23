# simplex-node / Isle / ParanoidX — Руководство пользователя (RU)

**Сборка:** b116 (Go), b81 (Flutter) | **Дата:** 2026-06-19

---

## 1. Обзор

simplex-node — узел суверенной цифровой сети, объединяющий:
- Приватный мессенджер на основе SimpleX (bridge к simplex-chat CLI)
- Цифровую экономику с серебряным обеспечением (Liquid Taler)
- Децентрализованное управление (DAO)
- Радио-трансляции
- AI-стюарда (Ollama)
- CryptoContainer (AES-256-GCM + argon2id)
- Форк ParanoidX (встроенный VPN/TOR, multi-network клиент)
- За Tor onion-сервисами

### Архитектура системы

```
simplex-node (Go :8080)
  ├── /api/* — REST API (100+ endpoints)
  ├── bridge — WebSocket ↔ simplex-chat CLI (:17225)
  ├── economy — Liquid Taler, банкноты, DAO, казначейство
  ├── radio — Планировщик, стриминг, M3U8
  ├── steward — AI-агент (Ollama)
  ├── container — CryptoContainer (AES-256-GCM + argon2id)
  └── docker stack — Tor, SMP, XFTP, Coturn
        └── 🧅 Tor onion-сервисы (5 скрытых сервисов)

Форк ParanoidX (/home/tomas/simplex-fork)
  ├── paranoid — Встроенный VPN/TOR (4 режима)
  ├── apptransport — Протокол AppTransport
  ├── bridge — SOCKS5 прокси + WS bridge
  ├── vpn — VPN-сервис
  └── simplexcli — SimpleX CLI клиент
```

---

## 2. Чат и Сообщения

### 2.1 Контакты
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/address/create` | GET | Создать локальный адрес чата |
| `/api/chat/contacts` | GET | Список всех контактов |
| `/api/chat/contact?id=@N` | GET | Информация о контакте |
| `/api/chat/contact/info` | GET | Инфо (количество + последнее сообщение) |
| `/api/chat/contact/alias` | POST | Переименовать контакт (`id` + `alias`) |
| `/api/chat/qr` | GET | QR-код контакта |
| `/api/chat/connect` | POST | Подключиться по приглашению |

### 2.2 Сообщения
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/history` | GET | История сообщений (опц. `chat_id`) |
| `/api/chat/stream` | GET | SSE стрим в реальном времени |
| `/api/chat/send` | POST | Отправить сообщение |
| `/api/chat/edit` | POST | Редактировать сообщение |
| `/api/chat/delete` | POST | Удалить сообщение |
| `/api/chat/clear` | GET | Очистить всю историю |
| `/api/chat/clear-old` | POST | Удалить сообщения старше N дней |
| `/api/chat/search` | GET | Поиск сообщений (`?q=`) |
| `/api/chat/search/advanced` | POST | Расширенный поиск (дата/отправитель/текст) |
| `/api/chat/stats` | GET | Статистика |
| `/api/chat/last-message` | GET | Последнее сообщение по чатам |
| `/api/chat/broadcast` | POST | Отправить всем контактам |

### 2.3 Функции сообщений
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/pin` | POST | Закрепить сообщение |
| `/api/chat/react` | POST | Добавить реакцию (emoji) |
| `/api/chat/typing` | POST | Индикатор набора текста |
| `/api/chat/schedule` | POST | Отложенная отправка |
| `/api/chat/forward` | POST | Переслать сообщение |
| `/api/chat/batch-forward` | POST | Массовая пересылка |
| `/api/chat/drafts` | GET/POST | Черновики сообщений |

### 2.4 Автоматизация
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/auto-reply` | GET/POST/PUT/DELETE | Правила автоответа (ключевые слова + regex) |
| `/api/chat/groups` | GET/POST/PUT/DELETE | Группы контактов |
| `/api/chat/labels` | GET/POST/DELETE | Метки сообщений |
| `/api/chat/content-filter` | POST | Фильтрация/блокировка сообщений |

### 2.5 Управление данными
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/backup` | GET/POST | Скачать/загрузить бэкап чата |
| `/api/chat/export` | GET | Экспорт чата (`?format=json|html`) |
| `/api/chat/archive` | POST | Архивировать старые сообщения на USB |
| `/api/chat/status` | GET | Статус bridge + количество сообщений |
| `/api/chat/bridge-health` | GET | Задержка, переподключения, uptime |
| `/api/chat/bridge-config` | GET | Конфигурация bridge |

### 2.6 Шаблоны
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/templates` | GET/POST/PUT/DELETE | Шаблоны сообщений (с категорией) |

---

## 3. Экономика (Liquid Taler)

### 3.1 Валюта с серебряным обеспечением
Liquid Taler (ng) — цифровая валюта, обеспеченная серебром.

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/economy/wheel` | GET | Колесо фортуны |
| `/api/economy/auto-mint` | GET | Автоматический минт банкнот |
| `/api/economy/crafting` | POST | Создание банкнот |
| `/api/economy/reinvest` | POST | Авто-реинвестирование дивидендов |
| `/api/economy/onboarding` | GET | Онбординг в экономику |
| `/api/economy/oracle` | GET | Текущая цена серебра |
| `/api/economy/deflate` | POST | Управление дефляцией |
| `/api/economy/tokenomics` | GET | Токеномика |
| `/api/economy/dividend-admin` | GET/POST | История дивидендов + ручной триггер |
| `/api/economy/rates` | GET | Курсы валют (EUR, GBP, JPY, BTC, XAG...) |
| `/api/economy/invoice-webhook-test` | GET | Тестовый вебхук инвойса |

### 3.2 Казначейство и Резервы
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/treasury/state` | GET | Состояние казначейства |
| `/api/treasury/proof-of-reserve` | GET | Доказательство резерва |
| `/api/reserve/por` | GET | PoR (старый endpoint) |
| `/api/reserve/proof` | GET | Расширенное доказательство с аудитом |
| `/api/treasury/usdt-deposits` | GET | История USDT депозитов |
| `/api/treasury/register-banknote` | POST | Регистрация банкноты |
| `/api/treasury/claim-dividends` | POST | Получение дивидендов |
| `/api/treasury/init-silver-round` | POST | Инициация silver round |
| `/api/treasury/auto-round` | GET | Автоматический silver round |

### 3.3 Серебряные активы
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/silver/mint` | POST | Создать silver-backed актив |
| `/api/silver/burn` | POST | Погасить актив |
| `/api/silver/list` | GET | Список silver активов |

### 3.4 RWA (Реальные Активы)
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/rwa/register` | POST | Зарегистрировать токенизированный актив |
| `/api/rwa/list` | GET | Список RWA |

### 3.5 Инвойсы
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/invoice/create` | POST | Создать инвойс |
| `/api/chat/invoice/list` | GET | Список инвойсов |
| `/api/chat/invoice/pay` | POST | Оплатить инвойс |
| `/api/chat/invoice/stats` | GET | Статистика инвойсов |
| `/api/chat/invoice/export-csv` | GET | Экспорт инвойсов в CSV |

---

## 4. AI и Стюард

### 4.1 AI Стюард (Ollama)
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/steward` | POST | Общение с AI Стюардом |
| `/api/ai/constitution` | GET | Конституционный анализ |
| `/api/ai/monitor` | GET | Метрики монитора |
| `/api/ai/steward-did` | GET | DID документ Стюарда |
| `/api/ai/radio-content` | GET | AI-генерация радио-контента |

### 4.2 Модерация
| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/admin/moderation-stats` | GET | Статистика модерации |

---

## 5. Вебхуки и Интеграции

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/chat/webhook` | GET/POST/PUT/DELETE | Конфигурация вебхука |
| `/api/admin/webhook-queue` | GET | Очередь вебхуков |
| `/api/webhook/whatsapp` | POST | WhatsApp интеграция |
| `/api/webhook/signal` | POST | Signal интеграция |
| `/api/webhook/matrix` | POST | Matrix интеграция |
| `/api/webhook/discord` | POST | Discord интеграция |

---

## 6. SimpleX Каналы (v6.5)

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/simplex/channel/create` | POST | Создать канал |
| `/api/simplex/channel/list` | GET | Список каналов |
| `/api/simplex/channel/join` | POST | Присоединиться к каналу |
| `/api/simplex/channel/post` | POST | Опубликовать в канал |

---

## 7. DID (Децентрализованная Идентичность)

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/did` | GET | DID документ узла |
| `/api/did/contact` | GET | DID документ контакта |

---

## 8. Меж-узел Relay (P2P)

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/relay/send` | POST | Отправить сообщение удаленному узлу |
| `/api/relay/receive` | GET | Получить сообщения |
| `/api/relay/history` | GET | История relay |

---

## 9. Радио

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/radio` | GET | Веб-радио плеер |
| `/api/radio/ai-content` | GET | AI-генерация радио-скриптов |

---

## 10. Информация об узле

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/version` | GET | Версия сборки |
| `/api/status` | GET | Статус сервера |
| `/api/health` | GET | Health check |
| `/api/admin/info` | GET | Полная информация об узле |
| `/api/admin/docker` | GET | Статус Docker контейнеров |
| `/api/admin/backup` | POST | Запустить бэкап |
| `/api/admin/metrics/system` | GET | Системные метрики (CPU, RAM, диск) |
| `/api/tracker/nodes` | GET | P2P трекер узлов |

---

## 11. Базы Данных

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/api/db/list` | GET | Список SQLite файлов |
| `/api/db/backup` | POST | Бэкап БД на USB |
| `/api/db/backup/list` | GET | Список бэкапов БД |
| `/api/db/restore` | POST | Восстановить БД |
| `/api/db/upload` | POST | Загрузить и восстановить БД |

---

## 12. Безопасность

### 12.1 Кошелек
| Endpoint | Описание |
|----------|----------|
| `/api/addresses` | Список адресов |
| `/api/rotate` | Ротация ключей |

### 12.2 CryptoContainer
| Endpoint | Описание |
|----------|----------|
| `/api/container/*` | Управление CryptoContainer |
| `/api/panic` | PANIC очистка (экстренная) |

### 12.3 Vault (16GB)
| Endpoint | Описание |
|----------|----------|
| `/api/vault/list` | Список файлов |
| `/api/vault/upload` | Загрузить файл |
| `/api/vault/download` | Скачать файл |
| `/api/vault/delete` | Удалить файл |
| `/api/vault/save-note` | Сохранить заметку |

### 12.4 Lock
| Endpoint | Описание |
|----------|----------|
| `/api/lock-status` | Статус блокировки |
| `/api/lock` | Заблокировать узел |
| `/api/unlock` | Разблокировать |
| `/api/change-lock-code` | Сменить код |

---

## 13. Диагностика и Аудит

| Endpoint | Описание |
|----------|----------|
| `/api/admin/audit-log` | Лог аудита безопасности |
| `/api/admin/metrics` | Метрики производительности |
| `/api/admin/diagnostics` | Полная диагностика |
| `/api/admin/status-page` | HTML страница статуса |
| `/api/admin/search-index` | Статус поискового индекса |
| `/api/admin/rate-limit-status` | Статус rate-limiter |
| `/api/admin/rate-limit-config` | Настройка rate-limiter |
| `/api/chat/analytics` | Аналитика чата |
| `/api/inquisitor/report` | Консолидированный отчет |

---

## 14. Telegram Боты

В узле работают 3 Telegram бота:
- **AskSteward** — команда `/steward` → AI Ollama
- **DarkPushkin** — экономика и администрирование
- **Torquemada** — инквизитор, получает отчеты

---

## 15. ParanoidX Форк

Находится в `/home/tomas/simplex-fork` (отдельный репозиторий).

| Компонент | LOC | Описание |
|-----------|-----|----------|
| AppTransport | 1,077 | Протокол передачи (конверт, очередь, replay) |
| Bridge | 616 | WS bridge + SOCKS5 прокси |
| Paranoid | 191 | VPN/TOR режимы (4) |
| VPN | 318 | VPN-сервис |
| SimpleX CLI | 344 | Клиент CLI |
| SMP | 186 | SMP протокол |
| Transport | 293 | Транспортный слой |

**Всего:** 20 Go файлов, ~3,693 LOC

### Режимы VPN/TOR (4)
1. **Full Tor** — весь трафик через Tor
2. **Split Tunnel** — чат через Tor, остальное напрямую
3. **VPN Only** — стандартный VPN без Tor
4. **Stealth** — рандомизированная маршрутизация

---

## 16. Быстрый старт

```bash
# Запуск узла
bash /home/tomas/simplex-node/scripts/launch-node.sh

# Проверка здоровья
curl http://localhost:8080/api/health

# Версия
curl http://localhost:8080/api/version

# Информация об узле
curl http://localhost:8080/api/admin/info

# Отправить сообщение
curl -X POST http://localhost:8080/api/chat/send \
  -H "Content-Type: application/json" \
  -d '{"chat_id":"@123456","text":"Привет из API"}'

# Цена серебра
curl http://localhost:8080/api/economy/oracle

# Бэкап
curl -X POST http://localhost:8080/api/admin/backup
```
