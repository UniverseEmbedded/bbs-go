$ErrorActionPreference = "Stop"

$baseUrl = "http://127.0.0.1:8082"

if (!(Test-Path ".\logs")) {
  New-Item -ItemType Directory -Path ".\logs" | Out-Null
}

try {
  $status = Invoke-RestMethod -Method Get -Uri "$baseUrl/api/install/status" -TimeoutSec 3
  if ($status.installed -eq $true) {
    Write-Host "已安装：installed=true"
    exit 0
  }
} catch {
  Write-Host "无法访问 $baseUrl，请先启动 bbs-go.exe（保持该窗口运行）。"
  throw
}

$body = @{
  siteTitle       = "bbs-go"
  siteDescription = "local"
  baseURL         = "/"
  dbConfig        = @{ type = "sqlite" }
  username        = "admin"
  password        = "admin123456"
  avatar          = ""
  language        = "zh-CN"
} | ConvertTo-Json

$resp = Invoke-RestMethod -Method Post -Uri "$baseUrl/api/install/install" -ContentType "application/json" -Body $body
if ($resp.success -ne $true) {
  throw ("安装失败：" + ($resp.message | Out-String))
}

$installed = $false
for ($i = 0; $i -lt 30; $i++) {
  Start-Sleep -Seconds 1
  $status2 = Invoke-RestMethod -Method Get -Uri "$baseUrl/api/install/status" -TimeoutSec 3
  if ($status2.data.installed -eq $true) {
    $installed = $true
    break
  }
}
if ($installed -ne $true) {
  throw "安装接口已返回成功，但 30 秒内 installed 仍为 false。请查看服务端日志。"
}

Write-Host "安装完成：installed=true"
Write-Host "请关闭并重新启动 bbs-go.exe，然后访问："
Write-Host "  Site  : $baseUrl/"
Write-Host "  Admin : $baseUrl/admin/"
