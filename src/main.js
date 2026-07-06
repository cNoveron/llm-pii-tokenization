const btnWill = document.getElementById('btn-will');
const btnPoa = document.getElementById('btn-poa');
const dropZone = document.getElementById('drop-zone');
const fileInput = document.getElementById('file-input');
const statusMsg = document.getElementById('status');
const pdfViewer = document.getElementById('pdf-viewer');

const factsSection = document.getElementById('facts-section');
const factsTableBody = document.querySelector('#facts-table tbody');

const qaSection = document.getElementById('qa-section');
const questionInput = document.getElementById('question-input');
const askBtn = document.getElementById('ask-btn');
const chatThread = document.getElementById('chat-thread');
const toggleBtn = document.getElementById('toggle-pii');
const suggestionsLabel = document.getElementById('suggestions-label');
const suggestionsBar = document.getElementById('suggestions-bar');

const SUGGESTED_QUESTIONS = [
  "What real property is included in the estate, and are there co-ownership arrangements?",
  "Are any assets held in trust and outside the probate estate?",
  "Does the document acknowledge any outstanding debts, mortgages, or liens?",
  "Who are the residuary beneficiaries and what share does each receive?",
  "Are there any conditions or contingencies attached to any bequest?",
  "Is there a no-contest clause?",
  "What powers are granted to the executor?",
  "Is there a pour-over will that directs assets into a living trust?",
  "Are there any special needs trust provisions?",
  "Does the document reference a marital deduction or QTIP trust?",
];

function renderSuggestions() {
  if (!suggestionsBar) return;
  suggestionsBar.innerHTML = '';
  SUGGESTED_QUESTIONS.forEach(q => {
    const chip = document.createElement('button');
    chip.className = 'suggestion-chip';
    chip.textContent = q;
    chip.addEventListener('click', () => {
      questionInput.value = q;
      askQuestion();
    });
    suggestionsBar.appendChild(chip);
  });
  suggestionsBar.hidden = false;
  if (suggestionsLabel) suggestionsLabel.hidden = false;
}

let sessionId = null;
window.lastUploadSessionId = null;

const tabScan = document.getElementById('tab-scan');
const tabAutomate = document.getElementById('tab-automate');
const scanPage = document.getElementById('scan-page');
const automatePage = document.getElementById('automate-page');

tabScan.addEventListener('click', () => {
  tabScan.classList.add('active');
  tabAutomate.classList.remove('active');
  scanPage.classList.remove('hidden');
  automatePage.classList.add('hidden');
});

tabAutomate.addEventListener('click', () => {
  tabAutomate.classList.add('active');
  tabScan.classList.remove('active');
  automatePage.classList.remove('hidden');
  scanPage.classList.add('hidden');
});

function setState(state, msg = '') {
  statusMsg.className = '';
  statusMsg.textContent = msg;

  if (state === 'ERROR') {
    statusMsg.classList.add('error');
  }

  switch (state) {
    case 'IDLE':
      sessionId = null;
      btnWill.disabled = false;
      btnPoa.disabled = false;
      fileInput.disabled = false;
      factsSection.hidden = true;
      qaSection.hidden = false;
      pdfViewer.hidden = true;
      pdfViewer.src = '';
      askBtn.disabled = true;
      questionInput.disabled = true;
      if (suggestionsLabel) suggestionsLabel.hidden = true;
      if (suggestionsBar) suggestionsBar.hidden = true;
      break;
    case 'LOADING':
      btnWill.disabled = true;
      btnPoa.disabled = true;
      fileInput.disabled = true;
      askBtn.disabled = true;
      questionInput.disabled = true;
      statusMsg.textContent = 'Uploading…';
      break;
    case 'READY':
      btnWill.disabled = false;
      btnPoa.disabled = false;
      fileInput.disabled = false;
      factsSection.hidden = false;
      qaSection.hidden = false;
      pdfViewer.hidden = false;
      askBtn.disabled = false;
      questionInput.disabled = false;
      statusMsg.textContent = 'Ready.';
      break;
    case 'ASKING':
      askBtn.disabled = true;
      questionInput.disabled = true;
      break;
    case 'ERROR':
      btnWill.disabled = false;
      btnPoa.disabled = false;
      fileInput.disabled = false;
      askBtn.disabled = !sessionId;
      questionInput.disabled = !sessionId;
      break;
  }
}

