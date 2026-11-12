#!/usr/bin/env bash
# Builds the connector for every platform and packs it with its manifest as ../anduin-builder.mcpb, the file the
# plugin's mcpServers points at. Needs macOS (lipo joins the two Mac builds into one). The output is byte-for-byte
# reproducible from the same source and Go version, so CI rebuilds it and fails if it differs from the commit.
set -euo pipefail
cd "$(dirname "$0")"

out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT
mkdir -p "$out/bin"

build() {
  CGO_ENABLED=0 GOOS="$1" GOARCH="$2" go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o "$3" .
}
build darwin arm64 "$out/darwin-arm64"
build darwin amd64 "$out/darwin-amd64"
lipo -create -output "$out/bin/anduin-builder" "$out/darwin-arm64" "$out/darwin-amd64"
build windows amd64 "$out/bin/anduin-builder.exe"
build linux amd64 "$out/bin/anduin-builder-linux"
cp manifest.json "$out/"

# Stored, not deflated, with fixed dates and modes: a zip that only changes when its files do. Git compresses it.
python3 - "$out" ../anduin-builder.mcpb <<'EOF'
import os, sys, zipfile
src, dest = sys.argv[1], sys.argv[2]
with zipfile.ZipFile(dest, "w", zipfile.ZIP_STORED) as z:
    for name in ["manifest.json", "bin/anduin-builder", "bin/anduin-builder.exe", "bin/anduin-builder-linux"]:
        info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
        info.create_system = 3
        info.external_attr = (0o100755 if name.startswith("bin/") else 0o100644) << 16
        with open(os.path.join(src, name), "rb") as f:
            z.writestr(info, f.read())
EOF
