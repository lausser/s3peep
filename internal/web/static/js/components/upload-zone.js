// S3 File Browser - Upload Zone Component
// Drag and drop file upload with visual feedback

(function() {
    'use strict';

    let dropZone = null;
    let overlay = null;
    let isDragging = false;

    // Initialize the component
    function init() {
        dropZone = document.getElementById('file-view');
        overlay = document.getElementById('drag-overlay');

        if (!dropZone) {
            console.error('File view container not found for drag-drop');
            return;
        }

        // Prevent default drag behaviors
        ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
            dropZone.addEventListener(eventName, preventDefaults, false);
            document.body.addEventListener(eventName, preventDefaults, false);
        });

        // Highlight drop zone when item is dragged over
        ['dragenter', 'dragover'].forEach(eventName => {
            dropZone.addEventListener(eventName, highlight, false);
        });

        // Remove highlight when item is dragged out
        ['dragleave', 'drop'].forEach(eventName => {
            dropZone.addEventListener(eventName, unhighlight, false);
        });

        // Handle dropped files
        dropZone.addEventListener('drop', handleDrop, false);
    }

    // Prevent default behaviors
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }

    // Show drag overlay
    function highlight(e) {
        if (!isFileDrag(e)) return;

        isDragging = true;
        if (overlay) {
            overlay.classList.remove('hidden');
        }
        dropZone.classList.add('drag-active');
    }

    // Hide drag overlay
    function unhighlight(e) {
        // Only unhighlight if we're leaving the drop zone, not entering a child
        if (e.relatedTarget && dropZone.contains(e.relatedTarget)) return;

        isDragging = false;
        if (overlay) {
            overlay.classList.add('hidden');
        }
        dropZone.classList.remove('drag-active');
    }

    // Check if drag contains files
    function isFileDrag(e) {
        if (e.dataTransfer.types) {
            for (let i = 0; i < e.dataTransfer.types.length; i++) {
                if (e.dataTransfer.types[i] === 'Files') {
                    return true;
                }
            }
        }
        return false;
    }

    // Handle dropped files
    function handleDrop(e) {
        unhighlight(e);

        const dt = e.dataTransfer;
        const files = dt.files;

        if (files.length === 0) return;

        // Get current bucket and path from state
        const bucket = window.State ? window.State.get('selectedBucket') : null;
        const path = window.State ? window.State.get('currentPath') : '';

        if (!bucket) {
            if (window.ErrorComponent) {
                window.ErrorComponent.error('No bucket selected. Please select a bucket first.');
            }
            return;
        }

        // Queue files for upload
        handleFiles(files, bucket, path);
    }

    // Process files for upload
    function handleFiles(files, bucket, path) {
        Array.from(files).forEach(file => {
            const key = path + file.name;

            // Add to upload queue
            const upload = {
                id: generateId(),
                file: file,
                bucket: bucket,
                key: key,
                status: 'pending',
                progress: 0,
                speed: 0,
                eta: 0
            };

            if (window.State) {
                window.State.addUpload(upload);
            }

            // Check for conflicts
            checkConflict(upload);
        });
    }

    // Check if file already exists
    async function checkConflict(upload) {
        try {
            // Check if file exists
            const exists = await checkFileExists(upload.bucket, upload.key);

            if (exists) {
                // Show conflict modal
                if (window.ConflictModal) {
                    window.ConflictModal.show(upload, handleConflictDecision);
                } else {
                    // Default to auto-rename if modal not available
                    upload.key = addTimestampToKey(upload.key);
                    startUpload(upload);
                }
            } else {
                startUpload(upload);
            }
        } catch (err) {
            console.error('Conflict check failed:', err);
            // Start upload anyway
            startUpload(upload);
        }
    }

    // Handle conflict resolution decision
    function handleConflictDecision(upload, decision) {
        switch (decision) {
            case 'replace':
                startUpload(upload);
                break;
            case 'rename':
                upload.key = addTimestampToKey(upload.key);
                startUpload(upload);
                break;
            case 'skip':
                if (window.State) {
                    window.State.removeUpload(upload.id);
                }
                break;
        }
    }

    // Start file upload
    async function startUpload(upload) {
        if (window.State) {
            window.State.updateUpload(upload.id, { status: 'uploading' });
        }

        try {
            await window.API.uploadFile(
                upload.bucket,
                upload.key,
                upload.file,
                true // overwrite (we already handled conflicts)
            );

            // Upload complete
            if (window.State) {
                window.State.updateUpload(upload.id, {
                    status: 'completed',
                    progress: 100
                });
            }

            if (window.ErrorComponent) {
                window.ErrorComponent.success(`Uploaded ${upload.file.name}`);
            }

            // Refresh file list
            if (window.FileList) {
                window.FileList.refresh();
            }

        } catch (err) {
            console.error('Upload failed:', err);

            if (window.State) {
                window.State.updateUpload(upload.id, {
                    status: 'failed',
                    error: err.message
                });
            }

            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
        }
    }

    // Check if file exists in S3
    async function checkFileExists(bucket, key) {
        try {
            await window.API.headObject(bucket, key);
            return true;
        } catch (err) {
            return false;
        }
    }

    // Add timestamp to key for auto-rename
    function addTimestampToKey(key) {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
        const parts = key.split('.');
        if (parts.length > 1) {
            const ext = parts.pop();
            return parts.join('.') + '_' + timestamp + '.' + ext;
        }
        return key + '_' + timestamp;
    }

    // Generate unique ID
    function generateId() {
        return 'upload-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    // Export
    window.UploadZone = {
        init,
        handleFiles
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
