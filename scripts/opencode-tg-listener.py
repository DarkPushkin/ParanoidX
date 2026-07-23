#!/usr/bin/env python3
"""
opencode-tg-listener.py — Telegram ↔ opencode bridge.

Polls Telegram for messages, forwards them directly to the shared opencode
session (this conversation). The AI responds and the reply goes back to Telegram.
"""

import json
import os
import re
import subprocess
import sys
import threading
import time
import urllib.error
import urllib.request

# ── Config ──────────────────────────────────────────────────────────────
TOKEN_FILE = os.path.expanduser("~/.config/opencode-tg-bot.token")
DATA_DIR = os.path.expanduser("~/.local/share/opencode-tg")
SESSION_FILE = os.path.join(DATA_DIR, "sessions.json")
OFFSET_FILE = os.path.join(DATA_DIR, "offset.txt")
WORK_DIR = os.path.expanduser("~/.opencode")

POLL_INTERVAL = 2
HEALTH_TIMEOUT = 30
REQUEST_TIMEOUT = 600
TG_TIMEOUT = 15

# Bypass proxy for localhost — unset proxy env vars
for var in ['HTTP_PROXY', 'HTTPS_PROXY', 'http_proxy', 'https_proxy', 'ALL_PROXY', 'all_proxy']:
    os.environ.pop(var, None)

# ── Logging ─────────────────────────────────────────────────────────────

def log(msg):
    t = time.strftime("%H:%M:%S")
    print(f"[{t}] {msg}", flush=True)

# ── HTTP helpers ────────────────────────────────────────────────────────

def http(url, method="GET", body=None, timeout=30):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    if data:
        req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            return resp.status, json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        txt = e.read().decode("utf-8", errors="replace")[:500]
        return e.code, txt
    except Exception as e:
        return -1, str(e)

# ── Token ───────────────────────────────────────────────────────────────

def load_token():
    try:
        with open(TOKEN_FILE) as f:
            tok = f.read().strip()
            if not tok:
                log(f"FATAL: token file {TOKEN_FILE} is empty")
                sys.exit(1)
            return tok
    except Exception as e:
        log(f"FATAL: cannot read {TOKEN_FILE}: {e}")
        sys.exit(1)

# ── Opencode server ─────────────────────────────────────────────────────

def find_existing_server():
    for port in range(1024, 65536):
        st, _ = http(f"http://127.0.0.1:{port}/global/health", timeout=1)
        if st == 200:
            log(f"Found existing opencode server on port {port}")
            return port
    return None

def start_opencode_server():
    existing = find_existing_server()
    if existing:
        return existing, None

    log("Starting opencode serve...")
    proc = subprocess.Popen(
        ["opencode", "serve", "--port", "0", "--hostname", "127.0.0.1"],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        cwd=WORK_DIR,
        text=True,
    )

    port = None
    t0 = time.time()

    while time.time() - t0 < HEALTH_TIMEOUT:
        line = proc.stdout.readline()
        if not line:
            if proc.poll() is not None:
                log(f"FATAL: opencode exited early (code={proc.returncode})")
                sys.exit(1)
            time.sleep(0.1)
            continue

        line = line.rstrip()
        log(f"[opencode] {line}")

        m = re.search(r"listening on https?://[^:]+:(\d+)", line, re.IGNORECASE)
        if m:
            port = int(m.group(1))
            log(f"Detected port {port}")
            break

    if port is None:
        log("FATAL: could not detect opencode port from output")
        proc.kill()
        sys.exit(1)

    t0 = time.time()
    while time.time() - t0 < HEALTH_TIMEOUT:
        st, data = http(f"http://127.0.0.1:{port}/global/health", timeout=2)
        if st == 200:
            log(f"Opencode server healthy (port {port})")
            return port, proc
        time.sleep(0.5)

    log("FATAL: opencode health check failed")
    proc.kill()
    sys.exit(1)

# ── Session management (thread-safe) ────────────────────────────────────

_sessions_lock = threading.Lock()

def load_sessions():
    try:
        with open(SESSION_FILE) as f:
            return json.load(f)
    except (FileNotFoundError, json.JSONDecodeError):
        return {}

def save_sessions(sessions):
    os.makedirs(DATA_DIR, exist_ok=True)
    with open(SESSION_FILE, "w") as f:
        json.dump(sessions, f, indent=2)

def get_session(chat_key):
    with _sessions_lock:
        return load_sessions().get(chat_key)

def set_session(chat_key, session_id):
    with _sessions_lock:
        s = load_sessions()
        s[chat_key] = session_id
        save_sessions(s)

def delete_session(chat_key):
    with _sessions_lock:
        s = load_sessions()
        s.pop(chat_key, None)
        save_sessions(s)

def create_session(base_url):
    st, data = http(f"{base_url}/session", method="POST", body={}, timeout=15)
    if st != 200:
        log(f"Create session failed: {data}")
        return None
    sid = data.get("id") if isinstance(data, dict) else None
    if sid:
        log(f"Created session {sid}")
    return sid

# ── Opencode message ────────────────────────────────────────────────────

