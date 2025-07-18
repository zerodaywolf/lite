from flask import Flask, render_template, request, jsonify, send_file
import subprocess
import os
import json
import re
from urllib.parse import urlparse
import tempfile
import threading
import time
import uuid
from werkzeug.utils import secure_filename

app = Flask(__name__)

# Store download progress
download_progress = {}

# Rate limiting storage (simple in-memory)
request_timestamps = {}

# Security configuration
ALLOWED_URL_SCHEMES = ['http', 'https']
ALLOWED_DOMAINS = [
    'youtube.com', 'www.youtube.com', 'youtu.be', 'm.youtube.com',
    'vimeo.com', 'dailymotion.com', 'twitch.tv', 'soundcloud.com',
    'twitter.com', 'x.com', 'tiktok.com', 'instagram.com',
    'facebook.com', 'reddit.com', 'streamable.com'
]
ALLOWED_DOWNLOAD_TYPES = ['video', 'audio', 'mp3', 'flac', 'm4a']
MAX_URL_LENGTH = 2048

def validate_url(url):
    """Validate URL against RCE and ensure it's from allowed domains"""
    if not url or not isinstance(url, str):
        return False
    
    # Check length
    if len(url) > MAX_URL_LENGTH:
        return False
    
    # Check for dangerous characters that could be used for command injection
    dangerous_chars = ['&', '|', ';', '`', '$', '(', ')', '<', '>', '"', "'", '\\', '\n', '\r']
    if any(char in url for char in dangerous_chars):
        return False
    
    try:
        parsed = urlparse(url)
        
        # Check scheme
        if parsed.scheme not in ALLOWED_URL_SCHEMES:
            return False
        
        # Check if domain is allowed
        domain = parsed.netloc.lower()
        # Remove port if present
        if ':' in domain:
            domain = domain.split(':')[0]
        
        # Check if domain is in allowed list or is a subdomain of allowed domains
        is_allowed = False
        for allowed_domain in ALLOWED_DOMAINS:
            if domain == allowed_domain or domain.endswith('.' + allowed_domain):
                is_allowed = True
                break
        
        return is_allowed
        
    except Exception:
        return False

def validate_format_id(format_id):
    """Validate format ID to prevent command injection"""
    if not format_id or not isinstance(format_id, str):
        return False
    
    # Allow 'best' as a special case
    if format_id == 'best':
        return True
    
    # Format IDs should be alphanumeric with some allowed special characters
    # Based on yt-dlp format ID patterns
    allowed_pattern = re.compile(r'^[a-zA-Z0-9_\-+]+$')
    
    # Check length (reasonable limit)
    if len(format_id) > 50:
        return False
    
    return bool(allowed_pattern.match(format_id))

def validate_download_type(download_type):
    """Validate download type against whitelist"""
    return download_type in ALLOWED_DOWNLOAD_TYPES

def sanitize_user_input(user_input):
    """Additional sanitization for user inputs"""
    if not isinstance(user_input, str):
        return ""
    
    # Remove any null bytes and control characters
    sanitized = user_input.replace('\x00', '').replace('\r', '').replace('\n', '')
    
    # Limit length
    return sanitized[:MAX_URL_LENGTH]

def check_rate_limit(client_ip, max_requests=10, window_seconds=60):
    """Simple rate limiting by IP address"""
    current_time = time.time()
    
    # Clean old entries
    cutoff_time = current_time - window_seconds
    if client_ip in request_timestamps:
        request_timestamps[client_ip] = [
            timestamp for timestamp in request_timestamps[client_ip] 
            if timestamp > cutoff_time
        ]
    
    # Check current request count
    if client_ip not in request_timestamps:
        request_timestamps[client_ip] = []
    
    if len(request_timestamps[client_ip]) >= max_requests:
        return False
    
    # Add current request
    request_timestamps[client_ip].append(current_time)
    return True

def extract_info(url):
    """Extract video information without downloading"""
    # Validate URL before processing
    if not validate_url(url):
        print(f"Invalid or potentially dangerous URL: {url}")
        return None
    
    try:
        # Use list form to prevent shell injection
        cmd = ['yt-dlp', '--dump-json', '--no-warnings', '--no-playlist', url]
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            return json.loads(result.stdout)
        else:
            print(f"yt-dlp error: {result.stderr}")
            return None
    except Exception as e:
        print(f"Exception in extract_info: {e}")
        return None

def get_audio_formats(info):
    """Extract audio-only formats from video info"""
    audio_formats = []
    if 'formats' in info:
        for fmt in info['formats']:
            # Audio-only formats (no video codec)
            if fmt.get('vcodec') == 'none' and fmt.get('acodec') != 'none':
                quality = fmt.get('abr', 'unknown')  # Audio bitrate
                ext = fmt.get('ext', 'unknown')
                filesize = fmt.get('filesize')
                filesize_str = f" ({filesize // 1024 // 1024}MB)" if filesize else ""
                
                format_desc = f"{quality}kbps {ext.upper()}{filesize_str}"
                audio_formats.append({
                    'format_id': fmt['format_id'],
                    'description': format_desc,
                    'quality': quality if str(quality).isdigit() else 0
                })
    
    # Sort by quality (descending)
    audio_formats.sort(key=lambda x: int(x['quality']) if str(x['quality']).isdigit() else 0, reverse=True)
    return audio_formats

