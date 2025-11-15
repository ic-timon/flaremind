#!/bin/bash
# 集成测试脚本 - 测试多个中文网站的爬取功能

echo "=========================================="
echo "FlareMind 集成测试"
echo "=========================================="

# 编译
echo "编译 FlareMind..."
go build -o flaremind.exe cmd/cli/main.go
if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

# 测试百度新闻
echo ""
echo "测试 1: 百度新闻"
echo "----------------------------------------"
./flaremind.exe -url https://news.baidu.com/ -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/baidu_news
echo ""

# 测试 Bilibili
echo "测试 2: Bilibili 热门"
echo "----------------------------------------"
./flaremind.exe -url https://www.bilibili.com/v/popular -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/bilibili
echo ""

# 测试抖音
echo "测试 3: 抖音"
echo "----------------------------------------"
./flaremind.exe -url https://www.douyin.com/ -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/douyin
echo ""

# 测试小红书
echo "测试 4: 小红书"
echo "----------------------------------------"
./flaremind.exe -url https://www.xiaohongshu.com/explore -depth 1 -pages 2 -rate 1 -delay 2000 -o test_output/xiaohongshu
echo ""

echo "=========================================="
echo "所有测试完成！"
echo "结果保存在 test_output/ 目录下"
echo "=========================================="

