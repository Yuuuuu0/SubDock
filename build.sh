#!/bin/bash

# SubDock 本地开发编译脚本
# 自动打包前端 + 后端

set -e  # 遇到错误立即退出

echo "========================================="
echo "  SubDock 本地编译脚本"
echo "========================================="

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}[1/5] 检查依赖...${NC}"
    
    if ! command -v node &> /dev/null; then
        echo -e "${RED}错误: 未找到 node，请先安装 Node.js${NC}"
        exit 1
    fi
    
    if ! command -v pnpm &> /dev/null; then
        echo -e "${RED}错误: 未找到 pnpm，请先安装 pnpm (npm install -g pnpm)${NC}"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: 未找到 go，请先安装 Go${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ 依赖检查通过${NC}"
}

# 构建前端
build_frontend() {
    echo ""
    echo -e "${BLUE}[2/5] 构建前端...${NC}"
    
    cd web
    
    # 安装依赖（如果需要）
    if [ ! -d "node_modules" ]; then
        echo "安装前端依赖..."
        pnpm install
    fi
    
    # 构建
    echo "编译 Vue 项目..."
    pnpm build
    
    cd ..
    echo -e "${GREEN}✓ 前端构建完成${NC}"
}

# 复制前端产物到后端
copy_frontend_dist() {
    echo ""
    echo -e "${BLUE}[3/5] 复制前端产物...${NC}"
    
    # 删除旧的 dist
    if [ -d "internal/router/dist" ]; then
        rm -rf internal/router/dist
        echo "已删除旧的前端产物"
    fi
    
    # 复制新的 dist
    cp -r web/dist internal/router/
    echo -e "${GREEN}✓ 前端产物已复制到 internal/router/dist${NC}"
}

# 构建后端
build_backend() {
    echo ""
    echo -e "${BLUE}[4/5] 构建后端...${NC}"
    
    # 下载依赖（如果需要）
    echo "下载 Go 依赖..."
    go mod download
    
    # 编译
    echo "编译 Go 项目..."
    go build -ldflags="-s -w" -o subdock .
    
    echo -e "${GREEN}✓ 后端构建完成${NC}"
}

# 显示结果
show_result() {
    echo ""
    echo -e "${BLUE}[5/5] 构建完成${NC}"
    echo "========================================="
    echo -e "${GREEN}✓ 编译成功！${NC}"
    echo ""
    echo "可执行文件: ./subdock"
    echo ""
    echo "运行方式:"
    echo "  DATA_DIR=./data ./subdock"
    echo ""
    echo "或者设置环境变量后运行:"
    echo "  export DATA_DIR=./data"
    echo "  export PORT=8080"
    echo "  ./subdock"
    echo "========================================="
}

# 主流程
main() {
    check_dependencies
    build_frontend
    copy_frontend_dist
    build_backend
    show_result
}

# 错误处理
trap 'echo -e "${RED}构建失败！${NC}"; exit 1' ERR

# 执行
main
