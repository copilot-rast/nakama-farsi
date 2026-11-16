var NEWS_STATUSES = ['draft', 'pending_review', 'published', 'archived', 'rejected'];
var NEWS_STATUS_COLOR = { draft: 'secondary', pending_review: 'warning', published: 'success', archived: 'secondary', rejected: 'error' };
var NEWS_NOTIFY = ['none', 'in_app', 'push', 'both'];
var NEWS_MAP_FIELDS = ['title', 'summary', 'body', 'cover', 'category', 'tags', 'lang', 'publishedAt', 'expiresAt', 'externalId', 'url', 'slug', 'metaTitle', 'metaDescription', 'keywords', 'canonicalUrl', 'ogImage', 'noindex'];
var NEWS_MAP_DEFAULT = { title: 'title', summary: 'summary', body: 'body', cover: 'image', category: 'category', tags: 'tags', lang: 'lang', publishedAt: 'publishedAt', expiresAt: 'expiresAt', externalId: 'id', url: 'url',
  slug: 'slug', metaTitle: 'seo.title', metaDescription: 'seo.description', keywords: 'keywords', canonicalUrl: 'canonicalUrl', ogImage: 'seo.image', noindex: 'noindex' };
var INGEST_STATUSES = ['created', 'updated', 'unchanged', 'ignored', 'rejected', 'error', 'preview'];
var INGEST_COLOR = { created: 'success', updated: 'primary', unchanged: 'secondary', ignored: 'secondary', rejected: 'error', error: 'error', preview: 'primary' };

PH_ICONS['news-review'] = '<path d="M173.66,98.34a8,8,0,0,1,0,11.32l-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35A8,8,0,0,1,173.66,98.34ZM232,128A104,104,0,1,1,128,24,104.11,104.11,0,0,1,232,128Zm-16,0a88,88,0,1,0-88,88A88.1,88.1,0,0,0,216,128Z"/>';
PH_ICONS['news-sources'] = '<path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm40-88a40,40,0,1,1-40-40A40,40,0,0,1,168,128Z"/>';
PH_ICONS['news-ingest'] = '<path d="M224,128a8,8,0,0,1-8,8H40a8,8,0,0,1,0-16H216A8,8,0,0,1,224,128ZM40,72H216a8,8,0,0,0,0-16H40a8,8,0,0,0,0,16ZM216,184H40a8,8,0,0,0,0,16H216a8,8,0,0,0,0-16Z"/>';
PH_ICONS['news-reactions'] = '<path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216ZM80,108a12,12,0,1,1,12,12A12,12,0,0,1,80,108Zm96,0a12,12,0,1,1-12-12A12,12,0,0,1,176,108Zm-1.07,48c-10.29,17.79-27.4,28-46.93,28s-36.63-10.2-46.92-28a8,8,0,1,1,13.84-8c7.47,12.91,19.21,20,33.08,20s25.61-7.1,33.07-20a8,8,0,0,1,13.86,8Z"/>';
PH_ICONS['news-settings'] = '<path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Z"/>';

function newsStatusTag(st) {
  return tag(t('ns.' + st), NEWS_STATUS_COLOR[st] || 'secondary');
}

function plainOf(v) {
  var s = i18nText(v) || '';
  if (/<[a-z][\s\S]*>/i.test(s)) {
    var d = document.createElement('div');
    d.innerHTML = s;
    s = d.textContent || '';
  } else {
    s = mdPlain(s);
  }
  return s.replace(/\s+/g, ' ').trim();
}

function clip(s, n) {
  s = String(s || '');
  return Array.from(s).length > n ? Array.from(s).slice(0, n - 1).join('') + '…' : s;
}

