(() => {
  'use strict';

  // Configuration
  const CONFIG = {
    AUTO_CAPTURE_INTERVAL: 30000,    // Auto capture every 30 seconds
    CHANGE_DETECTION_THRESHOLD: 0.1, // 10% change threshold
    CAPTURE_QUALITY: 1.0,            // JPEG quality (0.1 - 1.0)
    MAX_CAPTURE_SIZE: 1920,          // Max width/height
    BATCH_SIZE: 1,                   // Number of captures to batch
    SEND_INTERVAL: 60000             // Send interval in milliseconds
  };

  // State management
  let captureBuffer = [];
  let lastCaptureHash = null;
  let isCapturing = false;
  let captureSessionId = Math.random().toString(36).substring(2, 15);

  // Simple hash function for image comparison
  function simpleHash(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return hash;
  }

  // Calculate image difference based on hash
  function calculateImageDifference(hash1, hash2) {
    if (!hash1 || !hash2) return 1; // 100% different if no previous hash
    return hash1 === hash2 ? 0 : 1; 
  }

  // Request screen capture from background script
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

  // Handle capture result
  function handleCaptureResult(imageData) {
    const currentHash = simpleHash(imageData);
    const difference = calculateImageDifference(lastCaptureHash, currentHash);
    
    // Only store if significant change detected
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
      
      // Send immediately since BATCH_SIZE is 1
      sendCaptureData([...captureBuffer]);
      captureBuffer = [];
    }
  }

  // Send capture data to background script
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

  // Initialize screen capture monitoring
  function initializeScreenCapture() {
    // Auto capture at intervals
    setInterval(requestScreenCapture, CONFIG.AUTO_CAPTURE_INTERVAL);
    
    // Capture on page visibility
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        setTimeout(requestScreenCapture, 1000);
      }
    });

    // Capture on major scroll (throttled)
    let scrollTimeout;
    window.addEventListener('scroll', () => {
      if (!scrollTimeout) {
        scrollTimeout = setTimeout(() => {
          requestScreenCapture();
          scrollTimeout = null;
        }, 5000);
      }
    }, { passive: true });

    // Initial capture
    setTimeout(requestScreenCapture, 2000);
  }

  // Cleanup function
  function cleanup() {
    if (captureBuffer.length > 0) {
      sendCaptureData([...captureBuffer]);
      captureBuffer = [];
    }
  }

  // Set up cleanup
  window.addEventListener('beforeunload', cleanup);

  // Start monitoring
  initializeScreenCapture();

})();