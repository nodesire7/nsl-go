# PowerShell初始化脚本
# 用于Windows环境设置

Write-Host "正在初始化短链接系统..." -ForegroundColor Green

# 检查环境变量
if (-not $env:API_TOKEN) {
    Write-Host "警告: API_TOKEN 未设置，将使用默认值" -ForegroundColor Yellow
    $bytes = New-Object byte[] 32
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:API_TOKEN = [Convert]::ToHexString($bytes)
    Write-Host "生成的API Token: $env:API_TOKEN" -ForegroundColor Cyan
}

Write-Host "初始化完成！" -ForegroundColor Green

