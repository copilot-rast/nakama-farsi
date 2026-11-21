var schemaQuery = { module: '', q: '' };

function schemaSettings(cur) {
  var access = select(cur.access, ['public', 'session', 'apikey', 'console'].map(function (v) { return { value: v, label: t('set.schema.access.' + v) }; }), { name: 'access' });
  return { el: field(t('set.schema.access'), access), read: function () { return { access: access.value }; } };
}

function shape(s, depth) {
  depth = depth || 0;
  if (!s || typeof s !== 'object' || !s.type) return 'any';
  var pad = new Array(depth + 2).join('  ');
  var end = new Array(depth + 1).join('  ');
  if (s.type === 'object' && s.properties) {
    var keys = Object.keys(s.properties).sort();
    if (!keys.length) return '{}';
    return '{\n' + keys.map(function (k) { return pad + k + ': ' + shape(s.properties[k], depth + 1); }).join(',\n') + '\n' + end + '}' + (s.nullable ? ' | null' : '');
  }
  if (s.type === 'object' && s.additionalProperties) return '{ [key]: ' + shape(s.additionalProperties, depth + 1) + ' }';
  if (s.type === 'array') return shape(s.items, depth) + '[]' + (s.nullable ? ' | null' : '');
  return s.type + (s.format ? ' (' + s.format + ')' : '') + (s.nullable ? ' | null' : '');
}

function schemaPage(root) {
  var doc = null;
  var download = button({ label: t('sc.download'), data: { 'devkit-action': 'download-schema' }, onclick: function () {
    if (!doc) return;
    var blob = new Blob([JSON.stringify(doc, null, 2)], { type: 'application/json' });
    var a = h('a', { href: URL.createObjectURL(blob), download: 'rpc-schema.json' });
    document.body.appendChild(a);
    a.click();
    setTimeout(function () { URL.revokeObjectURL(a.href); a.remove(); }, 1000);
  } });
  var frame = pageFrame('rpc-schema', t('desc.rpc-schema'), [download]);
  clear(root).appendChild(frame.el);
  var bar = h('div', { class: 'flex flex-wrap gap-2' });
  var list = h('div', { 'data-devkit-list': 'rpc-schema', class: 'flex flex-col gap-2' });
  frame.content.appendChild(settingsCard('schema', schemaSettings));
  frame.content.appendChild(bar);
  frame.content.appendChild(list);

  api('GET', '/rpc-schema').then(function (d) {
    doc = d;
    var mods = {};
    d.rpcs.forEach(function (r) { mods[r.module] = true; });
    var search = input(schemaQuery.q, { placeholder: t('search'), name: 'q', label: t('search') });
    search.addEventListener('input', function () { schemaQuery.q = search.value.trim().toLowerCase(); paint(); });
    var mod = select(schemaQuery.module, [{ value: '', label: t('sc.module') + ': ' + t('all') }].concat(Object.keys(mods).sort().map(function (m) { return { value: m, label: m }; })), { name: 'module', label: t('sc.module') });
    mod.addEventListener('change', function () { schemaQuery.module = mod.value; paint(); });
    clear(bar).appendChild(h('div', { class: 'flex flex-wrap gap-2 w-full' }, h('div', { class: 'flex-1 min-w-[12rem]' }, search), mod));
    paint();
  }, function (err) { clear(list).appendChild(notice('error', t('error.load') + ' ' + errorText(err))); });

  function paint() {
    var items = doc.rpcs.filter(function (r) {
      return (!schemaQuery.module || r.module === schemaQuery.module) && (!schemaQuery.q || r.id.indexOf(schemaQuery.q) !== -1);
    });
    clear(list).appendChild(h('p', { class: 'text-xs text-secondary-500', 'data-devkit-count': '', text: t('sc.count', { n: items.length }) }));
    items.forEach(function (r) {
      var det = h('details', { class: 'dk-rpc bg-white dark:bg-secondary-800 rounded-md shadow', 'data-devkit-rpc': r.id },
        h('summary', { class: 'cursor-pointer px-4 py-2 flex flex-wrap items-center gap-2' }, mono(r.id), tag(r.module), tag(t('sc.format.' + r.errorFormat), r.errorFormat === 'envelope' ? 'success' : 'secondary')),
        h('div', { class: 'px-4 pb-4 grid md:grid-cols-2 gap-4' },
          h('div', {}, h('h3', { class: 'text-xs font-semibold mb-1', text: t('sc.request') }), h('pre', { class: 'dk-pre', dir: 'ltr', text: shape(r.request) })),
          h('div', {}, h('h3', { class: 'text-xs font-semibold mb-1', text: t('sc.response') }), h('pre', { class: 'dk-pre', dir: 'ltr', text: shape(r.response) })),
          h('div', { class: 'md:col-span-2 flex flex-wrap gap-1 items-center' }, h('span', { class: 'text-xs font-semibold me-2', text: t('sc.errors') }), (r.errors || []).map(function (c) { return tag(c, 'warning'); })),
          h('div', { class: 'md:col-span-2 text-xs text-secondary-500' }, t('sc.source') + ': ', mono(r.source + ' · ' + r.handler))));
      list.appendChild(det);
    });
  }
}
