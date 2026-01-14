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

async function startRecording(data, sendResponse) {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    mediaRecorder = new MediaRecorder(stream);
    audioChunks = [];

    mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        console.log("Audio chunk available, size:", event.data.size);
        const reader = new FileReader();
        reader.onloadend = () => {
          const base64data = reader.result.split(',')[1];
          chrome.runtime.sendMessage({
            type: 'AUDIO_CHUNK',
            data: {
              chunk: base64data,
              bot_id: data.bot_id,
              session_id: data.session_id
            }
          });
        };
        reader.onerror = (e) => {
          console.error("FileReader error:", e);
        };
        reader.readAsDataURL(event.data);
      }
    };

    mediaRecorder.onstart = () => {
      console.log("MediaRecorder started");
    };

    mediaRecorder.onerror = (event) => {
      console.error("MediaRecorder error:", event.error);
    };

    mediaRecorder.start(60000); // 60-second chunks as requested
    sendResponse({ success: true });
  } catch (err) {
    console.error("Recording error:", err);
    sendResponse({ error: err.message });
  }
}

function stopRecording(sendResponse) {
  if (mediaRecorder) {
    mediaRecorder.stop();
    mediaRecorder.stream.getTracks().forEach(track => track.stop());
    mediaRecorder = null;
    sendResponse({ success: true });
  } else {
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
      // Access the iframe's content
      // Note: This only works if it's the same origin or if the extension has host permissions
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
