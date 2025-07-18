// DOM elements
const videoUrlInput = document.getElementById('videoUrl');
const analyzeBtn = document.getElementById('analyzeBtn');
const loadingInfo = document.getElementById('loadingInfo');
const videoInfo = document.getElementById('videoInfo');
const videoTitle = document.getElementById('videoTitle');
const videoUploader = document.getElementById('videoUploader');
const videoDuration = document.getElementById('videoDuration');
const videoFormatSelect = document.getElementById('videoFormatSelect');
const audioFormatSelect = document.getElementById('audioFormatSelect');
const downloadBtn = document.getElementById('downloadBtn');
const downloadProgress = document.getElementById('downloadProgress');
const progressFill = document.getElementById('progressFill');
const progressText = document.getElementById('progressText');
const downloadStatus = document.getElementById('downloadStatus');
const downloadsList = document.getElementById('downloadsList');
const refreshBtn = document.getElementById('refreshBtn');

// Download type elements
const downloadTypeRadios = document.querySelectorAll('input[name="downloadType"]');
const videoFormatSection = document.getElementById('videoFormatSection');
const audioFormatSection = document.getElementById('audioFormatSection');
const mp3FormatSection = document.getElementById('mp3FormatSection');
const flacFormatSection = document.getElementById('flacFormatSection');
const m4aFormatSection = document.getElementById('m4aFormatSection');

// State
let currentDownloadId = null;
let progressInterval = null;

// Event listeners
analyzeBtn.addEventListener('click', analyzeVideo);
downloadBtn.addEventListener('click', startDownload);
refreshBtn.addEventListener('click', loadDownloads);
videoUrlInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        analyzeVideo();
    }
});

// Download type selection
downloadTypeRadios.forEach(radio => {
    radio.addEventListener('change', handleDownloadTypeChange);
});

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadDownloads();
    handleDownloadTypeChange(); // Set initial state
});

// Handle download type change
function handleDownloadTypeChange() {
    const selectedType = document.querySelector('input[name="downloadType"]:checked').value;
    
    // Hide all format sections
    videoFormatSection.classList.add('hidden');
    audioFormatSection.classList.add('hidden');
    mp3FormatSection.classList.add('hidden');
    flacFormatSection.classList.add('hidden');
    m4aFormatSection.classList.add('hidden');
    
    // Show relevant section
    if (selectedType === 'video') {
        videoFormatSection.classList.remove('hidden');
    } else if (selectedType === 'audio') {
        audioFormatSection.classList.remove('hidden');
    } else if (selectedType === 'mp3') {
        mp3FormatSection.classList.remove('hidden');
    } else if (selectedType === 'flac') {
        flacFormatSection.classList.remove('hidden');
    } else if (selectedType === 'm4a') {
        m4aFormatSection.classList.remove('hidden');
    }
}

// Utility functions
function showNotification(message, type = 'info') {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.classList.remove('hidden');
    
    setTimeout(() => {
        notification.classList.add('hidden');
    }, 5000);
}

function formatDuration(seconds) {
    if (!seconds) return 'Unknown';
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    } else {
        return `${minutes}:${secs.toString().padStart(2, '0')}`;
    }
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function isValidUrl(string) {
    try {
        new URL(string);
        return true;
    } catch (_) {
        return false;
    }
}

// Main functions
async function analyzeVideo() {
    const url = videoUrlInput.value.trim();
    
    if (!url) {
        showNotification('Please enter a URL', 'error');
        return;
    }
    
    if (!isValidUrl(url)) {
        showNotification('Please enter a valid URL', 'error');
        return;
    }
    
    // Reset UI
    videoInfo.classList.add('hidden');
    downloadProgress.classList.add('hidden');
    loadingInfo.classList.remove('hidden');
    analyzeBtn.disabled = true;
    
    try {
        const response = await fetch('/extract_info', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: url })
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Failed to analyze video');
        }
        
        // Populate video info
        videoTitle.textContent = data.title;
        videoUploader.textContent = data.uploader;
        videoDuration.textContent = formatDuration(data.duration);
        
        // Populate video format options
        videoFormatSelect.innerHTML = '<option value="best">Best Quality</option>';
        if (data.video_formats && data.video_formats.length > 0) {
            data.video_formats.forEach(format => {
                const option = document.createElement('option');
                option.value = format.format_id;
                option.textContent = format.description;
                videoFormatSelect.appendChild(option);
            });
        }

        // Populate audio format options
        audioFormatSelect.innerHTML = '<option value="best">Best Quality</option>';
        if (data.audio_formats && data.audio_formats.length > 0) {
            data.audio_formats.forEach(format => {
                const option = document.createElement('option');
                option.value = format.format_id;
                option.textContent = format.description;
                audioFormatSelect.appendChild(option);
            });
        }
        
        // Show video info
        loadingInfo.classList.add('hidden');
        videoInfo.classList.remove('hidden');
        
        showNotification('Video analyzed successfully!', 'success');
        
    } catch (error) {
        loadingInfo.classList.add('hidden');
        showNotification(error.message, 'error');
        console.error('Error analyzing video:', error);
    } finally {
        analyzeBtn.disabled = false;
    }
}

