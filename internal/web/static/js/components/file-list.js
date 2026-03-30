// S3 File Browser - File List Component
// Displays files and folders in a bucket with icons, sizes, and dates

(function() {
    'use strict';

    let container = null;
    let files = [];
    let filteredFiles = [];
    let isLoading = false;
    let selectedBucket = null;
    let currentPath = '';

    // Initialize the component
    function init() {
        container = document.getElementById('file-list');
        if (!container) {
            console.error('File list container not found');
            return;
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('bucket:selected', (bucket) => {
                selectedBucket = bucket;
                currentPath = '';
                loadFiles();
            });

            window.State.subscribe('folder:navigated', (path) => {
                currentPath = path;
                loadFiles();
            });

            window.State.subscribe('files:filtered', () => {
                filteredFiles = window.State.get('filteredFiles') || [];
                render();
            });

            window.State.subscribe('selection:changed', () => {
                updateSelectionUI();
            });
        }

        // Add click handler for file items (delegated)
        container.addEventListener('click', handleFileClick);

        // Add double-click handler for folders
        container.addEventListener('dblclick', handleFileDoubleClick);
    }

    // Load files from API
    async function loadFiles(page = 1) {
        if (!selectedBucket || isLoading) return;

        isLoading = true;
        showLoading();

        try {
            const pageSize = window.State ? window.State.get('pagination').pageSize : 100;
            const continuationToken = window.State ? window.State.getContinuationToken(page) : '';

            const result = await window.API.listObjects(
                selectedBucket,
                currentPath,
                continuationToken,
                pageSize
            );

            files = result.objects || [];

            // Update pagination state
            if (window.State) {
                window.State.update('pagination', {
                    currentPage: page,
                    isTruncated: result.is_truncated,
                    totalItems: null // S3 doesn't provide total count
                });

                // Store continuation token for next page
                if (result.next_continuation_token) {
                    window.State.storeContinuationToken(page + 1, result.next_continuation_token);
                }

                // Filter files
                window.State.filterFiles();
            }

            filteredFiles = window.State ? window.State.get('filteredFiles') || files : files;

            render();

            // Emit event
            document.dispatchEvent(new CustomEvent('files:loaded', {
                detail: { bucket: selectedBucket, path: currentPath, files }
            }));

        } catch (err) {
            console.error('Failed to load files:', err);
            showError('Failed to load files. Please try again.');

            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
        } finally {
            isLoading = false;
            hideLoading();
        }
    }

    // Handle file item click
    function handleFileClick(e) {
        const fileItem = e.target.closest('.file-item');
        if (!fileItem) return;

        const key = fileItem.dataset.key;
        const isFolder = fileItem.dataset.isFolder === 'true';

        // Handle checkbox click
        if (e.target.closest('.file-checkbox')) {
            e.stopPropagation();
            toggleSelection(key);
            return;
        }

        // Handle folder click (single click navigates)
        if (isFolder) {
            navigateToFolder(key);
            return;
        }

        // Handle file click - download
        downloadFile(key);
    }

    // Handle file double click (for folders)
    function handleFileDoubleClick(e) {
        const fileItem = e.target.closest('.file-item');
        if (!fileItem) return;

        const key = fileItem.dataset.key;
        const isFolder = fileItem.dataset.isFolder === 'true';

        if (isFolder) {
            navigateToFolder(key);
        }
    }

    // Navigate to folder
    function navigateToFolder(folderPath) {
        if (window.State) {
            window.State.navigateToFolder(folderPath);
        }
    }

    // Download a file
    async function downloadFile(key) {
        if (!selectedBucket || !key) return;

        try {
            // Show loading
            const fileItem = container.querySelector(`[data-key="${CSS.escape(key)}"]`);
            if (fileItem) {
                fileItem.classList.add('downloading');
            }

            const blob = await window.API.downloadFile(selectedBucket, key);

            // Get filename from key
            const filename = key.split('/').pop() || 'download';

            // Create download link
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);

            // Show success
            if (window.ErrorComponent) {
                window.ErrorComponent.success(`Downloaded ${filename}`);
            }

        } catch (err) {
            console.error('Download failed:', err);
            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
        } finally {
            // Remove loading state
            const fileItem = container.querySelector(`[data-key="${CSS.escape(key)}"]`);
            if (fileItem) {
                fileItem.classList.remove('downloading');
            }
        }
    }

    // Toggle file selection
    function toggleSelection(key) {
        if (window.State) {
            window.State.toggleFileSelection(key);
        }
    }

    // Update selection UI
    function updateSelectionUI() {
        const selectedFiles = window.State ? window.State.get('selectedFiles') : [];

        container.querySelectorAll('.file-item').forEach(item => {
            const key = item.dataset.key;
            const checkbox = item.querySelector('.file-checkbox');

            if (selectedFiles.includes(key)) {
                item.classList.add('selected');
                if (checkbox) checkbox.checked = true;
            } else {
                item.classList.remove('selected');
                if (checkbox) checkbox.checked = false;
            }
        });
    }

    // Render the file list
    function render() {
        if (!container) return;

        if (isLoading) {
            container.innerHTML = '';
            return;
        }

        if (filteredFiles.length === 0) {
            container.innerHTML = '';
            document.dispatchEvent(new CustomEvent('files:empty', {
                detail: {
                    hasFilter: window.State && window.State.get('fileFilter'),
                    isFolder: currentPath !== ''
                }
            }));
            return;
        }

        document.dispatchEvent(new CustomEvent('files:not-empty'));

        const html = filteredFiles.map(file => createFileHtml(file)).join('');
        container.innerHTML = html;

        // Update selection state
        updateSelectionUI();
    }

    // Create HTML for a single file
    function createFileHtml(file) {
        const isFolder = file.is_folder;
        const icon = isFolder ? '📁' : getFileIcon(file.file_type || 'other');
        const name = escapeHtml(file.name);
        const size = isFolder ? '' : formatSize(file.size);
        const date = formatDate(file.last_modified);

        const selectedFiles = window.State ? window.State.get('selectedFiles') : [];
        const isSelected = selectedFiles.includes(file.key);

        return `
            <div class="file-item ${isSelected ? 'selected' : ''}"
                 data-key="${escapeHtml(file.key)}"
                 data-is-folder="${isFolder}"
                 role="button"
                 tabindex="0"
                 aria-label="${isFolder ? 'Folder' : 'File'} ${name}">
                <input type="checkbox"
                       class="file-checkbox"
                       ${isSelected ? 'checked' : ''}
                       aria-label="Select ${name}">
                <div class="file-icon ${isFolder ? 'folder' : ''}">${icon}</div>
                <div class="file-info">
                    <div class="file-name">${name}</div>
                    ${!isFolder ? `<div class="file-path">${escapeHtml(file.key)}</div>` : ''}
                </div>
                <div class="file-meta">
                    ${size ? `<span class="file-size">${size}</span>` : ''}
                    <span class="file-date">${date}</span>
                </div>
            </div>
        `;
    }

    // Get icon for file type
    function getFileIcon(fileType) {
        const icons = {
            folder: '📁',
            image: '🖼️',
            document: '📄',
            archive: '📦',
            video: '🎬',
            audio: '🎵',
            code: '💻',
            other: '📎'
        };
        return icons[fileType] || icons.other;
    }

    // Format size
    function formatSize(bytes) {
        if (bytes === 0) return '0 B';
        if (!bytes || isNaN(bytes)) return '-';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // Format date
    function formatDate(dateStr) {
        if (!dateStr) return '-';

        const date = new Date(dateStr);
        if (isNaN(date.getTime())) return '-';

        return date.toLocaleString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // Escape HTML
    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Show loading state
    function showLoading() {
        const skeleton = document.getElementById('file-skeleton');
        if (skeleton) {
            skeleton.classList.remove('hidden');
        }
        if (container) {
            container.classList.add('hidden');
        }
    }

    // Hide loading state
    function hideLoading() {
        const skeleton = document.getElementById('file-skeleton');
        if (skeleton) {
            skeleton.classList.add('hidden');
        }
        if (container) {
            container.classList.remove('hidden');
        }
    }

    // Show error state
    function showError(message) {
        if (container) {
            container.innerHTML = `
                <div class="file-error">
                    <div class="error-icon">⚠️</div>
                    <p>${escapeHtml(message)}</p>
                    <button class="btn btn-secondary" onclick="FileList.refresh()">Retry</button>
                </div>
            `;
        }
    }

    // Refresh file list
    function refresh() {
        const page = window.State ? window.State.get('pagination').currentPage : 1;
        loadFiles(page);
    }

    // Go to specific page
    function goToPage(page) {
        loadFiles(page);
    }

    // Get current files
    function getFiles() {
        return files;
    }

    // Get filtered files
    function getFilteredFiles() {
        return filteredFiles;
    }

    // Get current path
    function getCurrentPath() {
        return currentPath;
    }

    // Export
    window.FileList = {
        init,
        refresh,
        goToPage,
        getFiles,
        getFilteredFiles,
        getCurrentPath,
        navigateToFolder,
        downloadFile
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
