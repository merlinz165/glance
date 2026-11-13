$ErrorActionPreference = 'Stop'
if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) { throw '请先安装 WSL：以管理员身份运行 wsl --install，然后重启电脑。' }
$linuxPath = (& wsl.exe --exec wslpath -a $PSScriptRoot).Trim()
if ($LASTEXITCODE -ne 0) { throw '请先启动 Ubuntu 完成 WSL 用户初始化。' }
Write-Host '启动后在浏览器打开 http://127.0.0.1:8080；账号配置 http://127.0.0.1:8091'
& wsl.exe --cd $linuxPath --exec bash ./start.sh
