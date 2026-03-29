package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/lausser/s3peep/internal/config"
	"github.com/lausser/s3peep/internal/s3"
)

var webAssets = map[string][]byte{}

func LoadWebAssets() error {
	webAssets["/index.html"] = []byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>S3 File Browser</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>S3 File Browser</h1>
            <div class="bucket-selector">
                <select id="bucketSelect">
                    <option value="">Select a bucket...</option>
                </select>
            </div>
        </header>
        
        <div class="breadcrumb" id="breadcrumb">
            <a href="#" data-path="">Home</a>
        </div>
        
        <div class="toolbar">
            <button id="uploadBtn" class="btn">Upload File</button>
            <button id="refreshBtn" class="btn">Refresh</button>
        </div>
        
        <div class="file-list" id="fileList">
            <div class="loading">Loading...</div>
        </div>
        
        <div class="status-bar" id="statusBar">
            Ready
        </div>
    </div>
    
    <input type="file" id="fileInput" style="display: none;">
    
    <script src="/static/app.js"></script>
</body>
</html>`)

	webAssets["/static/styles.css"] = []byte(`* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #f5f5f5;
    color: #333;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 20px;
    border-bottom: 1px solid #ddd;
}

h1 {
    font-size: 24px;
}

.bucket-selector select {
    padding: 8px 12px;
    font-size: 14px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background: white;
    min-width: 200px;
}

.breadcrumb {
    margin-bottom: 15px;
    font-size: 14px;
}

.breadcrumb a {
    color: #0066cc;
    text-decoration: none;
}

.breadcrumb a:hover {
    text-decoration: underline;
}

.breadcrumb span {
    margin: 0 5px;
    color: #999;
}

.toolbar {
    margin-bottom: 15px;
}

.btn {
    padding: 8px 16px;
    margin-right: 10px;
    border: none;
    border-radius: 4px;
    background: #0066cc;
    color: white;
    cursor: pointer;
    font-size: 14px;
}

.btn:hover {
    background: #0052a3;
}

.file-list {
    background: white;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    min-height: 300px;
}

.file-item {
    display: flex;
    align-items: center;
    padding: 12px 15px;
    border-bottom: 1px solid #eee;
    cursor: pointer;
}

.file-item:hover {
    background: #f8f9fa;
}

.file-icon {
    margin-right: 12px;
    font-size: 20px;
}

.file-icon.folder {
    color: #ffc107;
}

.file-name {
    flex: 1;
    font-weight: 500;
}

.file-meta {
    color: #666;
    font-size: 13px;
}

.file-size {
    width: 100px;
    text-align: right;
    color: #666;
    font-size: 13px;
}

.loading, .empty {
    padding: 40px;
    text-align: center;
    color: #666;
}

.status-bar {
    margin-top: 15px;
    padding: 10px;
    background: #f8f9fa;
    border-radius: 4px;
    font-size: 13px;
    color: #666;
}`)

	webAssets["/static/app.js"] = []byte(`const API_BASE = '/api';

let currentPath = '';
let currentBucket = '';
let files = [];

async function apiRequest(endpoint) {
    const response = await fetch(API_BASE + endpoint);
    if (!response.ok) {
        throw new Error(await response.text());
    }
    return response.json();
}

async function loadBuckets() {
    try {
        const buckets = await apiRequest('/buckets');
        const select = document.getElementById('bucketSelect');
        select.innerHTML = '<option value="">Select a bucket...</option>';
        buckets.forEach(bucket => {
            const option = document.createElement('option');
            option.value = bucket.name;
            option.textContent = bucket.name;
            if (bucket.name === currentBucket) {
                option.selected = true;
            }
            select.appendChild(option);
        });
        if (!currentBucket) {
            document.getElementById('fileList').innerHTML = '<div class="empty">Select a bucket to browse files</div>';
        }
    } catch (error) {
        showError(error.message);
    }
}

async function loadFiles(path = '') {
    currentPath = path;
    setStatus('Loading...');
    
    try {
        files = await apiRequest('/list?prefix=' + encodeURIComponent(path));
        renderFiles();
        updateBreadcrumb();
        setStatus('Showing ' + files.length + ' items');
    } catch (error) {
        showError(error.message);
        setStatus('Error loading files');
    }
}

