#!/bin/bash

# 定义目标操作系统和架构的组合
declare -a targets=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
)

# 获取当前目录
current_dir=$(pwd)

# 遍历所有目标组合
for target in "${targets[@]}"; do
    # 分割操作系统和架构
    GOOS=${target%/*}
    GOARCH=${target#*/}

    # 设置环境变量
    export GOOS
    export GOARCH

    # 定义输出文件名
    output_name="power-amplifier-$GOOS-$GOARCH"
    if [ "$GOOS" = "windows" ]; then
        output_name+=".exe"
    fi

    # 编译代码
    echo "Building for $GOOS/$GOARCH..."
    go build -o "$current_dir/buildProd/$output_name" main.go

    # 检查编译是否成功
    if [ $? -eq 0 ]; then
        echo "Build successful: $output_name"
    else
        echo "Build failed for $GOOS/$GOARCH"
    fi
done

# 恢复默认环境变量
unset GOOS
unset GOARCH

echo "编译完成"
