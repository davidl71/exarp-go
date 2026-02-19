# Git hooks

## Co-authored-by trailer (automatic)

To add a `Co-authored-by:` line to every commit without typing it each time:

1. **Use this repo’s hooks directory** (from repo root):
   ```bash
   git config core.hooksPath scripts/git-hooks
   ```

2. **Set the co-author** (name and email as you want it to appear):
   ```bash
   git config exarp-go.coauthor "Co Author Name <coauthor@example.com>"
   ```

After that, every commit (including `git commit -m "message"`) will get a trailing line:
`Co-authored-by: Co Author Name <coauthor@example.com>`.

- The hook does nothing if `exarp-go.coauthor` is not set.
- It does not add a second trailer if `Co-authored-by:` is already in the message.
- Merge and squash commits are left unchanged.

### Alternative: only this hook

If you prefer not to change `core.hooksPath`, install just this hook:

```bash
cp scripts/git-hooks/prepare-commit-msg .git/hooks/prepare-commit-msg
chmod +x .git/hooks/prepare-commit-msg
git config exarp-go.coauthor "Co Author Name <coauthor@example.com>"
```
