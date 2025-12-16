#!/bin/bash
# 一键安装脚本

set -e

echo "=========================================="
echo "  New short link (NSL GO) 一键安装脚本"
echo "=========================================="

# 检测系统架构
ARCH=$(uname -m)
OS=$(uname -s)

if [ "$OS" = "Linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        PLATFORM="linux-amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        PLATFORM="linux-arm64"
    else
        echo "不支持的架构: $ARCH"
        exit 1
    fi
elif [ "$OS" = "Darwin" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        PLATFORM="darwin-amd64"
    elif [ "$ARCH" = "arm64" ]; then
        PLATFORM="darwin-arm64"
    else
        echo "不支持的架构: $ARCH"
        exit 1
    fi
else
    echo "不支持的操作系统: $OS"
    exit 1
fi

# GitHub仓库
GITHUB_REPO="${GITHUB_REPO:-nodesire7/nsl-go}"

# 获取最新版本
echo "正在获取最新版本..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo "无法获取最新版本，使用默认版本"
    LATEST_VERSION="latest"
fi

# 下载文件
FILE_NAME="nsl-go-${PLATFORM}.tar.gz"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/${FILE_NAME}"

echo "正在下载: $DOWNLOAD_URL"
curl -L -o "$FILE_NAME" "$DOWNLOAD_URL"

# 解压
echo "正在解压..."
tar -xzf "$FILE_NAME"

# 移动到系统目录
INSTALL_DIR="/usr/local/bin"
echo "正在安装到 $INSTALL_DIR..."
sudo cp nsl-go "$INSTALL_DIR/nsl-go"
sudo chmod +x "$INSTALL_DIR/nsl-go"

# 创建配置目录
CONFIG_DIR="$HOME/.nsl-go"
mkdir -p "$CONFIG_DIR"
if [ -d "web" ]; then
    cp -r web "$CONFIG_DIR/"
fi

# 清理
rm -f "$FILE_NAME"
rm -rf release-*

echo ""
echo "=========================================="
echo "  安装完成！"
echo "=========================================="
echo "  可执行文件: $INSTALL_DIR/nsl-go"
echo "  配置目录: $CONFIG_DIR"
echo ""
echo "  运行命令: nsl-go"
echo "=========================================="

