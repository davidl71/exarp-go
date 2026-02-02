# Generating a TUI Demo with agent-tui

Use [agent-tui](https://github.com/pproenca/agent-tui) to drive the exarp-go Bubble Tea TUI (`exarp-go tui`) from a script, capture screenshots, and produce demo assets (text snapshots or images for docs/GIFs).

## Prerequisites

- **agent-tui** installed (CLI + daemon).
- **exarp-go** built: `make build` (binary at `bin/exarp-go`).
- Run from the **repository root** so `.todo2` (and DB) are found.

### Install agent-tui

```bash
# Quick install (binary to ~/.local/bin)
curl -fsSL https://raw.githubusercontent.com/pproenca/agent-tui/master/install.sh | bash

# Or via package managers
npm install -g agent-tui   # or pnpm/bun
```

## Quick manual demo

From the repo root:

```bash
# 1. Start daemon (if not already running)
agent-tui daemon start

# 2. Run exarp-go TUI inside agent-tui
agent-tui run ./bin/exarp-go tui

# 3. Wait for the TUI to show (e.g. status text or "Todo")
agent-tui wait "Todo" -t 5000 || true

# 4. Capture screenshots (repeat after keypresses for a sequence)
agent-tui screenshot                    # human-readable
agent-tui screenshot --json > demo-step1.json

agent-tui press ArrowDown ArrowDown     # navigate
agent-tui screenshot --json > demo-step2.json

agent-tui press Enter                  # open task (if any)
agent-tui screenshot --json > demo-step3.json

# 5. End session
agent-tui kill
```

Screenshot JSON contains a `screenshot` field (often terminal text or base64). Use it for docs or to generate images/GIFs with your own tooling.

## Scripted demo (recommended)

From the repo root, one command does everything (build + run demo):

```bash
make demo-tui
```

This target builds exarp-go if needed, then runs the demo script. Alternatively run the script directly:

```bash
./scripts/demo-tui-agent-tui.sh
```

The script (if `agent-tui` is on `PATH`):

1. Starts the agent-tui daemon if needed.
2. Runs `./bin/exarp-go tui` under agent-tui.
3. Waits for the TUI to be ready, sends a short key sequence, and saves screenshot outputs under `docs/demo/`.
4. Kills the session.

Outputs are written to `docs/demo/` (e.g. `step1.txt`, `step2.txt`, or JSON if you change the script). You can then:

- Paste text snapshots into docs.
- Decode JSON screenshot fields and convert to PNG/GIF (e.g. with a small script or ImageMagick if the format is supported).

## 3270 TUI demo (optional)

To demo the **3270 TUI** (`exarp-go tui3270`):

1. Start the 3270 server: `exarp-go tui3270 --daemon` (or in another terminal).
2. Use agent-tui to run a 3270 client (e.g. **x3270**) that connects to `localhost:3270`:

   ```bash
   agent-tui run x3270 -model 2 localhost
   ```

3. After connection, send PF keys / navigation (e.g. `agent-tui press F7`, `agent-tui press Tab`) and take screenshots the same way.

This requires a 3270 client installed (e.g. `x3270` on Linux/macOS).

## Turning screenshots into a GIF or video

- **Text-only output**: Use the saved `.txt` or JSON `screenshot` content in markdown (e.g. fenced code blocks) for a step-by-step doc.
- **Image output**: If agent-tui (or your pipeline) produces images, use ImageMagick or ffmpeg:

  ```bash
  # Example: images step1.png, step2.png, ... → demo.gif
  convert -delay 150 -loop 0 docs/demo/step*.png docs/demo/demo.gif
  ```

- **Live recording**: Run the TUI in a terminal and record the session with **asciinema** or **vhs** (e.g. `vhs docs/demo/script.tape`) for a reproducible demo; agent-tui is not required for that.

## References

- [agent-tui](https://github.com/pproenca/agent-tui) – TUI automation for AI agents (PTY, screenshot, input, wait).
- [3270 TUI implementation](./3270_TUI_IMPLEMENTATION.md) – exarp-go 3270 server and client usage.
- exarp-go Bubble Tea TUI: `internal/cli/tui.go`; entrypoint: `exarp-go tui`.
