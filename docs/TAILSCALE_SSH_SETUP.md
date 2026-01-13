# Tailscale SSH Configuration Guide

## Overview

Tailscale SSH allows secure SSH connections through your Tailscale mesh network without exposing services to the public internet.

## Prerequisites

- Tailscale account with admin access
- Tailscale client installed on both machines
- Both machines joined to the same Tailscale network

## Step 1: Enable SSH on Machines

### On Each Machine (Client and Server)

```bash
# Enable Tailscale SSH
sudo tailscale set --ssh=true

# Verify SSH is enabled
tailscale status --json | grep -i ssh
```

### Alternative: Enable via Admin Console

1. Go to https://login.tailscale.com/admin/machines
2. Find your machine (`davids-mac-mini.tailf62197.ts.net`)
3. Click on the machine name
4. Toggle "SSH" to enabled
5. Save changes

## Step 2: Configure ACL (Access Control List)

### Access Tailscale Admin Console

1. Go to https://login.tailscale.com/admin/acls
2. Click "Edit ACL file"

### Example ACL Configuration

Add SSH access rules to your ACL file:

```json
{
  "acls": [
    // Allow SSH from your current machine to davids-mac-mini
    {
      "action": "accept",
      "src": ["dlowes-imac18-3-1.tailf62197.ts.net"],
      "dst": ["davids-mac-mini.tailf62197.ts.net:22"]
    },
    // Or allow all machines in your tailnet to SSH to each other
    {
      "action": "accept",
      "src": ["autogroup:members"],
      "dst": ["autogroup:members:22"]
    }
  ],
  "ssh": [
    // Allow SSH connections with automatic user mapping
    {
      "action": "accept",
      "src": ["autogroup:members"],
      "dst": ["autogroup:members"],
      "users": ["autogroup:nonroot"]
    }
  ]
}
```

### Common SSH ACL Patterns

**Allow all members to SSH to all machines:**
```json
{
  "ssh": [
    {
      "action": "accept",
      "src": ["autogroup:members"],
      "dst": ["autogroup:members"],
      "users": ["autogroup:nonroot"]
    }
  ]
}
```

**Allow specific user to SSH to specific machine:**
```json
{
  "ssh": [
    {
      "action": "accept",
      "src": ["dlowes-imac18-3-1.tailf62197.ts.net"],
      "dst": ["davids-mac-mini.tailf62197.ts.net"],
      "users": ["dlowes"]
    }
  ]
}
```

**Allow with sudo access:**
```json
{
  "ssh": [
    {
      "action": "accept",
      "src": ["autogroup:members"],
      "dst": ["autogroup:members"],
      "users": ["autogroup:nonroot", "autogroup:members"]
    }
  ]
}
```

## Step 3: Save and Apply ACL

1. Click "Save" in the ACL editor
2. Wait for changes to propagate (usually < 1 minute)
3. Changes apply automatically to all machines

## Step 4: Verify Configuration

### Check SSH Status

```bash
# On client machine - check if SSH is enabled
tailscale status --json | python3 -m json.tool | grep -A 5 SSH

# Test SSH connection (RECOMMENDED: Use regular SSH)
ssh davids-mac-mini.tailf62197.ts.net

# Note: tailscale ssh is optional - regular SSH works perfectly
# The wrapper is mainly useful for MagicDNS resolution and userspace-networking mode
tailscale ssh davids-mac-mini.tailf62197.ts.net
```

**Important:** Per `tailscale ssh --help`: *"most users running the Tailscale SSH server will prefer to just use the normal 'ssh' command"*. Regular SSH routes through Tailscale's network and is more reliable.

### Check Connection

```bash
# Test connection with verbose output
ssh -v davids-mac-mini.tailf62197.ts.net "echo 'SSH working!'"

# Check host key
ssh-keyscan davids-mac-mini.tailf62197.ts.net
```

## Step 5: Configure SSH Client (Optional)

### Add to ~/.ssh/config

```ssh-config
Host *.ts.net
    UserKnownHostsFile ~/.ssh/known_hosts
    StrictHostKeyChecking accept-new
    IdentityFile ~/.ssh/id_ed25519

Host davids-mac-mini
    HostName davids-mac-mini.tailf62197.ts.net
    User dlowes
```

Then you can use: `ssh davids-mac-mini`

## Troubleshooting

### "Host key verification failed" with `tailscale ssh`

**Issue:** The `tailscale ssh` wrapper command may have issues with host key verification when hostnames include trailing periods or DNS resolution differs from regular SSH.

**Solution:** Use regular `ssh` command instead of `tailscale ssh`:

```bash
# Instead of: tailscale ssh davids-mac-mini.tailf62197.ts.net
# Use: 
ssh davids-mac-mini.tailf62197.ts.net
```

Regular SSH still routes through Tailscale's network and respects your SSH config. The `tailscale ssh` command is just a wrapper and not required for Tailscale connectivity.

**Why this happens:**
- `tailscale ssh` resolves hostnames differently, potentially adding trailing periods
- SSH's strict host key checking requires exact hostname matches
- Regular SSH uses DNS directly and respects `~/.ssh/config` properly

**Workaround if you must use `tailscale ssh`:**
1. Ensure host key is in known_hosts for both `hostname` and `hostname.` (with trailing period)
2. Add `StrictHostKeyChecking accept-new` to your SSH config for `*.ts.net` hosts

**Reference:** This is a known issue - see [GitHub Issue #9702](https://github.com/tailscale/tailscale/issues/9702)

### "SSH not enabled on this machine"

**Solution:**
```bash
sudo tailscale set --ssh=true
# Or enable via admin console
```

### "Permission denied" or ACL not working

**Check ACL:**
1. Verify ACL is saved in admin console
2. Check that source and destination match your machines
3. Ensure "users" array includes your user
4. Wait 1-2 minutes for ACL to propagate

**Debug:**
```bash
# Check what user Tailscale thinks you are
tailscale status --json | python3 -c "import sys, json; print(json.load(sys.stdin)['Self'].get('UserID'))"

# Check ACL from admin console
# https://login.tailscale.com/admin/acls
```

### "Host key verification failed"

**Solution:**
```bash
# Add host key to known_hosts
ssh-keyscan davids-mac-mini.tailf62197.ts.net >> ~/.ssh/known_hosts

# Or use accept-new in SSH config
```

### Connection timeout

**Check:**
1. Both machines are online in Tailscale
2. Tailscale status shows both machines connected
3. Firewall allows SSH (port 22)

```bash
# Check machine status
tailscale status

# Ping via Tailscale IP
ping 100.117.185.32  # Replace with actual Tailscale IP
```

## Security Best Practices

1. **Use ACLs:** Don't allow SSH from all sources - restrict to specific machines/users
2. **Principle of Least Privilege:** Only grant SSH access to machines/users that need it
3. **Use SSH Keys:** Configure SSH key authentication instead of passwords
4. **Regular Updates:** Keep Tailscale client updated
5. **Monitor Access:** Review Tailscale logs for SSH access patterns

## Additional Resources

- [Tailscale SSH Documentation](https://tailscale.com/kb/1193/tailscale-ssh/)
- [Tailscale ACL Documentation](https://tailscale.com/kb/1018/acls/)
- [SSH Configuration Guide](https://tailscale.com/kb/1193/tailscale-ssh/#ssh-configuration)

