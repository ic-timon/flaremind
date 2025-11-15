# FlareMind 集成测试脚本 - PowerShell 版本
# 测试多个中文网站的爬取功能

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "FlareMind 集成测试" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 编译
Write-Host "`n编译 FlareMind..." -ForegroundColor Yellow
go build -o flaremind.exe cmd/cli/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "编译失败" -ForegroundColor Red
    exit 1
}

# 创建输出目录
New-Item -ItemType Directory -Force -Path test_output | Out-Null

# 测试百度新闻
Write-Host "`n测试 1: 百度新闻" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
.\flaremind.exe -url https://news.baidu.com/ -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/baidu_news
Write-Host ""

# 测试 Bilibili
Write-Host "测试 2: Bilibili 热门" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
.\flaremind.exe -url https://www.bilibili.com/v/popular -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/bilibili
Write-Host ""

# 测试抖音
Write-Host "测试 3: 抖音" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
.\flaremind.exe -url https://www.douyin.com/ -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/douyin
Write-Host ""

# 测试小红书
Write-Host "测试 4: 小红书" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
.\flaremind.exe -url https://www.xiaohongshu.com/explore -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/xiaohongshu
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "所有测试完成！" -ForegroundColor Cyan
Write-Host "结果保存在 test_output/ 目录下" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

# 显示结果统计
Write-Host "`n结果统计:" -ForegroundColor Yellow
Get-ChildItem -Path test_output -Recurse -File -Filter *.md | 
    Group-Object DirectoryName | 
    ForEach-Object {
        Write-Host "  $($_.Name): $($_.Count) 个文件" -ForegroundColor White
    }

