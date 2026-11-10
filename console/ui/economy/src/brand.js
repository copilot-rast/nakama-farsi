var BRAND_KEY = 'nakama-brand';
var BRAND_SHADES = { 50: 0.9, 100: 0.8, 200: 0.6, 300: 0.4, 400: 0.1, 500: 0, 600: -0.1, 700: -0.25, 800: -0.58, 900: -0.7, 950: -0.8 };
var BRAND_DEFAULT_SHADES = { '244,244,253': 50, '233,232,252': 100, '209,205,248': 200, '181,175,245': 300, '130,117,240': 400, '107,89,238': 500, '102,84,226': 600, '93,77,206': 700, '69,58,154': 800, '59,49,130': 900, '48,40,106': 950 };
var BRAND_DEFAULT_ACCENT = '#6b59ee';
var BRAND_RGB_RE = /rgb\(\s*(\d{1,3})[\s,]+(\d{1,3})[\s,]+(\d{1,3})\s*\)/g;
var BRAND_HEX_RE = /^#?([0-9a-f]{3}|[0-9a-f]{6})$/i;
var BRAND = { data: null, loaded: false, promise: null, applied: '', titleName: '' };
var recolorObserver = null;
var nativeTitle = Object.getOwnPropertyDescriptor(Document.prototype, 'title');

function brandHex(v) {
  var m = BRAND_HEX_RE.exec(String(v || '').trim());
  if (!m) return '';
  var x = m[1].toLowerCase();
  if (x.length === 3) x = x[0] + x[0] + x[1] + x[1] + x[2] + x[2];
  return '#' + x;
}

function hexRgb(hex) {
  var x = brandHex(hex).slice(1);
  return [parseInt(x.slice(0, 2), 16), parseInt(x.slice(2, 4), 16), parseInt(x.slice(4, 6), 16)];
}

function rgbHex(rgb) {
  return '#' + rgb.map(function (c) { return ('0' + Math.round(c).toString(16)).slice(-2); }).join('');
}

function brandPalette(hex) {
  var base = hexRgb(hex);
  var out = {};
  Object.keys(BRAND_SHADES).forEach(function (k) {
    var a = BRAND_SHADES[k];
    out[k] = base.map(function (c) { return Math.round(a >= 0 ? c + (255 - c) * a : c * (1 + a)); });
  });
  return out;
}

