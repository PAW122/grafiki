package app

const PageTemplate = `<!DOCTYPE html>
<html lang="pl">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Galeria zdjec</title>
  <link rel="icon" type="image/x-icon" href="/favicon.ico">
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
      min-height: 100vh;
      font-family: 'Inter', 'Segoe UI', system-ui, sans-serif;
    }
    .app-wrapper {
      display: flex;
      min-height: 100vh;
    }
    .side-menu {
      width: 230px;
      background: #0f172a;
      color: #fff;
      display: flex;
      flex-direction: column;
      gap: 1.25rem;
      padding: 2rem 1.25rem;
    }
    .side-menu-title {
      font-weight: 700;
      letter-spacing: 0.05em;
      text-transform: uppercase;
      font-size: 0.9rem;
      opacity: 0.8;
    }
    .side-menu-links {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    .menu-link {
      border: none;
      border-radius: 12px;
      padding: 0.65rem 0.9rem;
      background: transparent;
      color: #cbd5f5;
      text-align: left;
      font-weight: 600;
      cursor: pointer;
      transition: background 0.18s ease, color 0.18s ease;
    }
    .menu-link:hover,
    .menu-link:focus-visible {
      background: rgba(255, 255, 255, 0.08);
      color: #fff;
      outline: none;
    }
    .menu-link.active {
      background: #2563eb;
      color: #fff;
    }
    .page-wrapper {
      flex: 1;
      display: flex;
      flex-direction: column;
      min-width: 0;
    }
    .view-section {
      display: none;
    }
    body[data-page-view="gallery"] .view-gallery {
      display: block;
    }
    body[data-page-view="submitted"] .view-submissions {
      display: block;
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
    .btn-tertiary {
      background: rgba(37, 99, 235, 0.12);
      color: #1d4ed8;
      box-shadow: inset 0 0 0 1px rgba(37, 99, 235, 0.2);
    }
    .btn-tertiary:hover {
      background: rgba(37, 99, 235, 0.24);
      color: #1e3a8a;
    }
    .inline-form {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
      align-items: center;
    }
    .inline-form input[type="text"] {
      border: 1px solid rgba(15, 23, 42, 0.15);
      border-radius: 999px;
      padding: 0.5rem 1.1rem;
      min-width: 200px;
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
    .submissions-panel {
      border: 1px solid rgba(148, 163, 184, 0.15);
    }
    .submissions-grid {
      margin-top: 1.25rem;
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 1rem;
    }
    .submission-card {
      border: 1px solid rgba(148, 163, 184, 0.25);
      border-radius: 18px;
      padding: 1rem;
      display: flex;
      flex-direction: column;
      gap: 0.8rem;
      background: #fff;
      transition: border 0.18s ease, box-shadow 0.18s ease;
    }
    .submission-card.active {
      border-color: #2563eb;
      box-shadow: 0 12px 32px rgba(37, 99, 235, 0.15);
    }
    .submission-name {
      font-weight: 700;
      font-size: 1rem;
    }
    .submission-meta {
      display: flex;
      gap: 0.65rem;
      font-size: 0.85rem;
      color: #475569;
      flex-wrap: wrap;
    }
    .submission-card-actions {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
      align-items: center;
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
    .folder-card:focus-visible {
      outline: 2px solid #2563eb;
      outline-offset: 2px;
    }
    .folder-card-body {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
      flex: 1;
    }
    .folder-card-actions {
      display: flex;
      justify-content: flex-end;
      margin-top: 0.4rem;
    }
    .folder-delete-btn {
      border: none;
      border-radius: 999px;
      padding: 0.35rem 0.9rem;
      font-size: 0.8rem;
      font-weight: 600;
      background: rgba(239, 68, 68, 0.12);
      color: #b91c1c;
      cursor: pointer;
      transition: background 0.18s ease, color 0.18s ease;
    }
    .folder-delete-btn:hover,
    .folder-delete-btn:focus-visible {
      background: rgba(239, 68, 68, 0.25);
      color: #7f1d1d;
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
    .submission-list {
      margin-top: 1.5rem;
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    .submission-entry {
      display: flex;
      gap: 1rem;
      padding: 1rem;
      background: #f8fafc;
      border: 1px solid rgba(148, 163, 184, 0.3);
      border-radius: 20px;
      flex-wrap: wrap;
    }
    .submission-preview {
      width: 120px;
      height: 120px;
      border-radius: 16px;
      overflow: hidden;
      background: #e2e8f0;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 700;
      color: #475569;
      flex-shrink: 0;
    }
    .submission-preview img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }
    .pdf-preview,
    .file-preview {
      font-size: 1.1rem;
      letter-spacing: 0.08em;
    }
    .submission-info {
      flex: 1;
      min-width: 220px;
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .submission-info h3 {
      margin: 0;
      font-size: 1rem;
      color: #0f172a;
    }
    .submission-info p {
      margin: 0;
      font-size: 0.92rem;
      color: #475569;
    }
    .submission-actions {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }
    .submissions-settings {
      margin: 1.25rem 0;
      padding: 1rem 1.25rem;
      background: #f1f5f9;
      border-radius: 16px;
      border: 1px solid rgba(148, 163, 184, 0.25);
      display: flex;
      flex-direction: column;
      gap: 1rem;
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
    .tile-actions {
      display: flex;
      gap: 0.35rem;
    }
    .filename {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .image-rename-btn {
      border: none;
      border-radius: 8px;
      padding: 0.35rem 0.8rem;
      background: rgba(59, 130, 246, 0.18);
      color: #1d4ed8;
      font-weight: 600;
      cursor: pointer;
      transition: background 0.18s ease, transform 0.18s ease;
    }
    .image-rename-btn:hover {
      background: rgba(59, 130, 246, 0.28);
      transform: translateY(-2px);
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
    @media (max-width: 960px) {
      .app-wrapper {
        flex-direction: column;
      }
      .side-menu {
        width: 100%;
        flex-direction: row;
        align-items: center;
        gap: 1rem;
      }
      .side-menu-links {
        flex-direction: row;
        flex-wrap: wrap;
      }
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
<body data-page-view="{{.View}}" data-logged-in="{{if .LoggedIn}}true{{else}}false{{end}}" data-upload-limit="{{.SubmissionUploadLimit}}" data-shared-mode="{{if .SharedMode}}true{{else}}false{{end}}" data-sub-shared-mode="{{if .SubmissionSharedMode}}true{{else}}false{{end}}" data-active-folder="{{if .ActiveFolder}}{{.ActiveFolder.Slug}}{{end}}" data-active-folder-id="{{if .ActiveFolder}}{{.ActiveFolder.ID}}{{end}}" data-active-folder-visibility="{{if .ActiveFolder}}{{.ActiveFolder.Visibility}}{{end}}" data-active-folder-share-token="{{if .ActiveFolder}}{{.ActiveFolder.SharedToken}}{{end}}" data-active-folder-share-url="{{if .ActiveFolder}}{{.ActiveFolder.ShareURL}}{{end}}" data-active-folder-share-views="{{if .ActiveFolder}}{{.ActiveFolder.SharedViews}}{{end}}" data-active-folder-name="{{if .ActiveFolder}}{{.ActiveFolder.Name}}{{end}}" data-sub-active-group="{{if .ActiveSubmissionGroup}}{{.ActiveSubmissionGroup.Slug}}{{end}}" data-sub-active-group-id="{{if .ActiveSubmissionGroup}}{{.ActiveSubmissionGroup.ID}}{{end}}" data-sub-active-group-visibility="{{if .ActiveSubmissionGroup}}{{.ActiveSubmissionGroup.Visibility}}{{end}}" data-sub-active-group-share-token="{{if .ActiveSubmissionGroup}}{{.ActiveSubmissionGroup.SharedToken}}{{end}}" data-sub-active-group-share-url="{{if .ActiveSubmissionGroup}}{{.ActiveSubmissionGroup.ShareURL}}{{end}}">
  <div class="app-wrapper">
    {{if .LoggedIn}}
    <aside class="side-menu">
      <div class="side-menu-title">Menu</div>
      <nav class="side-menu-links">
        <button type="button" class="menu-link {{if ne .View "submitted"}}active{{end}}" data-view-target="gallery">Galeria</button>
        <button type="button" class="menu-link {{if eq .View "submitted"}}active{{end}}" data-view-target="submitted">Przeslane</button>
      </nav>
    </aside>
    {{end}}
    <div class="page-wrapper">
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
    <section class="view-section view-gallery">
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
        <div class="folder-card {{if and $.ActiveFolder (eq $.ActiveFolder.Slug .Slug)}}active{{end}}" role="button" tabindex="0" data-slug="{{.Slug}}" data-folder-id="{{.ID}}" data-folder-name="{{.Name}}">
          <div class="folder-card-body">
            <div class="folder-name">{{.Name}}</div>
            <div class="folder-meta">
              <span class="badge {{.Visibility}}">
                {{if eq .Visibility "public"}}Publiczny{{else if eq .Visibility "shared"}}Udostepniony{{else}}Prywatny{{end}}
              </span>
              {{if eq .Visibility "shared"}}
              <span>{{.SharedViews}} wejsc</span>
              {{end}}
            </div>
          </div>
          {{if $.AllowFolderManagement}}
          <div class="folder-card-actions">
            <button type="button" class="folder-delete-btn" data-folder-id="{{.ID}}" data-folder-name="{{.Name}}">
              Usun folder
            </button>
          </div>
          {{end}}
        </div>
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
            <div class="tile-actions">
              <button type="button" class="image-rename-btn" data-name="{{.Name}}" data-folder="{{$.ActiveFolder.Slug}}">Zmien nazwe</button>
              <button type="button" class="delete-btn" data-name="{{.Name}}" data-folder="{{$.ActiveFolder.Slug}}">Usun</button>
            </div>
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
    </section>

    <section class="view-section view-submissions" id="submittedView">
      {{if or .AllowSubmissionManagement .SubmissionGroups}}
      <div class="section-card submissions-panel">
        <div class="section-header">
          <div>
            <h2>Przeslane pliki</h2>
            <p>Zarzadzaj grupami i przegladaj przeslane materialy.</p>
          </div>
          {{if .AllowSubmissionManagement}}
          <form id="newSubmissionGroupForm" class="inline-form">
            <input type="text" name="name" placeholder="Nowa grupa" required>
            <button class="btn btn-secondary" type="submit">Dodaj</button>
          </form>
          {{end}}
        </div>
        {{if .SubmissionGroups}}
        <div class="submissions-grid">
          {{range .SubmissionGroups}}
          <div class="submission-card {{if and $.ActiveSubmissionGroup (eq $.ActiveSubmissionGroup.ID .ID)}}active{{end}}">
            <div>
              <div class="submission-name">{{.Name}}</div>
              <div class="submission-meta">
                <span class="badge {{.Visibility}}">
                  {{if eq .Visibility "public"}}Publiczna{{else if eq .Visibility "shared"}}Udostepniony link{{else}}Prywatna{{end}}
                </span>
                {{if .SharedViews}}
                <span>{{.SharedViews}} wejsc</span>
                {{end}}
              </div>
            </div>
            <div class="submission-card-actions">
              <a class="btn btn-tertiary" href="/?view=submitted&group={{.Slug}}">Otworz</a>
              {{if $.AllowSubmissionManagement}}
              <button type="button" class="ghost submission-delete-btn" data-group-id="{{.ID}}" data-group-name="{{.Name}}">Usun</button>
              {{end}}
            </div>
          </div>
          {{end}}
        </div>
        {{else}}
        <p class="empty-state">Brak grup przeslan.</p>
        {{end}}
      </div>
      {{end}}

      <div class="section-card submissions-workspace {{if not .ActiveSubmissionGroup}}workspace-empty{{end}}" id="submissionsWorkspace">
        {{if .ActiveSubmissionGroup}}
        <div class="workspace-header">
          <div>
            <h2>{{.ActiveSubmissionGroup.Name}}</h2>
            <p class="workspace-subtitle">
              {{if eq .ActiveSubmissionGroup.Visibility "public"}}Grupa publiczna. Wszystkie pliki sa widoczne dla wszystkich.{{else if eq .ActiveSubmissionGroup.Visibility "shared"}}Grupa dostepna z tajnym linkiem. Odwiedzajacy widza jedynie swoje pliki.{{else}}Grupa prywatna. Dostepna tylko po zalogowaniu.{{end}}
            </p>
          </div>
          <div class="workspace-actions">
            <span class="badge {{.ActiveSubmissionGroup.Visibility}}">
              {{if eq .ActiveSubmissionGroup.Visibility "public"}}Publiczna{{else if eq .ActiveSubmissionGroup.Visibility "shared"}}Udostepniony link{{else}}Prywatna{{end}}
            </span>
            {{if .SubmissionSharedMode}}
            <span class="badge shared">Udostepniony link</span>
            {{end}}
          </div>
        </div>

        <div class="share-details" id="submissionShareDetails" {{if not .SubmissionShareLink}}hidden{{end}}>
          <strong>Link udostepniony</strong>
          <div class="share-link-row">
            <code id="submissionShareLinkValue" data-link="{{.SubmissionShareLink}}">{{if .SubmissionShareLink}}{{.SubmissionShareLink}}{{else}}Brak linku{{end}}</code>
            {{if .SubmissionShareLink}}
            <button type="button" class="ghost" id="submissionCopyLink">Kopiuj</button>
            {{end}}
            {{if .AllowSubmissionManagement}}
            <button type="button" class="ghost" id="submissionRegenerateLink">Nowy link</button>
            {{end}}
          </div>
        </div>

        {{if .AllowSubmissionManagement}}
        <form id="submissionGroupSettingsForm" class="submissions-settings">
          <label>
            Nazwa grupy
            <input type="text" name="groupName" value="{{.ActiveSubmissionGroup.Name}}" required>
          </label>
          <span class="section-label">Widocznosc</span>
          <div class="visibility-options">
            <label class="radio-option">
              <input type="radio" name="submissionVisibility" value="public" {{if eq .ActiveSubmissionGroup.Visibility "public"}}checked{{end}}>
              <span class="radio-description">
                <strong>Publiczna</strong>
                <span>Dostepna dla wszystkich odwiedzajacych.</span>
              </span>
            </label>
            <label class="radio-option">
              <input type="radio" name="submissionVisibility" value="shared" {{if eq .ActiveSubmissionGroup.Visibility "shared"}}checked{{end}}>
              <span class="radio-description">
                <strong>Udostepniony link</strong>
                <span>Wejscie tylko przez sekret link.</span>
              </span>
            </label>
            <label class="radio-option">
              <input type="radio" name="submissionVisibility" value="private" {{if eq .ActiveSubmissionGroup.Visibility "private"}}checked{{end}}>
              <span class="radio-description">
                <strong>Prywatna</strong>
                <span>Tylko administrator ma dostep.</span>
              </span>
            </label>
          </div>
          <div class="modal-actions">
            <button class="primary" type="submit">Zapisz</button>
          </div>
        </form>
        {{end}}

        {{if .SubmissionSharedMode}}
        <div class="info-panel">Ten widok pokazuje jedynie pliki przeslane z tego urzadzenia. Aby zobaczyc inne, uzyj wlasnego linku.</div>
        {{end}}

        {{if .AllowSubmissionUpload}}
        <form id="submissionUploadForm" class="upload-panel">
          <input type="hidden" name="group" value="{{.ActiveSubmissionGroup.Slug}}">
          {{if .SubmissionSharedMode}}
          <input type="hidden" name="token" value="{{.ActiveSubmissionGroup.SharedToken}}">
          {{end}}
          <label>
            Twoja nazwa
            <input type="text" name="name" placeholder="np. Jan Kowalski" required>
          </label>
          <label>
            Plik
            <input type="file" name="file" required accept=".jpg,.jpeg,.png,.gif,.bmp,.svg,.webp,.avif,.pdf">
          </label>
          <small>Maksymalny rozmiar {{.SubmissionUploadLimit}} MB. Dozwolone obrazy i PDF.</small>
          <button class="submit-btn" type="submit">Przeslij</button>
        </form>
        {{else}}
        <div class="info-panel">Wysylanie plikow jest wylaczone dla tej grupy.</div>
        {{end}}

        {{if .SubmissionEntries}}
        <div class="submission-list">
          {{range .SubmissionEntries}}
          <article class="submission-entry">
            <div class="submission-preview">
              {{if .IsImage}}
              <img src="{{.URL}}" alt="{{.Original}}">
              {{else if .IsPDF}}
              <div class="pdf-preview">PDF</div>
              {{else}}
              <div class="file-preview">Plik</div>
              {{end}}
            </div>
            <div class="submission-info">
              <h3>{{.Original}}</h3>
              <p>Dodane przez <strong>{{.UploadedBy}}</strong> • {{.UploadedAt}} • {{.SizeLabel}}</p>
              <div class="submission-actions">
                <a class="btn btn-secondary" href="{{.URL}}" target="_blank" rel="noopener">Podglad</a>
                <a class="btn btn-tertiary" href="{{.DownloadURL}}">Pobierz</a>
              </div>
            </div>
          </article>
          {{end}}
        </div>
        {{else}}
        <p class="empty">Brak plikow w tej grupie.</p>
        {{end}}
        {{else}}
        <p class="empty-state large">Wybierz grupe, aby zobaczyc przeslane pliki.</p>
        {{end}}
      </div>
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
        <label>
          Nazwa folderu
          <input type="text" name="folderName" id="folderNameInput" required>
        </label>
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
    </div>
  </div>

  <script>
    const state = (() => {
      const dataset = document.body?.dataset || {};
      return {
        pageView: dataset.pageView || 'gallery',
        loggedIn: dataset.loggedIn === 'true',
        sharedMode: dataset.sharedMode === 'true',
        activeFolder: dataset.activeFolder || '',
        activeFolderId: Number(dataset.activeFolderId || 0),
        activeFolderVisibility: dataset.activeFolderVisibility || '',
        activeFolderShareToken: dataset.activeFolderShareToken || '',
        activeFolderShareUrl: dataset.activeFolderShareUrl || '',
        activeFolderShareViews: Number(dataset.activeFolderShareViews || 0),
        activeFolderName: dataset.activeFolderName || '',
        submissionSharedMode: dataset.subSharedMode === 'true',
        activeSubmissionGroup: dataset.subActiveGroup || '',
        activeSubmissionGroupId: Number(dataset.subActiveGroupId || 0),
        activeSubmissionGroupVisibility: dataset.subActiveGroupVisibility || '',
        activeSubmissionShareToken: dataset.subActiveGroupShareToken || '',
        activeSubmissionShareUrl: dataset.subActiveGroupShareUrl || '',
        uploadLimit: Number(dataset.uploadLimit || 0)
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
    const folderNameInput = document.getElementById('folderNameInput');
    const shareDetails = document.getElementById('shareDetails');
    const shareLinkValue = document.getElementById('shareLinkValue');
    const shareViewsValue = document.getElementById('shareViewsValue');
    const copyShareLink = document.getElementById('copyShareLink');
    const regenerateLinkButton = document.getElementById('regenerateLinkButton');
    const downloadQrButton = document.getElementById('downloadQrButton');
    const viewSwitchButtons = document.querySelectorAll('[data-view-target]');
    const newSubmissionGroupForm = document.getElementById('newSubmissionGroupForm');
    const submissionGroupSettingsForm = document.getElementById('submissionGroupSettingsForm');
    const submissionUploadForm = document.getElementById('submissionUploadForm');
    const submissionDeleteButtons = document.querySelectorAll('.submission-delete-btn');
    const submissionCopyLinkButton = document.getElementById('submissionCopyLink');
    const submissionRegenerateLinkButton = document.getElementById('submissionRegenerateLink');
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

    submissionDeleteButtons.forEach(btn => {
      btn.addEventListener('click', async event => {
        event.preventDefault();
        if (!confirm('Czy na pewno usunac te grupe wraz z przeslanymi plikami?')) {
          return;
        }
        const id = btn.dataset.groupId;
        if (!id) {
          showMessage('Brak identyfikatora grupy', 'error');
          return;
        }
        try {
          await fetchJSON('/api/submissions/groups/' + id, { method: 'DELETE' });
          window.location.reload();
        } catch (err) {
          showMessage(err.message, 'error');
        }
      });
    });

    submissionGroupSettingsForm?.addEventListener('submit', async event => {
      event.preventDefault();
      if (!state.activeSubmissionGroupId) {
        showMessage('Brak wybranej grupy', 'error');
        return;
      }
      const formData = new FormData(submissionGroupSettingsForm);
      const name = String(formData.get('groupName') || '').trim();
      if (!name) {
        showMessage('Nazwa grupy jest wymagana', 'error');
        return;
      }
      const visibility = formData.get('submissionVisibility');
      try {
        const updated = await fetchJSON('/api/submissions/groups/' + state.activeSubmissionGroupId, {
          method: 'PATCH',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({name, visibility})
        });
        const slug = updated.slug || updated.Slug;
        const next = new URL(window.location.href);
        next.searchParams.set('view', 'submitted');
        if (slug) {
          next.searchParams.set('group', slug);
        }
        window.location.href = next.toString();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });

    submissionRegenerateLinkButton?.addEventListener('click', async () => {
      if (!state.activeSubmissionGroupId) {
        return;
      }
      try {
        await fetchJSON('/api/submissions/groups/' + state.activeSubmissionGroupId, {
          method: 'PATCH',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({regenerateLink: true})
        });
        window.location.reload();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });

    submissionCopyLinkButton?.addEventListener('click', async () => {
      const link = document.getElementById('submissionShareLinkValue')?.dataset.link;
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

    submissionUploadForm?.addEventListener('submit', async event => {
      event.preventDefault();
      const formData = new FormData(submissionUploadForm);
      const name = String(formData.get('name') || '').trim();
      if (!name) {
        showMessage('Podpisz sie przed wysylka', 'error');
        return;
      }
      if (!formData.get('file')) {
        showMessage('Wybierz plik do przeslania', 'error');
        return;
      }
      try {
        await fetchJSON('/api/submissions/upload', {
          method: 'POST',
          body: formData
        });
        window.location.reload();
      } catch (err) {
        showMessage(err.message, 'error');
      }
    });

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

    viewSwitchButtons.forEach(btn => {
      btn.addEventListener('click', event => {
        event.preventDefault();
        const target = btn.dataset.viewTarget;
        const url = new URL(window.location.href);
        if (target === 'gallery') {
          url.searchParams.delete('view');
          url.searchParams.delete('group');
        } else if (target === 'submitted') {
          url.searchParams.set('view', 'submitted');
          if (state.activeSubmissionGroup) {
            url.searchParams.set('group', state.activeSubmissionGroup);
          }
        }
        window.location.href = url.toString();
      });
    });

    document.querySelectorAll('.folder-card').forEach(card => {
      const slug = card.dataset.slug;
      if (!slug) {
        return;
      }
      const openFolder = () => {
        window.location.href = '/?folder=' + encodeURIComponent(slug);
      };
      card.addEventListener('click', () => {
        openFolder();
      });
      card.addEventListener('keydown', event => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault();
          openFolder();
        }
      });
    });

    document.querySelectorAll('.folder-delete-btn').forEach(btn => {
      btn.addEventListener('click', async event => {
        event.preventDefault();
        event.stopPropagation();
        const folderId = btn.dataset.folderId;
        const folderName = btn.dataset.folderName || '';
        if (!folderId) {
          showMessage('Nie mozna usunac folderu', 'error');
          return;
        }
        const confirmed = confirm('Czy na pewno chcesz usunac folder "' + (folderName || 'bez nazwy') + '" wraz ze wszystkimi grafikami oraz linkami?');
        if (!confirmed) return;
        try {
          await fetchJSON('/api/folders/' + folderId, {
            method: 'DELETE'
          });
          window.location.href = '/';
        } catch (err) {
          showMessage(err.message, 'error');
        }
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

    document.querySelectorAll('.image-rename-btn').forEach(btn => {
      btn.addEventListener('click', async event => {
        event.preventDefault();
        event.stopPropagation();
        const currentName = btn.dataset.name;
        const folder = btn.dataset.folder || state.activeFolder;
        if (!currentName || !folder) {
          showMessage('Brak danych do zmiany nazwy', 'error');
          return;
        }
        const proposed = prompt('Podaj nowa nazwe pliku (wraz z rozszerzeniem)', currentName);
        if (proposed === null) {
          return;
        }
        const trimmed = proposed.trim();
        if (!trimmed || trimmed === currentName) {
          return;
        }
        try {
          await fetchJSON('/api/images/rename', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({folder, oldName: currentName, newName: trimmed})
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

    if (newSubmissionGroupForm) {
      newSubmissionGroupForm.addEventListener('submit', async event => {
        event.preventDefault();
        const formData = new FormData(newSubmissionGroupForm);
        const name = String(formData.get('name') || '').trim();
        if (!name) {
          showMessage('Podaj nazwe grupy', 'error');
          return;
        }
        try {
          const group = await fetchJSON('/api/submissions/groups', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name})
          });
          const slug = group.slug || group.Slug;
          const next = new URL(window.location.href);
          next.searchParams.set('view', 'submitted');
          if (slug) {
            next.searchParams.set('group', slug);
          }
          window.location.href = next.toString();
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
        name: state.activeFolderName || '',
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
      if (folderNameInput) {
        folderNameInput.value = data.name || '';
        folderNameInput.focus();
        folderNameInput.select();
      }
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
      const payload = { visibility };
      if (folderNameInput) {
        const nameValue = folderNameInput.value.trim();
        if (!nameValue) {
          showMessage('Nazwa folderu nie moze byc pusta', 'error');
          folderNameInput.focus();
          return;
        }
        payload.name = nameValue;
      }
      try {
        const updated = await fetchJSON('/api/folders/' + state.activeFolderId, {
          method: 'PATCH',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(payload)
        });
        const slug = updated?.slug || updated?.Slug;
        if (slug && slug !== state.activeFolder) {
          window.location.href = '/?folder=' + encodeURIComponent(slug);
        } else {
          window.location.reload();
        }
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
