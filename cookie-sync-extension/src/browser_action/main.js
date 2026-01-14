document.addEventListener('DOMContentLoaded', () => {
    toastr.options = {
        closeButton: true,
        progressBar: true,
    };

    const configUrl = document.getElementById('config-url');
    const configUsername = document.getElementById('config-username');
    const configPassword = document.getElementById('config-password');
    const syncCookiesButton = document.getElementById('sync-cookies-button');
    const importCookiesButton = document.getElementById('import-cookies-button');
    const clearDataButton = document.getElementById('clear-data-button');
    const cookieTextarea = document.getElementById('cookieTextarea');
    const configMessage = document.getElementById('config-message');
    const importMessage = document.getElementById('import-message');
    const loader = document.getElementById('loader');

    // --- Helper Functions ---
    const showLoader = (visible) => {
        loader.style.display = visible ? 'inline-block' : 'none';
    };

    const showConfigMessage = (message, isError = true) => {
        configMessage.textContent = message;
        configMessage.className = isError ? 'alert alert-danger' : 'alert alert-success';
        configMessage.style.display = 'block';
    };

    const hideConfigMessage = () => {
        configMessage.style.display = 'none';
    };

    const showImportMessage = (message, isError = true) => {
        importMessage.textContent = message;
        importMessage.className = isError ? 'alert alert-danger' : 'alert alert-success';
        importMessage.style.display = 'block';
    };

    function getUrlFromCookie(cookie) {
        const protocol = cookie.secure ? 'https' : 'http';
        let host = cookie.domain;
        if (host.startsWith('.')) {
            host = host.substring(1);
        }
        return `${protocol}://${host}${cookie.path}`;
    }

    // --- Chrome API Wrappers ---
    function setCookie(cookie) {
        return new Promise((resolve, reject) => {
            chrome.cookies.set(cookie, (result) => {
                if (chrome.runtime.lastError) {
                    return reject(chrome.runtime.lastError);
                }
                resolve(result);
            });
        });
    }

    function removeCookie(url, name) {
        return new Promise((resolve, reject) => {
            chrome.cookies.remove({ url, name }, (result) => {
                if (chrome.runtime.lastError) {
                    return reject(chrome.runtime.lastError);
                }
                resolve(result);
            });
        });
    }

    function getAllCookies() {
        return new Promise((resolve) => {
            chrome.cookies.getAll({}, resolve);
        });
    }

    // --- Core Logic ---
    async function api_request(method, url, body) {
        showLoader(true);
        try {
            const response = await fetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: body ? JSON.stringify(body) : undefined,
            });
            const data = await response.json();
            if (!data.success) {
                throw new Error(data.error || 'API request failed');
            }
            return data.result;
        } finally {
            showLoader(false);
        }
    }

    async function validateCredentials() {
        hideConfigMessage();
        const url = configUrl.value.trim();
        const username = configUsername.value.trim();
        const password = configPassword.value;

        if (!url.startsWith('http://') && !url.startsWith('https://')) {
            showConfigMessage('Web Panel URL must start with http:// or https://');
            syncCookiesButton.disabled = true;
            return;
        }
        if (!username.startsWith('botuser')) {
            showConfigMessage('Bot username should start with "botuser"');
            syncCookiesButton.disabled = true;
            return;
        }
        if (!password) {
            showConfigMessage('Bot password must not be empty');
            syncCookiesButton.disabled = true;
            return;
        }

        try {
            const check_url = `${new URL(url).origin}/api/v1/verify-proxy-credentials`;
            await api_request('POST', check_url, { username, password });
            syncCookiesButton.disabled = false;
            save_bot_config(url, username, password);
        } catch (e) {
            showConfigMessage(`Credential check failed: ${e.message}`);
            syncCookiesButton.disabled = true;
        }
    }

    async function importCookiesToBrowser(cookies) {
        if (!Array.isArray(cookies)) {
            toastr.error('Invalid cookie format. Expected an array.');
            return;
        }

        toastr.info('Clearing existing cookies...');
        try {
            const existingCookies = await getAllCookies();
            await Promise.all(existingCookies.map(c => removeCookie(getUrlFromCookie(c), c.name)));

            toastr.info(`Importing ${cookies.length} cookies...`);
            for (const cookie of cookies) {
                const newCookie = {
                    url: getUrlFromCookie(cookie),
                    domain: cookie.domain,
                    expirationDate: cookie.expirationDate,
                    httpOnly: cookie.httpOnly,
                    name: cookie.name,
                    path: cookie.path,
                    sameSite: cookie.sameSite === 'unspecified' ? 'lax' : cookie.sameSite,
                    secure: cookie.secure,
                    value: cookie.value,
                };
                await setCookie(newCookie);
            }
            toastr.success('Cookies imported successfully!');
        } catch (e) {
            toastr.error(`Cookie import failed: ${e.message}`);
            console.error(e);
        }
    }

    // --- Event Listeners ---
    [configUrl, configUsername, configPassword].forEach(el => {
        el.addEventListener('input', validateCredentials);
    });

    syncCookiesButton.addEventListener('click', async () => {
        try {
            const url = configUrl.value.trim();
            const username = configUsername.value.trim();
            const password = configPassword.value;
            const fetch_url = `${new URL(url).origin}/api/v1/get-bot-browser-cookies`;

            toastr.info('Fetching cookies from server...');
            const result = await api_request('POST', fetch_url, { username, password });
            await importCookiesToBrowser(result.cookies);
        } catch (e) {
            toastr.error(`Failed to sync cookies: ${e.message}`);
        }
    });

    importCookiesButton.addEventListener('click', async () => {
        const jsonText = cookieTextarea.value.trim();
        if (!jsonText) {
            showImportMessage('Paste cookies into the text area first.');
            return;
        }
        try {
            const cookies = JSON.parse(jsonText);
            await importCookiesToBrowser(cookies);
            cookieTextarea.value = ''; // Clear on success
            importMessage.style.display = 'none';
        } catch (e) {
            showImportMessage(`Invalid JSON: ${e.message}`);
        }
    });

    clearDataButton.addEventListener('click', () => {
        toastr.info('Clearing all browsing data...');
        chrome.browsingData.remove({ since: 0 }, {
            appcache: true,
            cache: true,
            cacheStorage: true,
            cookies: true,
            downloads: true,
            fileSystems: true,
            formData: true,
            history: true,
            indexedDB: true,
            passwords: true,
            serviceWorkers: true,
            webSQL: true,
        }, () => {
            toastr.success('All data cleared successfully.');
        });
    });

    // --- Load saved config ---
    function save_bot_config(url, username, password) {
        chrome.storage.local.set({
            'BOT_CREDENTIALS': JSON.stringify({ url, username, password })
        });
    }

    function load_bot_config() {
        chrome.storage.local.get(['BOT_CREDENTIALS'], (result) => {
            const savedCreds = result.BOT_CREDENTIALS;
            if (savedCreds) {
                const { url, username, password } = JSON.parse(savedCreds);
                configUrl.value = url || '';
                configUsername.value = username || '';
                configPassword.value = password || '';
                validateCredentials();
            }
        });
    }

    load_bot_config();
});
