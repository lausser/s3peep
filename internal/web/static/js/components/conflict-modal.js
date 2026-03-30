// S3 File Browser - Conflict Modal Component
// Handle file name conflicts during upload

(function() {
    'use strict';

    let modal = null;
    let applyToAllCheckbox = null;
    let currentUpload = null;
    let onDecisionCallback = null;
    let applyToAll = false;

    // Initialize the component
    function init() {
        modal = document.getElementById('conflict-modal');
        if (!modal) {
            console.error('Conflict modal not found');
            return;
        }

        // Apply to all checkbox
        applyToAllCheckbox = document.getElementById('apply-to-all');

        // Close buttons
        modal.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', close);
        });

        // Decision buttons
        const btnSkip = document.getElementById('btn-skip');
        const btnRename = document.getElementById('btn-rename');
        const btnReplace = document.getElementById('btn-replace');

        if (btnSkip) {
            btnSkip.addEventListener('click', () => makeDecision('skip'));
        }

        if (btnRename) {
            btnRename.addEventListener('click', () => makeDecision('rename'));
        }

        if (btnReplace) {
            btnReplace.addEventListener('click', () => makeDecision('replace'));
        }

        // Close on backdrop click
        const backdrop = modal.querySelector('.modal-backdrop');
        if (backdrop) {
            backdrop.addEventListener('click', close);
        }

        // Close on Escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && isOpen()) {
                close();
            }
        });
    }

    // Show conflict modal
    function show(upload, onDecision) {
        currentUpload = upload;
        onDecisionCallback = onDecision;
        applyToAll = false;

        // Reset apply to all
        if (applyToAllCheckbox) {
            applyToAllCheckbox.checked = false;
        }

        // Load file info
        loadFileInfo(upload);

        // Show modal
        modal.classList.remove('hidden');

        // Focus replace button (default action)
        const btnReplace = document.getElementById('btn-replace');
        if (btnReplace) {
            btnReplace.focus();
        }
    }

    // Load file information into modal
    async function loadFileInfo(upload) {
        try {
            // Get existing file info
            const existingInfo = await getExistingFileInfo(upload.bucket, upload.key);

            // Update UI
            document.getElementById('conflict-existing-name').textContent = upload.file.name;
            document.getElementById('conflict-existing-size').textContent = existingInfo.size ?
                formatSize(existingInfo.size) : 'Unknown size';
            document.getElementById('conflict-existing-date').textContent = existingInfo.date ?
                formatDate(existingInfo.date) : 'Unknown date';

            document.getElementById('conflict-new-name').textContent = upload.file.name;
            document.getElementById('conflict-new-size').textContent = formatSize(upload.file.size);
            document.getElementById('conflict-new-date').textContent = 'Now';

        } catch (err) {
            console.error('Failed to load file info:', err);
            // Show basic info anyway
            document.getElementById('conflict-existing-name').textContent = upload.file.name;
            document.getElementById('conflict-new-name').textContent = upload.file.name;
        }
    }

    // Get existing file info from S3
    async function getExistingFileInfo(bucket, key) {
        try {
            // Try to get metadata via HEAD request
            const result = await window.API.headObject(bucket, key);
            return {
                size: result.content_length,
                date: result.last_modified
            };
        } catch (err) {
            return {};
        }
    }

    // Make decision
    function makeDecision(decision) {
        // Check apply to all
        if (applyToAllCheckbox) {
            applyToAll = applyToAllCheckbox.checked;
        }

        close();

        // Call callback
        if (onDecisionCallback) {
            onDecisionCallback(currentUpload, decision);
        }

        // If apply to all, store preference
        if (applyToAll) {
            window.uploadConflictPreference = decision;
        }
    }

    // Close modal
    function close() {
        modal.classList.add('hidden');
        currentUpload = null;
        onDecisionCallback = null;
    }

    // Check if modal is open
    function isOpen() {
        return !modal.classList.contains('hidden');
    }

    // Format size
    function formatSize(bytes) {
        if (!bytes || bytes === 0) return '0 B';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // Format date
    function formatDate(dateStr) {
        if (!dateStr) return '';

        const date = new Date(dateStr);
        if (isNaN(date.getTime())) return dateStr;

        return date.toLocaleString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // Export
    window.ConflictModal = {
        init,
        show,
        close,
        isOpen
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
