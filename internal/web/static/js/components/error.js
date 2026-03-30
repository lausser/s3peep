// S3 File Browser - Error Component
// Toast notifications for errors, warnings, and success messages

(function() {
    'use strict';

    // Container for toasts
    let toastContainer = null;
    let toasts = [];
    let toastIdCounter = 0;

    // Initialize toast container
    function init() {
        if (toastContainer) return;

        toastContainer = document.getElementById('toast-container');
        if (!toastContainer) {
            // Create container if it doesn't exist
            toastContainer = document.createElement('div');
            toastContainer.id = 'toast-container';
            toastContainer.className = 'toast-container';
            toastContainer.setAttribute('role', 'status');
            toastContainer.setAttribute('aria-live', 'polite');
            document.body.appendChild(toastContainer);
        }
    }

    // Create a toast element
    function createToastElement(id, title, message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.id = `toast-${id}`;
        toast.setAttribute('role', 'alert');

        const iconMap = {
            success: '✓',
            error: '✕',
            warning: '⚠',
            info: 'ℹ'
        };

        toast.innerHTML = `
            <div class="toast-header">
                <span class="toast-title">
                    <span class="toast-icon" aria-hidden="true">${iconMap[type] || iconMap.info}</span>
                    ${escapeHtml(title)}
                </span>
                <button class="toast-close" aria-label="Close notification" data-toast-id="${id}">
                    ✕
                </button>
            </div>
            ${message ? `<div class="toast-message">${escapeHtml(message)}</div>` : ''}
        `;

        // Add close handler
        const closeBtn = toast.querySelector('.toast-close');
        closeBtn.addEventListener('click', () => removeToast(id));

        return toast;
    }

    // Show a toast notification
    function show(message, type = 'info', duration = 5000) {
        init();

        const id = ++toastIdCounter;
        const title = getTitleForType(type);
        
        const toastEl = createToastElement(id, title, message, type);
        toastContainer.appendChild(toastEl);

        const toast = {
            id,
            element: toastEl,
            timeout: null
        };

        // Auto-remove after duration
        if (duration > 0) {
            toast.timeout = setTimeout(() => removeToast(id), duration);
        }

        toasts.push(toast);

        // Animate in
        requestAnimationFrame(() => {
            toastEl.style.animation = 'slideIn 0.3s ease';
        });

        return id;
    }

    // Show success toast
    function success(message, duration = 5000) {
        return show(message, 'success', duration);
    }

    // Show error toast
    function error(message, duration = 8000) {
        return show(message, 'error', duration);
    }

    // Show warning toast
    function warning(message, duration = 6000) {
        return show(message, 'warning', duration);
    }

    // Show info toast
    function info(message, duration = 5000) {
        return show(message, 'info', duration);
    }

    // Remove a specific toast
    function removeToast(id) {
        const index = toasts.findIndex(t => t.id === id);
        if (index === -1) return;

        const toast = toasts[index];
        
        // Clear timeout
        if (toast.timeout) {
            clearTimeout(toast.timeout);
        }

        // Animate out
        toast.element.style.animation = 'slideOut 0.3s ease';
        
        setTimeout(() => {
            if (toast.element.parentNode) {
                toast.element.parentNode.removeChild(toast.element);
            }
            toasts.splice(index, 1);
        }, 300);
    }

    // Remove all toasts
    function clearAll() {
        toasts.forEach(toast => {
            if (toast.timeout) {
                clearTimeout(toast.timeout);
            }
            if (toast.element.parentNode) {
                toast.element.parentNode.removeChild(toast.element);
            }
        });
        toasts = [];
    }

    // Get title based on type
    function getTitleForType(type) {
        const titles = {
            success: 'Success',
            error: 'Error',
            warning: 'Warning',
            info: 'Info'
        };
        return titles[type] || 'Notification';
    }

    // Escape HTML to prevent XSS
    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Handle API errors
    function handleApiError(err) {
        console.error('API Error:', err);

        let message = err.message || 'An unexpected error occurred';
        let duration = 8000;

        // Handle specific error codes
        if (err.code) {
            switch (err.code) {
                case 'INVALID_TOKEN':
                case 'HTTP_403':
                    message = 'Session expired. Please restart s3peep server.';
                    duration = 10000;
                    break;
                case 'BUCKET_NOT_FOUND':
                    message = 'Bucket not found or no access permission.';
                    break;
                case 'OBJECT_NOT_FOUND':
                    message = 'File or folder not found.';
                    break;
                case 'ACCESS_DENIED':
                    message = 'Access denied. Check your permissions.';
                    break;
                case 'FILE_TOO_LARGE':
                    message = 'File exceeds the 5GB limit.';
                    break;
                case 'NETWORK_ERROR':
                    message = 'Network error. Please check your connection.';
                    break;
                case 'UPLOAD_FAILED':
                    message = 'Upload failed. Please try again.';
                    break;
                case 'DELETE_FAILED':
                    message = 'Delete failed. Some items may not have been removed.';
                    break;
                case 'FOLDER_EXISTS':
                    message = 'A folder with this name already exists.';
                    break;
                case 'FILE_EXISTS':
                    message = 'A file with this name already exists.';
                    break;
            }
        }

        // Handle HTTP status codes
        if (err.status) {
            switch (err.status) {
                case 400:
                    message = message || 'Invalid request';
                    break;
                case 401:
                    case 403:
                    message = 'Access denied or session expired';
                    duration = 10000;
                    break;
                case 404:
                    message = message || 'Not found';
                    break;
                case 409:
                    message = message || 'Conflict - item already exists';
                    break;
                case 413:
                    message = 'File too large';
                    break;
                case 500:
                    case 502:
                    case 503:
                    message = 'Server error. Please try again later.';
                    break;
            }
        }

        return error(message, duration);
    }

    // Export
    window.ErrorComponent = {
        show,
        success,
        error,
        warning,
        info,
        remove: removeToast,
        clearAll,
        handleApiError
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
