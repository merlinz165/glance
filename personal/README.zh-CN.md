# 我的独立资讯屏

这是一份 Glance 的中文个人发行配置，附带本机配置服务和额度采集器。Mac 不参与运行。不含 iPhone、闲鱼、小红书接入。

## Linux（推荐，Intel / AMD 64 位）

解压安装包到你自己的目录，终端进入目录运行 `bash start.sh`。
打开 http://127.0.0.1:8080 。账号设置：http://127.0.0.1:8091 。
无需 Go、Python、Node 或 Docker。关闭启动窗口或按 Ctrl+C 停止。

## Windows 11

先安装 WSL Ubuntu：管理员 PowerShell 运行 `wsl --install -d Ubuntu`，按系统提示重启，并打开 Ubuntu 完成用户名设置。WSL 与 Windows 同时运行，无需切换已有的 Linux 双系统。
解压同一个 Linux 安装包，在 PowerShell 中运行：

```powershell
powershell -ExecutionPolicy Bypass -File .\start-windows.ps1
```

然后在 Windows 浏览器访问同样的两个 localhost 地址。首次账户登录在 Ubuntu 终端运行 `bash setup-accounts.sh`。WSL 安装需要管理员权限，此发行包不自动改动系统或设置开机服务。

## 账号配置

新闻开箱即用，无需任何凭据。配置页可随时保存 DeepSeek API Key 和 OpenAI **组织 Admin API Key**。密码框留空保留原值，输入 `-` 删除。凭据以本地明文 JSON 存放在 `personal-data/settings.json`（Linux 下权限 0600）；不要分享这个目录或把它提交 Git。在 Windows 挂载目录中由 Windows 文件权限保护。

CodexBar CLI 已随包附带。Codex 和 Claude 需要在这台电脑、同一个 Linux/WSL 用户环境独立登录；账号必须与平常使用的账号相同。运行 `bash setup-accounts.sh` 查看官方安装入口及登录菜单。完成后运行 `./CodexBarCLI dashboard --identity redacted` 检查。若提供商没有启用，请按 CodexBar 官方 CLI 文档配置： https://github.com/steipete/CodexBar/blob/v0.84.0/docs/cli-configuration.md 。

- Codex：账号服务返回的额度窗口。不是普通 ChatGPT 聊天所有模型的剩余次数。
- Claude：订阅用量。依赖该账号的登录有效性及平台接口。
- DeepSeek：官方余额接口，分别列出币种、充值和赠金。
- OpenAI：官方 Costs API 的 UTC 当前月份费用，支持分页。**费用不是余额**。真实余额通过配置页链接打开官方账单页面查看；本项目不使用不稳定的旧余额接口。

采集器每 5 分钟更新。页面也每 5 分钟刷新；显示各数据采集时间。错误/未配置单独显示，旧订阅数据超过 15 分钟不显示剩余数字。正常采集不发送推理请求。

## 配置与端口

`personal/glance.yml` 可编辑 RSS、栏目和刷新频率；端口 8080 是看板，8091 是配置与采集服务，仅监听本机。更换端口需同步修改配置和启动程序源码。

## 更新、退出与数据

保留 `personal-data` 和个人修改的 `personal/glance.yml`，在新目录解压新版后迁移它们。不要直接覆盖正在运行的安装目录。账号登录状态由官方 CLI 在当前用户目录管理。

## 来源与许可

Glance：https://github.com/glanceapp/glance （AGPL-3.0）。本 fork：https://github.com/merlinz165/glance 。
CodexBar v0.84.0：https://github.com/steipete/CodexBar （MIT；随包 LICENSE-CodexBar）。使用原版 Linux CLI，不维护额外 fork。配置与 companion 源码见本仓库 `personal` 和 `cmd/personal-dashboard`。
