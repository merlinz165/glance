#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"
export PATH="$PWD/tools:$PATH"
mkdir -p personal-data
export CODEXBAR_CONFIG="$PWD/personal-data/codexbar.json"
printf '在本机终端安装官方 Codex 和 Claude Code 后登录：\n\n'
printf 'Codex 官方安装说明：https://developers.openai.com/codex/cli/\n'
printf 'Claude 官方安装说明：https://code.claude.com/docs/en/setup\n\n'
printf '已安装时，选择登录（不会运行模型任务）：\n0) 下载官方客户端到本目录（不会全局安装）\n1) Codex\n2) Claude\n3) 检查额度\n'
read -r choice
case "$choice" in
  0) bash ./install-clients.sh ;;
  1) command -v codex >/dev/null || { echo '请先按官方说明安装 Codex CLI'; exit 1; }; codex login --device-auth; ./CodexBarCLI config enable --provider codex ;;
  2) command -v claude >/dev/null || { echo '请先按官方说明安装 Claude Code'; exit 1; }; claude auth login; ./CodexBarCLI config enable --provider claude ;;
  3) ./CodexBarCLI dashboard --identity redacted ;;
  *) echo '已退出' ;;
esac
