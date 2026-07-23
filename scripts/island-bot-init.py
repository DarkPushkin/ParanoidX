#!/usr/bin/env python3
"""
island-bot-init.py
Uses pexpect (available via python3-pexpect) to reliably initialize the "Island Royal Services" bot profile
using the official simplex-chat CLI binary.

This makes profile creation + address extraction fully automatic for the SimpleX transport.

Steps:
- Spawn CLI with --create-bot-display-name and --create-bot-allow-files + our SMP server.
- Wait for ready.
- Send /set bot commands with the full Island grimoire.
- Send /address and capture the simplex contact link.
- Print the link (bash captures it to write the full contact file with instructions + magic).

Then the main setup starts the CLI with -p for the WS gateway, and our Go bridge connects.

This solves the headless terminal issues and makes the "Soul of the Treasure Island" accessible via one contact in any stock SimpleX app.

Run via island-bot-setup.sh
"""

import pexpect
import sys
import os
import re
import time

def main():
    if len(sys.argv) < 5:
        print("Usage: island-bot-init.py <cli_bin> <db_dir> <smp_addr> <contact_file>")
        sys.exit(1)

    cli_bin = sys.argv[1]
    db = sys.argv[2]
    smp = sys.argv[3]
    contact_file = sys.argv[4]

    # Clean previous partial if wanted, but usually keep for idempotency
    # os.system(f"rm -rf {db}/* 2>/dev/null || true")

    cmd = f"{cli_bin} -d {db} --create-bot-display-name 'Остров — Королевские Сервисы' --create-bot-allow-files -s {smp} --yes-migrate --socks-proxy 127.0.0.1:9050 -p 5231"
    print(f"[init] Spawning: {cmd}", file=sys.stderr)

    child = pexpect.spawn(cmd, timeout=30, encoding='utf-8', echo=False)
    child.logfile_read = sys.stderr  # for debug in setup logs

    link = None
    try:
        # Wait for initial startup / prompt
        i = child.expect([
            r"type /help or /h for usage info",
            r"private message routing mode",
            r"db: ",
            pexpect.TIMEOUT,
            pexpect.EOF
        ], timeout=20)

        print(f"[init] Initial prompt seen (index {i})", file=sys.stderr)
        time.sleep(1)

        # Set as bot and commands (the grimoire)
        child.sendline("/set bot on")
        time.sleep(0.5)

        # Full menu - use single line for simplicity
        bot_cmds = "'/help':/help,'Кошелек':/wallet,'Радио':/radio,'Хранилище':/vault,'Маркет':/market,'Токенизатор':/tokenize,'ID/Паспорт':/id,'Каналы':/channels"
        child.sendline(f"/set bot commands {bot_cmds}")
        time.sleep(1)

        # Get the contact address
        child.sendline("/address")
        time.sleep(1.5)

        # Capture output - look for the link
        output = child.before + (child.after or "")
        # Typical output contains "simplex:/contact#..." or similar permanent address
        match = re.search(r"(simplex:/contact#[^\s]+|simplex:/[a-z0-9#?=&]+)", output, re.IGNORECASE)
        if match:
            link = match.group(1)
            print(f"[init] Captured link: {link}", file=sys.stderr)
        else:
            # Fallback: try to get more output or last lines
            child.sendline("/address")
            time.sleep(1)
            output2 = child.before + (child.after or "")
            match2 = re.search(r"(simplex:/[^\s]+)", output2, re.IGNORECASE)
            if match2:
                link = match2.group(1)
                print(f"[init] Captured link (retry): {link}", file=sys.stderr)

        if link:
            # Write the full beautiful contact file here? Or let bash do it.
            # For now, just print the pure link so bash can embed it in the full text.
            print(link)  # stdout for capture by setup.sh
        else:
            print("[init] Could not extract link automatically. Manual step needed.", file=sys.stderr)
            print("PLACEHOLDER")  # signal to bash

        # Graceful exit
        child.sendline("/quit")
        child.expect(pexpect.EOF, timeout=5)

    except pexpect.TIMEOUT as e:
        print(f"[init] Timeout during init: {e}", file=sys.stderr)
        print("PLACEHOLDER")
    except Exception as e:
        print(f"[init] Error: {e}", file=sys.stderr)
        print("PLACEHOLDER")
    finally:
        try:
            child.close()
        except:
            pass

if __name__ == "__main__":
    main()