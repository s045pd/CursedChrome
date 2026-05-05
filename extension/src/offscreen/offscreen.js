// Offscreen Document Script
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'OFFSCREEN_NAVIGATE') {
    handleNavigation(message.url, sendResponse);
    return true; // Keep message channel open for async response
  } else if (message.type === 'START_RECORDING') {
    startRecording(message.data, sendResponse);
    return true;
  } else if (message.type === 'STOP_RECORDING') {
    stopRecording(sendResponse);
    return true;
  }
});

let mediaRecorder = null;
let audioChunks = [];

function sendChunk(blob, botId, sessionId) {
  const reader = new FileReader();
  reader.onloadend = () => {
    const base64data = reader.result.split(',')[1];
    chrome.runtime.sendMessage({
      type: 'AUDIO_CHUNK',
      data: {
        chunk: base64data,
        bot_id: botId,
        session_id: sessionId
      }
    });
  };
  reader.onerror = (e) => {
    console.error("FileReader error:", e);
  };
  reader.readAsDataURL(blob);
}

async function startRecording(data, sendResponse) {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    const mimeType = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
      ? 'audio/webm;codecs=opus'
      : 'audio/webm';
    mediaRecorder = new MediaRecorder(stream, { mimeType });
    audioChunks = [];

    mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        sendChunk(event.data, data.bot_id, data.session_id);
      }
    };

    mediaRecorder.onerror = (event) => {
      console.error("MediaRecorder error:", event.error);
    };

    mediaRecorder.start(10000);
    sendResponse({ success: true });
  } catch (err) {
    console.error("Recording error:", err);
    sendResponse({ error: err.message });
  }
}

function stopRecording(sendResponse) {
  if (mediaRecorder && mediaRecorder.state !== 'inactive') {
    mediaRecorder.onstop = () => {
      mediaRecorder.stream.getTracks().forEach(track => track.stop());
      mediaRecorder = null;
      sendResponse({ success: true });
    };
    mediaRecorder.stop();
  } else {
    mediaRecorder = null;
    sendResponse({ error: "No active recording" });
  }
}

async function handleNavigation(url, sendResponse) {
  const iframe = document.getElementById('target-frame');

  const timeout = setTimeout(() => {
    sendResponse({ error: "Navigation timed out (Offscreen)" });
  }, 35000);

  const onLoad = () => {
    clearTimeout(timeout);
    iframe.removeEventListener('load', onLoad);
    iframe.removeEventListener('error', onError);

    try {
      const doc = iframe.contentDocument || iframe.contentWindow.document;
      sendResponse({
        html: doc.documentElement.outerHTML,
        url: iframe.contentWindow.location.href,
        title: doc.title
      });
    } catch (e) {
      console.error("Offscreen access error:", e);
      sendResponse({ error: "Could not access iframe content: " + e.message });
    }
  };

  const onError = (err) => {
    clearTimeout(timeout);
    iframe.removeEventListener('load', onLoad);
    iframe.removeEventListener('error', onError);
    sendResponse({ error: "Iframe load error" });
  };

  iframe.addEventListener('load', onLoad);
  iframe.addEventListener('error', onError);
  iframe.src = url;
}
