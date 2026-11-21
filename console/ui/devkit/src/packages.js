function packagesSettings(cur, defs) {
  var enabled = check(cur.enabled, t('set.packages.enabled'), 'enabled');
  var access = select(cur.access, ['apikey', 'public'].map(function (v) { return { value: v, label: t('set.packages.access.' + v) }; }), { name: 'access' });
  var prefixes = input((cur.namePrefixes || []).join(', '), { ltr: true, mono: true, name: 'namePrefixes', placeholder: 'com.example.' });
  var max = input(cur.maxPackageMB, { ltr: true, name: 'maxPackageMB', placeholder: String(defs.maxPackageMB) });
  var keep = input(cur.keepVersions, { ltr: true, name: 'keepVersions', placeholder: '0' });
  var url = input(cur.publicUrl, { ltr: true, name: 'publicUrl', placeholder: 'https://api.example.com/upm' });
  return {
    el: h('div', { class: 'grid md:grid-cols-2 gap-x-6' },
      h('div', {}, field(t('set.packages.access'), access), field(t('set.packages.prefixes'), prefixes, t('set.packages.prefixes.hint')),
        field(t('set.packages.url'), url, t('set.packages.url.hint'))),
      h('div', {}, enabled, field(t('set.packages.max'), max), field(t('set.packages.keep'), keep, t('set.packages.keep.hint')))),
    read: function () {
      var out = JSON.parse(JSON.stringify(cur));
      out.enabled = enabled.input.checked;
      out.access = access.value;
      out.namePrefixes = prefixes.value.split(/[,\s]+/).map(function (s) { return s.trim(); }).filter(Boolean);
      var m = numberValue(max), k = numberValue(keep);
      out.maxPackageMB = m == null ? defs.maxPackageMB : m;
      out.keepVersions = k == null ? 0 : k;
      out.publicUrl = url.value.trim();
      return out;
    },
    after: function () { dkMounted = null; dkSchedule(); }
  };
}

function snippet(title, text) {
  var pre = h('pre', { class: 'dk-pre', dir: 'ltr', text: text });
  var copy = button({ label: t('pk.copy'), onclick: function () {
    if (navigator.clipboard) navigator.clipboard.writeText(text).then(function () { toast('success', t('pk.copied')); });
  } });
  return h('div', { class: 'mb-3' }, h('div', { class: 'flex items-center justify-between mb-1' }, h('h3', { class: 'text-xs font-semibold', text: title }), copy), pre);
}

function packagesPage(root) {
  var frame = pageFrame('packages', t('desc.packages'));
  clear(root).appendChild(frame.el);
  var info = h('div', {});
  var list = h('div', { 'data-devkit-list': 'packages', class: 'flex flex-col gap-2' });
  frame.content.appendChild(settingsCard('packages', packagesSettings));
  frame.content.appendChild(info);
  frame.content.appendChild(list);
  paint();

  function paint() {
    api('GET', '/packages').then(function (d) {
      var url = d.registryUrl;
      var names = {};
      d.versions.forEach(function (v) { names[v.name.split('.').slice(0, 2).join('.')] = true; });
      var scopes = Object.keys(names).sort();
      if (!scopes.length) scopes = ['com.example'];
      var latest = {};
      d.versions.forEach(function (v) { latest[v.name] = v.version; });
      var deps = {};
      Object.keys(latest).sort().forEach(function (n) { deps[n] = latest[n]; });
      clear(info).appendChild(card([h('div', { class: 'p-4' },
        d.enabled ? null : notice('warning', t('pk.disabled')),
        h('p', { class: 'text-sm mb-3' }, t('pk.address') + ' ', mono(url)),
        h('p', { class: 'text-xs text-secondary-500 mb-3', text: d.access === 'public' ? t('pk.public') : t('pk.keyHint', { read: d.readScope, publish: d.publishScope }) }),
        d.access === 'public' ? null : snippet(t('pk.upmconfig'), '[npmAuth."' + url + '"]\ntoken = "<API key>"\nalwaysAuth = true\n'),
        snippet(t('pk.manifest'), JSON.stringify({ scopedRegistries: [{ name: t('pk.registryName'), url: url, scopes: scopes }], dependencies: deps }, null, 2)))]));
      clear(list);
      if (!d.versions.length) { list.appendChild(notice('info', t('pk.empty'))); return; }
      list.appendChild(table([
        { label: t('pk.name') }, { label: t('pk.version') }, { label: t('pk.size') }, { label: t('pk.by') }, { label: t('pk.at') }, { label: '' }
      ], d.versions.slice().reverse().map(function (v) {
        var del = button({ label: t('delete'), variant: 'danger', data: { 'devkit-package-delete': v.name + '@' + v.version }, onclick: function () {
          confirmBox(t('pk.deleteConfirm', { pkg: v.name + '@' + v.version }), t('delete')).then(function (ok) {
            if (!ok) return;
            api('DELETE', '/packages/' + encodeURIComponent(v.name) + '/' + encodeURIComponent(v.version)).then(function () { toast('success', t('pk.deleted')); paint(); },
              function (err) { toast('error', errorText(err)); });
          });
        } });
        return { id: v.name + '@' + v.version, cells: [h('div', {}, mono(v.name), v.displayName ? h('div', { class: 'text-xs text-secondary-500', text: v.displayName }) : null),
          mono(v.version), Math.round(v.size / 1024) + ' KB', v.publishedBy || '—', fmtDate(v.createdAt), del] };
      })));
    }, function (err) { clear(list).appendChild(notice('error', t('error.load') + ' ' + errorText(err))); });
  }
}
