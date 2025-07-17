package main

import (
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ClipboardEntry struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

type FileEntry struct {
	ID        string
	Filename  string
	CreatedAt time.Time
}

type PageData struct {
	ClipboardEntries []ClipboardEntry
	FileEntries      []FileEntry
	Message          string
}

var clipboardEntries = make(map[string]ClipboardEntry)
var fileEntries = make(map[string]FileEntry)

// Word lists for generating memorable URLs
var adjectives = []string{
	"happy", "bright", "calm", "swift", "clever", "gentle", "bold", "quiet",
	"warm", "cool", "fresh", "sweet", "sharp", "soft", "strong", "light",
	"dark", "clear", "deep", "wide", "tiny", "huge", "quick", "slow",
	"wild", "tame", "free", "safe", "pure", "wise", "kind", "brave",
	"neat", "cute", "fancy", "plain", "rich", "poor", "young", "old",
	"new", "blue", "red", "green", "gold", "pink", "gray", "white",
}

var nouns = []string{
	"cat", "dog", "bird", "fish", "lion", "bear", "wolf", "fox",
	"tree", "flower", "star", "moon", "sun", "cloud", "rain", "snow",
	"ocean", "river", "mountain", "valley", "forest", "desert", "island", "beach",
	"book", "pen", "key", "door", "window", "chair", "table", "lamp",
	"music", "song", "dance", "game", "story", "dream", "wish", "hope",
	"fire", "ice", "wind", "earth", "stone", "sand", "wave", "spark",
	"magic", "wonder", "peace", "joy", "love", "trust", "truth", "grace",
}

func generateReadableID() string {
	// Generate random indices
	adjIndex, err1 := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIndex, err2 := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	
	if err1 != nil || err2 != nil {
		// Fallback to timestamp if random fails
		return fmt.Sprintf("paste-%d", time.Now().Unix())
	}
	
	adj := adjectives[adjIndex.Int64()]
	noun := nouns[nounIndex.Int64()]
	
	return fmt.Sprintf("%s-%s", adj, noun)
}

func startCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Check every hour
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				cleanupExpiredEntries()
			}
		}
	}()
}

