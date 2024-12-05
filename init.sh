#!/bin/bash

# 定义日志和缓存目录
LOG_FILE="app.log"
CACHE_DIR="audio_cache"
DESKTOP_FILE="$HOME/Desktop/PowerAmplifier.lnk"

# 定义颜色
red="\033[31m"
green="\033[32m"
yellow="\033[33m"
blue="\033[34m"
reset="\033[0m"

# 是否是第一次启动
first_run=true

# 显示版权提示
show_copyright() {
    echo -e "${yellow}版权所有 (c) 2023 开源熊猫 @KY-PanDa${reset}"
    echo -e "${yellow}本应用主要用于升级传统模拟音箱、功放等设备，${reset}"
    echo -e "${yellow}提供音频播放、远程喊话等功能。${reset}"
    echo -e "${yellow}该应用基于 Go 语言编写，支持跨系统运行。${reset}"
    echo ""
}

# 清空音频缓存与日志
clear_cache_and_logs() {
    echo -e "${blue}清空音频缓存与日志...${reset}"

    # 检查缓存目录
    if [ -d "$CACHE_DIR" ]; then
        echo -e "${blue}正在清空缓存目录: $CACHE_DIR${reset}"
        rm -rf "$CACHE_DIR"/*
        echo -e "${green}缓存已清空。${reset}"
    else
        echo -e "${red}缓存目录不存在: $CACHE_DIR${reset}"
    fi

    # 检查日志文件
    if [ -f "$LOG_FILE" ]; then
        echo -e "${blue}正在删除日志文件: $LOG_FILE${reset}"
        rm -f "$LOG_FILE"
        echo -e "${green}日志文件已删除。${reset}"
    else
        echo -e "${red}日志文件不存在: $LOG_FILE${reset}"
    fi
}

# 设置开机启动
set_autostart() {
    local current_dir="$(pwd -W)"  # 使用 -W 获取 Windows 样式的路径
    echo -e "${blue}在 Windows 上设置开机启动...${reset}"

    # 确定快捷方式路径，使用 PowerShell 环境变量
    local startup_folder="$(powershell -Command "echo $env:APPDATA\\Microsoft\\Windows\\Start Menu\\Programs\\Startup")"
    local shortcut_path="$startup_folder\\PowerAmplifier.lnk"
    local target_path="$current_dir\\PowerAmplifier.exe"

    # 创建快捷方式目录（如果不存在）
    powershell -Command "if (-not (Test-Path '$startup_folder')) { New-Item -ItemType Directory -Path '$startup_folder' }"

    # 创建快捷方式
    powershell -ExecutionPolicy Bypass -File create_shortcut.ps1 -shortcutPath "$shortcut_path" -targetPath "$target_path"

    echo -e "${green}开机启动已设置。${reset}"
}





# 移除开机启动
remove_autostart() {
    echo -e "${blue}在 Windows 上移除开机启动...${reset}"

    # 确定快捷方式路径
    local startup_folder="$env:APPDATA\\Microsoft\\Windows\\Start Menu\\Programs\\Startup"
    local shortcut_path="$startup_folder\\PowerAmplifier.lnk"

    # 检查快捷方式是否存在并删除
    if [ -e "$shortcut_path" ]; then
        powershell -Command "Remove-Item '$shortcut_path'"
        echo -e "${green}开机启动已关闭。${reset}"
    else
        echo -e "${red}快捷方式不存在: $shortcut_path${reset}"
    fi
}


# 获取并展示 MAC 地址
show_mac_address() {
    if [ -f "cache.txt" ]; then
        local mac_address=$(cat cache.txt | tr -d '[:space:]')  # 读取并去除空格
        echo -e "${blue}您的设备序列号为：$mac_address${reset}"
    else
        echo -e "${red}未找到 cache.txt 文件。请先运行程序以获取设备序列号。${reset}"
    fi
}

# 运行 PowerAmplifier.exe
run_power_amplifier() {
    echo -e "${blue}正在运行 PowerAmplifier.exe...${reset}"
    ./"PowerAmplifier.exe" &  # 在后台运行 PowerAmplifier.exe
    power_amplifier_pid=$!    # 获取进程 ID
}

# 结束运行的 PowerAmplifier
exit_program() {
    if [ -n "$power_amplifier_pid" ]; then
        echo -e "${yellow}正在结束 PowerAmplifier.exe...${reset}"
        kill $power_amplifier_pid
        echo -e "${green}PowerAmplifier.exe 已结束。${reset}"
    else
        echo -e "${red}PowerAmplifier.exe 未在运行。${reset}"
    fi
}

# 主循环
while true; do
    if $first_run; then
        show_copyright
        first_run=false
    fi

    echo -e "${yellow}请输入序号选择功能:${reset}"
    echo -e "${red}1. 开启开机启动${reset}    ${green}2. 关闭开机启动${reset}"
    echo -e "${red}3. 清空音频缓存与日志${reset}    ${green}4. 设备绑定${reset}    ${red}5. 运行 PowerAmplifier${reset}"
    echo -e "${green}6. 结束 PowerAmplifier 程序${reset}    ${red}0. 退出主程序${reset}"

    read -p "请输入选择: " choice
    case $choice in
        1) set_autostart ;;
        2) remove_autostart ;;
        3) clear_cache_and_logs ;;
        4) show_mac_address ;;
        5) run_power_amplifier ;;
        6) exit_program ;;  # 调用结束 PowerAmplifier 功能
        0) echo -e "${yellow}退出主程序${reset}"; break ;;  # 退出主程序
        *) echo -e "${red}无效的选择，请再次尝试。${reset}" ;;
    esac
done

