# Docklog Release Checklist

This document outlines the standardized process for verifying a release of Docklog. Every version MUST pass the **Stability Audit** on both Windows and Linux/WSL.

## 1. Automated Stability Audit (The "Go/No-Go" Test)

Before any release, run the following scripts. They don't just check if it installs; they audit the binary for **portability** and **static linking**.

### Windows
```powershell
./scripts/verify-release.ps1
```
*Tests performed:*
- CGO Audit (Static linking verification)
- Multi-arch build verification (Linux, ARM)
- Binary integrity check (Standalone execution)

### Linux / WSL
```bash
chmod +x ./scripts/verify-release.sh
./scripts/verify-release.sh
```
*Tests performed:*
- CGO Audit
- Multi-arch build verification
- **Portability Test:** Binary is executed inside a clean `alpine` container to ensure it has no hidden dependencies on the host OS.

---

## 2. Manual Verification (Final Polish)

Once the automated audit passes, perform these final checks:

- [ ] **Docker Discovery:** Run `docklog check`.
- [ ] **TUI Colors:** Run `docklog start`. Verify color-coding is legible on your current terminal theme.
- [ ] **CLI Help:** Verify `docklog --help` is up to date with new flags.

## 3. Distribution

- [ ] Clear the `dist/` directory created by the audit scripts.
- [ ] Tag the commit: `git tag -a v0.X.X -m "Release version v0.X.X"`.
- [ ] Push tags: `git push origin v0.X.X`.

---

> [!IMPORTANT]
> If the **Portability Test** fails on Linux, DO NOT publish the release. It means the binary will likely crash on users' machines with "missing library" errors. Check that `CGO_ENABLED=0` is being used correctly.
