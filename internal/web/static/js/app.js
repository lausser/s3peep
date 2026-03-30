// S3 File Browser - Main Application
// Entry point and coordinator for all components

(function() {
    'use strict';

    // DOM element references
    const elements = {};

    // Initialize the application
    async function init() {
        console.log('S3 File Browser initializing...');

        // Cache DOM elements
        cacheElements();

        // Initialize authentication
        if (window.Auth && !window.Auth.init()) {
            // Auth failed - stop initialization
            return;
        }

        // Initialize state
        if (window.State) {
            window.State.init && window.State.init();
        }

        // Initialize UI components
        initComponents();
        
        // Initialize file browsing components (will be inactive until bucket selected)
        initFileBrowsingComponents();

        // Load profile info
        await loadProfile();

        // Check for default bucket and auto-navigate if present
        await checkDefaultBucket();

        console.log('S3 File Browser initialized');
    }

    // Cache DOM element references
    function cacheElements() {
        elements.bucketView = document.getElementById('bucket-view');
        elements.fileView = document.getElementById('file-view');
        elements.profileName = document.getElementById('profile-name');
        elements.themeToggle = document.getElementById('theme-toggle');
        elements.btnBack = document.getElementById('btn-back');
        elements.loadingOverlay = document.getElementById('loading-overlay');
    }

    // Initialize UI components
    function initComponents() {
        // Initialize error component first (for error reporting)
        if (window.ErrorComponent && window.ErrorComponent.init) {
            window.ErrorComponent.init();
        }

        // Initialize empty state
        if (window.EmptyState && window.EmptyState.init) {
            window.EmptyState.init();
        }

        // Initialize bucket filter
        if (window.BucketFilter && window.BucketFilter.init) {
            window.BucketFilter.init();
        }

        // Initialize bucket list
        if (window.BucketList && window.BucketList.init) {
            window.BucketList.init();
        }

        // Initialize theme toggle
        if (elements.themeToggle) {
            elements.themeToggle.addEventListener('click', () => {
                if (window.State) {
                    window.State.toggleTheme();
                    updateThemeIcon();
                }
            });
        }

        // Initialize back button
        if (elements.btnBack) {
            elements.btnBack.addEventListener('click', () => {
                goBackToBuckets();
            });
        }

        // Initialize upload button
        const btnUpload = document.getElementById('btn-upload');
        const fileInput = document.getElementById('file-input');
        if (btnUpload && fileInput) {
            btnUpload.addEventListener('click', () => {
                fileInput.click();
            });

            fileInput.addEventListener('change', (e) => {
                const files = e.target.files;
                if (files.length > 0) {
                    const bucket = window.State ? window.State.get('selectedBucket') : null;
                    const path = window.State ? window.State.get('currentPath') : '';
                    
                    if (bucket && window.UploadZone) {
                        window.UploadZone.handleFiles(files, bucket, path);
                    }
                    
                    // Reset input
                    fileInput.value = '';
                }
            });
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('view:changed', (view) => {
                switchView(view);
            });

            window.State.subscribe('theme:changed', () => {
                updateThemeIcon();
            });
        }
    }

    // Initialize file browsing components
    function initFileBrowsingComponents() {
        // Breadcrumb
        if (window.Breadcrumb && window.Breadcrumb.init) {
            window.Breadcrumb.init();
        }

        // File list
        if (window.FileList && window.FileList.init) {
            window.FileList.init();
        }

        // File filter
        if (window.FileFilter && window.FileFilter.init) {
            window.FileFilter.init();
        }

        // Pagination
        if (window.Pagination && window.Pagination.init) {
            window.Pagination.init();
        }

        // Upload components
        if (window.UploadZone && window.UploadZone.init) {
            window.UploadZone.init();
        }

        if (window.UploadProgress && window.UploadProgress.init) {
            window.UploadProgress.init();
        }

        if (window.ConflictModal && window.ConflictModal.init) {
            window.ConflictModal.init();
        }

        // Set initial theme
        updateThemeIcon();
    }

    // Load profile information
    async function loadProfile() {
        try {
            const profile = await window.API.getProfile();
            
            if (window.State) {
                window.State.set('profile', profile);
            }

            // Update UI
            if (elements.profileName && profile.name) {
                elements.profileName.textContent = profile.name;
            }

            return profile;
        } catch (err) {
            console.error('Failed to load profile:', err);
            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
        }
    }

    // Check for default bucket and auto-navigate
    async function checkDefaultBucket() {
        const profile = window.State ? window.State.get('profile') : null;
        
        if (profile && profile.bucket) {
            console.log(`Default bucket configured: ${profile.bucket}`);
            
            // Set the bucket filter to show only this bucket
            if (window.BucketFilter) {
                window.BucketFilter.setValue(profile.bucket);
            }
            
            // Wait a moment for the bucket list to load, then auto-select
            setTimeout(async () => {
                const buckets = window.BucketList ? window.BucketList.getBuckets() : [];
                const bucketExists = buckets.some(b => b.name === profile.bucket);
                
                if (bucketExists) {
                    console.log(`Auto-navigating to default bucket: ${profile.bucket}`);
                    if (window.BucketList) {
                        await window.BucketList.selectBucket(profile.bucket);
                    }
                } else {
                    console.warn(`Default bucket '${profile.bucket}' not found or not accessible`);
                    showToast(`Default bucket '${profile.bucket}' not found`, 'warning');
                }
            }, 500);
        }
    }

    // Switch between views
    function switchView(viewName) {
        if (!elements.bucketView || !elements.fileView) return;

        if (viewName === 'buckets') {
            elements.bucketView.classList.remove('hidden');
            elements.fileView.classList.add('hidden');
            document.title = 'S3 File Browser - Select Bucket';
        } else if (viewName === 'files') {
            elements.bucketView.classList.add('hidden');
            elements.fileView.classList.remove('hidden');
            const bucket = window.State ? window.State.get('selectedBucket') : '';
            document.title = bucket ? `S3 File Browser - ${bucket}` : 'S3 File Browser';
        }

        // Emit view changed event
        document.dispatchEvent(new CustomEvent('app:viewChanged', { detail: { view: viewName } }));
    }

    // Go back to bucket list
    function goBackToBuckets() {
        if (window.State) {
            window.State.goBackToBuckets();
        }
        switchView('buckets');
    }

    // Update theme icon based on current theme
    function updateThemeIcon() {
        if (!elements.themeToggle) return;
        
        const theme = window.State ? window.State.get('theme') : 'light';
        const icon = elements.themeToggle.querySelector('.icon');
        
        if (icon) {
            icon.textContent = theme === 'dark' ? '☀️' : '🌙';
        }
        
        elements.themeToggle.setAttribute('title', 
            theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'
        );
    }

    // Show loading overlay
    function showLoading(message = 'Loading...') {
        if (elements.loadingOverlay) {
            const textEl = elements.loadingOverlay.querySelector('.loading-text');
            if (textEl) textEl.textContent = message;
            elements.loadingOverlay.classList.remove('hidden');
        }
    }

    // Hide loading overlay
    function hideLoading() {
        if (elements.loadingOverlay) {
            elements.loadingOverlay.classList.add('hidden');
        }
    }

    // Show toast notification
    function showToast(message, type = 'info', duration = 5000) {
        if (window.ErrorComponent && window.ErrorComponent[type]) {
            window.ErrorComponent[type](message, duration);
        }
    }

    // Export public API
    window.App = {
        init,
        showLoading,
        hideLoading,
        showToast,
        goBackToBuckets,
        switchView
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
