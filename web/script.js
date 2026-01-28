const messagesDiv = document.getElementById('messages');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');
const statusDiv = document.getElementById('status');
const fileInput = document.getElementById('fileInput');
const fileLabel = document.getElementById('fileLabel');
const fileName = document.getElementById('fileName');

let ws = null;
let isOwnMessage = false;
let selectedFile = null;

function connect() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    statusDiv.textContent = 'Connected \u2713';
    statusDiv.className = 'status connected';
    messageInput.disabled = false;
    fileInput.disabled = false;
    updateSendButton();
    messageInput.focus();
  };

  ws.onclose = () => {
    statusDiv.textContent = 'Disconnected. Reconnecting...';
    statusDiv.className = 'status disconnected';
    messageInput.disabled = true;
    sendButton.disabled = true;
    fileInput.disabled = true;
    setTimeout(connect, 2000);
  };

  ws.onerror = error => {
    console.error('WebSocket error:', error);
    statusDiv.textContent = 'Connection error';
    statusDiv.className = 'status disconnected';
  };

  ws.onmessage = event => {
    if (isOwnMessage) {
      isOwnMessage = false;
      return;
    }

    const data = JSON.parse(event.data);
    const fileId = data.file?.id || data.id;
    addMessage(data.text || '', data.file, false, fileId || data.id, data.senderIp || '');
  };
}

function formatFileSize(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
}

function escapeHTML(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function addMessage(text, file, isOwn, messageId, senderIp) {
  const emptyState = messagesDiv.querySelector('.empty-state');
  if (emptyState) {
    emptyState.remove();
  }

  const messageDiv = document.createElement('div');
  messageDiv.className = `message ${isOwn ? 'own' : 'other'}`;

  const headerDiv = document.createElement('div');
  headerDiv.className = 'message-header';
  headerDiv.textContent = isOwn ? 'You' : senderIp || 'Unknown';
  messageDiv.appendChild(headerDiv);

  if (text) {
    const textDiv = document.createElement('div');
    textDiv.className = 'message-text';
    textDiv.textContent = text;
    messageDiv.appendChild(textDiv);
  }

  if (file) {
    const fileDiv = document.createElement('div');
    fileDiv.className = 'message-file';
    fileDiv.innerHTML = `
      <span class="file-icon">\uD83D\uDCCE</span>
      <span class="file-info">
        <span class="file-name-display">${escapeHTML(file.name)}</span>
        <span class="file-size">${formatFileSize(file.size)}</span>
      </span>
    `;
    messageDiv.appendChild(fileDiv);
  }

  const hasActions = (text && !isMobile()) || file;
  if (hasActions) {
    const actionsDiv = document.createElement('div');
    actionsDiv.className = 'message-actions';

    if (text && !isMobile()) {
      const copyBtn = document.createElement('button');
      copyBtn.className = 'action-btn copy-btn';
      copyBtn.textContent = 'Copy';
      copyBtn.onclick = () => copyMessage(text, copyBtn);
      actionsDiv.appendChild(copyBtn);
    }

    if (file) {
      const fileId = file.id || messageId;
      const downloadBtn = document.createElement('button');
      downloadBtn.className = 'action-btn download-btn';
      downloadBtn.textContent = 'Download';
      downloadBtn.onclick = () => downloadFile(fileId, file.name);
      actionsDiv.appendChild(downloadBtn);
    }

    messageDiv.appendChild(actionsDiv);
  }

  messagesDiv.appendChild(messageDiv);
  messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function isMobile() {
  return /Android|iPhone|iPad|iPod/i.test(navigator.userAgent);
}

function copyMessage(text, button) {
  function onSuccess() {
    button.classList.add('copied');
    button.textContent = 'Copied';
    setTimeout(() => {
      button.classList.remove('copied');
      button.textContent = 'Copy';
    }, 2000);
  }

  if (navigator.clipboard && window.isSecureContext) {
    navigator.clipboard
      .writeText(text)
      .then(onSuccess)
      .catch(() => {
        fallbackCopy(text, onSuccess);
      });
  } else {
    fallbackCopy(text, onSuccess);
  }
}

function fallbackCopy(text, onSuccess) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  try {
    document.execCommand('copy');
    onSuccess();
  } catch (err) {
    console.error('Copy failed:', err);
  }
  document.body.removeChild(textarea);
}

function downloadFile(fileId, fileName) {
  const url = `/file/${fileId}`;
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

async function uploadFile(file) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/upload', {
    method: 'POST',
    body: formData
  });

  if (!response.ok) {
    throw new Error('Upload failed');
  }

  return await response.json();
}

async function sendMessage() {
  const text = messageInput.value.trim();
  const hasFile = selectedFile !== null;

  if ((!text && !hasFile) || !ws || ws.readyState !== WebSocket.OPEN) {
    return;
  }

  let fileData = null;
  let fileId = null;

  if (hasFile) {
    try {
      sendButton.disabled = true;
      sendButton.textContent = '⌛';
      const uploadResult = await uploadFile(selectedFile);
      fileData = {
        name: uploadResult.name,
        size: uploadResult.size,
        type: uploadResult.type
      };
      fileId = uploadResult.id;
    } catch (error) {
      alert('Failed to upload file. Please try again.');
      sendButton.disabled = false;
      sendButton.textContent = '➤';
      return;
    }
  }

  const messageId = fileId || Date.now().toString();
  const message = {
    id: messageId,
    text: text,
    file: fileData ? { ...fileData, id: fileId } : null
  };

  isOwnMessage = true;
  ws.send(JSON.stringify(message));
  addMessage(text, fileData ? { ...fileData, id: fileId } : null, true, messageId, '');

  messageInput.value = '';
  messageInput.style.height = 'auto';
  selectedFile = null;
  fileInput.value = '';
  fileName.textContent = '';
  fileLabel.textContent = '📎';
  sendButton.textContent = '➤';
  updateSendButton();
  messageInput.focus();
}

function updateSendButton() {
  const hasText = messageInput.value.trim().length > 0;
  const hasFile = selectedFile !== null;
  sendButton.disabled = !hasText && !hasFile;
}

fileInput.addEventListener('change', e => {
  const file = e.target.files[0];
  if (file) {
    selectedFile = file;
    fileName.textContent = file.name;
  } else {
    selectedFile = null;
    fileName.textContent = '';
  }
  updateSendButton();
});

sendButton.addEventListener('click', sendMessage);

messageInput.addEventListener('keydown', e => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
});

messageInput.addEventListener('input', function () {
  this.style.height = 'auto';
  this.style.height = this.scrollHeight + 'px';
  updateSendButton();
});

// Fetch and display version
fetch('/api/version')
  .then(response => response.text())
  .then(version => {
    const versionSpan = document.querySelector('.version');
    if (versionSpan) {
      versionSpan.textContent = `v${version.trim()}`;
    }
  })
  .catch(err => console.error('Failed to fetch version:', err));

connect();
