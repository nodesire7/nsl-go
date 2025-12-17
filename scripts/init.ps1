# PowerShell初始化脚本
# 用于Windows环境设置

Write-Host "正在初始化短链接系统..." -ForegroundColor Green

# 检查环境变量
if (-not $env:JWT_SECRET) {
    Write-Host "警告: JWT_SECRET 未设置，将自动生成一个随机密钥（生产环境请自行配置并妥善保存）" -ForegroundColor Yellow
    $bytes = New-Object byte[] 32
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:JWT_SECRET = [Convert]::ToHexString($bytes)
    Write-Host "生成的JWT_SECRET: $env:JWT_SECRET" -ForegroundColor Cyan
}

Write-Host "初始化完成！" -ForegroundColor Green