function renderFiles() {
    const container = document.getElementById('fileList');
    
    if (files.length === 0) {
        container.innerHTML = '<div class="empty">This folder is empty</div>';
        return;
    }
    
    container.innerHTML = files.map(file => {
        const escapedName = file.name ? file.name.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;') : '';
        const key = file.key || '';
        const isFolder = file.is_folder || false;
        const size = file.size || 0;
        return '<div class="file-item" data-key="' + key + '" data-is-folder="' + isFolder + '">' +
            '<span class="file-icon ' + (isFolder ? 'folder' : 'file') + '">' + (isFolder ? '📁' : '📄') + '</span>' +
            '<span class="file-name">' + escapedName + '</span>' +
            '<span class="file-size">' + (isFolder ? '' : formatSize(size)) + '</span>' +
            '</div>';
    }).join('');
    
    container.querySelectorAll('.file-item').forEach(item => {
        item.addEventListener('click', () => {
            if (item.dataset.isFolder === 'true') {
                loadFiles(item.dataset.key);
            } else {
                downloadFile(item.dataset.key);
            }
        });
    });
}

function selectBucket(bucket) {
    currentBucket = bucket;
    currentPath = '';
    if (bucket) {
        loadFiles('');
    } else {
        document.getElementById('fileList').innerHTML = '<div class="empty">Select a bucket to browse files</div>';
    }
}

async function setBucket(bucket) {
    try {
        await fetch(API_BASE + '/buckets', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({bucket: bucket})
        });
        selectBucket(bucket);
    } catch (error) {
        showError(error.message);
    }
}

function updateBreadcrumb() {
    const breadcrumb = document.getElementById('breadcrumb');
    const parts = currentPath.split('/').filter(p => p);
    
    let html = '<a href="#" data-path="">Home</a>';
    let path = '';
    
    parts.forEach(part => {
        path += part + '/';
        const escapedPart = part.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        html += ' <span>/</span> <a href="#" data-path="' + path + '">' + escapedPart + '</a>';
    });
    
    breadcrumb.innerHTML = html;
    
    breadcrumb.querySelectorAll('a').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            loadFiles(link.dataset.path);
        });
    });
}

function downloadFile(key) {
    window.location.href = API_BASE + '/get?key=' + encodeURIComponent(key);
}

function formatSize(bytes) {
    if (!bytes) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024;
        i++;
    }
    return bytes.toFixed(1) + ' ' + units[i];
}

function setStatus(message) {
    document.getElementById('statusBar').textContent = message;
}

