var clArt = null;
function loadArt(force) {
  if (clArt && !force) return Promise.resolve(clArt);
  return api('GET', '/clans/art').then(function (d) { clArt = d.items || []; return clArt; });
}

function artOf(kind, key) {
  return (clArt || []).filter(function (a) { return a.kind === kind && a.key === key; })[0] || null;
}

function uploadImage(file) {
  var fd = new FormData();
  fd.append('file', file, file.name);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/economy/uploads', { method: 'POST', headers: headers, body: fd, credentials: 'same-origin' }).then(function (res) {
    return res.text().then(function (text) {
      var data = null;
      try { data = text ? JSON.parse(text) : null; } catch (e) { data = null; }
      if (!res.ok) throw new ClError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', 'file', null);
      return data || {};
    });
  });
}

function imageField(value, name) {
  var url = input(value || '', { name: name, ltr: true, placeholder: 'https://…' });
  var preview = h('div', { class: 'cl-thumb', 'data-clan-preview': name });
  var file = h('input', { type: 'file', accept: 'image/png,image/jpeg,image/webp', class: 'sr-only', name: name + '-file', 'aria-label': t('ga.upload') });
  var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400', hidden: true });
  var up = button({ label: t('ga.upload'), data: { 'clan-action': 'upload-image' }, onclick: function () { file.click(); } });
  function paint() {
    clear(preview);
    var v = url.value.trim();
    if (v) preview.appendChild(h('img', { src: v, alt: '' }));
    else preview.appendChild(h('span', { class: 'text-xs text-secondary-500', text: t('ga.noImage') }));
  }
  url.addEventListener('change', paint);
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    err.hidden = true;
    busy(up, true);
    uploadImage(f).then(function (res) {
      url.value = res.url || '';
      paint();
    }, function (e) {
      err.textContent = clErrText(e);
      err.hidden = false;
    }).then(function () { busy(up, false); file.value = ''; });
  });
  paint();
  var wrap = h('div', { class: 'flex gap-3 items-start' }, preview, h('div', { class: 'flex-1 min-w-0 flex flex-col gap-2' }, url, h('div', {}, up, file), err));
  wrap.input = url;
  return wrap;
}

function priceText(a) {
  return a.priceAmount ? t('ga.price', { n: fmtNum(a.priceAmount) }) + ' ' + a.priceCurrency : t('ga.free');
}