function renderFacts(facts) {
  factsTableBody.innerHTML = '';
  if (!facts || facts.length === 0) {
    const tr = document.createElement('tr');
    const td = document.createElement('td');
    td.colSpan = 2;
    td.textContent = 'No facts extracted.';
    tr.appendChild(td);
    factsTableBody.appendChild(tr);
    return;
  }
  for (const fact of facts) {
    const tr = document.createElement('tr');
    const tdLabel = document.createElement('td');
    tdLabel.textContent = fact.label;
    const tdValue = document.createElement('td');
    tdValue.textContent = fact.value;
    tr.appendChild(tdLabel);
    tr.appendChild(tdValue);
    factsTableBody.appendChild(tr);
  }
}

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function highlightTokens(text) {
  if (!text) return '';
  return escapeHtml(text).replace(
    /\{\{[A-Z_]+_\d+\}\}/g,
    m => `<mark class="pii-token">${m}</mark>`
  );
}

async function seedChatThread(uploadData) {
  chatThread.innerHTML = '';

  const wait = (ms) => new Promise(resolve => setTimeout(resolve, ms));

  function createSpinner(label) {
    const div = document.createElement('div');
    div.className = 'message system';
    div.innerHTML = `
      <span class="block-label">${label}</span>
      <div class="spinner" style="width:16px; height:16px; border-width:3px; margin: 4px 0 0 0;"></div>
    `;
    chatThread.appendChild(div);
    chatThread.scrollTop = chatThread.scrollHeight;
    return div;
  }

  const sysSpinner = createSpinner('Sending System Prompt...');
  await wait(1500);

  sysSpinner.innerHTML = `
    <span class="block-label">System Prompt</span>
    <p class="message-text">${escapeHtml(uploadData.system_prompt)}</p>
  `;
  chatThread.scrollTop = chatThread.scrollHeight;

  const ctxSpinner = createSpinner('Sending Document Context...');
  await wait(1500);

  const TRUNCATE_AT = 400;
  const raw = uploadData.tokenized_doc || '';
  const isTruncated = raw.length > TRUNCATE_AT;
  const truncated = isTruncated
    ? raw.slice(0, raw.lastIndexOf(' ', TRUNCATE_AT))
    : raw;

  ctxSpinner.className = 'message context';
  ctxSpinner.innerHTML = `
    <span class="block-label">Document Context (sent to LLM)</span>
    <p class="tokenized">${highlightTokens(truncated)}${isTruncated
      ? '<br><em class="truncation-notice">— truncated · full document sent to LLM —</em>'
      : ''}</p>
  `;
  chatThread.scrollTop = chatThread.scrollHeight;
}

function appendUserMessage(question, tokenizedQuestion) {
  const mode = toggleBtn.dataset.mode;
  const div = document.createElement('div');
  div.className = 'message user';
  div.innerHTML = `
    <p class="tokenized ${mode === 'detokenized' ? 'hidden' : ''}">${highlightTokens(tokenizedQuestion)}</p>
    <p class="detokenized ${mode === 'tokenized' ? 'hidden' : ''}">${escapeHtml(question)}</p>
  `;
  chatThread.appendChild(div);
  chatThread.scrollTop = chatThread.scrollHeight;
}

function appendAssistantMessage(data) {
  const mode = toggleBtn.dataset.mode;
  const div = document.createElement('div');
  div.className = 'message assistant';
  div.innerHTML = `
    <span class="block-label">Response</span>
    <p class="tokenized ${mode === 'detokenized' ? 'hidden' : ''}">${highlightTokens(data.tokenized_answer)}</p>
    <p class="detokenized ${mode === 'tokenized' ? 'hidden' : ''}">${escapeHtml(data.answer)}</p>
  `;
  chatThread.appendChild(div);
  chatThread.scrollTop = chatThread.scrollHeight;
}

