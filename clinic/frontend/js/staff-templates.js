document.addEventListener('DOMContentLoaded', () => {
  const user = requireRole('staff');
  if (!user) return;
  renderStaffChrome('templates', user);
  applyTheme();

  const SAMPLE = {
    patient_name: 'John Smith',
    appointment_date: '2025-01-15',
    appointment_time: '09:30',
    clinic_name: 'Pulse Clinic'
  };
  const form = $('#template-form');

  async function loadTemplates() {
    const rows = $('#tpl-rows');
    rows.innerHTML = `<tr><td colspan="5" class="muted">Loading…</td></tr>`;
    try {
      const d = await api('/api/templates');
      rows.innerHTML = '';
      if (!d.items.length) {
        rows.innerHTML = `<tr><td colspan="5" class="muted">No templates yet.</td></tr>`;
        return;
      }
      d.items.forEach(t => {
        rows.insertAdjacentHTML('beforeend', `<tr data-id="${t.id}">
          <td>#${t.id}</td>
          <td>${esc(t.name)}</td>
          <td><span class="badge">${esc(t.channel)}</span></td>
          <td>${esc(t.subject)}</td>
          <td>
            <button class="btn btn-small" data-action="edit">Edit</button>
            <button class="btn btn-small btn-danger" data-action="delete">Delete</button>
          </td>
        </tr>`);
      });
    } catch (e) {
      rows.innerHTML = `<tr><td colspan="5" class="error">${esc(e.message)}</td></tr>`;
    }
  }

  function openForm(tpl) {
    form.classList.remove('hidden');
    $('#template-form-title').textContent = tpl ? 'Edit template #' + tpl.id : 'New template';
    $('#tpl-id').value = tpl ? tpl.id : '';
    $('#tpl-name').value = tpl ? tpl.name : '';
    $('#tpl-channel').value = tpl ? tpl.channel : 'email';
    $('#tpl-subject').value = tpl ? tpl.subject : '';
    $('#tpl-body').value = tpl ? tpl.body : '';
    $('#tpl-error').classList.add('hidden');
    renderPreview();
    $('#tpl-name').focus();
  }

  function renderPreview() {
    const fill = s => s.replace(/\{\{\s*(\w+)\s*\}\}/g, (m, k) => (SAMPLE[k] !== undefined ? SAMPLE[k] : m));
    const subj = fill($('#tpl-subject').value);
    const body = fill($('#tpl-body').value);
    $('#tpl-preview').innerHTML =
      (subj ? `<strong>${esc(subj)}</strong><br>` : '') +
      (body ? esc(body).replace(/\n/g, '<br>') : '<em>(empty body)</em>');
  }
  $('#tpl-body').addEventListener('input', renderPreview);
  $('#tpl-subject').addEventListener('input', renderPreview);

  $$('.chip').forEach(chip => chip.addEventListener('click', () => {
    const ta = $('#tpl-body');
    const v = '{{' + chip.dataset.var + '}}';
    const pos = ta.selectionStart || ta.value.length;
    const end = ta.selectionEnd || pos;
    ta.value = ta.value.slice(0, pos) + v + ta.value.slice(end);
    ta.focus();
    renderPreview();
  }));

  $('#btn-new-template').addEventListener('click', () => openForm(null));
  $('#btn-close-template').addEventListener('click', () => form.classList.add('hidden'));

  $('#btn-save-template').addEventListener('click', async () => {
    const err = $('#tpl-error');
    err.classList.add('hidden');
    const payload = {
      name: $('#tpl-name').value.trim(),
      channel: $('#tpl-channel').value,
      subject: $('#tpl-subject').value.trim(),
      body: $('#tpl-body').value
    };
    const id = $('#tpl-id').value;
    try {
      if (id) await api('/api/templates/' + id, { method: 'PUT', body: payload });
      else await api('/api/templates', { method: 'POST', body: payload });
      toast('Template saved.', 'success');
      form.classList.add('hidden');
      loadTemplates();
    } catch (e) {
      err.textContent = e.message;
      err.classList.remove('hidden');
    }
  });

  // custom two-click confirm (NOT a native alert — different pattern)
  $('#tpl-rows').addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-action]');
    if (!btn) return;
    const tr = btn.closest('tr');
    const id = parseInt(tr.dataset.id, 10);

    if (btn.dataset.action === 'edit') {
      const d = await api('/api/templates');
      const tpl = d.items.find(t => t.id === id);
      if (tpl) openForm(tpl);
      return;
    }

    if (btn.dataset.action === 'delete') {
      if (btn.dataset.armed) {
        try {
          await api('/api/templates/' + id, { method: 'DELETE' });
          tr.remove();
          toast('Template #' + id + ' deleted.', 'success');
        } catch (err) { toast(err.message, 'error'); }
      } else {
        btn.dataset.armed = '1';
        btn.textContent = 'Confirm delete?';
        setTimeout(() => {
          if (btn.isConnected) { delete btn.dataset.armed; btn.textContent = 'Delete'; }
        }, 4000);
      }
    }
  });

  loadTemplates();
});