function galleryPage(root) {
  var state = { kind: '' };
  var frame = pageFrame('clans/gallery', t('desc.clans/gallery'), [
    button({ label: t('ga.new'), variant: 'primary', data: { 'clan-action': 'new-art' }, onclick: function () { artEditor(null, state.kind || 'emblem', reload); } })]);
  clear(root).appendChild(frame.el);
  var tabs = h('div', { class: 'cl-seg', role: 'tablist', 'aria-label': t('ga.kind') });
  ['', 'emblem', 'cover'].forEach(function (k) {
    var b = h('button', { type: 'button', role: 'tab', class: 'cl-seg-item', 'data-clan-kind': k || 'all', 'aria-selected': k === state.kind ? 'true' : 'false', text: k ? t('ga.kind.' + k) : t('all'), onclick: function () {
      state.kind = k;
      Array.prototype.forEach.call(tabs.children, function (x) { x.setAttribute('aria-selected', x === b ? 'true' : 'false'); });
      paint();
    } });
    tabs.appendChild(b);
  });
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(tabs);
  frame.content.appendChild(box);
  function paint() {
    var items = (clArt || []).filter(function (a) { return !state.kind || a.kind === state.kind; });
    clear(box);
    if (!items.length) { box.appendChild(emptyCard(t('ga.empty'))); return; }
    box.appendChild(h('div', { class: 'cl-gallery', 'data-clan-list': 'art' }, items.map(function (a) {
      var cardEl = h('button', { type: 'button', class: 'cl-art' + (a.kind === 'cover' ? ' cl-art--cover' : ''), 'data-clan-art': a.kind + ':' + a.key, onclick: function () { artEditor(a, a.kind, reload); } },
        h('div', { class: 'cl-art-img' }, a.imageUrl ? h('img', { src: a.imageUrl, alt: '', loading: 'lazy' }) : null),
        h('div', { class: 'cl-art-body' },
          h('span', { class: 'font-medium truncate', dir: 'auto', text: textOf(a.name, a.key) }), mono(a.key),
          h('div', { class: 'flex flex-wrap gap-1 mt-1' }, tag(t('ga.kind.' + a.kind), a.kind === 'cover' ? 'primary' : 'secondary'),
            a.isDefault ? tag(t('ga.default'), 'success') : null, !a.isActive ? tag(t('inactive'), 'warning') : null,
            a.minLevel > 1 ? tag(t('ga.fromLevel', { n: a.minLevel }), 'secondary') : null, a.priceAmount ? tag(priceText(a), 'warning') : null),
          h('span', { class: 'text-xs text-secondary-500 mt-1', text: t('ga.clans', { n: fmtNum(a.clans || 0) }) })));
      return cardEl;
    })));
  }
  function reload() {
    clear(box).appendChild(loading());
    Promise.all([loadArt(true), loadLanguages(), loadRefs('clans')]).then(paint, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function artEditor(a, kind, done) {
  var creating = !a;
  a = a || { kind: kind, name: {}, imageUrl: '', sort: 0, isActive: true, minLevel: 1, priceCurrency: '', priceAmount: 0, isDefault: false };
  var kindSel = select(a.kind, opt('ga.kind.', ['emblem', 'cover']), { name: 'kind' });
  var key = input(a.key || '', { name: 'key', ltr: true, mono: true });
  var name = translated(t('f.name'), 'name', a.name);
  var img = imageField(a.imageUrl, 'imageUrl');
  var sort = numInput(a.sort || 0, 'sort');
  var active = check(a.isActive, t('f.active'), 'isActive');
  var minLevel = numInput(a.minLevel || 1, 'minLevel');
  var cur = refSelect(refList('clans', 'currencies'), a.priceCurrency, 'priceCurrency', t('ga.noPrice'));
  var amount = numInput(a.priceAmount || 0, 'priceAmount');
  var def = check(a.isDefault, t('ga.isDefault'), 'isDefault');
  var msg = h('div', {});
  var body = h('div', { 'data-clan-form': 'art' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('ga.hint') }),
    grid(field(t('ga.kind'), kindSel), field(t('f.key'), key, t('hint.key'))), name,
    group(t('ga.image'), img, t('ga.hint.image')),
    grid(field(t('ga.minLevel'), minLevel, t('ga.hint.minLevel')), field(t('f.sort'), sort)),
    grid(field(t('ga.currency'), cur), field(t('ga.amount'), amount, t('ga.hint.price'))),
    active, def, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ga.hint.default') }), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-art' }, onclick: function () {
    clear(msg);
    var out = { kind: kindSel.value, key: key.value.trim(), name: name.read(), imageUrl: img.input.value.trim(), sort: intOr(sort, 0), isActive: active.input.checked,
      minLevel: intOr(minLevel, 1), priceCurrency: cur.value, priceAmount: cur.value ? intOr(amount, 0) : 0, isDefault: def.input.checked };
    busy(save, true);
    api(creating ? 'POST' : 'PUT', '/clans/art' + (creating ? '' : '/' + a.id), out).then(function (d) {
      clArt = d.items || []; closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (!creating) {
    var del = button({ label: t('delete'), variant: 'danger', data: { 'clan-action': 'delete-art' }, onclick: function () {
      confirmBox(t('ga.delete', { key: a.key }), t('delete')).then(function (ok) {
        if (!ok) return;
        busy(del, true);
        api('DELETE', '/clans/art/' + a.id).then(function () { closePanel(); toast('success', t('deleted')); done(); }, function (err) {
          busy(del, false);
          clear(msg).appendChild(notice('error', err.code === 'IN_USE' ? t('ga.inUse', { n: (err.meta && err.meta.clans) || 0 }) : clErrText(err)));
        });
      });
    } });
    foot = [del, h('div', { class: 'flex-1' })].concat(foot);
  }
  sidePanel(creating ? t('ga.new') : textOf(a.name, a.key), body, foot, 'art');
}
