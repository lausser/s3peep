// S3 File Browser - Bucket Filter Component
// Real-time filtering of bucket list

(function() {
    'use strict';

    let input = null;
    let debouncedFilter = null;
    const DEBOUNCE_MS = 300;

    // Initialize the component
    function init() {
        input = document.getElementById('bucket-filter');
        if (!input) {
            console.error('Bucket filter input not found');
            return;
        }

        // Create debounced filter function
        debouncedFilter = debounce((value) => {
            if (window.State) {
                window.State.setBucketFilter(value);
            }
        }, DEBOUNCE_MS);

        // Add input event listener
        input.addEventListener('input', (e) => {
            const value = e.target.value.trim();
            debouncedFilter(value);
        });

        // Handle keyboard shortcuts
        input.addEventListener('keydown', (e) => {
            // Clear on Escape
            if (e.key === 'Escape') {
                e.preventDefault();
                clear();
                input.blur();
            }
        });

        // Global keyboard shortcut (/) to focus filter
        document.addEventListener('keydown', (e) => {
            // Ignore if user is typing in another input
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
                return;
            }
            
            // Ignore if modal is open
            if (document.querySelector('.modal:not(.hidden)')) {
                return;
            }

            // Focus on /
            if (e.key === '/') {
                e.preventDefault();
                focus();
            }
        });

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('bucketFilter:changed', (value) => {
                if (input.value !== value) {
                    input.value = value;
                }
            });
        }

        // Handle clear button
        const clearBtn = document.getElementById('clear-bucket-filter');
        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                clear();
                focus();
            });
        }
    }

    // Debounce helper
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

    // Focus the filter input
    function focus() {
        if (input) {
            input.focus();
            input.select();
        }
    }

    // Clear the filter
    function clear() {
        if (input) {
            input.value = '';
            if (window.State) {
                window.State.setBucketFilter('');
            }
        }
    }

    // Get current filter value
    function getValue() {
        return input ? input.value : '';
    }

    // Set filter value programmatically
    function setValue(value) {
        if (input) {
            input.value = value;
            if (window.State) {
                window.State.setBucketFilter(value);
            }
        }
    }

    // Export
    window.BucketFilter = {
        init,
        focus,
        clear,
        getValue,
        setValue
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
