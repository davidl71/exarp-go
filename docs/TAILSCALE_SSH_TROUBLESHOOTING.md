# Tailscale SSH Troubleshooting Guide

## Common Issues and Solutions

### Issue: "Host key verification failed" with `tailscale ssh`

**Symptom:**
```
No ED25519 host key is known for hostname.tailf62197.ts.net. and you have requested strict checking.
Host key verification failed.
```

**Root Cause:**
According to `tailscale ssh --help`, the wrapper:
1. Automatically checks the destination server's SSH host key against the node's SSH host key as advertised via the Tailscale coordination server
2. Resolves hostnames using MagicDNS (even if --accept-dns=false)
3. Uses a ProxyCommand for userspace-networking mode

This automatic host key checking against Tailscale's coordination server can conflict with your local `~/.ssh/known_hosts` file, especially when hostnames are resolved differently (sometimes with trailing periods). SSH's strict host key checking requires exact hostname matches, causing verification to fail.

**Solutions (in order of preference):**

#### ✅ Solution 1: Use Regular SSH (Recommended)

**Use regular `ssh` instead of `tailscale ssh`:**

```bash
# Works perfectly - routes through Tailscale network
ssh davids-mac-mini.tailf62197.ts.net
```

**Why this works:**
- Regular SSH respects your `~/.ssh/config` settings
- Still routes through Tailscale's secure network (via DNS resolution)
- Better host key handling (uses your local known_hosts)
- More predictable behavior
- No conflicts with Tailscale's automatic host key checking

**Note:** As documented in `tailscale ssh --help`: *"The 'tailscale ssh' command is an optional wrapper around the system 'ssh' command... most users running the Tailscale SSH server will prefer to just use the normal 'ssh' command or their normal SSH client."*

#### Solution 2: Add Host Key for Both Variants

If you must use `tailscale ssh`, add the host key for both hostname variants:

```bash
# Add without trailing period
ssh-keyscan davids-mac-mini.tailf62197.ts.net >> ~/.ssh/known_hosts

# Add with trailing period (what tailscale ssh sees)
ssh-keyscan "davids-mac-mini.tailf62197.ts.net." >> ~/.ssh/known_hosts
```

#### Solution 3: Configure SSH to Accept New Keys

Add to `~/.ssh/config`:

```ssh-config
Host *.ts.net *.ts.net.
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts
```

**Note:** This may not work reliably with `tailscale ssh` wrapper.

### Issue: SSH Not Enabled

**Symptom:**
- Cannot connect via SSH
- "Connection refused" or "Permission denied"

**Solution:**

Enable SSH on both machines:

```bash
# On each machine
sudo tailscale set --ssh=true
```

Or via admin console:
1. Go to https://login.tailscale.com/admin/machines
2. Find the machine
3. Toggle "SSH" to enabled

### Issue: ACL Not Working

**Symptom:**
- "Permission denied" even though SSH is enabled
- Connection timeout

**Solution:**

1. **Check ACL in admin console:**
   - Visit https://login.tailscale.com/admin/acls
   - Ensure SSH rules are configured

2. **Verify ACL syntax:**
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

3. **Check source/destination:**
   - Verify machine names match exactly
   - Check both DNS names and IP addresses are allowed

4. **Wait for propagation:**
   - ACL changes take 1-2 minutes to propagate
   - Wait and try again

### Issue: Connection Timeout

**Symptom:**
- SSH connection hangs or times out
- "Connection refused"

**Checklist:**

1. **Verify machines are online:**
   ```bash
   tailscale status
   ```
   Both machines should show as "online"

2. **Test Tailscale connectivity:**
   ```bash
   ping <tailscale-ip>
   ```

3. **Check SSH service:**
   ```bash
   # On the server machine
   sudo systemctl status ssh
   # or
   sudo service ssh status
   ```

4. **Check firewall:**
   - Ensure port 22 is not blocked
   - Tailscale bypasses most firewalls, but check anyway

5. **Check SSH is listening:**
   ```bash
   # On the server machine
   sudo netstat -tlnp | grep :22
   # or
   sudo ss -tlnp | grep :22
   ```

## Best Practices

### 1. Use Regular SSH

**Recommendation:** Use `ssh` instead of `tailscale ssh`:

```bash
ssh hostname.tailf62197.ts.net
```

**Benefits:**
- More reliable host key handling
- Respects SSH config
- Standard SSH behavior
- Still uses Tailscale network

### 2. Configure SSH Config

Create `~/.ssh/config` entries:

```ssh-config
# Tailscale hosts - accept new keys automatically
Host *.ts.net *.ts.net.
    StrictHostKeyChecking accept-new
    UserKnownHostsFile ~/.ssh/known_hosts

# Specific host with alias
Host davids-mac-mini
    HostName davids-mac-mini.tailf62197.ts.net
    User davidl
    StrictHostKeyChecking accept-new
```

Then use: `ssh davids-mac-mini`

### 3. Keep Host Keys Updated

If you reinstall a machine or change SSH keys:

```bash
# Remove old key
ssh-keygen -R hostname.tailf62197.ts.net

# Add new key
ssh-keyscan hostname.tailf62197.ts.net >> ~/.ssh/known_hosts
```

### 4. Monitor ACL Changes

- Always save ACL changes in admin console
- Wait 1-2 minutes for propagation
- Test with `tailscale status` to verify connectivity

## Known Issues

### GitHub Issue #9702

There's a known issue with `tailscale ssh` command and host key verification:
- [GitHub Issue #9702](https://github.com/tailscale/tailscale/issues/9702)

**Workaround:** Use regular `ssh` command instead of `tailscale ssh`.

## Quick Reference

| Problem | Solution |
|---------|----------|
| Host key verification failed | Use `ssh` instead of `tailscale ssh` |
| SSH not enabled | `sudo tailscale set --ssh=true` or admin console |
| Permission denied | Check ACL configuration |
| Connection timeout | Verify machines online, check SSH service |
| ACL not working | Verify syntax, wait for propagation |

## Additional Resources

- [Tailscale SSH Documentation](https://tailscale.com/kb/1193/tailscale-ssh/)
- [Tailscale ACL Documentation](https://tailscale.com/kb/1018/acls/)
- [SSH Configuration Guide](https://tailscale.com/kb/1193/tailscale-ssh/#ssh-configuration)
- [GitHub: Tailscale SSH Issues](https://github.com/tailscale/tailscale/issues?q=is%3Aissue+is%3Aopen+ssh)