async function startDownload() {
    const url = videoUrlInput.value.trim();
    const downloadType = document.querySelector('input[name="downloadType"]:checked').value;
    
    let format = 'best';
    if (downloadType === 'video') {
        format = videoFormatSelect.value;
    } else if (downloadType === 'audio') {
        format = audioFormatSelect.value;
    }
    // For MP3, we don't need a specific format as it will be converted
    
    if (!url) {
        showNotification('Please enter a URL', 'error');
        return;
    }
    
    downloadBtn.disabled = true;
    downloadProgress.classList.remove('hidden');
    progressFill.style.width = '0%';
    progressText.textContent = '0%';
    downloadStatus.textContent = '';
    downloadStatus.className = '';
    
    try {
        const response = await fetch('/download', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                url: url,
                format: format,
                type: downloadType
            })
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Failed to start download');
        }
        
        currentDownloadId = data.download_id;
        showNotification('Download started!', 'info');
        
        // Start progress tracking
        trackProgress();
        
    } catch (error) {
        downloadBtn.disabled = false;
        downloadProgress.classList.add('hidden');
        showNotification(error.message, 'error');
        console.error('Error starting download:', error);
    }
}

function trackProgress() {
    if (!currentDownloadId) return;
    
    progressInterval = setInterval(async () => {
        try {
            const response = await fetch(`/progress/${currentDownloadId}`);
            const data = await response.json();
            
            if (data.status === 'downloading') {
                const progress = Math.round(data.progress);
                progressFill.style.width = `${progress}%`;
                progressText.textContent = `${progress}%`;
            } else if (data.status === 'completed') {
                progressFill.style.width = '100%';
                progressText.textContent = '100%';
                downloadStatus.innerHTML = '✅ Download completed! <a href="#" id="downloadLink">Click to download</a>';
                downloadStatus.className = 'success';
                
                // Auto-trigger browser download
                const downloadUrl = `/download_file/${currentDownloadId}`;
                const link = document.createElement('a');
                link.href = downloadUrl;
                link.download = data.filename || 'download';
                document.body.appendChild(link);
                link.click();
                document.body.removeChild(link);
                
                // Also add click handler to the manual link
                const downloadLink = document.getElementById('downloadLink');
                if (downloadLink) {
                    downloadLink.onclick = (e) => {
                        e.preventDefault();
                        window.location.href = downloadUrl;
                    };
                }
                
                clearInterval(progressInterval);
                downloadBtn.disabled = false;
                currentDownloadId = null;
                
                showNotification('Download started in your browser!', 'success');
                loadDownloads(); // Refresh downloads list
                
            } else if (data.status === 'error') {
                downloadStatus.textContent = `❌ Download failed: ${data.error || 'Unknown error'}`;
                downloadStatus.className = 'error';
                
                clearInterval(progressInterval);
                downloadBtn.disabled = false;
                currentDownloadId = null;
                
                showNotification('Download failed', 'error');
            }
            
        } catch (error) {
            console.error('Error tracking progress:', error);
            clearInterval(progressInterval);
            downloadBtn.disabled = false;
            currentDownloadId = null;
        }
    }, 1000);
}

async function loadDownloads() {
    try {
        const response = await fetch('/downloads');
        const downloads = await response.json();
        
        if (downloads.length === 0) {
            downloadsList.innerHTML = '<p class="no-downloads">No completed downloads</p>';
        } else {
            downloadsList.innerHTML = '';
            downloads.forEach(download => {
                const item = document.createElement('div');
                item.className = 'download-item';
                item.innerHTML = `
                    <div class="file-name">${download.filename}</div>
                    <div class="download-actions">
                        <a href="/download_file/${download.download_id}" class="download-link">
                            <i class="fas fa-download"></i> Download
                        </a>
                    </div>
                `;
                downloadsList.appendChild(item);
            });
        }
    } catch (error) {
        console.error('Error loading downloads:', error);
        showNotification('Failed to load downloads', 'error');
    }
}

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (progressInterval) {
        clearInterval(progressInterval);
    }
}); 