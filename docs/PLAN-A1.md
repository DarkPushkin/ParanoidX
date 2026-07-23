# A1.2 / A1.3 Status + Pre-Reboot Save (see full session plan for details)

Current (post deep review + Phase 0 + save before update/reboot): Automatic bot listener (2s poll, passwordless, exact "шаг (...) завершен. жду в главной консоли." template), version checkpoints (A1.3-* via script + bot), royal-common.sh central paths, marketplace /api/market stubs, 20+ bot cmds (plan/build/gobuild, edit, market_*, checkpoint, test, royal_*, etc.), testing harness, silver funnel (20/80 pro-rata ng to banknote shares), RWA incl. island_nfc_passport, radio anns with "чёрной дыры" text.

Phase 0 executed: listener fixes (gobuild, offset max-all, signals always, pw remnants removed), common.sh + hardcode reduction, test-royal clean, node-up checks, gofmt, main TODO for extract, VERSION updated, A0 garbage noted, full verify (test PASS, math 53 denom, royal, anns), version-checkpoint A1.3 + exact signals.

**SAVE CONTEXT + STATE BEFORE REBOOT/UPDATE (2026-06-03)**: Disk was full; created /home/tomas/A1-backups/pre-reboot-save-20260603-204116/ (src snapshot with all fixes + current session-plan + live-critical registries/silver/anns + MANIFEST with PIDs/rollback) + .tar.gz. PRE-REBOOT-SAVE.txt in live data + src backup. Freed ~7G. Live data + tor HS survive reboot. Post-reboot: launch via scripts/launch-*.sh ; restore from pre-reboot-save or kept A1.3-202630 data if needed. Full plan + context in the save package.

Full detailed plan + execution log + recovery: in /home/tomas/.grok/sessions/.../plan.md (the primary living document).

Services may be down after reboot/update — (re)start with:
  /home/tomas/simplex-node/scripts/launch-node.sh
  /home/tomas/simplex-node/scripts/launch-bot-listener.sh

See scripts/ for signal_step_done.sh, version-checkpoint.sh, listener, royal-common.sh, etc. Send commands to @torquemada878_bot for signals/DONE.

(Refer to session plan for Phase 1+ : robust parse + tests, real TRON, royal signed control, market UI, etc. Old A0 content removed for compactness.)

**Added (disk incident follow-up)**: comprehensive disk metrics in /api/status + /api/disk-check (root %, data/vault/backups), background 60s checker that sends TG alerts on low space via send-to, preflight warning in launch-node.sh, enhanced dashboard storage card with gauges + "Проверить диск" button, bot "disk"/"disk_check" cmds. Full list of future metrics (reserve graphs, economy velocity, alerts.log) in session plan.md "Observability" section.
