#!/bin/bash
set -euo pipefail
# tg-radio-sync.sh — Скачивает музыку с Telegram-канала @RadioArmageddonFM в радио-плейлист
# Использует python-telegram-bot и yt-dlp

# Очищаем прокси — Tor мешает прямым запросам
for var in http_proxy https_proxy HTTP_PROXY HTTPS_PROXY ALL_PROXY all_proxy; do
  unset "$var"
done

DATA_DIR="${DATA_DIR:-$HOME/.local/share/simplex-node}"
LOG="$DATA_DIR/logs/tg-radio-sync.log"
mkdir -p "$DATA_DIR/logs"
exec >> "$LOG" 2>&1

echo "===== TG Radio Sync $(date) ====="

TOKEN_FILE="$HOME/.config/opencode-tg-bot.token"
CHANNEL="${1:-@RadioArmageddonFM}"
LIMIT="${2:-20}"

if [ ! -f "$TOKEN_FILE" ]; then
  echo "No bot token found at $TOKEN_FILE"
  exit 1
fi

TOKEN=$(cat "$TOKEN_FILE")

# Get recent messages from the channel (bot must be admin)
echo "Fetching updates from $CHANNEL..."

# Use the Bot API to get chat messages
# Since bot needs to be admin to read channel messages,
# we get the last message via getChat and try getUpdates

RESP=$(curl -s --max-time 15 "https://api.telegram.org/bot${TOKEN}/getUpdates?timeout=5&allowed_updates=[\"message\",\"channel_post\"]")

echo "Got updates response: $(echo "$RESP" | head -c 200)"

# Also try getting chat administrators
CHAT_INFO=$(http_proxy="" https_proxy="" HTTP_PROXY="" HTTPS_PROXY="" ALL_PROXY="" \
  curl -s --max-time 10 "https://api.telegram.org/bot${TOKEN}/getChat?chat_id=${CHANNEL}")

echo "Chat info: $(echo "$CHAT_INFO" | head -c 200)"

# Check if bot can read messages from the channel
HAS_ACCESS=$(echo "$CHAT_INFO" | python3 -c "import sys,json; d=json.load(sys.stdin); print('ok' if d.get('ok') else 'no')" 2>/dev/null)

if [ "$HAS_ACCESS" != "ok" ]; then
  echo "❌ Bot can't access $CHANNEL. Add bot as admin to the channel."
  echo "Send this link to the channel admin: https://t.me/torquemada878_bot?startchannel=admin"
  exit 0
fi

# Try to get the last N messages from the channel
# Bot API doesn't support reading channel history directly unless bot is admin
# Alternative: use getUpdates if bot has been added as admin recently

echo "Updates with channel posts:"
echo "$RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
if not d.get('ok'):
    print(f'API error: {d}')
    sys.exit(0)
results = d.get('result', [])
print(f'Total updates: {len(results)}')
for u in results[-10:]:
    msg = u.get('channel_post') or u.get('message')
    if not msg:
        continue
    chat = msg.get('chat', {})
    if chat.get('username','') != '${CHANNEL#@}':
        continue
    msg_id = msg.get('message_id')
    date = msg.get('date')
    text = msg.get('text', '') or msg.get('caption', '') or ''
    audio = msg.get('audio')
    video = msg.get('video')
    document = msg.get('document')
    print(f'  msg#{msg_id} date={date} text={text[:100]}')
    if audio: print(f'    AUDIO: {audio.get(\"title\",\"\")} - {audio.get(\"performer\",\"\")} ({audio.get(\"file_id\",\"\")})')
    if video: print(f'    VIDEO: {video.get(\"file_id\",\"\")}')" 2>&1 || true

# If we have channel posts, download audio files
echo "$RESP" | python3 -c "
import sys, json, os, subprocess, hashlib

d = json.load(sys.stdin)
if not d.get('ok'):
    sys.exit(0)

TOKEN = open('$TOKEN_FILE').read().strip()
RADIO_DIR = os.path.expanduser('$DATA_DIR/radio')
STATE_FILE = os.path.join(RADIO_DIR, 'tg_scraper_state.json')
os.makedirs(RADIO_DIR, exist_ok=True)

def load_state():
    try:
        with open(STATE_FILE) as f: return json.load(f)
    except: return {'downloaded': []}

def save_state(state):
    with open(STATE_FILE, 'w') as f: json.dump(state, f, indent=2)

state = load_state()
downloaded_ids = set(d.get('msg_id') for d in state.get('downloaded', []))

results = d.get('result', [])
new_count = 0

for u in results:
    msg = u.get('channel_post') or u.get('message')
    if not msg: continue
    chat = msg.get('chat', {})
    if chat.get('username','') != 'RadioArmageddonFM': continue

    msg_id = msg.get('message_id')
    if msg_id in downloaded_ids: continue

    audio = msg.get('audio')
    video = msg.get('video')
    document = msg.get('document')
    text = msg.get('text', '') or msg.get('caption', '') or ''

    # Check for YouTube links in text
    import re
    yt_urls = re.findall(r'(https?://(?:www\.)?(?:youtube\.com|youtu\.be)/[^\s<>\"\']+)', text)
    for url in yt_urls:
        url = url.rstrip('.,;:!?)')
        safe = hashlib.sha256(url.encode()).hexdigest()[:16]
        output = os.path.join(RADIO_DIR, f'{safe}.%(ext)s')
        env = os.environ.copy()
        for k in ['http_proxy','https_proxy','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','all_proxy']:
            env.pop(k, None)
        cmd = f'python3 -m yt_dlp -x --audio-format mp3 --audio-quality 0 -o \"{output}\" --proxy \"\" \"{url}\"'
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=300, env=env)
        if result.returncode == 0:
            print(f'  Downloaded: {url}')
            state.setdefault('downloaded', []).append({'msg_id': msg_id, 'url': url, 'title': text[:50], 'date': str(msg.get('date'))})
            new_count += 1
        else:
            print(f'  Failed: {result.stderr[:100]}')

    # Download audio file if present
    if audio:
        file_id = audio.get('file_id')
        fname = f\"{audio.get('performer','unknown')} - {audio.get('title','unknown')}.mp3\".replace('/', '_')
        fpath = os.path.join(RADIO_DIR, fname)
        if not os.path.exists(fpath):
            print(f'  Audio file: {fname}')
            state.setdefault('downloaded', []).append({'msg_id': msg_id, 'file_id': file_id, 'title': fname, 'date': str(msg.get('date'))})
            new_count += 1

    # Download video/audio file if present
    media = audio or video or document
    if media and not audio:
        file_id = media.get('file_id')
        if file_id and msg_id not in downloaded_ids:
            print(f'  Media file_id: {file_id}')
            state.setdefault('downloaded', []).append({'msg_id': msg_id, 'file_id': file_id, 'title': text[:50], 'date': str(msg.get('date'))})
            new_count += 1

if new_count > 0:
    save_state(state)
    print(f'New tracks: {new_count}')
else:
    print('No new tracks')
" 2>&1 || true

echo "===== Sync Complete ====="
