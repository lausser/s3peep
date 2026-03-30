// S3 File Browser - Bucket List Component
// Displays and manages the list of S3 buckets

(function() {
    'use strict';

    let container = null;
    let buckets = [];
    let filteredBuckets = [];
    let isLoading = false;

    // Initialize the component
    function init() {
        container = document.getElementById('bucket-list');
        if (!container) {
            console.error('Bucket list container not found');
            return;
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('buckets:loaded', (data) => {
                buckets = data.buckets || [];
                render();
            });

            window.State.subscribe('bucketFilter:changed', () => {
                filterBuckets();
            });
        }

        // Initial load
        loadBuckets();
    }

    // Load buckets from API
    async function loadBuckets() {
        if (isLoading) return;
        
        isLoading = true;
        showLoading();

        try {
            const result = await window.API.listBuckets();
            buckets = result.buckets || [];
            
            // Update state
            if (window.State) {
                window.State.set('buckets', buckets);
                window.State.filterBuckets();
            }
            
            // Get filtered buckets
            filteredBuckets = window.State ? window.State.get('filteredBuckets') || buckets : buckets;
            
            render();
            
            // Emit event
            document.dispatchEvent(new CustomEvent('buckets:loaded', { detail: { buckets } }));
        } catch (err) {
            console.error('Failed to load buckets:', err);
            showError('Failed to load buckets. Please try again.');
            
            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
        } finally {
            isLoading = false;
            hideLoading();
        }
    }

    // Filter buckets based on current filter text
    function filterBuckets() {
        if (!window.State) return;
        
        const filterText = window.State.get('bucketFilter') || '';
        
        if (!filterText) {
            filteredBuckets = [...buckets];
        } else {
            const lowerFilter = filterText.toLowerCase();
            filteredBuckets = buckets.filter(bucket => 
                bucket.name.toLowerCase().includes(lowerFilter)
            );
        }
        
        window.State.set('filteredBuckets', filteredBuckets);
        render();
    }

    // Render the bucket list
    function render() {
        if (!container) return;

        if (isLoading) {
            container.innerHTML = '';
            return;
        }

        if (filteredBuckets.length === 0) {
            container.innerHTML = '';
            // Show empty state - the empty-state component handles this
            document.dispatchEvent(new CustomEvent('buckets:empty', { 
                detail: { hasFilter: window.State && window.State.get('bucketFilter') } 
            }));
            return;
        }

        // Hide empty state
        document.dispatchEvent(new CustomEvent('buckets:not-empty'));

        const html = filteredBuckets.map(bucket => createBucketHtml(bucket)).join('');
        container.innerHTML = html;

        // Add click handlers
        container.querySelectorAll('.bucket-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const bucketName = item.dataset.bucketName;
                selectBucket(bucketName);
            });
        });
    }

    // Create HTML for a single bucket
    function createBucketHtml(bucket) {
        const icon = '🪣';
        const name = escapeHtml(bucket.name);
        const region = bucket.region ? escapeHtml(bucket.region) : 'Unknown region';
        
        return `
            <div class="bucket-item" data-bucket-name="${name}" role="button" tabindex="0" aria-label="Bucket ${name}">
                <div class="bucket-icon">${icon}</div>
                <div class="bucket-info">
                    <div class="bucket-name">${name}</div>
                    <div class="bucket-meta">${region}</div>
                </div>
                <div class="bucket-arrow">→</div>
            </div>
        `;
    }

    // Select a bucket
    async function selectBucket(bucketName) {
        if (!bucketName) return;

        // Show loading indicator on the item
        const item = container.querySelector(`[data-bucket-name="${bucketName}"]`);
        if (item) {
            item.classList.add('loading');
        }

        try {
            // Select bucket via API
            await window.API.selectBucket(bucketName);
            
            // Update state
            if (window.State) {
                window.State.selectBucket(bucketName);
            }
            
            // Emit event
            document.dispatchEvent(new CustomEvent('bucket:selected', { 
                detail: { bucket: bucketName } 
            }));
            
        } catch (err) {
            console.error('Failed to select bucket:', err);
            if (window.ErrorComponent) {
                window.ErrorComponent.handleApiError(err);
            }
            
            if (item) {
                item.classList.remove('loading');
            }
        }
    }

    // Show loading state
    function showLoading() {
        const skeleton = document.getElementById('bucket-skeleton');
        if (skeleton) {
            skeleton.classList.remove('hidden');
        }
        if (container) {
            container.classList.add('hidden');
        }
    }

    // Hide loading state
    function hideLoading() {
        const skeleton = document.getElementById('bucket-skeleton');
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
                <div class="bucket-error">
                    <div class="error-icon">⚠️</div>
                    <p>${escapeHtml(message)}</p>
                    <button class="btn btn-secondary" onclick="BucketList.refresh()">Retry</button>
                </div>
            `;
        }
    }

    // Escape HTML
    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Refresh bucket list
    function refresh() {
        loadBuckets();
    }

    // Get current buckets
    function getBuckets() {
        return buckets;
    }

    // Get filtered buckets
    function getFilteredBuckets() {
        return filteredBuckets;
    }

    // Export
    window.BucketList = {
        init,
        refresh,
        getBuckets,
        getFilteredBuckets,
        selectBucket
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