def handle_llm_message(token, base_url, chat_id, session_id, text):
    body = {"parts": [{"type": "text", "text": text}]}
    log(f"Sending to opencode session {session_id} (chat {chat_id})...")
    st, data = http(
        f"{base_url}/session/{session_id}/message",
        method="POST", body=body, timeout=REQUEST_TIMEOUT,
    )
    if st != 200:
        err = data if isinstance(data, str) else json.dumps(data)
        reply = f"\u26a0\ufe0f opencode error (status {st}): {err[:300]}"
    else:
        parts = data.get("parts", []) if isinstance(data, dict) else []
        texts = []
        for p in parts:
            if isinstance(p, dict) and p.get("type") == "text":
                t = (p.get("text") or "").strip()
                if t:
                    texts.append(t)
        reply = "\n\n".join(texts) if texts else "\u2705 Done"
    send_reply(token, chat_id, reply)

# ── Telegram helpers ────────────────────────────────────────────────────

def tg_call(token, method, body=None, timeout=TG_TIMEOUT):
    url = f"https://api.telegram.org/bot{token}/{method}"
    return http(url, method="POST" if body else "GET", body=body, timeout=timeout)

def send_reply(token, chat_id, text):
    if not text:
        return
    if len(text) > 4096:
        text = text[:4093] + "..."
    st, data = tg_call(token, "sendMessage", {
        "chat_id": chat_id,
        "text": text,
    })
    if st == 200:
        mid = ""
        if isinstance(data, dict):
            mid = data.get("result", {}).get("message_id", "") if isinstance(data.get("result"), dict) else ""
        log(f"Sent to chat {chat_id} (msg_id={mid}): {text[:60]}...")
    else:
        log(f"sendMessage error (chat {chat_id}): {data}")

def get_updates(token, offset):
    params = f"?timeout=10&allowed_updates=[\"message\"]"
    if offset:
        params += f"&offset={offset}"
    url = f"https://api.telegram.org/bot{token}/getUpdates{params}"
    st, data = http(url, timeout=TG_TIMEOUT)
    if st != 200:
        return []
    return data.get("result", [])

# ── Main loop ───────────────────────────────────────────────────────────

def main():
    os.makedirs(DATA_DIR, exist_ok=True)
    token = load_token()

    port, server_proc = start_opencode_server()
    base_url = f"http://127.0.0.1:{port}"

    log(f"Loaded session(s) from {SESSION_FILE}")

    offset = None
    try:
        with open(OFFSET_FILE) as f:
            offset = int(f.read().strip())
            log(f"Resumed poll at offset {offset}")
    except (FileNotFoundError, ValueError):
        updates = get_updates(token, None)
        if updates:
            offset = max(u["update_id"] for u in updates if "update_id" in u)
            log(f"Bootstrapped offset to {offset}")

    log("Starting poll loop")

    while True:
        try:
            if server_proc is not None and server_proc.poll() is not None:
                log("Server died, restarting...")
                port, server_proc = start_opencode_server()
                base_url = f"http://127.0.0.1:{port}"

            updates = get_updates(token, offset)

            for update in updates:
                up_id = update.get("update_id")
                if up_id is not None:
                    offset = up_id + 1

                msg = update.get("message", {})
                chat = msg.get("chat", {})
                chat_id = chat.get("id")
                text = (msg.get("text") or "").strip()
                caption = (msg.get("caption") or "").strip()

                # Handle photo messages — use caption as text
                if not text and caption:
                    text = caption
                    photo = msg.get("photo", [])
                    if photo:
                        log(f"Received photo with caption from chat {chat_id}")
                    # We store the caption but don't download the photo yet

                if not chat_id or not text:
                    continue

                log(f"Received from chat {chat_id}: {text[:80]}")
                cmd = text.lower()

                # ── Commands (processed inline, instant reply) ──
                if cmd in ("/start", "/new", "/reset"):
                    delete_session(str(chat_id))
                    send_reply(token, chat_id,
                        "\U0001f504 Session reset! Send any message to start fresh.")
                    continue

                if cmd == "/help":
                    send_reply(token, chat_id,
                        "\U0001f916 OpenCode Telegram Bot\n\n"
                        "Send any message and I'll respond via opencode.\n\n"
                        "Commands:\n"
                        "/new \u2014 reset conversation\n"
                        "/help \u2014 this message\n"
                        "/status \u2014 check connection")
                    continue

                if cmd == "/status":
                    st, _ = http(f"{base_url}/global/health", timeout=5)
                    status = "\u2705 healthy" if st == 200 else f"\u26a0\ufe0f error ({st})"
                    send_reply(token, chat_id,
                        f"opencode server {status} (port {port})")
                    continue

                # ── LLM message (dispatched to background thread) ──
                chat_key = str(chat_id)
                session_id = get_session(chat_key)

                if not session_id:
                    log(f"Creating session for chat {chat_id}")
                    session_id = create_session(base_url)
                    if not session_id:
                        send_reply(token, chat_id,
                            "\u26a0\ufe0f Failed to create session, try again later.")
                        continue
                    set_session(chat_key, session_id)

                send_reply(token, chat_id, "\u23f3 Processing...")
                t = threading.Thread(
                    target=handle_llm_message,
                    args=(token, base_url, chat_id, session_id, text),
                    daemon=True,
                )
                t.start()

            if offset:
                with open(OFFSET_FILE, "w") as f:
                    f.write(str(offset))

        except KeyboardInterrupt:
            log("Shutting down...")
            if server_proc:
                server_proc.terminate()
            break
        except Exception as e:
            log(f"Error: {e}")

        time.sleep(POLL_INTERVAL)

if __name__ == "__main__":
    main()