def download_video(url, format_id, download_id, download_type='video'):
    """Download video or audio in a separate thread"""
    # Validate all inputs before processing
    if not validate_url(url):
        download_progress[download_id] = {'status': 'error', 'progress': 0, 'error': 'Invalid or dangerous URL'}
        return
    
    if not validate_format_id(format_id):
        download_progress[download_id] = {'status': 'error', 'progress': 0, 'error': 'Invalid format ID'}
        return
    
    if not validate_download_type(download_type):
        download_progress[download_id] = {'status': 'error', 'progress': 0, 'error': 'Invalid download type'}
        return
    
    try:
        # Create temporary downloads directory
        temp_dir = tempfile.mkdtemp()
        
        # Build yt-dlp command
        cmd = ['yt-dlp']
        
        # Set output template
        output_template = os.path.join(temp_dir, '%(title)s.%(ext)s')
        
        if download_type == 'mp3':
            # Audio-only download with MP3 conversion (best quality)
            cmd.extend([
                '--extract-audio',
                '--audio-format', 'mp3',
                '--audio-quality', '0',  # Best quality (320kbps equivalent)
                '--output', output_template
            ])
        elif download_type == 'flac':
            # Lossless FLAC audio
            cmd.extend([
                '--extract-audio',
                '--audio-format', 'flac',
                '--output', output_template
            ])
        elif download_type == 'm4a':
            # High quality M4A audio
            cmd.extend([
                '--extract-audio',
                '--audio-format', 'm4a',
                '--audio-quality', '0',  # Best quality
                '--output', output_template
            ])
        elif download_type == 'audio' and format_id and format_id != 'best':
            # Specific audio format
            cmd.extend([
                '-f', format_id,
                '--output', output_template
            ])
        else:
            # Video download
            if format_id and format_id != 'best':
                cmd.extend(['-f', format_id])
            cmd.extend([
                '--output', output_template
            ])
        
        cmd.extend([
            '--no-playlist',
            url
        ])
        
        download_progress[download_id] = {'status': 'downloading', 'progress': 0}
        
        # Run download
        process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, 
                                 universal_newlines=True, bufsize=1)
        
        for line in process.stdout:
            # Parse progress from youtube-dl output
            if '[download]' in line and '%' in line:
                try:
                    progress_match = re.search(r'(\d+\.?\d*)%', line)
                    if progress_match:
                        progress = float(progress_match.group(1))
                        download_progress[download_id]['progress'] = progress
                except:
                    pass
            # Handle ffmpeg conversion progress for audio formats
            elif '[ffmpeg]' in line and download_type in ['mp3', 'flac', 'm4a']:
                download_progress[download_id]['progress'] = 95  # Almost done
        
        process.wait()
        
        if process.returncode == 0:
            # Find the downloaded file
            downloaded_files = []
            for root, dirs, files in os.walk(temp_dir):
                for file in files:
                    downloaded_files.append(os.path.join(root, file))
            
            if downloaded_files:
                file_path = downloaded_files[0]  # Take the first (should be only) file
                download_progress[download_id] = {
                    'status': 'completed', 
                    'progress': 100,
                    'file_path': file_path,
                    'filename': os.path.basename(file_path)
                }
            else:
                download_progress[download_id] = {'status': 'error', 'progress': 0, 'error': 'No file found after download'}
        else:
            download_progress[download_id] = {'status': 'error', 'progress': 0}
            
    except Exception as e:
        download_progress[download_id] = {'status': 'error', 'progress': 0, 'error': str(e)}

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/extract_info', methods=['POST'])
def extract_video_info():
    """Extract video information for format selection"""
    # Rate limiting
    client_ip = request.environ.get('HTTP_X_FORWARDED_FOR', request.environ.get('REMOTE_ADDR', 'unknown'))
    if not check_rate_limit(client_ip, max_requests=20, window_seconds=60):
        return jsonify({'error': 'Rate limit exceeded. Please try again later.'}), 429
    
    # Validate Content-Type
    if not request.is_json:
        return jsonify({'error': 'Content-Type must be application/json'}), 400
    
    data = request.get_json()
    if not data:
        return jsonify({'error': 'Invalid request data'}), 400
    
    url = sanitize_user_input(data.get('url', '')).strip()
    
    if not url:
        return jsonify({'error': 'URL is required'}), 400
    
    # Validate URL with comprehensive security checks
    if not validate_url(url):
        return jsonify({'error': 'Invalid or potentially dangerous URL. Only supported video platforms are allowed.'}), 400
    
    info = extract_info(url)
    if not info:
        return jsonify({'error': 'Could not extract video information'}), 400
    
    # Extract available video formats
    video_formats = []
    if 'formats' in info:
        for fmt in info['formats']:
            # Video formats (has video codec and may have audio)
            if fmt.get('vcodec') != 'none':
                quality = fmt.get('height', 'unknown')
                ext = fmt.get('ext', 'unknown')
                filesize = fmt.get('filesize')
                filesize_str = f" ({filesize // 1024 // 1024}MB)" if filesize else ""
                
                format_desc = f"{quality}p {ext.upper()}{filesize_str}"
                video_formats.append({
                    'format_id': fmt['format_id'],
                    'description': format_desc,
                    'quality': quality
                })
    
    # Sort video formats by quality (descending)
    video_formats.sort(key=lambda x: int(x['quality']) if str(x['quality']).isdigit() else 0, reverse=True)
    
    # Extract audio formats
    audio_formats = get_audio_formats(info)
    
    return jsonify({
        'title': info.get('title', 'Unknown'),
        'uploader': info.get('uploader', 'Unknown'),
        'duration': info.get('duration'),
        'video_formats': video_formats,
        'audio_formats': audio_formats
    })

