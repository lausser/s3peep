// S3 File Browser - Empty State Component
// Shows empty state messages for bucket and file lists

(function() {
    'use strict';

    // Initialize the component
    function init() {
        // Subscribe to events
        document.addEventListener('buckets:empty', (e) => {
            showBucketEmpty(e.detail.hasFilter);
        });

        document.addEventListener('buckets:not-empty', () => {
            hideBucketEmpty();
        });

        document.addEventListener('files:empty', (e) => {
            showFileEmpty(e.detail.hasFilter, e.detail.isFolder);
        });

        document.addEventListener('files:not-empty', () => {
            hideFileEmpty();
        });
    }

    // Show bucket empty state
    function showBucketEmpty(hasFilter) {
        const emptyEl = document.getElementById('bucket-empty');
        if (!emptyEl) return;

        if (hasFilter) {
            emptyEl.querySelector('h3').textContent = 'No buckets match';
            emptyEl.querySelector('p').textContent = 'No buckets match your filter criteria.';
            const clearBtn = emptyEl.querySelector('button');
            if (clearBtn) clearBtn.classList.remove('hidden');
        } else {
            emptyEl.querySelector('h3').textContent = 'No buckets found';
            emptyEl.querySelector('p').textContent = 'No S3 buckets are accessible with your credentials.';
            const clearBtn = emptyEl.querySelector('button');
            if (clearBtn) clearBtn.classList.add('hidden');
        }

        emptyEl.classList.remove('hidden');
        
        // Hide list
        const list = document.getElementById('bucket-list');
        if (list) list.classList.add('hidden');
    }

    // Hide bucket empty state
    function hideBucketEmpty() {
        const emptyEl = document.getElementById('bucket-empty');
        if (emptyEl) {
            emptyEl.classList.add('hidden');
        }

        // Show list
        const list = document.getElementById('bucket-list');
        if (list) list.classList.remove('hidden');
    }

    // Show file empty state
    function showFileEmpty(hasFilter, isFolder) {
        const emptyEl = document.getElementById('file-empty');
        const filterEmptyEl = document.getElementById('file-filter-empty');
        
        if (hasFilter) {
            // Show filter empty state
            if (filterEmptyEl) filterEmptyEl.classList.remove('hidden');
            if (emptyEl) emptyEl.classList.add('hidden');
        } else {
            // Show folder empty state
            if (filterEmptyEl) filterEmptyEl.classList.add('hidden');
            if (emptyEl) {
                emptyEl.querySelector('h3').textContent = isFolder ? 'This folder is empty' : 'No files';
                emptyEl.classList.remove('hidden');
            }
        }

        // Hide file list
        const list = document.getElementById('file-list');
        if (list) list.classList.add('hidden');
    }

    // Hide file empty state
    function hideFileEmpty() {
        const emptyEl = document.getElementById('file-empty');
        const filterEmptyEl = document.getElementById('file-filter-empty');
        
        if (emptyEl) emptyEl.classList.add('hidden');
        if (filterEmptyEl) filterEmptyEl.classList.add('hidden');

        // Show file list
        const list = document.getElementById('file-list');
        if (list) list.classList.remove('hidden');
    }

    // Export
    window.EmptyState = {
        init,
        showBucketEmpty,
        hideBucketEmpty,
        showFileEmpty,
        hideFileEmpty
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
