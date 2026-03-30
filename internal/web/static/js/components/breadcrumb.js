// S3 File Browser - Breadcrumb Component
// Navigate folder hierarchy

(function() {
    'use strict';

    let container = null;

    // Initialize the component
    function init() {
        container = document.getElementById('breadcrumb');
        if (!container) {
            console.error('Breadcrumb container not found');
            return;
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('folder:navigated', (path) => {
                render(path);
            });

            window.State.subscribe('bucket:selected', () => {
                render('');
            });
        }

        // Add click handler
        container.addEventListener('click', handleBreadcrumbClick);
    }

    // Handle breadcrumb click
    function handleBreadcrumbClick(e) {
        const link = e.target.closest('[data-path]');
        if (!link) return;

        e.preventDefault();

        const path = link.dataset.path;
        navigateToPath(path);
    }

    // Navigate to a specific path
    function navigateToPath(path) {
        if (window.State) {
            window.State.navigateToFolder(path);
        }
    }

    // Render breadcrumb
    function render(currentPath) {
        if (!container) return;

        const bucket = window.State ? window.State.get('selectedBucket') : '';
        if (!bucket) {
            container.innerHTML = '';
            return;
        }

        const parts = currentPath.split('/').filter(p => p);
        const items = [];

        // Home link
        items.push({
            name: bucket,
            path: '',
            isCurrent: parts.length === 0
        });

        // Path segments
        let accumulatedPath = '';
        parts.forEach((part, index) => {
            accumulatedPath += part + '/';
            items.push({
                name: part,
                path: accumulatedPath,
                isCurrent: index === parts.length - 1
            });
        });

        // Build HTML
        const html = items.map((item, index) => {
            if (item.isCurrent) {
                return `<span class="breadcrumb-current">${escapeHtml(item.name)}</span>`;
            }

            return `
                <a href="#" data-path="${escapeHtml(item.path)}" class="breadcrumb-link">
                    ${escapeHtml(item.name)}
                </a>
                <span class="breadcrumb-separator">/</span>
            `;
        }).join('');

        container.innerHTML = html;
    }

    // Escape HTML
    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Get current path from breadcrumb
    function getCurrentPath() {
        const current = container ? container.querySelector('.breadcrumb-current') : null;
        if (!current) return '';

        // Find all links and build path
        const links = container.querySelectorAll('.breadcrumb-link');
        let path = '';
        links.forEach(link => {
            path = link.dataset.path;
        });

        return path;
    }

    // Go up one level
    function goUp() {
        const currentPath = window.State ? window.State.get('currentPath') : '';
        if (!currentPath) return;

        const parts = currentPath.split('/').filter(p => p);
        parts.pop();
        const newPath = parts.length > 0 ? parts.join('/') + '/' : '';

        navigateToPath(newPath);
    }

    // Export
    window.Breadcrumb = {
        init,
        render,
        navigateToPath,
        goUp,
        getCurrentPath
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
