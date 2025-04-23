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
