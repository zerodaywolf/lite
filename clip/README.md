# Clip - Private Text & File Sharing

A simple, lightweight Go web application that provides secure, temporary sharing of text and files with memorable URLs. Perfect for quick, private sharing between devices or people.

## 🚀 Features

- **📋 Text Sharing**: Paste text and get a memorable URL (like `happy-cat` or `blue-moon`)
- **📁 File Sharing**: Upload files (up to 1GB) and get shareable links
- **🔗 Memorable URLs**: Easy-to-remember two-word combinations instead of long random strings
- **⏰ Auto-Expiry**: All content automatically deletes after 12 hours for privacy
- **📱 Responsive**: Works perfectly on desktop and mobile devices
- **⚡ Lightweight**: Uses only Go standard library (no external dependencies)
- **🔒 Private**: Content only accessible via specific URLs, no browsing/listing

## 🎯 URL Examples

Instead of ugly URLs like `/c/a1b2c3d4e5f6`, you get:
- `/c/happy-cat` - for text shares
- `/f/blue-moon` - for file shares

## ⏰ Auto-Cleanup

- **12-hour expiry**: All content is automatically deleted after 12 hours
- **Hourly cleanup**: Server runs cleanup every hour to remove expired content
- **Privacy-focused**: No permanent storage of user content

## 🛠️ How to Run

1. Make sure you have Go installed on your system
2. Navigate to this directory
3. Run the application:

```bash
go run main.go
```

4. Open your browser and go to: `http://localhost:8000`




## 📖 Usage Flow

### Sharing Text
1. Go to your Clip instance (e.g., `localhost:8000`)
2. Paste your text in the left panel
3. Click "Generate Link"
4. Share the memorable URL (e.g., `localhost:8000/c/swift-river`)

### Sharing Files
1. Drag & drop or select files in the right panel
2. Click "Generate Links"
3. Share the file URLs (e.g., `localhost:8000/f/calm-star`)

### Accessing Shared Content
- **Text**: Opens in a clean, readable format with copy functionality
- **Files**: Downloads immediately with original filename
- **Expiry**: Shows remaining time before auto-deletion

## 🔧 Customization

You can easily customize:
- **Port**: Change `:8000` in `main.go`
- **Expiry time**: Change `12 * time.Hour` in the cleanup functions
- **File size limit**: Change `1 << 30` (1GB) in `main.go`
- **Word lists**: Modify `adjectives` and `nouns` arrays for different URL styles
- **Cleanup frequency**: Change `1 * time.Hour` in `startCleanupRoutine()`

## 🔒 Security Features

- **No content listing**: Can't browse all content without specific URLs
- **Path validation**: Prevents directory traversal attacks  
- **File size limits**: Prevents abuse with oversized uploads
- **Auto-cleanup**: Ensures no permanent data retention
- **Safe filenames**: Handles malicious filename attempts

## 💡 Use Cases

- **Cross-device sharing**: Quickly share text/files between your devices
- **Temporary file sharing**: Send files without email attachment limits
- **Code snippets**: Share code with memorable URLs
- **Meeting notes**: Quick sharing during meetings
- **Screenshots**: Upload and share images instantly

## 🏗️ Architecture

- **Backend**: Pure Go with standard library only
- **Frontend**: Vanilla HTML/CSS/JavaScript (no frameworks)
- **Storage**: Local filesystem with automatic cleanup
- **Memory**: In-memory indexing with disk persistence
- **URLs**: Cryptographically secure random word combinations 