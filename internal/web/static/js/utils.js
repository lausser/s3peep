// S3 File Browser - Utility Functions
// Shared utility functions used across the application

/**
 * Debounce function to limit how often a function can fire
 * @param {Function} func - Function to debounce
 * @param {number} wait - Milliseconds to wait
 * @returns {Function} Debounced function
 */
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

/**
 * Format bytes to human-readable string
 * @param {number} bytes - Size in bytes
 * @param {number} decimals - Number of decimal places
 * @returns {string} Formatted size (e.g., "1.5 MB")
 */
function formatSize(bytes, decimals = 2) {
    if (bytes === 0) return '0 B';
    if (!bytes || isNaN(bytes)) return '-';
    
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
    
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

/**
 * Format date to localized string
 * @param {string|Date} date - Date to format
 * @param {Object} options - Intl.DateTimeFormat options
 * @returns {string} Formatted date
 */
function formatDate(date, options = {}) {
    if (!date) return '-';
    
    const d = new Date(date);
    if (isNaN(d.getTime())) return '-';
    
    const defaultOptions = {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        ...options
    };
    
    return d.toLocaleString(undefined, defaultOptions);
}

/**
 * Escape HTML special characters to prevent XSS
 * @param {string} text - Text to escape
 * @returns {string} Escaped text
 */
function escapeHtml(text) {
    if (!text) return '';
    
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Get file extension from filename
 * @param {string} filename - Filename
 * @returns {string} Lowercase extension with dot
 */
function getFileExtension(filename) {
    if (!filename) return '';
    const ext = filename.split('.').pop();
    return ext ? '.' + ext.toLowerCase() : '';
}

/**
 * Determine file type from extension
 * @param {string} filename - Filename
 * @param {boolean} isFolder - Whether this is a folder
 * @returns {string} File type category
 */
function getFileType(filename, isFolder = false) {
    if (isFolder) return 'folder';
    
    const ext = getFileExtension(filename);
    
    const typeMap = {
        image: ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.bmp', '.ico', '.tiff'],
        document: ['.pdf', '.doc', '.docx', '.txt', '.md', '.csv', '.xls', '.xlsx', '.ppt', '.pptx', '.odt', '.ods'],
        archive: ['.zip', '.tar', '.gz', '.bz2', '.7z', '.rar', '.tgz', '.xz'],
        video: ['.mp4', '.avi', '.mov', '.mkv', '.webm', '.flv', '.wmv'],
        audio: ['.mp3', '.wav', '.flac', '.aac', '.ogg', '.m4a', '.wma'],
        code: ['.js', '.py', '.go', '.java', '.cpp', '.c', '.h', '.hpp', '.html', '.css', '.json', '.xml', '.yaml', '.yml', '.ts', '.tsx', '.jsx', '.php', '.rb', '.rs', '.swift', '.kt']
    };
    
    for (const [type, extensions] of Object.entries(typeMap)) {
        if (extensions.includes(ext)) {
            return type;
        }
    }
    
    return 'other';
}

/**
 * Get icon for file type
 * @param {string} fileType - File type from getFileType()
 * @returns {string} Icon character/emoji
 */
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

/**
 * Parse S3 key to get folder path and filename
 * @param {string} key - S3 object key
 * @returns {Object} {folder, filename, isFolder}
 */
function parseS3Key(key) {
    if (!key) return { folder: '', filename: '', isFolder: false };
    
    const isFolder = key.endsWith('/');
    const parts = key.split('/');
    
    if (isFolder) {
        parts.pop(); // Remove empty string at end
        const filename = parts.pop() + '/';
        const folder = parts.join('/') + (parts.length > 0 ? '/' : '');
        return { folder, filename, isFolder: true };
    } else {
        const filename = parts.pop();
        const folder = parts.join('/') + (parts.length > 0 ? '/' : '');
        return { folder, filename, isFolder: false };
    }
}

/**
 * Generate UUID v4
 * @returns {string} UUID string
 */
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

/**
 * Deep clone an object
 * @param {Object} obj - Object to clone
 * @returns {Object} Cloned object
 */
function deepClone(obj) {
    return JSON.parse(JSON.stringify(obj));
}

/**
 * Check if a string matches a filter (case-insensitive)
 * @param {string} text - Text to check
 * @param {string} filter - Filter string
 * @returns {boolean} Whether text matches filter
 */
function matchesFilter(text, filter) {
    if (!filter) return true;
    if (!text) return false;
    
    return text.toLowerCase().includes(filter.toLowerCase());
}

/**
 * Truncate text with ellipsis
 * @param {string} text - Text to truncate
 * @param {number} maxLength - Maximum length
 * @returns {string} Truncated text
 */
function truncate(text, maxLength = 50) {
    if (!text || text.length <= maxLength) return text;
    return text.substring(0, maxLength - 3) + '...';
}

/**
 * Validate S3 bucket name
 * @param {string} name - Bucket name to validate
 * @returns {Object} {valid: boolean, error: string|null}
 */
function validateBucketName(name) {
    if (!name) {
        return { valid: false, error: 'Bucket name is required' };
    }
    
    if (name.length < 3 || name.length > 63) {
        return { valid: false, error: 'Bucket name must be between 3 and 63 characters' };
    }
    
    if (!/^[a-z0-9.-]+$/.test(name)) {
        return { valid: false, error: 'Bucket name can only contain lowercase letters, numbers, dots, and hyphens' };
    }
    
    if (/^[.-]|[.-]$/.test(name)) {
        return { valid: false, error: 'Bucket name cannot start or end with a dot or hyphen' };
    }
    
    if (/\.\./.test(name)) {
        return { valid: false, error: 'Bucket name cannot contain consecutive dots' };
    }
    
    if (/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(name)) {
        return { valid: false, error: 'Bucket name cannot be formatted as an IP address' };
    }
    
    return { valid: true, error: null };
}

/**
 * Validate S3 key (folder or file path)
 * @param {string} key - Key to validate
 * @returns {Object} {valid: boolean, error: string|null}
 */
function validateS3Key(key) {
    if (!key) {
        return { valid: false, error: 'Key is required' };
    }
    
    if (key.length > 1024) {
        return { valid: false, error: 'Key must be less than 1024 characters' };
    }
    
    // Check for invalid characters (control characters)
    if (/[\x00-\x1f\x7f]/.test(key)) {
        return { valid: false, error: 'Key contains invalid characters' };
    }
    
    return { valid: true, error: null };
}

/**
 * Sanitize filename for display
 * @param {string} filename - Filename to sanitize
 * @returns {string} Sanitized filename
 */
function sanitizeFilename(filename) {
    if (!filename) return '';
    
    // Remove control characters and path traversal attempts
    return filename.replace(/[\x00-\x1f\x7f]/g, '').replace(/\.\./g, '');
}

/**
 * Format upload speed
 * @param {number} bytesPerSecond - Speed in bytes/sec
 * @returns {string} Formatted speed
 */
function formatSpeed(bytesPerSecond) {
    return formatSize(bytesPerSecond) + '/s';
}

/**
 * Calculate ETA for upload
 * @param {number} bytesUploaded - Bytes uploaded so far
 * @param {number} bytesTotal - Total bytes
 * @param {number} speedBps - Current speed in bytes/sec
 * @returns {string} Formatted ETA
 */
function formatETA(bytesUploaded, bytesTotal, speedBps) {
    if (!speedBps || speedBps <= 0) return 'Calculating...';
    
    const remaining = bytesTotal - bytesUploaded;
    const seconds = Math.ceil(remaining / speedBps);
    
    if (seconds < 60) {
        return seconds + 's';
    } else if (seconds < 3600) {
        return Math.floor(seconds / 60) + 'm ' + (seconds % 60) + 's';
    } else {
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        return hours + 'h ' + minutes + 'm';
    }
}

/**
 * Add timestamp suffix to filename before extension
 * @param {string} filename - Original filename
 * @returns {string} Filename with timestamp
 */
function addTimestampToFilename(filename) {
    if (!filename) return '';
    
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
    const lastDot = filename.lastIndexOf('.');
    
    if (lastDot === -1) {
        return filename + '_' + timestamp;
    }
    
    const name = filename.substring(0, lastDot);
    const ext = filename.substring(lastDot);
    return name + '_' + timestamp + ext;
}

/**
 * Copy text to clipboard
 * @param {string} text - Text to copy
 * @returns {Promise<boolean>} Success status
 */
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        console.error('Failed to copy:', err);
        return false;
    }
}

