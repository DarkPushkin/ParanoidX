# The Island Project / Saint Mary Liberty Island - Digital Infrastructure (simplex-node A0+)

This node (with royal mode) is the **private digital backbone** for Saint Mary Liberty Island (stmaria.org / m007.org / markbank.org) and The Island Project.

## Core (A0 baseline)
- Private SMP/XFTP relays with persistent .onion.
- Full dashboard, Vault 2GB (E2EE capable), real voice/video (hidden coturn TURN), global lock, orchestration.
- See VERSION-A0, README A0 section, docs/A0-release-notes.md, PLAN-A0.md.

## New: Silver Token (Crypto-Taler / TLR) + Funding Loop + Funnel
- **TRON USDT TRC20 inflows** (cheap, popular) fund broker silver purchases for next "round" of tokenization of silver batch into reserve.
- Royal node triggers/ controls the round: USDT -> silver from broker -> tokenize batch (new ng in reserve, % to Island treasury).
- **Banknotes as shares**: registered physical/digital Mark Bank banknotes (denomination based) receive **pro-rata nanograms of silver** as dividends from **each new tokenized batch**.
- Mathematical funnel / black hole: dividends incentivize buying/holding banknotes (pay USDT -> funds more silver -> more tokenization -> more dividends -> more demand for shares/banknotes -> continuous accumulation of physical silver into the Island reserve from world market).
- Demo: in dashboard treasury card (or curl): simulate USDT, check (sees log), init round (processes batch, updates reserve, distributes divs to sample holders, credits dividend-*.txt in vault, logs rounds). See silver_reserve_ng.txt, banknotes_registry.json, silver_rounds.log, vault/dividend*.

## Tokenization of Everything Tool (Royal Controlled)
- Use silver tokens/ng as the "base" / unit of value / backing for tokenizing arbitrary physical items' value.
- First examples (as specified):
  - Silver coins: register physical coin, token = ownership/claim, backed by its silver ng in reserve.
  - Physical banknotes Mark Bank: register serials (holders claim share rights + dividends), tokenized for digital use while physical secure.
  - NFC Island passports: token = right to receive physical passport with NFC chip (for auth/payment in system, can hold token ref or TLR balance).
- Process: register item (type, serial, backing_ng from silver, holder, NFC UID, proofs/photos to Vault) -> issue asset token (SILVER-BACKED-..., with backing noted).
- Demo: curl POST /api/rwa/register with json (type, serial, backing_ng, holder, nfc_uid), then /api/rwa/list.
- Royal approves/controls all issuances across the node network.

## Royal Node
- Special mode (royal.enabled marker, set in startup).
- Full control over other (sub) nodes for **tokenization of everything**, rounds, dividends, reserve.
- Commands via secure SMP (E2EE - existing infra).

**Статус A0+ (эта сессия)**: Полный цикл TRON->round->dividends->accrued в реестре + RWA register (вкл. NFC passport) + Register Banknote Share форма + live state в дашборде + royal detection + vault proofs + funnel math в UI/логах/плане. Всё demoable. См. PLAN-A0.md (дополнение с research: novel combo, Kinesis closest но не то же самое). Запусти дашборд или curls — воронка работает.
- Sub nodes: local use, report to royal.
- This node (royal) acts as the "Mark Bank" controller / central issuer for the Island.

## Integration with Services & Island Priorities
- Vault: proofs, RWA metadata, dividend credits, passport data.
- Radio (future): announcements of rounds, "library" audio content about silver/funnel/Island.
- Media Channels (future): listings of tokenized assets, private subs for holders.
- Billing (future): price tokenization services, round participation in TLR/silver or USDT.
- Wallet (future): full TLR (dividends, mints), TRON USDT (inflows/payments).
- SMP: private comms for citizens, node-to-node control (royal-sub), ownership transfers of RWA/tokens.
- Supports Island priorities: global library (via media+Vault+tokenized docs), citizenship (paid TLR + passport token + NFC), banking/payments (wallet + TLR + USDT on-ramp + dividends), company reg, real estate (via tokenizer paid in silver), science data sharing.

## The Funnel / Black Hole for World Silver
USDT (public on-ramp, TRON cheap) -> buy silver from broker -> new batch tokenized (reserve + treasury + dividends to banknote shares) -> holders benefit -> demand for banknotes (shares) rises -> more USDT capital in -> more silver bought -> ... self-reinforcing accumulation of physical silver into the sovereign reserve.

## Next / Polish (see PLAN-A0.md todos)
- Royal explicit mode + sub node control via SMP.
- Full RWA UI wizard in dashboard.
- Billing with TLR prices, payments log, hooks.
- Radio stream from Vault + channels for RWA/library.
- Silver wallet (TLR units, dividends view).
- Physical audits/PoR for silver, legal wrappers for "Island law".
- Scale: many subs, more RWAs (Island properties etc.).

See full detailed plan in docs/PLAN-A0.md (or sessions backup).

This is A0+ implementation of the vision: private + real silver backed + hierarchical control + tokenization of everything + economic funnel for accumulation.

## Alignment with declared Saint Mary Liberty Island / Mark Bank (stmaria.org, markbank.org, m007.org)
Real project exists: sovereign micronation declaration in neutral waters (specific coords), Mark Bank as central bank issuing Taler TLR (1 TLR = 1 troy oz 99.99 silver, 70% physical reserved via SolCity US vaults partner, 30% ops), physical + crypto banknotes (signed PDFs), citizenship (passport benefits), company reg, library priority, etc. Laws, notifications, general license published.

Our royal node + TRON USDT on-ramp + broker silver batches -> tokenized in reserve + 20% treasury / 80% pro-rata ng dividends to registered Mark Bank banknote "shares" (equity) + RWA "tokenization of everything" (silver coins, physical banknotes as shares, NFC Island passports with payment chip as first citizenship+pay instruments) + SimpleX private comms/Vault-as-library/radio+channels monetized in ng unit = the **private digital sovereign infrastructure layer** that makes the economic "mathematical black hole funnel" operational.

Holding a registered physical Mark Bank banknote = ownership of a share that earns nanograms of the next broker silver batch (self-reinforcing: dividends attract more capital/holders -> more USDT inflows -> more silver accumulation into Island reserve from world market).

No prior art matches the full combo (Kinesis yields from platform fees not exact per-batch pro-rata to specific banknote equity shares + no royal micronation control + no SimpleX private + no "tokenize everything" with silver base for passports etc.). This is novel extension realizing the Island's own currency + sovereignty vision with a powerful accumulation engine.

Run the node, use the treasury card in dashboard (or curl the APIs) to demo the loop and RWA. Set real funded TRON addr in tron_treasury.txt for live public inflows. Royal marker (royal.enabled) + is_royal in state.
