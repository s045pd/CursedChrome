#!/bin/bash

# 聊天功能测试脚本

echo "=== Mac聊天功能测试 ==="
echo ""

# 测试本地消息功能
echo "1. 测试本地消息功能..."
if command -v write >/dev/null 2>&1; then
    echo "✓ write 命令可用"
else
    echo "✗ write 命令不可用"
fi

if command -v talk >/dev/null 2>&1; then
    echo "✓ talk 命令可用"
else
    echo "✗ talk 命令不可用"
fi

if command -v wall >/dev/null 2>&1; then
    echo "✓ wall 命令可用"
else
    echo "✗ wall 命令不可用"
fi

echo ""

# 测试消息接收状态
echo "2. 当前消息接收状态:"
mesg

echo ""

# 测试网络工具
echo "3. 测试网络工具..."
if command -v nc >/dev/null 2>&1; then
    echo "✓ netcat 可用"
else
    echo "✗ netcat 不可用"
fi

if command -v ssh >/dev/null 2>&1; then
    echo "✓ SSH 可用"
else
    echo "✗ SSH 不可用"
fi

echo ""

# 显示可用用户
echo "4. 当前登录用户:"
who

echo ""

# 显示网络接口
echo "5. 网络接口信息:"
ifconfig | grep "inet " | grep -v 127.0.0.1

echo ""
echo "测试完成！"