/**
 * Download blob as file
 * @param {Blob} blob - Blob to download
 * @param {string} filename - Filename for download
 */
function downloadBlob(blob, filename) {
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
}

/**
 * Throttle function to limit execution rate
 * @param {Function} func - Function to throttle
 * @param {number} limit - Time limit in milliseconds
 * @returns {Function} Throttled function
 */
function throttle(func, limit) {
    let inThrottle;
    return function(...args) {
        if (!inThrottle) {
            func.apply(this, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    };
}

/**
 * Wait for specified duration
 * @param {number} ms - Milliseconds to wait
 * @returns {Promise<void>}
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Retry a function with exponential backoff
 * @param {Function} fn - Function to retry
 * @param {number} maxRetries - Maximum retry attempts
 * @param {number} delay - Initial delay in ms
 * @returns {Promise<any>} Function result
 */
async function retryWithBackoff(fn, maxRetries = 3, delay = 1000) {
    let lastError;
    
    for (let i = 0; i <= maxRetries; i++) {
        try {
            return await fn();
        } catch (error) {
            lastError = error;
            if (i === maxRetries) break;
            
            const waitTime = delay * Math.pow(2, i);
            await sleep(waitTime);
        }
    }
    
    throw lastError;
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        debounce,
        formatSize,
        formatDate,
        escapeHtml,
        getFileExtension,
        getFileType,
        getFileIcon,
        parseS3Key,
        generateUUID,
        deepClone,
        matchesFilter,
        truncate,
        validateBucketName,
        validateS3Key,
        sanitizeFilename,
        formatSpeed,
        formatETA,
        addTimestampToFilename,
        copyToClipboard,
        downloadBlob,
        throttle,
        sleep,
        retryWithBackoff
    };
}
