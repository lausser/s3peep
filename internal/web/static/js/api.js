// S3 File Browser - API Client
// Handles all HTTP communication with the backend

(function() {
    'use strict';

    // Get the token from the URL path
    function getTokenFromUrl() {
        const path = window.location.pathname;
        const parts = path.split('/').filter(p => p);
        return parts[0] || '';
    }

    // Get the base API URL
    function getBaseUrl() {
        const token = getTokenFromUrl();
        if (!token) {
            console.error('No token found in URL');
            return null;
        }
        return `/${token}/api`;
    }

    // Make an authenticated API request
    async function request(method, endpoint, options = {}) {
        const baseUrl = getBaseUrl();
        if (!baseUrl) {
            throw new Error('Not authenticated - no token in URL');
        }

        const url = `${baseUrl}${endpoint}`;
        
        const config = {
            method: method.toUpperCase(),
            headers: {
                'Accept': 'application/json',
                ...options.headers
            },
            ...options
        };

        // Add body for non-GET requests
        if (options.body && method.toUpperCase() !== 'GET') {
            if (typeof options.body === 'object' && !(options.body instanceof FormData)) {
                config.body = JSON.stringify(options.body);
                config.headers['Content-Type'] = 'application/json';
            } else {
                config.body = options.body;
            }
        }

        try {
            const response = await fetch(url, config);
            
            // Handle non-OK responses
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({
                    error: `HTTP ${response.status}: ${response.statusText}`,
                    code: `HTTP_${response.status}`
                }));
                
                const error = new Error(errorData.error || `Request failed: ${response.statusText}`);
                error.code = errorData.code || `HTTP_${response.status}`;
                error.status = response.status;
                error.details = errorData.details;
                throw error;
            }

            // Return JSON for JSON responses, text for others
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            
            return await response.text();
            
        } catch (err) {
            // Network errors or other fetch failures
            if (err.name === 'TypeError' && err.message.includes('fetch')) {
                err.code = 'NETWORK_ERROR';
                err.message = 'Network error - please check your connection';
            }
            throw err;
        }
    }

    // Handle errors consistently
    function handleError(err) {
        console.error('API Error:', err);
        
        // Show toast notification if Error component is available
        if (window.ErrorComponent && window.ErrorComponent.show) {
            window.ErrorComponent.show(err.message, 'error');
        }
        
        // Redirect to error page for auth errors
        if (err.code === 'INVALID_TOKEN' || err.code === 'HTTP_403') {
            sessionStorage.removeItem('s3peep_token');
            window.ErrorComponent && window.ErrorComponent.show(
                'Session expired. Please restart s3peep server.', 
                'error',
                10000
            );
        }
        
        throw err;
    }

    // Retry with exponential backoff
    async function retryWithBackoff(fn, maxRetries = 3, delay = 1000) {
        let lastError;
        
        for (let i = 0; i <= maxRetries; i++) {
            try {
                return await fn();
            } catch (error) {
                lastError = error;
                if (i === maxRetries) break;
                
                // Only retry on network errors or 5xx errors
                if (error.code === 'NETWORK_ERROR' || (error.status && error.status >= 500)) {
                    await new Promise(resolve => setTimeout(resolve, delay * Math.pow(2, i)));
                } else {
                    throw error;
                }
            }
        }
        
        throw lastError;
    }

    // API endpoints
    const API = {
        // Token
        getToken: getTokenFromUrl,
        getBaseUrl: getBaseUrl,

        // Buckets
        async listBuckets() {
            return retryWithBackoff(() => request('GET', '/buckets'));
        },

        async selectBucket(bucketName) {
            return retryWithBackoff(() => request('POST', '/buckets/select', {
                body: { bucket: bucketName }
            }));
        },

        // Objects
        async listObjects(bucket, prefix = '', continuationToken = '', maxKeys = 100) {
            const params = new URLSearchParams();
            if (prefix) params.append('prefix', prefix);
            if (continuationToken) params.append('continuation_token', continuationToken);
            params.append('max_keys', maxKeys.toString());
            
            return retryWithBackoff(() => 
                request('GET', `/buckets/${encodeURIComponent(bucket)}/objects?${params}`)
            );
        },

        async downloadFile(bucket, key) {
            const params = new URLSearchParams({ key });
            const url = `${getBaseUrl()}/buckets/${encodeURIComponent(bucket)}/download?${params}`;
            
            const response = await fetch(url);
            if (!response.ok) {
                throw new Error(`Download failed: ${response.statusText}`);
            }
            
            return response.blob();
        },

        async uploadFile(bucket, key, file, overwrite = false) {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('key', key);
            formData.append('overwrite', overwrite.toString());
            
            return request('POST', `/buckets/${encodeURIComponent(bucket)}/upload`, {
                body: formData
            });
        },

        async createFolder(bucket, folderPath) {
            return request('PUT', `/buckets/${encodeURIComponent(bucket)}/folders`, {
                body: { folder_path: folderPath }
            });
        },

        async deleteObjects(bucket, keys) {
            return request('DELETE', `/buckets/${encodeURIComponent(bucket)}/objects`, {
                body: { keys }
            });
        },

        // Profile
        async getProfile() {
            return retryWithBackoff(() => request('GET', '/profile'));
        },

        // Head object (check if exists, get metadata)
        async headObject(bucket, key) {
            // HEAD requests don't work well with CORS/fetch in some cases
            // Use a GET with a special header or just try to get metadata via list
            // For now, return a dummy implementation
            return retryWithBackoff(() => 
                request('GET', `/buckets/${encodeURIComponent(bucket)}/objects?prefix=${encodeURIComponent(key)}&max_keys=1`)
            );
        }
    };

    // Export to global scope
    window.API = API;
    window.handleApiError = handleError;
})();
