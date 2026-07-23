#!/bin/bash
# Send reports to Torquemada bot — supports text + document + photo.
# Usage:
#   send-to-torquemada.sh "text message"
#   send-to-torquemada.sh -d /path/to/file.md "caption"
#   send-to-torquemada.sh -p /path/to/photo.png "caption"
set -euo pipefail

TOKEN_FILE="${HOME}/.config/torquemada-bot.token"
CHAT_FILE="${HOME}/.config/torquemada-bot.chat"

# Fallback: read from simplex-node config
if [ ! -f "$TOKEN_FILE" ] || [ ! -f "$CHAT_FILE" ]; then
  CONFIG="${HOME}/.local/share/simplex-node/simplex-node.json"
  if [ -f "$CONFIG" ]; then
    python3 -c "
import json
with open('$CONFIG') as f:
    d = json.load(f)
with open('$TOKEN_FILE', 'w') as f:
    f.write(d['torquemada_token'])
with open('$CHAT_FILE', 'w') as f:
    f.write(str(d['torquemada_chat_id']))
" 2>/dev/null
  fi
fi

if [ ! -f "$TOKEN_FILE" ]; then
  echo "ERROR: no token file"
  exit 1
fi

TOKEN=$(tr -d '\n\r' < "$TOKEN_FILE")
CHAT_ID=$(tr -d '\n\r' < "$CHAT_FILE")

case "${1:-}" in
  -d|--document)
    FILE="${2:?No file specified}"
    CAPTION="${3:-}"
    [ -f "$FILE" ] || { echo "ERROR: file not found: $FILE"; exit 1; }
    curl -s -X POST "https://api.telegram.org/bot${TOKEN}/sendDocument" \
      -F "chat_id=${CHAT_ID}" \
      -F "document=@${FILE}" \
      -F "caption=${CAPTION}" | python3 -c '
import sys, json
d = json.load(sys.stdin)
if d.get("ok"):
    print("✅ Torquemada doc sent, message_id:", d["result"]["message_id"])
else:
    print("❌ Error:", d.get("description", d))
'
    ;;
  -p|--photo)
    FILE="${2:?No file specified}"
    CAPTION="${3:-}"
    [ -f "$FILE" ] || { echo "ERROR: file not found: $FILE"; exit 1; }
    curl -s -X POST "https://api.telegram.org/bot${TOKEN}/sendPhoto" \
      -F "chat_id=${CHAT_ID}" \
      -F "photo=@${FILE}" \
      -F "caption=${CAPTION}" | python3 -c '
import sys, json
d = json.load(sys.stdin)
if d.get("ok"):
    print("✅ Torquemada photo sent, message_id:", d["result"]["message_id"])
else:
    print("❌ Error:", d.get("description", d))
'
    ;;
  -*)
    echo "Usage: $0 [-d file caption | -p file caption | text]"
    exit 1
    ;;
  *)
    MSG="${1:-Torquemada update}"
    curl -s -X POST "https://api.telegram.org/bot${TOKEN}/sendMessage" \
      -H "Content-Type: application/json" \
      -d "$(python3 -c "
import json
print(json.dumps({'chat_id': ${CHAT_ID}, 'text': '''${MSG}'''[:4000]}))
")" | python3 -c '
import sys, json
d = json.load(sys.stdin)
if d.get("ok"):
    print("✅ Torquemada msg sent, message_id:", d["result"]["message_id"])
else:
    print("❌ Error:", d.get("description", d))
'
    ;;
esac
