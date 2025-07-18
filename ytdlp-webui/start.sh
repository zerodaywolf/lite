#!/bin/bash

# YouTube-dl Web UI Startup Script

echo "🎬 Starting YouTube-dl Web UI..."

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "❌ Virtual environment not found. Creating it..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Install/upgrade dependencies
echo "📦 Installing dependencies..."
pip install --quiet --upgrade pip
pip install --quiet Flask yt-dlp

# Ensure yt-dlp is available and up to date
echo "🔄 Updating yt-dlp to latest version..."
pip install --quiet --upgrade yt-dlp

if ! command -v yt-dlp &> /dev/null; then
    echo "❌ Error: yt-dlp installation failed!"
    exit 1
fi

echo "✅ Using yt-dlp version: $(yt-dlp --version)"

# Create downloads directory
mkdir -p downloads

echo "🚀 Starting web server..."
echo "📱 Open your browser and go to: http://localhost:5000"
echo "🛑 Press Ctrl+C to stop the server"
echo ""

# Start the Flask application
python app.py 