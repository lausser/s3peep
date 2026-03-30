// S3 File Browser - Upload Progress Component
// Display upload progress with progress bars

(function() {
    'use strict';

    let panel = null;
    let list = null;
    let closeBtn = null;

    // Initialize the component
    function init() {
        panel = document.getElementById('upload-panel');
        list = document.getElementById('upload-list');
        closeBtn = document.getElementById('btn-close-uploads');

        if (!panel || !list) {
            console.error('Upload panel elements not found');
            return;
        }

        // Close button
        if (closeBtn) {
            closeBtn.addEventListener('click', hide);
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('upload:added', (upload) => {
                show();
                addUploadItem(upload);
            });

            window.State.subscribe('upload:updated', (upload) => {
                updateUploadItem(upload);
            });

            window.State.subscribe('upload:removed', (upload) => {
                removeUploadItem(upload.id);
            });
        }
    }

    // Show upload panel
    function show() {
        if (panel) {
            panel.classList.remove('hidden');
        }
    }

    // Hide upload panel
    function hide() {
        if (panel) {
            panel.classList.add('hidden');
        }
    }

    // Add upload item to list
    function addUploadItem(upload) {
        const item = document.createElement('div');
        item.className = 'upload-item';
        item.id = 'upload-' + upload.id;
        item.innerHTML = createUploadItemHtml(upload);
        list.appendChild(item);
    }

    // Update upload item
    function updateUploadItem(upload) {
        const item = document.getElementById('upload-' + upload.id);
        if (!item) return;

        item.innerHTML = createUploadItemHtml(upload);

        // Auto-remove completed uploads after 5 seconds
        if (upload.status === 'completed') {
            setTimeout(() => {
                if (window.State) {
                    window.State.removeUpload(upload.id);
                }
            }, 5000);
        }
    }

    // Remove upload item
    function removeUploadItem(uploadId) {
        const item = document.getElementById('upload-' + uploadId);
        if (item) {
            item.style.opacity = '0';
            setTimeout(() => {
                item.remove();
                // Hide panel if empty
                if (list.children.length === 0) {
                    hide();
                }
            }, 300);
        }
    }

    // Create HTML for upload item
    function createUploadItemHtml(upload) {
        const statusIcons = {
            pending: '⏳',
            uploading: '📤',
            completed: '✅',
            failed: '❌'
        };

        const statusText = {
            pending: 'Pending',
            uploading: 'Uploading',
            completed: 'Complete',
            failed: 'Failed'
        };

        const progress = upload.progress || 0;
        const speed = upload.speed ? formatSpeed(upload.speed) : '';
        const eta = upload.eta ? formatEta(upload.eta) : '';

        let metaText = statusText[upload.status];
        if (upload.status === 'uploading' && speed) {
            metaText += ` - ${speed}`;
            if (eta) {
                metaText += ` (${eta})`;
            }
        }

        return `
            <div class="upload-item-header">
                <span class="upload-item-name" title="${escapeHtml(upload.file.name)}">
                    ${statusIcons[upload.status]} ${escapeHtml(upload.file.name)}
                </span>
            </div>
            <div class="upload-progress">
                <div class="upload-progress-bar" style="width: ${progress}%"></div>
            </div>
            <div class="upload-item-meta">
                <span>${metaText}</span>
                <span>${progress}%</span>
            </div>
        `;
    }

    // Format speed
    function formatSpeed(bytesPerSecond) {
        if (bytesPerSecond < 1024) {
            return bytesPerSecond + ' B/s';
        } else if (bytesPerSecond < 1024 * 1024) {
            return (bytesPerSecond / 1024).toFixed(1) + ' KB/s';
        } else {
            return (bytesPerSecond / (1024 * 1024)).toFixed(1) + ' MB/s';
        }
    }

    // Format ETA
    function formatEta(seconds) {
        if (seconds < 60) {
            return seconds + 's';
        } else if (seconds < 3600) {
            return Math.floor(seconds / 60) + 'm';
        } else {
            return Math.floor(seconds / 3600) + 'h';
        }
    }

    // Escape HTML
    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Export
    window.UploadProgress = {
        init,
        show,
        hide
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