func cleanupExpiredEntries() {
	now := time.Now()
	cutoff := now.Add(-12 * time.Hour) // 12 hours ago
	
	// Clean clipboard entries
	for id, entry := range clipboardEntries {
		if entry.CreatedAt.Before(cutoff) {
			// Remove from memory
			delete(clipboardEntries, id)
			
			// Remove file
			filename := fmt.Sprintf("clipboard_%s.txt", id)
			filepath := filepath.Join("uploads", filename)
			if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to remove expired clipboard file %s: %v", filepath, err)
			}
			
			log.Printf("Cleaned up expired clipboard entry: %s", id)
		}
	}
	
	// Clean file entries
	for id, entry := range fileEntries {
		if entry.CreatedAt.Before(cutoff) {
			// Remove from memory
			delete(fileEntries, id)
			
			// Find and remove the actual file
			files, err := os.ReadDir("uploads")
			if err == nil {
				for _, file := range files {
					if strings.HasPrefix(file.Name(), id) {
						filepath := filepath.Join("uploads", file.Name())
						if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
							log.Printf("Failed to remove expired file %s: %v", filepath, err)
						}
						break
					}
				}
			}
			
			log.Printf("Cleaned up expired file entry: %s (%s)", id, entry.Filename)
		}
	}
	
	log.Printf("Cleanup completed at %s", now.Format("2006-01-02 15:04:05"))
}

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	// Start the cleanup routine
	startCleanupRoutine()

	// Route handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/clipboard", clipboardHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/c/", clipboardViewHandler)
	http.HandleFunc("/f/", fileViewHandler)

	// Start server
	fmt.Println("Server starting on http://localhost:8000")
	fmt.Println("Auto-cleanup: Entries expire after 12 hours")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Convert maps to slices for template rendering, sorted by creation time
	var clipEntries []ClipboardEntry
	var fileEnts []FileEntry
	
	for _, entry := range clipboardEntries {
		clipEntries = append(clipEntries, entry)
	}
	
	for _, entry := range fileEntries {
		fileEnts = append(fileEnts, entry)
	}

	data := PageData{
		ClipboardEntries: clipEntries,
		FileEntries:      fileEnts,
	}

	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Clip - Private Text & File Sharing</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f5f5f5;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }
        
        .header {
            background-color: #2c3e50;
            color: white;
            padding: 1rem;
            text-align: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .header p {
            margin-top: 0.5rem;
            opacity: 0.9;
            font-size: 0.9rem;
        }
        
        .container {
            display: flex;
            flex: 1;
            gap: 2rem;
            padding: 2rem;
            max-width: 1200px;
            margin: 0 auto;
            width: 100%;
        }
        
        .panel {
            flex: 1;
            background: white;
            border-radius: 8px;
            padding: 2rem;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
        }
        
        .panel h2 {
            color: #2c3e50;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #3498db;
        }
        
        .clipboard-panel textarea {
            flex: 1;
            border: 2px solid #e0e0e0;
            border-radius: 4px;
            padding: 1rem;
            font-family: 'Courier New', monospace;
            font-size: 14px;
            resize: none;
            outline: none;
            margin-bottom: 1rem;
            min-height: 200px;
        }
        
        .clipboard-panel textarea:focus {
            border-color: #3498db;
        }
        
        .upload-panel {
            display: flex;
            flex-direction: column;
        }
        
        .upload-area {
            border: 2px dashed #3498db;
            border-radius: 8px;
            padding: 3rem;
            text-align: center;
            background-color: #f8f9fa;
            margin-bottom: 2rem;
            transition: all 0.3s ease;
        }
        
        .upload-area:hover {
            background-color: #e8f4f8;
            border-color: #2980b9;
        }
        
        .upload-area input[type="file"] {
            display: none;
        }
        
        .upload-btn {
            background-color: #3498db;
            color: white;
            border: none;
            padding: 1rem 2rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            transition: background-color 0.3s ease;
        }
        
        .upload-btn:hover {
            background-color: #2980b9;
        }
        
        .btn {
            background-color: #27ae60;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            transition: background-color 0.3s ease;
        }
        
        .btn:hover {
            background-color: #219a52;
        }
        
        .links-section {
            margin-top: 2rem;
            padding: 2rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        .links-section h2 {
            color: #2c3e50;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #3498db;
        }
        
        .links-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
        }
        
        .link-group h3 {
            color: #34495e;
            margin-bottom: 1rem;
            font-size: 1.1rem;
        }
        
        .link-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem;
            border: 1px solid #eee;
            border-radius: 4px;
            margin-bottom: 0.5rem;
            background-color: #fafafa;
        }
        
        .link-item:hover {
            background-color: #f0f0f0;
        }
        
        .link-url {
            color: #3498db;
            text-decoration: none;
            font-family: monospace;
            font-size: 0.9rem;
            font-weight: 500;
        }
        
        .link-url:hover {
            text-decoration: underline;
        }
        
        .link-time {
            color: #7f8c8d;
            font-size: 0.8rem;
        }
        
        .empty-state {
            text-align: center;
            color: #7f8c8d;
            font-style: italic;
            padding: 2rem;
        }
        
        .copy-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            cursor: pointer;
            font-size: 0.8rem;
        }
        
        .copy-btn:hover {
            background: #2980b9;
        }
        
        .delete-btn {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            cursor: pointer;
            font-size: 0.8rem;
            margin-left: 0.25rem;
        }
        
        .delete-btn:hover {
            background: #c0392b;
        }
        
        .action-buttons {
            display: flex;
            gap: 0.25rem;
            align-items: center;
        }
        
        .expiry-notice {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 0.75rem;
            border-radius: 4px;
            margin-bottom: 1rem;
            font-size: 0.9rem;
        }
        
        @media (max-width: 768px) {
            .container {
                flex-direction: column;
                padding: 1rem;
            }
            
            .panel {
                padding: 1rem;
            }
            
            .links-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Clip - Private Sharing</h1>
        <p>Share text & files with memorable URLs • Auto-expires in 12 hours</p>
    </div>
    
    <div class="expiry-notice">
        ⏰ <strong>Auto-Cleanup:</strong> All content is automatically deleted after 12 hours for privacy and security.
    </div>
    
    <div class="container">
        <!-- Left Panel: New Clipboard Entry -->
        <div class="panel clipboard-panel">
            <h2>📋 Share Text</h2>
            <form method="POST" action="/clipboard">
                <textarea name="content" placeholder="Paste your text here..." required></textarea>
                <button type="submit" class="btn">Generate Link</button>
            </form>
        </div>
        
        <!-- Right Panel: File Upload -->
        <div class="panel upload-panel">
            <h2>📁 Share Files</h2>
            <form method="POST" action="/upload" enctype="multipart/form-data">
                <div class="upload-area" onclick="document.getElementById('fileInput').click()">
                    <input type="file" name="file" id="fileInput" multiple style="display: none;">
                    <div style="font-size: 3rem; margin-bottom: 1rem;">📎</div>
                    <div style="font-size: 1.2rem; margin-bottom: 1rem;">Drop files here or click to browse</div>
                    <div class="upload-btn" style="pointer-events: none;">
                        Choose Files
                    </div>
                </div>
                <button type="submit" class="btn">Generate Links</button>
            </form>
        </div>
    </div>
    
    <div class="links-section">
        <h2>🔗 Your Recent Links</h2>
        <div class="links-grid">
            <div class="link-group">
                <h3>📋 Text Shares</h3>
                {{if .ClipboardEntries}}
                    {{range .ClipboardEntries}}
                    <div class="link-item" id="clipboard-{{.ID}}">
                        <a href="/c/{{.ID}}" class="link-url" target="_blank">{{.ID}}</a>
                        <div style="display: flex; gap: 0.5rem; align-items: center;">
                            <span class="link-time">{{.CreatedAt.Format "Jan 2, 15:04"}}</span>
                            <div class="action-buttons">
                                <button class="copy-btn" onclick="copyToClipboard('{{.ID}}', 'c')">Copy</button>
                                <button class="delete-btn" onclick="deleteShare('{{.ID}}', 'c')" title="Delete share">🗑️</button>
                            </div>
                        </div>
                    </div>
                    {{end}}
                {{else}}
                    <div class="empty-state">No text shares yet</div>
                {{end}}
            </div>
            
            <div class="link-group">
                <h3>📁 File Shares</h3>
                {{if .FileEntries}}
                    {{range .FileEntries}}
                    <div class="link-item" id="file-{{.ID}}">
                        <a href="/f/{{.ID}}" class="link-url" target="_blank">{{.ID}} <small>({{.Filename}})</small></a>
                        <div style="display: flex; gap: 0.5rem; align-items: center;">
                            <span class="link-time">{{.CreatedAt.Format "Jan 2, 15:04"}}</span>
                            <div class="action-buttons">
                                <button class="copy-btn" onclick="copyToClipboard('{{.ID}}', 'f')">Copy</button>
                                <button class="delete-btn" onclick="deleteShare('{{.ID}}', 'f')" title="Delete share">🗑️</button>
                            </div>
                        </div>
                    </div>
                    {{end}}
                {{else}}
                    <div class="empty-state">No file shares yet</div>
                {{end}}
            </div>
        </div>
    </div>
    
    {{if .Message}}
    <div class="message">{{.Message}}</div>
    {{end}}
    
    <script>
        // Handle drag and drop
        const uploadArea = document.querySelector('.upload-area');
        const fileInput = document.getElementById('fileInput');
        
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.style.backgroundColor = '#e8f4f8';
        });
        
        uploadArea.addEventListener('dragleave', (e) => {
            e.preventDefault();
            uploadArea.style.backgroundColor = '#f8f9fa';
        });
        
        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.style.backgroundColor = '#f8f9fa';
            fileInput.files = e.dataTransfer.files;
            updateFileDisplay();
        });
        
        fileInput.addEventListener('change', updateFileDisplay);
        
        function updateFileDisplay() {
            const files = fileInput.files;
            const uploadBtn = document.querySelector('.upload-btn');
            if (files.length > 0) {
                uploadBtn.textContent = files.length === 1 ? '1 file selected' : files.length + ' files selected';
            } else {
                uploadBtn.textContent = 'Choose Files';
            }
        }
        
        function copyToClipboard(id, type) {
            const url = window.location.origin + '/' + type + '/' + id;
            navigator.clipboard.writeText(url).then(() => {
                // Brief visual feedback
                event.target.textContent = 'Copied!';
                setTimeout(() => {
                    event.target.textContent = 'Copy';
                }, 1000);
            });
        }
        
        function deleteShare(id, type) {
            if (!confirm('Are you sure you want to delete this share? This action cannot be undone.')) {
                return;
            }
            
            const deleteUrl = '/delete/' + type + '/' + id;
            
            fetch(deleteUrl, {
                method: 'POST',
                headers: {
                    'Accept': 'application/json'
                }
            })
            .then(response => {
                if (response.ok) {
                    // Remove the item from the DOM
                    const elementId = (type === 'c' ? 'clipboard-' : 'file-') + id;
                    const element = document.getElementById(elementId);
                    if (element) {
                        element.style.opacity = '0.5';
                        element.style.transition = 'opacity 0.3s ease';
                        setTimeout(() => {
                            element.remove();
                            
                            // Check if sections are empty and show empty state
                            updateEmptyStates();
                        }, 300);
                    }
                } else {
                    alert('Failed to delete share. Please try again.');
                }
            })
            .catch(error => {
                console.error('Error deleting share:', error);
                alert('Failed to delete share. Please try again.');
            });
        }
        
                 function updateEmptyStates() {
             // Check clipboard section
             const clipboardSection = document.querySelector('.link-group:first-child');
             const clipboardItems = clipboardSection.querySelectorAll('.link-item');
             if (clipboardItems.length === 0) {
                 clipboardSection.innerHTML = '<h3>&#128203; Text Shares</h3><div class="empty-state">No text shares yet</div>';
             }
             
             // Check file section
             const fileSection = document.querySelector('.link-group:last-child');
             const fileItems = fileSection.querySelectorAll('.link-item');
             if (fileItems.length === 0) {
                 fileSection.innerHTML = '<h3>&#128193; File Shares</h3><div class="empty-state">No file shares yet</div>';
             }
         }
    </script>
