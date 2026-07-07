#!/usr/bin/env bash
# Regenerate assets/AppIcon.icns from tools/appicon (macOS only: needs sips + iconutil).
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

go run ./tools/appicon 1024 "$tmp/icon_1024.png"

set="$tmp/AppIcon.iconset"
mkdir -p "$set" assets
while read -r sz name; do
	sips -z "$sz" "$sz" "$tmp/icon_1024.png" --out "$set/icon_${name}.png" >/dev/null
done <<'SIZES'
16 16x16
32 16x16@2x
32 32x32
64 32x32@2x
128 128x128
256 128x128@2x
256 256x256
512 256x256@2x
512 512x512
1024 512x512@2x
SIZES

iconutil -c icns "$set" -o assets/AppIcon.icns
echo "→ assets/AppIcon.icns"
