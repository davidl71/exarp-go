# Research: Go implementations (GitHub) and CGO vs LSPF/C++

**Date:** 2026-04-05  
**Method:** `gh search repos` / `gh search code` (GitHub CLI).  
**Scope:** TN3270 / 3270 UI in Go, ISPF-adjacent tooling, and whether **daniel64/lspf** (or similar) can be **imported via cgo**.

---

## 1. Executive summary

- **Pure Go “ISPF Dialogue Manager”** equivalents on GitHub are **essentially absent**. The main open **ISPF DM + panel language** stack is **[daniel64/lspf](https://github.com/daniel64/lspf)** — **C++**, not Go.
- **Go activity clusters around TN3270 datastreams and emulators**, led by **[racingmars/go3270](https://github.com/racingmars/go3270)** (server library) and several **client / proxy / test** projects.
- **CGO** can call **C** (and often **C++** only via a **C wrapper**). **LSPF does not ship a stable C API** for embedding; treating it as a **library** would be a large fork, not an import. Realistic options: **stay on Go + go3270**, or **inter-process** integration (shell out / socket), not cgo into LSPF internals.

---

## 2. GitHub search results (representative)

Commands used (examples):

```bash
gh search repos "IBM 3270" --language=Go --sort=stars --limit=15
gh search repos ispf --limit=15
gh search repos go3270 --limit=10
gh search code --language=go "SELECT PANEL" --limit=5   # noisy; not ISPF-specific
```

### 2.1 TN3270 / 3270 in Go (server, client, tools)

| Repo | Role | Notes |
|------|------|--------|
| [racingmars/go3270](https://github.com/racingmars/go3270) | **3270 server library** (Go) | Used by exarp-go 3270 TUI; **primary reference** for `Screen` / `Tx` patterns. |
| [racingmars/proxy3270](https://github.com/racingmars/proxy3270) | 3270 proxy (Go) | Operational pattern: accept TN3270, rewrite, forward. |
| [mattheusv/go3270](https://github.com/mattheusv/go3270) | Interface to **x3270** | Client/scripting style, not a DM. |
| [rthorp/go3270](https://github.com/rthorp/go3270) | x3270 access | Older; check maintenance vs `mattheusv`. |
| [wuzuf/go-tn3270](https://github.com/wuzuf/go-tn3270) | TN3270 | Low-level / protocol-oriented (verify API). |
| [cyberdotgent/route3270](https://github.com/cyberdotgent/route3270) | Connection **router** | Pattern: demux sessions / backends. |
| [mflorence99/go-3270](https://github.com/mflorence99/go-3270) | **Browser** 3270 emulator | WASM/UI angle; different stack from terminal TN3270 servers. |
| [3270io/3270Web](https://github.com/3270io/3270Web) | Web 3270 + **recording** | Go; workflow/session capture. |
| [msradam/xk6-tn3270](https://github.com/msradam/xk6-tn3270) | **k6** TN3270 | Load-test pattern for green-screen apps. |
| [moshix/minesweeper](https://github.com/moshix/minesweeper) | Game for 3270 | Go + go3270; UX reference. |
| [moshix/ansitool-](https://github.com/moshix/ansitool-) | CP310 / screen tooling | Adjacent to presentation (not DM). |

### 2.2 “ISPF” name collisions

`gh search repos ispf` returns **daniel64/lspf** (C++ ISPF DM clone) plus many **mainframe REXX/ISPF dialog** repos (z/OS assets, not Go), **unrelated** projects (e.g. ML paper acronyms), and **oxidecomputer/ispf** (Internet packet format — **not** IBM ISPF).

**Conclusion:** refine queries with `language:Go` **and** keywords like `3270`, `tn3270`, `go3270`, `mainframe terminal`.

---

## 3. LSPF (daniel64/lspf) and Go

- **Repository:** [github.com/daniel64/lspf](https://github.com/daniel64/lspf)  
- **Language:** **C++** (ISPF Dialogue Manager services, panel language, tables, Lib/LMM-style APIs, ncurses-oriented display per README).  
- **Navigation pattern:** declarative **`plib/`** panels (e.g. `PMAINP01`) with **`)PROC` / `TRANS`** routing — see internal research on `plib/pmainp01` and `pmain0a.cpp`.

There is **no maintained Go port** of this stack in the same class on GitHub as of this search.

---

## 4. Can we import LSPF (or C++) using CGO?

### 4.1 What CGO does

- **CGO** links Go to **C** call conventions (`import "C"` + `#cgo` LDFLAGS/CFLAGS).
- **C++** classes are **not** exposed to Go directly; you need **`extern "C"`** wrappers and usually **`libfoo.so` / `libfoo.a`** with a **defined ABI**.

### 4.2 LSPF specifically

- LSPF is a **full application framework** (`pApplication`, variable pool, panel engine, services), not a small **embeddable C library**.
- Upstream does not advertise **“embed in another process via libc API”**.
- Wrapping it in cgo would require:
  - exposing a **minimal C API** (init, run one transaction, pass buffers),
  - **lifetime/threading** alignment (Boost threads, ncurses),
  - **distribution** (ship `.so`, libc++, panel/message libraries, profile paths),
  - **security** and **supply-chain** review for a binary that runs end-user code paths.

**Practical verdict:** **not a good fit** for exarp-go’s 3270 TUI unless the project **forks LSPF** with an explicit **embedded mode** and C shim. **Prefer Go-native** UI on `go3270` and **borrow LSPF ideas** (TRANS tables, panel metadata, PNTS) as data structures in Go.

### 4.3 When CGO *is* reasonable

- Wrapping **small C libraries** (e.g. codecs, parsers) with a clear **C** header.
- Accepting **cross-compile pain** (macOS ↔ Linux, static binaries) and **no pure `CGO_ENABLED=0`** builds for that code path.

### 4.4 Alternatives to cgo for “use LSPF”

1. **Process boundary:** run LSPF as its own terminal app; integrate at **file/MCP/task** level only (no shared TN3270 session unless proxying).
2. **TN3270 proxy / door:** connect users via **3270BBS-style** proxy between two servers (see `docs/3270_TUI_IMPLEMENTATION.md`).
3. **Pattern copy only:** implement **`TRANS`-like** routing tables and **panel defs** in Go (YAML/JSON or embedded structs), not the full LSPF executor.

---

## 5. Recommendations for exarp-go

1. **Stay on [racingmars/go3270](https://github.com/racingmars/go3270)** for presentation; continue to mine **LSPF panel semantics** as **documentation**, not a runtime dependency.
2. If routing grows, add a **declarative command/option table** (LSPF `TRANS` analogue) with tests; optional **`gh`** periodic repo search for new `go3270` consumers.
3. **Avoid cgo → LSPF** unless there is a **maintained C embedding API** and a product decision to **bundle** C++ runtime artifacts.
4. For reproducibility, re-run:

   ```bash
   gh search repos go3270 --sort=stars --limit=25
   gh search repos tn3270 --language=Go --sort=updated --limit=25
   ```

---

## 6. Related docs in this repo

- [TUI3270_LSPF_IMPLEMENTATION_RESEARCH.md](./TUI3270_LSPF_IMPLEMENTATION_RESEARCH.md)
- [3270_TUI_IMPLEMENTATION.md](./3270_TUI_IMPLEMENTATION.md)
- [ISPF_PATTERNS_RESEARCH.md](./ISPF_PATTERNS_RESEARCH.md)

---

## 7. References (external)

- Go wiki — **cgo**: [https://go.dev/wiki/cgo](https://go.dev/wiki/cgo)  
- **go3270:** [https://github.com/racingmars/go3270](https://github.com/racingmars/go3270)  
- **lspf:** [https://github.com/daniel64/lspf](https://github.com/daniel64/lspf)  
