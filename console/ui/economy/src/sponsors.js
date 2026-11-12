PH_ICONS.sponsors = '<path d="M254.3,107.91,228.78,56.85a16,16,0,0,0-21.47-7.15L182.44,62.13,130.05,48.27a8.14,8.14,0,0,0-4.1,0L73.56,62.13,48.69,49.7a16,16,0,0,0-21.47,7.15L1.7,107.9a16,16,0,0,0,7.15,21.47l27,13.51,55.49,39.63a8.06,8.06,0,0,0,2.71,1.25l64,16a8,8,0,0,0,7.6-2.1l55.07-55.08,26.42-13.21a16,16,0,0,0,7.15-21.46Zm-54.89,33.37L165,113.72a8,8,0,0,0-10.68.61C136.51,132.27,116.66,130,104,122L147.24,80h31.81l27.21,54.41ZM41.53,64,62,74.22,36.43,125.27,16,115.06Zm116,119.13L99.42,168.61l-49.2-35.14,28-56L128,64.28l9.8,2.59-45,43.68-.08.09a16,16,0,0,0,2.72,24.81c20.56,13.13,45.37,11,64.91-5L188,152.66Zm62-57.87-25.52-51L214.47,64,240,115.06Zm-87.75,92.67a8,8,0,0,1-7.75,6.06,8.13,8.13,0,0,1-1.95-.24L80.41,213.33a7.89,7.89,0,0,1-2.71-1.25L51.35,193.26a8,8,0,0,1,9.3-13l25.11,17.94L126,208.24A8,8,0,0,1,131.82,217.94Z"/>';
PH_ICONS.sponsorDuo = '<path d="M200,152l-40,40L96,176,40,136,72.68,70.63,128,56l55.32,14.63L183.6,72H144L98.34,116.29a8,8,0,0,0,1.38,12.42C117.23,139.9,141,139.13,160,120Z" opacity="0.2"/><path d="M254.3,107.91,228.78,56.85a16,16,0,0,0-21.47-7.15L182.44,62.13,130.05,48.27a8.14,8.14,0,0,0-4.1,0L73.56,62.13,48.69,49.7a16,16,0,0,0-21.47,7.15L1.7,107.9a16,16,0,0,0,7.15,21.47l27,13.51,55.49,39.63a8.06,8.06,0,0,0,2.71,1.25l64,16a8,8,0,0,0,7.6-2.1l55.07-55.08,26.42-13.21a16,16,0,0,0,7.15-21.46Zm-54.89,33.37L165,113.72a8,8,0,0,0-10.68.61C136.51,132.27,116.66,130,104,122L147.24,80h31.81l27.21,54.41ZM41.53,64,62,74.22,36.43,125.27,16,115.06Zm116,119.13L99.42,168.61l-49.2-35.14,28-56L128,64.28l9.8,2.59-45,43.68-.08.09a16,16,0,0,0,2.72,24.81c20.56,13.13,45.37,11,64.91-5L188,152.66Zm62-57.87-25.52-51L214.47,64,240,115.06Zm-87.75,92.67a8,8,0,0,1-7.75,6.06,8.13,8.13,0,0,1-1.95-.24L80.41,213.33a7.89,7.89,0,0,1-2.71-1.25L51.35,193.26a8,8,0,0,1,9.3-13l25.11,17.94L126,208.24A8,8,0,0,1,131.82,217.94Z"/>';
PH_ICONS.download = '<path d="M224,144v64a8,8,0,0,1-8,8H40a8,8,0,0,1-8-8V144a8,8,0,0,1,16,0v56H208V144a8,8,0,0,1,16,0Zm-101.66,5.66a8,8,0,0,0,11.32,0l40-40a8,8,0,0,0-11.32-11.32L136,124.69V32a8,8,0,0,0-16,0v92.69L93.66,98.34a8,8,0,0,0-11.32,11.32Z"/>';

if (TAG_KEY_TYPES.indexOf('sponsor') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'sponsor');

var SP_LINKS = ['website', 'instagram', 'telegram', 'x', 'youtube', 'aparat', 'whatsapp', 'app', 'other'];
var SP_STATES = ['live', 'scheduled', 'ended', 'off'];