async function uploadFile(file) {
  if (file.type !== 'application/pdf') {
    setState('ERROR', 'Please upload a PDF file.');
    return;
  }
  setState('LOADING');
  try {
    const fileUrl = URL.createObjectURL(file);
    pdfViewer.src = fileUrl;

    const form = new FormData();
    form.append('file', file);
    const res = await fetch('/upload', { method: 'POST', body: form });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Upload failed');
    sessionId = data.session_id;
    window.lastUploadSessionId = data.session_id;
    renderFacts(data.facts);

    statusMsg.textContent = 'Initializing chat...';
    pdfViewer.hidden = false;
    factsSection.hidden = false;
    qaSection.hidden = false;

    await seedChatThread(data);

    renderSuggestions();
    questionInput.value = '';
    setState('READY');
  } catch (err) {
    setState('ERROR', err.message);
  }
}

async function loadSample(name) {
  setState('LOADING');
  try {
    const res = await fetch(`/sample/${name}`);
    if (!res.ok) throw new Error(`Failed to load sample: ${res.statusText}`);
    const blob = await res.blob();
    const file = new File([blob], `${name}.pdf`, { type: 'application/pdf' });
    await uploadFile(file);
  } catch (err) {
    setState('ERROR', err.message);
  }
}

async function askQuestion() {
  const question = questionInput.value.trim();
  if (!question || !sessionId) return;

  setState('ASKING');
  try {
    const res = await fetch('/ask', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ session_id: sessionId, question: question })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Question failed');
    appendUserMessage(question, data.tokenized_prompt);
    appendAssistantMessage(data);
    questionInput.value = '';
    setState('READY');
  } catch (err) {
    setState('ERROR', err.message);
  }
}

const appOverlay = document.getElementById('app-overlay');
const willSampleTooltip = document.getElementById('will-sample-tooltip');

function dismissOnboarding() {
  if (appOverlay) appOverlay.style.display = 'none';
  if (willSampleTooltip) willSampleTooltip.style.display = 'none';
  if (btnWill) btnWill.style.zIndex = 'auto';
}

if (appOverlay) {
  appOverlay.addEventListener('click', dismissOnboarding);
}

btnWill.addEventListener('click', () => {
  dismissOnboarding();
  loadSample('will');
});
btnPoa.addEventListener('click', () => loadSample('poa'));

fileInput.addEventListener('change', (e) => {
  if (e.target.files.length > 0) {
    uploadFile(e.target.files[0]);
  }
});

dropZone.addEventListener('dragover', (e) => {
  e.preventDefault();
  dropZone.classList.add('dragover');
});

dropZone.addEventListener('dragleave', (e) => {
  e.preventDefault();
  dropZone.classList.remove('dragover');
});

dropZone.addEventListener('drop', (e) => {
  e.preventDefault();
  dropZone.classList.remove('dragover');
  if (e.dataTransfer.files.length > 0) {
    uploadFile(e.dataTransfer.files[0]);
  }
});

askBtn.addEventListener('click', askQuestion);
questionInput.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') askQuestion();
});

toggleBtn.addEventListener('click', () => {
  const next = toggleBtn.dataset.mode === 'tokenized' ? 'detokenized' : 'tokenized';
  toggleBtn.dataset.mode = next;
  toggleBtn.textContent = next === 'tokenized' ? '🔒 Show real data' : '🔓 Hide real data';

  document.querySelectorAll('.message.user .tokenized, .message.assistant .tokenized, .automation-card .tokenized, #doc-modal-body .tokenized')
    .forEach(el => el.classList.toggle('hidden', next === 'detokenized'));
  document.querySelectorAll('.message.user .detokenized, .message.assistant .detokenized, .automation-card .detokenized, #doc-modal-body .detokenized')
    .forEach(el => el.classList.toggle('hidden', next === 'tokenized'));
});

setState('IDLE');

