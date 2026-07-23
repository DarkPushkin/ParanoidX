#!/bin/bash
# Sync Radio/ folder (project root) → server radio data directory.
# Detects new/modified/deleted files and logs changes.
# Designed to run every minute via cron.
set -euo pipefail

SRC="${HOME}/simplex-node/Radio"
DST="${HOME}/.local/share/simplex-node/radio"
LOG="${HOME}/.local/share/simplex-node/radio/sync.log"

mkdir -p "$DST"

changes=0

# Process new/modified files: copy Radio/*.mp3 → server radio dir
for f in "$SRC"/*.mp3; do
    [ -f "$f" ] || continue
    name=$(basename "$f")
    dst="$DST/$name"
    if [ ! -f "$dst" ] || [ "$(md5sum "$f" | cut -d' ' -f1)" != "$(md5sum "$dst" 2>/dev/null | cut -d' ' -f1)" ]; then
        cp "$f" "$dst"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] COPIED  $name" >> "$LOG"
        changes=$((changes + 1))
    fi
done

# Detect deleted files: files in DST but not in SRC
for f in "$DST"/*.mp3; do
    [ -f "$f" ] || continue
    name=$(basename "$f")
    src="$SRC/$name"
    if [ ! -f "$src" ]; then
        rm "$f"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] DELETED $name" >> "$LOG"
        changes=$((changes + 1))
    fi
done

# Log empty sync
if [ "$changes" -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] no changes" >> "$LOG"
fi

# Keep log manageable (last 200 lines)
tail -n 200 "$LOG" > "${LOG}.tmp" && mv "${LOG}.tmp" "$LOG"

exit 0
