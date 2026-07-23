# Testing Royal Node Services (A1 focus)

## Principles
- Invariants: 80/20 split exact (int), pro-rata by denom sum==1, accrued monotonic, RWA backing <= reserve, royal gate enforced.
- Reproducible fixtures in testdata/royal-fixtures/.
- Levels: unit (Go), integration (test-royal.sh + curl/python), E2E (multi round + dynamic holders), sub-agent (spawn_subagent with docs/tester-prompt-royal.txt).
- Use launch-node.sh or controlled bin start + always pkill.

## Fixtures
- minimal/: reserve=0, 2 banknotes (1+10), royal.enabled, empty rwa/logs/vault.
- Add goldens: expected_ann_after_123usdt.txt etc.

## Harness
- scripts/test-royal.sh : launches, runs simulate/init/register/RWA, asserts via python (exact ng, text contains "чёрной дыры"), checks vault anns, radio, payments.
- Run: ./scripts/test-royal.sh (uses synthetic if no fixture).

## Sub-agent-tester
Use:
spawn_subagent(
  prompt = open("docs/tester-prompt-royal.txt").read() + " Use A1 data or reset to fixture. Report to /tmp/royal-tester-report.txt",
  subagent_type="general-purpose",
  capability_mode="execute",
  description="Autonomous A1 royal funnel regression"
)
The agent will launch/kill, curl, python math, assert, produce report.

## Go units (TODO add)
Extract from main.go:
func calcNewSilverNg(usdt float64) int64 { return int64(usdt * 1e9) }
func proRata(holders []Holder, pool int64) []Dividend { ... }
Table tests for 1+10+42, batch 1e12 etc.

## Bot reports
After test or phase: ./scripts/send-to-torquemada.sh "A1 tester: 6/6 checks PASS. Reserve grew, divs pro-rata, NFC RWA ok. #A1 #royal #funnel"

See plan.md for full + data needs.
