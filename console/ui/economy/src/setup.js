var SETUP_ORDER = ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review'];
var SETUP_ICONS = { brand: 'image', languages: 'languages', currency: 'currencies', payment: 'payments', sms: 'comments', email: 'send', review: 'check' };
var SETUP_STATUS_COLOR = { pending: 'warning', in_progress: 'primary', completed: 'success', skipped: 'secondary' };
var STEP_COLOR = { done: 'success', skipped: 'secondary', pending: 'warning' };
var QUICK_LANGS = [
  { code: 'fa', native: 'فارسی', dir: 'rtl', fa: 'فارسی', en: 'Persian' },
  { code: 'en', native: 'English', dir: 'ltr', fa: 'انگلیسی', en: 'English' },
  { code: 'ar', native: 'العربية', dir: 'rtl', fa: 'عربی', en: 'Arabic' },
  { code: 'tr', native: 'Türkçe', dir: 'ltr', fa: 'ترکی', en: 'Turkish' },
  { code: 'ru', native: 'Русский', dir: 'ltr', fa: 'روسی', en: 'Russian' },
  { code: 'de', native: 'Deutsch', dir: 'ltr', fa: 'آلمانی', en: 'German' },
  { code: 'fr', native: 'Français', dir: 'ltr', fa: 'فرانسوی', en: 'French' },
  { code: 'es', native: 'Español', dir: 'ltr', fa: 'اسپانیایی', en: 'Spanish' },
  { code: 'ur', native: 'اردو', dir: 'rtl', fa: 'اردو', en: 'Urdu' },
  { code: 'zh', native: '中文', dir: 'ltr', fa: 'چینی', en: 'Chinese' }
];
var setupClaimFor = null;
var wiz = null;

function setupStepStatus(state, k) {
  return (state && state.steps && state.steps[k]) || 'pending';
}

function quietBase(method, path, body) {
  var tk = token();
  if (!tk) return Promise.resolve(null);
  var headers = { Accept: 'application/json', Authorization: 'Bearer ' + tk };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  return fetch(BASE_BASE + path, { method: method, headers: headers, body: body === undefined ? undefined : JSON.stringify(body) }).then(function (res) {
    return res.ok ? res.json() : null;
  }, function () { return null; });
}

