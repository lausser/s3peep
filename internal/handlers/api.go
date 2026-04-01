// Package handlers provides HTTP request handlers for the S3 File Browser API
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/lausser/s3peep/internal/config"
	"github.com/lausser/s3peep/internal/s3"
)

// APIHandler handles HTTP requests for the S3 File Browser
type APIHandler struct {
	config    *config.Config
	configPath string
	s3Client  *s3.Client
	token     string
	debug     bool
}

// NewAPIHandler creates a new API handler with the given dependencies
func NewAPIHandler(cfg *config.Config, cfgPath string, s3Client *s3.Client, token string, debug bool) *APIHandler {
	return &APIHandler{
		config:     cfg,
		configPath: cfgPath,
		s3Client:   s3Client,
		token:      token,
		debug:      debug,
	}
}

// Handle is the main HTTP handler that routes requests
func (h *APIHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Debug logging
	if h.debug {
		log.Printf("[DEBUG] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract token from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	
	// Check if token is present
	if len(pathParts) == 0 || pathParts[0] == "" {
		h.serveIndex(w, r)
		return
	}
	
	requestToken := pathParts[0]
	
	// Validate token
	if requestToken != h.token {
		if h.debug {
			log.Printf("[DEBUG] Invalid token from %s", r.RemoteAddr)
		}
		h.writeError(w, http.StatusForbidden, "INVALID_TOKEN", "Invalid or expired session token. Please restart s3peep.")
		return
	}
	
	// Get the actual path after token
	var subPath string
	if len(pathParts) > 1 {
		subPath = "/" + strings.Join(pathParts[1:], "/")
	}
	
	// Route to appropriate handler
	switch {
	case subPath == "" || subPath == "/":
		h.serveIndex(w, r)
	case strings.HasPrefix(subPath, "/static/"):
		h.serveStatic(w, r, subPath)
	case strings.HasPrefix(subPath, "/api/"):
		h.handleAPI(w, r, subPath)
	default:
		h.serveIndex(w, r)
	}
}

// serveIndex serves the main HTML page
func (h *APIHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	content := []byte(`<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>S3 File Browser</title>
    <style>
        /* Inline critical CSS */
        .hidden { display: none !important; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .view { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .filter-input { width: 100%; max-width: 400px; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 16px; }
        .bucket-list, .file-list { margin-top: 20px; }
        .bucket-item, .file-item { padding: 15px; border: 1px solid #eee; border-radius: 4px; margin-bottom: 10px; cursor: pointer; transition: background 0.2s; display: flex; align-items: center; justify-content: space-between; }
        .bucket-item:hover, .file-item:hover { background: #f0f0f0; }
        .file-item.selected { background: #e3f2fd; border-color: #2196f3; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin-right: 10px; }
        .btn-primary { background: #0066cc; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-danger:disabled { background: #6c757d; cursor: not-allowed; }
        .file-checkbox { margin-right: 10px; cursor: pointer; }
        .delete-btn { background: #dc3545; color: white; border: none; padding: 5px 10px; border-radius: 4px; cursor: pointer; font-size: 12px; }
        .delete-btn:hover { background: #c82333; }
        .btn-secondary { background: #6c757d; color: white; }
        .empty-state { text-align: center; padding: 60px 20px; color: #666; }
        .error { color: #dc3545; padding: 20px; background: #f8d7da; border-radius: 4px; margin: 20px 0; }
        .loading { text-align: center; padding: 40px; color: #666; }
    </style>
</head>
<body>
    <div id="app">
        <div class="header">
            <h1>S3 File Browser</h1>
            <div id="profile-name">Loading...</div>
        </div>

        <!-- Bucket View -->
        <div id="bucket-view" class="view">
            <h2>Select a Bucket</h2>
            <input type="text" id="bucket-filter" class="filter-input" placeholder="Filter buckets...">
            <div id="bucket-loading" class="loading">Loading buckets...</div>
            <div id="bucket-list" class="bucket-list hidden"></div>
            <div id="bucket-empty" class="empty-state hidden">
                <h3>No buckets found</h3>
                <p>No buckets match your filter criteria.</p>
                <button id="clear-bucket-filter" class="btn btn-primary">Clear Filter</button>
            </div>
            <div id="bucket-error" class="error hidden"></div>
        </div>

        <!-- File View -->
        <div id="file-view" class="view hidden">
            <div style="margin-bottom: 20px;">
                <button id="btn-back" class="btn btn-secondary">← Back to Buckets</button>
            </div>
            <div id="breadcrumb" style="margin-bottom: 20px; padding: 10px; background: #f8f9fa; border-radius: 4px;"></div>
            <div style="margin-bottom: 20px;">
                <button id="btn-create-folder" class="btn btn-primary">+ New Folder</button>
                <button id="btn-upload" class="btn btn-primary">+ Upload File</button>
                <button id="btn-delete-selected" class="btn btn-danger" disabled>Delete Selected</button>
                <input type="file" id="file-upload-input" class="hidden">
            </div>
            <input type="text" id="file-filter" class="filter-input" placeholder="Filter files...">
            <div id="file-loading" class="loading hidden">Loading files...</div>
            <div id="file-list" class="file-list"></div>
            <div id="file-empty" class="empty-state hidden">
                <h3>This folder is empty</h3>
            </div>
        </div>
    </div>

    <script>
        // Simple inline app
        (function() {
            'use strict';
            
            const state = {
                buckets: [],
                filteredBuckets: [],
                currentBucket: null,
                currentPath: '',
                files: [],
                filteredFiles: [],
                selectedItems: new Set()
            };

            // Get token from URL
            const token = window.location.pathname.split('/')[1];
            const baseUrl = '/' + token + '/api';

            // DOM elements
            const els = {
                bucketView: document.getElementById('bucket-view'),
                fileView: document.getElementById('file-view'),
                bucketList: document.getElementById('bucket-list'),
                bucketLoading: document.getElementById('bucket-loading'),
                bucketEmpty: document.getElementById('bucket-empty'),
                bucketError: document.getElementById('bucket-error'),
                bucketFilter: document.getElementById('bucket-filter'),
                clearBucketFilter: document.getElementById('clear-bucket-filter'),
                fileList: document.getElementById('file-list'),
                fileLoading: document.getElementById('file-loading'),
                fileEmpty: document.getElementById('file-empty'),
                fileFilter: document.getElementById('file-filter'),
                breadcrumb: document.getElementById('breadcrumb'),
                btnBack: document.getElementById('btn-back'),
                btnCreateFolder: document.getElementById('btn-create-folder'),
                btnUpload: document.getElementById('btn-upload'),
                btnDeleteSelected: document.getElementById('btn-delete-selected'),
                fileUploadInput: document.getElementById('file-upload-input'),
                profileName: document.getElementById('profile-name')
            };

            // API helper
            async function api(method, endpoint, body) {
                const options = { 
                    method,
                    headers: {}
                };
                if (body) {
                    options.headers['Content-Type'] = 'application/json';
                    options.body = JSON.stringify(body);
                }
                const response = await fetch(baseUrl + endpoint, options);
                if (!response.ok) throw new Error('API error: ' + response.status);
                return response.json();
            }

            // Load buckets
            async function loadBuckets() {
                try {
                    els.bucketLoading.classList.remove('hidden');
                    els.bucketList.classList.add('hidden');
                    els.bucketEmpty.classList.add('hidden');
                    els.bucketError.classList.add('hidden');

                    const result = await api('GET', '/buckets');
                    state.buckets = result.buckets || [];
                    state.filteredBuckets = [...state.buckets];
                    
                    renderBuckets();
                    
                    // Load profile
                    const profile = await api('GET', '/profile');
                    els.profileName.textContent = profile.name || 'Unknown';
                    
                } catch (err) {
                    console.error('Failed to load buckets:', err);
                    els.bucketError.textContent = 'Error: ' + err.message;
                    els.bucketError.classList.remove('hidden');
                } finally {
                    els.bucketLoading.classList.add('hidden');
                }
            }

            // Render buckets
            function renderBuckets() {
                if (state.filteredBuckets.length === 0) {
                    els.bucketList.classList.add('hidden');
                    els.bucketEmpty.classList.remove('hidden');
                    return;
                }

                els.bucketList.innerHTML = state.filteredBuckets.map(b => 
                    '<div class="bucket-item" data-name="' + escapeHtml(b.name) + '">' +
                    '🪣 ' + escapeHtml(b.name) + '</div>'
                ).join('');

                els.bucketList.classList.remove('hidden');
                els.bucketEmpty.classList.add('hidden');

                // Add click handlers
                els.bucketList.querySelectorAll('.bucket-item').forEach(item => {
                    item.addEventListener('click', () => selectBucket(item.dataset.name));
                });
            }

            // Select bucket
            async function selectBucket(name) {
                try {
                    await api('POST', '/buckets/select', { bucket: name });
                    state.currentBucket = name;
                    state.currentPath = '';
                    
                    els.bucketView.classList.add('hidden');
                    els.fileView.classList.remove('hidden');
                    
                    loadFiles();
                } catch (err) {
                    alert('Failed to select bucket: ' + err.message);
                }
            }

            // Load files
            async function loadFiles() {
                try {
                    els.fileLoading.classList.remove('hidden');
                    els.fileList.innerHTML = '';
                    els.fileEmpty.classList.add('hidden');

                    const result = await api('GET', '/buckets/' + encodeURIComponent(state.currentBucket) + '/objects?prefix=' + encodeURIComponent(state.currentPath));
                    state.files = result.objects || [];
                    
                    renderFiles();
                } catch (err) {
                    console.error('Failed to load files:', err);
                    els.fileList.innerHTML = '<div class="error">Error: ' + escapeHtml(err.message) + '</div>';
                } finally {
                    els.fileLoading.classList.add('hidden');
                }
            }

            // Render breadcrumb
            function renderBreadcrumb() {
                if (!state.currentBucket) {
                    els.breadcrumb.innerHTML = '';
                    return;
                }
                
                let html = '<span class="breadcrumb-item" data-path="">' + escapeHtml(state.currentBucket) + '</span>';
                
                if (state.currentPath) {
                    const parts = state.currentPath.split('/').filter(p => p);
                    let currentPath = '';
                    parts.forEach((part, index) => {
                        currentPath += part + '/';
                        const isLast = index === parts.length - 1;
                        html += ' / <span class="breadcrumb-item' + (isLast ? ' breadcrumb-current' : '') + '" data-path="' + currentPath + '">' + escapeHtml(part) + '</span>';
                    });
                }
                
                els.breadcrumb.innerHTML = html;
                
                // Add click handlers to breadcrumb items (except current)
                els.breadcrumb.querySelectorAll('.breadcrumb-item:not(.breadcrumb-current)').forEach(item => {
                    item.style.cursor = 'pointer';
                    item.style.color = '#0066cc';
                    item.addEventListener('click', () => {
                        state.currentPath = item.dataset.path;
                        loadFiles();
                    });
                });
            }

            // Render files
            function renderFiles() {
                // Update breadcrumb
                renderBreadcrumb();
                
                // Clear selection when rendering
                state.selectedItems.clear();
                updateDeleteButton();
                
                // Filter files based on current filter
                const filter = els.fileFilter.value.toLowerCase();
                const filesToShow = filter 
                    ? state.files.filter(f => f.name.toLowerCase().includes(filter))
                    : state.files;
                
                if (filesToShow.length === 0 && !state.currentPath) {
                    els.fileEmpty.classList.remove('hidden');
                    els.fileList.innerHTML = '';
                    return;
                }

                let html = '';
                
                // Add ".." parent folder link if not at root
                if (state.currentPath) {
                    const parentPath = state.currentPath.replace(/[^/]+\/$/, '');
                    html += '<div class="file-item parent-folder" data-key="' + parentPath + '" data-folder="true">' +
                        '<span>📁 .. (Parent Folder)</span>' +
                        '</div>';
                }
                
                // Add files and folders
                html += filesToShow.map(f => {
                    const icon = f.is_folder ? '📁' : '📄';
                    const isSelected = state.selectedItems.has(f.key);
                    return '<div class="file-item' + (isSelected ? ' selected' : '') + '" data-key="' + escapeHtml(f.key) + '" data-folder="' + f.is_folder + '" data-name="' + escapeHtml(f.name) + '">' +
                        '<span class="file-content">' +
                        '<input type="checkbox" class="file-checkbox" data-key="' + escapeHtml(f.key) + '">' +
                        icon + ' ' + escapeHtml(f.name) +
                        '</span>' +
                        '<button class="delete-btn" data-key="' + escapeHtml(f.key) + '" data-name="' + escapeHtml(f.name) + '">Delete</button>' +
                        '</div>';
                }).join('');

                els.fileList.innerHTML = html;
                els.fileEmpty.classList.add('hidden');

                // Add click handlers for navigation and checkboxes
                els.fileList.querySelectorAll('.file-item').forEach(item => {
                    const checkbox = item.querySelector('.file-checkbox');
                    const deleteBtn = item.querySelector('.delete-btn');
                    const content = item.querySelector('.file-content');
                    
                    // Checkbox change handler
                    if (checkbox) {
                        checkbox.addEventListener('change', (e) => {
                            e.stopPropagation();
                            if (e.target.checked) {
                                state.selectedItems.add(item.dataset.key);
                                item.classList.add('selected');
                            } else {
                                state.selectedItems.delete(item.dataset.key);
                                item.classList.remove('selected');
                            }
                            updateDeleteButton();
                        });
                    }
                    
                    // Individual delete button handler
                    if (deleteBtn) {
                        deleteBtn.addEventListener('click', (e) => {
                            e.stopPropagation();
                            const name = deleteBtn.dataset.name;
                            const key = deleteBtn.dataset.key;
                            if (confirm('Are you sure you want to delete "' + name + '"?')) {
                                deleteItems([key]);
                            }
                        });
                    }
                    
                    // Click on item (not checkbox or delete button) for navigation/download
                    item.addEventListener('click', (e) => {
                        if (e.target === checkbox || e.target === deleteBtn) return;
                        
                        if (item.dataset.folder === 'true') {
                            state.currentPath = item.dataset.key;
                            els.fileFilter.value = ''; // Clear filter when navigating
                            loadFiles();
                        } else {
                            downloadFile(item.dataset.key);
                        }
                    });
                });
            }
            
            // Update delete button state
            function updateDeleteButton() {
                els.btnDeleteSelected.disabled = state.selectedItems.size === 0;
                els.btnDeleteSelected.textContent = state.selectedItems.size > 0 
                    ? 'Delete Selected (' + state.selectedItems.size + ')' 
                    : 'Delete Selected';
            }
            
            // Delete items
            async function deleteItems(keys) {
                if (!keys || keys.length === 0) return;
                
                // Check if any folders are being deleted and get their contents
                let allKeys = [...keys];
                let foldersWithContents = [];
                
                for (const key of keys) {
                    if (key.endsWith('/')) {
                        // This is a folder, check if it has contents
                        try {
                            const response = await fetch(baseUrl + '/buckets/' + 
                                encodeURIComponent(state.currentBucket) + '/objects?prefix=' + 
                                encodeURIComponent(key));
                            const data = await response.json();
                            const objects = data.objects || [];
                            
                            if (objects.length > 1) {
                                // More than just the folder marker
                                foldersWithContents.push({
                                    key: key,
                                    count: objects.length
                                });
                                // Add all child objects to the delete list
                                objects.forEach(obj => {
                                    if (!allKeys.includes(obj.key)) {
                                        allKeys.push(obj.key);
                                    }
                                });
                            }
                        } catch (err) {
                            console.error('Failed to check folder contents:', err);
                        }
                    }
                }
                
                // Build confirmation message
                let confirmMsg = '';
                if (foldersWithContents.length > 0) {
                    const folderNames = foldersWithContents.map(f => {
                        const name = f.key.replace(/\/$/, '').split('/').pop();
                        return '"' + name + '" (' + f.count + ' items)';
                    }).join(', ');
                    confirmMsg = 'You are deleting ' + foldersWithContents.length + ' folder(s) with contents: ' + folderNames + '.\n\n';
                }
                
                const itemCount = allKeys.length;
                const itemLabel = itemCount === 1 ? 'item' : 'items';
                confirmMsg += 'Are you sure you want to delete ' + itemCount + ' ' + itemLabel + '?';
                
                if (!confirm(confirmMsg)) {
                    return; // User cancelled
                }
                
                try {
                    const response = await fetch(baseUrl + '/buckets/' + encodeURIComponent(state.currentBucket) + '/objects', {
                        method: 'DELETE',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ keys: allKeys })
                    });
                    
                    if (!response.ok) {
                        const error = await response.json();
                        throw new Error(error.error || 'Delete failed');
                    }
                    
                    const result = await response.json();
                    
                    // Clear selection
                    keys.forEach(key => state.selectedItems.delete(key));
                    updateDeleteButton();
                    
                    // Show results
                    if (result.failed && result.failed.length > 0) {
                        const failedNames = result.failed.map(f => f.key.split('/').pop()).join(', ');
                        alert('Deleted ' + result.deleted.length + ' items. Failed to delete: ' + failedNames);
                    } else {
                        alert('Successfully deleted ' + result.deleted.length + ' items');
                    }
                    
                    loadFiles(); // Refresh file list
                } catch (err) {
                    console.error('Failed to delete:', err);
                    alert('Failed to delete: ' + err.message);
                }
            }

            // Filter files
            els.fileFilter.addEventListener('input', (e) => {
                renderFiles();
            });

            // Download file using anchor tag (more reliable than window.open)
            function downloadFile(key) {
                const downloadUrl = baseUrl + '/buckets/' + encodeURIComponent(state.currentBucket) + '/download?key=' + encodeURIComponent(key);
                const link = document.createElement('a');
                link.href = downloadUrl;
                link.download = key.split('/').pop();
                link.style.display = 'none';
                document.body.appendChild(link);
                link.click();
                setTimeout(() => {
                    document.body.removeChild(link);
                }, 100);
            }

            // Filter buckets
            els.bucketFilter.addEventListener('input', (e) => {
                const filter = e.target.value.toLowerCase();
                state.filteredBuckets = filter 
                    ? state.buckets.filter(b => b.name.toLowerCase().includes(filter))
                    : [...state.buckets];
                renderBuckets();
            });

            // Clear filter
            els.clearBucketFilter.addEventListener('click', () => {
                els.bucketFilter.value = '';
                state.filteredBuckets = [...state.buckets];
                renderBuckets();
            });

            // Back button
            els.btnBack.addEventListener('click', () => {
                els.fileView.classList.add('hidden');
                els.bucketView.classList.remove('hidden');
                state.currentBucket = null;
            });

            // Delete selected button
            els.btnDeleteSelected.addEventListener('click', () => {
                const keys = Array.from(state.selectedItems);
                if (keys.length === 0) return;
                
                const names = keys.map(k => k.split('/').pop()).join(', ');
                if (confirm('Are you sure you want to delete ' + keys.length + ' item(s): ' + names + '?')) {
                    deleteItems(keys);
                }
            });

            // Create folder button
            els.btnCreateFolder.addEventListener('click', () => {
                const folderName = prompt('Enter folder name:');
                if (folderName && folderName.trim()) {
                    createFolder(folderName.trim());
                }
            });

            // Upload button
            els.btnUpload.addEventListener('click', () => {
                els.fileUploadInput.click();
            });

            // File upload input change
            els.fileUploadInput.addEventListener('change', (e) => {
                const file = e.target.files[0];
                if (file) {
                    uploadFile(file);
                    els.fileUploadInput.value = ''; // Reset input
                }
            });

            // Create folder
            async function createFolder(folderName) {
                try {
                    const folderPath = state.currentPath + folderName + '/';
                    await api('PUT', '/buckets/' + encodeURIComponent(state.currentBucket) + '/folders', {
                        folder_path: folderPath
                    });
                    loadFiles(); // Refresh file list
                } catch (err) {
                    console.error('Failed to create folder:', err);
                    alert('Failed to create folder: ' + err.message);
                }
            }

            // Upload file
            async function uploadFile(file) {
                try {
                    const formData = new FormData();
                    const key = state.currentPath + file.name;
                    formData.append('file', file);
                    formData.append('key', key);
                    formData.append('overwrite', 'false');

                    const response = await fetch(baseUrl + '/buckets/' + encodeURIComponent(state.currentBucket) + '/upload', {
                        method: 'POST',
                        body: formData
                    });

                    if (!response.ok) {
                        const error = await response.json();
                        throw new Error(error.error || 'Upload failed');
                    }

                    loadFiles(); // Refresh file list
                } catch (err) {
                    console.error('Failed to upload file:', err);
                    alert('Failed to upload file: ' + err.message);
                }
            }

            // Escape HTML
            function escapeHtml(text) {
                const div = document.createElement('div');
                div.textContent = text;
                return div.innerHTML;
            }

            // Initialize
            loadBuckets();
        })();
    </script>
</body>
</html>`)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

// serveStatic serves static assets (CSS, JS)
func (h *APIHandler) serveStatic(w http.ResponseWriter, r *http.Request, subPath string) {
	// Remove /static/ prefix
	assetPath := strings.TrimPrefix(subPath, "/static/")
	
	// Try multiple possible locations
	possiblePaths := []string{
		filepath.Join("internal", "web", "static", assetPath),
		filepath.Join("..", "..", "internal", "web", "static", assetPath),
		filepath.Join("/home", "lausser", "git", "s3peep", "internal", "web", "static", assetPath),
	}
	
	var content []byte
	var err error
	
	for _, filePath := range possiblePaths {
		content, err = os.ReadFile(filePath)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		log.Printf("Asset not found: %s (tried multiple paths)", assetPath)
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "Asset not found: "+assetPath)
		return
	}
	
	// Set content type based on extension
	if strings.HasSuffix(assetPath, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if strings.HasSuffix(assetPath, ".js") {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}
	
	w.Write(content)
}

// handleAPI routes API requests
func (h *APIHandler) handleAPI(w http.ResponseWriter, r *http.Request, apiPath string) {
	// Remove /api/ prefix
	endpoint := strings.TrimPrefix(apiPath, "/api/")
	
	switch {
	// Buckets endpoints
	case endpoint == "buckets":
		if r.Method == http.MethodGet {
			h.listBuckets(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	case endpoint == "buckets/select":
		if r.Method == http.MethodPost {
			h.selectBucket(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	case strings.HasPrefix(endpoint, "buckets/") && strings.HasSuffix(endpoint, "/objects"):
		h.handleBucketObjects(w, r, endpoint)
	case strings.HasPrefix(endpoint, "buckets/") && strings.HasSuffix(endpoint, "/download"):
		h.handleDownload(w, r, endpoint)
	case strings.HasPrefix(endpoint, "buckets/") && strings.HasSuffix(endpoint, "/upload"):
		h.handleUpload(w, r, endpoint)
	case strings.HasPrefix(endpoint, "buckets/") && strings.HasSuffix(endpoint, "/folders"):
		h.handleCreateFolder(w, r, endpoint)
	
	// Profile endpoint
	case endpoint == "profile":
		if r.Method == http.MethodGet {
			h.getProfile(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	
	// Upload progress endpoint
	case strings.HasPrefix(endpoint, "upload/") && strings.HasSuffix(endpoint, "/progress"):
		h.getUploadProgress(w, r, endpoint)
	
	default:
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "API endpoint not found")
	}
}

// handleBucketObjects handles listing objects in a bucket
func (h *APIHandler) handleBucketObjects(w http.ResponseWriter, r *http.Request, endpoint string) {
	// Extract bucket name from endpoint: buckets/{bucket}/objects
	parts := strings.Split(endpoint, "/")
	if len(parts) < 3 {
		h.writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid bucket path")
		return
	}
	
	bucket := parts[1]
	
	if r.Method == http.MethodGet {
		h.listObjects(w, r, bucket)
	} else if r.Method == http.MethodDelete {
		h.deleteObjects(w, r, bucket)
	} else {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// listBuckets returns all accessible buckets
func (h *APIHandler) listBuckets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	buckets, err := h.s3Client.ListBuckets(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "S3_ERROR", fmt.Sprintf("Failed to list buckets: %v", err))
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"buckets": buckets,
	})
}

// selectBucket sets the active bucket
func (h *APIHandler) selectBucket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Bucket string `json:"bucket"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}
	
	if req.Bucket == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUCKET", "Bucket name is required")
		return
	}
	
	// Update profile
	profile := config.GetActiveProfile(h.config)
	if profile == nil {
		h.writeError(w, http.StatusInternalServerError, "NO_PROFILE", "No active profile")
		return
	}
	
	profile.Bucket = req.Bucket
	
	// Save config
	if err := config.Save(h.config, h.configPath); err != nil {
		h.writeError(w, http.StatusInternalServerError, "SAVE_ERROR", fmt.Sprintf("Failed to save config: %v", err))
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"bucket":  req.Bucket,
		"message": "Bucket selected successfully",
	})
}

// listObjects returns objects in a bucket with pagination
func (h *APIHandler) listObjects(w http.ResponseWriter, r *http.Request, bucket string) {
	ctx := r.Context()
	
	// Get query parameters
	prefix := r.URL.Query().Get("prefix")
	continuationToken := r.URL.Query().Get("continuation_token")
	maxKeysStr := r.URL.Query().Get("max_keys")
	
	maxKeys := 100 // Default
	if maxKeysStr != "" {
		// Parse max_keys, validate against allowed values
		parsed, err := parseInt(maxKeysStr)
		if err == nil && (parsed == 25 || parsed == 50 || parsed == 100 || parsed == 250) {
			maxKeys = parsed
		}
	}
	
	result, err := h.s3Client.ListObjectsPaginated(ctx, bucket, prefix, continuationToken, maxKeys)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "S3_ERROR", fmt.Sprintf("Failed to list objects: %v", err))
		return
	}
	
	h.writeJSON(w, http.StatusOK, result)
}

// handleDownload handles file downloads
func (h *APIHandler) handleDownload(w http.ResponseWriter, r *http.Request, endpoint string) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	
	// Extract bucket name
	parts := strings.Split(endpoint, "/")
	if len(parts) < 3 {
		h.writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid bucket path")
		return
	}
	
	bucket := parts[1]
	key := r.URL.Query().Get("key")
	
	if key == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_KEY", "File key is required")
		return
	}
	
	ctx := r.Context()
	
	// Set download headers (no Content-Length due to AWS SDK v2 bug with MinIO)
	filename := path.Base(key)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	
	// Stream the file
	body, err := h.s3Client.GetObject(ctx, bucket, key)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "DOWNLOAD_ERROR", fmt.Sprintf("Failed to download file: %v", err))
		return
	}
	defer body.Close()
	
	io.Copy(w, body)
}

// handleUpload handles file uploads
func (h *APIHandler) handleUpload(w http.ResponseWriter, r *http.Request, endpoint string) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	
	// Extract bucket name
	parts := strings.Split(endpoint, "/")
	if len(parts) < 3 {
		h.writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid bucket path")
		return
	}
	
	bucket := parts[1]
	
	// Parse multipart form
	r.ParseMultipartForm(32 << 20) // 32MB max memory for form
	
	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_FILE", "No file provided")
		return
	}
	defer file.Close()
	
	key := r.FormValue("key")
	if key == "" {
		key = header.Filename
	}
	
	overwrite := r.FormValue("overwrite") == "true"
	
	ctx := r.Context()
	
	// Check if file exists
	exists, _ := h.s3Client.ObjectExists(ctx, bucket, key)
	if exists && !overwrite {
		h.writeError(w, http.StatusConflict, "FILE_EXISTS", "File already exists")
		return
	}
	
	// Upload file
	err = h.s3Client.UploadObject(ctx, bucket, key, file, header.Size)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "UPLOAD_ERROR", fmt.Sprintf("Failed to upload file: %v", err))
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "completed",
		"key":      key,
		"size":     header.Size,
		"filename": header.Filename,
	})
}

// handleCreateFolder handles folder creation
func (h *APIHandler) handleCreateFolder(w http.ResponseWriter, r *http.Request, endpoint string) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	
	// Extract bucket name
	parts := strings.Split(endpoint, "/")
	if len(parts) < 3 {
		h.writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid bucket path")
		return
	}
	
	bucket := parts[1]
	
	var req struct {
		FolderPath string `json:"folder_path"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}
	
	if req.FolderPath == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_PATH", "Folder path is required")
		return
	}
	
	// Ensure path ends with /
	if !strings.HasSuffix(req.FolderPath, "/") {
		req.FolderPath += "/"
	}
	
	ctx := r.Context()
	
	// Check if folder exists
	exists, _ := h.s3Client.ObjectExists(ctx, bucket, req.FolderPath)
	if exists {
		h.writeError(w, http.StatusConflict, "FOLDER_EXISTS", "Folder already exists")
		return
	}
	
	// Create folder (empty object with trailing slash)
	err := h.s3Client.CreateFolder(ctx, bucket, req.FolderPath)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "CREATE_ERROR", fmt.Sprintf("Failed to create folder: %v", err))
		return
	}
	
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "created",
		"key":    req.FolderPath,
	})
}

