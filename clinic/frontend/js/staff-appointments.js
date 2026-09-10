document.addEventListener('DOMContentLoaded', () => {
  const user = requireRole('staff');
  if (!user) return;
  renderStaffChrome('appointments', user);
  applyTheme();

  const tbody = $('#appts-body');

  function renderRow(a) {
    const tr = document.createElement('tr');
    tr.dataset.id = a.id;
    tr.innerHTML = `
      <td>#${a.id}</td>
      <td>${esc(a.patientName)}</td>
      <td>${esc(a.date)}</td>
      <td>${esc(a.time)}</td>
      <td>${esc(a.reason)}</td>
      <td><span class="badge badge-${esc(a.status)}">${esc(a.status)}</span></td>
      <td class="att-cell">${a.attachmentId ? `<a class="download-link" href="/api/files/${a.attachmentId}">Download</a>` : '—'}</td>
      <td class="row-actions">
        <button class="btn btn-small" data-action="confirm">Confirm</button>
        <button class="btn btn-small" data-action="cancel">Cancel</button>
        <button class="btn btn-small" data-action="attach">Attach</button>
        <button class="btn btn-small btn-danger" data-action="delete">Delete</button>
      </td>`;
    tbody.appendChild(tr);
  }

  async function loadAppointments() {
    tbody.innerHTML = `<tr><td colspan="8" class="muted">Loading appointments…</td></tr>`;
    const status = $('#filter-status').value;
    const date = $('#filter-date').value;
    const qs = [];
    if (status) qs.push('status=' + encodeURIComponent(status));
    if (date) qs.push('date=' + encodeURIComponent(date));
    try {
      const d = await api('/api/appointments' + (qs.length ? '?' + qs.join('&') : ''));
      tbody.innerHTML = '';
      if (!d.items.length) {
        tbody.innerHTML = `<tr><td colspan="8" class="muted">No appointments found.</td></tr>`;
        return;
      }
      d.items.forEach(renderRow);
    } catch (e) {
      tbody.innerHTML = `<tr><td colspan="8" class="error">Failed to load: ${esc(e.message)}</td></tr>`;
    }
  }

  // ---- create form ----
  $('#btn-new-appt').addEventListener('click', () => {
    const f = $('#appt-form');
    f.classList.toggle('hidden');
    if (!f.classList.contains('hidden')) $('#appt-patient').focus();
  });
  $('#btn-cancel-form').addEventListener('click', () => $('#appt-form').classList.add('hidden'));

  $('#btn-save-appt').addEventListener('click', async () => {
    const err = $('#appt-form-error');
    err.classList.add('hidden');
    try {
      const a = await api('/api/appointments', {
        method: 'POST',
        body: {
          patientName: $('#appt-patient').value.trim(),
          date: $('#appt-date').value,
          time: $('#appt-time').value,
          reason: $('#appt-reason').value.trim()
        }
      });
      toast(`Appointment #${a.id} created.`, 'success');
      $('#appt-form').classList.add('hidden');
      $('#appt-patient').value = $('#appt-date').value = $('#appt-time').value = $('#appt-reason').value = '';
      loadAppointments();
    } catch (e) {
      err.textContent = e.message;
      err.classList.remove('hidden');
    }
  });

  $('#btn-apply-filter').addEventListener('click', loadAppointments);

  // ---- row actions (event delegation) ----
  tbody.addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-action]');
    if (!btn) return;
    const tr = btn.closest('tr');
    const id = parseInt(tr.dataset.id, 10);
    const action = btn.dataset.action;

    if (action === 'confirm' || action === 'cancel') {
      try {
        const a = await api(`/api/appointments/${id}/status`, {
          method: 'PUT',
          body: { status: action === 'confirm' ? 'confirmed' : 'cancelled' }
        });
        const badge = tr.querySelector('.badge');
        badge.className = `badge badge-${a.status}`;
        badge.textContent = a.status;
        toast(`Appointment #${id} → ${a.status}`, 'success');
        // note: action buttons intentionally stay enabled
      } catch (err) { toast(err.message, 'error'); }
      return;
    }

    if (action === 'delete') {
      // native browser confirm — an alert to handle in automation
      if (!window.confirm(`Delete appointment #${id}? This cannot be undone.`)) return;
      try {
        await api(`/api/appointments/${id}`, { method: 'DELETE' });
        tr.remove();
        toast(`Appointment #${id} deleted.`, 'success');
      } catch (err) { toast(err.message, 'error'); }
      return;
    }

    if (action === 'attach') {
      const input = $('#row-attach-input');
      input.value = '';
      input.onchange = async () => {
        const file = input.files[0];
        if (!file) return;
        const fd = new FormData();
        fd.append('file', file);
        try {
          const res = await api(`/api/appointments/${id}/attachment`, { method: 'POST', body: fd });
          tr.querySelector('.att-cell').innerHTML = `<a class="download-link" href="${res.url}">Download</a>`;
          toast(`File "${res.name}" attached to #${id}.`, 'success');
        } catch (err) { toast(err.message, 'error'); }
      };
      input.click();
      return;
    }
  });

  loadAppointments();
});