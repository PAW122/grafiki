package main

const pageTemplate = `<!DOCTYPE html>
<html lang="pl">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Galeria zdjec</title>
  <style>
    :root {
      color-scheme: light dark;
      font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
    }
    body {
      margin: 0;
      background: #f3f4f8;
      color: #1f2933;
    }
    .topbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 1rem 2rem;
      background: linear-gradient(135deg, #3a7bd5, #00d2ff);
      color: #fff;
      box-shadow: 0 8px 24px rgba(58, 123, 213, 0.35);
    }
    .brand {
      font-weight: 600;
      font-size: 1.25rem;
      letter-spacing: 0.04em;
    }
    .top-actions {
      display: flex;
      gap: 0.75rem;
    }
    .btn {
      border: none;
      border-radius: 999px;
      padding: 0.55rem 1.25rem;
      font-size: 0.95rem;
      font-weight: 500;
      cursor: pointer;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .btn-primary {
      background: rgba(255, 255, 255, 0.16);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.35);
    }
    .btn-secondary {
      background: rgba(15, 23, 42, 0.12);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.25);
    }
    .btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 30px rgba(15, 23, 42, 0.25);
    }
    .hidden-input {
      display: none;
    }
    .page {
      padding: 2rem clamp(1rem, 4vw, 3rem);
      max-width: 1200px;
      margin: 0 auto;
    }
    .info-panel {
      background: #ffffff;
      border-radius: 18px;
      padding: 1.5rem;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.08);
      margin-bottom: 1.5rem;
    }
    .info-panel p {
      margin: 0 0 0.4rem;
      color: #52606d;
      font-size: 0.95rem;
    }
    .info-panel code {
      background: rgba(82, 96, 109, 0.08);
      padding: 0.25rem 0.45rem;
      border-radius: 6px;
      font-family: 'JetBrains Mono', 'SFMono-Regular', ui-monospace, monospace;
      font-size: 0.85rem;
    }
    .upload-panel {
      display: grid;
      gap: 1rem;
      margin-top: 1.25rem;
    }
    .upload-panel form {
      display: flex;
      flex-wrap: wrap;
      gap: 0.75rem;
      align-items: center;
    }
    .upload-panel input[type="file"],
    .upload-panel input[type="text"] {
      flex: 1 1 240px;
      max-width: 320px;
      padding: 0.65rem 0.8rem;
      border-radius: 10px;
      border: 1px solid rgba(82, 96, 109, 0.2);
      font-size: 0.95rem;
    }
    .upload-panel .submit-btn {
      flex: 0 0 auto;
      background: #2563eb;
      color: #fff;
      padding: 0.65rem 1.4rem;
      border-radius: 999px;
      border: none;
      cursor: pointer;
      font-weight: 500;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .upload-panel .submit-btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 30px rgba(37, 99, 235, 0.3);
    }
    .gallery {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
      gap: 1.25rem;
    }
    .tile {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }
    .thumb {
      position: relative;
      border: none;
      border-radius: 16px;
      overflow: hidden;
      cursor: zoom-in;
      padding: 0;
      background: #131722;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.25);
      transition: transform 0.2s ease, box-shadow 0.2s ease;
      min-height: 200px;
      display: block;
    }
    .thumb:hover {
      transform: translateY(-4px);
      box-shadow: 0 18px 55px rgba(15, 23, 42, 0.28);
    }
    .thumb img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
      transition: transform 0.25s ease;
    }
    .tile-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 0.75rem;
      font-size: 0.9rem;
      color: #364152;
    }
    .filename {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .delete-btn {
      border: none;
      border-radius: 8px;
      padding: 0.4rem 0.75rem;
      background: rgba(239, 68, 68, 0.16);
      color: #dc2626;
      font-weight: 500;
      cursor: pointer;
      transition: background 0.18s ease, transform 0.18s ease;
    }
    .delete-btn:hover {
      background: rgba(239, 68, 68, 0.28);
      transform: translateY(-2px);
    }
    .empty {
      text-align: center;
      font-size: 1.1rem;
      color: #52606d;
      padding: 3rem 0;
    }
    .fullscreen-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.92);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1000;
      padding: 2rem;
    }
    .fullscreen-backdrop.active {
      display: flex;
    }
    .fullscreen-backdrop img {
      max-width: 95vw;
      max-height: 95vh;
      border-radius: 20px;
      box-shadow: 0 20px 45px rgba(15, 23, 42, 0.55);
      transition: transform 0.2s ease;
      transform-origin: center;
    }
    .fullscreen-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 1.5rem;
      width: min(900px, 95vw);
    }
    .zoom-controls {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      width: min(420px, 90vw);
      padding: 0.65rem 1rem;
      border-radius: 999px;
      background: rgba(12, 18, 31, 0.75);
      box-shadow: 0 15px 40px rgba(0, 0, 0, 0.35);
      color: #e2e8f0;
      position: fixed;
      bottom: 2rem;
      left: 50%;
      transform: translateX(-50%);
      z-index: 1105;
      backdrop-filter: blur(12px);
      border: 1px solid rgba(148, 163, 184, 0.35);
    }
    .zoom-controls label {
      font-size: 0.8rem;
      text-transform: uppercase;
      letter-spacing: 0.04em;
      color: #cbd5f5;
      white-space: nowrap;
    }
    .zoom-controls input[type="range"] {
      flex: 1;
      accent-color: #38bdf8;
      cursor: pointer;
    }
    .zoom-value {
      font-variant-numeric: tabular-nums;
      min-width: 3ch;
      text-align: right;
    }
    .modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.6);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1100;
      padding: 1.5rem;
    }
    .modal-backdrop.active {
      display: flex;
    }
    .modal {
      background: #ffffff;
      border-radius: 20px;
      padding: 2rem;
      width: min(360px, 90vw);
      box-shadow: 0 20px 50px rgba(15, 23, 42, 0.35);
      display: grid;
      gap: 1rem;
    }
    .modal h2 {
      margin: 0;
      font-size: 1.25rem;
      color: #111827;
      text-align: center;
    }
    .modal label {
      display: grid;
      gap: 0.35rem;
      font-size: 0.9rem;
      color: #364152;
    }
    .modal input {
      padding: 0.65rem 0.8rem;
      border-radius: 10px;
      border: 1px solid rgba(82, 96, 109, 0.2);
      font-size: 0.95rem;
      background: #f8fafc;
      color: #111827;
      transition: border 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    }
    .modal input:focus {
      outline: none;
      border-color: rgba(37, 99, 235, 0.6);
      box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.2);
      background: #ffffff;
    }
    .modal-actions {
      display: flex;
      gap: 0.75rem;
      justify-content: center;
    }
    .modal .primary {
      flex: 1;
      background: #2563eb;
      color: #fff;
      border: none;
      border-radius: 12px;
      padding: 0.65rem;
      font-weight: 600;
      cursor: pointer;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .modal .primary:hover {
      transform: translateY(-2px);
      box-shadow: 0 12px 32px rgba(37, 99, 235, 0.35);
    }
    .modal .ghost {
      flex: 1;
      background: #f8fafc;
      border: 1px solid rgba(148, 163, 184, 0.6);
      border-radius: 12px;
      padding: 0.65rem;
      font-weight: 500;
      cursor: pointer;
      color: #364152;
      transition: transform 0.18s ease, box-shadow 0.18s ease, background 0.18s ease;
    }
    .modal .ghost:hover {
      transform: translateY(-2px);
      box-shadow: 0 12px 32px rgba(15, 23, 42, 0.12);
      background: #eef2f8;
    }
    .toast {
      position: fixed;
      bottom: 2rem;
      right: 2rem;
      background: #111827;
      color: #fff;
      padding: 0.85rem 1.25rem;
      border-radius: 12px;
      box-shadow: 0 15px 40px rgba(17, 24, 39, 0.35);
      opacity: 0;
      transform: translateY(20px);
      pointer-events: none;
      transition: opacity 0.2s ease, transform 0.2s ease;
      z-index: 1200;
      font-size: 0.95rem;
    }
    .toast.visible {
      opacity: 1;
      transform: translateY(0);
    }
    .toast[data-type="error"] {
      background: #dc2626;
      box-shadow: 0 15px 40px rgba(220, 38, 38, 0.35);
    }
    @media (max-width: 720px) {
      .topbar {
        flex-direction: column;
        gap: 1rem;
        text-align: center;
      }
      .upload-panel form {
        flex-direction: column;
        align-items: stretch;
      }
      .toast {
        left: 1rem;
        right: 1rem;
      }
      .zoom-controls {
        flex-direction: column;
        align-items: stretch;
        border-radius: 16px;
        gap: 0.5rem;
        width: calc(100% - 2rem);
        bottom: 1rem;
        padding: 0.75rem 1rem;
      }
    }
  </style>
</head>
<body>
  <header class="topbar">
    <div class="brand">Galeria zdjec</div>
    <div class="top-actions">
      {{if .LoggedIn}}
      <label for="quickUploadInput" class="btn btn-primary" id="quickUploadTrigger">Dodaj zdjecie</label>
      <input type="file" id="quickUploadInput" class="hidden-input" accept=".jpg,.jpeg,.png,.gif,.bmp,.svg,.webp,.avif">
      <button id="logoutButton" class="btn btn-secondary" type="button">Wyloguj</button>
      {{else}}
      <button id="loginButton" class="btn btn-primary" type="button">Zaloguj</button>
      {{end}}
    </div>
  </header>
  <main class="page">
    {{if .Images}}
    <section class="gallery">
      {{range .Images}}
      <div class="tile" data-name="{{.Name}}">
        <button type="button" class="thumb" data-src="{{.URL}}" aria-label="Zobacz {{.Name}}">
          <img src="{{.URL}}" alt="{{.Name}}">
        </button>
        <div class="tile-meta">
          <span class="filename" title="{{.Name}}">{{.Name}}</span>
          {{if $.LoggedIn}}
          <button type="button" class="delete-btn" data-name="{{.Name}}">Usun</button>
          {{end}}
        </div>
      </div>
      {{end}}
    </section>
    {{else}}
    <p class="empty">Brak obrazow w katalogu.</p>
    {{end}}
  </main>

  <div class="fullscreen-backdrop" id="backdrop" role="dialog" aria-modal="true">
    <div class="fullscreen-content">
      <img id="fullImage" alt="">
      <div class="zoom-controls" id="zoomControls" hidden>
        <label for="zoomSlider">Powiekszenie</label>
        <input type="range" id="zoomSlider" min="100" max="250" step="10" value="100">
        <span class="zoom-value" id="zoomValue">100%</span>
      </div>
    </div>
  </div>

  <div class="modal-backdrop" id="loginModal">
    <form class="modal" id="loginForm" autocomplete="off">
      <h2>Panel administratora</h2>
      <label>
        Login
        <input type="text" name="username" autocomplete="off" autocapitalize="none" spellcheck="false" required>
      </label>
      <label>
        Haslo
        <input type="password" name="password" autocomplete="off" required>
      </label>
      <div class="modal-actions">
        <button class="primary" type="submit">Zaloguj</button>
        <button class="ghost" type="button" id="loginCancel">Anuluj</button>
      </div>
    </form>
  </div>

  <div class="toast" id="statusMessage" role="status" aria-live="polite"></div>

  <script>
    const backdrop = document.getElementById('backdrop');
    const fullImage = document.getElementById('fullImage');
    const loginModal = document.getElementById('loginModal');
    const loginButton = document.getElementById('loginButton');
    const logoutButton = document.getElementById('logoutButton');
    const loginForm = document.getElementById('loginForm');
    const loginCancel = document.getElementById('loginCancel');
    const uploadForm = document.getElementById('uploadForm');
    const quickUploadInput = document.getElementById('quickUploadInput');
    const messageEl = document.getElementById('statusMessage');
    const zoomSlider = document.getElementById('zoomSlider');
    const zoomControls = document.getElementById('zoomControls');
    const zoomValue = document.getElementById('zoomValue');
    let hideToast;

    function showMessage(text, type = 'info') {
      if (!messageEl) return;
      messageEl.textContent = text;
      messageEl.dataset.type = type;
      messageEl.classList.add('visible');
      clearTimeout(hideToast);
      hideToast = setTimeout(() => {
        messageEl.classList.remove('visible');
      }, 4000);
    }

    function setZoom(value) {
      if (!fullImage) return;
      const scale = value / 100;
      fullImage.style.transform = 'scale(' + scale + ')';
      if (zoomValue) {
        zoomValue.textContent = value + '%';
      }
    }

    function resetZoom() {
      if (zoomSlider) {
        zoomSlider.value = '100';
      }
      setZoom(100);
    }

    function openFullscreen(src, alt) {
      fullImage.src = src;
      fullImage.alt = alt;
      resetZoom();
      if (zoomControls) {
        zoomControls.hidden = false;
      }
      backdrop.classList.add('active');
    }

    function closeFullscreen() {
      backdrop.classList.remove('active');
      if (zoomControls) {
        zoomControls.hidden = true;
      }
      resetZoom();
      fullImage.src = '';
      fullImage.alt = '';
    }

    document.querySelectorAll('.thumb').forEach(btn => {
      btn.addEventListener('click', () => {
        const src = btn.dataset.src;
        const alt = btn.closest('.tile')?.dataset.name || '';
        if (backdrop.classList.contains('active') && fullImage.src.endsWith(src)) {
          closeFullscreen();
        } else {
          openFullscreen(src, alt);
        }
      });
    });

    backdrop.addEventListener('click', closeFullscreen);
    document.addEventListener('keydown', event => {
      if (event.key === 'Escape') {
        if (backdrop.classList.contains('active')) {
          closeFullscreen();
        }
        if (loginModal.classList.contains('active')) {
          loginModal.classList.remove('active');
        }
      }
    });

    if (loginButton) {
      loginButton.addEventListener('click', () => {
        loginModal.classList.add('active');
        loginForm.reset();
        loginForm.querySelector('input[name="username"]').focus();
      });
    }

    if (loginCancel) {
      loginCancel.addEventListener('click', () => {
        loginModal.classList.remove('active');
      });
    }

    if (loginModal) {
      loginModal.addEventListener('click', event => {
        if (event.target === loginModal) {
          loginModal.classList.remove('active');
        }
      });
    }

    async function fetchJSON(url, options = {}) {
      const response = await fetch(url, options);
      const text = await response.text();
      let data;
      try {
        data = text ? JSON.parse(text) : {};
      } catch (_) {
        data = {};
      }
      if (!response.ok) {
        const message = data.error || text || 'Wystapil blad';
        throw new Error(message);
      }
      return data;
    }

    if (loginForm) {
      loginForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(loginForm);
        const payload = {
          username: formData.get('username'),
          password: formData.get('password')
        };
        try {
          await fetchJSON('/api/login', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (logoutButton) {
      logoutButton.addEventListener('click', async () => {
        try {
          await fetchJSON('/api/logout', { method: 'POST' });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (uploadForm) {
      uploadForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(uploadForm);
        try {
          await fetchJSON('/api/upload', {
            method: 'POST',
            body: formData
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    if (quickUploadInput) {
      quickUploadInput.addEventListener('change', async () => {
        if (!quickUploadInput.files.length) {
          return;
        }
        const formData = new FormData();
        formData.append('file', quickUploadInput.files[0]);
        try {
          await fetchJSON('/api/upload', {
            method: 'POST',
            body: formData
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        } finally {
          quickUploadInput.value = '';
        }
      });
    }

    if (zoomSlider) {
      zoomSlider.addEventListener('input', () => {
        const value = Number(zoomSlider.value) || 100;
        setZoom(value);
      });
    }

    if (zoomControls) {
      ['click', 'mousedown', 'pointerdown', 'touchstart'].forEach(type => {
        zoomControls.addEventListener(type, event => {
          event.stopPropagation();
        });
      });
    }

    document.querySelectorAll('.delete-btn').forEach(btn => {
      btn.addEventListener('click', async event => {
        event.stopPropagation();
        const name = btn.dataset.name;
        if (!name) return;
        const confirmed = confirm('Czy na pewno chcesz usunac plik "' + name + '"?');
        if (!confirmed) return;
        try {
          await fetchJSON('/api/delete', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name})
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    });
  </script>
</body>
</html>
`