// deleteObjects handles object deletion
func (h *APIHandler) deleteObjects(w http.ResponseWriter, r *http.Request, bucket string) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}
	
	var req struct {
		Keys []string `json:"keys"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON body")
		return
	}
	
	if len(req.Keys) == 0 {
		h.writeError(w, http.StatusBadRequest, "MISSING_KEYS", "No keys provided")
		return
	}
	
	ctx := r.Context()
	
	deleted := []string{}
	failed := []map[string]interface{}{}
	
	for _, key := range req.Keys {
		err := h.s3Client.DeleteObject(ctx, bucket, key)
		if err != nil {
			failed = append(failed, map[string]interface{}{
				"key":   key,
				"error": err.Error(),
				"code":  "DELETE_FAILED",
			})
		} else {
			deleted = append(deleted, key)
		}
	}
	
	response := map[string]interface{}{
		"deleted": deleted,
		"failed":  failed,
	}
	
	if len(failed) > 0 {
		w.WriteHeader(http.StatusPartialContent)
	}
	
	h.writeJSON(w, http.StatusOK, response)
}

// getProfile returns profile information
func (h *APIHandler) getProfile(w http.ResponseWriter, r *http.Request) {
	profile := config.GetActiveProfile(h.config)
	if profile == nil {
		h.writeError(w, http.StatusInternalServerError, "NO_PROFILE", "No active profile")
		return
	}
	
	// Return profile info without sensitive fields
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"name":         profile.Name,
		"region":       profile.Region,
		"bucket":       profile.Bucket,
		"endpoint_url": profile.EndpointURL,
	})
}

// getUploadProgress returns upload progress (placeholder for now)
func (h *APIHandler) getUploadProgress(w http.ResponseWriter, r *http.Request, endpoint string) {
	// Extract upload ID
	parts := strings.Split(endpoint, "/")
	if len(parts) < 3 {
		h.writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid upload path")
		return
	}
	
	uploadID := parts[1]
	
	// This is a placeholder - full upload tracking would require additional infrastructure
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"upload_id":           uploadID,
		"status":              "completed",
		"progress_percentage": 100,
		"bytes_uploaded":      0,
		"bytes_total":         0,
		"speed_mbps":          0,
		"eta_seconds":         0,
	})
}

// writeJSON writes a JSON response
func (h *APIHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response
func (h *APIHandler) writeError(w http.ResponseWriter, status int, code string, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error":   message,
		"code":    code,
		"details": "",
	})
}

// parseInt safely parses an integer from string
func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid integer")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// Time-related helpers
type SessionToken struct {
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ProfileName string    `json:"profile_name"`
}

// ValidateToken checks if a token is valid
func (h *APIHandler) ValidateToken(token string) bool {
	return token == h.token
}
