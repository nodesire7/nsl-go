#!/bin/bash
# 初始化脚本
# 用于设置环境和初始化数据库

echo "正在初始化短链接系统..."

# 检查环境变量
if [ -z "$JWT_SECRET" ]; then
    echo "警告: JWT_SECRET 未设置，将自动生成一个随机密钥（生产环境请自行配置并妥善保存）"
    export JWT_SECRET=$(openssl rand -hex 32)
    echo "生成的JWT_SECRET: $JWT_SECRET"
fi

# 等待数据库就绪
echo "等待PostgreSQL就绪..."
until pg_isready -h ${DB_HOST:-localhost} -p ${DB_PORT:-5432} -U ${DB_USER:-postgres}; do
    echo "等待数据库连接..."
    sleep 2
done

echo "初始化完成！"

