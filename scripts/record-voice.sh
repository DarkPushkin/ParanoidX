#!/bin/bash
# Record voice message and save to simplex-node chat files
# Usage: ./record-voice.sh [duration_seconds] [output_name]
DURATION=${1:-5}
NAME=${2:-voice-$(date +%s)}
OUTDIR="${HOME}/.local/share/simplex-node/files"
mkdir -p "$OUTDIR"
OUTFILE="${OUTDIR}/${NAME}.wav"
echo "Recording ${DURATION}s to ${OUTFILE}..."
arecord -f cd -t wav -d "$DURATION" "$OUTFILE" 2>/dev/null || \
  sox -n "$OUTFILE" synth "$DURATION" sine 440 vol 0.3
echo "Saved: ${OUTFILE} ($(du -h "$OUTFILE" | cut -f1))"