function seoSnippetField() {
  return { name: 'seoPreview', type: 'static', readonly: true, render: function (v) {
    var box = h('div', { class: 'eco-serp rounded-md px-4 py-3', 'data-news-serp': '' });
    function paint() {
      clear(box);
      var title = i18nText(v.metaTitle) || i18nText(v.title) || t('seo.noTitle');
      var desc = i18nText(v.metaDescription) || i18nText(v.summary) || plainOf(v.body);
      var url = (NEWS_SET.publicBaseUrl || 'https://example.com') + String(NEWS_SET.publicPath || '/news/{slug}').replace('{slug}', v.slug || t('seo.autoSlug'));
      box.appendChild(h('div', { class: 'eco-serp-url', dir: 'ltr', text: url }));
      box.appendChild(h('div', { class: 'eco-serp-title', dir: 'auto', 'data-serp-title': '', text: clip(title, 60) }));
      box.appendChild(h('div', { class: 'eco-serp-desc', dir: 'auto', 'data-serp-desc': '', text: clip(desc, 160) }));
      if (v.noindex) box.appendChild(h('div', { class: 'mt-1' }, tag(t('seo.noindexOn'), 'warning')));
    }
    var pending = 0;
    function later() { if (pending) return; pending = requestAnimationFrame(function () { pending = 0; paint(); }); }
    setTimeout(function () {
      var form = box.closest('form');
      if (!form) return;
      form.addEventListener('input', later);
      form.addEventListener('change', later);
      form.addEventListener('click', later);
      form.addEventListener('keyup', later);
    }, 0);
    paint();
    return h('div', {}, h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mb-1', text: t('seo.preview') }), box);
  } };
}

function keywordsInput(cur, change) {
  var inp = textInput((cur || []).join('، '), { label: t('f.keywords'), placeholder: t('f.keywords.ph'), oninput: function (val) {
    change(val.split(/[,،\n]/).map(function (x) { return x.trim(); }).filter(Boolean));
  } });
  inp.setAttribute('data-news-keywords', '');
  return inp;
}

var NEWS_SET = {};
function newsSettings() {
  return pubApi('GET', '/news-settings').then(function (s) { NEWS_SET = s || {}; return s; }, function () { return null; });
}

RESOURCES.news = {
  api: 'publication',
  prepare: newsSettings,
  filters: [
    { name: 'status', label: 'filter.status', options: opts('ns.', NEWS_STATUSES) },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'news' } },
    { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: [{ keyType: 'news' }, { keyType: 'general' }] },
    { name: 'expired', label: 'filter.expired', options: function () { return [{ value: 'false', label: t('news.notExpired') }, { value: 'true', label: t('news.expired') }]; } },
    { name: 'isPinned', label: 'filter.pinned', options: function () { return [{ value: 'true', label: t('news.pinned') }, { value: 'false', label: t('news.notPinned') }]; } }
  ],
  columns: [
    { label: 'col.title', render: function (it) {
      var pic = h('div', { class: 'h-10 w-16 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
      pic.appendChild(it.coverUrl ? img(it.coverUrl) : icon('image', 'h-4 w-4'));
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic,
        h('div', { class: 'min-w-0' },
          h('div', { class: 'font-medium flex items-center gap-1.5 min-w-0' }, it.isPinned ? h('span', { class: 'eco-text-warning', title: t('news.pinned') }, icon('pin', 'h-3.5 w-3.5')) : null, h('span', { class: 'eco-clamp', text: i18nText(it.title) })),
          h('div', { class: 'text-xs text-secondary-500 eco-clamp flex flex-wrap items-center eco-meta' },
            it.slug ? h('span', { dir: 'auto', 'data-news-slug': '', text: '/' + it.slug }) : null,
            it.source && it.source !== 'console' ? h('span', { 'data-news-source': it.source, text: it.source.replace('webhook:', t('news.from') + ' ') }) : null)));
    }, skeleton: '70%' },
    { label: 'col.category', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, it.categoryName ? tag(i18nText(it.categoryName), 'secondary') : h('span', { class: 'text-secondary-400', text: t('none') }), taxonomyCell(it));
    }, width: 160 },
    { label: 'col.comments', render: function (it) {
      var box = h('div', { class: 'flex flex-wrap items-center gap-1' });
      box.appendChild(it.allowComment ? h('span', { class: 'tabular-nums text-sm', title: t('col.comments'), text: '💬 ' + fmtNum(it.commentCount || 0) }) : tag(t('news.commentsOff'), 'secondary'));
      if (it.pendingComments) box.appendChild(h('a', { href: hashFor('comments', null, { status: 'pending', newsId: it.id }), class: 'eco-tag eco-tag--warning inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', text: t('cm.pendingN', { n: it.pendingComments }) }));
      if (it.reactionCount) box.appendChild(h('span', { class: 'tabular-nums text-xs text-secondary-600 dark:text-secondary-300', 'data-news-reactions': '', text: t('news.reactionsN', { n: fmtNum(it.reactionCount) }) }));
      return box;
    }, width: 150 },
    { label: 'col.published', render: function (it) {
      var future = it.publishedAt && new Date(it.publishedAt).getTime() > Date.now();
      return h('div', { class: 'flex flex-col items-start gap-0.5' }, dateCell(it.publishedAt),
        future && it.status === 'published' ? tag(t('news.scheduled'), 'primary') : null,
        it.expired ? tag(t('news.expired'), 'secondary') : it.expiresAt ? h('span', { class: 'text-xs text-secondary-500', text: t('news.until', { t: fmtDate(it.expiresAt, true) }) }) : null);
    }, width: 190 },
    { label: 'col.status', render: function (it) { return h('span', { 'data-news-status': it.status }, newsStatusTag(it.status)); }, width: 130 }
  ],
  fields: [
    tr('title', { required: true, max: 200 }),
    tr('summary', { multiline: true, max: 1000, optional: true }),
    tr('body', { rich: true, required: true, max: 200000 }),
    { name: 'coverUrl', type: 'image', label: 'f.cover', optional: true },
    { type: 'row', fields: [categoryField('news'), tagsField('news')] },
    { type: 'row', fields: [
      { name: 'status', type: 'select', label: 'f.news.status', hint: 'f.news.status.hint', required: true, noPlaceholder: true, options: opts('ns.', NEWS_STATUSES) },
      { name: 'isPinned', type: 'bool', inline: true, label: 'f.isPinned', hint: 'f.isPinned.hint' }
    ] },
    { name: 'rejectReason', type: 'text', label: 'f.news.rejectReason', optional: true, when: function (v) { return v.status === 'rejected'; } },
    { type: 'row', fields: [
      { name: 'publishedAt', type: 'datetime', label: 'f.publishedAt', hint: 'f.publishedAt.hint', optional: true },
      { name: 'expiresAt', type: 'datetime', label: 'f.news.expiresAt', hint: 'f.news.expiresAt.hint', optional: true,
        check: function (v, vals) { return v && vals.publishedAt && new Date(v) <= new Date(vals.publishedAt) ? t('code.EXPIRES_BEFORE_PUBLISH') : null; } }
    ] },
    { type: 'heading', label: 'f.news.engage' },
    { type: 'row', fields: [
      { name: 'allowComment', type: 'bool', inline: true, label: 'f.allowComment', hint: 'f.allowComment.hint' },
      { name: 'freeComment', type: 'bool', inline: true, label: 'f.freeComment', hint: 'f.freeComment.hint', empty: false, when: function (v) { return !!v.allowComment; } }
    ] },
    { name: 'allowReaction', type: 'bool', inline: true, label: 'f.news.allowReaction', hint: 'f.news.allowReaction.hint' },
    { type: 'heading', label: 'f.news.notify' },
    { name: 'notifyMode', type: 'select', label: 'f.news.notifyMode', hint: 'f.news.notifyMode.hint', required: true, noPlaceholder: true, options: opts('nn.', NEWS_NOTIFY) },
    { type: 'row', fields: [
      { name: 'notifyAudience.audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.news.audience', hint: 'f.news.audience.hint', optional: true, empty: [], when: function (v) { return v.notifyMode && v.notifyMode !== 'none'; } },
      { name: 'notifyAudience.excludeTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.news.exclude', optional: true, empty: [], when: function (v) { return v.notifyMode && v.notifyMode !== 'none'; } }
    ] },
    { name: 'notifyState', type: 'static', readonly: true, render: function (v) {
      if (v.notifiedAt) return notice('success', t('news.notifiedAt', { t: fmtDate(v.notifiedAt, true) }), v.notificationId ? h('a', { href: hashFor('notifications', v.notificationId), class: 'eco-link text-sm', text: t('ar.openNote') }) : null);
      return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('news.notifyOnce') });
    } },
    { type: 'heading', label: 'f.news.seo' },
    seoSnippetField(),
    { name: 'slug', type: 'text', label: 'f.news.slug', hint: 'f.news.slug.hint', optional: true,
      check: function (v) { return v && !/^[\p{L}\p{N}]+(-[\p{L}\p{N}]+)*$/u.test(v.trim()) ? t('code.INVALID_SLUG') : null; } },
    tr('metaTitle', { label: 'f.news.metaTitle', hint: 'f.news.metaTitle.hint', max: 200, counter: true, optional: true }),
    tr('metaDescription', { label: 'f.news.metaDescription', hint: 'f.news.metaDescription.hint', multiline: true, rows: 2, max: 400, counter: true, optional: true }),
    { name: 'keywords', type: 'custom', label: 'f.keywords', hint: 'f.keywords.hint', optional: true, build: function (cur, change) { return keywordsInput(cur, change); } },
    { type: 'row', fields: [
      { name: 'canonicalUrl', type: 'text', ltr: true, label: 'f.news.canonical', hint: 'f.news.canonical.hint', optional: true,
        check: function (v) { return v && !/^https?:\/\/[^\s"'<>`]+$/i.test(v.trim()) ? t('code.INVALID_URL') : null; } },
      { name: 'noindex', type: 'bool', inline: true, label: 'f.news.noindex', hint: 'f.news.noindex.hint' }
    ] },
    { name: 'ogImage', type: 'image', label: 'f.news.ogImage', hint: 'f.news.ogImage.hint', optional: true },
    { type: 'heading', label: 'f.news.origin', when: function (v) { return !!v.id; } },
    { name: 'origin', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
      return dl([
        ['news.o.source', h('span', { dir: 'ltr', text: v.source || 'console' })],
        ['news.o.external', v.externalId ? mono(v.externalId) : null],
        ['news.o.url', v.sourceUrl ? h('a', { href: v.sourceUrl, target: '_blank', rel: 'noopener noreferrer', class: 'eco-link', dir: 'ltr', text: v.sourceUrl }) : null],
        ['news.o.created', v.createdBy ? h('span', { text: v.createdBy + ' · ' + fmtDate(v.createdAt, true) }) : h('span', { text: fmtDate(v.createdAt, true) })],
        ['news.o.approved', v.approvedAt ? h('span', { text: (v.approvedBy || '') + ' · ' + fmtDate(v.approvedAt, true) }) : null],
        ['news.o.updated', v.updatedBy ? h('span', { text: v.updatedBy }) : null]
      ]);
    } },
    { type: 'heading', label: 'f.actions' },
    { name: 'actions', type: 'actions', empty: [] },
    { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
  ],
  defaults: function (query) {
    return { title: {}, summary: {}, body: {}, coverUrl: '', categoryId: (query && query.categoryId) || null, tagIds: [], status: 'published', publishedAt: null, expiresAt: null, isPinned: false,
      allowComment: true, freeComment: false, allowReaction: true, notifyMode: 'none', notifyAudience: { audienceTagIds: [], excludeTagIds: [] }, slug: '', metaTitle: {}, metaDescription: {}, keywords: [],
      canonicalUrl: '', ogImage: '', noindex: false, actions: [], metadata: {} };
  },
  toForm: function (v) {
    if (v.bodyFormat !== 'html') {
      var body = {};
      Object.keys(v.body || {}).forEach(function (k) { body[k] = mdHtml(v.body[k]); });
      v.body = body;
    }
    var a = v.notifyAudience || {};
    v.notifyAudience = { audienceTagIds: a.audienceTagIds || [], excludeTagIds: a.excludeTagIds || [] };
    v.keywords = v.keywords || [];
    v.tagIds = v.tagIds || [];
    return v;
  },
  fixPayload: function (p) {
    p.bodyFormat = 'html';
    if (!p.coverUrl) p.coverUrl = '';
    if (!p.publishedAt) delete p.publishedAt;
    if (!p.expiresAt) delete p.expiresAt;
    p.slug = (p.slug || '').trim();
    p.canonicalUrl = (p.canonicalUrl || '').trim();
    p.ogImage = p.ogImage || '';
    var a = p.notifyAudience || {};
    p.notifyAudience = p.notifyMode && p.notifyMode !== 'none' ? { audienceTagIds: a.audienceTagIds || [], excludeTagIds: a.excludeTagIds || [] } : null;
    ['seoPreview', 'notifyState', 'origin'].forEach(function (k) { delete p[k]; });
    return p;
  },
  afterSave: function () { forgetRefs('news'); }
};

function newsReviewPage(container, query) {
  var frame = pageFrame(t('nav.news-review'), t('desc.news-review'));
  clear(container).appendChild(frame.el);
  var state = { page: Math.max(1, parseInt(query.page, 10) || 1), total: 0, items: [], sel: {} };
  var box = h('div', {});
  var count = h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', 'data-review-count': '' });
  var reason = textInput('', { label: t('rv.reason'), placeholder: t('rv.reason.ph') });
  reason.setAttribute('data-review-reason', '');
  var approve = button({ variant: 'primary', size: 'sm', icon: 'check', label: t('rv.approve'), onclick: function () { act('approve', approve); } });
  var reject = button({ variant: 'danger', size: 'sm', icon: 'close', label: t('rv.reject'), onclick: function () { act('reject', reject); } });
  approve.setAttribute('data-review', 'approve');
  reject.setAttribute('data-review', 'reject');
  var all = checkbox(false, t('rv.all'), function (on) { state.items.forEach(function (it) { if (on) state.sel[it.id] = true; else delete state.sel[it.id]; }); paint(); });
  all.setAttribute('data-review-all', '');
  frame.content.appendChild(card(h('div', { class: 'p-3 flex flex-col gap-2 eco-bulk', 'data-review-bar': '' },
    h('div', { class: 'flex flex-wrap items-center gap-3' }, all, count),
    h('div', { class: 'flex flex-wrap items-end gap-2' }, h('div', { class: 'min-w-0 flex-1' }, reason), approve, reject))));
  frame.content.appendChild(h('div', { class: 'mt-3' }, box));
  function selected() { return Object.keys(state.sel); }
  function paintCount() {
    var n = selected().length;
    count.textContent = n ? t('rv.selectedN', { n: fmtNum(n) }) : t('rv.pendingN', { n: fmtNum(state.total) });
    approve.disabled = reject.disabled = !n;
  }
  function act(action, btn) {
    var ids = selected();
    if (!ids.length) return;
    setBusy(btn, true);
    pubApi('POST', '/news-review', { ids: ids, action: action, reason: reason.value.trim() }).then(function (res) {
      toast('success', t(action === 'approve' ? 'rv.approved' : 'rv.rejected', { n: fmtNum(res.changed || 0) }));
      state.sel = {};
      reason.value = '';
      load();
    }, function (err) { toast('error', errorText(err)); }).then(function () { setBusy(btn, false); paintCount(); });
  }
  function paint() {
    clear(box);
    if (!state.items.length) {
      box.appendChild(card(emptyBlock(t('rv.empty'), t('rv.empty.hint'))));
      paintCount();
      return;
    }
    var cols = [{ label: '', width: 40 }, { label: t('col.title') }, { label: t('col.source'), width: 160 }, { label: t('col.published'), width: 170 }, { label: t('col.actions'), width: 200 }];
    box.appendChild(table(cols, state.items.map(function (it) {
      var cb = h('input', { type: 'checkbox', class: 'eco-check', 'aria-label': i18nText(it.title), 'data-review-pick': it.id });
      cb.checked = !!state.sel[it.id];
      cb.addEventListener('change', function () { if (cb.checked) state.sel[it.id] = true; else delete state.sel[it.id]; paintCount(); });
      var one = function (action) {
        return function (e) {
          var b = e.currentTarget;
          setBusy(b, true);
          pubApi('POST', '/news-review', { ids: [it.id], action: action, reason: reason.value.trim() }).then(function () {
            toast('success', t(action === 'approve' ? 'rv.approved' : 'rv.rejected', { n: fmtNum(1) }));
            delete state.sel[it.id];
            load();
          }, function (err) { setBusy(b, false); toast('error', errorText(err)); });
        };
      };
      return { item: it, cells: [
        cb,
        h('div', { class: 'min-w-0' },
          h('a', { href: hashFor('news', it.id), class: 'eco-link font-medium eco-clamp', text: i18nText(it.title) }),
          i18nText(it.summary) ? h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: i18nText(it.summary) }) : null),
        h('span', { class: 'text-sm', dir: 'ltr', text: it.source || 'console' }),
        dateCell(it.publishedAt),
        h('div', { class: 'flex flex-wrap gap-1 justify-end' },
          button({ variant: 'primary', size: 'sm', label: t('rv.approve'), onclick: one('approve') }),
          button({ variant: 'ghost', size: 'sm', label: t('rv.reject'), onclick: one('reject') }))
      ] };
    })));
    var pg = h('div', { class: 'flex justify-end py-3' });
    pg.appendChild(pager({ page: state.page, limit: 50, total: state.total, loading: false }, function (p) { state.page = p; load(); }));
    box.appendChild(pg);
    paintCount();
  }
  function load() {
    clear(box).appendChild(table([{ label: '' }, { label: t('col.title') }], skeletonRows([{}, {}], 4)));
    pubApi('GET', '/news' + qs({ status: 'pending_review', page: state.page, limit: 50 })).then(function (res) {
      state.items = res.items || [];
      state.total = res.total || 0;
      paint();
    }, function (err) { clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

function sourceEndpoint(v) {
  var base = location.protocol + '//' + location.hostname + (NEWS_SET.apiPort ? ':' + NEWS_SET.apiPort : '');
  return { path: '/v2/publication/news/ingest/' + (v.key || '…'), base: base };
}

function secretField() {
  return { name: 'secretBox', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
    var out = h('div', { class: 'flex flex-col gap-2', 'data-src-secret': '' });
    var shown = h('div', { 'aria-live': 'polite' });
    var hint = h('p', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: v.secretSet ? t('src.secret.set', { hint: v.secretHint || '' }) : t('src.secret.none') });
    function run(kind, btn) {
      var go2 = function () {
        setBusy(btn, true);
        pubApi('POST', '/news-sources/' + encodeURIComponent(v.id) + '/secret/' + kind).then(function (res) {
          clear(shown).appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' },
            h('code', { class: 'eco-secret-out font-mono text-sm px-2 py-1 rounded-md', dir: 'ltr', 'data-src-secret-value': '', text: res.secret }),
            button({ variant: 'ghost', size: 'sm', icon: 'copy', label: t('up.copy'), onclick: function () { copyText(res.secret).then(function () { toast('success', t('up.copied')); }); } })));
          hint.textContent = t('src.secret.set', { hint: res.secretHint || '' });
          if (kind === 'rotate') toast('success', t('src.rotated'));
        }, function (err) { toast('error', errorText(err)); }).then(function () { setBusy(btn, false); });
      };
      if (kind !== 'rotate') { go2(); return; }
      confirmModal({ title: t('src.rotate.title'), text: t('src.rotate.text'), ok: t('src.rotate'), danger: true }).then(function (yes) { if (yes) go2(); });
    }
    var reveal = button({ variant: 'secondary', size: 'sm', icon: 'eye', label: t('src.reveal'), onclick: function (e) { run('reveal', e.currentTarget); } });
    var rotate = button({ variant: 'danger', size: 'sm', label: t('src.rotate'), onclick: function (e) { run('rotate', e.currentTarget); } });
    reveal.setAttribute('data-src-reveal', '');
    rotate.setAttribute('data-src-rotate', '');
    out.appendChild(hint);
    out.appendChild(h('div', { class: 'flex flex-wrap gap-2' }, reveal, rotate));
    out.appendChild(shown);
    return out;
  } };
}

