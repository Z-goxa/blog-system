#!/bin/bash

echo "🚀 Blog System - 快速启动脚本"
echo "================================="

# 检查 MariaDB 是否运行
if ! systemctl is-active --quiet mariadb; then
    echo "⚠️  MariaDB 未运行，正在启动..."
    sudo systemctl start mariadb
fi

# 初始化数据库
echo "📥 初始化数据库..."
mysql -u root -p <<EOF
CREATE DATABASE IF NOT EXISTS blog_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE blog_db;
SOURCE sql/schema.sql;
EOF

# 后端依赖安装
echo "📦 安装后端依赖..."
cd backend
if [ ! -d "vendor" ]; then
    go mod tidy
    go mod vendor
fi

# 前端依赖安装
echo "📦 安装前端依赖..."
cd ../frontend
if [ ! -d "node_modules" ]; then
    npm install
fi

echo ""
echo "✅ 安装完成！"
echo ""
echo "🎯 启动方式："
echo "   1. 开发模式: cd backend && wails dev"
echo "   2. 生产构建: cd backend && wails build"
echo ""
echo "Happy Coding! 🐉"
