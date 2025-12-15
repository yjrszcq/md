#!/bin/sh

# 构建前端
cd ./web
npm install
npm run build
cd ../

# 复制前端到后端目录
rm -rf ./md/web
cp -r ./web/dist ./md/web

# 构建后端
cd ./md
go build
echo "md build finished"

# 本地运行
./md -p 9900 -log ./logs -data ./data -reg=true -ai_key=md-ai-encrypt-key-2024
