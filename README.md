# Quickies - Small Services Collection

A collection of small, lightweight services and utilities.

## Available Services

### 📋 Clip
**Path:** `clip/`

A simple Go web application that provides:
- **Clipboard functionality** (left panel) - paste and save text like pastebin
- **File upload functionality** (right panel) - drag & drop or browse to upload files up to 1GB

**Features:**
- Built with Go standard library only (no dependencies)
- Responsive web interface
- File download links
- Automatic file deduplication
- 1GB upload limit

**To run:**
```bash
cd clip
go run main.go
# Visit http://localhost:8000
```

---

## Adding New Services

When adding new small services to this collection:

1. Create a new folder with a descriptive name
2. Include a README.md with setup instructions
3. Keep services lightweight and focused on a single purpose
4. Update this main README.md to document the new service 