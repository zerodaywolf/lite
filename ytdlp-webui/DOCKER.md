# Docker Deployment

This project uses a lightweight Python Docker image for easy deployment with minimal setup.

## Lightweight Python Image

This Docker image uses `python:3.11-slim` as the base, which provides:
- **Minimal** - Small footprint while including pip and essential tools
- **Simple** - Single-stage build with direct pip installation
- **Complete** - All necessary dependencies in one place

This image includes ffmpeg for full audio/video processing capabilities, enabling:
- MP3/FLAC/M4A audio extraction
- Video format conversion
- Audio quality optimization
- All yt-dlp post-processing features

## Building the Image

### Using the build script (recommended)
```bash
chmod +x build.sh
./build.sh
```

### Manual build
```bash
docker build -t ytdlp-webui:latest .
```

## Running the Container

### Using Docker directly
```bash
docker run -d -p 8000:8000 --name ytdlp-webui ytdlp-webui:latest
```

### With persistent downloads volume
```bash
docker run -d -p 8000:8000 -v $(pwd)/downloads:/app/downloads --name ytdlp-webui ytdlp-webui:latest
```

## Accessing the Application

Once running, access the web interface at:
- http://localhost:8000

## Image Details

- **Base Image**: `python:3.11-slim`
- **Build Type**: Single-stage build with direct pip installation
- **Includes**: Python 3.11, yt-dlp, Flask, and ffmpeg for audio/video processing
- **Security**: Minimal Linux packages, only essential dependencies
- **Size**: Lightweight while including all necessary tools
- **Simplicity**: Direct pip install in final image, no complex multi-stage copying

## Troubleshooting

### Viewing logs
```bash
docker logs ytdlp-webui
```

### Health check
The container includes a health check that verifies the Flask app is responding.

### Security

The lightweight image provides enhanced security by:
- Minimal base system (python:3.11-slim)
- Only essential packages installed
- Clean package cache removal
- Read-only filesystem compatibility

### Input Validation & Security

The application includes comprehensive security measures to prevent RCE attacks:
- **URL Validation**: Only allows HTTP/HTTPS URLs from supported video platforms
- **Domain Allowlist**: Restricts downloads to trusted video platforms (YouTube, Vimeo, etc.)
- **Input Sanitization**: Removes dangerous characters and limits input length
- **Format ID Validation**: Ensures format IDs match expected patterns
- **Command Injection Prevention**: Uses subprocess lists instead of shell execution
- **Request Validation**: Validates all JSON inputs and parameter types
- **Rate Limiting**: Prevents abuse with IP-based request limiting (20 info requests, 5 downloads per minute)
- **Content-Type Validation**: Ensures only JSON requests are processed

## Development

For development with shell access, you can use the builder stage:
```bash
docker build --target builder -t ytdlp-webui:dev .
docker run -it ytdlp-webui:dev /bin/bash
``` 