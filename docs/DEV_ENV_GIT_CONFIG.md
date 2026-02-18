# Development environment: Git and Make shortcuts

Optional git and make settings applied when creating or using the dev environment.

## Git: less verbose output

Git can be configured to turn off advice/hint messages (e.g. “use git add…”, “use git restore…”, push/merge hints). The development Ansible playbook can set these for the dev user so new environments get quieter git output.

**Applied by Ansible (development playbook):** When you run `ansible-playbook playbooks/development.yml`, the **common** role sets `advice.*` to `false` for the user running the playbook (see `ansible/roles/common/tasks/main.yml`). That matches the list below.

**Apply manually (current machine):**

```bash
for key in statusHints statusUoption addEmptyPathspec addEmbeddedRepo checkoutAmbiguousRemoteBranchName commitBeforeMerge detachedHead diverging implicitIdentity mergeConflict nestedTag pushNonFFCurrent pushNonFFMatching pushUpdateRejected pushRefNeedsUpdate rebaseInProgress resetQuiet rmHints updateRefs waitingForEditor ignoredHook resolveConflict forceDeleteBranch skippedCherryPicks amWorkDir; do
  git config --global advice.$key false
done
```

**Revert (re-enable hints):**

```bash
git config --global --unset-all advice.statusHints
# Repeat for other keys, or reset the advice namespace if your git supports it.
```

**References:** `.cursor/rules/make-shortcuts.mdc`, `make status` / `make st` in the project Makefile (also uses quieter status when run from repo root).

## Make shortcuts (repo root)

From repo root use:

- **Build:** `make b`
- **Repo root:** `make r` or `make root` → use with `cd $(make r)` in scripts
- **Git status:** `make status` or `make st`
- **Git push/pull:** `make p` / `make pl`

From a subdir: use **`r`** then make (e.g. `r && make st`) or **`make -C $(git rev-parse --show-toplevel) st`**. See `.cursor/rules/make-shortcuts.mdc` and `docs/EXARP_CLI_SHORTCUTS.md`.

## See also

- **Ansible:** `ansible/README.md`, `docs/ANSIBLE_SETUP.md`, `ansible/run-dev-setup.sh`
- **npm/SSL:** `docs/NPM_CERTIFICATE_FIX.md` (if you hit certificate errors with npx/npm)
