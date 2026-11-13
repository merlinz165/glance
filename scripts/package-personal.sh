#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
GO=${GO:-go}
out=dist/personal-dashboard-linux-amd64
mkdir -p "$out/personal"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "$GO" build -trimpath -o "$out/glance" .
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "$GO" build -trimpath -o "$out/personal-dashboard" ./cmd/personal-dashboard
cp personal/glance.yml "$out/personal/"
cp personal/*.sh personal/*.ps1 personal/README.zh-CN.md "$out/"
cp LICENSE "$out/LICENSE-Glance"
mkdir -p .tools
if [ ! -f .tools/CodexBarCLI-v0.84.0-linux-musl-x86_64.tar.gz ]; then
 curl -fL --retry 2 https://github.com/steipete/CodexBar/releases/download/v0.84.0/CodexBarCLI-v0.84.0-linux-musl-x86_64.tar.gz -o .tools/CodexBarCLI-v0.84.0-linux-musl-x86_64.tar.gz
fi
tar -xzf .tools/CodexBarCLI-v0.84.0-linux-musl-x86_64.tar.gz -C "$out"
curl -fLsS https://raw.githubusercontent.com/steipete/CodexBar/v0.84.0/LICENSE -o "$out/LICENSE-CodexBar"
chmod u+x "$out"/*.sh
COPYFILE_DISABLE=1 tar -czf dist/personal-dashboard-linux-amd64.tar.gz -C dist personal-dashboard-linux-amd64
printf '安装包：%s\n' "$PWD/dist/personal-dashboard-linux-amd64.tar.gz"
