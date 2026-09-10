document.addEventListener('DOMContentLoaded', () => {
  const user = requireRole('staff');
  if (!user) return;
  renderStaffChrome('hours', user);
  applyTheme();

  async function load() {
    const body = $('#hours-body');
    body.innerHTML = `<tr><td colspan="4" class="muted">Loading…</td></tr>`;
    try {
      const d = await api('/api/clinic/hours');
      body.innerHTML = '';
      d.hours.forEach(h => {
        body.insertAdjacentHTML('beforeend', `<tr data-day="${h.day}">
          <td>${h.day}</td>
          <td><input type="time" class="hours-open" value="${h.open}" aria-label="${h.day} opens"></td>
          <td><input type="time" class="hours-close" value="${h.close}" aria-label="${h.day} closes"></td>
          <td><input type="checkbox" class="hours-closed" ${h.closed ? 'checked' : ''} aria-label="${h.day} closed"></td>
        </tr>`);
      });
    } catch (e) {
      body.innerHTML = `<tr><td colspan="4" class="error">${esc(e.message)}</td></tr>`;
    }
  }

  $('#btn-save-hours').addEventListener('click', async () => {
    const rows = $$('#hours-body tr');
    const hours = rows.map(tr => ({
      day: tr.dataset.day,
      open: tr.querySelector('.hours-open').value,
      close: tr.querySelector('.hours-close').value,
      closed: tr.querySelector('.hours-closed').checked
    }));
    try {
      await api('/api/clinic/hours', { method: 'PUT', body: { hours } });
      toast('Clinic hours saved.', 'success');
    } catch (e) {
      toast(e.message, 'error');
    }
  });

  load();
});