function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text);
  var ta = h('textarea', { style: 'position:fixed;opacity:0' });
  ta.value = text;
  document.body.appendChild(ta);
  ta.select();
  try { document.execCommand('copy'); } catch (e) { ta.remove(); return Promise.resolve(); }
  ta.remove();
  return Promise.resolve();
}

function liveStatic(build) {
  return function (v) {
    var box = h('div', {});
    function paint() { clear(box).appendChild(build(v)); }
    var pending = 0;
    function later() { if (pending) return; pending = requestAnimationFrame(function () { pending = 0; paint(); }); }
    setTimeout(function () {
      var form = box.closest('form');
      if (!form) return;
      form.addEventListener('input', later);
      form.addEventListener('change', later);
    }, 0);
    paint();
    return box;
  };
}

function endpointField() {
  return { name: 'endpointBox', type: 'static', readonly: true, render: liveStatic(function (v) {
    var ep = sourceEndpoint(v);
    var hmac = v.authMode === 'hmac';
    var sample = hmac
      ? 'TS=$(date +%s)\nSIG=$(printf "%s.%s" "$TS" "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -hex | cut -d" " -f2)\ncurl -X POST <API>' + ep.path + ' -H "Content-Type: application/json" -H "' + (v.timestampHeader || 'X-News-Timestamp') + ': $TS" -H "' + (v.signatureHeader || 'X-News-Signature') + ': sha256=$SIG" -d "$BODY"'
      : 'curl -X POST <API>' + ep.path + ' -H "Content-Type: application/json" -H "' + (v.tokenHeader || 'X-News-Token') + ': $SECRET" -d \'{"id":"1","title":"…","body":"<p>…</p>"}\'';
    return h('div', { class: 'flex flex-col gap-2', 'data-src-endpoint': '' },
      h('div', { class: 'flex flex-wrap items-center gap-2' }, tag('POST', 'primary'), h('code', { class: 'font-mono text-sm', dir: 'ltr', text: ep.path })),
      h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t(hmac ? 'src.how.hmac' : 'src.how.token') }),
      h('pre', { class: 'eco-code-sample text-xs rounded-md p-2 overflow-x-auto', dir: 'ltr', text: sample }));
  }) };
}

