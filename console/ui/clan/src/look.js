var clCountries = null;
function loadCountries() {
  if (clCountries) return Promise.resolve(clCountries);
  var out = [];
  function page(n) {
    return api('GET', '/countries' + qs({ isActive: 'true', limit: 100, page: n })).then(function (d) {
      out = out.concat(d.items || []);
      if ((d.items || []).length === 100 && out.length < (d.total || 0) && n < 5) return page(n + 1);
      return out;
    });
  }
  return page(1).then(function (list) { clCountries = list; return list; }, function () { clCountries = out; return out; });
}

function countryName(code) {
  if (!code) return '';
  var c = (clCountries || []).filter(function (x) { return x.code === code; })[0];
  return c ? ((c.flag ? c.flag + ' ' : '') + (typeof c.name === 'object' ? textOf(c.name, code) : c.name || code)) : code;
}

function tagName(key) {
  var list = (clClanSettings && clClanSettings.profile && clClanSettings.profile.tags) || [];
  var x = list.filter(function (g) { return g.key === key; })[0];
  return x ? textOf(x.name, key) : key;
}

function lookHeader(c) {
  var cover = h('div', { class: 'cl-cover', 'data-clan-cover': c.coverKey || '' });
  if (c.coverUrl) cover.appendChild(h('img', { src: c.coverUrl, alt: '' }));
  if (c.color) cover.style.backgroundColor = c.color;
  var emblem = h('div', { class: 'cl-emblem', 'data-clan-emblem': c.emblemKey || '' });
  if (c.emblemUrl || c.avatarUrl) emblem.appendChild(h('img', { src: c.emblemUrl || c.avatarUrl, alt: '' }));
  else emblem.appendChild(h('span', { text: (c.name || '?').slice(0, 1) }));
  var pend = c.pendingUploads || {};
  var waiting = ['emblem', 'cover'].filter(function (k) { return pend[k]; });
  return h('div', { class: 'cl-look mb-4', 'data-clan-look': '' }, cover,
    h('div', { class: 'cl-look-row' }, emblem,
      h('div', { class: 'flex flex-col gap-1 min-w-0 flex-1' },
        h('span', { class: 'font-semibold text-lg truncate', dir: 'auto', text: c.name }),
        c.motto ? h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', dir: 'auto', 'data-clan-value': 'motto', text: '«' + c.motto + '»' }) : null,
        h('div', { class: 'flex flex-wrap gap-1 items-center' },
          c.color ? h('span', { class: 'cl-swatch', 'data-clan-value': 'color', title: c.color, 'aria-label': t('lk.color') + ' ' + c.color, style: 'background:' + c.color }) : null,
          c.country ? tag(countryName(c.country), 'secondary') : null,
          c.lang ? tag(c.lang, 'secondary') : null,
          (c.tags || []).map(function (k) { return tag(tagName(k), 'primary'); })),
        waiting.length ? h('span', { class: 'text-xs text-warning-700 dark:text-warning-400', text: t('lk.waiting', { what: waiting.map(function (k) { return t('ga.kind.' + k); }).join('، ') }) }) : null,
        c.nextChangeAt && new Date(c.nextChangeAt) > new Date() ? h('span', { class: 'text-xs text-secondary-500', text: t('lk.next', { at: fmtDate(c.nextChangeAt) }) }) : null)));
}

function artSelect(kind, value, name) {
  var list = (clArt || []).filter(function (a) { return a.kind === kind; });
  var opts = [{ value: '', label: t('none') }].concat(list.map(function (a) {
    return { value: a.key, label: textOf(a.name, a.key) + ' · ' + a.key + (a.isActive ? '' : ' (' + t('inactive') + ')') + (a.minLevel > 1 ? ' · ' + t('ga.fromLevel', { n: a.minLevel }) : '') };
  }));
  if (value && !list.some(function (a) { return a.key === value; })) opts.push({ value: value, label: value + ' (' + t('missing') + ')' });
  return select(value || '', opts, { name: name, label: name });
}

