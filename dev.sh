#!/bin/sh

# Ensure web directory has a placeholder for go:embed (ignores dot files)
mkdir -p ./md/web
[ -z "$(ls -A ./md/web 2>/dev/null)" ] && echo "dev" > ./md/web/placeholder.txt

# Build and start backend
cd ./md
echo "Building backend..."
if ! go build; then
  echo "Backend build failed!"
  exit 1
fi

echo "Starting backend on port 9900..."
./md -p 9900 -log ./logs -data ./data -reg=true -ai_key=md-ai-encrypt-key-2024 &
BACKEND_PID=$!

# Wait for backend to be ready (check if port is listening)
echo "Waiting for backend to start..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  if nc -z 127.0.0.1 9900 2>/dev/null; then
    echo "Backend is ready!"
    break
  fi
  if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo "Backend process died!"
    exit 1
  fi
  sleep 1
done

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
