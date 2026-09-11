# MobileCLI 10x Fork

This public fork maintains a safe, repeatable physical-iPhone control surface for Alex's local tooling.

- Keep canonical product and workflow documentation in this repository. Keep dated work packets and completed evidence in Google Drive dated folders, following the central Google Drive `AGENTS.md`.
- Keep build, test, release, and implementation documentation required by this repository here.
- Bind every unauthenticated device-control endpoint to loopback only. Do not expose WebDriverAgent, DeviceKit, or MobileCLI control ports on LAN interfaces.
- Preserve the complete upstream agent build pipeline. Never commit placeholder embedded agents, signed apps, device identifiers, profiles, screenshots, credentials, tokens, or authenticated runtime state.
- Keep `upstream` pointed at `https://github.com/mobile-next/mobilecli.git`; keep 10x-specific changes focused and easy to rebase.
- Verify forwarding changes with focused socket tests and a listener inspection before release.
