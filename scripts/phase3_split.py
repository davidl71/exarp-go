#!/usr/bin/env python3
"""scripts/phase3_split.py — Phase 3 large-file splits with Claude-friendly annotations.

Splits 11 files (each >750 lines) into 2–3 parts, and adds navigation aids to
every output file:
  • Section separator comment before each top-level declaration
  • File-level Contents TOC listing all functions/types
  • "See also:" hint naming sibling split files

Runs all 11 file groups in parallel via ThreadPoolExecutor(max_workers=6).

Usage:
    cd /path/to/exarp-go
    python3 scripts/phase3_split.py
    make b QUIET=1
"""
import re
import os
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed

T = "internal/tools"
SEP_W = 79  # separator line width (chars)

# ─── File I/O and import helpers ──────────────────────────────────────────────

def rlines(path):
    with open(path) as f:
        return f.readlines()

def header_end(ls):
    """Return 0-based index of first body line (after the import block)."""
    in_imp = False
    for i, l in enumerate(ls):
        s = l.strip()
        if re.match(r'^import\s*\(', s):
            in_imp = True
        elif in_imp and s == ")":
            return i + 1
        elif s.startswith('import "'):
            return i + 1
    for i, l in enumerate(ls):
        if l.strip().startswith("package "):
            return i + 1
    return 0

def parse_imps(ls):
    """Return list of (alias_or_None, import_path, raw_line)."""
    result, in_imp = [], False
    for l in ls:
        s = l.strip()
        if re.match(r'^import\s*\(', s):
            in_imp = True; continue
        if in_imp:
            if s == ")": break
            if s and not s.startswith("//"):
                m = re.match(r'^(\w+\s+)?"([^"]+)"', s)
                if m:
                    a = m.group(1).strip() if m.group(1) else None
                    result.append((a, m.group(2), l))
    return result

def pkg_id(alias, path):
    if alias: return alias
    seg = path.split("/")[-1]
    if re.match(r'^v\d+$', seg) and len(path.split("/")) > 1:
        seg = path.split("/")[-2]
    return seg.replace("-", "")

def filter_imps(imps, body):
    """Keep only imports actually referenced (by pkgname.) in body."""
    out = []
    for a, p, raw in imps:
        name = pkg_id(a, p)
        if name in ("_", ".") or re.search(r'\b' + re.escape(name) + r'\.', body):
            out.append(raw)
    return out

# ─── Claude-friendly annotation generation ────────────────────────────────────

def find_toplevel(chunk_lines):
    """
    Find top-level declarations (func, type, var, const) in chunk_lines.

    Returns list of (sep_insert_idx, display_name, doc_first_line) where:
      sep_insert_idx — 0-based line index to insert the separator (before
                       the doc comment, not the func keyword itself)
      display_name   — "FuncName" or "MethodName (ReceiverType)"
      doc_first_line — first line of the doc comment, or ""
    """
    result = []
    for i, line in enumerate(chunk_lines):
        if not re.match(r'^(func|type|var|const)\s', line):
            continue

        # Extract display name (methods get receiver type appended)
        m_method = re.match(r'^func\s+\(([^)]+)\)\s+(\w+)', line)
        if m_method:
            recv_type = m_method.group(1).split()[-1].lstrip("*")
            name = f"{m_method.group(2)} ({recv_type})"
        else:
            m2 = re.match(r'^(?:func|type|var|const)\s+(\w+)', line)
            if not m2:
                continue  # skip unnamed blocks like "const (" or "var ("
            name = m2.group(1)

        # Walk backward to find where the doc comment starts
        j = i - 1
        sep_idx = i  # default: separator goes at the func line
        # Skip blank lines immediately before the declaration
        while j >= 0 and chunk_lines[j].strip() == "":
            j -= 1
        # Walk back through contiguous // comment lines
        while j >= 0 and chunk_lines[j].strip().startswith("//"):
            sep_idx = j   # move separator up to encompass the comment
            j -= 1

        # Get doc text from the first comment line (the topmost one)
        doc = ""
        if sep_idx < i:  # there is a comment
            doc = chunk_lines[sep_idx].strip().lstrip("/").strip()

        result.append((sep_idx, name, doc))
    return result

def make_sep(name):
    """Return a single-line separator: // ─── name ────────... (SEP_W chars)."""
    prefix = f"// ─── {name} "
    return prefix + "─" * max(2, SEP_W - len(prefix)) + "\n"

def make_toc(funcs):
    """Return a // ─── Contents ─── block listing all declarations."""
    if not funcs:
        return ""
    top = f"// ─── Contents {'─' * (SEP_W - 16)}\n"
    bot = f"// {'─' * (SEP_W - 3)}\n"
    lines = [top]
    for _, name, doc in funcs:
        desc = f" — {doc}" if doc else ""
        lines.append(f"//   {name}{desc}\n")
    lines.append(bot)
    return "".join(lines)

