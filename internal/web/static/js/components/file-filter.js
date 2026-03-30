// S3 File Browser - File Filter Component
// Real-time filtering of files by name

(function() {
    'use strict';

    let input = null;
    let debouncedFilter = null;
    const DEBOUNCE_MS = 300;

    // Initialize the component
    function init() {
        input = document.getElementById('file-filter');
        if (!input) {
            console.error('File filter input not found');
            return;
        }

        // Create debounced filter function
        debouncedFilter = debounce((value) => {
            if (window.State) {
                window.State.setFileFilter(value);
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

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('fileFilter:changed', (value) => {
                if (input.value !== value) {
                    input.value = value;
                }
            });

            // Show filter when bucket is selected
            window.State.subscribe('view:changed', (view) => {
                if (view === 'files') {
                    input.disabled = false;
                } else {
                    input.disabled = true;
                    clear();
                }
            });
        }

        // Handle clear button
        const clearBtn = document.getElementById('clear-file-filter');
        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                clear();
                focus();
            });
        }

        // Initially disable until bucket selected
        input.disabled = true;
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
        if (input && !input.disabled) {
            input.focus();
            input.select();
        }
    }

    // Clear the filter
    function clear() {
        if (input) {
            input.value = '';
            if (window.State) {
                window.State.setFileFilter('');
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
                window.State.setFileFilter(value);
            }
        }
    }

    // Enable/disable input
    function setEnabled(enabled) {
        if (input) {
            input.disabled = !enabled;
        }
    }

    // Export
    window.FileFilter = {
        init,
        focus,
        clear,
        getValue,
        setValue,
        setEnabled
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
