#!/bin/bash

echo "🚀 Running ytdlp-webui container..."

# Stop any existing container
docker stop ytdlp-webui 2>/dev/null || true
docker rm ytdlp-webui 2>/dev/null || true

# Run the container
docker run -d \
    --name ytdlp-webui \
    -p 8000:8000 \
    ytdlp-webui:latest

echo "✅ Container started!"
echo "🌐 Open http://localhost:8000 in your browser"
echo "📋 View logs: docker logs ytdlp-webui"
echo "🛑 Stop container: docker stop ytdlp-webui" 