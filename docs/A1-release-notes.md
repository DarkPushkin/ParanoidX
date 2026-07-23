# A1 Release Notes — Ранняя альфа 1 (заготовка для минимально функционирующей королевской ноды)

**Дата**: 2026-06-03
**Версия**: A1 (см. VERSION-A1 в корне, бэкапы simplex-node-A1-20260603-*, ~/.local/share/simplex-node-A1-*)
**Бэкапы**: /home/tomas/simplex-node-A1-20260603-152104/ (source + VERSION-A1), data backup, plan backups, /home/tomas/A1-backups/PLAN-A1-*.md

## Что такое A1
Ранняя альфа 1 — "заготовка / подготовка" для **минимально функционирующей королевской ноды** Saint Mary Liberty Island (stmaria.org / m007.org / markbank.org).

Это эволюция A0: добавлена полная серебряная экономика + royal control prep на базе уже изобретённого (TRON USDT funding loop для оплаты брокеру, токенизация партий серебра в резерв, банкноты Mark Bank = акции с pro-rata ng дивидендами с каждой партии (20% казначейство Острова, 80% holders), "токенизация всего" tool с первыми примерами silver_coin / mark_banknote / island_nfc_passport с NFC чипом, математическая воронка-чёрная дыра, royal marker + is_royal, auto radio anns с русским текстом про воронку, channels monetized в ng, billing skeleton, полная Treasury UI в дашборде).

A1 уже позволяет:
- Запустить (только через scripts/launch-node.sh !) → royal node с ~53.7 кг в резерве.
- Симулировать USDT inflows → init round (royal) → проверить math (reserve growth, 20/80, pro-rata на 1+10+42 TLR shares, accrued persist, vault proofs, round log, radio ann с "чёрной дыры").
- Register banknote shares и RWA (вкл. NFC passport).
- Видеть всё в дашборде (royal badge, таблицы, формы).
- Funnel работает: дивиденды incentivize holding banknotes → больше capital → больше silver accumulation.

**Это база для завершения "минимально функционирующей королевской ноды"** (real funded TRON, SMP signed royal→subs control, auto rounds, PoR, claims, wallet) + **comprehensive testing** (unit+integration+E2E + суб-агент-тестер на spawn_subagent) + Telegram bot @torquemada878_bot для планов/отчётов.

## Основные фичи A1 (royal prep snapshot)
- TRON USDT TRC20 treasury (TronGrid live query + simulate-usdt, log в treasury_usdt.log).
- Royal (royal.enabled, isRoyalNode helper, is_royal в state/treasury).
- Init silver round: 20% treasury, 80% pro-rata по denomination_tlr (banknotes=shares), robust reserve parse, persisted accrued, vault dividend/share/ann proofs, silver_rounds.log, billing hook, radio ann с точным "РАУНД ТОКЕНИЗАЦИИ СЕРЕБРА ... 20% ... 80% ... математической воронки — чёрной дыры для мирового серебра".
- Register banknote (как equity для всех будущих партий).
- RWA tokenizer: silver_coin, mark_banknote, island_nfc_passport (nfc_uid + SILVER-BACKED- token + backing check).
- State: /api/treasury/state (reserve ng/g/oz, shares с accrued, rwa, rounds, is_royal, prices, funnel note).
- Billing: prices + payments.log (round/rwa/channel в ng).
- Dashboard: полная Treasury (royal badge, live, Sim/Init/Register forms для banknote + RWA+NFC).
- Channels + radio tie-in (Black Hole News, anns из rounds).
- Launch hygiene: scripts/launch-node.sh с warnings, aggressive kill, disown (против background spam).
- Docs: research (Kinesis closest но не то; реальный stmaria/markbank alignment — нода + royal как цифровой Mark Bank + funnel engine), ISLAND.md, client-info с vision + hygiene, VERSION-A1, README A1 section, PLAN-A1.md (с detailed testing + sub-agent-tester + bot).
- Live data (A1 snapshot): 53.704 кг, royal=true, 3 banknotes (1/10/42 TLR с accrued), 3 RWA (вкл. NFC-PASS-001 uid 04AABBCCDD), funnel demo (250 USDT round прошёл), channels, anns с текстом.

## Известные ограничения A1
- Реальный funded TRON addr (TYourReal...) + live broker/SolCity — placeholder + simulate (для тестов).
- Royal control над subs — marker + soft gate + skeleton; полный SMP signed commands + dashboard panel — в A1.1.
- Backing enforcement для RWA, claim dividends, PoR stubs, auto threshold rounds — частично (в план).
- Полноценные тесты (unit для math, integration скрипт test-royal.sh, суб-агент-тестер, fixtures/goldens) — в процессе (A1.1).
- Telegram bot @torquemada878_bot integration (scripts/send-to-...) — скрипт + интеграция в flows, но token/chat_id от пользователя (placeholder в A1 fixture).
- E2EE client-side, physical reserved full, hot wallet — как в A0.
- Нет full sub network live.

## Roadmap после A1 (из утверждённого плана)
1. A1.0 formalize (done) + A1.1 completion (real TRON, royal SMP control, RWA enforce + claims + PoR, testing harness + sub-agent-tester + bot integration).
2. A1.2 security/packaging/release A1.
3. A2: auto funnel, sub network, full wallet/billing, first investors demo (real funded + public inflows).
4. A3+: scale, more RWA (Island properties), legal, citizenship flows, full Island layer.

См. полный план: docs/PLAN-A1.md (в main + A1-backups; включает testing strategy, sub-agent-tester prompt, bot integration, roadmap). docs/ISLAND.md, VERSION-A1, client-connection-info.txt.

## Изменения с A0
- Полная silver/royal/funnel/RWA/tokenization-of-everything (по точному запросу пользователя про TRON, royal, банкноты=акции, NFC passports, чёрную дыру).
- Hygiene launch script + warnings.
- Research + docs обновления (novelty + real project alignment).
- Бэкапы A1 + markers.
- Фокус на тестировании и Telegram отчётах (новое в A1).

**A1 готов как заготовка для королевской ноды + база для тестов и следующей фазы. Запускай только launch-node.sh, используй Treasury для демо funnel.**

(Адаптировано из A0 notes + live A1 snapshot.)
