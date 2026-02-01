# Example: Cursor-Style Plan

This file shows the **plan structure** recommended by [Cursor — Creating Plans](https://cursor.com/learn/creating-plans). Use exarp-go to generate a project-specific plan (default: `.cursor/plans/<project-slug>.plan.md` so Cursor shows Build):

```bash
exarp-go -tool report -args '{"action":"plan"}'
# or custom path:
exarp-go -tool report -args '{"action":"plan","output_path":".cursor/plans/my-feature.plan.md"}'
```

Generated plans include YAML frontmatter with `status: draft`; Cursor may show "Built" when you build from the plan.

---

# Plan: Music Tracking Web App

**Generated:** 2026-02-01

---

## 1. Purpose & Success Criteria

**Purpose:** Build a web app to track music listening habits—connect to Spotify, analyze patterns, and share stats with friends.

**Success criteria:**

- Users can log in with Google OAuth (no passwords).
- Users can connect a Spotify account and see top artists/songs (weekly/monthly).
- Dashboard shows listening time trends and genre breakdown.
- Public profile page allows sharing stats with friends.
- App is mobile-friendly and each milestone is independently shippable.

---

## 2. Technical Foundation

- **Frontend:** Next.js App Router, TypeScript, Tailwind CSS, shadcn/ui
- **Backend:** API Routes, Postgres with Drizzle ORM
- **Auth:** Google OAuth
- **Deploy:** Vercel
- **Invariants:** TypeScript strict mode; component tests for critical paths; no direct DB access from client

---

## 3. Iterative Milestones

Each milestone is independently valuable. Check off as done.

- [ ] **Set up project** — Next.js, TypeScript, Tailwind, repo structure
- [ ] **Google OAuth** — Sign in / sign out, session handling
- [ ] **Spotify connection** — Connect account, store tokens securely
- [ ] **Dashboard — Top artists/songs** — Weekly and monthly views
- [ ] **Listening time trends** — Charts and time-range filters
- [ ] **Genre breakdown** — Summary and filters
- [ ] **Public profile page** — Shareable stats, privacy controls
- [ ] **Mobile-friendly UI** — Responsive layout and touch targets

---

## 4. Open Questions

- Preferred Spotify data refresh interval (e.g. daily vs on-demand)?
- Rate limits and caching strategy for Spotify API?
- What profile fields are shareable by default (e.g. top 5 artists only)?

---

*Structure from [Creating Plans](https://cursor.com/learn/creating-plans). Generate your own with `report(action="plan")`.*
