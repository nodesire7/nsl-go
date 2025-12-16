#!/bin/bash
# 初始化脚本
# 用于设置环境和初始化数据库

echo "正在初始化短链接系统..."

# 检查环境变量
if [ -z "$API_TOKEN" ]; then
    echo "警告: API_TOKEN 未设置，将使用默认值"
    export API_TOKEN=$(openssl rand -hex 32)
    echo "生成的API Token: $API_TOKEN"
fi

# 等待数据库就绪
echo "等待PostgreSQL就绪..."
until pg_isready -h ${DB_HOST:-localhost} -p ${DB_PORT:-5432} -U ${DB_USER:-postgres}; do
    echo "等待数据库连接..."
    sleep 2
done

echo "初始化完成！"