</body>
</html>
`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func clipboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if strings.TrimSpace(content) == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	// Generate unique readable ID
	id := generateReadableID()
	
	// Ensure uniqueness (though collisions are extremely rare)
	for _, exists := clipboardEntries[id]; exists; {
		id = generateReadableID()
	}
	
	entry := ClipboardEntry{
		ID:        id,
		Content:   content,
		CreatedAt: time.Now(),
	}
	
	clipboardEntries[id] = entry

	// Save to file with ID
	filename := fmt.Sprintf("clipboard_%s.txt", id)
	filepath := filepath.Join("uploads", filename)
	
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		log.Printf("Failed to save clipboard content: %v", err)
	}

	log.Printf("Created clipboard entry: %s", id)

	// Redirect back to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(1 << 30) // 1GB max memory
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range files {
		// Generate unique readable ID for each file
		id := generateReadableID()
		
		// Ensure uniqueness
		for _, exists := fileEntries[id]; exists; {
			id = generateReadableID()
		}
		
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open file %s: %v", fileHeader.Filename, err)
			continue
		}
		defer file.Close()

		// Use original filename but store with unique ID
		originalFilename := filepath.Base(fileHeader.Filename)
		if originalFilename == "" || originalFilename == "." {
			originalFilename = "upload_" + fmt.Sprintf("%d", time.Now().Unix())
		}

		// Store file with unique ID as filename
		ext := filepath.Ext(originalFilename)
		destPath := filepath.Join("uploads", id+ext)

		// Create the destination file
		destFile, err := os.Create(destPath)
		if err != nil {
			log.Printf("Failed to create file %s: %v", destPath, err)
			continue
		}
		defer destFile.Close()

		// Copy the uploaded file to the destination
		_, err = io.Copy(destFile, file)
		if err != nil {
			log.Printf("Failed to copy file %s: %v", destPath, err)
			continue
		}

		// Store file entry with original filename for display
		entry := FileEntry{
			ID:        id,
			Filename:  originalFilename,
			CreatedAt: time.Now(),
		}
		
		fileEntries[id] = entry
		log.Printf("Created file entry: %s (%s)", id, originalFilename)
	}

	// Redirect back to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/delete/")
	parts := strings.SplitN(path, "/", 2)
	
	if len(parts) != 2 {
		http.Error(w, "Invalid delete path", http.StatusBadRequest)
		return
	}
	
	entryType := parts[0] // "c" for clipboard or "f" for file
	id := parts[1]
	
	var deleted bool
	
	switch entryType {
	case "c":
		// Delete clipboard entry
		if _, exists := clipboardEntries[id]; exists {
			delete(clipboardEntries, id)
			
			// Remove file
			filename := fmt.Sprintf("clipboard_%s.txt", id)
			filepath := filepath.Join("uploads", filename)
			if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to remove clipboard file %s: %v", filepath, err)
			}
			
			deleted = true
			log.Printf("Deleted clipboard entry: %s", id)
		}
		
	case "f":
		// Delete file entry
		if _, exists := fileEntries[id]; exists {
			delete(fileEntries, id)
			
			// Find and remove the actual file
			files, err := os.ReadDir("uploads")
			if err == nil {
				for _, file := range files {
					if strings.HasPrefix(file.Name(), id) {
						filepath := filepath.Join("uploads", file.Name())
						if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
							log.Printf("Failed to remove file %s: %v", filepath, err)
						}
						break
					}
				}
			}
			
			deleted = true
			log.Printf("Deleted file entry: %s", id)
		}
		
	default:
		http.Error(w, "Invalid entry type", http.StatusBadRequest)
		return
	}
	
	if !deleted {
		http.Error(w, "Entry not found", http.StatusNotFound)
		return
	}
	
	// Return success for AJAX requests
	if r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true}`))
		return
	}
	
	// Redirect back to home for form submissions
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clipboardViewHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/c/")
	if id == "" {
		http.Error(w, "Invalid clipboard ID", http.StatusBadRequest)
		return
	}

	entry, exists := clipboardEntries[id]
	if !exists {
		http.Error(w, "Clipboard entry not found or expired", http.StatusNotFound)
		return
	}

	// Calculate remaining time
	expiresAt := entry.CreatedAt.Add(12 * time.Hour)
	timeLeft := time.Until(expiresAt)
	
	// Simple HTML page to display clipboard content
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>📋 {{.ID}} - Clip</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: #f5f5f5;
            margin: 0;
            padding: 2rem;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            padding: 2rem;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 2px solid #3498db;
            padding-bottom: 1rem;
            margin-bottom: 2rem;
        }
        .content {
            background: #f8f9fa;
            border: 1px solid #e0e0e0;
            border-radius: 4px;
            padding: 1rem;
            font-family: 'Courier New', monospace;
            white-space: pre-wrap;
            word-wrap: break-word;
            line-height: 1.5;
        }
        .meta {
            margin-top: 1rem;
            color: #7f8c8d;
            font-size: 0.9rem;
        }
        .btn {
            background-color: #3498db;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            margin-top: 1rem;
            margin-right: 0.5rem;
        }
        .expiry-warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 0.75rem;
            border-radius: 4px;
            margin-bottom: 1rem;
            font-size: 0.9rem;
        }
        .url-display {
            background: #e8f4f8;
            border: 1px solid #3498db;
            padding: 0.75rem;
            border-radius: 4px;
            font-family: monospace;
            margin-bottom: 1rem;
            color: #2c3e50;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="url-display">
            🔗 <strong>Share URL:</strong> {{.URL}}
        </div>
        
        {{if .TimeLeft}}
        <div class="expiry-warning">
            ⏰ <strong>Auto-expires in:</strong> {{.TimeLeft}}
        </div>
        {{end}}
        
        <div class="header">
            <h1>📋 {{.ID}}</h1>
            <p>Created: {{.CreatedAt.Format "January 2, 2006 at 15:04 MST"}}</p>
        </div>
        <div class="content">{{.Content}}</div>
        <div class="meta">
            <a href="/" class="btn">← Back to Home</a>
            <button class="btn" onclick="copyToClipboard()">Copy Content</button>
        </div>
    </div>
    
    <script>
        function copyToClipboard() {
            const content = document.querySelector('.content').textContent;
            navigator.clipboard.writeText(content).then(() => {
                event.target.textContent = 'Copied!';
                setTimeout(() => {
                    event.target.textContent = 'Copy Content';
                }, 1000);
            });
        }
    </script>
</body>
</html>
`

	data := struct {
		ClipboardEntry
		URL      string
		TimeLeft string
	}{
		ClipboardEntry: entry,
		URL:            r.Host + r.URL.Path,
		TimeLeft:       formatDuration(timeLeft),
	}

	t, err := template.New("clipboard").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func fileViewHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/f/")
	if id == "" {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	entry, exists := fileEntries[id]
	if !exists {
		http.Error(w, "File not found or expired", http.StatusNotFound)
		return
	}

	// Find the actual file (could have different extension)
	var filepath string
	files, err := os.ReadDir("uploads")
	if err != nil {
		http.Error(w, "Error reading uploads directory", http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), id) {
			filepath = "uploads/" + file.Name()
			break
		}
	}

	if filepath == "" {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	// Set filename for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", entry.Filename))
	
	// Serve the file
	http.ServeFile(w, r, filepath)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Expired"
	}
	
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
} 