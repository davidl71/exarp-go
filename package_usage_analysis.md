# Package Usage Analysis
## Packages to Keep vs Remove

### ✅ KEEP (User Requested)
- **1password**: 503.0 MB - Password manager (user requested to keep)
- **google-chrome-stable**: 373.7 MB - Browser (user requested to keep)  
- **golang packages**: User requested to keep golang (but see note about golang-1.22-src below)

### ✅ KEEP (Essential/In Use)
- **QEMU x86_64 packages**: ~57 MB - You have a macOS VM, need x86_64 QEMU
  - `qemu-system-x86` - Essential for your macOS VM
  - `qemu-system` - Meta package
  - `qemu-system-common` - Common files
  - `qemu-system-data` - Data files
  - `qemu-system-gui` - GUI support
  - `qemu-utils` - Utilities
  - `qemu-img` - Disk image tools
  - `libvirt` packages - Managing VMs

### ❌ REMOVE (Not Used)
1. **LibreOffice**: ~144.4 MB
   - Status: Not installed (already removed) or 0 packages found
   - Action: Already removed or was a false positive

2. **golang-1.22-src**: 119.0 MB
   - Status: Source code, not needed for building/running Go programs
   - Action: Safe to remove (keep golang-1.22-go which is the compiler)
   - Command: `sudo apt remove --purge golang-1.22-src`

3. **Old Kernel Headers**: ~82 MB
   - Status: Running kernel 6.14.0-37, but have headers for 6.8.0-90
   - Packages:
     - `linux-headers-6.8.0-90` (82.0 MB)
     - `linux-headers-6.8.0-90-generic` (part of the above)
   - Action: Safe to remove (not needed for current kernel)
   - Command: `sudo apt remove --purge linux-headers-6.8.0-90 linux-headers-6.8.0-90-generic`

4. **QEMU Non-x86 Architectures**: ~765 MB
   - Status: You only need x86_64 for your macOS VM
   - Packages that can be removed:
     - `qemu-efi-aarch64` (322.1 MB) - ARM64 firmware
     - `qemu-efi-arm` (128.0 MB) - ARM firmware  
     - `qemu-system-arm` (57.6 MB) - ARM emulation
     - `qemu-system-mips` (60.6 MB) - MIPS emulation
     - `qemu-system-ppc` - PowerPC emulation
     - `qemu-system-s390x` - IBM S390x emulation
     - `qemu-system-sparc` - SPARC emulation
     - `qemu-system-misc` (213.5 MB) - Miscellaneous architectures
   - Action: Safe to remove if you only use x86_64 VMs
   - Command: `sudo apt remove --purge qemu-efi-aarch64 qemu-efi-arm qemu-system-arm qemu-system-mips qemu-system-ppc qemu-system-s390x qemu-system-sparc qemu-system-misc`

### ⚠️ CONSIDER (Maybe Remove)
- **LLVM packages**: Multiple versions (libllvm18, libllvm20, libllvm20.1-amdgpu)
  - Total: ~357 MB
  - Status: Needed for GPU drivers (amdgpu), but might have unused version
  - Action: Keep for now (needed for GPU)

### 📊 Summary

**Safe to Remove:**
- golang-1.22-src: 119.0 MB
- Old kernel headers: ~82 MB  
- QEMU non-x86 packages: ~765 MB
- **Total: ~966 MB**

**Potentially Remove (if LibreOffice exists):**
- LibreOffice: ~144.4 MB
- **Total if included: ~1.1 GB**

## Commands to Remove (Run in order)

```bash
# 1. Remove Go source code (keep compiler)
sudo apt remove --purge golang-1.22-src

# 2. Remove old kernel headers
sudo apt remove --purge linux-headers-6.8.0-90 linux-headers-6.8.0-90-generic

# 3. Remove QEMU non-x86 architectures (if you only use x86_64 VMs)
sudo apt remove --purge qemu-efi-aarch64 qemu-efi-arm qemu-system-arm qemu-system-mips qemu-system-ppc qemu-system-s390x qemu-system-sparc qemu-system-misc

# 4. Clean up after removal
sudo apt autoremove -y
sudo apt autoclean
```

## Verification

After removal, verify your VM still works:
```bash
virsh list --all
qemu-system-x86_64 --version
```
