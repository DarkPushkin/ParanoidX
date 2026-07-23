#!/bin/bash
# Regenerate QR codes from current address files
set -euo pipefail

DATA_DIR="${1:-$HOME/.local/share/simplex-node}"

SMP_FILE="$DATA_DIR/smp_client_address.txt"
XFTP_FILE="$DATA_DIR/xftp_client_address.txt"
AUDITOR_FILE="$DATA_DIR/auditor_onion.txt"
ADDR_FILE="$DATA_DIR/addresses.json"

SMP_CLIENT=$(cat "$SMP_FILE" 2>/dev/null || echo "")
XFTP_CLIENT=$(cat "$XFTP_FILE" 2>/dev/null || echo "")
AUDITOR=$(cat "$AUDITOR_FILE" 2>/dev/null || echo "")
ICAL=$(python3 -c "import json; d=json.load(open('$ADDR_FILE')); print(d.get('ice',''))" 2>/dev/null || echo "")
CONTACT=$(cat "$DATA_DIR/island_contact_link.txt" 2>/dev/null || echo "")

UPDATED=0

if command -v qrencode >/dev/null 2>&1; then
  if [ -n "$SMP_CLIENT" ]; then
    qrencode -s 6 -o "$DATA_DIR/qr-smp.png" "$SMP_CLIENT" 2>/dev/null
    echo "qr-smp.png: $SMP_CLIENT"
    UPDATED=1
  fi
  if [ -n "$XFTP_CLIENT" ]; then
    qrencode -s 6 -o "$DATA_DIR/qr-xftp.png" "$XFTP_CLIENT" 2>/dev/null
    echo "qr-xftp.png: $XFTP_CLIENT"
    UPDATED=1
  fi
  if [ -n "$AUDITOR" ]; then
    qrencode -s 6 -o "$DATA_DIR/qr-auditor.png" "http://$AUDITOR/auditor" 2>/dev/null
    echo "qr-auditor.png: http://$AUDITOR/auditor"
    UPDATED=1
  fi
  if [ -n "$ICAL" ]; then
    qrencode -s 6 -o "$DATA_DIR/qr-ice.png" "turn:$ICAL:3478?transport=tcp" 2>/dev/null
    echo "qr-ice.png: turn:$ICAL:3478?transport=tcp"
    UPDATED=1
  fi
  if [ -n "$CONTACT" ]; then
    qrencode -s 6 -o "$DATA_DIR/qr-island-services.png" "$CONTACT" 2>/dev/null
    echo "qr-island-services.png: $CONTACT"
    UPDATED=1
  fi
  if [ "$UPDATED" -eq 1 ]; then
    echo "QR codes regenerated in $DATA_DIR"
  else
    echo "No address data found — QR codes unchanged"
  fi
else
  echo "qrencode not found — install: sudo apt install qrencode"
  exit 1
fi