function maybeClaimSetup() {
  var tk = token();
  if (!tk || wiz || setupClaimFor === tk || /^#\/login/.test(location.hash)) return;
  if (!document.querySelector('.x-drawer a[href="#/players"]')) return;
  setupClaimFor = tk;
  quietBase('POST', '/setup/claim').then(function (res) {
    if (res && res.claimed && res.setup && !wiz) openWizard(res.setup, true);
  });
}

function putSetup(body) {
  return baseApi('PUT', '/setup', body).then(function (st) {
    if (wiz) { wiz.state = st; paintBar(); }
    return st;
  });
}

function focusables(root) {
  return Array.prototype.filter.call(root.querySelectorAll('a[href],button:not([disabled]),input:not([disabled]):not([type="hidden"]):not(.hidden),select:not([disabled]),textarea:not([disabled]),summary,[tabindex]:not([tabindex="-1"])'), function (el) {
    return el.offsetParent !== null || el === document.activeElement;
  });
}

function openWizard(state, auto, step) {
  if (wiz) { goStep(step || wiz.step); return; }
  var dir = document.documentElement.getAttribute('dir') || 'ltr';
  var bar = h('ol', { class: 'setup-steps', 'aria-label': t('setup.steps') });
  var body = h('div', { class: 'setup-body' });
  var footer = h('div', { class: 'setup-footer' });
  var later = button({ variant: 'ghost', size: 'sm', label: t('setup.later'), onclick: function () { closeWizard(true); } });
  later.setAttribute('data-setup-later', '');
  var title = h('div', { class: 'setup-title' },
    h('div', { class: 'setup-title-mark' }, icon('flask', 'h-5 w-5')),
    h('div', { class: 'min-w-0' },
      h('h1', { id: 'setup-title', class: 'text-lg font-semibold truncate', text: t('setup.title') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 truncate', text: t('setup.subtitle') })));
  var el = h('div', { class: 'setup-overlay', 'data-economy': 'setup', 'data-l10n-skip': '', dir: dir, role: 'dialog', 'aria-modal': 'true', 'aria-labelledby': 'setup-title' },
    h('header', { class: 'setup-header' }, title, later),
    h('nav', { class: 'setup-nav', 'aria-label': t('setup.steps') }, bar),
    h('main', { class: 'setup-main' }, h('div', { class: 'setup-inner' }, body)),
    footer);
  wiz = { el: el, state: state, step: null, ctl: null, bar: bar, body: body, footer: footer, prevFocus: document.activeElement, auto: !!auto };
  document.addEventListener('keydown', trapKeys, true);
  document.addEventListener('keydown', wizardEsc);
  document.documentElement.classList.add('setup-open');
  document.body.appendChild(el);
  if (!auto && state && (state.status === 'pending' || state.status === 'skipped')) putSetup({ status: 'in_progress' }).then(null, function () {});
  goStep(step || (state && state.nextStep) || 'brand', true);
}

function trapKeys(e) {
  if (e.key !== 'Tab' || !wiz || document.querySelector('[data-economy="dialog"]')) return;
  trapTab(e, wiz.el);
}

function trapTab(e, root) {
  var list = focusables(root).filter(function (el) { return el.getAttribute('tabindex') !== '-1'; });
  if (!list.length) return;
  var first = list[0];
  var last = list[list.length - 1];
  var active = document.activeElement;
  if (active && root.contains(active) && list.indexOf(active) === -1) {
    var after = list.some(function (el) { return active.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING; });
    var before = list.some(function (el) { return active.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_PRECEDING; });
    if (!e.shiftKey && !after) { e.preventDefault(); first.focus(); }
    else if (e.shiftKey && !before) { e.preventDefault(); last.focus(); }
    return;
  }
  if (e.shiftKey && (document.activeElement === first || !root.contains(document.activeElement))) { e.preventDefault(); last.focus(); }
  else if (!e.shiftKey && (document.activeElement === last || !root.contains(document.activeElement))) { e.preventDefault(); first.focus(); }
}

function wizardEsc(e) {
  if (e.key !== 'Escape' || !wiz || e.defaultPrevented) return;
  if (document.querySelector('[data-economy="dialog"]')) return;
  e.preventDefault();
  confirmModal({ title: t('setup.later.title'), text: t('setup.later.text'), ok: t('setup.later'), cancel: t('setup.stay'), icon: 'alert' }).then(function (yes) {
    if (yes) closeWizard(true);
  });
}

function closeWizard(later, done) {
  if (!wiz) return;
  var w = wiz;
  var go = function () {
    w.el.remove();
    document.removeEventListener('keydown', wizardEsc);
    document.removeEventListener('keydown', trapKeys, true);
    document.documentElement.classList.remove('setup-open');
    wiz = null;
    if (w.prevFocus && w.prevFocus.isConnected) { try { w.prevFocus.focus(); } catch (e) {} }
    psRefresh('setup');
  };
  if (done) { go(); return; }
  if (later && (!w.ctl || !w.ctl.dirty || !w.ctl.dirty())) {
    putSetup({ status: 'skipped' }).then(null, function () {});
    toast('success', t('setup.later.toast'));
    go();
    return;
  }
  confirmModal({ title: t('setup.unsaved.title'), text: t('setup.unsaved.text'), ok: t('setup.unsaved.ok'), danger: true }).then(function (yes) {
    if (!yes) return;
    putSetup({ status: 'skipped' }).then(null, function () {});
    toast('success', t('setup.later.toast'));
    go();
  });
}

function paintBar() {
  if (!wiz) return;
  clear(wiz.bar);
  SETUP_ORDER.forEach(function (k, i) {
    var st = k === 'review' ? (wiz.state && wiz.state.status === 'completed' ? 'done' : 'pending') : setupStepStatus(wiz.state, k);
    var on = wiz.step === k;
    var mark = st === 'done' ? icon('check', 'h-3.5 w-3.5') : st === 'skipped' ? h('span', { text: '–' }) : h('span', { text: localDigits(i + 1) });
    var b = h('button', { type: 'button', class: 'setup-step setup-step--' + st + (on ? ' setup-step--on' : ''), 'data-step': k, 'aria-current': on ? 'step' : null,
      'aria-label': localDigits(i + 1) + '. ' + t('setup.step.' + k) + ' (' + t('setup.st.' + st) + ')',
      onclick: function () { goStep(k); } },
      h('span', { class: 'setup-step-num', 'aria-hidden': 'true' }, mark),
      h('span', { class: 'setup-step-label', text: t('setup.step.' + k) }));
    wiz.bar.appendChild(h('li', {}, b));
  });
  var cur = wiz.bar.querySelector('.setup-step--on');
  if (cur && cur.scrollIntoView) { try { cur.scrollIntoView({ block: 'nearest', inline: 'center' }); } catch (e) {} }
}

function goStep(k, force) {
  if (!wiz) return;
  if (!force && wiz.step === k) return;
  var leave = !force && wiz.ctl && wiz.ctl.dirty && wiz.ctl.dirty() ? confirmModal({ title: t('setup.unsaved.title'), text: t('setup.unsaved.text'), ok: t('setup.unsaved.ok'), danger: true }) : Promise.resolve(true);
  leave.then(function (yes) {
    if (!yes || !wiz) return;
    wiz.step = k;
    paintBar();
    renderStep(k);
  });
}

function nextOf(k) {
  var i = SETUP_ORDER.indexOf(k);
  return SETUP_ORDER[Math.min(i + 1, SETUP_ORDER.length - 1)];
}

function prevOf(k) {
  var i = SETUP_ORDER.indexOf(k);
  return SETUP_ORDER[Math.max(i - 1, 0)];
}

function renderStep(k) {
  var w = wiz;
  var i = SETUP_ORDER.indexOf(k);
  var st = k === 'review' ? null : setupStepStatus(w.state, k);
  var content = h('div', { class: 'setup-content' });
  clear(w.body).appendChild(h('section', { class: 'setup-section', 'aria-labelledby': 'setup-step-title', 'data-setup-step': k },
    h('div', { class: 'setup-section-head' },
      h('div', { class: 'setup-section-icon' }, icon(SETUP_ICONS[k], 'h-5 w-5')),
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('setup.stepOf', { n: localDigits(i + 1), total: localDigits(SETUP_ORDER.length) }) }),
        h('h2', { id: 'setup-step-title', class: 'text-xl font-semibold', tabindex: '-1', text: t('setup.step.' + k + '.title') }),
        h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mt-1', text: t('setup.step.' + k + '.desc') })),
      st ? tag(t('setup.st.' + st), STEP_COLOR[st], 'shrink-0') : null),
    content));
  var ctl = STEP_BUILDERS[k](content, w);
  w.ctl = ctl;
  paintFooter(k, ctl);
  var hd = w.body.querySelector('#setup-step-title');
  if (hd) { try { hd.focus({ preventScroll: true }); } catch (e) { hd.focus(); } }
  var main = w.el.querySelector('.setup-main');
  if (main) main.scrollTop = 0;
}

function paintFooter(k, ctl) {
  var w = wiz;
  clear(w.footer);
  var first = k === SETUP_ORDER[0];
  var review = k === 'review';
  var back = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('setup.back'), disabled: first, onclick: function () { goStep(prevOf(k)); } });
  var skip = review ? null : button({ variant: 'muted', size: 'sm', label: t('setup.skip'), onclick: function () {
    setBusy(skip, true);
    putSetup({ steps: stepPatch(k, 'skipped') }).then(function () { goStep(nextOf(k), true); }, function (err) { setBusy(skip, false); toast('error', errorText(err)); });
  } });
  if (skip) skip.setAttribute('data-setup-skip', '');
  var primary = button({ variant: 'primary', size: 'sm', icon: review ? 'check' : null, label: review ? t('setup.finish') : t('setup.saveNext'), onclick: function () {
    setBusy(primary, true, t('saving'));
    Promise.resolve(ctl.save()).then(function (result) {
      if (!result || !wiz) { setBusy(primary, false); return; }
      if (review) {
        return putSetup({ status: 'completed' }).then(function () {
          toast('success', t('setup.done.toast'));
          closeWizard(false, true);
        });
      }
      return putSetup({ steps: stepPatch(k, result) }).then(function () { goStep(nextOf(k), true); });
    }).then(null, function (err) {
      setBusy(primary, false);
      toast('error', errorText(err));
    });
  } });
  primary.setAttribute('data-setup-next', '');
  w.footer.appendChild(h('div', { class: 'setup-footer-inner' },
    back,
    h('div', { class: 'flex items-center gap-2' }, skip, primary)));
}

function stepPatch(k, v) {
  var o = {};
  o[k] = v;
  return o;
}

function stepSkeleton(rows) {
  var box = h('div', { class: 'setup-card' });
  for (var i = 0; i < rows; i++) box.appendChild(h('div', { class: 'flex items-center gap-3 py-3 border-b border-secondary-200 dark:border-secondary-700' }, bar('2rem', 'h-8'), h('div', { class: 'flex-1' }, bar('40%', 'h-3 mb-2'), bar('25%', 'h-3')), bar('3rem', 'h-5 rounded-full')));
  return box;
}

function questionSkeleton(cards) {
  var box = h('div', {}, h('div', { class: 'setup-card mb-4' }, bar('12rem', 'h-4 mb-3'), h('div', { class: 'flex gap-2' }, bar('6rem', 'h-9'), bar('6rem', 'h-9'))));
  for (var i = 0; i < cards; i++) box.appendChild(h('div', { class: 'setup-card mb-3' }, bar('8rem', 'h-4 mb-4'), bar('100%', 'h-12 mb-4'), skeletonField(), skeletonField()));
  return box;
}

function card2(children, extra) {
  return h('div', { class: 'setup-card' + (extra ? ' ' + extra : '') }, children);
}

function failBox(container, err, retry) {
  clear(container).appendChild(errorBlock(err, retry));
}

function stepBrand(box) {
  var editor = null;
  box.appendChild(card2(brandFormSkeleton()));
  function load() {
    Promise.all([baseApi('GET', '/brand'), freshLanguages()]).then(function (all) {
      if (!box.isConnected) return;
      editor = brandEditor(all[0]);
      clear(box).appendChild(card2(editor.el));
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () { return editor ? saveBrand(editor).then(function (ok) { return ok ? 'done' : false; }) : false; },
    dirty: function () { return !!editor && editor.dirty(); }
  };
}

function quickLangBody(q) {
  var name = { fa: q.fa, en: q.en };
  if (LANGS.def !== 'fa' && LANGS.def !== 'en') name[LANGS.def] = q.native;
  Object.keys(name).forEach(function (k) { if (LANGS.codes.indexOf(k) === -1 && k !== q.code) delete name[k]; });
  name[q.code] = q.native;
  return { code: q.code, name: name, nativeName: q.native, direction: q.dir, isActive: true, isDefault: false, sort: LANGS.items.length + 1 };
}

function stepLanguages(box) {
  box.appendChild(stepSkeleton(3));
  var busy = false;
  function load() {
    baseApi('GET', '/languages?limit=100').then(function (res) {
      if (!box.isConnected) return;
      paint((res && res.items) || []);
    }, function (err) { failBox(box, err, load); });
  }
  function act(p, done) {
    if (busy) return;
    busy = true;
    p.then(function () {
      busy = false;
      languagesChanged();
      if (done) toast('success', done);
      load();
    }, function (err) {
      busy = false;
      toast('error', errorText(err));
      load();
    });
  }
  function paint(items) {
    var list = h('ul', { class: 'setup-list', 'aria-label': t('setup.lang.list') });
    items.forEach(function (it) {
      var sw = toggle(it.isActive, null, function (on) {
        act(baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: on })), t(on ? 'lang.activated' : 'lang.deactivated', { name: it.nativeName || it.code }));
      }, 'sm');
      sw.setAttribute('aria-label', t('setup.lang.active', { name: it.nativeName || it.code }));
      if (it.isDefault) { sw.classList.add('opacity-60', 'pointer-events-none'); sw.setAttribute('aria-disabled', 'true'); sw.removeAttribute('tabindex'); }
      var def = it.isDefault ? null : button({ variant: 'link', size: 'sm', label: t('setup.lang.makeDefault'), onclick: function () {
        act(baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: true, isDefault: true })), t('setup.lang.defaultSet', { name: it.nativeName || it.code }));
      } });
      list.appendChild(h('li', { class: 'setup-list-row', 'data-lang': it.code },
        h('div', { class: 'min-w-0 flex-1' }, languageCell(it)),
        h('div', { class: 'flex items-center gap-3 shrink-0' }, def, h('div', { class: 'flex items-center gap-2 text-xs text-secondary-600 dark:text-secondary-300' }, sw, h('span', { text: it.isActive ? t('active') : t('inactive') })))));
    });
    var have = items.map(function (x) { return x.code; });
    var chips = h('div', { class: 'setup-chips' });
    QUICK_LANGS.filter(function (q) { return have.indexOf(q.code) === -1; }).forEach(function (q) {
      chips.appendChild(h('button', { type: 'button', class: 'setup-chip', 'data-add-lang': q.code, onclick: function () {
        act(baseApi('POST', '/languages', quickLangBody(q)), t('setup.lang.added', { name: q.native }));
      } }, icon('plus', 'h-3.5 w-3.5'), h('span', { dir: q.dir, text: q.native }), h('span', { class: 'setup-chip-code', dir: 'ltr', text: q.code })));
    });
    clear(box).appendChild(card2(list));
    box.appendChild(card2([
      h('div', { class: 'text-sm font-medium mb-1', text: t('setup.lang.quick') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('setup.lang.quick.hint') }),
      chips.children.length ? chips : h('p', { class: 'text-sm text-secondary-500', text: t('setup.lang.quick.none') })
    ], 'mt-4'));
  }
  load();
  return { save: function () { return 'done'; }, dirty: function () { return false; } };
}

