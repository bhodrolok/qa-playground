document.addEventListener('DOMContentLoaded', async () => {
  const user = requireRole('staff');
  if (!user) return;
  renderStaffChrome('dashboard', user);
  applyTheme();

  // ---- stats (slow endpoint: skeleton -> real values) ----
  api('/api/stats').then(st => {
    $('#stat-upcoming').textContent = st.upcomingAppointments;
    $('#stat-today').textContent = st.confirmedToday;
    $('#stat-patients').textContent = st.totalPatients;
    $('#stat-cancelled').textContent = st.cancelledTotal;
    $$('.stat-value').forEach(el => el.classList.remove('skeleton'));
  }).catch(e => toast(e.message, 'error'));

  // ---- live metrics (values change on every poll) ----
  async function refreshLive() {
    try {
      const d = await api('/api/stats/live');
      $('#live-queue').textContent = d.queueNow;
      $('#live-wait').textContent = d.avgWaitMinutes;
      $('#live-staff').textContent = d.onlineStaff;
      $('#live-updated').textContent = 'Updated ' + new Date().toLocaleTimeString();
    } catch (e) { /* keep last values */ }
  }
  refreshLive();
  setInterval(refreshLive, 8000);

  // ---- activity feed (items appear/disappear over time) ----
  let lastId = 0;
  const feed = $('#activity-feed');
  async function pollActivity() {
    try {
      const d = await api('/api/activity?after=' + lastId);
      if (d.events && d.events.length) {
        feed.querySelector('li.muted')?.remove();
        for (const ev of d.events) {
          const li = document.createElement('li');
          li.id = 'feed-item-' + ev.id;
          li.textContent = `${new Date(ev.timestamp).toLocaleTimeString()} — ${ev.message}`;
          feed.prepend(li);
          lastId = Math.max(lastId, ev.id);
        }
        while (feed.children.length > 15) feed.lastElementChild.remove();
      } else if (lastId === 0 && d.latest > 0) {
        lastId = d.latest;
      }
    } catch (e) { /* transient */ }
  }
  pollActivity();
  setInterval(pollActivity, 4000);

  // ---- recent appointments ----
  api('/api/appointments').then(d => {
    const body = $('#recent-body');
    body.innerHTML = '';
    d.items.slice(0, 5).forEach(a => {
      body.insertAdjacentHTML('beforeend', `<tr>
        <td>${esc(a.patientName)}</td>
        <td>${esc(a.date)}</td>
        <td>${esc(a.time)}</td>
        <td><span class="badge badge-${esc(a.status)}">${esc(a.status)}</span></td>
      </tr>`);
    });
  }).catch(e => toast(e.message, 'error'));
});