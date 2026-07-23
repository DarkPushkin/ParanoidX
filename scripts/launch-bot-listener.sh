#!/bin/bash
# Safe launch for automatic bot listener (polls every 2s, executes commands, signals DONE)

source "$(dirname "$0")/royal-common.sh" 2>/dev/null || true
: "${LISTENER:=${SCRIPTS_DIR:-$(dirname "$0")}/royal-telegram-command-listener.sh}"
: "${LOG:=${DATA_DIR:-$HOME/.local/share/simplex-node}/logs/bot-listener.log}"
: "${PIDFILE:=/tmp/royal-bot-listener.pid}"

mkdir -p "$(dirname "$LOG")"

# Clean old
if [ -f "$PIDFILE" ]; then
  kill $(cat "$PIDFILE") 2>/dev/null || true
fi
sleep 0.5

# Bootstrap offset to latest on (re)start: prevents replay of old messages after reboot/tmp clear or restart.
# This ensures only *new* messages after launch are received/processed/logged as USER.
if [ -f "$TOKEN_FILE" ]; then
  TOKEN=$(cat "$TOKEN_FILE" 2>/dev/null)
  LATEST=$(curl -s "https://api.telegram.org/bot${TOKEN}/getUpdates?limit=1" 2>/dev/null | python3 -c '
import sys,json
d=json.load(sys.stdin)
res=d.get("result",[])
if res:
  print(res[-1].get("update_id",0))
else:
  print(0)
' 2>/dev/null || echo 0)
  if [ "$LATEST" -gt 0 ]; then
    echo $LATEST > "$OFFSET_FILE"
    echo "Bootstrap: offset set to $LATEST (only future msgs will be processed)"
  fi
fi

echo "Launching automatic Telegram listener (2s poll)..."
# Use while in subshell to avoid & issues in some tools
nohup bash -c "
  while true; do
    \"$LISTENER\" once
    sleep 2
  done
" >> "$LOG" 2>&1 &
echo $! > "$PIDFILE"

echo "Started PID $(cat $PIDFILE). Log: $LOG"
echo "Commands from bot will auto-execute, '✅ DONE' sent when finished."
echo "Stop: kill \$(cat $PIDFILE)"
