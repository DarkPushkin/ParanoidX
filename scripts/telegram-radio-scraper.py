#!/usr/bin/env python3
"""Telegram Radio Scraper — парсит каналы Telegram, скачивает музыку в радио-плейлист simplex-node."""
import json, os, re, subprocess, sys, time, logging, tempfile, hashlib, argparse
from pathlib import Path
from urllib.request import urlopen, Request, build_opener, ProxyHandler
from urllib.parse import urlparse

os.environ.pop('http_proxy', None); os.environ.pop('https_proxy', None)
os.environ.pop('HTTP_PROXY', None); os.environ.pop('HTTPS_PROXY', None)
os.environ.pop('ALL_PROXY', None); os.environ.pop('all_proxy', None)

logging.basicConfig(level=logging.INFO, format='%(asctime)s [%(levelname)s] %(message)s')
log = logging.getLogger('tg-radio')

DATA_DIR = os.environ.get('DATA_DIR', os.path.expanduser('~/.local/share/simplex-node'))
RADIO_DIR = os.path.join(DATA_DIR, 'radio')
STATE_FILE = os.path.join(DATA_DIR, 'radio', 'tg_scraper_state.json')
YTDLP = os.environ.get('YTDLP', 'python3 -m yt_dlp')
BOT_TOKEN = ''

def load_state():
    try:
        with open(STATE_FILE) as f:
            return json.load(f)
    except: return {'channels': {}, 'downloaded': []}

def save_state(state):
    os.makedirs(os.path.dirname(STATE_FILE), exist_ok=True)
    with open(STATE_FILE, 'w') as f:
        json.dump(state, f, indent=2)

def is_already_downloaded(url, state):
    return any(d.get('url') == url for d in state.get('downloaded', []))

def mark_downloaded(url, title, filepath, state):
    if 'downloaded' not in state:
        state['downloaded'] = []
    state['downloaded'].append({
        'url': url, 'title': title, 'file': filepath, 'date': time.strftime('%Y-%m-%dT%H:%M:%S')
    })
    save_state(state)

def extract_urls(text):
    urls = re.findall(r'(https?://[^\s<>"\']+)', text)
    return [u.rstrip('.,;:!?)') for u in urls]

def is_youtube_url(url):
    domains = ['youtube.com', 'youtu.be', 'm.youtube.com', 'www.youtube.com']
    return any(d in url.lower() for d in domains)

def download_ytdlp(url, output_dir):
    try:
        safe_name = hashlib.sha256(url.encode()).hexdigest()[:16]
        output = os.path.join(output_dir, f'{safe_name}.%(ext)s')
        env = os.environ.copy()
        for k in ['http_proxy','https_proxy','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','all_proxy']:
            env.pop(k, None)
        cmd = f'{YTDLP} -x --audio-format mp3 --audio-quality 0 -o "{output}" --proxy "" "{url}"'
        log.info(f'Downloading: {url}')
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=300, env=env)
        if result.returncode != 0:
            log.warning(f'yt-dlp failed: {result.stderr[:200]}')
            return None
        for f in os.listdir(output_dir):
            if f.startswith(safe_name):
                return os.path.join(output_dir, f)
    except subprocess.TimeoutExpired:
        log.warning(f'Timeout downloading {url}')
    except Exception as e:
        log.warning(f'Download error: {e}')
    return None

def download_audio_file(url, output_dir):
    try:
        parsed = urlparse(url)
        name = os.path.basename(parsed.path) or hashlib.sha256(url.encode()).hexdigest()[:16]
        if not name.endswith(('.mp3', '.ogg', '.wav', '.flac', '.m4a', '.aac')):
            name += '.mp3'
        path = os.path.join(output_dir, name)
        if os.path.exists(path):
            return path
        opener = build_opener(ProxyHandler({}))
        req = Request(url, headers={'User-Agent': 'Mozilla/5.0'})
        with opener.open(req, timeout=60) as resp:
            data = resp.read()
        with open(path, 'wb') as f:
            f.write(data)
        return path
    except Exception as e:
        log.warning(f'Direct download error: {e}')
    return None

def scrape_telegram_channel(channel, limit=20):
    """Scrape messages from t.me/s/<channel>"""
    messages = []
    url = f'https://t.me/s/{channel}?before={limit}'
    try:
        opener = build_opener(ProxyHandler({}))
        req = Request(url, headers={'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36'})
        with opener.open(req, timeout=30) as resp:
            html = resp.read().decode('utf-8', errors='replace')
        texts = re.findall(r'<div class="tgme_widget_message_text[^>]*>(.*?)</div>', html, re.DOTALL)
        links = re.findall(r'<a[^>]+href="(https?://[^"]+)"', html)
        for t in texts:
            clean = re.sub(r'<[^>]+>', '', t).strip()
            if clean:
                messages.append({'text': clean})
        for l in links:
            if any(u in l for u in ['youtube.com', 'youtu.be', 'soundcloud.com', 'bandcamp.com']):
                messages.append({'url': l})
        return messages, links
    except Exception as e:
        log.warning(f'Scrape failed for {channel}: {e}')
    return [], []

def process_channel(channel, state):
    log.info(f'Scraping @{channel}...')
    messages, urls = scrape_telegram_channel(channel)
    if not messages and not urls:
        log.info(f'No messages from @{channel} (needs bot admin)')
        return

    new_count = 0
    for msg in messages:
        text = msg.get('text', '')
        urls_in_msg = extract_urls(text)
        for u in urls_in_msg:
            if is_already_downloaded(u, state):
                continue
            if is_youtube_url(u):
                path = download_ytdlp(u, RADIO_DIR)
                if path:
                    mark_downloaded(u, text[:50], path, state)
                    new_count += 1
            else:
                path = download_audio_file(u, RADIO_DIR)
                if path:
                    mark_downloaded(u, text[:50], path, state)
                    new_count += 1

    for u in urls:
        if u in [d.get('url') for d in state.get('downloaded', [])]:
            continue
        if is_youtube_url(u):
            path = download_ytdlp(u, RADIO_DIR)
            if path:
                mark_downloaded(u, '', path, state)
                new_count += 1

    if new_count > 0:
        log.info(f'Downloaded {new_count} new tracks from @{channel}')
        run_sync()

def run_sync():
    log.info('Running radio sync...')
    subprocess.run(['curl', '-s', '-X', 'POST', 'http://127.0.0.1:8080/api/admin/disk-cleanup'], capture_output=True)

def poll_loop(channels, interval=300):
    state = load_state()
    log.info(f'Starting poll loop for channels: {channels} (interval={interval}s)')
    while True:
        for ch in channels:
            process_channel(ch.strip(), state)
        time.sleep(interval)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Telegram Radio Scraper')
    parser.add_argument('--channels', nargs='+', default=['RadioArmageddonFM'],
                        help='Telegram channels to scrape')
    parser.add_argument('--interval', type=int, default=300,
                        help='Poll interval in seconds (default: 300)')
    parser.add_argument('--once', action='store_true',
                        help='Run once and exit')
    args = parser.parse_args()
    if args.once:
        state = load_state()
        for ch in args.channels:
            process_channel(ch, state)
    else:
        poll_loop(args.channels, args.interval)