// Drag and Drop implementation for Automate Tab
const draggableStages = document.querySelectorAll('.draggable-stage');
const automateDropZone = document.getElementById('automate-drop-zone');
const actionsThread = document.getElementById('actions-thread');
const arrowCanvas = document.getElementById('arrow-canvas');

function initializePivot() {
  if (!actionsThread) return;
  if (!actionsThread.querySelector('.pivot-spine')) {
    const spine = document.createElement('div');
    spine.className = 'pivot-spine';
    actionsThread.appendChild(spine);
  }
  if (!actionsThread.querySelector('.pivot-node')) {
    const node = document.createElement('div');
    node.className = 'pivot-node';
    node.innerHTML = `
      <div class="pivot-node-circle"></div>
      <span class="pivot-node-label">Pipeline Pivot</span>
    `;
    actionsThread.appendChild(node);
  }
}

function initializeArrowCanvas() {
  if (!arrowCanvas) return;
  arrowCanvas.innerHTML = `
    <defs>
      <marker id="arrowhead" markerWidth="10" markerHeight="7" 
      refX="8" refY="3.5" orient="auto">
        <polygon points="0 0, 10 3.5, 0 7" fill="#1a5276" />
      </marker>
    </defs>
  `;
}

function setupCardDot(card) {
  let dot = card.querySelector('.card-output-dot');
  if (!dot) {
    dot = document.createElement('div');
    dot.className = 'card-output-dot';
    card.appendChild(dot);
    
    dot.addEventListener('mousedown', (e) => {
      e.preventDefault();
      e.stopPropagation();
      
      const threadRect = actionsThread.getBoundingClientRect();
      const dotRect = dot.getBoundingClientRect();
      const startX = (dotRect.left + dotRect.width / 2) - threadRect.left;
      const startY = (dotRect.top + dotRect.height / 2) - threadRect.top;
      
      const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
      path.setAttribute('stroke', '#1a5276');
      path.setAttribute('stroke-width', '2');
      path.setAttribute('fill', 'none');
      path.setAttribute('marker-end', 'url(#arrowhead)');
      path.setAttribute('stroke-dasharray', '4');
      arrowCanvas.appendChild(path);
      
      const onMouseMove = (moveEvent) => {
        const curX = moveEvent.clientX - threadRect.left;
        const curY = moveEvent.clientY - threadRect.top;
        
        const cp1x = startX + 60;
        const cp1y = startY;
        const cp2x = curX - 60;
        const cp2y = curY;
        const d = `M ${startX} ${startY} C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${curX} ${curY}`;
        path.setAttribute('d', d);
      };
      
      const onMouseUp = (upEvent) => {
        window.removeEventListener('mousemove', onMouseMove);
        window.removeEventListener('mouseup', onMouseUp);
        
        path.removeAttribute('stroke-dasharray');
        
        const releaseX = upEvent.clientX - threadRect.left;
        const releaseY = upEvent.clientY - threadRect.top;
        
        let snapped = false;
        const cards = actionsThread.querySelectorAll('.automation-card');
        for (const targetCard of cards) {
          if (targetCard === card) continue;
          
          const cardRect = targetCard.getBoundingClientRect();
          const targetLeft = cardRect.left - threadRect.left;
          const targetTop = cardRect.top - threadRect.top;
          const targetHeight = cardRect.height;
          
          const snapTargetX = targetLeft;
          const snapTargetY = targetTop + targetHeight / 2;
          
          const dist = Math.hypot(releaseX - snapTargetX, releaseY - snapTargetY);
          if (dist < 100) {
            const cp1x = startX + 60;
            const cp1y = startY;
            const cp2x = snapTargetX - 60;
            const cp2y = snapTargetY;
            const d = `M ${startX} ${startY} C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${snapTargetX} ${snapTargetY}`;
            path.setAttribute('d', d);
            snapped = true;
            break;
          }
        }
        
        if (!snapped) {
          const curX = upEvent.clientX - threadRect.left;
          const curY = upEvent.clientY - threadRect.top;
          const totalDragDist = Math.hypot(curX - startX, curY - startY);
          if (totalDragDist < 15) {
            path.remove();
          }
        }
      };
      
      window.addEventListener('mousemove', onMouseMove);
      window.addEventListener('mouseup', onMouseUp);
    });
  }
}