function spLogo(it, size) {
  var box = h('div', { class: 'eco-sp-logo shrink-0 rounded-md overflow-hidden flex items-center justify-center', style: it.color ? '--sp-color:' + it.color : null });
  if (it.logoUrl) box.appendChild(img(it.logoUrl, 'sponsors'));
  else box.appendChild(icon('sponsorDuo', size || 'h-5 w-5'));
  return box;
}

function spDownload(id, days) {
  var tk = token();
  return fetch(PUB_BASE + '/sponsors/' + encodeURIComponent(id) + '/report.csv' + qs({ days: days }), { headers: { Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) return res.json().then(function (b) { throw new ApiError(res.status, b && b.error, b && b.message); }, function () { throw new ApiError(res.status, null, ''); });
    var name = 'sponsor.csv';
    var m = /filename="([^"]+)"/.exec(res.headers.get('Content-Disposition') || '');
    if (m) name = m[1];
    return res.blob().then(function (blob) {
      var url = URL.createObjectURL(blob);
      var a = h('a', { href: url, download: name, hidden: true });
      document.body.appendChild(a);
      a.click();
      setTimeout(function () { URL.revokeObjectURL(url); a.remove(); }, 1000);
      return name;
    });
  });
}

function sponsorReportField() {
  return { name: 'report', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
    if (!v.id) return h('span', {});
    var days = 14;
    var box = h('div', { class: 'eco-sp-report', 'data-sp-report': '' });
    var body = h('div', {});
    var csvBtn = button({ variant: 'ghost', size: 'sm', icon: 'download', label: t('sp.csv'), onclick: function () {
      setBusy(csvBtn, true);
      spDownload(v.id, days).then(function (name) { toast('success', t('sp.csv.done', { name: name })); }, function (err) { toast('error', errorText(err)); }).then(function () { setBusy(csvBtn, false); });
    } });
    csvBtn.setAttribute('data-sp-csv', '');
    var seg = segmented([7, 14, 30, 90].map(function (n) { return { value: String(n), label: t('sp.daysN', { n: fmtNum(n) }) }; }), String(days), function (x) { days = Number(x); load(); }, t('sp.range'));
    seg.setAttribute('data-sp-range', '');
    box.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, seg, csvBtn));
    box.appendChild(body);
    function tile(label, value, attr) {
      var el = h('div', { class: 'eco-sp-tile rounded-md px-3 py-2' }, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: label }),
        h('div', { class: 'text-xl font-semibold tabular-nums', text: value }));
      el.setAttribute(attr, '');
      return el;
    }
    function load() {
      clear(body).appendChild(h('div', { class: 'flex flex-col gap-2' }, bar('100%', 'h-14'), bar('100%', 'h-24'), bar('60%')));
      pubApi('GET', '/sponsors/' + encodeURIComponent(v.id) + '/report' + qs({ days: days })).then(function (res) {
        clear(body);
        body.appendChild(h('div', { class: 'eco-sp-tiles' },
          tile(t('sp.r.views'), fmtNum(res.views || 0), 'data-sp-views'),
          tile(t('sp.r.clicks'), fmtNum(res.clicks || 0), 'data-sp-clicks'),
          tile(t('sp.r.ctr'), res.views ? t('sp.r.ctrN', { n: fmtNum(Math.round(((res.clicks || 0) / res.views) * 1000) / 10) }) : '—', 'data-sp-ctr'),
          tile(t('sp.r.players'), fmtNum(res.players || 0), 'data-sp-players')));
        var max = 1;
        (res.days || []).forEach(function (d) { if (d.views > max) max = d.views; if (d.clicks > max) max = d.clicks; });
        var bars = h('div', { class: 'eco-bn-bars mt-3', 'aria-hidden': 'true', 'data-sp-bars': String((res.days || []).length) });
        (res.days || []).forEach(function (d) {
          bars.appendChild(h('div', { class: 'eco-bn-day', title: d.day },
            h('span', { class: 'eco-bn-v', style: 'block-size:' + Math.round((d.views / max) * 100) + '%' }),
            h('span', { class: 'eco-bn-c', style: 'block-size:' + Math.round((d.clicks / max) * 100) + '%' })));
        });
        body.appendChild(bars);
        body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-3 text-xs text-secondary-600 dark:text-secondary-300 mt-1' },
          h('span', { class: 'eco-bn-key eco-bn-key-v', text: t('bn.stats.viewsKey') }), h('span', { class: 'eco-bn-key eco-bn-key-c', text: t('bn.stats.clicksKey') }),
          h('span', { text: t('sp.r.note', { from: fmtDay(res.from), to: fmtDay(res.to), zone: res.timezone || bnZone() }) })));
        var tbl = h('div', { class: 'flex flex-col mt-3 eco-tk-list', 'data-sp-banners': '' });
        (res.banners || []).forEach(function (b) {
          tbl.appendChild(h('a', { href: hashFor('banners', b.id), class: 'flex flex-wrap items-center gap-2 py-2 text-sm', 'data-sp-banner': b.key },
            h('span', { class: 'font-medium truncate flex-1 min-w-0', text: i18nText(b.title) || b.key }),
            h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('bn.viewsClicks', { v: fmtNum(b.views || 0), c: fmtNum(b.clicks || 0) }) })));
        });
        (res.placements || []).forEach(function (p) {
          tbl.appendChild(h('a', { href: hashFor('banner-lists', p.id), class: 'flex items-center gap-2 py-2 text-sm', 'data-sp-placement': p.key },
            icon('banner-lists', 'h-4 w-4'), h('span', { class: 'truncate flex-1', text: t('sp.r.hosts', { name: i18nText(p.name) || p.key }) })));
        });
        if (!(res.banners || []).length && !(res.placements || []).length) tbl.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('sp.r.none') }));
        body.appendChild(h('h4', { class: 'text-sm font-medium mt-4', text: t('sp.r.where') }));
        body.appendChild(tbl);
      }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
    }
    load();
    return box;
  } };
}

