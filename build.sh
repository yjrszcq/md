#!/bin/sh

GO_BIN="${GO_BIN:-$HOME/sdk/go1.27.3/bin/go}"
if [ ! -x "$GO_BIN" ]; then
  echo "Go 1.27.3 SDK not found: $GO_BIN"
  exit 1
fi

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
"$GO_BIN" build
echo "md build finished"

# 本地运行
./md -p 9900 -log ./logs -data ./data -reg=true -ai_key=md-ai-encrypt-key-2024