function ingestResults(sum) {
  var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-src-results': '' });
  (sum.results || []).forEach(function (r) {
    var p = r.preview || {};
    list.appendChild(h('div', { class: 'py-2 flex flex-col gap-1', 'data-src-result': r.status },
      h('div', { class: 'flex flex-wrap items-center gap-2' },
        tag(t('ig.r.' + r.status), r.status === 'failed' ? 'error' : r.status === 'created' ? 'success' : r.status === 'updated' ? 'primary' : 'secondary'),
        r.externalId ? mono(r.externalId) : null,
        r.newsId ? h('a', { href: hashFor('news', r.newsId), class: 'eco-link text-sm', text: t('ig.openNews') }) : null,
        r.code ? h('span', { class: 'text-xs eco-text-error', text: has('code.' + r.code) ? t('code.' + r.code) : r.code }) : null),
      p.title ? h('div', { class: 'text-sm font-medium', dir: 'auto', text: p.title + (p.lang ? ' (' + p.lang + ')' : '') }) : null,
      p.bodyHtml ? (function () { var d = h('div', { class: 'eco-md text-sm eco-clamp-3', dir: 'auto' }); d.innerHTML = p.bodyHtml; return d; })() : null,
      (r.notes || []).length ? h('div', { class: 'flex flex-wrap gap-1' }, r.notes.map(function (n) { return tag(has('ig.note.' + n) ? t('ig.note.' + n) : n, 'warning'); })) : null));
  });
  return list;
}