function fmtDay(d) {
  if (!d) return '';
  try {
    return new Intl.DateTimeFormat(locale(), { timeZone: 'UTC', month: 'short', day: 'numeric' }).format(new Date(d + 'T00:00:00Z'));
  } catch (e) {
    return d;
  }
}

RESOURCES.sponsors = {
  api: 'publication',
  prepare: bannerSettings,
  filters: [
    { name: 'state', label: 'filter.bn.state', options: opts('bn.state.', SP_STATES) },
    { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: [{ keyType: 'sponsor' }, { keyType: 'general' }] }
  ],
  columns: [
    { label: 'col.sp.sponsor', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, spLogo(it),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '50%' },
    { label: 'col.bn.when', render: function (it) {
      var box = h('div', { class: 'flex flex-col items-start gap-1 text-xs' }, bnStateTag(it.state));
      if (it.startsAt) box.appendChild(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('bn.from', { date: zoneFmt(it.startsAt) }) }));
      if (it.endsAt) box.appendChild(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('bn.to', { date: zoneFmt(it.endsAt) }) }));
      return box;
    }, width: 190 },
    { label: 'col.sp.where', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm' },
        h('a', { href: hashFor('banners', null, { sponsorId: it.id }), class: 'eco-link tabular-nums whitespace-nowrap', 'data-sp-banner-count': '', text: t('sp.bannersN', { n: fmtNum(it.banners || 0) }) }),
        it.placements ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: t('sp.placementsN', { n: fmtNum(it.placements) }) }) : null);
    }, width: 150 },
    { label: 'col.bn.stats', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums whitespace-nowrap', 'data-sp-row-stats': '' },
        h('span', { text: t('bn.viewsClicks', { v: fmtNum(it.views7 || 0), c: fmtNum(it.clicks7 || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: ctrText(it.views7 || 0, it.clicks7 || 0) }));
    }, width: 150 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.sp.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.sp.active.hint' }
    ] },
    tr('name', { label: 'f.sp.name', required: true, max: 100 }),
    tr('description', { multiline: true, max: 1000, optional: true, hint: 'f.sp.description.hint' }),
    { type: 'row', fields: [
      { name: 'logoUrl', type: 'image', label: 'f.sp.logo', hint: 'f.sp.logo.hint', optional: true },
      { name: 'color', type: 'text', ltr: true, label: 'f.sp.color', hint: 'f.bn.color.hint', optional: true, check: function (v) { return v && !/^#[0-9a-fA-F]{6}$/.test(v) ? t('code.INVALID_COLOR') : null; } }
    ] },
    { name: 'websiteUrl', type: 'text', ltr: true, label: 'f.sp.website', hint: 'f.sp.website.hint', optional: true, check: function (v) { return v && !/^https?:\/\/[^\s"'<>`]+$/i.test(v) ? t('code.INVALID_URL') : null; } },
    { name: 'links', type: 'rows', label: 'f.sp.links', hint: 'f.sp.links.hint', addLabel: 'sp.link.add', emptyText: 'sp.link.empty', max: 8, empty: [], optional: true,
      columns: [
        { name: 'type', type: 'select', label: 'f.sp.link.type', options: opts('sp.link.', SP_LINKS) },
        { name: 'url', type: 'text', ltr: true, label: 'f.sp.link.url', wide: true, required: true },
        { name: 'label', type: 'i18n', label: 'f.sp.link.label', optional: true, max: 40, wide: true }
      ],
      check: function (rows) {
        if (!Array.isArray(rows)) return null;
        for (var i = 0; i < rows.length; i++) {
          if (!/^https?:\/\/[^\s"'<>`]+$/i.test((rows[i] || {}).url || '')) return { index: i, message: t('code.INVALID_URL') };
        }
        return null;
      } },
    tagsField('sponsor'),
    { type: 'heading', label: 'sp.h.contact' },
    { name: 'contactNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('sp.contact.note') }); } },
    { type: 'row', fields: [
      { name: 'contactName', type: 'text', label: 'f.sp.contactName', optional: true },
      { name: 'contactEmail', type: 'text', ltr: true, label: 'f.sp.contactEmail', optional: true, check: function (v) { return v && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.trim()) ? t('code.INVALID_EMAIL') : null; } },
      { name: 'contactPhone', type: 'text', ltr: true, label: 'f.sp.contactPhone', optional: true }
    ] },
    { name: 'notes', type: 'textarea', label: 'f.sp.notes', optional: true },
    { type: 'heading', label: 'sp.h.when' },
    { name: 'zoneNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('sp.zoneNote', { zone: bnZone() }) }); } },
    { type: 'row', fields: [
      { name: 'startsAt', type: 'custom', label: 'f.sp.startsAt', optional: true, build: function (cur, change) { return zoneDateInput(cur, change); } },
      { name: 'endsAt', type: 'custom', label: 'f.sp.endsAt', optional: true, build: function (cur, change) { return zoneDateInput(cur, change); },
        check: function (v, vals) { return v && vals.startsAt && new Date(v) <= new Date(vals.startsAt) ? t('code.INVALID_WINDOW') : null; } }
    ] },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' },
    { type: 'heading', label: 'sp.h.report', when: function (v) { return !!v.id; } },
    sponsorReportField()
  ],
  defaults: function () {
    return { key: '', isActive: true, name: {}, description: {}, logoUrl: '', color: '', websiteUrl: '', links: [], tagIds: [], contactName: '', contactEmail: '', contactPhone: '', notes: '', startsAt: null, endsAt: null, sort: 0 };
  },
  toForm: function (v) {
    v.links = (v.links || []).map(function (l) { return { type: l.type, url: l.url, label: l.label || {} }; });
    v.color = v.color || '';
    return v;
  },
  fixPayload: function (p) {
    p.links = (p.links || []).map(function (l) { return { type: l.type || 'website', url: (l.url || '').trim(), label: l.label || {} }; });
    p.color = (p.color || '').toLowerCase();
    p.logoUrl = p.logoUrl || '';
    p.websiteUrl = (p.websiteUrl || '').trim();
    p.contactEmail = (p.contactEmail || '').trim();
    p.contactPhone = toLatinDigits(p.contactPhone || '').trim();
    p.startsAt = p.startsAt || '';
    p.endsAt = p.endsAt || '';
    if (p.sort == null) p.sort = 0;
    delete p.report;
    delete p.zoneNote;
    delete p.contactNote;
    return p;
  },
  afterSave: function () { forgetRefs('sponsors'); }
};