// Run initializations
initializePivot();
initializeArrowCanvas();

draggableStages.forEach(stage => {
  stage.addEventListener('dragstart', (e) => {
    e.dataTransfer.setData('text/plain', stage.dataset.stage);
    stage.classList.add('dragging');
  });

  stage.addEventListener('dragend', () => {
    stage.classList.remove('dragging');
  });
});

if (automateDropZone) {
  automateDropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    automateDropZone.classList.add('dragover');
  });

  automateDropZone.addEventListener('dragenter', (e) => {
    e.preventDefault();
    automateDropZone.classList.add('dragover');
  });

  automateDropZone.addEventListener('dragleave', () => {
    automateDropZone.classList.remove('dragover');
  });

  automateDropZone.addEventListener('drop', async (e) => {
    e.preventDefault();
    automateDropZone.classList.remove('dragover');

    const stageId = e.dataTransfer.getData('text/plain');
    if (!stageId) return;

    let stageTitle = '';
    let endpoint = '';
    if (stageId === '1') {
      stageTitle = 'Asset Discovery';
      endpoint = '/automate/discover';
    } else if (stageId === '2a') {
      stageTitle = 'Account Balances Verification';
      endpoint = '/automate/retrieve/balances';
    } else if (stageId === '2b') {
      stageTitle = 'Property Deed Lookup';
      endpoint = '/automate/retrieve/deeds';
    } else if (stageId === '2c') {
      stageTitle = 'Vital Records Request';
      endpoint = '/automate/retrieve/records';
    } else if (stageId === '3a') {
      stageTitle = 'Petition for Probate Drafting';
      endpoint = '/automate/actions/probate';
    } else if (stageId === '3b') {
      stageTitle = 'Asset Transfer Notice';
      endpoint = '/automate/actions/transfer';
    } else if (stageId === '3c') {
      stageTitle = 'Notice to Creditors';
      endpoint = '/automate/actions/creditors';
    } else {
      return;
    }

    const card = document.createElement('div');
    card.className = 'automation-card running';
    
    const mode = toggleBtn.dataset.mode || 'tokenized';

    card.innerHTML = `
      <h4>
        <span>${stageTitle}</span>
        <span class="status-badge">Running</span>
      </h4>
      <div class="result-content">
        <div class="spinner" style="width:16px; height:16px; border-width:3px; margin:0 8px 0 0;"></div>
        <span>Executing LLM query...</span>
      </div>
    `;
    setupCardDot(card);

    actionsThread.appendChild(card);
    actionsThread.scrollTop = actionsThread.scrollHeight;

    const currentSessionId = sessionId || window.lastUploadSessionId;

    if (!currentSessionId) {
      card.className = 'automation-card error';
      card.innerHTML = `
        <h4>
          <span>${stageTitle}</span>
          <span class="status-badge">Error</span>
        </h4>
        <div class="result-content">
          <p style="margin: 0; color: #d9534f; font-weight: 500;">No active document session. Please upload a PDF or load a sample on the Scan tab first.</p>
        </div>
      `;
      setupCardDot(card);
      return;
    }

    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: currentSessionId })
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Automation stage failed');
      }

      card.className = 'automation-card success';

      const formatResult = (text) => {
        if (!text) return '';
        const lines = text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
        let html = '';
        let inList = false;
        for (const line of lines) {
          if (line.startsWith('-') || line.startsWith('*') || /^\d+\./.test(line)) {
            if (!inList) {
              html += '<ul>';
              inList = true;
            }
            const cleanLine = line.replace(/^[-*\d\.]+\s*/, '');
            html += `<li>${escapeHtml(cleanLine)}</li>`;
          } else {
            if (inList) {
              html += '</ul>';
              inList = false;
            }
            html += `<p>${escapeHtml(line)}</p>`;
          }
        }
        if (inList) {
          html += '</ul>';
        }
        return html || escapeHtml(text);
      };

      const formatResultTokenized = (text) => {
        if (!text) return '';
        const lines = text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
        let html = '';
        let inList = false;
        for (const line of lines) {
          if (line.startsWith('-') || line.startsWith('*') || /^\d+\./.test(line)) {
            if (!inList) {
              html += '<ul>';
              inList = true;
            }
            const cleanLine = line.replace(/^[-*\d\.]+\s*/, '');
            html += `<li>${highlightTokens(cleanLine)}</li>`;
          } else {
            if (inList) {
              html += '</ul>';
              inList = false;
            }
            html += `<p>${highlightTokens(line)}</p>`;
          }
        }
        if (inList) {
          html += '</ul>';
        }
        return html || highlightTokens(text);
      };

      const formattedAnswer = formatResult(data.answer);
      const formattedTokenized = formatResultTokenized(data.tokenized_answer);

      card.innerHTML = `
        <h4>
          <span>${stageTitle}</span>
          <span class="status-badge">Success</span>
        </h4>
        <div class="result-content">
          <div class="tokenized ${mode === 'detokenized' ? 'hidden' : ''}">${formattedTokenized}</div>
          <div class="detokenized ${mode === 'tokenized' ? 'hidden' : ''}">${formattedAnswer}</div>
          <div style="margin-top: 12px; display: flex;">
            <button class="btn-generate-doc" data-stage="${stageId}">
              📄 Generate Document Draft
            </button>
          </div>
        </div>
      `;
      setupCardDot(card);

      const genBtn = card.querySelector('.btn-generate-doc');
      if (genBtn) {
        genBtn.addEventListener('click', async () => {
          const originalText = genBtn.innerHTML;
          genBtn.disabled = true;
          genBtn.innerHTML = `<span class="spinner" style="width:12px; height:12px; border-width:2px; display:inline-block; vertical-align:middle; margin-right:6px; animation: spin 1s linear infinite;"></span> Generating...`;
          try {
            const docRes = await fetch('/automate/generate-document', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ session_id: currentSessionId, stage_id: stageId })
            });
            if (!docRes.ok) {
              const errText = await docRes.text();
              throw new Error(errText || `Server returned status ${docRes.status}`);
            }
            const docData = await docRes.json();
            showDocModal(docData.document_title, docData.tokenized_content, docData.content);
          } catch (err) {
            alert(err.message);
          } finally {
            genBtn.disabled = false;
            genBtn.innerHTML = originalText;
          }
        });
      }
    } catch (err) {
      card.className = 'automation-card error';
      card.innerHTML = `
        <h4>
          <span>${stageTitle}</span>
          <span class="status-badge">Error</span>
        </h4>
        <div class="result-content">
          <p style="margin: 0; color: #d9534f; font-weight: 500;">Failed: ${escapeHtml(err.message)}</p>
        </div>
      `;
      setupCardDot(card);
    }
  });
}

