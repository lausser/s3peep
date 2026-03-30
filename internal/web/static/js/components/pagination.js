// S3 File Browser - Pagination Component
// Page through large folders using S3 continuation tokens

(function() {
    'use strict';

    let container = null;
    let btnFirst = null;
    let btnPrev = null;
    let btnNext = null;
    let btnLast = null;
    let pageInfo = null;
    let pageSizeSelect = null;

    // Initialize the component
    function init() {
        container = document.getElementById('pagination');
        if (!container) {
            console.error('Pagination container not found');
            return;
        }

        // Cache elements
        btnFirst = document.getElementById('btn-first');
        btnPrev = document.getElementById('btn-prev');
        btnNext = document.getElementById('btn-next');
        btnLast = document.getElementById('btn-last');
        pageInfo = document.getElementById('page-info');
        pageSizeSelect = document.getElementById('page-size');

        // Add event listeners
        if (btnFirst) {
            btnFirst.addEventListener('click', () => goToPage(1));
        }

        if (btnPrev) {
            btnPrev.addEventListener('click', () => {
                const currentPage = getCurrentPage();
                if (currentPage > 1) {
                    goToPage(currentPage - 1);
                }
            });
        }

        if (btnNext) {
            btnNext.addEventListener('click', () => {
                const currentPage = getCurrentPage();
                goToPage(currentPage + 1);
            });
        }

        if (btnLast) {
            // S3 doesn't support "last page", so we disable this
            btnLast.disabled = true;
            btnLast.title = 'Last page not available with S3 pagination';
        }

        if (pageSizeSelect) {
            pageSizeSelect.addEventListener('change', (e) => {
                const size = parseInt(e.target.value, 10);
                setPageSize(size);
            });
        }

        // Subscribe to state changes
        if (window.State) {
            window.State.subscribe('pagination:changed', () => {
                updateUI();
            });
        }
    }

    // Get current page from state
    function getCurrentPage() {
        if (window.State) {
            return window.State.get('pagination').currentPage;
        }
        return 1;
    }

    // Go to a specific page
    function goToPage(page) {
        if (page < 1) return;

        // Update state
        if (window.State) {
            window.State.setPage(page);
        }

        // Load files for this page
        if (window.FileList) {
            window.FileList.goToPage(page);
        }

        // Update UI
        updateUI();
    }

    // Set page size
    function setPageSize(size) {
        if (window.State) {
            window.State.setPageSize(size);
        }

        // Reload first page with new size
        goToPage(1);
    }

    // Update UI based on current state
    function updateUI() {
        if (!window.State) return;

        const pagination = window.State.get('pagination');
        const currentPage = pagination.currentPage;
        const isTruncated = pagination.isTruncated;

        // Update page info text
        if (pageInfo) {
            if (isTruncated) {
                pageInfo.textContent = `Page ${currentPage}`;
            } else {
                pageInfo.textContent = `Page ${currentPage} (last)`;
            }
        }

        // Update button states
        if (btnFirst) {
            btnFirst.disabled = currentPage === 1;
        }

        if (btnPrev) {
            btnPrev.disabled = currentPage === 1;
        }

        if (btnNext) {
            btnNext.disabled = !isTruncated;
        }

        // Last button is always disabled (S3 limitation)
        if (btnLast) {
            btnLast.disabled = true;
        }

        // Update page size select
        if (pageSizeSelect) {
            pageSizeSelect.value = pagination.pageSize.toString();
        }
    }

    // Show pagination
    function show() {
        if (container) {
            container.classList.remove('hidden');
        }
    }

    // Hide pagination
    function hide() {
        if (container) {
            container.classList.add('hidden');
        }
    }

    // Export
    window.Pagination = {
        init,
        goToPage,
        setPageSize,
        show,
        hide,
        updateUI
    };

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
