# yt-dlp Web UI

A simple, modern web interface for [yt-dlp](https://github.com/yt-dlp/yt-dlp) that allows you to download videos from YouTube and other supported sites through a user-friendly web browser interface.

> **Note**: This uses yt-dlp, a more actively maintained and feature-rich fork of youtube-dl with better YouTube support and regular updates.

## Features

- 🎥 **Video Analysis**: Automatically extract video information and available formats
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile devices
- ⚡ **Real-time Progress**: Live download progress tracking
- 💾 **Direct Browser Downloads**: Files download directly to your device - no server storage
- 🎵 **High-Quality Audio**: MP3 (320kbps), FLAC (lossless), M4A, and original formats
- 🎨 **Modern UI**: Clean, intuitive interface with smooth animations
- 🔧 **Format Selection**: Choose from available video qualities and formats

## Installation

### Prerequisites

- Python 3.7 or higher  
- yt-dlp (will be installed via requirements.txt)

### Setup

1. **Clone or download this repository**:
   ```bash
   git clone <repository-url>
   cd ytdl
   ```

2. **Install Python dependencies**:
   ```bash
   pip install -r requirements.txt
   ```

3. **Install yt-dlp** (if not already installed):
   ```bash
   pip install yt-dlp
   ```
   
   Or update to the latest version:
   ```bash
   pip install --upgrade yt-dlp
   ```

## Usage

1. **Start the web server**:
   ```bash
   python app.py
   ```

2. **Open your web browser** and navigate to:
   ```
   http://localhost:5000
   ```

3. **Download videos**:
   - Paste a YouTube URL (or any supported site URL) into the input field
   - Click "Analyze" to extract video information and available formats
   - Select your preferred quality/format and download type
   - Click "Download" to start the download
   - Monitor progress in real-time
   - **Files download directly to your browser** - no server storage needed!

## Supported Sites

This web UI supports all sites that yt-dlp supports, including:

- YouTube (with enhanced support)
- Vimeo
- Facebook
- Twitter
- Instagram
- Twitch
- TikTok
- And many more (1900+ sites)

For a complete list, visit: [yt-dlp supported sites](https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md)

## Configuration

### Download Behavior

Files are downloaded directly to your browser's default download folder. No files are permanently stored on the server - they are processed temporarily and then automatically cleaned up.

### Server Configuration

You can modify the server settings in `app.py`:

- **Host**: Change `host='0.0.0.0'` to `host='127.0.0.1'` for local-only access
- **Port**: Change `port=5000` to your preferred port
- **Debug**: Set `debug=False` for production use

### yt-dlp Options

You can customize yt-dlp behavior by modifying the command-line options in the `download_video()` function in `app.py`.

## API Endpoints

The web UI also provides a REST API:

- `GET /` - Main web interface
- `POST /extract_info` - Extract video information
- `POST /download` - Start video download
- `GET /progress/<download_id>` - Get download progress
- `GET /downloads` - List downloaded files

## Troubleshooting

### Common Issues

1. **"yt-dlp command not found"**:
   - Make sure yt-dlp is installed: `pip install yt-dlp`
   - Try updating: `pip install --upgrade yt-dlp`

2. **Download fails with "Unable to extract video info"**:
   - The URL might not be supported
   - yt-dlp might need updating (it's actively maintained)
   - Check if the video is private or restricted

3. **Browser download issues**:
   - Check your browser's download settings
   - Make sure downloads aren't blocked by browser security settings

4. **Slow download speeds**:
   - This depends on your internet connection and the source server
   - Consider selecting a lower quality format

### Updating yt-dlp

yt-dlp is actively maintained and regularly updated to support changes in video sites. Update it regularly:

```bash
pip install --upgrade yt-dlp
```

## Security Considerations

- This application is intended for personal use
- Be careful when exposing it to the internet (consider authentication)
- Only download content you have the right to download
- Respect copyright and terms of service of video platforms

## Contributing

Feel free to submit issues, feature requests, or pull requests to improve this web UI.

## License

This project is open source. Please respect the licenses of youtube-dl and other dependencies.

---

**Note**: This web UI is a wrapper around yt-dlp. All download capabilities and site support come from the excellent yt-dlp project, which is a more actively maintained and feature-rich fork of youtube-dl. 