function testField() {
  return { name: 'testBox', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
    var ta = textInput('{\n  "id": "1",\n  "title": "…",\n  "body": "<p>…</p>"\n}', { multiline: true, rows: 6, ltr: true, mono: true, label: t('src.test.payload') });
    ta.setAttribute('data-src-test-payload', '');
    var out = h('div', { 'aria-live': 'polite', 'data-src-test-out': '' });
    function run(save, btn) {
      clear(out);
      setBusy(btn, true);
      pubApi('POST', '/news-sources/' + encodeURIComponent(v.id) + '/test', { payload: ta.value, save: save }).then(function (res) {
        if (!res.ok) { out.appendChild(notice('error', (has('code.' + res.code) ? t('code.' + res.code) : res.code) + (res.message ? ': ' + res.message : ''))); return; }
        var s = res.summary || {};
        out.appendChild(notice(save ? 'success' : 'info', t(save ? 'src.test.saved' : 'src.test.preview', { c: fmtNum(s.created || 0), u: fmtNum(s.updated || 0), n: fmtNum((s.unchanged || 0) + (s.ignored || 0)), f: fmtNum(s.failed || 0) })));
        out.appendChild(ingestResults(s));
        forgetRefs('news');
      }, function (err) { out.appendChild(notice('error', errorText(err))); }).then(function () { setBusy(btn, false); });
    }
    var preview = button({ variant: 'secondary', size: 'sm', label: t('src.test.run'), onclick: function (e) { run(false, e.currentTarget); } });
    var save = button({ variant: 'ghost', size: 'sm', label: t('src.test.save'), onclick: function (e) { run(true, e.currentTarget); } });
    preview.setAttribute('data-src-test', 'preview');
    save.setAttribute('data-src-test', 'save');
    return h('div', { class: 'flex flex-col gap-2' }, h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('src.test.hint') }), ta, h('div', { class: 'flex flex-wrap gap-2' }, preview, save), out);
  } };
}