var SETUP_CURRENCY_FIELDS = [
  { type: 'row', fields: [
    { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, keyRe: CURRENCY_KEY_RE, lockOnEdit: true },
    { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('type.', ['soft', 'hard']) }
  ] },
  { name: 'name', type: 'i18n', label: 'f.name', required: true, max: 128 },
  { name: 'defaultValue', type: 'number', label: 'setup.cur.defaultValue', hint: 'setup.cur.defaultValue.hint', min: 0 },
  { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true }
];

function stepCurrency(box) {
  var form = null;
  var current = null;
  var dirty = false;
  var items = [];
  var top = h('div', {});
  box.appendChild(card2([bar('10rem', 'h-4 mb-3'), bar('100%', 'h-10 mb-5'), h('div', { class: 'eco-form-row' }, skeletonField(), skeletonField()), skeletonField('h-16'), skeletonField(), skeletonField('h-16')]));
  function build(item) {
    current = item;
    dirty = false;
    var values = clone(item) || { type: 'soft', name: {}, defaultValue: 0, sort: 0, metadata: {} };
    form = buildForm(SETUP_CURRENCY_FIELDS, values, { onchange: function () { dirty = true; } });
    dirty = false;
    var num = form.el.querySelector('input[inputmode="numeric"]');
    if (num) {
      num.setAttribute('dir', 'ltr');
      num.value = toLatinDigits(num.value);
      num.addEventListener('input', function () { latinize(num); });
      num.addEventListener('blur', function () { num.value = toLatinDigits(num.value); });
    }
    return form.el;
  }
  function paint() {
    var formBox = h('div', {});
    var pick = null;
    if (items.length) {
      var options = items.map(function (c) { return { value: c.id, label: (i18nText(c.name) || c.key) + ' (' + c.key + ')' }; });
      options.push({ value: '', label: t('setup.cur.new') });
      pick = selectInput(current ? current.id : '', options, function (v) {
        var next = items.filter(function (c) { return c.id === v; })[0] || null;
        clear(formBox).appendChild(build(next));
        clear(top);
      }, { label: t('setup.cur.pick'), placeholder: false });
    }
    clear(box).appendChild(card2([
      pick ? field(t('setup.cur.pick'), pick, t('setup.cur.pick.hint')) : h('div', { class: 'mb-4' }, notice('info', t('setup.cur.none'))),
      top,
      formBox
    ]));
    formBox.appendChild(build(current));
  }
  function load() {
    api('GET', '/currencies?limit=100').then(function (res) {
      if (!box.isConnected) return;
      items = ((res && res.items) || []).slice().sort(function (a, b) { return (a.sort || 0) - (b.sort || 0) || String(a.key).localeCompare(String(b.key)); });
      current = items[0] || null;
      paint();
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () {
      if (!form) return false;
      clear(top);
      if (!form.validate()) return false;
      var payload = form.payload();
      var body = current ? { key: current.key, type: current.type, name: current.name, defaultValue: current.defaultValue, iconUrl: current.iconUrl, categoryId: current.categoryId || null, tagIds: current.tagIds || [], metadata: current.metadata || {}, sort: current.sort || 0 } : { metadata: {}, sort: 0 };
      Object.keys(payload).forEach(function (k) { body[k] = payload[k]; });
      if (body.iconUrl == null) body.iconUrl = '';
      return api(current ? 'PUT' : 'POST', '/currencies' + (current ? '/' + encodeURIComponent(current.id) : ''), body).then(function (saved) {
        toast('success', t('saved'));
        dirty = false;
        if (saved && saved.id) {
          var idx = items.map(function (c) { return c.id; }).indexOf(saved.id);
          if (idx === -1) items.push(saved); else items[idx] = saved;
          current = saved;
        }
        return 'done';
      }, function (err) {
        if (!form.showServerError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        return false;
      });
    },
    dirty: function () { return dirty; }
  };
}

function needQuestion(labelKey, value, onchange) {
  var seg = segmented([{ value: 'yes', label: t('yes') }, { value: 'no', label: t('setup.no') }], value, onchange, t(labelKey), 'md');
  seg.setAttribute('data-need', '');
  return card2(h('div', { class: 'flex flex-col sm:flex-row sm:items-center justify-between gap-3' },
    h('div', { class: 'font-medium', text: t(labelKey) }), seg), 'mb-4');
}

function sandboxBox(on, onchange, confirmOff) {
  var text = h('p', { class: 'text-xs mt-1.5' });
  var boxEl = h('div', { 'data-sandbox-box': '' });
  var sw = toggle(on, t('pv.sandbox'), function (v) {
    if (!v && confirmOff) {
      confirmModal({ title: t('pv.live.title'), text: t('pv.live.text'), ok: t('pv.live.ok'), danger: true, icon: 'alert' }).then(function (yes) {
        if (!yes) { sw.set(true); return; }
        on = false; paint(); onchange(false);
      });
      return;
    }
    on = v; paint(); onchange(v);
  });
  boxEl.appendChild(h('div', { class: 'flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), sw));
  boxEl.appendChild(text);
  function paint() {
    boxEl.className = 'rounded-md px-3 py-2.5 mb-4 setup-sandbox ' + (on ? 'eco-sandbox-note' : 'eco-live-note');
    text.textContent = on ? t('setup.sandbox.on') : t('setup.sandbox.off');
  }
  paint();
  return boxEl;
}

function stepPayment(box) {
  var methods = [];
  var picked = {};
  var cfgs = {};
  var els = {};
  var need = null;
  var dirty = false;
  var top = h('div', {});
  box.appendChild(questionSkeleton(2));
  function methodCard(m) {
    var cfg = cfgs[m.method];
    var fieldEls = {};
    var fieldsBox = h('div', { class: 'setup-method-fields' });
    var sandboxField = (m.fields || []).filter(function (f) { return f.name === 'sandbox'; })[0];
    if (sandboxField) fieldsBox.appendChild(sandboxBox(cfg.sandbox !== false, function (v) { cfg.sandbox = v; dirty = true; }, true));
    (m.fields || []).forEach(function (f) {
      if (f.type === 'bool') return;
      var label = has('pm.field.' + f.name) ? t('pm.field.' + f.name) : f.name;
      var ctl = textInput(cfg[f.name] == null ? '' : cfg[f.name], { ltr: true, mono: true, multiline: f.type === 'textarea', rows: 3, label: label, oninput: function (v) { cfg[f.name] = v; dirty = true; fieldEls[f.name].setError(''); } });
      ctl.classList.add('text-sm');
      if (f.type === 'secret') ctl.classList.add('eco-secret');
      var hint = f.type === 'secret' && cfg[f.name] ? t('pm.secret.hint') : f.type === 'secret' && sandboxField ? t('setup.pay.secret.sandbox') : null;
      fieldEls[f.name] = field(label, ctl, hint, { optional: !f.required });
      fieldsBox.appendChild(fieldEls[f.name]);
    });
    if (!(m.fields || []).length) fieldsBox.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('pm.noFields') }));
    var errBox = h('div', {});
    var check = checkbox(picked[m.method], has('method.' + m.method) ? t('method.' + m.method) : m.method, function (on) {
      picked[m.method] = on;
      dirty = true;
      fieldsBox.hidden = !on;
      el.classList.toggle('setup-method--on', on);
    });
    check.setAttribute('data-pick-method', m.method);
    fieldsBox.hidden = !picked[m.method];
    var el = card2([
      h('div', { class: 'flex items-center justify-between gap-3' }, check,
        h('div', { class: 'flex items-center gap-1.5' }, mono(m.method), tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'success' : 'secondary'))),
      errBox,
      fieldsBox
    ], 'setup-method mb-3' + (picked[m.method] ? ' setup-method--on' : ''));
    el.setAttribute('data-method', m.method);
    els[m.method] = { el: el, fields: fieldEls, err: errBox };
    return el;
  }
  function paint() {
    var list = h('div', { 'data-methods': '' });
    methods.forEach(function (m) { list.appendChild(methodCard(m)); });
    list.hidden = need !== 'yes';
    var noBox = h('div', { hidden: need !== 'no' }, notice('info', t('setup.pay.no')));
    clear(box).appendChild(needQuestion('setup.pay.need', need, function (v) { need = v; dirty = true; list.hidden = v !== 'yes'; noBox.hidden = v !== 'no'; }));
    box.appendChild(top);
    box.appendChild(noBox);
    box.appendChild(list);
  }
  function load() {
    api('GET', '/payment-methods').then(function (res) {
      if (!box.isConnected) return;
      methods = Array.isArray(res) ? res : (res && (res.items || res.methods)) || [];
      methods.forEach(function (m) {
        picked[m.method] = !!m.isActive;
        cfgs[m.method] = clone(m.config || {}) || {};
        if (!m.configured && !m.isActive && (m.fields || []).some(function (f) { return f.name === 'sandbox'; })) cfgs[m.method].sandbox = true;
      });
      need = methods.some(function (m) { return m.isActive; }) ? 'yes' : null;
      paint();
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () {
      clear(top);
      if (!need) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.err.answer')))); return false; }
      if (need === 'no') return 'skipped';
      var changed = methods.filter(function (m) { return picked[m.method] !== !!m.isActive || (picked[m.method] && JSON.stringify(cfgs[m.method]) !== JSON.stringify(m.config || {})); });
      if (!methods.some(function (m) { return picked[m.method]; })) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.pay.pickOne')))); return false; }
      var failed = false;
      return changed.reduce(function (p, m) {
        return p.then(function () {
          var e = els[m.method];
          clear(e.err);
          Object.keys(e.fields).forEach(function (k) { e.fields[k].setError(''); });
          return api('PUT', '/payment-methods/' + encodeURIComponent(m.method), { isActive: !!picked[m.method], config: cfgs[m.method] }).then(function (res) {
            if (res && res.method) { m.isActive = res.isActive; m.config = res.config; m.configured = res.configured; cfgs[m.method] = clone(res.config || {}) || {}; }
          }, function (err) {
            failed = true;
            if (err.field && e.fields[err.field]) e.fields[err.field].setError(errorText(err));
            else e.err.appendChild(h('div', { class: 'mt-3' }, notice('error', errorText(err))));
            e.el.scrollIntoView({ block: 'center', behavior: 'smooth' });
          });
        });
      }, Promise.resolve()).then(function () {
        if (failed) return false;
        dirty = false;
        if (changed.length) toast('success', t('saved'));
        return 'done';
      });
    },
    dirty: function () { return dirty; }
  };
}

function stepProvider(channel) {
  return function (box) {
    var catalog = [];
    var items = [];
    var need = null;
    var kind = null;
    var cfg = {};
    var st = { sandbox: true };
    var fieldEls = {};
    var dirty = false;
    var top = h('div', {});
    var formBox = h('div', {});
    var testBoxEl = h('div', {});
    box.appendChild(questionSkeleton(1));
    function current() { return items.filter(function (p) { return p.kind === kind; })[0] || null; }
    function buildForm2() {
      var p = current();
      var def = catalog.filter(function (k) { return k.channel === channel && k.kind === kind; })[0] || { fields: [] };
      cfg = p ? clone(p.config || {}) || {} : {};
      st = { sandbox: p ? !!p.sandbox : true };
      fieldEls = {};
      clear(formBox);
      formBox.appendChild(sandboxBox(st.sandbox, function (v) { st.sandbox = v; dirty = true; paintTest(); }, true));
      (def.fields || []).forEach(function (f) {
        var w = providerField(f, cfg, function () { dirty = true; if (fieldEls[f.name]) fieldEls[f.name].setError(''); });
        fieldEls[f.name] = w;
        formBox.appendChild(w);
      });
      if (!p) formBox.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 -mt-2 mb-2', text: t('setup.pv.new') }));
      paintTest();
    }
    function saveProvider() {
      clear(top);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      var p = current();
      var def = catalog.filter(function (k) { return k.channel === channel && k.kind === kind; })[0] || { fields: [] };
      var out = {};
      (def.fields || []).forEach(function (f) { if (cfg[f.name] !== undefined && cfg[f.name] !== null) out[f.name] = String(cfg[f.name]); });
      var body = { isActive: true, isDefault: true, sandbox: st.sandbox, config: out };
      var call = p ? pubApi('PUT', '/providers/' + encodeURIComponent(p.id), body) : pubApi('POST', '/providers', Object.assign({ channel: channel, kind: kind }, body));
      return call.then(function (saved) {
        var idx = items.map(function (x) { return x.id; }).indexOf(saved.id);
        if (idx === -1) items.push(saved); else items[idx] = saved;
        items.forEach(function (x) { if (x.id !== saved.id) x.isDefault = false; });
        dirty = false;
        buildForm2();
        return true;
      }, function (err) {
        if (err.field && fieldEls[err.field]) { fieldEls[err.field].setError(errorText(err)); fieldEls[err.field].scrollIntoView({ block: 'center', behavior: 'smooth' }); }
        else top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        return false;
      });
    }
    function paintTest() {
      clear(testBoxEl);
      var email = channel === 'email';
      var p = current();
      var to = textInput('', { ltr: true, mono: true, label: t('pv.test.to'), placeholder: email ? 'name@example.com' : '09xxxxxxxxx', inputmode: email ? 'email' : 'tel' });
      to.classList.add('text-sm');
      var subject = email ? textInput('', { label: t('pv.test.subject'), placeholder: t('pv.test.subject.ph') }) : null;
      if (subject) subject.classList.add('text-sm');
      var msg = textInput(t('pv.test.message.default'), { multiline: true, rows: 2, label: t('pv.test.message') });
      msg.classList.add('text-sm');
      var out = h('div', { 'aria-live': 'polite' });
      var live = !st.sandbox;
      var sendBtn = button({ variant: 'link', size: 'sm', icon: 'send', label: t('setup.test.send'), disabled: live, onclick: function () {
        clear(out);
        setBusy(sendBtn, true, t('nt.sending'));
        var before = dirty || !current() ? saveProvider() : Promise.resolve(true);
        before.then(function (ok) {
          var cur = current();
          if (!ok || !cur) { setBusy(sendBtn, false); return; }
          if (!cur.sandbox) { setBusy(sendBtn, false); out.appendChild(notice('error', t('setup.test.liveBlocked'))); return; }
          var payload = { to: toLatinDigits(to.value).trim(), message: msg.value };
          if (subject) payload.subject = subject.value;
          return pubApi('POST', '/providers/' + encodeURIComponent(cur.id) + '/test', payload).then(function (r) {
            setBusy(sendBtn, false);
            if (r && r.ok && r.sandbox) out.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-start gap-2', 'data-test-result': 'sandbox' }, icon('flask', 'h-4 w-4 shrink-0 mt-0.5'), h('span', {}, h('span', { text: t('pv.test.sandbox') + ' ' }), r.reply ? mono(r.reply) : null)));
            else if (r && r.ok) out.appendChild(h('div', { 'data-test-result': 'ok' }, notice('success', t('pv.test.ok'))));
            else out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', t('pv.test.failed') + ' ' + ((r && r.error) || ''))));
          }, function (err) {
            setBusy(sendBtn, false);
            var bad = err.field === 'to';
            out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', errorText(err))));
            if (bad) to.focus();
          });
        });
      } });
      testBoxEl.appendChild(card2([
        h('div', { class: 'flex items-center gap-2 font-medium mb-1' }, icon('send', 'h-4 w-4'), h('span', { text: t('setup.test') })),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: live ? t('setup.test.liveBlocked') : t('setup.test.hint') }),
        field(t('pv.test.to'), to),
        subject ? field(t('pv.test.subject'), subject, null, { optional: true }) : null,
        field(t('pv.test.message'), msg),
        h('div', { class: 'flex justify-end -mt-2' }, sendBtn),
        out
      ], 'mt-4'));
      if (p) testBoxEl.setAttribute('data-provider-id', p.id);
    }
    function paint() {
      var kinds = catalog.filter(function (k) { return k.channel === channel; });
      var seg = segmented(kinds.map(function (k) { return { value: k.kind, label: kindLabel(k.kind) }; }), kind, function (v) { kind = v; dirty = true; buildForm2(); clear(top); }, t('pv.kind'), 'md');
      seg.setAttribute('data-kind', '');
      var yesBox = h('div', { hidden: need !== 'yes' },
        card2([
          field(t('setup.pv.kind'), seg, t('setup.pv.kind.hint')),
          top,
          formBox
        ]),
        testBoxEl);
      var noBox = h('div', { hidden: need !== 'no' }, notice('info', t('setup.' + channel + '.no')));
      clear(box).appendChild(needQuestion('setup.' + channel + '.need', need, function (v) { need = v; dirty = true; yesBox.hidden = v !== 'yes'; noBox.hidden = v !== 'no'; }));
      box.appendChild(noBox);
      box.appendChild(yesBox);
      buildForm2();
    }
    function load() {
      pubApi('GET', '/providers').then(function (res) {
        if (!box.isConnected) return;
        catalog = (res && res.catalog) || [];
        items = ((res && res.items) || []).filter(function (p) { return p.channel === channel; });
        var def = items.filter(function (p) { return p.isDefault; })[0] || items.filter(function (p) { return p.isActive; })[0] || items[0];
        kind = def ? def.kind : (catalog.filter(function (k) { return k.channel === channel; })[0] || {}).kind;
        need = items.some(function (p) { return p.isActive; }) ? 'yes' : null;
        paint();
        dirty = false;
      }, function (err) { failBox(box, err, load); });
    }
    load();
    return {
      save: function () {
        clear(top);
        if (!need) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.err.answer')))); var q = box.querySelector('[data-need]'); if (q) q.scrollIntoView({ block: 'center' }); return false; }
        if (need === 'no') return 'skipped';
        return saveProvider().then(function (ok) { if (ok) toast('success', t('saved')); return ok ? 'done' : false; });
      },
      dirty: function () { return dirty; }
    };
  };
}

function stepReview(box, w) {
  box.appendChild(stepSkeleton(6));
  function load() {
    Promise.all([
      baseApi('GET', '/setup'),
      baseApi('GET', '/brand'),
      baseApi('GET', '/languages?limit=100'),
      api('GET', '/currencies?limit=100').then(null, function () { return null; }),
      api('GET', '/payment-methods').then(null, function () { return null; }),
      pubApi('GET', '/providers').then(null, function () { return null; })
    ]).then(function (all) {
      if (!box.isConnected || !wiz) return;
      wiz.state = all[0];
      paintBar();
      var b = all[1] || {};
      var langs = (all[2] && all[2].items) || [];
      var curs = ((all[3] && all[3].items) || []).slice().sort(function (a, c) { return (a.sort || 0) - (c.sort || 0); });
      var pms = Array.isArray(all[4]) ? all[4] : [];
      var pvs = (all[5] && all[5].items) || [];
      var summary = {
        brand: brandCustom(b) ? h('span', { class: 'flex items-center gap-2 min-w-0' }, b.accent ? h('span', { class: 'setup-dot', style: 'background-color:' + b.accent }) : null, h('span', { class: 'truncate', text: brandText(b.name) || b.shortName || t('setup.rv.brand.custom') })) : t('setup.rv.brand.default'),
        languages: langs.filter(function (l) { return l.isActive; }).map(function (l) { return (l.nativeName || l.code) + (l.isDefault ? ' (' + t('lang.default') + ')' : ''); }).join(t('setup.rv.sep')),
        currency: curs.length ? (i18nText(curs[0].name) || curs[0].key) + ' · ' + t('setup.rv.balance', { n: localDigits(fmtNum(curs[0].defaultValue || 0)) }) : t('setup.rv.none'),
        payment: pms.filter(function (m) { return m.isActive; }).map(function (m) { return t('method.' + m.method) + (m.config && m.config.sandbox ? ' (' + t('pv.sandbox.short') + ')' : ''); }).join(t('setup.rv.sep')) || t('setup.rv.none'),
        sms: provSummary(pvs, 'sms'),
        email: provSummary(pvs, 'email')
      };
      var list = h('ul', { class: 'setup-list' });
      SETUP_ORDER.slice(0, -1).forEach(function (k) {
        var s = setupStepStatus(wiz.state, k);
        list.appendChild(h('li', { class: 'setup-list-row', 'data-review': k },
          h('div', { class: 'setup-section-icon setup-section-icon--sm' }, icon(SETUP_ICONS[k], 'h-4 w-4')),
          h('div', { class: 'min-w-0 flex-1' },
            h('div', { class: 'font-medium', text: t('setup.step.' + k) }),
            h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 truncate' }, summary[k])),
          h('div', { class: 'flex items-center gap-2 shrink-0' },
            tag(t('setup.st.' + s), STEP_COLOR[s]),
            button({ variant: 'link', size: 'sm', label: t('setup.rv.edit'), title: t('setup.rv.editStep', { step: t('setup.step.' + k) }), onclick: function () { goStep(k); } }))));
      });
      var left = SETUP_ORDER.slice(0, -1).filter(function (k) { return setupStepStatus(wiz.state, k) === 'pending'; }).length;
      clear(box).appendChild(card2(list));
      box.appendChild(h('div', { class: 'mt-4' }, notice(left ? 'info' : 'success', left ? t('setup.rv.left', { n: localDigits(left) }) : t('setup.rv.ready'))));
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return { save: function () { return 'done'; }, dirty: function () { return false; } };
}

function provSummary(pvs, channel) {
  var list = pvs.filter(function (p) { return p.channel === channel && p.isActive; });
  if (!list.length) return t('setup.rv.none');
  var d = list.filter(function (p) { return p.isDefault; })[0] || list[0];
  return kindLabel(d.kind) + (d.sandbox ? ' (' + t('pv.sandbox.short') + ')' : '');
}

var STEP_BUILDERS = {
  brand: stepBrand,
  languages: stepLanguages,
  currency: stepCurrency,
  payment: stepPayment,
  sms: stepProvider('sms'),
  email: stepProvider('email'),
  review: stepReview
};

function setupPage(container) {
  var openBtn = button({ variant: 'primary', size: 'page', icon: 'flask', label: t('setup.open'), onclick: function () { openFromPage(); } });
  openBtn.setAttribute('data-setup-open', '');
  var frame = tabFrame(t('desc.setup'), openBtn);
  clear(container).appendChild(frame.el);
  var body = h('div', {});
  frame.content.appendChild(body);
  var state = null;
  function openFromPage(step) {
    openWizard(state, false, step);
  }
  function skeleton() {
    clear(body).appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'flex items-center justify-between mb-4' }, bar('10rem', 'h-5'), bar('5rem', 'h-5 rounded-full')), bar('100%', 'h-2 mb-4 rounded-full'), stepSkeleton(6))));
  }
  function load() {
    skeleton();
    baseApi('GET', '/setup').then(function (st) {
      if (!container.isConnected) return;
      state = st;
      var steps = SETUP_ORDER.slice(0, -1);
      var finished = steps.filter(function (k) { return setupStepStatus(st, k) !== 'pending'; }).length;
      var list = h('ul', { class: 'setup-list' });
      steps.forEach(function (k, i) {
        var s = setupStepStatus(st, k);
        list.appendChild(h('li', { class: 'setup-list-row', 'data-setup-row': k },
          h('div', { class: 'setup-section-icon setup-section-icon--sm' }, icon(SETUP_ICONS[k], 'h-4 w-4')),
          h('div', { class: 'min-w-0 flex-1' },
            h('div', { class: 'font-medium', text: localDigits(i + 1) + '. ' + t('setup.step.' + k) }),
            h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 truncate', text: t('setup.step.' + k + '.desc') })),
          h('div', { class: 'flex items-center gap-2 shrink-0' }, tag(t('setup.st.' + s), STEP_COLOR[s]),
            button({ variant: 'link', size: 'sm', label: t('setup.rv.edit'), title: t('setup.rv.editStep', { step: t('setup.step.' + k) }), onclick: function () { openFromPage(k); } }))));
      });
      var meta = [];
      if (st.firstOpenedAt) meta.push([t('setup.meta.opened'), (st.firstOpenedBy || '') + ' · ' + fmtDate(st.firstOpenedAt)]);
      if (st.completedAt) meta.push([t('setup.meta.completed'), (st.completedBy || '') + ' · ' + fmtDate(st.completedAt)]);
      clear(body).appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex flex-wrap items-center justify-between gap-3 mb-3' },
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'text-lg font-medium', text: t('setup.title') }), tag(t('setup.status.' + st.status), SETUP_STATUS_COLOR[st.status] || 'secondary')),
          h('span', { class: 'text-sm text-secondary-500 tabular-nums', text: t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) })),
        h('div', { class: 'setup-progress', role: 'progressbar', 'aria-valuemin': '0', 'aria-valuemax': String(steps.length), 'aria-valuenow': String(finished), 'aria-label': t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) },
          h('div', { class: 'setup-progress-bar', style: 'width:' + Math.round(finished / steps.length * 100) + '%' })),
        meta.length ? h('dl', { class: 'eco-dl mt-4 mb-2' }, meta.map(function (m) { return [h('dt', { class: 'text-secondary-500', text: m[0] }), h('dd', { text: m[1] })]; })) : null,
        list)));
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}
