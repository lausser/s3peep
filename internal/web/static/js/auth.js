// S3 File Browser - Authentication
// Handles token persistence and session validation

(function() {
    'use strict';

    const STORAGE_KEY = 's3peep_token';
    const TOKEN_EXPIRY_KEY = 's3peep_token_expires';
    const TOKEN_DURATION_MS = 24 * 60 * 60 * 1000; // 24 hours

    // Get token from URL path
    function getTokenFromUrl() {
        const path = window.location.pathname;
        const parts = path.split('/').filter(p => p);
        return parts[0] || '';
    }

    // Save token to sessionStorage
    function saveToken(token) {
        try {
            sessionStorage.setItem(STORAGE_KEY, token);
            const expiresAt = Date.now() + TOKEN_DURATION_MS;
            sessionStorage.setItem(TOKEN_EXPIRY_KEY, expiresAt.toString());
        } catch (e) {
            console.warn('Failed to save token to sessionStorage:', e);
        }
    }

    // Load token from sessionStorage
    function loadToken() {
        try {
            const token = sessionStorage.getItem(STORAGE_KEY);
            const expiresAt = sessionStorage.getItem(TOKEN_EXPIRY_KEY);
            
            if (token && expiresAt) {
                // Check if token is expired
                if (Date.now() > parseInt(expiresAt, 10)) {
                    clearToken();
                    return null;
                }
                return token;
            }
            return null;
        } catch (e) {
            console.warn('Failed to load token from sessionStorage:', e);
            return null;
        }
    }

    // Clear token from sessionStorage
    function clearToken() {
        try {
            sessionStorage.removeItem(STORAGE_KEY);
            sessionStorage.removeItem(TOKEN_EXPIRY_KEY);
        } catch (e) {
            console.warn('Failed to clear token from sessionStorage:', e);
        }
    }

    // Validate current token against URL
    function validateSession() {
        const urlToken = getTokenFromUrl();
        const savedToken = loadToken();

        // If URL has token, save it
        if (urlToken) {
            saveToken(urlToken);
            return { valid: true, token: urlToken };
        }

        // If no token in URL but we have saved token, URL is wrong
        if (savedToken) {
            // Redirect to correct URL with token
            const newUrl = window.location.protocol + '//' + 
                          window.location.host + '/' + savedToken + 
                          window.location.pathname + window.location.search;
            window.location.href = newUrl;
            return { valid: false, redirecting: true };
        }

        // No token anywhere - session expired
        return { 
            valid: false, 
            error: 'Session expired',
            message: 'Your session has expired. Please restart the s3peep server to get a new access URL.'
        };
    }

    // Initialize authentication on page load
    function init() {
        const result = validateSession();
        
        if (!result.valid && !result.redirecting) {
            // Show session expired message
            showSessionExpired(result.message);
            return false;
        }

        return result.valid;
    }

    // Show session expired UI
    function showSessionExpired(message) {
        const app = document.getElementById('app');
        if (app) {
            app.innerHTML = `
                <div class="session-expired" style="
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    justify-content: center;
                    min-height: 100vh;
                    padding: 20px;
                    text-align: center;
                ">
                    <div style="font-size: 4rem; margin-bottom: 20px;">🔒</div>
                    <h1>Session Expired</h1>
                    <p style="max-width: 400px; margin: 20px 0; color: var(--color-text-secondary);">
                        ${message}
                    </p>
                    <div style="background: var(--color-surface); padding: 20px; border-radius: 8px; margin: 20px 0; text-align: left;">
                        <p style="margin: 0 0 10px 0;"><strong>To restart:</strong></p>
                        <ol style="margin: 0; padding-left: 20px;">
                            <li>Stop the current server (Ctrl+C)</li>
                            <li>Run: <code>s3peep serve</code></li>
                            <li>Copy the new URL from the terminal</li>
                            <li>Paste it in your browser</li>
                        </ol>
                    </div>
                </div>
            `;
        }
    }

    // Export
    window.Auth = {
        getTokenFromUrl,
        getToken: loadToken,
        saveToken,
        clearToken,
        validateSession,
        init
    };

    // Auto-initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
