#!/bin/bash

# 简单实时聊天工具
# 使用方法: ./simple-chat.sh [目标IP] [用户名]

TARGET_IP=${1:-"192.168.1.100"}
USERNAME=${2:-$(whoami)}

echo "=== 简单实时聊天 ==="
echo "目标: $USERNAME@$TARGET_IP"
echo "输入 'exit' 退出"
echo ""

# 确保可以接收消息
mesg y

# 创建聊天日志文件
CHAT_LOG="$HOME/.simple_chat_$(date +%Y%m%d_%H%M%S).log"

echo "聊天开始时间: $(date)" > "$CHAT_LOG"
echo "聊天日志: $CHAT_LOG"
echo ""

while true; do
    read -p "你: " message
    
    if [ "$message" = "exit" ]; then
        echo "退出聊天"
        break
    fi
    
    # 记录到本地日志
    echo "[$(date '+%H:%M:%S')] 你: $message" >> "$CHAT_LOG"
    
    # 发送到远程Mac
    echo "[$(date '+%H:%M:%S')] $(whoami): $message" | ssh $USERNAME@$TARGET_IP "cat >> ~/.simple_chat_$(date +%Y%m%d_%H%M%S).log"
    
    # 尝试使用write命令实时显示
    echo "$message" | ssh $USERNAME@$TARGET_IP "write $(whoami)@$(hostname)" 2>/dev/null || true
    
    echo "消息已发送"
done

echo "聊天结束，日志保存在: $CHAT_LOG"





