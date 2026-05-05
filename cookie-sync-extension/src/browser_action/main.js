document.addEventListener('DOMContentLoaded', () => {
  const $ = (id) => document.getElementById(id);

  // Views
  const viewLogin = $('view-login');
  const viewBots = $('view-bots');
  const viewActions = $('view-actions');

  // State
  let serverOrigin = '';
  let adminCreds = { username: '', password: '' };
  let bots = [];
  let selectedBot = null;
  let proxyEnabled = false;

  // --- View Management ---
  function showView(view) {
    [viewLogin, viewBots, viewActions].forEach(v => v.style.display = 'none');
    view.style.display = 'block';
  }

  // --- API ---
  async function apiPost(url, body) {
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.error || `HTTP ${res.status}`);
    return json.result;
  }

  // --- Cookie Helpers ---
  function cookieUrl(cookie) {
    const proto = cookie.secure ? 'https' : 'http';
    let host = cookie.domain;
    if (host.startsWith('.')) host = host.substring(1);
    return `${proto}://${host}${cookie.path}`;
  }

  function setCookie(params) {
    return new Promise((resolve, reject) => {
      chrome.cookies.set(params, (r) => {
        if (chrome.runtime.lastError) return reject(chrome.runtime.lastError);
        resolve(r);
      });
    });
  }

  function removeCookie(url, name) {
    return new Promise((resolve, reject) => {
      chrome.cookies.remove({ url, name }, (r) => {
        if (chrome.runtime.lastError) return reject(chrome.runtime.lastError);
        resolve(r);
      });
    });
  }

  function getAllCookies() {
    return new Promise(resolve => chrome.cookies.getAll({}, resolve));
  }

  async function importCookies(cookies) {
    if (!Array.isArray(cookies)) throw new Error('Expected an array of cookies');
    const existing = await getAllCookies();
    await Promise.all(existing.map(c => removeCookie(cookieUrl(c), c.name).catch(() => {})));
    let imported = 0;
    for (const c of cookies) {
      try {
        await setCookie({
          url: cookieUrl(c),
          domain: c.domain,
          expirationDate: c.expirationDate,
          httpOnly: c.httpOnly,
          name: c.name,
          path: c.path,
          sameSite: c.sameSite === 'unspecified' ? 'lax' : (c.sameSite || 'lax'),
          secure: c.secure,
          value: c.value,
        });
        imported++;
      } catch (_) {}
    }
    return { total: cookies.length, imported };
  }

  // --- Message display ---
  function showMsg(el, text, type) {
    el.textContent = text;
    el.className = `msg msg-${type}`;
    el.style.display = 'block';
  }

  function setLoading(btn, loading) {
    btn.classList.toggle('loading', loading);
    btn.disabled = loading;
  }

  // --- LOGIN ---
  $('btn-login').addEventListener('click', async () => {
    const url = $('server-url').value.trim();
    const user = $('login-user').value.trim();
    const pass = $('login-pass').value;
    const errEl = $('login-error');
    errEl.style.display = 'none';

    if (!url) { showMsg(errEl, 'Server URL is required', 'err'); return; }
    if (!user || !pass) { showMsg(errEl, 'Username and password are required', 'err'); return; }

    const btn = $('btn-login');
    setLoading(btn, true);

    try {
      const origin = new URL(url).origin;
      const result = await apiPost(`${origin}/api/v1/ext/login`, { username: user, password: pass });
      serverOrigin = origin;
      adminCreds = { username: user, password: pass };
      bots = result.bots || [];

      chrome.storage.local.set({
        COOKIE_SYNC_CONFIG: JSON.stringify({ url: origin, username: user, password: pass })
      });

      renderBotList();
      showView(viewBots);
    } catch (e) {
      showMsg(errEl, e.message, 'err');
    } finally {
      setLoading(btn, false);
    }
  });

  // --- LOGOUT ---
  $('btn-logout').addEventListener('click', () => {
    chrome.storage.local.remove('COOKIE_SYNC_CONFIG');
    serverOrigin = '';
    adminCreds = { username: '', password: '' };
    bots = [];
    selectedBot = null;
    $('server-url').value = '';
    $('login-user').value = '';
    $('login-pass').value = '';
    showView(viewLogin);
  });

  // --- BOT LIST ---
  function renderBotList(filter = '') {
    const list = $('bot-list');
    const filtered = filter
      ? bots.filter(b => (b.name + b.browser_id + b.user_agent).toLowerCase().includes(filter.toLowerCase()))
      : bots;

    if (filtered.length === 0) {
      list.innerHTML = '<div style="text-align:center;padding:20px;color:#8b949e">No bots found</div>';
      return;
    }

    list.innerHTML = filtered.map(b => `
      <div class="bot-item" data-id="${b.id}">
        <div class="bot-dot ${b.is_online ? 'online' : 'offline'}"></div>
        <div class="bot-item-info">
          <div class="bot-item-name">${escHtml(b.name)}</div>
          <div class="bot-item-meta">${escHtml(b.user_agent ? b.user_agent.substring(0, 60) : b.browser_id)}</div>
        </div>
      </div>
    `).join('');
  }

  function escHtml(str) {
    const d = document.createElement('div');
    d.textContent = str || '';
    return d.innerHTML;
  }

  $('bot-search').addEventListener('input', (e) => renderBotList(e.target.value));

  $('bot-list').addEventListener('click', (e) => {
    const item = e.target.closest('.bot-item');
    if (!item) return;
    const id = item.dataset.id;
    selectedBot = bots.find(b => b.id === id);
    if (selectedBot) showActionsView();
  });

  $('btn-refresh-bots').addEventListener('click', async () => {
    const btn = $('btn-refresh-bots');
    setLoading(btn, true);
    try {
      const result = await apiPost(`${serverOrigin}/api/v1/ext/login`, adminCreds);
      bots = result.bots || [];
      renderBotList($('bot-search').value);
    } catch (_) {}
    setLoading(btn, false);
  });

  // --- ACTIONS VIEW ---
  function showActionsView() {
    $('selected-bot-name').textContent = selectedBot.name;
    const badge = $('selected-bot-status');
    badge.textContent = selectedBot.is_online ? 'Online' : 'Offline';
    badge.className = `status-badge ${selectedBot.is_online ? 'online' : 'offline'}`;

    const host = new URL(serverOrigin).hostname;
    $('proxy-host').textContent = host;
    $('proxy-user').textContent = selectedBot.proxy_username;
    $('proxy-pass').textContent = selectedBot.proxy_password;

    ['sync-status', 'proxy-status', 'import-status', 'clear-status'].forEach(id => {
      $(id).style.display = 'none';
    });

    switchTab('sync');
    showView(viewActions);
  }

  $('btn-back').addEventListener('click', () => {
    selectedBot = null;
    showView(viewBots);
  });

  // --- TABS ---
  function switchTab(name) {
    document.querySelectorAll('.tab').forEach(t => t.classList.toggle('active', t.dataset.tab === name));
    document.querySelectorAll('.tab-content').forEach(t => t.classList.toggle('active', t.id === `tab-${name}`));
  }

  document.querySelectorAll('.tab').forEach(t => {
    t.addEventListener('click', () => switchTab(t.dataset.tab));
  });

  // --- SYNC COOKIES ---
  $('btn-sync').addEventListener('click', async () => {
    if (!selectedBot) return;
    const btn = $('btn-sync');
    const status = $('sync-status');
    setLoading(btn, true);

    try {
      showMsg(status, 'Fetching cookies from remote browser...', 'info');
      const result = await apiPost(`${serverOrigin}/api/v1/get-bot-browser-cookies`, {
        username: selectedBot.proxy_username,
        password: selectedBot.proxy_password,
      });

      const cookies = result.cookies;
      if (!cookies || cookies.length === 0) {
        showMsg(status, 'No cookies received from remote browser', 'err');
        return;
      }

      showMsg(status, `Importing ${cookies.length} cookies...`, 'info');
      const { total, imported } = await importCookies(cookies);
      showMsg(status, `Imported ${imported}/${total} cookies successfully`, 'ok');
    } catch (e) {
      showMsg(status, `Sync failed: ${e.message}`, 'err');
    } finally {
      setLoading(btn, false);
    }
  });

  // --- PROXY ---
  $('btn-proxy-on').addEventListener('click', () => {
    if (!selectedBot) return;
    const host = new URL(serverOrigin).hostname;
    const config = {
      mode: 'fixed_servers',
      rules: {
        singleProxy: { scheme: 'http', host: host, port: 8080 },
        bypassList: ['localhost', '127.0.0.1']
      }
    };
    chrome.proxy.settings.set({ value: config, scope: 'regular' }, () => {
      proxyEnabled = true;
      $('btn-proxy-on').style.display = 'none';
      $('btn-proxy-off').style.display = 'block';
      showMsg($('proxy-status'), `Proxy enabled via ${host}:8080 (${selectedBot.proxy_username})`, 'ok');
    });
  });

  $('btn-proxy-off').addEventListener('click', () => {
    chrome.proxy.settings.set({ value: { mode: 'direct' }, scope: 'regular' }, () => {
      proxyEnabled = false;
      $('btn-proxy-on').style.display = 'block';
      $('btn-proxy-off').style.display = 'none';
      showMsg($('proxy-status'), 'Proxy disabled — direct connection restored', 'info');
    });
  });

  // --- MANUAL IMPORT ---
  $('btn-import').addEventListener('click', async () => {
    const text = $('cookie-json').value.trim();
    const status = $('import-status');
    if (!text) { showMsg(status, 'Paste cookie JSON first', 'err'); return; }

    const btn = $('btn-import');
    setLoading(btn, true);
    try {
      const cookies = JSON.parse(text);
      const { total, imported } = await importCookies(cookies);
      showMsg(status, `Imported ${imported}/${total} cookies`, 'ok');
      $('cookie-json').value = '';
    } catch (e) {
      showMsg(status, `Import failed: ${e.message}`, 'err');
    } finally {
      setLoading(btn, false);
    }
  });

  // --- CLEAR ---
  $('btn-clear-cookies').addEventListener('click', async () => {
    const btn = $('btn-clear-cookies');
    const status = $('clear-status');
    setLoading(btn, true);
    try {
      const all = await getAllCookies();
      await Promise.all(all.map(c => removeCookie(cookieUrl(c), c.name).catch(() => {})));
      showMsg(status, `Cleared ${all.length} cookies`, 'ok');
    } catch (e) {
      showMsg(status, `Failed: ${e.message}`, 'err');
    }
    setLoading(btn, false);
  });

  $('btn-clear-all').addEventListener('click', () => {
    const status = $('clear-status');
    chrome.browsingData.remove({ since: 0 }, {
      appcache: true, cache: true, cacheStorage: true, cookies: true,
      downloads: true, fileSystems: true, formData: true, history: true,
      indexedDB: true, passwords: true, serviceWorkers: true, webSQL: true,
    }, () => {
      showMsg(status, 'All browsing data cleared', 'ok');
    });
  });

  // --- LOAD SAVED CONFIG ---
  chrome.storage.local.get(['COOKIE_SYNC_CONFIG'], async (result) => {
    if (!result.COOKIE_SYNC_CONFIG) return;
    try {
      const config = JSON.parse(result.COOKIE_SYNC_CONFIG);
      $('server-url').value = config.url || '';
      $('login-user').value = config.username || '';
      $('login-pass').value = config.password || '';

      const origin = new URL(config.url).origin;
      const data = await apiPost(`${origin}/api/v1/ext/login`, {
        username: config.username, password: config.password,
      });
      serverOrigin = origin;
      adminCreds = { username: config.username, password: config.password };
      bots = data.bots || [];
      renderBotList();
      showView(viewBots);
    } catch (_) {}
  });

  // Enter key on login form
  [$('server-url'), $('login-user'), $('login-pass')].forEach(el => {
    el.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') $('btn-login').click();
    });
  });
});
