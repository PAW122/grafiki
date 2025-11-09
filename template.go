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
    * {
      box-sizing: border-box;
    }
    body {
      margin: 0;
      background: #f3f4f8;
      color: #1f2933;
    }
    .topbar {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      padding: 1.25rem clamp(1rem, 4vw, 3rem);
      background: linear-gradient(135deg, #3a7bd5, #00d2ff);
      color: #fff;
      box-shadow: 0 12px 30px rgba(58, 123, 213, 0.35);
    }
    .brand {
      font-weight: 700;
      font-size: 1.4rem;
      letter-spacing: 0.04em;
    }
    .top-actions {
      display: flex;
      gap: 0.75rem;
      align-items: center;
      flex-wrap: wrap;
    }
    .btn {
      border: none;
      border-radius: 999px;
      padding: 0.6rem 1.4rem;
      font-size: 0.95rem;
      font-weight: 500;
      cursor: pointer;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .btn:disabled,
    .btn[aria-disabled="true"] {
      opacity: 0.6;
      cursor: not-allowed;
    }
    .btn-primary {
      background: rgba(255, 255, 255, 0.18);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.4);
    }
    .btn-secondary {
      background: rgba(15, 23, 42, 0.18);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.1);
    }
    .btn:hover:not(:disabled) {
      transform: translateY(-2px);
      box-shadow: 0 12px 30px rgba(15, 23, 42, 0.3);
    }
    .hidden-input {
      display: none;
    }
    .page {
      padding: 2rem clamp(1rem, 4vw, 3rem);
      max-width: 1280px;
      margin: 0 auto 3rem;
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }
    .section-card {
      background: #fff;
      border-radius: 20px;
      padding: 1.5rem;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.09);
    }
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
    }
    .section-header h2 {
      margin: 0;
      font-size: 1.1rem;
    }
    .section-header p {
      margin: 0.25rem 0 0;
      color: #64748b;
      font-size: 0.93rem;
    }
    .folders-panel {
      border: 1px solid rgba(148, 163, 184, 0.2);
    }
    .folders-grid {
      margin-top: 1.25rem;
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
    }
    .folder-card {
      border: 1px solid rgba(148, 163, 184, 0.2);
      border-radius: 16px;
      padding: 1rem;
      background: #f8fafc;
      text-align: left;
      cursor: pointer;
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
      transition: border 0.18s ease, transform 0.18s ease, background 0.18s ease;
    }
    .folder-card:hover {
      transform: translateY(-3px);
      border-color: rgba(37, 99, 235, 0.4);
    }
    .folder-card.active {
      border-color: #2563eb;
      background: #e0ebff;
    }
    .folder-name {
      font-weight: 600;
      font-size: 1rem;
      color: #111827;
    }
    .folder-meta {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 0.85rem;
      color: #475569;
    }
    .badge {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      padding: 0.15rem 0.7rem;
      border-radius: 999px;
      font-size: 0.7rem;
      font-weight: 600;
      letter-spacing: 0.05em;
      text-transform: uppercase;
    }
    .badge.public {
      background: rgba(34, 197, 94, 0.15);
      color: #15803d;
    }
    .badge.shared {
      background: rgba(249, 115, 22, 0.15);
      color: #c2410c;
    }
    .badge.private {
      background: rgba(59, 130, 246, 0.15);
      color: #1d4ed8;
    }
    .workspace {
      border: 1px solid rgba(148, 163, 184, 0.2);
    }
    .workspace-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      flex-wrap: wrap;
    }
    .workspace-header h2 {
      margin: 0;
      font-size: 1.3rem;
    }
    .workspace-subtitle {
      margin: 0.35rem 0 0;
      color: #64748b;
      font-size: 0.95rem;
    }
    .workspace-actions {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 0.5rem;
    }
    .info-panel {
      margin-top: 1rem;
      padding: 1rem;
      border-radius: 14px;
      background: rgba(14, 165, 233, 0.1);
      color: #0369a1;
      border: 1px solid rgba(14, 165, 233, 0.3);
      font-size: 0.92rem;
    }
    .upload-panel {
      margin-top: 1.25rem;
      padding: 1rem;
      border-radius: 16px;
      background: #f8fafc;
      border: 1px solid rgba(148, 163, 184, 0.2);
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
      min-width: 200px;
      padding: 0.65rem 0.85rem;
      border-radius: 10px;
      border: 1px solid rgba(148, 163, 184, 0.5);
      font-size: 0.95rem;
    }
    .upload-panel .submit-btn {
      border: none;
      border-radius: 10px;
      background: #2563eb;
      color: #fff;
      padding: 0.65rem 1.5rem;
      font-weight: 600;
      cursor: pointer;
      transition: box-shadow 0.18s ease, transform 0.18s ease;
    }
    .upload-panel .submit-btn:hover {
      transform: translateY(-2px);
      box-shadow: 0 10px 25px rgba(37, 99, 235, 0.35);
    }
    .gallery {
      margin-top: 1.5rem;
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
      gap: 1.25rem;
    }
    .tile {
      display: flex;
      flex-direction: column;
      gap: 0.65rem;
    }
    .thumb {
      position: relative;
      border: none;
      border-radius: 16px;
      overflow: hidden;
      cursor: zoom-in;
      padding: 0;
      background: #0f172a;
      box-shadow: 0 12px 40px rgba(15, 23, 42, 0.25);
      min-height: 200px;
      display: block;
      transition: transform 0.2s ease, box-shadow 0.2s ease;
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
      padding: 0.35rem 0.8rem;
      background: rgba(239, 68, 68, 0.18);
      color: #b91c1c;
      font-weight: 600;
      cursor: pointer;
      transition: background 0.18s ease, transform 0.18s ease;
    }
    .delete-btn:hover {
      background: rgba(239, 68, 68, 0.28);
      transform: translateY(-2px);
    }
    .empty,
    .empty-state {
      text-align: center;
      color: #64748b;
      padding: 2.5rem 1rem;
      font-size: 1rem;
    }
    .empty-state.large {
      font-size: 1.1rem;
      padding: 3rem 1rem;
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
    .fullscreen-content {
      width: min(90vw, 1100px);
      min-width: min(80vw, 1100px);
      height: min(90vh, 900px);
      min-height: min(80vh, 900px);
      display: flex;
      flex-direction: column;
      gap: 1rem;
      align-items: center;
      justify-content: center;
      overflow: hidden;
      border: 2px solid rgba(255, 255, 255, 0.25);
      border-radius: 24px;
      background: rgba(15, 23, 42, 0.85);
      box-shadow: 0 30px 80px rgba(0, 0, 0, 0.55);
    }
    .fullscreen-content img {
      max-width: 100%;
      max-height: calc(100% - 2.5rem);
      object-fit: contain;
      border-radius: 18px;
      box-shadow: 0 25px 70px rgba(0, 0, 0, 0.65);
      transition: transform 0.15s ease;
      cursor: grab;
      touch-action: none;
      will-change: transform;
    }
    .fullscreen-content img.dragging {
      cursor: grabbing;
    }
    .zoom-controls {
      position: fixed;
      bottom: 1.25rem;
      left: 50%;
      transform: translateX(-50%);
      z-index: 1101;
      min-width: min(420px, 90vw);
      display: flex;
      align-items: center;
      gap: 0.75rem;
      background: rgba(15, 23, 42, 0.85);
      border-radius: 999px;
      padding: 0.5rem 1.25rem;
      color: #fff;
      font-size: 0.9rem;
      box-shadow: 0 12px 30px rgba(15, 23, 42, 0.45);
      backdrop-filter: blur(12px);
    }
    .zoom-controls input[type="range"] {
      flex: 1;
    }
    .modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(15, 23, 42, 0.5);
      display: none;
      align-items: center;
      justify-content: center;
      z-index: 1001;
      padding: 1rem;
    }
    .modal-backdrop.active {
      display: flex;
    }
    .modal {
      background: #fff;
      border-radius: 24px;
      padding: 1.75rem 2rem;
      width: min(460px, 92vw);
      max-height: min(90vh, 720px);
      margin: 0.5rem;
      display: flex;
      flex-direction: column;
      gap: 1.1rem;
      box-shadow: 0 30px 70px rgba(15, 23, 42, 0.28);
      overflow-y: auto;
      scrollbar-width: thin;
    }
    .modal.modal-large {
      width: min(640px, 94vw);
      max-height: min(92vh, 760px);
      padding: 2rem 2.35rem;
      gap: 1.35rem;
    }
    .modal h2 {
      margin: 0;
      font-size: 1.35rem;
      color: #0f172a;
    }
    .modal-subtitle {
      margin: 0;
      font-size: 0.95rem;
      color: #475569;
    }
    .modal-section {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .section-label {
      font-size: 0.95rem;
      font-weight: 600;
      color: #0f172a;
    }
    .visibility-options {
      display: grid;
      gap: 0.75rem;
      margin-top: 0.35rem;
    }
    .radio-option {
      display: flex;
      gap: 0.75rem;
      padding: 0.65rem 0.75rem;
      border: 1px solid rgba(148, 163, 184, 0.6);
      border-radius: 14px;
      align-items: flex-start;
      background: rgba(248, 250, 252, 0.75);
    }
    .radio-option input {
      margin-top: 0.2rem;
    }
    .radio-description {
      display: flex;
      flex-direction: column;
      gap: 0.15rem;
      font-size: 0.88rem;
      color: #475569;
    }
    .radio-description strong {
      font-size: 0.95rem;
      color: #0f172a;
    }
    .modal label {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
      font-size: 0.9rem;
      color: #475569;
    }
    .modal input,
    .modal textarea {
      border-radius: 12px;
      border: 1px solid rgba(148, 163, 184, 0.5);
      padding: 0.6rem 0.8rem;
      font-size: 0.95rem;
    }
    .modal-actions {
      display: flex;
      gap: 0.75rem;
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
    .modal .ghost {
      flex: 1;
      background: transparent;
      border: 1px solid rgba(148, 163, 184, 0.7);
      border-radius: 12px;
      padding: 0.65rem;
      font-weight: 500;
      cursor: pointer;
      color: #475569;
      transition: transform 0.18s ease, box-shadow 0.18s ease;
    }
    .share-details {
      border: 1px dashed rgba(148, 163, 184, 0.7);
      border-radius: 12px;
      padding: 0.75rem;
      background: #f8fafc;
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    .share-details strong {
      font-size: 0.9rem;
      color: #0f172a;
    }
    .share-link-row {
      display: flex;
      gap: 0.5rem;
      align-items: center;
    }
    .share-link-row code {
      flex: 1;
      background: rgba(15, 23, 42, 0.08);
      padding: 0.35rem 0.6rem;
      border-radius: 8px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
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
      transform: translateY(15px);
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
      .section-header {
        flex-direction: column;
        align-items: flex-start;
      }
      .topbar {
        flex-direction: column;
        align-items: flex-start;
      }
      .workspace-actions {
        width: 100%;
        justify-content: flex-start;
      }
      .share-link-row {
        flex-direction: column;
        align-items: stretch;
      }
      .modal,
      .modal.modal-large {
        width: calc(100% - 1rem);
        margin: 0.5rem;
        max-height: 90vh;
        padding: 1.5rem 1.25rem;
      }
    }
  </style>
</head>
<body data-shared-mode="{{if .SharedMode}}true{{else}}false{{end}}" data-active-folder="{{if .ActiveFolder}}{{.ActiveFolder.Slug}}{{end}}" data-active-folder-id="{{if .ActiveFolder}}{{.ActiveFolder.ID}}{{end}}" data-active-folder-visibility="{{if .ActiveFolder}}{{.ActiveFolder.Visibility}}{{end}}" data-active-folder-share-token="{{if .ActiveFolder}}{{.ActiveFolder.SharedToken}}{{end}}" data-active-folder-share-url="{{if .ActiveFolder}}{{.ActiveFolder.ShareURL}}{{end}}" data-active-folder-share-views="{{if .ActiveFolder}}{{.ActiveFolder.SharedViews}}{{end}}">
  <header class="topbar">
    <div class="brand">Galeria zdjec</div>
    <div class="top-actions">
      {{if and .AllowFolderManagement .ActiveFolder (not .SharedMode)}}
      <label for="quickUploadInput" class="btn btn-primary" id="quickUploadTrigger">Szybkie dodawanie</label>
      <input type="file" id="quickUploadInput" class="hidden-input" accept=".jpg,.jpeg,.png,.gif,.bmp,.svg,.webp,.avif">
      {{end}}
      {{if .LoggedIn}}
      <button id="logoutButton" class="btn btn-secondary" type="button">Wyloguj</button>
      {{else}}
      <button id="loginButton" class="btn btn-primary" type="button">Zaloguj</button>
      {{end}}
    </div>
  </header>
  <main class="page">
    {{if not .SharedMode}}
    <section class="section-card folders-panel">
      <div class="section-header">
        <div>
          <h2>Foldery</h2>
          <p>Wybierz katalog, aby zobaczyc jego pliki.</p>
        </div>
        {{if .AllowFolderManagement}}
        <button class="btn btn-secondary" type="button" id="newFolderButton">Nowy folder</button>
        {{end}}
      </div>
      {{if .Folders}}
      <div class="folders-grid">
        {{range .Folders}}
        <button type="button" class="folder-card {{if and $.ActiveFolder (eq $.ActiveFolder.Slug .Slug)}}active{{end}}" data-slug="{{.Slug}}">
          <div class="folder-name">{{.Name}}</div>
          <div class="folder-meta">
            <span class="badge {{.Visibility}}">
              {{if eq .Visibility "public"}}Publiczny{{else if eq .Visibility "shared"}}Udostepniony{{else}}Prywatny{{end}}
            </span>
            {{if eq .Visibility "shared"}}
            <span>{{.SharedViews}} wejsc</span>
            {{end}}
          </div>
        </button>
        {{end}}
      </div>
      {{else}}
      <p class="empty-state">Brak folderow. Zaloguj sie, aby utworzyc pierwszy.</p>
      {{end}}
    </section>
    {{end}}

    <section class="section-card workspace {{if not .ActiveFolder}}workspace-empty{{end}}" id="workspace">
      {{if .ActiveFolder}}
      <div class="workspace-header">
        <div>
          <h2>{{.ActiveFolder.Name}}</h2>
          <p class="workspace-subtitle">
            {{if eq .ActiveFolder.Visibility "public"}}Folder widoczny dla wszystkich uzytkownikow.{{else if eq .ActiveFolder.Visibility "shared"}}Folder udostepniony przez tajny link.{{else}}Folder widoczny tylko po zalogowaniu.{{end}}
          </p>
        </div>
        <div class="workspace-actions">
          <span class="badge {{.ActiveFolder.Visibility}}">
            {{if eq .ActiveFolder.Visibility "public"}}Publiczny{{else if eq .ActiveFolder.Visibility "shared"}}Udostepniony{{else}}Prywatny{{end}}
          </span>
          {{if .SharedMode}}
          <span class="badge shared">Tryb linku</span>
          {{end}}
          {{if and .AllowFolderManagement (not .SharedMode)}}
          <button class="btn btn-secondary" type="button" id="folderSettingsButton">Ustawienia folderu</button>
          {{end}}
        </div>
      </div>

      {{if and .AllowFolderManagement (not .SharedMode)}}
      <div class="upload-panel">
        <form id="uploadForm">
          <input type="hidden" name="folder" value="{{.ActiveFolder.Slug}}">
          <input type="file" name="file" required accept=".jpg,.jpeg,.png,.gif,.bmp,.svg,.webp,.avif">
          <input type="text" name="name" placeholder="Nazwa pliku (opcjonalnie)">
          <button class="submit-btn" type="submit">Przeslij</button>
        </form>
      </div>
      {{else if .SharedMode}}
      <div class="info-panel">Ten folder jest udostepniony tylko do odczytu.</div>
      {{else if not .LoggedIn}}
      <div class="info-panel">Zaloguj sie, aby zarzadzac plikami w tym folderze.</div>
      {{end}}

      {{if .Images}}
      <section class="gallery" data-folder="{{.ActiveFolder.Slug}}">
        {{range .Images}}
        <div class="tile" data-name="{{.Name}}">
          <button type="button" class="thumb" data-src="{{.URL}}" aria-label="Zobacz {{.Name}}">
            <img src="{{.URL}}" alt="{{.Name}}">
          </button>
          <div class="tile-meta">
            <span class="filename" title="{{.Name}}">{{.Name}}</span>
            {{if $.AllowFolderManagement}}
            <button type="button" class="delete-btn" data-name="{{.Name}}" data-folder="{{$.ActiveFolder.Slug}}">Usun</button>
            {{end}}
          </div>
        </div>
        {{end}}
      </section>
      {{else}}
      <p class="empty">Brak obrazow w tym folderze.</p>
      {{end}}
      {{else}}
      <div class="empty-state large">
        {{if .SharedMode}}
        Folder nie jest juz udostepniony lub link wygasl.
        {{else}}
        Wybierz folder z listy powyzej, aby zobaczyc jego zawartosc.
        {{end}}
      </div>
      {{end}}
    </section>
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

  <div class="modal-backdrop" id="newFolderModal">
    <form class="modal" id="newFolderForm" autocomplete="off">
      <h2>Nowy folder</h2>
      <label>
        Nazwa folderu
        <input type="text" name="name" placeholder="np. Zajecia-1" required>
      </label>
      <div class="modal-actions">
        <button class="primary" type="submit">Utworz</button>
        <button class="ghost" type="button" id="newFolderCancel">Anuluj</button>
      </div>
    </form>
  </div>

  <div class="modal-backdrop" id="folderSettingsModal">
    <form class="modal modal-large" id="folderSettingsForm">
      <div class="modal-section">
        <h2>Ustawienia folderu</h2>
        <p class="modal-subtitle">Dostosuj widocznosc oraz linki udostepnione dla tego katalogu.</p>
      </div>
      <div class="modal-section">
        <span class="section-label">Widocznosc</span>
        <div class="visibility-options">
          <label class="radio-option">
            <input type="radio" name="visibility" value="public">
            <span class="radio-description">
              <strong>Publiczny</strong>
              <span>Widoczny dla wszystkich odwiedzajacych strone.</span>
            </span>
          </label>
          <label class="radio-option">
            <input type="radio" name="visibility" value="shared">
            <span class="radio-description">
              <strong>Udostepniony link</strong>
              <span>Dostep tylko przez tajny link, idealne do wspoldzielenia.</span>
            </span>
          </label>
          <label class="radio-option">
            <input type="radio" name="visibility" value="private">
            <span class="radio-description">
              <strong>Prywatny</strong>
              <span>Widoczny tylko po zalogowaniu jako administrator.</span>
            </span>
          </label>
        </div>
      </div>
      <div class="modal-section">
        <div class="share-details" id="shareDetails" hidden>
          <strong>Udostepniony link</strong>
          <div class="share-link-row">
            <code id="shareLinkValue"></code>
            <button type="button" class="ghost" id="copyShareLink">Kopiuj</button>
          </div>
          <div class="share-link-row">
            <span>Wejscia: <strong id="shareViewsValue">0</strong></span>
            <button type="button" class="ghost" id="regenerateLinkButton">Nowy link</button>
            <button type="button" class="ghost" id="downloadQrButton">Pobierz QR</button>
          </div>
        </div>
      </div>
      <div class="modal-actions">
        <button class="primary" type="submit">Zapisz</button>
        <button class="ghost" type="button" id="folderSettingsCancel">Zamknij</button>
      </div>
    </form>
  </div>

  <div class="toast" id="statusMessage" role="status" aria-live="polite"></div>

  <script>
    const state = (() => {
      const dataset = document.body?.dataset || {};
      return {
        sharedMode: dataset.sharedMode === 'true',
        activeFolder: dataset.activeFolder || '',
        activeFolderId: Number(dataset.activeFolderId || 0),
        activeFolderVisibility: dataset.activeFolderVisibility || '',
        activeFolderShareToken: dataset.activeFolderShareToken || '',
        activeFolderShareUrl: dataset.activeFolderShareUrl || '',
        activeFolderShareViews: Number(dataset.activeFolderShareViews || 0)
      };
    })();

    const backdrop = document.getElementById('backdrop');
    const fullImage = document.getElementById('fullImage');
    const loginModal = document.getElementById('loginModal');
    const loginButton = document.getElementById('loginButton');
    const logoutButton = document.getElementById('logoutButton');
    const loginForm = document.getElementById('loginForm');
    const loginCancel = document.getElementById('loginCancel');
    const uploadForm = document.getElementById('uploadForm');
    const quickUploadInput = document.getElementById('quickUploadInput');
    const quickUploadTrigger = document.getElementById('quickUploadTrigger');
    const messageEl = document.getElementById('statusMessage');
    const zoomSlider = document.getElementById('zoomSlider');
    const zoomControls = document.getElementById('zoomControls');
    const zoomValue = document.getElementById('zoomValue');
    const newFolderButton = document.getElementById('newFolderButton');
    const newFolderModal = document.getElementById('newFolderModal');
    const newFolderForm = document.getElementById('newFolderForm');
    const newFolderCancel = document.getElementById('newFolderCancel');
    const folderSettingsButton = document.getElementById('folderSettingsButton');
    const folderSettingsModal = document.getElementById('folderSettingsModal');
    const folderSettingsForm = document.getElementById('folderSettingsForm');
    const folderSettingsCancel = document.getElementById('folderSettingsCancel');
    const shareDetails = document.getElementById('shareDetails');
    const shareLinkValue = document.getElementById('shareLinkValue');
    const shareViewsValue = document.getElementById('shareViewsValue');
    const copyShareLink = document.getElementById('copyShareLink');
    const regenerateLinkButton = document.getElementById('regenerateLinkButton');
    const downloadQrButton = document.getElementById('downloadQrButton');
    let hideToast;

    const zoomState = {
      value: 100,
      scale: 1,
      panX: 0,
      panY: 0,
      dragging: false,
      pointerId: null,
      lastX: 0,
      lastY: 0,
      panFrame: null
    };

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

    function openModal(modal) {
      modal?.classList.add('active');
    }

    function closeModal(modal) {
      modal?.classList.remove('active');
    }

    function applyImageTransform(immediate = false) {
      if (!fullImage) return;
      const apply = () => {
        fullImage.style.transform = 'translate3d(' + zoomState.panX + 'px, ' + zoomState.panY + 'px, 0) scale(' + zoomState.scale + ')';
      };
      if (immediate) {
        if (zoomState.panFrame) {
          cancelAnimationFrame(zoomState.panFrame);
          zoomState.panFrame = null;
        }
        apply();
        return;
      }
      if (zoomState.panFrame) {
        cancelAnimationFrame(zoomState.panFrame);
      }
      zoomState.panFrame = requestAnimationFrame(() => {
        zoomState.panFrame = null;
        apply();
      });
    }

    function setZoom(value) {
      const numeric = Number(value);
      const clamped = Math.min(250, Math.max(100, Number.isFinite(numeric) ? numeric : 100));
      zoomState.value = clamped;
      zoomState.scale = clamped / 100;
      applyImageTransform(true);
      if (zoomSlider && zoomSlider.value !== String(clamped)) {
        zoomSlider.value = String(clamped);
      }
      if (zoomValue) {
        zoomValue.textContent = clamped + '%';
      }
    }

    function resetView() {
      const activePointer = zoomState.pointerId;
      zoomState.pointerId = null;
      zoomState.dragging = false;
      zoomState.panX = 0;
      zoomState.panY = 0;
      if (fullImage) {
        fullImage.classList.remove('dragging');
        fullImage.style.transition = '';
        if (typeof fullImage.releasePointerCapture === 'function' && activePointer !== null) {
          try {
            fullImage.releasePointerCapture(activePointer);
          } catch (_) {}
        }
      }
      if (zoomSlider && zoomSlider.value !== '100') {
        zoomSlider.value = '100';
      }
      setZoom(100);
    }

    function startPan(event) {
      if (!fullImage) return;
      if (event.pointerType === 'mouse' && event.button !== 0) {
        return;
      }
      event.preventDefault();
      zoomState.dragging = true;
      zoomState.pointerId = event.pointerId;
      zoomState.lastX = event.clientX;
      zoomState.lastY = event.clientY;
      fullImage.setPointerCapture?.(event.pointerId);
      fullImage.style.transition = 'none';
      fullImage.classList.add('dragging');
    }

    function movePan(event) {
      if (!zoomState.dragging || event.pointerId !== zoomState.pointerId) {
        return;
      }
      event.preventDefault();
      zoomState.panX += event.clientX - zoomState.lastX;
      zoomState.panY += event.clientY - zoomState.lastY;
      zoomState.lastX = event.clientX;
      zoomState.lastY = event.clientY;
      applyImageTransform();
    }

    function endPan(event) {
      if (!zoomState.dragging || event.pointerId !== zoomState.pointerId) {
        return;
      }
      zoomState.dragging = false;
      zoomState.pointerId = null;
      fullImage?.classList.remove('dragging');
      if (fullImage) {
        fullImage.style.transition = '';
      }
      fullImage?.releasePointerCapture?.(event.pointerId);
    }

    function openFullscreen(src, alt) {
      if (!backdrop || !fullImage) return;
      fullImage.src = src;
      fullImage.alt = alt;
      resetView();
      if (zoomControls) {
        zoomControls.hidden = false;
      }
      backdrop.classList.add('active');
    }

    function closeFullscreen() {
      backdrop?.classList.remove('active');
      if (zoomControls) {
        zoomControls.hidden = true;
      }
      resetView();
      if (fullImage) {
        fullImage.src = '';
        fullImage.alt = '';
      }
    }

    document.querySelectorAll('.thumb').forEach(btn => {
      btn.addEventListener('click', () => {
        const src = btn.dataset.src;
        const alt = btn.closest('.tile')?.dataset.name || '';
        if (backdrop?.classList.contains('active') && fullImage?.src.endsWith(src)) {
          closeFullscreen();
        } else {
          openFullscreen(src, alt);
        }
      });
    });

    if (fullImage) {
      fullImage.addEventListener('pointerdown', startPan);
      fullImage.addEventListener('pointermove', movePan);
      ['pointerup', 'pointercancel', 'pointerleave'].forEach(type => {
        fullImage.addEventListener(type, endPan);
      });
    }

    backdrop?.addEventListener('click', event => {
      if (event.target === backdrop) {
        closeFullscreen();
      }
    });

    document.addEventListener('keydown', event => {
      if (event.key === 'Escape') {
        closeFullscreen();
        [loginModal, newFolderModal, folderSettingsModal].forEach(closeModal);
      }
    });

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
    if (loginButton) {
      loginButton.addEventListener('click', () => {
        openModal(loginModal);
        loginForm?.reset();
        loginForm?.querySelector('input[name="username"]')?.focus();
      });
    }

    loginCancel?.addEventListener('click', () => closeModal(loginModal));

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

    logoutButton?.addEventListener('click', async () => {
      try {
        await fetchJSON('/api/logout', { method: 'POST' });
        window.location.reload();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });

    document.querySelectorAll('.folder-card').forEach(card => {
      card.addEventListener('click', () => {
        const slug = card.dataset.slug;
        if (!slug) return;
        window.location.href = '/?folder=' + encodeURIComponent(slug);
      });
    });

    if (uploadForm) {
      uploadForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(uploadForm);
        if (!formData.get('folder')) {
          showMessage('Najpierw wybierz folder', 'error');
          return;
        }
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
        if (!state.activeFolder) {
          showMessage('Wybierz aktywny folder przed przeslaniem pliku', 'error');
          quickUploadInput.value = '';
          return;
        }
        const formData = new FormData();
        formData.append('folder', state.activeFolder);
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
    } else if (quickUploadTrigger && !state.activeFolder) {
      quickUploadTrigger.setAttribute('aria-disabled', 'true');
    }

    document.querySelectorAll('.delete-btn').forEach(btn => {
      btn.addEventListener('click', async event => {
        event.stopPropagation();
        const name = btn.dataset.name;
        const folder = btn.dataset.folder || state.activeFolder;
        if (!name || !folder) {
          showMessage('Nie mozna usunac bez folderu', 'error');
          return;
        }
        const confirmed = confirm('Czy na pewno chcesz usunac plik "' + name + '"?');
        if (!confirmed) return;
        try {
          await fetchJSON('/api/delete', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name, folder})
          });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    });

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
    newFolderButton?.addEventListener('click', () => {
      newFolderForm?.reset();
      openModal(newFolderModal);
      newFolderForm?.querySelector('input[name="name"]')?.focus();
    });

    newFolderCancel?.addEventListener('click', () => closeModal(newFolderModal));

    if (newFolderForm) {
      newFolderForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(newFolderForm);
        const payload = { name: formData.get('name') };
        try {
          const folder = await fetchJSON('/api/folders', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
          });
          closeModal(newFolderModal);
          const slug = folder.slug || folder.Slug;
          window.location.href = slug ? '/?folder=' + encodeURIComponent(slug) : window.location.href;
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    }

    function updateShareDetails(data) {
      if (!shareDetails) return;
      if (data.visibility !== 'shared') {
        shareDetails.hidden = true;
        return;
      }
      shareDetails.hidden = false;
      const link = data.shareUrl || (data.sharedToken ? window.location.origin + '/shared/' + data.sharedToken : '');
      shareLinkValue.textContent = link || 'Brak linku';
      shareLinkValue.dataset.link = link;
      shareViewsValue.textContent = String(data.sharedViews ?? 0);
      copyShareLink.disabled = !link;
      regenerateLinkButton.disabled = !data.id;
      downloadQrButton.disabled = !data.id;
    }

    function currentFolderData() {
      return {
        id: state.activeFolderId,
        visibility: state.activeFolderVisibility || 'private',
        sharedToken: state.activeFolderShareToken || '',
        shareUrl: state.activeFolderShareUrl || '',
        sharedViews: state.activeFolderShareViews || 0
      };
    }

    folderSettingsButton?.addEventListener('click', () => {
      if (!state.activeFolderId) {
        showMessage('Brak wybranego folderu', 'error');
        return;
      }
      const data = currentFolderData();
      folderSettingsForm?.querySelectorAll('input[name="visibility"]').forEach(radio => {
        radio.checked = radio.value === data.visibility;
      });
      updateShareDetails({...data, visibility: data.visibility});
      openModal(folderSettingsModal);
    });

    folderSettingsCancel?.addEventListener('click', () => closeModal(folderSettingsModal));

    folderSettingsForm?.addEventListener('submit', async event => {
      event.preventDefault();
      if (!state.activeFolderId) return;
      const visibility = folderSettingsForm.elements['visibility'].value;
      try {
        await fetchJSON('/api/folders/' + state.activeFolderId, {
          method: 'PATCH',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({visibility})
        });
        window.location.reload();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });
    regenerateLinkButton?.addEventListener('click', async () => {
      if (!state.activeFolderId) return;
      try {
        await fetchJSON('/api/folders/' + state.activeFolderId, {
          method: 'PATCH',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({regenerateLink: true})
        });
        window.location.reload();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });

    downloadQrButton?.addEventListener('click', () => {
      if (!state.activeFolderId) return;
      window.open('/api/folders/' + state.activeFolderId + '/qr', '_blank');
    });

    copyShareLink?.addEventListener('click', async () => {
      const link = shareLinkValue?.dataset.link;
      if (!link) {
        showMessage('Brak linku do skopiowania', 'error');
        return;
      }
      try {
        await navigator.clipboard.writeText(link);
        showMessage('Skopiowano link');
      } catch (_) {
        showMessage('Nie udalo sie skopiowac linku', 'error');
      }
    });
  </script>
</body>
</html>
`
