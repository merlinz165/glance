#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"
chmod u+x ./personal-dashboard ./glance ./CodexBarCLI
export PATH="$PWD/tools:$PATH"
mkdir -p personal-data
export CODEXBAR_CONFIG="$PWD/personal-data/codexbar.json"
if [ ! -f "$CODEXBAR_CONFIG" ]; then
 ./CodexBarCLI config enable --provider codex
 ./CodexBarCLI config enable --provider claude
fi
exec ./personal-dashboard
