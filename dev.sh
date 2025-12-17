#!/bin/sh

# Ensure web directory exists for go:embed (can be empty in dev mode)
mkdir -p ./md/web

# Build and start backend
cd ./md
go build
./md -p 9900 -log ./logs -data ./data -reg=true -ai_key=md-ai-encrypt-key-2024 &
BACKEND_PID=$!
cd ../

# Install frontend dependencies if needed
cd ./web
if [ ! -d "node_modules" ]; then
  echo "Installing frontend dependencies..."
  npm install
fi

# Start frontend in dev mode
npm run dev

# Kill backend when frontend exits
kill $BACKEND_PID 2>/dev/null