@app.route('/download', methods=['POST'])
def download():
    """Start video or audio download"""
    # Rate limiting (stricter for downloads)
    client_ip = request.environ.get('HTTP_X_FORWARDED_FOR', request.environ.get('REMOTE_ADDR', 'unknown'))
    if not check_rate_limit(client_ip, max_requests=5, window_seconds=60):
        return jsonify({'error': 'Rate limit exceeded. Please try again later.'}), 429
    
    # Validate Content-Type
    if not request.is_json:
        return jsonify({'error': 'Content-Type must be application/json'}), 400
    
    data = request.get_json()
    if not data:
        return jsonify({'error': 'Invalid request data'}), 400
    
    # Sanitize and validate all inputs
    url = sanitize_user_input(data.get('url', '')).strip()
    format_id = sanitize_user_input(data.get('format', 'best')).strip()
    download_type = sanitize_user_input(data.get('type', 'video')).strip()
    
    if not url:
        return jsonify({'error': 'URL is required'}), 400
    
    # Validate URL
    if not validate_url(url):
        return jsonify({'error': 'Invalid or potentially dangerous URL. Only supported video platforms are allowed.'}), 400
    
    # Validate format ID
    if not validate_format_id(format_id):
        return jsonify({'error': 'Invalid format ID'}), 400
    
    # Validate download type
    if not validate_download_type(download_type):
        return jsonify({'error': 'Invalid download type. Allowed types: video, audio, mp3, flac, m4a'}), 400
    
    # Generate unique download ID
    download_id = str(int(time.time() * 1000))
    
    # Start download in background thread
    thread = threading.Thread(target=download_video, args=(url, format_id, download_id, download_type))
    thread.daemon = True
    thread.start()
    
    return jsonify({'download_id': download_id})

@app.route('/progress/<download_id>')
def get_progress(download_id):
    """Get download progress"""
    progress = download_progress.get(download_id, {'status': 'not_found', 'progress': 0})
    return jsonify(progress)

@app.route('/download_file/<download_id>')
def download_file(download_id):
    """Serve the downloaded file to the user's browser"""
    progress = download_progress.get(download_id)
    
    if not progress or progress.get('status') != 'completed':
        return jsonify({'error': 'Download not found or not completed'}), 404
    
    file_path = progress.get('file_path')
    if not file_path or not os.path.exists(file_path):
        return jsonify({'error': 'File not found'}), 404
    
    filename = progress.get('filename', 'download')
    
    # Serve the file for download
    response = send_file(
        file_path,
        as_attachment=True,
        download_name=filename,
        mimetype='application/octet-stream'
    )
    
    # Schedule file cleanup after sending
    def cleanup_file():
        try:
            if os.path.exists(file_path):
                os.remove(file_path)
            # Try to remove the temp directory if it's empty
            temp_dir = os.path.dirname(file_path)
            try:
                os.rmdir(temp_dir)
            except OSError:
                pass  # Directory not empty or other error
            # Remove from progress tracking
            if download_id in download_progress:
                del download_progress[download_id]
        except Exception as e:
            print(f"Error cleaning up file: {e}")
    
    # Schedule cleanup after a delay to allow file to be sent
    threading.Timer(30, cleanup_file).start()
    
    return response

@app.route('/downloads')
def list_downloads():
    """List completed downloads"""
    completed_downloads = []
    for download_id, progress in download_progress.items():
        if progress.get('status') == 'completed':
            completed_downloads.append({
                'download_id': download_id,
                'filename': progress.get('filename', 'Unknown'),
                'ready': True
            })
    
    return jsonify(completed_downloads)

if __name__ == '__main__':
    app.run(debug=False, host='0.0.0.0', port=8000) 