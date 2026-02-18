# Fixing npm Certificate / SSL Issues

If you see errors like **"unable to get local issuer certificate"** or **`UNABLE_TO_GET_ISSUER_CERT_LOCALLY`** when running `npm install`, `npx`, or other npm commands, use one of the options below.

## Option 1: Add your CA certificate (recommended)

If you have a corporate or custom CA certificate (e.g. from IT):

**Using environment variable (session-only):**
```bash
export NODE_EXTRA_CA_CERTS=/path/to/your/ca-bundle.pem
```

**Using npm config (persistent):**
```bash
npm config set cafile /path/to/your/ca-bundle.pem
```

To combine Mozilla’s bundle with your CA (e.g. corporate proxy):
```bash
curl -o ~/.npm.certs.pem https://curl.se/ca/cacert.pem
cat /path/to/your-extra-ca.pem >> ~/.npm.certs.pem
npm config set cafile ~/.npm.certs.pem
```

## Option 2: Disable strict SSL (temporary only)

Use only when you cannot add the correct CA and accept the security risk:

```bash
npm config set strict-ssl false
```

To re-enable later:
```bash
npm config set strict-ssl true
```

## Option 3: Use an up-to-date CA bundle (old macOS / old Node bundle)

Node uses its own **bundled** Mozilla CA store (fixed at Node release time). OpenSSL and some builds use a **system or Homebrew** CA file that may be outdated on older macOS. If you see "unable to get local issuer certificate" and you're on older macOS or an older Node install, try pointing at a fresh Mozilla bundle:

```bash
# Download current Mozilla CA bundle (from curl.se)
curl -o ~/.npm.cacert.pem https://curl.se/ca/cacert.pem

# Use it for npm
npm config set cafile ~/.npm.cacert.pem

# Optional: use for Node in general (e.g. MCP servers)
export NODE_EXTRA_CA_CERTS=~/.npm.cacert.pem
```

Then run `npm config set strict-ssl true` if you had turned it off, and retry.

## Option 4: Update Node.js and npm

Newer Node.js (18+) and npm ship with updated CA bundles and TLS behavior. Upgrade if you’re on an old version:

```bash
node -v
npm -v
# Upgrade via nvm, official installer, or your package manager
```

## Verify

```bash
npm config get cafile
npm config get strict-ssl
npm install
```

## References

- [npm cafile / ca config](https://docs.npmjs.com/cli/v10/using-npm/config#cafile)
- [Node.js NODE_EXTRA_CA_CERTS](https://nodejs.org/api/cli.html#node_extra_ca_certsfile)
- [Stack Overflow: npm add root CA](https://stackoverflow.com/questions/23788564/npm-add-root-ca)
