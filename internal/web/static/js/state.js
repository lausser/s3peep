// S3 File Browser - State Management
// Simple event-driven state management without external libraries

(function() {
    'use strict';

    // Application state
    const state = {
        // Authentication
        token: null,
        tokenExpiresAt: null,
        
        // Navigation
        currentView: 'buckets', // 'buckets' | 'files'
        selectedBucket: null,
        currentPath: '',
        
        // Buckets
        buckets: [],
        bucketFilter: '',
        filteredBuckets: [],
        
        // Files
        files: [],
        fileFilter: '',
        filteredFiles: [],
        
        // Pagination
        pagination: {
            currentPage: 1,
            pageSize: 100,
            isTruncated: false,
            continuationTokens: {}, // page number -> token
            totalItems: null
        },
        
        // Uploads
        uploads: [], // UploadTask[]
        
        // Selection
        selectedFiles: [], // file keys
        
        // UI State
        theme: 'light',
        isLoading: false,
        profile: null
    };

    // Event listeners
    const listeners = {};

    // Subscribe to state changes
    function subscribe(event, callback) {
        if (!listeners[event]) {
            listeners[event] = [];
        }
        listeners[event].push(callback);
        
        // Return unsubscribe function
        return () => {
            const index = listeners[event].indexOf(callback);
            if (index > -1) {
                listeners[event].splice(index, 1);
            }
        };
    }

    // Emit event to subscribers
    function emit(event, data) {
        if (listeners[event]) {
            listeners[event].forEach(callback => {
                try {
                    callback(data, state);
                } catch (err) {
                    console.error('Error in state listener:', err);
                }
            });
        }
    }

    // Get current state (shallow copy)
    function getState() {
        return { ...state };
    }

    // Get specific state property
    function get(key) {
        return state[key];
    }

    // Set state property
    function set(key, value) {
        const oldValue = state[key];
        state[key] = value;
        emit('state:changed', { key, value, oldValue });
        emit(`state:${key}:changed`, value);
    }

    // Update nested state
    function update(key, updates) {
        if (typeof state[key] === 'object' && state[key] !== null) {
            const oldValue = { ...state[key] };
            state[key] = { ...state[key], ...updates };
            emit('state:changed', { key, value: state[key], oldValue });
            emit(`state:${key}:changed`, state[key]);
        } else {
            set(key, updates);
        }
    }

    // Reset state to initial values
    function reset() {
        state.currentView = 'buckets';
        state.selectedBucket = null;
        state.currentPath = '';
        state.buckets = [];
        state.bucketFilter = '';
        state.filteredBuckets = [];
        state.files = [];
        state.fileFilter = '';
        state.filteredFiles = [];
        state.pagination = {
            currentPage: 1,
            pageSize: 100,
            isTruncated: false,
            continuationTokens: {},
            totalItems: null
        };
        state.selectedFiles = [];
        state.isLoading = false;
        emit('state:reset');
    }

    // Persist theme to sessionStorage
    function persistTheme() {
        try {
            sessionStorage.setItem('s3peep_theme', state.theme);
        } catch (e) {
            // Ignore storage errors
        }
    }

    // Load theme from sessionStorage
    function loadTheme() {
        try {
            const saved = sessionStorage.getItem('s3peep_theme');
            if (saved) {
                state.theme = saved;
            }
        } catch (e) {
            // Ignore storage errors
        }
    }

    // Filter buckets based on filter text
    function filterBuckets() {
        const filter = state.bucketFilter.toLowerCase();
        if (!filter) {
            state.filteredBuckets = [...state.buckets];
        } else {
            state.filteredBuckets = state.buckets.filter(bucket => 
                bucket.name.toLowerCase().includes(filter)
            );
        }
        emit('buckets:filtered', state.filteredBuckets);
    }

    // Filter files based on filter text (client-side, current page only)
    function filterFiles() {
        const filter = state.fileFilter.toLowerCase();
        if (!filter) {
            state.filteredFiles = [...state.files];
        } else {
            state.filteredFiles = state.files.filter(file => 
                file.name.toLowerCase().includes(filter)
            );
        }
        emit('files:filtered', state.filteredFiles);
    }

    // Set bucket filter
    function setBucketFilter(filter) {
        state.bucketFilter = filter;
        filterBuckets();
        emit('bucketFilter:changed', filter);
    }

    // Set file filter
    function setFileFilter(filter) {
        state.fileFilter = filter;
        filterFiles();
        emit('fileFilter:changed', filter);
    }

    // Select bucket
    function selectBucket(bucket) {
        state.selectedBucket = bucket;
        state.currentPath = '';
        state.currentView = 'files';
        state.files = [];
        state.fileFilter = '';
        state.filteredFiles = [];
        state.pagination = {
            currentPage: 1,
            pageSize: 100,
            isTruncated: false,
            continuationTokens: {},
            totalItems: null
        };
        state.selectedFiles = [];
        emit('bucket:selected', bucket);
    }

    // Navigate to folder
    function navigateToFolder(folderPath) {
        state.currentPath = folderPath;
        state.files = [];
        state.fileFilter = '';
        state.filteredFiles = [];
        state.pagination = {
            currentPage: 1,
            pageSize: 100,
            isTruncated: false,
            continuationTokens: {},
            totalItems: null
        };
        state.selectedFiles = [];
        emit('folder:navigated', folderPath);
    }

    // Go back to bucket list
    function goBackToBuckets() {
        state.currentView = 'buckets';
        state.selectedBucket = null;
        state.currentPath = '';
        state.files = [];
        state.fileFilter = '';
        state.filteredFiles = [];
        state.selectedFiles = [];
        emit('view:changed', 'buckets');
    }

    // Set page size
    function setPageSize(size) {
        state.pagination.pageSize = size;
        state.pagination.currentPage = 1;
        state.pagination.continuationTokens = {};
        emit('pagination:changed', state.pagination);
    }

    // Set current page
    function setPage(page) {
        state.pagination.currentPage = page;
        emit('pagination:changed', state.pagination);
    }

    // Store continuation token for a page
    function storeContinuationToken(page, token) {
        state.pagination.continuationTokens[page] = token;
    }

    // Get continuation token for a page
    function getContinuationToken(page) {
        return state.pagination.continuationTokens[page] || '';
    }

    // Toggle file selection
    function toggleFileSelection(key) {
        const index = state.selectedFiles.indexOf(key);
        if (index > -1) {
            state.selectedFiles.splice(index, 1);
        } else {
            state.selectedFiles.push(key);
        }
        emit('selection:changed', [...state.selectedFiles]);
    }

    // Select all files
    function selectAllFiles() {
        state.selectedFiles = state.filteredFiles.map(f => f.key);
        emit('selection:changed', [...state.selectedFiles]);
    }

    // Deselect all files
    function deselectAllFiles() {
        state.selectedFiles = [];
        emit('selection:changed', []);
    }

    // Add upload task
    function addUpload(upload) {
        state.uploads.push(upload);
        emit('upload:added', upload);
    }

    // Update upload
    function updateUpload(uploadId, updates) {
        const upload = state.uploads.find(u => u.id === uploadId);
        if (upload) {
            Object.assign(upload, updates);
            emit('upload:updated', upload);
        }
    }

    // Remove upload
    function removeUpload(uploadId) {
        const index = state.uploads.findIndex(u => u.id === uploadId);
        if (index > -1) {
            const upload = state.uploads.splice(index, 1)[0];
            emit('upload:removed', upload);
        }
    }

    // Toggle theme
    function toggleTheme() {
        state.theme = state.theme === 'light' ? 'dark' : 'light';
        persistTheme();
        emit('theme:changed', state.theme);
    }

    // Set theme
    function setTheme(theme) {
        state.theme = theme;
        persistTheme();
        emit('theme:changed', theme);
    }

    // Initialize
    loadTheme();

    // Export
    window.State = {
        get: get,
        getState: getState,
        set: set,
        update: update,
        reset: reset,
        subscribe: subscribe,
        emit: emit,
        filterBuckets: filterBuckets,
        filterFiles: filterFiles,
        setBucketFilter: setBucketFilter,
        setFileFilter: setFileFilter,
        selectBucket: selectBucket,
        navigateToFolder: navigateToFolder,
        goBackToBuckets: goBackToBuckets,
        setPageSize: setPageSize,
        setPage: setPage,
        storeContinuationToken: storeContinuationToken,
        getContinuationToken: getContinuationToken,
        toggleFileSelection: toggleFileSelection,
        selectAllFiles: selectAllFiles,
        deselectAllFiles: deselectAllFiles,
        addUpload: addUpload,
        updateUpload: updateUpload,
        removeUpload: removeUpload,
        toggleTheme: toggleTheme,
        setTheme: setTheme
    };
})();