def annotate_body(chunk_lines, funcs):
    """
    Insert a section separator before each top-level declaration (before
    its doc comment when present).  Returns annotated body as a string.
    """
    sep_at = {sep_idx: name for sep_idx, name, _ in funcs}
    out = []
    for i, line in enumerate(chunk_lines):
        if i in sep_at:
            out.append(make_sep(sep_at[i]))
        out.append(line)
    return "".join(out)

# ─── File writing ──────────────────────────────────────────────────────────────

def wfile(out_path, desc, see_also, chunk, imps):
    """
    Write a split file:
      1. File comment + See-also hint
      2. package tools
      3. Filtered import block
      4. Contents TOC
      5. Annotated body (section separators inserted)
    """
    funcs = find_toplevel(chunk)
    body  = annotate_body(chunk, funcs).lstrip("\n")
    if not body.endswith("\n"):
        body += "\n"
    used  = filter_imps(imps, body)
    fname = os.path.basename(out_path)

    with open(out_path, "w") as f:
        f.write(f"// {fname} — {desc}\n")
        if see_also:
            f.write(f"// See also: {', '.join(see_also)}\n")
        f.write("package tools\n")
        if used:
            f.write("\nimport (\n")
            for l in used:
                f.write(l)
            f.write(")\n")
        f.write("\n")
        toc = make_toc(funcs)
        if toc:
            f.write(toc)
            f.write("\n")
        f.write(body)

    nf = sum(1 for l in chunk if l.strip())
    print(f"  {fname}: ~{nf} non-blank lines, {len(used)} imports, {len(funcs)} decls")

# ─── Per-file-group processor ─────────────────────────────────────────────────

def process_file(src_name, parts):
    """
    Read src_name, split into parts, annotate each part, write outputs.

    parts: list of (out_name, desc, start_0idx_or_None, end_0idx_or_None)
      - first part: start=None → use dynamically computed header_end
      - subsequent: start=<0-based line index>
      - end=None → end of file
    """
    src   = os.path.join(T, src_name)
    ls    = rlines(src)
    he    = header_end(ls)
    imps  = parse_imps(ls)
    names = [p[0] for p in parts]

    for i, (out_name, desc, start, end) in enumerate(parts):
        s     = he if i == 0 else start
        chunk = ls[s:end] if end is not None else ls[s:]
        see_also = [n for n in names if n != out_name]
        wfile(os.path.join(T, out_name), desc, see_also, chunk, imps)

# ─── Split specifications ──────────────────────────────────────────────────────
# Each entry: (source_filename, [(out_name, description, start, end), ...])
#
# • start/end are 0-based indices (exclusive end), matching Python slice notation.
# • First part uses start=None (→ header_end computed at runtime).
# • Split points are verified against grep output for each file.
#
# Naming convention for new files: original_base + _suffix.go

