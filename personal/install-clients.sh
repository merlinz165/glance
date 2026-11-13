#!/usr/bin/env bash
# Optional first-run installation, scoped to this extracted directory.
set -euo pipefail
cd -- "$(dirname -- "$0")"
[ "$(uname -s)" = Linux ] && [ "$(uname -m)" = x86_64 ] || { echo '此包仅支持 Linux/WSL x86_64'; exit 1; }
command -v curl >/dev/null || { echo '需要 curl，请使用系统包管理器安装后重试'; exit 1; }
mkdir -p tools
if [ ! -x tools/codex ]; then
 curl -fL --retry 2 https://github.com/openai/codex/releases/download/rust-v0.169.0/codex-x86_64-unknown-linux-musl.tar.gz -o tools/codex-download.tar.gz
 tar -xzf tools/codex-download.tar.gz -C tools
 mv tools/codex-x86_64-unknown-linux-musl tools/codex
 chmod u+x tools/codex
fi
if [ ! -x tools/claude ]; then
 base=https://downloads.claude.ai/claude-code-releases
 version=$(curl -fsSL "$base/stable")
 [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo '无法取得 Claude 稳定版本'; exit 1; }
 manifest=$(curl -fsSL "$base/$version/manifest.json" | tr -d '\n\r')
 if [[ "$manifest" =~ \"linux-x64\"[[:space:]]*:[[:space:]]*\{[^\}]*\"checksum\"[[:space:]]*:[[:space:]]*\"([a-f0-9]{64})\" ]]; then
  expected="${BASH_REMATCH[1]}"
 else
  echo 'Claude 下载清单格式无法识别'; exit 1
 fi
 curl -fL --retry 2 "$base/$version/linux-x64/claude" -o tools/claude-download
 printf '%s  %s\n' "$expected" tools/claude-download | sha256sum --check -
 mv tools/claude-download tools/claude
 chmod u+x tools/claude
fi
printf '客户端已下载到本目录 tools。现在重新运行 bash setup-accounts.sh，选择登录。\n'
