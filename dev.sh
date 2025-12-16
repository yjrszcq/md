#!/bin/sh

# Build and start backend
cd ./md
go build
./md -p 9900 -log ./logs -data ./data -reg=true -ai_key=md-ai-encrypt-key-2024 &
BACKEND_PID=$!
cd ../

# Start frontend in dev mode
cd ./web
npm run dev

# Kill backend when frontend exits
kill $BACKEND_PID 2>/dev/null