function showError(message) {
    const container = document.getElementById('fileList');
    const escapedMsg = message.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    container.innerHTML = '<div class="error">' + escapedMsg + '</div>';
}

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('bucketSelect').addEventListener('change', (e) => {
        setBucket(e.target.value);
    });
    
    document.getElementById('uploadBtn').addEventListener('click', () => {
        document.getElementById('fileInput').click();
    });
    
    document.getElementById('fileInput').addEventListener('change', () => {
        const file = document.getElementById('fileInput').files[0];
        if (file) {
            setStatus('Upload not implemented yet');
        }
    });
    
    document.getElementById('refreshBtn').addEventListener('click', () => {
        if (currentBucket) {
            loadFiles(currentPath);
        } else {
            loadBuckets();
        }
    });
    
    loadBuckets();
});`)

	return nil
}

// S3Service defines the S3 operations used by the API handler.
type S3Service interface {
	ListBuckets(ctx context.Context) ([]s3.Bucket, error)
	ListObjects(ctx context.Context, prefix string) ([]s3.FileObject, error)
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
	SetBucket(bucket string)
}

type APIHandler struct {
	cfg         *config.Config
	configPath  string
	s3Client    S3Service
}

func NewAPIHandler(cfg *config.Config, configPath string, s3Client S3Service) *APIHandler {
	if len(webAssets) == 0 {
		LoadWebAssets()
	}
	return &APIHandler{
		cfg:        cfg,
		configPath: configPath,
		s3Client:   s3Client,
	}
}

func (h *APIHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	
	if strings.HasPrefix(path, "static/") {
		h.serveStatic(w, r, path)
		return
	}
	
	if path == "api" || strings.HasPrefix(path, "api/") {
		apiPath := strings.TrimPrefix(path, "api/")
		if apiPath == "" {
			apiPath = "/"
		}
		h.handleAPI(w, r, strings.SplitN(apiPath, "/", 2))
		return
	}

	if path == "" || path == "index.html" {
		h.handleIndex(w, r)
		return
	}

	http.NotFound(w, r)
}

func (h *APIHandler) serveStatic(w http.ResponseWriter, r *http.Request, path string) {
	contentTypes := map[string]string{
		"static/styles.css": "text/css",
		"static/app.js":     "application/javascript",
	}

	contentType, ok := contentTypes[path]
	if !ok {
		contentType = "application/octet-stream"
	}

	data, ok := webAssets["/"+path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func (h *APIHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>S3 File Browser</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .file { padding: 8px; border-bottom: 1px solid #eee; }
        .folder { font-weight: bold; color: #0066cc; }
        .file-info { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <h1>S3 File Browser</h1>
    <p>Loading...</p>
    <div id="content"></div>
    <script>
        async function loadFiles(prefix = '') {
            const resp = await fetch('/api/list?prefix=' + encodeURIComponent(prefix));
            const files = await resp.json();
            let html = '<a href="#" onclick="loadFiles(\'\'); return false;">Home</a>';
            if (prefix) {
                const parent = prefix.split('/').slice(0, -1).join('/');
                html += ' | <a href="#" onclick="loadFiles(\'' + encodeURIComponent(parent) + '\'); return false;">Parent</a>';
            }
            html += '<br><br>';
            files.forEach(f => {
                if (f.is_folder) {
                    html += '<div class="file folder">📁 ' + f.name + '</div>';
                } else {
                    html += '<div class="file">📄 ' + f.name + ' <span class="file-info">(' + f.size + ' bytes)</span></div>';
                }
            });
            document.getElementById('content').innerHTML = html;
        }
        loadFiles();
    </script>
</body>
</html>`))
}

func (h *APIHandler) handleAPI(w http.ResponseWriter, r *http.Request, segments []string) {
	if len(segments) == 0 || segments[0] == "" {
		http.NotFound(w, r)
		return
	}

	switch segments[0] {
	case "list":
		h.handleList(w, r)
	case "get":
		h.handleGet(w, r)
	case "profile":
		h.handleProfile(w, r)
	case "buckets":
		h.handleBuckets(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *APIHandler) handleList(w http.ResponseWriter, r *http.Request) {
	profile := config.GetActiveProfile(h.cfg)
	if profile == nil || profile.Bucket == "" {
		buckets, err := h.s3Client.ListBuckets(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list buckets: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buckets)
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = ""
	}

	files, err := h.s3Client.ListObjects(r.Context(), prefix)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list objects: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *APIHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	obj, err := h.s3Client.GetObject(r.Context(), key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path.Base(key)))
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, obj)
}

func (h *APIHandler) handleBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		buckets, err := h.s3Client.ListBuckets(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list buckets: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buckets)
		return
	}

	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data struct {
			Bucket string `json:"bucket"`
		}
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		profile := config.GetActiveProfile(h.cfg)
		if profile == nil {
			http.Error(w, "No active profile", http.StatusBadRequest)
			return
		}

		profile.Bucket = data.Bucket
		if err := config.Save(h.cfg, h.configPath); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.s3Client.SetBucket(data.Bucket)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "bucket": data.Bucket})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *APIHandler) handleProfile(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Query().Get("action") {
	case "list":
		h.handleProfileList(w, r)
	case "switch":
		h.handleProfileSwitch(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *APIHandler) handleProfileList(w http.ResponseWriter, r *http.Request) {
	profiles := make([]string, len(h.cfg.Profiles))
	for i, p := range h.cfg.Profiles {
		profiles[i] = p.Name
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_profile": h.cfg.ActiveProfile,
		"profiles":      profiles,
	})
}

func (h *APIHandler) handleProfileSwitch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := config.SwitchProfile(h.cfg, name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := config.Save(h.cfg, h.configPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.handleProfileList(w, r)
}

func (h *APIHandler) handleProfileAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var newProfile config.Profile
	if err := json.Unmarshal(body, &newProfile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := config.AddProfile(h.cfg, newProfile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := config.Save(h.cfg, h.configPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