function mappingFields() {
  var rows = [];
  for (var i = 0; i < NEWS_MAP_FIELDS.length; i += 2) {
    rows.push({ type: 'row', fields: NEWS_MAP_FIELDS.slice(i, i + 2).map(function (k) {
      return { name: 'mapping.' + k, type: 'text', ltr: true, label: 'map.' + k, optional: k !== 'title', required: k === 'title',
        check: function (v) { return v && !/^\$?\.?([A-Za-z_][A-Za-z0-9_-]*)?(\[[0-9]{1,4}\])*(\.([A-Za-z_][A-Za-z0-9_-]*)?(\[[0-9]{1,4}\])*)*$/.test(v.trim()) ? t('code.INVALID_PATH') : null; } };
    }) });
  }
  return rows;
}

RESOURCES['news-sources'] = {
  api: 'publication',
  prepare: newsSettings,
  filters: [activeFilter(), { name: 'publishMode', label: 'filter.src.publish', options: opts('src.publish.', ['review', 'auto']) }],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.name || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.endpoint)));
    }, skeleton: '50%' },
    { label: 'col.src.mode', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('src.auth.' + it.authMode), 'secondary'), tag(t('src.publish.' + it.publishMode), it.publishMode === 'auto' ? 'primary' : 'warning'));
    }, width: 210 },
    { label: 'col.src.traffic', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums', 'data-src-traffic': '' },
        h('span', { text: t('src.news', { n: fmtNum(it.newsCount || 0), p: fmtNum(it.pendingCount || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('src.day', { n: fmtNum(it.received24h || 0), f: fmtNum(it.failed24h || 0) }) }),
        it.lastReceivedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(it.lastReceivedAt, true) }) : null);
    }, width: 200 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: function () { return null; },
  defaults: function () {
    return { key: '', name: '', isActive: true, authMode: 'token', tokenHeader: 'X-News-Token', signatureHeader: 'X-News-Signature', timestampHeader: 'X-News-Timestamp', skewSeconds: 300,
      itemsPath: '', mapping: Object.assign({}, NEWS_MAP_DEFAULT), bodyFormat: 'auto', defaultCategoryId: null, defaultLang: '', defaultTagIds: [], publishMode: 'review', updateMode: 'update',
      publishedAtMode: 'source', rateLimitPerMinute: 60, copyImages: false, notifyMode: 'none', allowComment: true, allowReaction: true };
  },
  toForm: function (v) { v.mapping = Object.assign({}, NEWS_MAP_DEFAULT, v.mapping || {}); return v; },
  fixPayload: function (p) {
    ['endpointBox', 'secretBox', 'testBox'].forEach(function (k) { delete p[k]; });
    var m = {};
    NEWS_MAP_FIELDS.forEach(function (k) { m[k] = ((p.mapping || {})[k] || '').trim(); });
    p.mapping = m;
    p.defaultLang = p.defaultLang || '';
    p.defaultCategoryId = p.defaultCategoryId || '';
    p.itemsPath = (p.itemsPath || '').trim();
    if (!p.secret) delete p.secret;
    return p;
  },
  afterSave: function () { forgetRefs('news-sources'); }
};
RESOURCES['news-sources'].fields = [
  { type: 'row', fields: [
    { name: 'key', type: 'key', label: 'f.key', hint: 'f.src.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_-]{1,39}$/, keyErr: 'err.src.key' },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.src.active.hint' }
  ] },
  { name: 'name', type: 'text', label: 'f.src.name', optional: true },
  endpointField(),
  { type: 'heading', label: 'src.h.auth' },
  { name: 'authMode', type: 'select', label: 'f.src.auth', hint: 'f.src.auth.hint', required: true, noPlaceholder: true, options: opts('src.auth.', ['token', 'hmac']) },
  { name: 'tokenHeader', type: 'text', ltr: true, label: 'f.src.tokenHeader', required: true, when: function (v) { return v.authMode === 'token'; } },
  { type: 'row', fields: [
    { name: 'signatureHeader', type: 'text', ltr: true, label: 'f.src.signatureHeader', required: true, when: function (v) { return v.authMode === 'hmac'; } },
    { name: 'timestampHeader', type: 'text', ltr: true, label: 'f.src.timestampHeader', required: true, when: function (v) { return v.authMode === 'hmac'; } }
  ] },
  { name: 'skewSeconds', type: 'number', label: 'f.src.skew', hint: 'f.src.skew.hint', min: 30, max: 3600, required: true, when: function (v) { return v.authMode === 'hmac'; } },
  { name: 'secret', type: 'text', ltr: true, label: 'f.src.secret', hint: 'f.src.secret.hint', optional: true, when: function (v) { return !v.id; } },
  secretField(),
  { type: 'heading', label: 'src.h.mapping' },
  { type: 'row', fields: [
    { name: 'itemsPath', type: 'text', ltr: true, label: 'f.src.itemsPath', hint: 'f.src.itemsPath.hint', optional: true },
    { name: 'bodyFormat', type: 'select', label: 'f.src.bodyFormat', hint: 'f.src.bodyFormat.hint', required: true, noPlaceholder: true, options: opts('src.fmt.', ['auto', 'html', 'markdown', 'text']) }
  ] }
].concat(mappingFields(), [
  { type: 'heading', label: 'src.h.defaults' },
  { type: 'row', fields: [
    { name: 'defaultCategoryId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'news' }, label: 'f.src.category', optional: true },
    { name: 'defaultLang', type: 'select', label: 'f.src.lang', hint: 'f.src.lang.hint', optional: true, options: function () { return LANGS.active.map(function (l) { return { value: l.code, label: langName(l.code) }; }); } }
  ] },
  { name: 'defaultTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: [{ keyType: 'news' }, { keyType: 'general' }], label: 'f.src.tags', optional: true, empty: [] },
  { type: 'heading', label: 'src.h.publish' },
  { type: 'row', fields: [
    { name: 'publishMode', type: 'select', label: 'f.src.publish', hint: 'f.src.publish.hint', required: true, noPlaceholder: true, options: opts('src.publish.', ['review', 'auto']) },
    { name: 'updateMode', type: 'select', label: 'f.src.update', hint: 'f.src.update.hint', required: true, noPlaceholder: true, options: opts('src.update.', ['update', 'ignore']) }
  ] },
  { type: 'row', fields: [
    { name: 'publishedAtMode', type: 'select', label: 'f.src.time', required: true, noPlaceholder: true, options: opts('src.time.', ['source', 'received']) },
    { name: 'notifyMode', type: 'select', label: 'f.news.notifyMode', required: true, noPlaceholder: true, options: opts('nn.', NEWS_NOTIFY) }
  ] },
  { type: 'row', fields: [
    { name: 'allowComment', type: 'bool', inline: true, label: 'f.allowComment' },
    { name: 'allowReaction', type: 'bool', inline: true, label: 'f.news.allowReaction' }
  ] },
  { type: 'row', fields: [
    { name: 'rateLimitPerMinute', type: 'number', label: 'f.src.rate', hint: 'f.src.rate.hint', min: 0, max: 10000, required: true },
    { name: 'copyImages', type: 'bool', inline: true, label: 'f.src.copyImages', hint: 'f.src.copyImages.hint' }
  ] },
  { type: 'heading', label: 'src.h.test', when: function (v) { return !!v.id; } },
  testField()
]);

function ingestDetail(it) {
  var body = h('div', { class: 'flex flex-col gap-3', 'data-ingest-detail': it.id }, bar('60%'), bar('90%'));
  var title = t('ig.detail');
  var closeBtn = button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: function () { el.remove(); } });
  var el = sidePanel(title, body, closeBtn, function () { el.remove(); });
  document.body.appendChild(el);
  pubApi('GET', '/news-ingest-log/' + encodeURIComponent(it.id)).then(function (d) {
    clear(body);
    body.appendChild(dl([
      ['col.time', h('span', { text: fmtDate(d.createdAt, true) })],
      ['col.source', mono(d.sourceKey)],
      ['col.status', tag(t('ig.' + d.status), INGEST_COLOR[d.status])],
      ['ig.http', h('span', { class: 'tabular-nums', text: String(d.httpStatus) })],
      ['ig.error', d.errorCode ? h('span', { text: (has('code.' + d.errorCode) ? t('code.' + d.errorCode) : d.errorCode) + (d.message ? ' (' + d.message + ')' : '') }) : null],
      ['ig.from', d.remoteAddr ? mono(d.remoteAddr) : null],
      ['ig.by', d.createdBy ? h('span', { text: d.createdBy }) : null]
    ]));
    if ((d.results || []).length) body.appendChild(ingestResults({ results: d.results }));
    body.appendChild(h('h4', { class: 'text-sm font-medium', text: t('ig.payload') }));
    body.appendChild(h('pre', { class: 'eco-code-sample text-xs rounded-md p-2 overflow-x-auto whitespace-pre-wrap break-all', dir: 'ltr', 'data-ingest-payload': '', text: d.payload || t('ig.noPayload') }));
  }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
}

LISTS['news-ingest'] = {
  path: 'news-ingest-log',
  onRow: function (it) { ingestDetail(it); },
  filters: [
    { name: 'sourceId', label: 'filter.src', type: 'ref', resource: 'news-sources', api: 'publication' },
    { name: 'status', label: 'filter.status', options: opts('ig.', INGEST_STATUSES) },
    { name: 'isTest', label: 'filter.ig.test', options: function () { return [{ value: 'false', label: t('ig.live') }, { value: 'true', label: t('ig.test') }]; } }
  ],
  columns: [
    { label: 'col.time', render: function (it) { return dateCell(it.createdAt); }, width: 170 },
    { label: 'col.source', render: function (it) { return h('div', { class: 'flex flex-wrap items-center gap-1' }, mono(it.sourceKey), it.isTest ? tag(t('ig.test'), 'primary') : null); }, width: 180 },
    { label: 'col.status', render: function (it) { return h('span', { 'data-ingest-status': it.status }, tag(t('ig.' + it.status), INGEST_COLOR[it.status])); }, width: 120 },
    { label: 'col.ig.items', render: function (it) {
      return h('span', { class: 'text-sm tabular-nums whitespace-nowrap', text: t('ig.counts', { c: fmtNum(it.createdCount || 0), u: fmtNum(it.updatedCount || 0), n: fmtNum(it.unchangedCount || 0), f: fmtNum(it.failedCount || 0) }) });
    }, width: 220 },
    { label: 'col.ig.error', render: function (it) { return it.errorCode ? h('span', { class: 'text-sm eco-text-error', text: has('code.' + it.errorCode) ? t('code.' + it.errorCode) : it.errorCode }) : h('span', { class: 'text-secondary-400', text: '-' }); }, skeleton: '40%' }
  ]
};

RESOURCES['news-reactions'] = {
  api: 'publication',
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) {
      var ic = /^(https?:)?\//.test(it.icon || '') ? img(it.icon) : h('span', { class: 'text-2xl leading-none', text: it.icon });
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, h('div', { class: 'h-9 w-9 shrink-0 flex items-center justify-center' }, ic),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '45%' },
    { label: 'col.rx.uses', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.uses || 0) }); }, width: 120 },
    { label: 'col.sort', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.sort || 0) }); }, width: 90 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.rx.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.rx.active.hint' }
    ] },
    { name: 'icon', type: 'text', label: 'f.rx.icon', hint: 'f.rx.icon.hint', required: true },
    tr('name', { label: 'f.rx.name', required: true, max: 40 }),
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function () { return { key: '', icon: '', name: {}, sort: 0, isActive: true }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; },
  afterSave: function () { forgetRefs('news-reactions'); }
};

