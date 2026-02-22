# Ansible Galaxy and Offline Flow

**Tag hints:** `#docs`

> How to use Ansible Galaxy roles/collections in exarp-go, including offline/air-gapped environments.

## Requirements File

Galaxy dependencies are declared in `ansible/requirements.yml`:

```yaml
collections:
  - name: community.general
    version: ">=8.0.0"
```

## Online Installation

```bash
# Install all requirements
ansible-galaxy install -r ansible/requirements.yml

# Install collections only
ansible-galaxy collection install -r ansible/requirements.yml

# Force reinstall
ansible-galaxy collection install -r ansible/requirements.yml --force
```

## Offline / Air-Gapped Flow

For environments without internet access:

### 1. Download on a Connected Machine

```bash
# Download collections as tarballs
ansible-galaxy collection download -r ansible/requirements.yml -p ./ansible/offline-collections/

# This creates .tar.gz files in the target directory
```

### 2. Transfer to Air-Gapped Machine

Copy the `ansible/offline-collections/` directory to the target machine via USB, SCP, or other secure transfer method.

### 3. Install from Local Tarballs

```bash
# Install from local directory
ansible-galaxy collection install -r ansible/requirements.yml \
  -p ~/.ansible/collections \
  --offline \
  -s ./ansible/offline-collections/

# Or install individual tarballs
ansible-galaxy collection install ./ansible/offline-collections/community-general-8.0.0.tar.gz
```

### 4. Verify Installation

```bash
ansible-galaxy collection list
```

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/go.yml`) handles Galaxy dependencies automatically:

```yaml
- name: Install Ansible Galaxy requirements
  run: ansible-galaxy install -r ansible/requirements.yml
```

For self-hosted runners in restricted networks, pre-install collections on the runner or use the offline flow above.

## Docker Smoke Test

The Makefile provides a Docker-based syntax check that includes Galaxy installation:

```bash
make ansible-docker-smoke
```

This runs `ansible-playbook --syntax-check` inside an Ubuntu container with all Galaxy dependencies installed.

## Adding New Dependencies

1. Add the collection/role to `ansible/requirements.yml`
2. Run `ansible-galaxy install -r ansible/requirements.yml` locally
3. Test with `make ansible-docker-smoke`
4. For offline users: update the download step to include new dependencies

## Integration Test

To run the Ansible integration test:

```bash
# Syntax check (no target hosts needed)
ansible-playbook ansible/playbooks/site.yml --syntax-check

# Dry run against localhost
ansible-playbook ansible/playbooks/site.yml --check --connection=local

# Docker smoke test (recommended for CI)
make ansible-docker-smoke
```