function lookForm(c, clanId, changed) {
  var prof = (clClanSettings && clClanSettings.profile) || {};
  var emblem = artSelect('emblem', c.emblemKey, 'emblemKey');
  var cover = artSelect('cover', c.coverKey, 'coverKey');
  var color = colorField(c.color);
  color.input.setAttribute('name', 'look-color');
  var swatches = h('div', { class: 'flex flex-wrap gap-1 mt-1' }, (prof.colors || []).map(function (hex) {
    return h('button', { type: 'button', class: 'cl-swatch cl-swatch--pick', title: hex, 'aria-label': hex, style: 'background:' + hex, onclick: function () { color.input.value = hex; color.input.dispatchEvent(new Event('input')); } });
  }));
  var motto = input(c.motto, { name: 'motto' });
  var countries = [{ value: '', label: t('none') }].concat((clCountries || []).map(function (x) { return { value: x.code, label: countryName(x.code) + ' · ' + x.code }; }));
  if (c.country && !(clCountries || []).some(function (x) { return x.code === c.country; })) countries.push({ value: c.country, label: c.country });
  var country = (clCountries || []).length ? select(c.country || '', countries, { name: 'country', label: t('lk.country') }) : input(c.country, { name: 'country', ltr: true, mono: true, placeholder: 'IR' });
  var langs = [{ value: '', label: t('none') }].concat((clLanguages || []).map(function (l) { return { value: l.code, label: l.name + ' · ' + l.code }; }));
  if (c.lang && !(clLanguages || []).some(function (l) { return l.code === c.lang; })) langs.push({ value: c.lang, label: c.lang });
  var lang = select(c.lang || '', langs, { name: 'lang', label: t('cl.lang') });
  var tagBoxes = (prof.tags || []).map(function (g) {
    var b = check((c.tags || []).indexOf(g.key) !== -1, textOf(g.name, g.key), 'tag');
    b.input.value = g.key;
    return b;
  });
  var tagBox = tagBoxes.length ? h('div', { class: 'flex flex-wrap gap-x-4', 'data-clan-tags': '' }, tagBoxes) : h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('lk.noTags') });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-look' }, onclick: function () {
    clear(msg);
    busy(save, true);
    api('PUT', '/clans/list/' + clanId, { emblemKey: emblem.value, coverKey: cover.value, color: color.input.value.trim(), motto: motto.value.trim(), country: String(country.value || '').trim().toUpperCase(),
      lang: lang.value, tags: tagBoxes.filter(function (b) { return b.input.checked; }).map(function (b) { return b.input.value; }) }).then(function (out) {
      toast('success', t('saved')); changed(out);
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var reset = button({ label: t('lk.reset'), variant: 'danger', data: { 'clan-action': 'reset-look' }, onclick: function () {
    confirmBox(t('lk.resetConfirm', { name: c.name }), t('lk.reset')).then(function (ok) {
      if (!ok) return;
      busy(reset, true);
      api('POST', '/clans/list/' + clanId + '/reset-look', {}).then(function (out) { toast('success', t('lk.resetDone')); changed(out); }, function (err) { busy(reset, false); fieldError(msg, err); });
    });
  } });
  return section(t('lk.title'), h('div', { 'data-clan-form': 'look' },
    h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('lk.hint') }),
    grid(field(t('lk.emblem'), emblem), field(t('lk.cover'), cover)),
    grid(group(t('lk.color'), h('div', {}, color, swatches)), field(t('lk.motto'), motto, prof.mottoMax ? t('lk.hint.motto', { n: prof.mottoMax }) : null)),
    grid(field(t('lk.country'), country), field(t('cl.lang'), lang)),
    group(t('lk.tags'), tagBox, prof.maxTags ? t('lk.hint.tags', { n: prof.maxTags }) : null),
    msg, h('div', { class: 'flex flex-wrap justify-between gap-2' }, reset, save)), 'look');
}