function luminance(rgb) {
  var v = rgb.map(function (c) {
    c = c / 255;
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
}

function contrastRatio(a, b) {
  var la = luminance(a);
  var lb = luminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

function brandContrast(hex) {
  var p = brandPalette(hex);
  return {
    light: contrastRatio(p[500], [255, 255, 255]),
    dark: contrastRatio(p[400], [17, 24, 39])
  };
}

function contrastLevel(ratio) {
  if (ratio >= 4.5) return 'good';
  if (ratio >= 3) return 'large';
  return 'poor';
}

function brandText(map) {
  if (!map || typeof map !== 'object') return '';
  var l = lang();
  if (map[l]) return map[l];
  if (map[LANGS.def]) return map[LANGS.def];
  var keys = Object.keys(map).sort();
  for (var i = 0; i < keys.length; i++) if (map[keys[i]]) return map[keys[i]];
  return '';
}

function brandName() {
  return BRAND.data ? brandText(BRAND.data.name) : '';
}

function brandCustom(b) {
  b = b || BRAND.data;
  return !!(b && (b.logoUrl || b.logoDarkUrl || b.markUrl || b.accent || b.shortName || brandText(b.name) || brandText(b.welcome)));
}

function brandSet(data) {
  BRAND.data = data && typeof data === 'object' ? data : null;
  BRAND.loaded = true;
  try {
    if (brandCustom(BRAND.data)) localStorage.setItem(BRAND_KEY, JSON.stringify(BRAND.data));
    else localStorage.removeItem(BRAND_KEY);
  } catch (e) {}
  BRAND.applied = '';
  applyBrand();
}

function brandLoad(force) {
  if (BRAND.promise && !force) return BRAND.promise;
  BRAND.promise = fetch(BASE_BASE + '/brand/public', { cache: 'no-cache', credentials: 'omit', headers: { Accept: 'application/json' } }).then(function (res) {
    if (!res.ok) throw new Error('brand ' + res.status);
    return res.json();
  }).then(function (data) {
    brandSet(data);
    return BRAND.data;
  }, function () {
    BRAND.promise = null;
    return BRAND.data;
  });
  return BRAND.promise;
}

var SP_INK_LIGHT_BG = [223, 228, 238];
var SP_INK_DARK_BG = [74, 77, 95];

function pickShade(p, shades, bg) {
  for (var i = 0; i < shades.length; i++) if (contrastRatio(p[shades[i]], bg) >= 4.5) return shades[i];
  return shades[shades.length - 1];
}

function spatialAccent(hex) {
  var p = brandPalette(hex);
  var a = pickShade(p, [500, 600, 700, 800, 900], [255, 255, 255]);
  var ink = pickShade(p, [600, 700, 800, 900, 950], SP_INK_LIGHT_BG);
  var inkDark = pickShade(p, [300, 200, 100, 50], SP_INK_DARK_BG);
  return {
    accent: p[a], ink: p[ink], inkDark: p[inkDark],
    shades: { accent: +a, ink: +ink, inkDark: +inkDark },
    ratios: { onAccent: contrastRatio(p[a], [255, 255, 255]), ink: contrastRatio(p[ink], SP_INK_LIGHT_BG), inkDark: contrastRatio(p[inkDark], SP_INK_DARK_BG) }
  };
}

function brandCss(hex) {
  var p = brandPalette(hex);
  var sp = spatialAccent(hex);
  return 'html:root{' + Object.keys(p).map(function (k) { return '--primary-color-' + k + ':' + p[k].join(' ') + ';'; }).join('') +
    '--sp-accent:rgb(' + sp.accent.join(' ') + ');--sp-accent-ink:rgb(' + sp.ink.join(' ') + ');--sp-accent-ink-dark:rgb(' + sp.inkDark.join(' ') + ');}';
}

function recolorEl(el) {
  if (!el || el.nodeType !== 1) return;
  var s = el.getAttribute('style');
  if (!s || s.indexOf('rgb(') === -1) return;
  var out = s.replace(BRAND_RGB_RE, function (m, r, g, b) {
    var n = BRAND_DEFAULT_SHADES[+r + ',' + +g + ',' + +b];
    return n ? 'rgb(var(--primary-color-' + n + '))' : m;
  });
  if (out !== s) el.setAttribute('style', out);
}

function recolorTree(node) {
  if (!node || node.nodeType !== 1) return;
  recolorEl(node);
  var list = node.querySelectorAll('[style*="rgb("]');
  for (var i = 0; i < list.length; i++) recolorEl(list[i]);
}

function startRecolor() {
  if (recolorObserver) return;
  recolorTree(document.documentElement);
  recolorObserver = new MutationObserver(function (records) {
    for (var i = 0; i < records.length; i++) {
      var r = records[i];
      if (r.type === 'attributes') recolorEl(r.target);
      else for (var j = 0; j < r.addedNodes.length; j++) recolorTree(r.addedNodes[j]);
    }
  });
  recolorObserver.observe(document.documentElement, { subtree: true, childList: true, attributes: true, attributeFilter: ['style'] });
}

function applyAccent(hex) {
  var style = document.getElementById('brand-accent');
  hex = brandHex(hex);
  if (!hex) {
    if (style) style.remove();
    return;
  }
  var css = brandCss(hex);
  if (!style) {
    style = document.createElement('style');
    style.id = 'brand-accent';
  }
  if (style.textContent !== css) style.textContent = css;
  if (style.parentNode !== document.head) document.head.appendChild(style);
  startRecolor();
}

function applyFavicon(url) {
  var links = document.querySelectorAll('link[rel~="icon"]');
  if (!links.length && url) {
    var link = document.createElement('link');
    link.rel = 'icon';
    document.head.appendChild(link);
    links = [link];
  }
  for (var i = 0; i < links.length; i++) {
    var l = links[i];
    if (!l.hasAttribute('data-brand-default')) l.setAttribute('data-brand-default', l.getAttribute('href') || '/favicon.ico');
    var next = url || l.getAttribute('data-brand-default');
    if (l.getAttribute('href') !== next) l.setAttribute('href', next);
  }
}

function brandTitle(v) {
  v = String(v == null ? '' : v);
  if (BRAND.titleName) v = v.split(BRAND.titleName).join('Nakama');
  var name = brandName();
  BRAND.titleName = name;
  if (!name) return v;
  var parts = v.split(' | ');
  var last = parts.length - 1;
  parts[last] = parts[last].replace(/\bNakama\b/, name);
  return parts.join(' | ');
}

function hookTitle() {
  if (!nativeTitle || !nativeTitle.set) return;
  Object.defineProperty(document, 'title', {
    configurable: true,
    get: function () { return nativeTitle.get.call(document); },
    set: function (v) { nativeTitle.set.call(document, brandTitle(v)); }
  });
}

function refreshTitle() {
  var cur = nativeTitle.get.call(document);
  var next = brandTitle(cur);
  if (next !== cur) nativeTitle.set.call(document, next);
}

function brandLogoEl(b, sizeCls, compact) {
  var wrap = h('span', { class: 'brand-logo ' + sizeCls, 'data-brand-logo': '', 'data-l10n-skip': '' });
  var name = brandText(b.name);
  var alt = name || b.shortName || '';
  if (b.logoUrl && !compact) {
    wrap.appendChild(h('img', { src: b.logoUrl, alt: alt, class: 'brand-logo-img' + (b.logoDarkUrl ? ' brand-only-light' : '') }));
    if (b.logoDarkUrl) wrap.appendChild(h('img', { src: b.logoDarkUrl, alt: alt, class: 'brand-logo-img brand-only-dark' }));
    return wrap;
  }
  var mark = b.markUrl;
  if (mark) wrap.appendChild(h('img', { src: mark, alt: compact ? alt : '', class: 'brand-logo-mark' }));
  var label = compact ? (b.shortName || name).slice(0, 2) : (name || b.shortName);
  if (!mark || !compact) wrap.appendChild(h('span', { class: 'brand-logo-name', text: label || 'Nakama' }));
  return wrap;
}

function placeLogo(anchor, b, compact) {
  var natives = [];
  for (var i = 0; i < anchor.children.length; i++) if (!anchor.children[i].hasAttribute('data-brand-logo')) natives.push(anchor.children[i]);
  var ours = anchor.querySelector(':scope > [data-brand-logo]');
  var show = b && (b.logoUrl || b.markUrl || brandText(b.name) || b.shortName);
  if (!show) {
    if (ours) ours.remove();
    natives.forEach(function (n) { if (n.hasAttribute('data-brand-hidden')) { n.style.display = ''; n.removeAttribute('data-brand-hidden'); } });
    return;
  }
  var img = anchor.querySelector('img:not([data-brand-logo] img)');
  var size = img && /\bh-(\d+)\b/.exec(img.className) ? 'brand-h-' + /\bh-(\d+)\b/.exec(img.className)[1] : 'brand-h-7';
  var key = lang() + '|' + compact + '|' + size + '|' + JSON.stringify([b.name, b.shortName, b.logoUrl, b.logoDarkUrl, b.markUrl]);
  if (!ours || ours.getAttribute('data-key') !== key) {
    var built = brandLogoEl(b, size, compact);
    built.setAttribute('data-key', key);
    if (ours) ours.replaceWith(built);
    else anchor.insertBefore(built, anchor.firstChild);
  }
  natives.forEach(function (n) {
    if (n.style.display !== 'none') n.style.display = 'none';
    n.setAttribute('data-brand-hidden', '');
  });
}

function applyLogos(b) {
  var drawers = document.querySelectorAll('.x-drawer a[href="#/"]');
  for (var i = 0; i < drawers.length; i++) {
    var a = drawers[i];
    if (!a.querySelector('img') || a.closest('nav')) continue;
    var drawer = a.closest('.x-drawer');
    var compact = !!drawer && drawer.getBoundingClientRect().width > 0 && drawer.getBoundingClientRect().width < 100;
    placeLogo(a, b, compact);
  }
  var login = document.querySelectorAll('a[href="/"]');
  for (var j = 0; j < login.length; j++) if (login[j].querySelector('img')) placeLogo(login[j], b, false);
}

function applyLogin(b) {
  var input = document.querySelector('form input[name="username"]');
  var heads = document.querySelectorAll('h1[data-brand-hidden]');
  var native = null;
  if (input) {
    var box = input.closest('form');
    for (var k = 0; k < 4 && box && !native; k++) {
      box = box.parentElement;
      native = box && box.querySelector('h1:not([data-brand-signin])');
    }
  }
  var name = b ? brandText(b.name) || b.shortName : '';
  var welcome = b ? brandText(b.welcome) : '';
  var ours = document.querySelector('[data-brand-signin]');
  if (!native || (!name && !welcome)) {
    if (ours) ours.remove();
    for (var i = 0; i < heads.length; i++) { heads[i].style.display = ''; heads[i].removeAttribute('data-brand-hidden'); }
    return;
  }
  var key = lang() + '|' + name + '|' + welcome;
  if (!ours || ours.getAttribute('data-key') !== key || ours.previousElementSibling !== native) {
    var built = h('div', { 'data-brand-signin': '', 'data-l10n-skip': '', 'data-key': key, class: 'brand-signin', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('h1', { class: native.className, text: name ? t('brand.signIn', { name: name }) : native.textContent }),
      welcome ? h('p', { class: 'brand-welcome', text: welcome }) : null);
    if (ours) ours.remove();
    native.after(built);
  }
  if (native.style.display !== 'none') native.style.display = 'none';
  native.setAttribute('data-brand-hidden', '');
}

function applyBrand() {
  var b = BRAND.data;
  var key = lang() + '|' + JSON.stringify(b);
  if (BRAND.applied !== key) {
    BRAND.applied = key;
    applyAccent(b && b.accent);
    applyFavicon(b && b.markUrl ? b.markUrl : '');
    refreshTitle();
  }
  applyLogos(b);
  applyLogin(b);
}

(function () {
  try {
    var cached = localStorage.getItem(BRAND_KEY);
    if (cached) BRAND.data = JSON.parse(cached);
  } catch (e) { BRAND.data = null; }
  hookTitle();
  if (BRAND.data) {
    applyAccent(BRAND.data.accent);
    applyFavicon(BRAND.data.markUrl || '');
    refreshTitle();
  }
  brandLoad();
})();

var BRAND_SWATCHES = ['#6b59ee', '#2563eb', '#0891b2', '#059669', '#d97706', '#dc2626', '#db2777', '#475569'];

function contrastLine(labelKey, ratio) {
  var level = contrastLevel(ratio);
  var color = level === 'good' ? 'success' : level === 'large' ? 'warning' : 'error';
  return h('div', { class: 'flex items-center justify-between gap-3 text-sm', 'data-contrast': labelKey },
    h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t(labelKey) }),
    h('span', { class: 'flex items-center gap-2' },
      h('span', { class: 'tabular-nums', dir: 'ltr', text: fmtNum(ratio, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + ':' + fmtNum(1) }),
      tag(t('brand.contrast.' + level), color)));
}

function brandPreview(state) {
  var box = h('div', { class: 'brand-preview', 'aria-hidden': 'true' });
  function paint() {
    clear(box);
    var hex = brandHex(state.accent) || BRAND_DEFAULT_ACCENT;
    var p = brandPalette(hex);
    ['light', 'dark'].forEach(function (mode) {
      var dark = mode === 'dark';
      var logoUrl = dark ? state.logoDarkUrl || state.logoUrl : state.logoUrl;
      var name = brandText(state.name) || state.shortName || 'Nakama';
      var head = h('div', { class: 'brand-preview-head' });
      if (logoUrl) head.appendChild(h('img', { src: logoUrl, alt: '', class: 'brand-preview-logo' }));
      else {
        if (state.markUrl) head.appendChild(h('img', { src: state.markUrl, alt: '', class: 'brand-preview-mark' }));
        head.appendChild(h('span', { class: 'brand-preview-name', text: name }));
      }
      var accent = 'rgb(' + p[dark ? 600 : 500].join(',') + ')';
      var link = 'rgb(' + p[dark ? 300 : 700].join(',') + ')';
      box.appendChild(h('div', { class: 'brand-preview-pane brand-preview-' + mode },
        h('div', { class: 'brand-preview-mode', text: t(dark ? 'brand.preview.dark' : 'brand.preview.light') }),
        head,
        h('div', { class: 'brand-preview-nav', style: 'border-inline-start-color:' + link + ';color:' + link + ';background-color:' + (dark ? 'rgb(30,41,59)' : 'rgb(248,250,252)') }, t('brand.preview.nav')),
        h('div', { class: 'brand-preview-row' },
          h('span', { class: 'brand-preview-btn', style: 'background-color:' + accent, text: t('save') }),
          h('span', { class: 'brand-preview-link', style: 'color:' + link, text: t('brand.preview.link') }))));
    });
  }
  paint();
  box.paint = paint;
  return box;
}

function brandImage(value, extra, onchange) {
  var el = imageInput(value, onchange);
  el.classList.add('brand-img');
  if (extra) el.classList.add(extra);
  return el;
}

function brandEditor(data) {
  data = data || {};
  var state = {
    name: trValue(data.name), shortName: data.shortName || '', logoUrl: data.logoUrl || '', logoDarkUrl: data.logoDarkUrl || '',
    markUrl: data.markUrl || '', accent: brandHex(data.accent), welcome: trValue(data.welcome)
  };
  var initial = null;
  var preview = brandPreview(state);
  var onchange = function () { preview.paint(); paintContrast(); };

  var nameCtl = translatedInput(state.name, { name: 'name', label: t('brand.f.name'), max: 64, oninput: function (v) { state.name = v; nameField.setError(''); nameCtl.markLang(null); onchange(); } });
  var nameField = field(t('brand.f.name'), nameCtl, t('brand.f.name.hint'), { optional: true });
  var shortCtl = textInput(state.shortName, { label: t('brand.f.shortName'), oninput: function (v) { state.shortName = v; shortField.setError(''); onchange(); } });
  var shortField = field(t('brand.f.shortName'), shortCtl, t('brand.f.shortName.hint'), { optional: true });
  var logoField = field(t('brand.f.logo'), brandImage(state.logoUrl, '', function (v) { state.logoUrl = v || ''; logoField.setError(''); onchange(); }), t('brand.f.logo.hint'), { optional: true });
  var darkField = field(t('brand.f.logoDark'), brandImage(state.logoDarkUrl, 'brand-img--dark', function (v) { state.logoDarkUrl = v || ''; darkField.setError(''); onchange(); }), t('brand.f.logoDark.hint'), { optional: true });
  var markField = field(t('brand.f.mark'), brandImage(state.markUrl, '', function (v) { state.markUrl = v || ''; markField.setError(''); onchange(); }), t('brand.f.mark.hint'), { optional: true });

  var picker = h('input', { type: 'color', class: 'brand-color-input', 'aria-label': t('brand.f.accent.pick'), value: state.accent || BRAND_DEFAULT_ACCENT });
  var hexInput = textInput(state.accent, { ltr: true, mono: true, label: t('brand.f.accent'), placeholder: BRAND_DEFAULT_ACCENT });
  hexInput.classList.add('text-sm');
  hexInput.setAttribute('maxlength', '7');
  var swatches = h('div', { class: 'brand-swatches', role: 'group', 'aria-label': t('brand.f.accent.presets') });
  BRAND_SWATCHES.forEach(function (c) {
    swatches.appendChild(h('button', { type: 'button', class: 'brand-swatch', style: 'background-color:' + c, title: c, 'aria-label': c, 'data-color': c, onclick: function () { setAccent(c); } }));
  });
  var resetBtn = button({ variant: 'link', size: 'sm', label: t('brand.f.accent.reset'), onclick: function () { setAccent(''); } });
  var contrastBox = h('div', { class: 'brand-contrast', 'aria-live': 'polite' });
  function setAccent(v, fromText) {
    state.accent = v ? brandHex(v) || v : '';
    if (!fromText) hexInput.value = state.accent;
    if (brandHex(state.accent)) picker.value = brandHex(state.accent);
    accentField.setError('');
    Array.prototype.forEach.call(swatches.children, function (s) { s.setAttribute('aria-pressed', s.getAttribute('data-color') === brandHex(state.accent) ? 'true' : 'false'); });
    onchange();
  }
  picker.addEventListener('input', function () { setAccent(picker.value); });
  hexInput.addEventListener('input', function () {
    var v = hexInput.value.trim();
    if (v && v[0] !== '#') v = '#' + v;
    setAccent(v, true);
  });
  function paintContrast() {
    clear(contrastBox);
    var hex = brandHex(state.accent);
    if (state.accent && !hex) return;
    var c = brandContrast(hex || BRAND_DEFAULT_ACCENT);
    contrastBox.appendChild(contrastLine('brand.contrast.onWhite', c.light));
    contrastBox.appendChild(contrastLine('brand.contrast.onDark', c.dark));
    if (contrastLevel(c.light) === 'poor' || contrastLevel(c.dark) === 'poor') contrastBox.appendChild(h('p', { class: 'text-xs eco-text-warning mt-1', text: t('brand.contrast.tip') }));
  }
  var accentField = field(t('brand.f.accent'), h('div', { class: 'flex flex-col gap-2' },
    h('div', { class: 'flex items-center gap-2' }, picker, h('div', { class: 'brand-hex' }, hexInput), resetBtn),
    swatches), t('brand.f.accent.hint'), { optional: true });

  var welcomeCtl = translatedInput(state.welcome, { name: 'welcome', label: t('brand.f.welcome'), multiline: true, rows: 2, max: 280, counter: true, oninput: function (v) { state.welcome = v; welcomeField.setError(''); welcomeCtl.markLang(null); } });
  var welcomeField = field(t('brand.f.welcome'), welcomeCtl, t('brand.f.welcome.hint'), { optional: true });

  var fieldsEl = { name: nameField, shortName: shortField, logoUrl: logoField, logoDarkUrl: darkField, markUrl: markField, accent: accentField, welcome: welcomeField };
  setAccent(state.accent);
  var el = h('div', { class: 'brand-editor' },
    h('div', { class: 'brand-editor-main' },
      nameField, shortField,
      h('div', { class: 'eco-form-row' }, logoField, darkField),
      markField, accentField, contrastBox, welcomeField),
    h('aside', { class: 'brand-editor-aside' },
      h('div', { class: 'text-sm font-medium text-secondary-700 dark:text-secondary-200 mb-2', text: t('brand.preview') }),
      preview));

  function values() {
    return { name: trClean(state.name), shortName: state.shortName.trim(), logoUrl: (state.logoUrl || '').trim(), logoDarkUrl: (state.logoDarkUrl || '').trim(), markUrl: (state.markUrl || '').trim(), accent: brandHex(state.accent), welcome: trClean(state.welcome) };
  }
  initial = JSON.stringify(values());
  function validate() {
    var ok = true;
    var hit = trCheck(state.name, { max: 64 });
    if (hit) { nameCtl.markLang(hit.lang); nameField.setError(t('code.' + hit.code, { lang: langName(hit.lang) })); ok = false; }
    var wh = trCheck(state.welcome, { max: 280 });
    if (wh) { welcomeCtl.markLang(wh.lang); welcomeField.setError(t('code.' + wh.code, { lang: langName(wh.lang) })); ok = false; }
    if (Array.from(state.shortName.trim()).length > 24) { shortField.setError(t('brand.err.short')); ok = false; }
    if (state.accent && !brandHex(state.accent)) { accentField.setError(t('brand.err.color')); ok = false; }
    if (!ok) { var bad = el.querySelector('.eco-invalid'); if (bad) bad.scrollIntoView({ block: 'center', behavior: 'smooth' }); }
    return ok;
  }
  function showError(err) {
    var f = err && err.field && fieldsEl[err.field];
    if (!f) return false;
    if (err.field === 'name') nameCtl.markLang(trServerLang(err, state.name, { max: 64 }));
    if (err.field === 'welcome') welcomeCtl.markLang(trServerLang(err, state.welcome, { max: 280 }));
    f.setError(err.code === 'INVALID_COLOR' ? t('brand.err.color') : errorText(err));
    f.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return true;
  }
  return {
    el: el, values: values, validate: validate, showError: showError,
    dirty: function () { return JSON.stringify(values()) !== initial; },
    clean: function () { initial = JSON.stringify(values()); }
  };
}

function brandFormSkeleton() {
  return h('div', { class: 'brand-editor' },
    h('div', { class: 'brand-editor-main' }, skeletonField('h-16'), skeletonField(),
      h('div', { class: 'eco-form-row' }, skeletonField('h-16'), skeletonField('h-16')), skeletonField('h-16'), skeletonField(), skeletonField('h-16')),
    h('aside', { class: 'brand-editor-aside' }, bar('5rem', 'h-3 mb-2'), bar('100%', 'h-36 mb-3'), bar('100%', 'h-36')));
}

function saveBrand(editor) {
  if (!editor.validate()) return Promise.resolve(false);
  return baseApi('PUT', '/brand', editor.values()).then(function (res) {
    brandSet(res);
    editor.clean();
    return true;
  }, function (err) {
    if (!editor.showError(err)) toast('error', errorText(err));
    return false;
  });
}

function brandPage(container) {
  var frame = tabFrame(t('desc.brand'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'p-4 md:p-6' }, brandFormSkeleton());
  frame.content.appendChild(card(body));
  function load() {
    Promise.all([baseApi('GET', '/brand'), freshLanguages()]).then(function (all) {
      if (!container.isConnected) return;
      var editor = brandEditor(all[0]);
      var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () {
        setBusy(saveBtn, true, t('saving'));
        saveBrand(editor).then(function (ok) { setBusy(saveBtn, false); if (ok) toast('success', t('saved')); });
      } });
      clear(body).appendChild(editor.el);
      body.appendChild(h('div', { class: 'flex justify-end pt-2' }, saveBtn));
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}
