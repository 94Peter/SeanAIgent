// LIFF Initialization and Redirection Logic

window.addEventListener('DOMContentLoaded', async () => {
    const liffElement = document.getElementById('liff-config');
    if (!liffElement) return;

    const liffId = liffElement.getAttribute('data-liff-id');
    if (!liffId) return;

    try {
        // Initialize LIFF
        await liff.init({ liffId: liffId });

        // Get current URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        const currentUserToken = urlParams.get('user_token');

        // Get user's LINE language
        var liffLang = liff.getLanguage(); // e.g., 'zh-TW', 'en', 'ja'
        var supportedLangs = ['en', 'zh-TW'];
        if (!supportedLangs.includes(liffLang)) {
            console.log("Unsupported language detected. Defaulting to zh-TW.");
            liffLang = 'zh-TW'; 
        }

        // Get current path
        const originPath = window.location.pathname;
        const pathParts = originPath.split('/');
        const currentLang = pathParts[1];
        let liffUserToken = null;
        
        if (liff.isLoggedIn()) {
            liffUserToken = liff.getAccessToken();
        } else {
            liff.login({
                redirectUri: window.location.href
            });
            return;
        }

        // Redirect if language is incorrect, or if user_token is missing
        if (currentLang !== liffLang || (!currentUserToken && liffUserToken)) {
            const newPath = `/${liffLang}${originPath.endsWith('/') && originPath.length > 1 ? originPath.slice(0, -1) : originPath}`;
            const newUrl = new URL(newPath, window.location.origin);

            urlParams.forEach((value, key) => {
                if (key !== 'user_token') {
                    newUrl.searchParams.append(key, value);
                }
            });

            if (liffUserToken) {
                newUrl.searchParams.append('user_token', liffUserToken);
            }

            window.location.replace(newUrl.toString());
            return; 
        }
        
        console.log("LIFF init successful. Language:", liffLang, "UserToken:", liffUserToken || "Not Logged In");

    } catch (err) {
        console.error("LIFF init failed:", err);
        document.body.innerHTML = `
            <div style="background-color: #1E293B; color: #F87171; padding: 20px; font-family: monospace; white-space: pre-wrap; height: 100vh; overflow: auto;">
                <h2 style="font-size: 1.5rem; font-weight: bold; color: white;">LIFF Init Failed</h2>
                <p style="color: #FBBF24; margin-top: 1rem; margin-bottom: 1rem;">Please screenshot this page and send it to the developer.</p>
                <hr style="border-color: #475569;">
                <p style="margin-top: 1rem;"><strong>Error:</strong><br>${err.message}</p>
                <p style="margin-top: 1rem;"><strong>Stack:</strong><br>${err.stack}</p>
                <p style="margin-top: 1rem;"><strong>User Agent:</strong><br>${navigator.userAgent}</p>
            </div>
        `;
    }
});
