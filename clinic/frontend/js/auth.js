document.addEventListener('DOMContentLoaded', async () => {
  // branding on auth pages
  try {
    const b = await api('/api/clinic/branding');
    const t = $('#clinic-name'); if (t) t.textContent = b.clinicName;
    const tag = $('#clinic-tagline'); if (tag) tag.textContent = b.tagline;
  } catch (e) { /* fall back to defaults */ }

  // prefill email after successful registration
  const last = localStorage.getItem('lastRegisteredEmail');
  if (last && $('#login-form')) {
    $('#email').value = last;
    localStorage.removeItem('lastRegisteredEmail');
  }

  // ---- LOGIN ----
  const loginForm = $('#login-form');
  if (loginForm) {
    loginForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const err = $('#login-error');
      err.classList.add('hidden');

      const email = $('#email').value.trim();
      const password = $('#password').value;
      if (!email || !password) {
        err.textContent = 'Please fill in both fields.';
        err.classList.remove('hidden');
        return;
      }

      const btn = $('#btn-login');
      btn.disabled = true;
      btn.textContent = 'Signing in…';
      try {
        await new Promise(res => setTimeout(res, 600)); // simulated latency
        const data = await api('/api/auth/login', { method: 'POST', body: { email, password } });
        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));
        location.href = homePath(data.user.role);
      } catch (ex) {
        err.textContent = ex.message || 'Login failed.';
        err.classList.remove('hidden');
        btn.disabled = false;
        btn.textContent = 'Sign in';
      }
    });
  }

  // ---- REGISTER ----
  const regForm = $('#register-form');
  if (regForm) {
    regForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const err = $('#reg-error');
      err.classList.add('hidden');

      const name = $('#reg-name').value.trim();
      const email = $('#reg-email').value.trim();
      const password = $('#reg-password').value;
      const confirm = $('#reg-confirm').value;
      const role = $('#reg-role').value;

      const showErr = m => { err.textContent = m; err.classList.remove('hidden'); };
      if (!name) return showErr('Name is required.');
      if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email)) return showErr('Please enter a valid email address.');
      if (password.length < 8) return showErr('Password must be at least 8 characters.');
      if (password !== confirm) return showErr('Passwords do not match.');

      const btn = $('#btn-register');
      btn.disabled = true;
      btn.textContent = 'Creating account…';
      try {
        await api('/api/auth/register', { method: 'POST', body: { name, email, password, role } });
        localStorage.setItem('lastRegisteredEmail', email);
        location.href = 'index.html';
      } catch (ex) {
        showErr(ex.status === 409 ? 'This email is already registered.' : (ex.message || 'Registration failed.'));
        btn.disabled = false;
        btn.textContent = 'Create account';
      }
    });
  }
});