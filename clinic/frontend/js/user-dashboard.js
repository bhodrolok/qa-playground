document.addEventListener('DOMContentLoaded', async () => {
  const user = requireRole('user');
  if (!user) return;
  $('#user-greeting').textContent = `Welcome, ${user.name}`;
  $('#user-email-top').textContent = user.email;

  // ---- profile ----
  api('/api/me').then(me => {
    $('#profile-name').textContent = me.name;
    $('#profile-email').textContent = me.email;
    $('#member-since').textContent = new Date(me.createdAt).toLocaleDateString();
  }).catch(e => toast(e.message, 'error'));

  $('#btn-logout-user').addEventListener('click', logout);

  // ---- booking: chained dynamic loading ----
  let selectedSlot = null;
  const slotBtns = $('#slot-buttons');

  $('#book-date').addEventListener('change', async () => {
    selectedSlot = null;
    $('#btn-book').disabled = true;
    slotBtns.innerHTML = '';
    const date = $('#book-date').value;
    if (!date) return;

    $('#slots-hint').classList.add('hidden');
    $('#slots-spinner').classList.remove('hidden');
    try {
      const d = await api('/api/slots?date=' + encodeURIComponent(date));
      $('#slots-spinner').classList.add('hidden');
      slotBtns.innerHTML = '';
      d.slots.forEach(s => {
        const b = document.createElement('button');
        b.type = 'button';
        b.className = 'slot';
        b.textContent = s.time;
        if (s.available) {
          b.addEventListener('click', () => {
            $$('.slot.selected').forEach(x => x.classList.remove('selected'));
            b.classList.add('selected');
            selectedSlot = s.time;
            $('#btn-book').disabled = false;
          });
        } else {
          b.disabled = true;
        }
        slotBtns.appendChild(b);
      });
      if (!d.slots.some(s => s.available)) {
        slotBtns.insertAdjacentHTML('beforeend', '<p class="muted">No slots available on this date.</p>');
      }
    } catch (e) {
      $('#slots-spinner').classList.add('hidden');
      $('#slots-hint').textContent = 'Failed to load slots: ' + e.message;
      $('#slots-hint').classList.remove('hidden');
    }
  });

  $('#btn-book').addEventListener('click', async () => {
    const err = $('#book-error');
    err.classList.add('hidden');
    if (!selectedSlot) return;
    // note: intentionally no disable/debounce during submit
    try {
      const a = await api('/api/me/appointments', {
        method: 'POST',
        body: {
          date: $('#book-date').value,
          time: selectedSlot,
          reason: $('#book-reason').value.trim()
        }
      });
      toast(`Booked! Appointment #${a.id} on ${a.date} at ${a.time}.`, 'success');
      // note: the appointments table below is intentionally NOT refreshed here
    } catch (e) {
      err.textContent = e.message;
      err.classList.remove('hidden');
    }
  });

  // ---- my appointments ----
  async function loadMyAppts() {
    const body = $('#my-appts-body');
    body.innerHTML = `<tr><td colspan="5" class="muted">Loading…</td></tr>`;
    try {
      const d = await api('/api/me/appointments');
      body.innerHTML = '';
      if (!d.items.length) {
        body.innerHTML = `<tr><td colspan="5" class="muted">No appointments yet. Book one above.</td></tr>`;
        return;
      }
      d.items.forEach(a => {
        const tr = document.createElement('tr');
        tr.dataset.id = a.id;
        tr.innerHTML = `
          <td>${esc(a.date)}</td>
          <td>${esc(a.time)}</td>
          <td>${esc(a.reason)}</td>
          <td><span class="badge badge-${esc(a.status)}">${esc(a.status)}</span></td>
          <td>${a.status === 'cancelled' ? '' : `<button class="btn btn-small" data-action="cancel-appt">Cancel</button>`}</td>`;
        body.appendChild(tr);
      });
    } catch (e) {
      body.innerHTML = `<tr><td colspan="5" class="error">${esc(e.message)}</td></tr>`;
    }
  }

  $('#btn-refresh-appts').addEventListener('click', loadMyAppts);

  $('#my-appts-body').addEventListener('click', async (e) => {
    const btn = e.target.closest('button[data-action="cancel-appt"]');
    if (!btn) return;
    const tr = btn.closest('tr');
    const id = parseInt(tr.dataset.id, 10);
    try {
      const a = await api(`/api/me/appointments/${id}/cancel`, { method: 'POST' });
      const badge = tr.querySelector('.badge');
      badge.className = `badge badge-${a.status}`;
      badge.textContent = a.status;
      btn.remove();
      toast(`Appointment #${id} cancelled.`, 'success');
    } catch (err) { toast(err.message, 'error'); }
  });

  loadMyAppts();

  // ---- notifications (polling) ----
  async function pollNotifs() {
    try {
      const d = await api('/api/notifications');
      const list = $('#notif-list');
      list.innerHTML = '';
      d.items.forEach(n => {
        const li = document.createElement('li');
        li.textContent = n.title + ' — ' + n.body;
        list.appendChild(li);
      });
    } catch (e) { /* transient */ }
  }
  pollNotifs();
  setInterval(pollNotifs, 12000);
});