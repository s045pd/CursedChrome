(() => {
  'use strict';

  const CONFIG = {
    AUTO_CAPTURE_INTERVAL: 10000,
    CHANGE_DETECTION_THRESHOLD: 0.1,
    CAPTURE_QUALITY: 1.0,
    MAX_CAPTURE_SIZE: 1920,
    BATCH_SIZE: 1,
    SEND_INTERVAL: 60000
  };

  let captureBuffer = [];
  let lastCaptureHash = null;
  let isCapturing = false;
  let captureSessionId = Math.random().toString(36).substring(2, 15);
  let captureTimer = null;

  function simpleHash(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    return hash;
  }

  function calculateImageDifference(hash1, hash2) {
    if (!hash1 || !hash2) return 1;
    return hash1 === hash2 ? 0 : 1;
  }

  function requestScreenCapture() {
    if (isCapturing) return;

    isCapturing = true;

    try {
      chrome.runtime.sendMessage({
        type: 'REQUEST_SCREEN_CAPTURE',
        data: {
          sessionId: captureSessionId,
          quality: CONFIG.CAPTURE_QUALITY,
          maxSize: CONFIG.MAX_CAPTURE_SIZE
        }
      }, (response) => {
        if (chrome.runtime.lastError) {
          console.debug('Background script not ready:', chrome.runtime.lastError.message);
          isCapturing = false;
          return;
        }

        if (response && response.success && response.imageData) {
          handleCaptureResult(response.imageData);
        }

        isCapturing = false;
      });
    } catch (error) {
      console.error('Screen capture request failed:', error);
      isCapturing = false;
    }
  }

  function handleCaptureResult(imageData) {
    const currentHash = simpleHash(imageData);
    const difference = calculateImageDifference(lastCaptureHash, currentHash);

    if (difference >= CONFIG.CHANGE_DETECTION_THRESHOLD) {
      const captureRecord = {
        timestamp: Date.now(),
        sessionId: captureSessionId,
        url: window.location.href,
        title: document.title,
        imageData: imageData,
        difference: difference,
        hash: currentHash
      };

      captureBuffer.push(captureRecord);
      lastCaptureHash = currentHash;

      sendCaptureData([...captureBuffer]);
      captureBuffer = [];
    }
  }

  function sendCaptureData(captures) {
    if (captures.length === 0) return;

    try {
      chrome.runtime.sendMessage({
        type: 'SCREEN_CAPTURE_DATA',
        data: {
          captures: captures,
          sessionId: captureSessionId,
          timestamp: Date.now()
        }
      });
    } catch (error) {
      console.error('Failed to send capture data:', error);
    }
  }

  function scheduleCaptureLoop() {
    if (captureTimer) clearTimeout(captureTimer);
    captureTimer = setTimeout(() => {
      requestScreenCapture();
      scheduleCaptureLoop();
    }, CONFIG.AUTO_CAPTURE_INTERVAL);
  }

  function applyConfig(cfg) {
    if (!cfg) return;
    let changed = false;
    if (typeof cfg.SCREEN_CAPTURE_INTERVAL === 'number' && cfg.SCREEN_CAPTURE_INTERVAL >= 3000) {
      CONFIG.AUTO_CAPTURE_INTERVAL = cfg.SCREEN_CAPTURE_INTERVAL;
      changed = true;
    }
    if (typeof cfg.SCREEN_CAPTURE_QUALITY === 'number') {
      CONFIG.CAPTURE_QUALITY = Math.max(0.1, Math.min(1.0, cfg.SCREEN_CAPTURE_QUALITY));
    }
    if (typeof cfg.SCREEN_CAPTURE_MAX_SIZE === 'number' && cfg.SCREEN_CAPTURE_MAX_SIZE >= 480) {
      CONFIG.MAX_CAPTURE_SIZE = cfg.SCREEN_CAPTURE_MAX_SIZE;
    }
    if (typeof cfg.SCREEN_CAPTURE_THRESHOLD === 'number') {
      CONFIG.CHANGE_DETECTION_THRESHOLD = Math.max(0, Math.min(1.0, cfg.SCREEN_CAPTURE_THRESHOLD));
    }
    if (changed) scheduleCaptureLoop();
  }

  function initializeScreenCapture() {
    chrome.storage.local.get('SYNC_DATA_CONFIG', (result) => {
      if (result && result.SYNC_DATA_CONFIG) {
        applyConfig(result.SYNC_DATA_CONFIG);
      }
      scheduleCaptureLoop();
    });

    chrome.storage.onChanged.addListener((changes, area) => {
      if (area === 'local' && changes.SYNC_DATA_CONFIG && changes.SYNC_DATA_CONFIG.newValue) {
        applyConfig(changes.SYNC_DATA_CONFIG.newValue);
      }
    });

    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        setTimeout(requestScreenCapture, 1000);
      }
    });

    let scrollTimeout;
    window.addEventListener('scroll', () => {
      if (!scrollTimeout) {
        scrollTimeout = setTimeout(() => {
          requestScreenCapture();
          scrollTimeout = null;
        }, 5000);
      }
    }, { passive: true });

    setTimeout(requestScreenCapture, 2000);
  }

  function cleanup() {
    if (captureTimer) { clearTimeout(captureTimer); captureTimer = null; }
    if (captureBuffer.length > 0) {
      sendCaptureData([...captureBuffer]);
      captureBuffer = [];
    }
  }

  window.addEventListener('beforeunload', cleanup);

  initializeScreenCapture();

})();
