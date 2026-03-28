const API_BASE = '/api';

let currentPath = '';
let files = [];

async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const response = await fetch(url, {
        ...options,
        headers: {
            ...options.headers,
        },
    });
    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }
    return response.json();
}

async function loadFiles(path = '') {
    currentPath = path;
    setStatus('Loading...');
    
    try {
        files = await apiRequest(`/list?prefix=${encodeURIComponent(path)}`);
        renderFiles();
        updateBreadcrumb();
        setStatus(`Showing ${files.length} items`);
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
    
    container.innerHTML = files.map(file => `
        <div class="file-item" data-key="${file.key}" data-is-folder="${file.is_folder}">
            <span class="file-icon ${file.is_folder ? 'folder' : 'file'}">
                ${file.is_folder ? '📁' : '📄'}
            </span>
            <span class="file-name">${escapeHtml(file.name)}</span>
            <span class="file-size">${file.is_folder ? '' : formatSize(file.size)}</span>
            <span class="file-meta">${file.last_modified ? formatDate(file.last_modified) : ''}</span>
        </div>
    `).join('');
    
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

function updateBreadcrumb() {
    const breadcrumb = document.getElementById('breadcrumb');
    const parts = currentPath.split('/').filter(p => p);
    
    let html = '<a href="#" data-path="">Home</a>';
    let path = '';
    
    parts.forEach(part => {
        path += part + '/';
        html += ` <span>/</span> <a href="#" data-path="${path}">${escapeHtml(part)}</a>`;
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
    const url = `${API_BASE}/get?key=${encodeURIComponent(key)}`;
    window.location.href = url;
}

async function uploadFile() {
    const input = document.getElementById('fileInput');
    const file = input.files[0];
    
    if (!file) return;
    
    const key = currentPath + file.name;
    setStatus(`Uploading ${file.name}...`);
    
    try {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('key', key);
        
        const response = await fetch(`${API_BASE}/upload`, {
            method: 'POST',
            body: formData,
        });
        
        if (!response.ok) {
            throw new Error(await response.text());
        }
        
        setStatus(`Uploaded ${file.name}`);
        loadFiles(currentPath);
    } catch (error) {
        showError(error.message);
        setStatus('Upload failed');
    }
    
    input.value = '';
}

function formatSize(bytes) {
    if (!bytes) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
        bytes /= 1024;
        i++;
    }
    return `${bytes.toFixed(1)} ${units[i]}`;
}

function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function setStatus(message) {
    document.getElementById('statusBar').textContent = message;
}

function showError(message) {
    const container = document.getElementById('fileList');
    container.innerHTML = `<div class="error">Error: ${escapeHtml(message)}</div>`;
}

function init() {
    document.getElementById('uploadBtn').addEventListener('click', () => {
        document.getElementById('fileInput').click();
    });
    
    document.getElementById('fileInput').addEventListener('change', uploadFile);
    
    document.getElementById('refreshBtn').addEventListener('click', () => {
        loadFiles(currentPath);
    });
    
    loadFiles();
}

document.addEventListener('DOMContentLoaded', init);
