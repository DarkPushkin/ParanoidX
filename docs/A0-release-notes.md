# A0 Release Notes — Нулевая альфа simplex-node

**Дата**: 2026-06-03
**Версия**: A0 (см. VERSION-A0 в корне и бэкапах simplex-node-A0-*)

## Что такое A0
Минимально функционирующий продукт на базе изобретённого в истории проекта:
- Полноценные приватные релеи SMP (сообщения) + XFTP (файлы/медиа) с постоянными .onion (bind-mount persistence, custom Tor Alpine + su-exec).
- Красивый дашборд (vanilla JS, мобильный): статус (uptime + реальные docker ps), адреса с truncate/COPY/QR, ICE/TURN с pasteLines для клиентов SimpleX, Vault 2GB (upload/list/previews/notes/quota), голосовые записи авто в Vault, реальные голос/видео (SMP-сигналинг + hidden coturn TURN с динамическими кредами ~12ч и 2-tab тестером), глобальный 6-значный server-side lock (точный UX), ротация адресов.
- Оркестрация: startup.sh (всё от certs/bootstrap до красивого блока в терминале + client-connection-info.txt), systemd.
- Строгая приватность/доступ: всё в onion где возможно, дашборд/API только 127.0.0.1 или dashboard-onion.

A0 уже usable как standalone приватный "суверенный узел" для SimpleX (релеи + звонки + хранилище + контроль).

**Бэкап A0**: /home/tomas/simplex-node-A0-... (source + data + VERSION-A0).

## Основные фичи (подробно)
(Копия/адаптация из плана и README: relays, voice, vault, dashboard, lock, ops и т.д.)

## Известные ограничения A0
- Физическая reserved для Vault (fallocate) — в процессе добавления (см. roadmap).
- Полноценный hot wallet / E2EE client-side в upload — базово (рекомендация), улучшения в следующих фазах.
- Монетизация и новые сервисы (радио, каналы) — skeleton + prototypes в roadmap.
- Нет Prometheus/health полного, backup.sh — в полировке.
- Токен серебра (Crypto-Taler TLR, нанограммы физ. серебра в хранилище Mark Bank per stmaria/markbank) — интеграция начинается после A0.

## Roadmap после A0 (из утверждённого плана)
1. A0.0 formalize + A0.1 hardening (vault reserved/E2EE, voice polish, billing skeleton, docs).
2. Новые сервисы: радио-стрим из Vault (полный), анонимные медиа-каналы с монетизацией трафика.
3. Токенизированный актив: wallet для TLR/silver_ng, pricing сервисов в токене, integration с billing.
4. Эволюция в The Island Project / Saint Mary Liberty Island: библиотека (Vault+radio+channels), banking в TLR, citizenship, real estate tokens, "мобильный банкинг" и т.д. Нода как цифровое сердце Острова (коммуникации SimpleX, глобальная библиотека, экономика в физ. серебре).

См. полный план в корне plan.md (или A0-бэкап) и docs/ISLAND.md (будет).

## Изменения с предыдущих
(История итераций: persistent onion, lock UX, vault quota+reserved, voice с tester и pasteLines, etc. — все закалены.)

**Готов к использованию как A0 alpha для приватных нужд и базы The Island Project.**
