# rrctl Prebuilt Binaries

This directory contains precompiled binaries for multiple platforms.

## Available Binaries

| Platform | Binary | Architecture |
|----------|--------|--------------|
| Linux | `rrctl-linux-amd64` | x86_64 (Intel/AMD) |
| macOS | `rrctl-darwin-amd64` | x86_64 (Intel) |
| macOS | `rrctl-darwin-arm64` | ARM64 (Apple Silicon M1/M2/M3) |
| Windows | `rrctl-windows-amd64.exe` | x86_64 (Intel/AMD) |

## Quick Install

### Linux / macOS

```bash
# One-line installer (recommended)
curl -fsSL https://raw.githubusercontent.com/kushin77/elevatedIQ/main/rrctl-opensource/install.sh | bash

# Or manual download
# Linux
curl -fsSL -o rrctl https://github.com/kushin77/elevatedIQ/raw/main/rrctl-opensource/rrctl-linux-amd64
chmod +x rrctl
sudo mv rrctl /usr/local/bin/

# macOS Intel
curl -fsSL -o rrctl https://github.com/kushin77/elevatedIQ/raw/main/rrctl-opensource/rrctl-darwin-amd64
chmod +x rrctl
sudo mv rrctl /usr/local/bin/

# macOS Apple Silicon
curl -fsSL -o rrctl https://github.com/kushin77/elevatedIQ/raw/main/rrctl-opensource/rrctl-darwin-arm64
chmod +x rrctl
sudo mv rrctl /usr/local/bin/
```

### Windows

1. Download `rrctl-windows-amd64.exe` from the releases
2. Rename to `rrctl.exe`
3. Add to PATH or place in `C:\Windows\System32`

## Verification

```bash
# Check version
rrctl version

# Run help
rrctl --help
```

## Building from Source

If you prefer to build from source:

```bash
git clone https://github.com/kushin77/elevatedIQ.git
cd elevatedIQ/rrctl-opensource
go build -o rrctl .
```

## Security

All binaries are built from the source code in this repository. You can verify the build process by:

1. Checking the GitHub Actions workflow
2. Building from source and comparing checksums
3. Running `go version` on the binary to see build info

## Support

- Issues: [GitHub Issues](https://github.com/kushin77/elevatedIQ/issues)
- Documentation: See [README.md](./README.md)
- Examples: See [EXAMPLES.md](./EXAMPLES.md)
