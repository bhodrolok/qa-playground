const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

function esc(s) {
  return String(s).replace(/[&<>"']/g, c => (
    { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
  ));
}

class ApiError extends Error {
  constructor(message, status, data) { super(message); this.status = status; this.data = data; }
}

async function api(path, { method = 'GET', body = null } = {}) {
  const headers = {};
  const token = localStorage.getItem('token');
  if (token) headers['Authorization'] = 'Bearer ' + token;
  if (body !== null && !(body instanceof FormData)) headers['Content-Type'] = 'application/json';

  const res = await fetch(path, {
    method,
    headers,
    body: body instanceof FormData ? body : (body !== null ? JSON.stringify(body) : undefined)
  });

  let data = null;
  const text = await res.text();
  if (text) { try { data = JSON.parse(text); } catch (e) { /* non-JSON */ } }

  if (!res.ok) throw new ApiError((data && data.error) || ('HTTP ' + res.status), res.status, data);
  return data;
}

function toast(msg, type = 'info') {
  let el = $('#toast');
  if (!el) { el = document.createElement('div'); el.id = 'toast'; document.body.appendChild(el); }
  el.textContent = msg;
  el.className = 'toast show ' + type;
  clearTimeout(el._t);
  el._t = setTimeout(() => el.classList.remove('show'), 4000);
}

function loginRedirectPath() {
  const p = location.pathname;
  if (p.includes('/staff/') || p.includes('/user/')) return '../index.html';
  return 'index.html';
}

function homePath(role) {
  return role === 'staff' ? 'staff/dashboard.html' : 'user/dashboard.html';
}

function requireRole(role) {
  const token = localStorage.getItem('token');
  const user = JSON.parse(localStorage.getItem('user') || 'null');
  if (!token || !user) { location.replace(loginRedirectPath()); return null; }
  if (role && user.role !== role) { location.replace(homePath(user.role)); return null; }
  return user;
}

function logout() {
  api('/api/auth/logout', { method: 'POST' }).catch(() => {});
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  location.replace(loginRedirectPath());
}

// ---- staff chrome (sidebar + topbar) ----
const STAFF_PAGES = [
  { key: 'dashboard',  label: 'Dashboard',          href: 'dashboard.html' },
  { key: 'appointments', label: 'Appointments',     href: 'appointments.html' },
  { key: 'hours',      label: 'Clinic Hours',       href: 'hours.html' },
  { key: 'branding',   label: 'Branding & Settings', href: 'branding.html' },
  { key: 'templates',  label: 'Message Templates',  href: 'templates.html' },
];

function renderStaffChrome(activeKey, user) {
  const nav = $('#sidebar');
  nav.innerHTML = STAFF_PAGES.map(p =>
    `<a href="${p.href}" class="nav-link${p.key === activeKey ? ' active' : ''}" data-nav="${p.key}"><span>${p.label}</span></a>`
  ).join('');

  const tb = $('#topbar');
  tb.innerHTML = `
    <input id="global-search" type="search" placeholder="Search (non-functional)…" aria-label="Search">
    <div class="topbar-right">
      <span id="user-email" class="muted">${esc(user.email)}</span>
      <button id="btn-logout" class="btn btn-ghost">Log out</button>
    </div>`;
  $('#btn-logout').addEventListener('click', logout);
}

async function applyTheme() {
  try {
    const b = await api('/api/clinic/branding');
    if (b.clinicName) document.title = b.clinicName;
    if (b.primaryColor) document.documentElement.style.setProperty('--primary', b.primaryColor);
  } catch (e) { /* theme is cosmetic */ }
}