#!/bin/bash
# Launch the Telegram ↔ opencode bridge.
# Starts the listener which spawns opencode serve internally.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LISTENER="$SCRIPT_DIR/opencode-tg-listener.py"
LOG_DIR="$HOME/.local/share/opencode-tg"
PIDFILE="/tmp/opencode-tg-listener.pid"

mkdir -p "$LOG_DIR"

# Kill any previous instance
if [ -f "$PIDFILE" ]; then
  kill "$(cat "$PIDFILE")" 2>/dev/null || true
  sleep 1
fi

echo "Launching opencode Telegram bridge..."
echo "Log: $LOG_DIR/listener.log"
nohup python3 "$LISTENER" >> "$LOG_DIR/listener.log" 2>&1 &
PID=$!
echo $PID > "$PIDFILE"
echo "Started PID $PID"
echo "Stop: kill \$(cat $PIDFILE)"
