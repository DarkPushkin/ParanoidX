# Production Cycle — Standard Operating Procedure

**Version:** 1.0
**Date:** 2026-06-08
**Project:** simplex-node / Saint Mary Liberty Island

---

## Cycle Flow

Each development cycle follows exactly 8 steps. No step is skipped.

```
  ┌─────────────────────────────────────────────────────┐
  │                    CYCLE START                       │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  1. BACKUP                                           │
  │     ├── Source snapshot (tar.gz)                     │
  │     ├── Live data snapshot (registries, ledger,      │
  │     │   vault files, configs)                        │
  │     └── Save to /home/tomas/A1-backups/cycle-N/      │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  2. RE-WRITE THEPLAN                                  │
  │     ├── Read THEPLAN.md                               │
  │     ├── Read all docs/*.md for context                │
  │     ├── Incorporate new requests from admin           │
  │     ├── Update version, phase, status                 │
  │     └── Verify consistency with docs/EVOLUTION-PLAN.md│
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  3. REPORT RE-WRITTEN PLAN TO ADMIN BOT               │
  │     ├── Send summary to @torquemada878_bot            │
  │     ├── Include: version, changes, next cycle goals   │
  │     └── Use: scripts/send-to-torquemada.sh            │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  4. CHOOSE 1-2-3 STEPS FOR THIS CYCLE                 │
  │     ├── From THEPLAN or EVOLUTION-PLAN.md             │
  │     ├── Prioritize: bugs > security > features        │
  │     ├── max 3 steps per cycle (focus)                 │
  │     ├── Report chosen steps to admin bot              │
  │     └── Update todo list                              │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  5. BUILD (according to cycle plan)                   │
  │     ├── One step at a time                            │
  │     ├── After each step: `go build` must pass         │
  │     ├── After each step: `go vet` must pass           │
  │     └── Commit after each completed step              │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  6. RUN TESTS + DEBUG                                 │
  │     ├── Integration: scripts/test-royal.sh            │
  │     ├── If test fails: debug, fix, re-test            │
  │     ├── Race: go test -race ./internal/...            │
  │     ├── Lint: go vet ./...                            │
  │     └── Build: go build -o /dev/null ./cmd/...        │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  7. CREATE REPORT                                     │
  │     ├── What was done (steps completed)               │
  │     ├── What changed (files modified, new files)      │
  │     ├── Test results (pass/fail, coverage if added)   │
  │     ├── Issues found                                  │
  │     ├── What's next (recommendations for next cycle)  │
  │     └── Send report to admin bot                      │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
  ┌─────────────────────────────────────────────────────┐
  │  8. CALL ADMIN TO INTERACT                            │
  │     ├── Present report summary                        │
  │     ├── Ask: approve, adjust, or reject for next      │
  │     ├── If approve → start next cycle (goto step 1)   │
  │     ├── If adjust → update plan, re-run cycle         │
  │     └── If reject → stop, document reasoning          │
  └─────────────────────────────────────────────────────┘
                            │
                            ▼
                    WAIT FOR ADMIN INPUT

```

---

## Rules

### R1: Backup always comes first
No code changes before backup. If backup fails, cycle stops.

### R2: Max 3 steps per cycle
Small, focused cycles. No 10-step epics. If a step is > 4 hours of work, split it.

### R3: Build must compile after every step
After each completed step, run `go build ./cmd/...`. If it doesn't compile, fix before next step.

### R4: Tests before report
Step 6 is mandatory. If tests fail and can't be fixed, document as known issue in the report.

### R5: Admin approval gates cycle start
Cycle starts only after admin interaction in step 8. Never auto-start a new cycle.

### R6: One person — one cycle
All steps in a cycle are done by the same person (AI). No handoffs mid-cycle.

---

## Backup Format

```
~/A1-backups/cycle-N/
├── MANIFEST.txt          — what, when, why
├── src/                  — git archive of source
├── data/                 — live data snapshot
│   ├── banknotes_registry.json
│   ├── ledger.json
│   ├── rwa_registry.json
│   ├── silver_reserve_ng.txt
│   ├── silver_rounds.log
│   ├── treasury_usdt.log
│   ├── channels.json
│   ├── billing_payments.log
│   └── vault/            — key files only
├── plans/                — THEPLAN.md, EVOLUTION-PLAN.md, docs/
└── configs/              — royal-bot.token, etc. if changed
```

---

## Report Template

```markdown
## Cycle N Report

**Date:** YYYY-MM-DD
**Status:** ✅ Complete / ⚠️ Partial / ❌ Failed

### Steps Completed
1. [Step name] — ✅/⚠️/❌ — brief result
2. [Step name] — ✅/⚠️/❌ — brief result
3. [Step name] — ✅/⚠️/❌ — brief result

### Files Changed
- path/to/file.go — what changed
- path/to/new_file.go — created, purpose

### Test Results
- scripts/test-royal.sh: ✅ PASS / ❌ FAIL
- go vet ./...: ✅ PASS / ❌ FAIL
- go build ./cmd/...: ✅ PASS / ❌ FAIL
- go test -race: ✅ PASS / ❌ FAIL

### Issues Found
- Bug/issue encountered, current status

### Next Cycle Recommendations
- Suggested steps for next cycle
```

---

## Version Tracking

Each cycle updates `VERSION` file:
```
CYCLE-1  (A1 -> A1.1)
CYCLE-2  (A1.1 -> A1.2)
...
```

`THEPLAN.md` header includes current cycle number and date.

---

*Document created: 2026-06-08*
*Part of simplex-node production SOP*