var NEWS_SETTINGS_FIELDS = [
  { type: 'heading', label: 'nst2.h.client' },
  { type: 'row', fields: [
    { name: 'commentsPerMinute', type: 'number', label: 'f.nst.comments', hint: 'f.nst.comments.hint', min: 1, max: 600, required: true },
    { name: 'reactionsPerMinute', type: 'number', label: 'f.nst.reactions', hint: 'f.nst.reactions.hint', min: 1, max: 600, required: true }
  ] },
  { type: 'row', fields: [
    { name: 'unreadScope', type: 'select', label: 'f.nst.unread', hint: 'f.nst.unread.hint', required: true, noPlaceholder: true, options: opts('nst2.unread.', ['signup', 'all']) },
    { name: 'showExpired', type: 'bool', inline: true, label: 'f.nst.showExpired', hint: 'f.nst.showExpired.hint' }
  ] },
  { type: 'heading', label: 'nst2.h.publish' },
  { type: 'row', fields: [
    { name: 'realtimeOnPublish', type: 'bool', inline: true, label: 'f.nst.realtime', hint: 'f.nst.realtime.hint' },
    { name: 'notifyCategoryKey', type: 'ref', resource: 'notify-categories', api: pubApi, valueKey: 'key', label: 'f.nst.category', hint: 'f.nst.category.hint', required: true }
  ] },
  { type: 'heading', label: 'nst2.h.ingest' },
  { name: 'fillDefaultLang', type: 'bool', inline: true, label: 'f.nst.fill', hint: 'f.nst.fill.hint' },
  { type: 'row', fields: [
    { name: 'logRetentionDays', type: 'number', label: 'f.nst.retention', min: 1, max: 365, required: true },
    { name: 'logPayloadBytes', type: 'number', label: 'f.nst.payload', hint: 'f.nst.payload.hint', min: 0, max: 262144, required: true }
  ] },
  { type: 'heading', label: 'nst2.h.seo' },
  { name: 'slugStyle', type: 'select', label: 'f.nst.slug', hint: 'f.nst.slug.hint', required: true, noPlaceholder: true, options: opts('nst2.slug.', ['native', 'latin', 'id']) },
  { name: 'publicEnabled', type: 'bool', inline: true, label: 'f.nst.public', hint: 'f.nst.public.hint' },
  { type: 'row', fields: [
    { name: 'publicBaseUrl', type: 'text', ltr: true, label: 'f.nst.base', hint: 'f.nst.base.hint', optional: true,
      check: function (v, vals) { return (vals.publicEnabled || vals.rssEnabled) && !v ? t('code.PUBLIC_BASE_URL_REQUIRED') : v && !/^https?:\/\/[^\s"'<>`]+$/i.test(v.trim()) ? t('code.INVALID_URL') : null; } },
    { name: 'publicPath', type: 'text', ltr: true, label: 'f.nst.path', hint: 'f.nst.path.hint', required: true,
      check: function (v) { return v && (v.charAt(0) !== '/' || v.indexOf('{slug}') === -1) ? t('code.INVALID_PATH') : null; } }
  ] },
  { type: 'row', fields: [
    { name: 'rssEnabled', type: 'bool', inline: true, label: 'f.nst.rss', hint: 'f.nst.rss.hint' },
    { name: 'publicCacheSeconds', type: 'number', label: 'f.nst.cache', min: 0, max: 86400, required: true }
  ] },
  tr('siteTitle', { label: 'f.nst.siteTitle', hint: 'f.nst.siteTitle.hint', max: 120, optional: true }),
  { name: 'publicLinks', type: 'static', readonly: true, render: function () {
    return h('div', { class: 'flex flex-col gap-1 text-xs', 'data-news-public-links': '' },
      h('p', { class: 'text-secondary-600 dark:text-secondary-300', text: t('nst2.links') }),
      ['/v2/publication/public/news', '/v2/publication/public/news/{slug}', '/v2/publication/public/sitemap.xml', '/v2/publication/public/rss.xml'].map(function (p) { return h('code', { class: 'font-mono', dir: 'ltr', text: 'GET ' + p }); }));
  } }
];

function newsSettingsPage(container) {
  var frame = pageFrame(t('nav.news-settings'), t('desc.news-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(pubApi, '/news-settings', null, NEWS_SETTINGS_FIELDS, {
    id: 'news',
    fixPayload: function (p) {
      delete p.updatedAt;
      delete p.updatedBy;
      delete p.publicLinks;
      p.publicBaseUrl = (p.publicBaseUrl || '').trim();
      return p;
    },
    afterSave: function (res) { NEWS_SET = res || NEWS_SET; }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['news-review'] = function (root, route) { newsReviewPage(root, route.query || {}); };
CUSTOM_PAGES['news-settings'] = function (root) { newsSettingsPage(root); };
