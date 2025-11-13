# trackpad — use your phone as a trackpad for macOS

A tiny Go server + web UI that turns your phone into a wireless trackpad for your Mac. Start the server on your Mac, open the UI on your phone (LAN or via ngrok), paste the pairing token, and control the cursor with touch gestures.

Why this is useful

- Control your Mac remotely using your phone (presentations, couch browsing, media control).
- No native app required on the phone — use the browser UI in `public/index.html`.

Supported gestures

- Single-finger: cursor movement
- Two-finger: scroll (vertical/horizontal)
- Pinch: zoom

Requirements

- Go 1.20+
- `task` (Taskfile runner)
- `ngrok` (optional for remote access) or phone on same Wi‑Fi for LAN access

QuickStart (fastest way to try)

1) Create `.env` with defaults (run once):

```bash
cd /path/to/trackpad
task init-env
```

This creates `.env` with a generated `TOKEN`. Edit `.env` to set `NGROK_AUTHTOKEN` if you want to run `task ngrok`.

2) Start the server (terminal A):

```bash
# reads TOKEN from .env
task server

# or override inline
TOKEN=mysupersecret task server
```

3) Start ngrok (terminal B) — or use LAN:

```bash
# opens HTTP tunnel to port 8080 (or HTTP_PORT from .env)
task ngrok
```

- If ngrok reports an endpoint already online, copy the printed public URL and open it on your phone instead of starting a new tunnel.
  - You may copy the ngrok URL and generate a QR code for easy access from your phone. Go there to generate a QR code: https://www.qr-code-generator.com/
- If you don't want to use ngrok, connect over the local network:
- LAN: find your Mac IP and open it on the phone:

```bash
ipconfig getifaddr en0
# open http://<MAC_IP>:8080/ on your phone
```

4) On the phone

- Open the ngrok HTTPS URL (or the LAN URL).
- Paste the `TOKEN` value from `.env` (or the inline token you used) into the UI and tap Connect.
- Use touch gestures to control the Mac cursor.

Minimal troubleshooting

- No connection: ensure `task server` is running and listening:

```bash
lsof -iTCP -sTCP:LISTEN -n -P
```

- Phone can't reach Mac on LAN: bind server to all interfaces:

```bash
HOST=0.0.0.0 PORT=8080 task server
```

Security

- Keep `TOKEN` private. Do not commit `.env` to source control.
- ngrok exposes a public URL — stop the tunnel when you're done.

Files of interest

- `Taskfile.yaml` — tasks: `init-env`, `server`, `ngrok`.
- `main.go` — server & pairing logic.
- `public/index.html` — phone UI.

That's it — run `task init-env`, `task server`, `task ngrok`, open the URL on your phone, paste the token, and try gestures.
