# MacBook Pro M4 Max 128GB — AI-лошадка + Кросс-платформенная разработка

## Почему этот MacBook идеален

### 1. Unified Memory — ключевое преимущество

У большинства PC видеокарта имеет **24GB VRAM** (RTX 5090 за $2000+). У тебя — **128GB unified memory**, которая доступна и CPU, и GPU одновременно. Это значит:

- **Llama 3.1 70B** (Q4, ~40GB) — влезает целиком с запасом
- **Qwen3-Coder-Next 80B** (Q4, ~48GB) — основная кодовая модель — влезает
- **DeepSeek-V4** (MoE, 284B total) — можно запустить в Q2
- **Не нужно выбирать**, какая модель поместится — помещаются все

### 2. Пропускная способность памяти

M4 Max выдаёт **546 GB/s**. Это больше, чем у GMKtec EVO-X2 (256 GB/s) и любого другого mini PC на рынке. Для LLM скорость генерации токенов упирается **именно в bandwidth памяти**, а не в вычислительную мощность.

Сравнение скорости на 70B-модели:
| Устройство | tok/s |
|-----------|-------|
| RTX 4090 (24GB) | 15-18 (но не влезает, надо offloading) |
| GMKtec EVO-X2 (128GB) | 4-8 |
| **MacBook M4 Max (128GB)** | **11-12** |
| Mac Studio M3 Ultra (256GB) | 15-18 |

### 3. MLX — родной фреймворк Apple

MLX — это Apple-оптимизированный рантайм для нейросетей. Он работает **напрямую с Metal GPU** и unified memory, без прослоек CUDA/ROCm/Vulkan. На M4 Max MLX даёт **на 20-30% больше tok/s**, чем llama.cpp.

Ollama 0.19+ использует MLX как бэкенд по умолчанию на M-серии — никаких плясок с бубном.

### 4. Энергопотребление

Весь MacBook потребляет **40-80W** при полной загрузке GPU. Для сравнения: RTX 4090 одна жрёт 450W. Это значит:
- Тишина (вентиляторы еле слышно)
- Можно держать включённым 24/7 за ~€5 электроэнергии в месяц
- Не греет офис

### 5. Tailscale — как локальная сеть, только через интернет

Tailscale на скорость не влияет (<1% оверхеда, микросекунды задержки). Это просто WireGuard, который делает так, будто MacBook и сервер simplex-node в одной локальной сети, даже если они физически в разных местах.

---

## Инструменты для кросс-платформенной разработки

### Flutter SDK (уже установлен)

Flutter 3.44.1 + Dart 3.12.1 у тебя уже стоят. Для **Linux desktop** нужно добавить:

```bash
# Включить поддержку Linux desktop
flutter config --enable-linux-desktop

# Проверить, что появилась
flutter devices
# Должен показать: Linux (desktop)
```

### Основной стек

| Задача | Инструмент | Почему |
|--------|-----------|--------|
| **Кроссплатформенный UI** | Flutter + Material 3 | Один код — Linux, macOS, Windows, Web, Android, iOS |
| **API клиент** | `http` пакет (уже есть) | typed client к simplex-node REST API |
| **Состояние** | Provider (уже есть) | Достаточно для этого размера приложения |
| **QR-коды** | qr_flutter (уже есть) | Для банкнот, платежей |
| **Биометрия** | local_auth (уже есть) | Face ID / Touch ID для кошелька |
| **Файлы** | file_picker (уже есть) | Для vault |
| **Хранение** | shared_preferences (уже есть) | Настройки, последний pubkey |
| **Графики** | fl_chart (в royal_app) | Для экономической аналитики |

### Что ещё пригодится

```bash
# Dart-инструменты
dart pub global activate very_good_cli    # лучший шаблон Flutter-проектов
dart pub global activate melos            # монорепозиторий (у нас уже есть)
dart pub global activate dart_code_metrics # статический анализ

# Инструменты разработчика
brew install --cask visual-studio-code
brew install --cask postman              # тестирование API
brew install --cask notion               # документация

# Для AI
brew install jq                           # работа с JSON в терминале
```

---

## План: The Isle — Linux Desktop

### Текущее состояние

Приложение имеет:
- **5 экранов**: Dashboard, Wallet, Vault, Market, Radio
- **3 shared пакета**: api_client, models, widgets
- **Provider** для состояния
- **Material 3** дизайн

### Что нужно сделать

1. **Исправить импорты** — код использует `package:shared_widgets/` и `package:models/`, но реальные имена пакетов — `widgets` и `models`
2. **Включить Linux Desktop** — `flutter config --enable-linux-desktop`
3. **Создать linux/ директорию** — Flutter сгенерирует её сам при первом `flutter create --platforms=linux .`
4. **Настроить иконку и title** для Linux-окна
5. **Собрать и запустить** — `flutter run -d linux`
6. **Добавить POS-терминал** — экран приёма платежей
7. **Улучшить Radio** — реальный плеер вместо заглушки

### Архитектура

```
apps/isle_app/
├── lib/
│   ├── main.dart
│   ├── screens/
│   │   ├── dashboard_screen.dart    ← главная с балансом
│   │   ├── wallet_screen.dart       ← кошелёк, переводы
│   │   ├── vault_screen.dart        ← файловое хранилище
│   │   ├── market_screen.dart       ← маркетплейс
│   │   └── pos_screen.dart          ← NEW: POS-терминал
│   ├── services/
│   │   └── isle_api_service.dart
│   └── widgets/                     ← локальные виджеты
└── linux/                           ← Flutter сгенерирует
    ├── CMakeLists.txt
    ├── main.cc
    └── my_application.cc
```

---

## Как пользоваться

### На MacBook (AI-сервер)

После `setup-macbook-ai.sh`:
```bash
ai-status        # проверить, что работает
ai-info          # узнать Tailscale IP для подключения
```

### На сервере simplex-node (через opencode)

Я подключаюсь к MacBook по Tailscale и гоняю код через Qwen3-Coder-Next.
Для разработки Linux Desktop тебе не нужен MacBook — всё собирается на сервере:

```bash
cd apps/isle_app
flutter pub get
flutter run -d linux   # или flutter build linux
```
