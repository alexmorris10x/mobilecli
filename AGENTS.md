# MobileCLI 10x Fork

This public fork maintains a safe, repeatable physical-iPhone control surface for Alex's local tooling.

- Keep canonical product and workflow documentation in `/Users/10x/10x-os/90-internal/Physical iPhone Control/`.
- Keep build, test, release, and implementation documentation required by this repository here.
- Bind every unauthenticated device-control endpoint to loopback only. Do not expose WebDriverAgent, DeviceKit, or MobileCLI control ports on LAN interfaces.
- Preserve the complete upstream agent build pipeline. Never commit placeholder embedded agents, signed apps, device identifiers, profiles, screenshots, credentials, tokens, or authenticated runtime state.
- Keep `upstream` pointed at `https://github.com/mobile-next/mobilecli.git`; keep 10x-specific changes focused and easy to rebase.
- Verify forwarding changes with focused socket tests and a listener inspection before release.
