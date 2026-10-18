/*
 * Nakama Console localization runtime.
 *
 * The console UI is shipped as a prebuilt bundle (console/ui/dist) whose
 * templates contain hardcoded English text. This runtime localizes it in place
 * without modifying the bundle:
 *
 *   - selects the active locale (default fa-IR) and sets <html lang dir>;
 *   - redirects default-locale Intl formatting to the active locale, so dates
 *     use the Solar Hijri (Jalali) calendar and Persian digits;
 *   - translates text nodes, user-facing attributes and the document title
 *     through the catalogs in console/ui/l10n/locales, using exact matches,
 *     placeholder patterns, plural forms and CSS-selector contexts;
 *   - mirrors directional icons and keeps technical content left-to-right.
 *
 * The build step (scripts/build.mjs) prepends `var __NAKAMA_L10N_BUILD__ = {...}`
 * and ships this file to dist/static. In Node (tests) the module exports the
 * pure functions instead of starting.
 */
(function (root, factory) {
  'use strict';
  var api = factory();
  if (typeof module === 'object' && module.exports) {
    module.exports = api;
  } else if (typeof __NAKAMA_L10N_BUILD__ !== 'undefined') {
    root.NakamaL10n = api.start(__NAKAMA_L10N_BUILD__, root);
  }
})(typeof self !== 'undefined' ? self : this, function () {
  'use strict';

  var STORAGE_KEY = 'nakama.console.locale';
  var DEBUG_KEY = 'nakama.console.l10n.debug';
  var FSI = '\u2068';
  var LRI = '\u2066';
  var RLI = '\u2067';
  var PDI = '\u2069';
  var RTL_CHARS = /[\u0590-\u08FF\uFB1D-\uFDFF\uFE70-\uFEFF]/;
  var SVG_NS = 'http://www.w3.org/2000/svg';
  var PERSIAN_DIGITS = '۰۱۲۳۴۵۶۷۸۹';
  var PLACEHOLDER = /\{(\w+)\}/g;
  var TRANSLATED_ATTRIBUTES = ['placeholder', 'title', 'aria-label', 'alt', 'aria-placeholder', 'aria-description'];
  // Never translate or reshape text inside these: code, editors, user input.
  var SKIP_SELECTOR = [
    'script', 'style', 'noscript', 'code', 'pre', 'kbd', 'samp', 'textarea', 'svg title',
    '[contenteditable=""]', '[contenteditable="true"]', '.monaco-editor', '.monaco-diff-editor',
    '.ProseMirror', '.cm-editor', '[translate="no"]', '.notranslate', '[data-l10n-skip]'
  ].join(',');
  // Inside these, text is translated but numbers keep their Latin digits.
  var LATIN_DIGITS_SELECTOR = '.font-mono, [data-l10n-latn], [dir="ltr"], a[href^="http"]';
  var NUMERIC = /^[+\-\u2212]?(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?\s?(?:%|[kKMB])?$/;
  // Placeholder names imply what they may capture, so short patterns such as
  // "{from} to {to}" cannot swallow arbitrary English sentences.
  var PLACEHOLDER_TYPES = {
    numeric: { names: /^(count|n|from|to|min|max|amount|seconds|percent|row|step|factor)$/, re: '([+\\-\\u2212]?[\\d\\u06F0-\\u06F9][\\d\\u06F0-\\u06F9.,\\u066B\\u066C]*\\s?[%\\u066AkKMB]?)' },
    token: { names: /^(field|id|provider|version|word|path|status)$/, re: '(\\S+?)' }
  };
  // Nested placeholders must themselves be catalog phrases (or already localized).
  var NESTED_PLACEHOLDERS = /^(time)$/;

  function normalize(text) {
    return String(text).replace(/\s+/g, ' ').trim();
  }

  function escapeRegExp(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  // toPersianDigits converts ASCII digits and the separators between them.
  function toPersianDigits(s) {
    return String(s)
      .replace(/(\d),(?=\d{3}\b)/g, '$1٬')
      .replace(/(\d)\.(?=\d)/g, '$1٫')
      .replace(/%/g, '٪')
      .replace(/\d/g, function (d) { return PERSIAN_DIGITS[+d]; });
  }

  function isNumeric(s) {
    return NUMERIC.test(s) && s.replace(/\D/g, '').length <= 12;
  }

  function hasStrongLTR(s) {
    return /[A-Za-z\u00C0-\u024F]/.test(s);
  }

  // flatten turns a namespaced catalog into {"ns.key": value} pairs.
  function flatten(catalog) {
    var out = {};
    Object.keys(catalog).forEach(function (ns) {
      if (ns.charAt(0) === '$') return;
      var entries = catalog[ns];
      Object.keys(entries).forEach(function (key) {
        out[ns + '.' + key] = entries[key];
      });
    });
    return out;
  }

  // compile builds the lookup index used by translate() from a source (English)
  // catalog and a target catalog with the same keys.
  function compile(source, target, locale) {
    var src = flatten(source);
    var dst = flatten(target);
    var index = { locale: locale, exact: Object.create(null), patterns: [], contexts: [], missing: [], keys: 0 };
    Object.keys(src).forEach(function (id) {
      var entry = src[id];
      var message = typeof entry === 'string' ? entry : entry.message;
      var context = typeof entry === 'string' ? null : entry.context || null;
      var translation = dst[id];
      if (typeof translation !== 'string' || translation === '') {
        index.missing.push(id);
        return;
      }
      index.keys++;
      var forms = translation.split(' | ');
      var sourceForms = message.split(' | ');
      // In "1 minute | {count} minutes" the singular form spells the count out;
      // capture that number as the same placeholder.
      var plural = sourceForms.length > 1 ? /\{(\w+)\}/.exec(message) : null;
      sourceForms.forEach(function (form) {
        form = normalize(form);
        if (plural && form.indexOf('{') === -1) form = form.replace(/\b1\b/, '{' + plural[1] + '}');
        var names = [];
        var literal = form.replace(PLACEHOLDER, function (_, name) { names.push(name); return ''; });
        var rule = { id: id, forms: forms, names: names, context: context, literal: literal.length };
        if (names.length) {
          var parts = form.split(PLACEHOLDER);
          var source = '^';
          for (var i = 0; i < parts.length; i++) {
            source += i % 2 ? placeholderPattern(parts[i]) : escapeRegExp(parts[i]);
          }
          rule.re = new RegExp(source + '$');
          (context ? index.contexts : index.patterns).push(rule);
        } else if (context) {
          rule.text = form;
          index.contexts.push(rule);
        } else if (!(form in index.exact)) {
          index.exact[form] = rule;
        }
      });
    });
    // Most specific patterns first so "{a} to {b}" does not shadow longer ones.
    var bySpecificity = function (a, b) { return b.literal - a.literal || a.names.length - b.names.length; };
    index.patterns.sort(bySpecificity);
    index.contexts.sort(bySpecificity);
    return index;
  }

  function placeholderPattern(name) {
    if (PLACEHOLDER_TYPES.numeric.names.test(name)) return PLACEHOLDER_TYPES.numeric.re;
    if (PLACEHOLDER_TYPES.token.names.test(name)) return PLACEHOLDER_TYPES.token.re;
    return '(.+?)';
  }

  function pluralForm(rule, values, locale) {
    if (rule.forms.length === 1) return rule.forms[0];
    var n = NaN;
    for (var i = 0; i < rule.names.length && isNaN(n); i++) n = parseFloat(values[rule.names[i]]);
    var category = isNaN(n) ? 'other' : new Intl.PluralRules(locale).select(n);
    return category === 'one' ? rule.forms[0] : rule.forms[rule.forms.length - 1];
  }

  function render(index, rule, match, depth, options, element) {
    var values = {};
    for (var i = 0; i < rule.names.length; i++) {
      var raw = match ? match[i + 1] : '';
      var inner = depth < 2 ? lookup(index, raw, element, depth + 1, options) : null;
      if (inner === null && NESTED_PLACEHOLDERS.test(rule.names[i]) && !RTL_CHARS.test(raw) && !isNumeric(raw)) return null;
      var value = inner !== null ? inner : raw;
      if (options.digits && isNumeric(value)) value = toPersianDigits(value);
      else if (inner === null && options.isolate && hasStrongLTR(value)) value = FSI + value + PDI;
      values[rule.names[i]] = value;
    }
    return pluralForm(rule, match ? values : {}, index.locale).replace(PLACEHOLDER, function (all, name) {
      return name in values ? values[name] : all;
    });
  }

  function lookup(index, text, element, depth, options) {
    var i, rule, m, out;
    if (element && index.contexts.length) {
      for (i = 0; i < index.contexts.length; i++) {
        rule = index.contexts[i];
        if (!element.closest || !element.closest(rule.context)) continue;
        if (rule.re) {
          m = rule.re.exec(text);
          if (m && (out = render(index, rule, m, depth, options, element)) !== null) return out;
        } else if (rule.text === text) {
          return render(index, rule, null, depth, options, element);
        }
      }
    }
    rule = index.exact[text];
    if (rule) return render(index, rule, null, depth, options, element);
    for (i = 0; i < index.patterns.length; i++) {
      rule = index.patterns[i];
      m = rule.re.exec(text);
      if (m && (out = render(index, rule, m, depth, options, element)) !== null) return out;
    }
    return null;
  }

  // translate returns the localized text for an English UI string, or null
  // when the catalog has no entry. Surrounding whitespace is preserved.
  function translate(index, text, element, options) {
    options = options || { digits: true, isolate: true };
    var key = normalize(text);
    if (!key) return null;
    var out = lookup(index, key, element || null, 0, options);
    if (out === null) {
      if (options.digits && isNumeric(key)) out = toPersianDigits(key);
      else return null;
    }
    if (options.isolate) out = isolateTechnical(out);
    var lead = /^\s*/.exec(text)[0];
    var trail = /\s*$/.exec(text)[0];
    // A fragment that starts with punctuation attaches to the preceding element.
    if (/^[.,:;!?\u060C\u061B\u061F]/.test(out)) lead = '';
    return lead + out + trail;
  }

  // isolateTechnical keeps digit groups joined by separators (UUIDs, ranges,
  // versions written in Latin digits) in order inside right-to-left sentences.
  function isolateTechnical(text) {
    if (!RTL_CHARS.test(text)) return text;
    return text.replace(/(^|[^\w\u2066\u2068.\-])(\d[\dA-Za-z]*(?:[-:/.][\dA-Za-z]+)+)/g, function (all, before, token) {
      return before + LRI + token + PDI;
    });
  }

  // translateTitle localizes "Page - Sub | Nakama (host)" document titles.
  function translateTitle(index, title) {
    var parts = String(title).split(' | ');
    var page = parts[0].split(' - ').map(function (segment) {
      var t = translate(index, segment, null, { digits: false, isolate: false });
      return t === null ? segment : t;
    }).join(' - ');
    parts[0] = page;
    return parts.join(' | ');
  }

  function resolveLocale(build, win) {
    var supported = build.locales || {};
    var pick = function (value) {
      if (!value) return null;
      if (supported[value]) return value;
      var lower = String(value).toLowerCase();
      var keys = Object.keys(supported);
      for (var i = 0; i < keys.length; i++) {
        if (keys[i].toLowerCase() === lower || keys[i].split('-')[0].toLowerCase() === lower) return keys[i];
      }
      return null;
    };
    var fromQuery = null;
    try {
      fromQuery = pick(new URLSearchParams(win.location.search).get('lang'));
      if (fromQuery) win.localStorage.setItem(STORAGE_KEY, fromQuery);
    } catch (e) { /* storage unavailable */ }
    if (fromQuery) return fromQuery;
    try {
      var stored = pick(win.localStorage.getItem(STORAGE_KEY));
      if (stored) return stored;
    } catch (e) { /* storage unavailable */ }
    return build.defaultLocale;
  }

  // patchIntl makes formatting that relies on the browser default locale use
  // the active locale. Calls with an explicit locale are left untouched because
  // parts of the console parse that output (date picker, number inputs).
  function patchIntl(win, locale) {
    var Intl = win.Intl;
    var isDefault = function (l) { return l === undefined || l === null || (Array.isArray(l) && l.length === 0); };
    var wrap = function (Native, choose, decorate) {
      var Wrapped = function (locales, options) {
        var instance = new Native(choose(locales, options), options);
        return decorate ? decorate(instance, locales, options) : instance;
      };
      Wrapped.prototype = Native.prototype;
      Wrapped.supportedLocalesOf = Native.supportedLocalesOf.bind(Native);
      Object.defineProperty(Wrapped, 'name', { value: Native.name });
      return Wrapped;
    };
    var NativeDTF = Intl.DateTimeFormat;
    var NativeNF = Intl.NumberFormat;
    var NativeRTF = Intl.RelativeTimeFormat;
    Intl.DateTimeFormat = wrap(NativeDTF, function (l) { return isDefault(l) ? locale : l; });
    Intl.NumberFormat = wrap(NativeNF, function (l, o) {
      // Currency amounts are display-only; localize their presentation but keep the currency.
      if (isDefault(l) || (o && o.style === 'currency')) return locale;
      return l;
    }, function (instance, l, o) {
      if (!isDefault(l)) return instance;
      // formatToParts consumers (number-flow) parse digits with parseInt, so keep
      // Latin digits there; the Persian glyphs come from the FD font in CSS.
      var latin = new NativeNF(locale + '-u-nu-latn', o);
      instance.formatToParts = function (n) { return latin.formatToParts(n); };
      return instance;
    });
    if (NativeRTF) {
      var probe = new NativeRTF(locale, { numeric: 'always' });
      var past = probe.format(-2, 'day');
      var future = probe.format(2, 'day');
      var common = 0;
      while (common < past.length && past[common] === future[common]) common++;
      // e.g. fa-IR "۲ روز پیش" vs "۲ روز بعد" -> past marker "پیش" at the end.
      var pastSuffix = common > 0 && common < past.length ? new RegExp('\\s*' + escapeRegExp(past.slice(common)) + '$') : null;
      Intl.RelativeTimeFormat = wrap(NativeRTF, function (l) {
        return isDefault(l) || l === 'en' ? locale : l;
      }, function (instance, l) {
        if (!(isDefault(l) || l === 'en') || !pastSuffix) return instance;
        // The console turns "5 minutes ago" into a duration with .replace(" ago", "").
        // Emit that English suffix for past values so the replacement keeps working;
        // the catalog pattern "{time} ago" renders whatever remains in Persian.
        var format = instance.format.bind(instance);
        instance.format = function (value, unit) {
          var out = format(value, unit);
          return value < 0 || Object.is(value, -0) ? out.replace(pastSuffix, ' ago') : out;
        };
        return instance;
      });
    }
    ['toLocaleString', 'toLocaleDateString', 'toLocaleTimeString'].forEach(function (method) {
      var native = win.Date.prototype[method];
      win.Date.prototype[method] = function (locales, options) {
        return native.call(this, isDefault(locales) ? locale : locales, options);
      };
    });
  }

  // localizeMailto translates the subject and body of a mailto: URL (the
  // console composes the password-reset email this way). No bidi controls are
  // added so mail clients still recognise the link in the body.
  function localizeMailto(index, url) {
    var q = url.indexOf('?');
    if (q === -1) return url;
    var params = url.slice(q + 1).split('&').map(function (pair) {
      var eq = pair.indexOf('=');
      var name = eq === -1 ? pair : pair.slice(0, eq);
      if (eq === -1 || !/^(subject|body)$/i.test(name)) return pair;
      var value;
      try { value = decodeURIComponent(pair.slice(eq + 1)); } catch (e) { return pair; }
      var out = translate(index, value, null, { digits: false, isolate: false });
      return out === null ? pair : name + '=' + encodeURIComponent(out.trim());
    });
    return url.slice(0, q + 1) + params.join('&');
  }

  function patchMailto(win, index) {
    var open = win.open;
    win.open = function (url) {
      var args = Array.prototype.slice.call(arguments);
      if (typeof url === 'string' && /^mailto:/i.test(url)) args[0] = localizeMailto(index, url);
      return open.apply(this, args);
    };
  }

  function start(build, win) {
    var doc = win.document;
    var locale = resolveLocale(build, win);
    var info = build.locales[locale] || { dir: 'ltr', lang: locale };
    var html = doc.documentElement;
    var index = null;
    var missing = Object.create(null);
    var api = {
      locale: locale,
      dir: info.dir,
      locales: build.locales,
      version: build.version,
      setLocale: function (next) {
        try { win.localStorage.setItem(STORAGE_KEY, next); } catch (e) { /* storage unavailable */ }
        var url = new URL(win.location.href);
        url.searchParams.delete('lang');
        win.location.replace(url.toString());
      },
      translate: function (text) { return index ? translate(index, text) : text; },
      formatNumber: function (n, options) { return new win.Intl.NumberFormat(info.lang, options).format(n); },
      formatDate: function (d, options) { return new win.Intl.DateTimeFormat(info.lang, options).format(d); },
      missing: function () { return Object.keys(missing); }
    };

    var applyRoot = function () {
      if (html.getAttribute('lang') !== info.lang) html.setAttribute('lang', info.lang);
      if (html.getAttribute('dir') !== info.dir) html.setAttribute('dir', info.dir);
    };
    applyRoot();
    html.classList.add('l10n-' + info.dir);
    // The console sets <html lang> itself on boot; keep ours.
    new win.MutationObserver(applyRoot).observe(html, { attributes: true, attributeFilter: ['lang', 'dir'] });

    var catalog = build.catalogs && build.catalogs[locale];
    if (!catalog) {
      win.addEventListener('DOMContentLoaded', function () {
        mountSwitcher(win, api);
        new win.MutationObserver(function () { api.dock(); }).observe(doc.body, { childList: true, subtree: true });
      });
      return api;
    }

    patchIntl(win, info.lang);
    index = compile(build.source, catalog, info.lang);
    patchMailto(win, index);
    var debug = false;
    try { debug = win.localStorage.getItem(DEBUG_KEY) === '1'; } catch (e) { /* storage unavailable */ }
    var textDone = new WeakMap();
    var attrDone = new WeakMap();
    var mirror = Object.create(null);
    (build.mirrorIcons || []).forEach(function (svg) { mirror[normalizeSvg(doc, svg)] = true; });

    var skip = function (el) { return !el || (el.closest && el.closest(SKIP_SELECTOR)); };
    var options = function (el) {
      return { digits: !(el && el.closest && el.closest(LATIN_DIGITS_SELECTOR)), isolate: true };
    };
    var note = function (text) {
      if (debug && /[A-Za-z]{2}/.test(text)) missing[normalize(text)] = true;
    };

    var translateText = function (node) {
      var value = node.nodeValue;
      if (!value || textDone.get(node) === value || !/\S/.test(value)) return;
      var el = node.parentElement;
      if (skip(el)) return;
      var out = translate(index, value, el, options(el));
      if (out === null) {
        note(value);
        textDone.set(node, value);
        if (info.dir === 'rtl' && !markLatinProse(el, value)) isolateLatin(node, value);
        return;
      }
      // Chart labels are SVG text in a left-to-right box (ECharts positions them
      // with text-anchor), so embed Persian there as an explicit RTL run.
      if (info.dir === 'rtl' && el.namespaceURI === SVG_NS && RTL_CHARS.test(out)) out = RLI + out.trim() + PDI;
      textDone.set(node, out);
      if (out !== value) node.nodeValue = out;
    };

    // Untranslated English prose (external feeds, server or user text) reads
    // correctly only as a left-to-right paragraph.
    var markLatinProse = function (el, value) {
      if (!el || el.namespaceURI === SVG_NS) return false;
      if (el.hasAttribute('data-l10n-ltr')) return true;
      var words = normalize(value).split(' ').filter(function (w) { return /[A-Za-z]{2}/.test(w); });
      if (words.length < 3) return false;
      if (RTL_CHARS.test(el.textContent) || el.closest('td, th')) return false;
      el.setAttribute('data-l10n-ltr', '');
      return true;
    };

    // Other multi-token Latin text (cron expressions, values with spaces) keeps
    // the cell's alignment but must not be reordered: embed it as an LTR run.
    // Single tokens (IDs, usernames) are left untouched so copying them is exact.
    var isolateLatin = function (node, value) {
      var key = normalize(value);
      if (!/\s/.test(key) || RTL_CHARS.test(key) || !/[A-Za-z0-9]/.test(key)) return;
      if (node.parentElement && node.parentElement.namespaceURI === SVG_NS) return;
      var out = value.replace(key, LRI + key + PDI);
      textDone.set(node, out);
      node.nodeValue = out;
    };

    var translateAttributes = function (el) {
      if (skip(el)) return;
      var done = attrDone.get(el);
      for (var i = 0; i < TRANSLATED_ATTRIBUTES.length; i++) {
        var name = TRANSLATED_ATTRIBUTES[i];
        var value = el.getAttribute(name);
        if (!value || (done && done[name] === value)) continue;
        var out = translate(index, value, el, { digits: true, isolate: false });
        if (out === null) note(value);
        if (!done) attrDone.set(el, (done = {}));
        done[name] = out === null ? value : out;
        if (out !== null && out !== value) el.setAttribute(name, out);
      }
      if (el.tagName === 'svg' && info.dir === 'rtl' && !el.hasAttribute('data-l10n-icon')) {
        el.setAttribute('data-l10n-icon', mirror[normalizeSvg(doc, el.innerHTML)] ? 'mirror' : '');
      }
    };

    var walk = function (rootNode) {
      if (rootNode.nodeType === 3) return translateText(rootNode);
      if (rootNode.nodeType !== 1 && rootNode.nodeType !== 11) return;
      if (rootNode.nodeType === 1) {
        if (skip(rootNode)) return;
        translateAttributes(rootNode);
      }
      var walker = doc.createTreeWalker(rootNode, 5 /* SHOW_ELEMENT | SHOW_TEXT */, {
        acceptNode: function (n) {
          if (n.nodeType === 1 && n.matches(SKIP_SELECTOR)) return 2; // FILTER_REJECT skips the subtree
          return 1;
        }
      });
      var n;
      while ((n = walker.nextNode())) {
        if (n.nodeType === 3) translateText(n);
        else translateAttributes(n);
      }
    };

    var titleDone = null;
    var translateDocTitle = function () {
      var current = doc.title;
      if (current === titleDone) return;
      titleDone = translateTitle(index, current);
      if (titleDone !== current) doc.title = titleDone;
    };

    var observer = new win.MutationObserver(function (records) {
      if (api.dock) api.dock();
      for (var i = 0; i < records.length; i++) {
        var r = records[i];
        if (r.type === 'characterData') {
          translateText(r.target);
        } else if (r.type === 'attributes') {
          translateAttributes(r.target);
        } else {
          for (var j = 0; j < r.addedNodes.length; j++) {
            var added = r.addedNodes[j];
            if (added.parentNode && (added.parentNode.nodeName === 'TITLE' || added.nodeName === 'TITLE')) translateDocTitle();
            else walk(added);
          }
          if (r.target.nodeName === 'TITLE') translateDocTitle();
        }
      }
    });
    observer.observe(html, {
      childList: true,
      subtree: true,
      characterData: true,
      attributes: true,
      attributeFilter: TRANSLATED_ATTRIBUTES.concat(['type'])
    });
    translateDocTitle();
    if (doc.body) walk(doc.body);
    win.addEventListener('DOMContentLoaded', function () {
      walk(doc.body);
      mountSwitcher(win, api);
    });
    return api;
  }

  function normalizeSvg(doc, markup) {
    var holder = doc.createElementNS(SVG_NS, 'svg');
    holder.innerHTML = markup;
    return holder.innerHTML.replace(/\s+/g, ' ').trim();
  }

  // mountSwitcher adds a small language toggle to the console header (or a
  // floating one on pages without a header, such as sign in).
  function mountSwitcher(win, api) {
    var doc = win.document;
    var codes = Object.keys(api.locales);
    if (codes.length < 2) return;
    var next = codes[(codes.indexOf(api.locale) + 1) % codes.length];
    var button = doc.createElement('button');
    button.type = 'button';
    button.className = 'l10n-switcher';
    button.setAttribute('data-l10n-skip', '');
    button.setAttribute('lang', api.locales[next].lang);
    button.setAttribute('title', api.locales[next].switchLabel || api.locales[next].name);
    button.setAttribute('aria-label', api.locales[next].switchLabel || api.locales[next].name);
    button.textContent = api.locales[next].short || next;
    button.addEventListener('click', function () { api.setLocale(next); });
    doc.body.appendChild(button);
    // Dock into the console top bar (next to the account menu) when it exists.
    var scheduled = false;
    api.dock = function () {
      if (scheduled) return;
      scheduled = true;
      win.requestAnimationFrame(function () {
        scheduled = false;
        var counter = doc.querySelector('number-flow-vue');
        var bar = counter && counter.closest('.justify-end');
        var menu = bar && bar.querySelector(':scope > .x-popover');
        if (menu) {
          if (button.parentElement !== bar) {
            bar.appendChild(button);
            button.classList.add('l10n-switcher--docked');
          }
        } else if (button.parentElement !== doc.body) {
          doc.body.appendChild(button);
          button.classList.remove('l10n-switcher--docked');
        }
      });
    };
    api.dock();
  }

  return {
    normalize: normalize,
    toPersianDigits: toPersianDigits,
    isNumeric: isNumeric,
    flatten: flatten,
    compile: compile,
    translate: translate,
    translateTitle: translateTitle,
    localizeMailto: localizeMailto,
    resolveLocale: resolveLocale,
    patchIntl: patchIntl,
    start: start
  };
});
