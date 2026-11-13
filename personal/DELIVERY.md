# 交付与验证

范围：独立中文 RSS 看板、本机凭据配置、DeepSeek 官方余额、OpenAI 官方月费用、CodexBar 订阅额度适配、Linux 可移植包和 Windows WSL 启动入口。

本地验证：Go 全量测试（含 race）、go vet、默认 YAML config:validate、Mac 原生构建运行、浏览器查看实际 RSS 和未配置状态、浏览器保存空凭据配置。API 使用 fixture 测试，未访问用户的真实账号。

Linux 目标通过 GitHub Actions 构建并启动打包后的程序，验证配置、CodexBar 可执行文件、配置页和 HTTP API。Windows WSL 未在真实 Windows 设备验证。首次真实账号登录及余额核对由用户在目标电脑完成。

OpenAI 真正充值余额暂为官方账单页面入口；不宣称费用等于余额。GitHub fork 默认关闭 Issues，因此不更改仓库设置，以 PR 承载交付记录。

Linux 打包启动已通过 Actions：https://github.com/merlinz165/glance/actions/runs/43423737367 。