SPLITS = [
    # ── task_analysis_tags.go (1926 lines) → 3 files ─────────────────────────
    # Split 1 @ 0-idx 656: before discoverTagsFromMarkdown (grep: 657:func)
    # Split 2 @ 0-idx 1254: before buildBatchTagPrompt    (grep: 1255:func)
    ("task_analysis_tags.go", [
        ("task_analysis_tags.go",
         "Tag rules, NoiseTags, and tag-analysis action handlers.",
         None, 656),
        ("task_analysis_tags_discover.go",
         "Tag discovery: markdown docs, project tag caching, and batch size helpers.",
         656, 1254),
        ("task_analysis_tags_llm.go",
         "Tag enrichment: LLM batch prompts, parsing, scoring, Ollama/FM bridges.",
         1254, None),
    ]),

    # ── protobuf_helpers.go (1722 lines) → 3 files ───────────────────────────
    # Split 1 @ 0-idx 571: before GoScorecardResultToProto (grep: 572:func)
    # Split 2 @ 0-idx 1149: before GitToolsResponseToMap   (grep: 1150:func)
    ("protobuf_helpers.go", [
        ("protobuf_helpers.go",
         "Protobuf helpers: Memory, Context, Report, Health, and Metrics conversions.",
         None, 571),
        ("protobuf_helpers_report.go",
         "Protobuf helpers: Scorecard, BriefingData, TaskWorkflow, and Session conversions.",
         571, 1149),
        ("protobuf_helpers_tools.go",
         "Protobuf helpers: GitTools, Testing, MemoryMaint, TaskAnalysis, and Ollama conversions.",
         1149, None),
    ]),

    # ── report_plan.go (1184 lines) → 2 files ────────────────────────────────
    # Split @ 0-idx 537: before generatePlanMarkdown (grep: 538:func)
    ("report_plan.go", [
        ("report_plan.go",
         "Report plan: plan handler, wave formatting, and scorecard plans.",
         None, 537),
        ("report_plan_generate.go",
         "Report plan: plan markdown generation, YAML repair, and display helpers.",
         537, None),
    ]),

    # ── task_workflow_maintenance.go (999 lines) → 2 files ───────────────────
    # Split @ 0-idx 651: before handleTaskWorkflowFixEmptyDescriptions (grep: 652:func)
    ("task_workflow_maintenance.go", [
        ("task_workflow_maintenance.go",
         "Task workflow maintenance: sync, sanity-check, clarity, and cleanup handlers.",
         None, 651),
        ("task_workflow_maintenance_helpers.go",
         "Task workflow maintenance: fix descriptions, clarity/stale formatters, link planning.",
         651, None),
    ]),

    # ── git_tools.go (930 lines) → 2 files ───────────────────────────────────
    # Split @ 0-idx 466: before handleBranchesAction (grep: 467:func)
    ("git_tools.go", [
        ("git_tools.go",
         "Git tools: types, branch/commit helpers, and HandleGitToolsNative dispatcher.",
         None, 466),
        ("git_tools_actions.go",
         "Git tools: action handlers (branches, tasks, diff, graph, merge, set-branch).",
         466, None),
    ]),

    # ── task_workflow_create_ai.go (895 lines) → 2 files ─────────────────────
    # Split @ 0-idx 581: before handleTaskWorkflowSummarize (grep: 582:func)
    ("task_workflow_create_ai.go", [
        ("task_workflow_create_ai.go",
         "Task workflow: create, batch-create, single-create, enrich, fix-IDs, and estimate helpers.",
         None, 581),
        ("task_workflow_ai_run.go",
         "Task workflow: summarize and run-with-AI handlers.",
         581, None),
    ]),

    # ── ollama_native.go (882 lines) → 2 files ───────────────────────────────
    # Split @ 0-idx 482: before handleOllamaPull (grep: 483:func)
    ("ollama_native.go", [
        ("ollama_native.go",
         "Ollama native: types, dispatcher, availability, text generation, status, and models.",
         None, 482),
        ("ollama_native_handlers.go",
         "Ollama native: pull, hardware info, docs, quality check, and summary handlers.",
         482, None),
    ]),

    # ── session_helpers.go (881 lines) → 2 files ─────────────────────────────
    # Split @ 0-idx 481: before checkHandoffAlert (grep: 482:func)
    # GetSessionStatus (451) stays in file 1; handoff CRUD moves to file 2.
    ("session_helpers.go", [
        ("session_helpers.go",
         "Session helpers: types, agent/mode detection, hints, task summaries, and session status.",
         None, 481),
        ("session_helpers_handoff.go",
         "Session helpers: handoff CRUD, git status, and suggested next actions.",
         481, None),
    ]),

    # ── task_analysis_deps.go (828 lines) → 2 files ──────────────────────────
    # Split @ 0-idx 436: before handleTaskAnalysisComplexity (grep: 437:func)
    # fmtTime (432) stays with the execution-plan formatters that call it.
    ("task_analysis_deps.go", [
        ("task_analysis_deps.go",
         "Task analysis: dependency, summary, execution-plan handlers, formatters, and fmtTime.",
         None, 436),
        ("task_analysis_deps_analysis.go",
         "Task analysis: complexity, parallelization, fix-deps, validate, and noise handlers.",
         436, None),
    ]),

    # ── task_discovery_native.go (785 lines) → 2 files ───────────────────────
    # Split @ 0-idx 454: before enhanceTaskWithAppleFM (grep: 455:func)
    ("task_discovery_native.go", [
        ("task_discovery_native.go",
         "Task discovery: main handler, comment scanner, markdown scanner, and orphan finder.",
         None, 454),
        ("task_discovery_native_scanners.go",
         "Task discovery: Apple FM enhancement, planning doc scanner, and git JSON scanner.",
         454, None),
    ]),

    # ── memory_maint.go (770 lines) → 2 files ────────────────────────────────
    # Split @ 0-idx 547: before handleMemoryMaintDream (grep: 548:func)
    # Keeps handlers (health, GC, prune, consolidate) in file 1;
    # analytical utilities (dream, similarity, dedup, merge) in file 2.
    ("memory_maint.go", [
        ("memory_maint.go",
         "Memory maintenance: dispatcher, health, GC, prune, and consolidate handlers.",
         None, 547),
        ("memory_maint_utils.go",
         "Memory maintenance: dream, similarity scoring, duplicate groups, and memory merge.",
         547, None),
    ]),
]

# ─── Entry point ──────────────────────────────────────────────────────────────

def main():
    n = sum(len(parts) for _, parts in SPLITS)
    print(f"Phase 3: {len(SPLITS)} source files → {n} output files (parallel, max_workers=6)\n")

    errors = []
    with ThreadPoolExecutor(max_workers=6) as ex:
        futures = {ex.submit(process_file, src, parts): src for src, parts in SPLITS}
        for fut in as_completed(futures):
            src = futures[fut]
            try:
                fut.result()
                print(f"✓ {src}")
            except Exception as e:
                print(f"✗ {src}: {e}", file=sys.stderr)
                errors.append((src, str(e)))

    if errors:
        print(f"\n{len(errors)} error(s):", file=sys.stderr)
        for src, e in errors:
            print(f"  {src}: {e}", file=sys.stderr)
        sys.exit(1)
    else:
        print(f"\nAll {n} files written. Run: make b QUIET=1")

if __name__ == "__main__":
    main()