const modelSel = document.getElementById('model-selector');
const mistralTt = document.getElementById('mistral-tooltip');
const closeMistralTt = document.getElementById('close-mistral-tooltip');
let mistralTtDismissed = false;

if (modelSel) {
  modelSel.addEventListener('mouseenter', () => {
    if (!mistralTtDismissed && mistralTt) {
      mistralTt.style.display = 'block';
    }
  });
}

if (closeMistralTt) {
  closeMistralTt.addEventListener('click', (e) => {
    e.stopPropagation();
    mistralTtDismissed = true;
    if (mistralTt) mistralTt.style.display = 'none';
  });
}

const modelSelAutomate = document.getElementById('automate-model-selector');
const mistralTtAutomate = document.getElementById('mistral-tooltip-automate');
const closeMistralTtAutomate = document.getElementById('close-mistral-tooltip-automate');
let mistralTtAutomateDismissed = false;

if (modelSelAutomate) {
  modelSelAutomate.addEventListener('mouseenter', () => {
    if (!mistralTtAutomateDismissed && mistralTtAutomate) {
      mistralTtAutomate.style.display = 'block';
    }
  });
}

if (closeMistralTtAutomate) {
  closeMistralTtAutomate.addEventListener('click', (e) => {
    e.stopPropagation();
    mistralTtAutomateDismissed = true;
    if (mistralTtAutomate) mistralTtAutomate.style.display = 'none';
  });
}

