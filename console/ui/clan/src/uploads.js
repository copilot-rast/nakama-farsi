var clPendingUploads = 0;
var clPendingAt = 0;
function refreshPendingUploads(force) {
  var now = Date.now();
  if (!force && now - clPendingAt < 60000) return;
  clPendingAt = now;
  var tk = token();
  if (!tk) return;
  fetch(CL_API + '/clans/uploads?status=pending&limit=1', { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    return res.ok ? res.json() : null;
  }).then(function (d) {
    var n = (d && d.total) || 0;
    if (n !== clPendingUploads) { clPendingUploads = n; clSchedule(); }
  }, function () { return null; });
}

function uploadBlob(id) {
  var tk = token();
  return fetch(CL_API + '/clans/uploads/' + id + '/file', { headers: tk ? { Authorization: 'Bearer ' + tk } : {}, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) throw new Error('HTTP_' + res.status);
    return res.blob();
  }).then(function (b) { return URL.createObjectURL(b); });
}

function reasonBox() {
  return new Promise(function (resolve) {
    var root;
    var area = input('', { name: 'reason', multiline: true, rows: 3, placeholder: t('up.reasonHint') });
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-clan': 'confirm', 'data-l10n-skip': '', class: 'cl-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'cl-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true', 'aria-label': t('up.reject') },
        h('h3', { class: 'font-semibold mb-2', text: t('up.reject') }),
        field(t('up.reason'), area, t('up.reasonNote')),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(null); } }),
          button({ label: t('up.reject'), variant: 'danger', data: { 'clan-confirm': 'ok' }, onclick: function () { done(area.value.trim()); } }))));
    document.body.appendChild(root);
    area.focus();
  });
}

var CL_UPLOAD_COLOR = { pending: 'warning', approved: 'success', rejected: 'error', replaced: 'secondary' };

function uploadsPage(root) {
  var state = { page: 1, limit: 24, total: 0, status: 'pending' };
  var frame = pageFrame('clans/uploads', t('desc.clans/uploads'));
  clear(root).appendChild(frame.el);
  var tabs = h('div', { class: 'cl-seg', role: 'tablist', 'aria-label': t('f.state') });
  ['pending', 'approved', 'rejected', 'replaced', 'all'].forEach(function (s) {
    var b = h('button', { type: 'button', role: 'tab', class: 'cl-seg-item', 'data-clan-status': s, 'aria-selected': s === state.status ? 'true' : 'false', text: t('up.st.' + s), onclick: function () {
      state.status = s; state.page = 1;
      Array.prototype.forEach.call(tabs.children, function (x) { x.setAttribute('aria-selected', x === b ? 'true' : 'false'); });
      reload();
    } });
    tabs.appendChild(b);
  });
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(tabs);
  frame.content.appendChild(box);
  var urls = [];
  function reload() {
    urls.forEach(function (u) { URL.revokeObjectURL(u); });
    urls = [];
    clear(box).appendChild(loading());
    api('GET', '/clans/uploads' + qs({ status: state.status, page: state.page, limit: state.limit })).then(function (d) {
      state.total = d.total || 0;
      var items = d.items || [];
      clear(box);
      if (!items.length) box.appendChild(emptyCard(t('up.empty.' + state.status)));
      else box.appendChild(h('div', { class: 'cl-gallery cl-uploads', 'data-clan-list': 'uploads' }, items.map(function (u) {
        var img = h('div', { class: 'cl-art-img' }, h('span', { class: 'text-xs text-secondary-500', text: t('loading') }));
        uploadBlob(u.id).then(function (src) { urls.push(src); clear(img).appendChild(h('img', { src: src, alt: t('up.previewOf', { name: u.clanName }) })); }, function () { clear(img).appendChild(h('span', { class: 'text-xs text-secondary-500', text: t('ga.noImage') })); });
        var actions = h('div', { class: 'flex flex-wrap gap-2 mt-2' });
        if (u.status === 'pending') {
          var ok = button({ label: t('up.approve'), variant: 'primary', data: { 'clan-action': 'approve' }, onclick: function () {
            busy(ok, true);
            api('POST', '/clans/uploads/' + u.id + '/approve', {}).then(function () { toast('success', t('up.approved')); refreshPendingUploads(true); reload(); }, function (err) { busy(ok, false); toast('error', clErrText(err)); });
          } });
          var no = button({ label: t('up.reject'), variant: 'danger', data: { 'clan-action': 'reject' }, onclick: function () {
            reasonBox().then(function (reason) {
              if (reason === null) return;
              busy(no, true);
              api('POST', '/clans/uploads/' + u.id + '/reject', { reason: reason }).then(function () { toast('success', t('up.rejected')); refreshPendingUploads(true); reload(); }, function (err) { busy(no, false); toast('error', clErrText(err)); });
            });
          } });
          actions.appendChild(ok);
          actions.appendChild(no);
        }
        return h('div', { class: 'cl-art cl-upload' + (u.kind === 'cover' ? ' cl-art--cover' : ''), 'data-clan-upload': u.id }, img,
          h('div', { class: 'cl-art-body' },
            h('a', { href: '#/clans/list', class: 'eco-link font-medium truncate', dir: 'auto', text: u.clanName || u.clanId.slice(0, 8), onclick: function (e) { e.preventDefault(); clanPanel(u.clanId, u.clanName); } }),
            h('div', { class: 'flex flex-wrap gap-1' }, tag(t('ga.kind.' + u.kind), 'secondary'), tag(t('up.st.' + u.status), CL_UPLOAD_COLOR[u.status] || 'secondary')),
            h('span', { class: 'text-xs text-secondary-500', text: t('up.by', { name: u.username || '—' }) + ' · ' + fmtDate(u.createdAt) }),
            h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: t('up.size', { n: fmtNum(Math.max(1, Math.round(u.size / 1024))) }) + ' · ' + (u.contentType || '') }),
            u.reviewedBy ? h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: t('up.reviewed', { name: u.reviewedBy }) + (u.reviewedAt ? ' · ' + fmtDate(u.reviewedAt) : '') }) : null,
            u.reason ? h('span', { class: 'text-xs text-error-600 dark:text-error-400', dir: 'auto', 'data-clan-reason': '', text: u.reason }) : null,
            actions));
      })));
      box.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}
