document.addEventListener('DOMContentLoaded', async () => {
  const user = requireRole('staff');
  if (!user) return;
  renderStaffChrome('branding', user);

  function showLogo(url) {
    $('#logo-preview').src = url;
    $('#logo-preview').classList.remove('hidden');
    $('#logo-placeholder').classList.add('hidden');
  }

  async function load() {
    try {
      const b = await api('/api/clinic/branding');
      $('#brand-name').value = b.clinicName;
      $('#brand-tagline').value = b.tagline;
      $('#brand-color').value = b.primaryColor;
      if (b.logoUrl) showLogo(b.logoUrl);
    } catch (e) { toast(e.message, 'error'); }
  }

  // local preview before upload (dynamic element)
  $('#logo-input').addEventListener('change', () => {
    const file = $('#logo-input').files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => showLogo(reader.result);
    reader.readAsDataURL(file);
  });

  $('#btn-upload-logo').addEventListener('click', async () => {
    const file = $('#logo-input').files[0];
    if (!file) { toast('Choose a file first.', 'error'); return; }
    const fd = new FormData();
    fd.append('file', file);
    try {
      const res = await api('/api/clinic/branding/logo', { method: 'POST', body: fd });
      showLogo(res.logoUrl);
      toast('Logo uploaded.', 'success');
    } catch (e) { toast(e.message, 'error'); }
  });

  $('#btn-save-branding').addEventListener('click', async () => {
    try {
      await api('/api/clinic/branding', {
        method: 'PUT',
        body: {
          clinicName: $('#brand-name').value.trim(),
          tagline: $('#brand-tagline').value.trim(),
          primaryColor: $('#brand-color').value.trim()
        }
      });
      applyTheme();
      toast('Branding saved.', 'success');
    } catch (e) { toast(e.message, 'error'); }
  });

  load();
});