const btnRationale = document.getElementById('btn-rationale');
const modalOverlay = document.getElementById('modal-overlay');
const closeModal = document.getElementById('close-modal');
let onboardingShown = false;

if (btnRationale && modalOverlay && closeModal) {
  btnRationale.addEventListener('click', () => {
    modalOverlay.style.display = 'flex';
  });

  function handleModalClose() {
    modalOverlay.style.display = 'none';
    if (!onboardingShown) {
      onboardingShown = true;
      if (appOverlay) appOverlay.style.display = 'block';
      if (willSampleTooltip) willSampleTooltip.style.display = 'flex';
    }
  }

  closeModal.addEventListener('click', handleModalClose);
  modalOverlay.addEventListener('click', (e) => {
    if (e.target === modalOverlay) {
      handleModalClose();
    }
  });
}

// Document Preview Modal JS Logic
const docModalOverlay = document.getElementById('doc-modal-overlay');
const docModalTitle = document.getElementById('doc-modal-title');
const docPaperTokenized = document.getElementById('doc-paper-tokenized');
const docPaperDetokenized = document.getElementById('doc-paper-detokenized');
const closeDocModal = document.getElementById('close-doc-modal');
const btnCopyDoc = document.getElementById('btn-copy-doc');
const btnDownloadDoc = document.getElementById('btn-download-doc');

let currentDocTokenized = '';
let currentDocDetokenized = '';
let currentDocTitle = '';

function showDocModal(title, tokenized, detokenized) {
  if (!docModalOverlay || !docModalTitle || !docPaperTokenized || !docPaperDetokenized) return;
  currentDocTitle = title;
  currentDocTokenized = tokenized;
  currentDocDetokenized = detokenized;
  
  docModalTitle.textContent = title;
  docPaperTokenized.innerHTML = highlightTokens(tokenized);
  docPaperDetokenized.textContent = detokenized;
  
  const mode = toggleBtn.dataset.mode || 'tokenized';
  docPaperTokenized.classList.toggle('hidden', mode === 'detokenized');
  docPaperDetokenized.classList.toggle('hidden', mode === 'tokenized');
  
  docModalOverlay.style.display = 'flex';
}

if (closeDocModal && docModalOverlay) {
  closeDocModal.addEventListener('click', () => {
    docModalOverlay.style.display = 'none';
  });
  docModalOverlay.addEventListener('click', (e) => {
    if (e.target === docModalOverlay) {
      docModalOverlay.style.display = 'none';
    }
  });
}

if (btnCopyDoc) {
  btnCopyDoc.addEventListener('click', () => {
    const mode = toggleBtn.dataset.mode || 'tokenized';
    const textToCopy = mode === 'tokenized' ? currentDocTokenized : currentDocDetokenized;
    navigator.clipboard.writeText(textToCopy).then(() => {
      const originalText = btnCopyDoc.innerHTML;
      btnCopyDoc.innerHTML = `<svg style="width:16px; height:16px; fill:currentColor;" viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg> <span>Copied!</span>`;
      setTimeout(() => {
        btnCopyDoc.innerHTML = originalText;
      }, 2000);
    });
  });
}

if (btnDownloadDoc) {
  btnDownloadDoc.addEventListener('click', () => {
    const mode = toggleBtn.dataset.mode || 'tokenized';
    const textToDownload = mode === 'tokenized' ? currentDocTokenized : currentDocDetokenized;
    const blob = new Blob([textToDownload], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${currentDocTitle.toLowerCase().replace(/[^a-z0-9]+/g, '_')}_${mode}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  });
}

