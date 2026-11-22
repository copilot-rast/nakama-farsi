var ND_EXAMPLES = {
  persian: { date: '1404/10/11', datetime: '1404/10/11 14:30', range: '1404/10/01 – 1404/10/30' },
  gregory: { date: '2026/01/01', datetime: '2026/01/01 14:30', range: '2026/01/01 – 2026/01/31' }
};

function ndSame(a, b) {
  if (a === b) return true;
  if (!a || !b) return !a && !b;
  if (typeof a === 'object' || typeof b === 'object') return !!a && !!b && a.from === b.from && a.to === b.to;
  return false;
}

function ndCleanValue(v, mode) {
  if (v === undefined || v === null || v === '') return null;
  if (mode === 'range') {
    if (typeof v !== 'object') return null;
    var from = dayToJdn(v.from) != null ? v.from : null;
    var to = dayToJdn(v.to) != null ? v.to : null;
    return from || to ? { from: from, to: to } : null;
  }
  if (mode === 'date') {
    var s = String(v);
    if (dayToJdn(s.slice(0, 10)) != null && /^\d{4}-\d{2}-\d{2}$/.test(s)) return s;
    var d = new Date(s);
    return isNaN(d.getTime()) ? null : d.toISOString().slice(0, 10);
  }
  if (mode === 'time') {
    var t = ndParseTime(ndLatin(String(v)));
    return t ? ndPad(t.h) + ':' + ndPad(t.mi) : null;
  }
  var ms = new Date(v).getTime();
  return isNaN(ms) ? null : new Date(ms).toISOString();
}

function ndBoundJdn(v, mode, zone) {
  if (!v) return null;
  if (mode === 'date' || mode === 'range') return dayToJdn(String(v).slice(0, 10));
  return isoToJdn(v, zone);
}

function ndInput(o) {
  o = o || {};
  var mode = ['date', 'datetime', 'time', 'range'].indexOf(o.mode) === -1 ? 'datetime' : o.mode;
  var lang = ndLang();
  var cal = o.calendar || activeCalendar(lang);
  var zone = o.zone || ndZone();
  var value = ndCleanValue(o.value, mode);
  var min = o.min || null;
  var max = o.max || null;
  var id = o.id || ('nd-in-' + (++ndSeq));
  var zoneId = id + '-zone';
  var wrap = ndEl('div', { class: 'nd-field' + (o.className ? ' ' + o.className : ''), 'data-nd': mode, 'data-calendar': cal, 'data-l10n-skip': '' });
  var box = ndEl('div', { class: 'nd-box' });
  var input = ndEl('input', { type: 'text', id: id, class: 'nd-input' + (o.inputClass ? ' ' + o.inputClass : ''), name: o.name || null, autocomplete: 'off', spellcheck: 'false', dir: 'ltr', inputmode: ndMobile() ? 'none' : mode === 'time' ? 'numeric' : 'text', 'aria-label': o.label || o.name || null, 'aria-haspopup': 'dialog', 'aria-expanded': 'false', 'aria-describedby': o.zoneLine === false ? null : zoneId, placeholder: '\u200e' + (o.placeholder || (mode === 'time' ? ndDigits('14:30', lang) : ndDigits(ND_EXAMPLES[cal][mode], lang))) });
  Object.keys(o.attrs || {}).forEach(function (k) { input.setAttribute(k, o.attrs[k]); });
  if (o.required) input.setAttribute('aria-required', 'true');
  var clearBtn = ndEl('button', { type: 'button', class: 'nd-clear', 'aria-label': ndt('clear'), title: ndt('clear'), html: ndIcon('x'), 'data-nd-clear': '', hidden: true });
  var openBtn = ndEl('button', { type: 'button', class: 'nd-open', 'aria-label': ndt('open'), title: ndt('open'), 'aria-haspopup': 'dialog', 'aria-expanded': 'false', html: ndIcon(mode === 'time' ? 'clock' : 'calendar'), 'data-nd-open': '' });
  var line = o.zoneLine === false ? null : ndEl('p', { class: 'nd-zone', id: zoneId, 'data-nd-zone': zone, dir: lang === 'fa' ? 'rtl' : 'ltr' });
  Object.keys(o.zoneAttrs || {}).forEach(function (k) { if (line) line.setAttribute(k, o.zoneAttrs[k]); });
  var err = ndEl('p', { class: 'nd-error', hidden: true, 'data-nd-error': '' });
  box.appendChild(input);
  box.appendChild(clearBtn);
  box.appendChild(openBtn);
  wrap.appendChild(box);
  if (line) wrap.appendChild(line);
  wrap.appendChild(err);
  var disabled = false;
  var picker = null;
  var quiet = false;

  function shown() {
    return ndNumeric(value, mode, { calendar: cal, zone: zone, lang: lang });
  }

  function paintLine() {
    if (!line) return;
    var label = o.zoneLabel || zone;
    if (mode === 'datetime' && value) line.textContent = ndt('zoneAt', { time: ndFormat(value, { calendar: cal, zone: zone, style: 'long' }), zone: label });
    else if (mode === 'date' && value) line.textContent = ndt('zoneAt', { time: ndFormat(value, { calendar: cal, style: 'long' }), zone: label });
    else line.textContent = ndt('zone', { zone: label });
  }

  function paint(keepText) {
    if (!keepText) input.value = shown();
    clearBtn.hidden = !value || disabled || !!o.required;
    paintLine();
    wrap.setAttribute('data-value', value == null ? '' : typeof value === 'object' ? (value.from || '') + '/' + (value.to || '') : value);
  }

  function fire() {
    quiet = true;
    try {
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    } finally { quiet = false; }
  }

  function set(v, opts) {
    v = ndCleanValue(v, mode);
    var changed = !ndSame(v, value);
    value = v;
    err.hidden = true;
    err.removeAttribute('role');
    input.removeAttribute('aria-invalid');
    paint(opts && opts.keepText);
    if (changed && !(opts && opts.silent)) {
      if (o.onchange) o.onchange(value);
      if (!(opts && opts.keepText)) fire();
    }
  }

  function invalid() {
    input.setAttribute('aria-invalid', 'true');
    err.textContent = ndt(mode === 'time' ? 'invalidTime' : 'invalid', { example: ndDigits(ND_EXAMPLES[cal][mode] || '14:30', lang) });
    err.hidden = false;
    err.setAttribute('role', 'alert');
  }

  function fromText(final) {
    var r = ndParse(input.value, mode, zone);
    if (!r) {
      if (final && o.passInvalid && o.onchange) { value = null; o.onchange(ndLatin(input.value).trim()); return; }
      if (final) invalid();
      return;
    }
    if (r.empty) { set(null, { keepText: !final }); return; }
    var v = r.value;
    if (mode === 'datetime' && v) {
      if (min && new Date(v) < new Date(min)) { if (final) invalid(); return; }
      if (max && new Date(v) > new Date(max)) { if (final) invalid(); return; }
    }
    set(v, { keepText: !final });
  }

  function open() {
    if (disabled || picker) return;
    if (ndParse(input.value, mode, zone) && !ndSame(ndParse(input.value, mode, zone).value, value)) fromText(true);
    picker = openPicker({
      mode: mode,
      calendar: cal,
      zone: zone,
      label: o.label || null,
      min: min,
      max: max,
      minJdn: ndBoundJdn(min, mode, zone),
      maxJdn: ndBoundJdn(max, mode, zone),
      required: !!o.required,
      box: box,
      mount: wrap,
      getValue: function () { return value; },
      commit: function (v) { set(v); },
      onClose: function (restore) {
        picker = null;
        input.setAttribute('aria-expanded', 'false');
        openBtn.setAttribute('aria-expanded', 'false');
        wrap.classList.remove('is-open');
        if (restore) (ndMobile() ? openBtn : input).focus({ preventScroll: true });
      }
    });
    input.setAttribute('aria-expanded', 'true');
    input.setAttribute('aria-controls', picker.pop.id);
    openBtn.setAttribute('aria-expanded', 'true');
    wrap.classList.add('is-open');
  }

  input.addEventListener('input', function () { if (!quiet) fromText(false); });
  input.addEventListener('change', function () { if (!quiet) fromText(true); });
  input.addEventListener('blur', function () { if (!picker && !input.getAttribute('aria-invalid')) paint(); });
  input.addEventListener('keydown', function (e) {
    if ((e.key === 'ArrowDown' && (e.altKey || !input.value)) || (e.key === ' ' && !input.value)) { e.preventDefault(); open(); }
    else if (e.key === 'Enter') { fromText(true); if (!input.getAttribute('aria-invalid')) paint(); }
  });
  input.addEventListener('click', function () { if (ndMobile()) open(); });
  input.addEventListener('focus', function () { if (ndMobile()) { input.blur(); open(); } });
  openBtn.addEventListener('click', function (e) { e.preventDefault(); if (picker) closePicker(true); else open(); });
  clearBtn.addEventListener('click', function (e) { e.preventDefault(); set(null); input.focus(); });

  Object.defineProperty(wrap, 'value', { configurable: true, get: function () { return value; }, set: function (v) { set(v, { silent: true }); } });
  Object.defineProperty(wrap, 'disabled', { configurable: true, get: function () { return disabled; }, set: function (v) {
    disabled = !!v;
    input.disabled = disabled;
    openBtn.disabled = disabled;
    wrap.classList.toggle('is-disabled', disabled);
    if (disabled && picker) closePicker(false);
    paint();
  } });
  wrap.read = function () { return value; };
  wrap.set = function (v, opts) { set(v, opts || {}); };
  wrap.input = input;
  wrap.open = open;
  wrap.close = function () { if (picker) closePicker(false); };
  wrap.focus = function () { input.focus(); };
  wrap.setMin = function (v) { min = v || null; };
  wrap.setMax = function (v) { max = v || null; };
  wrap.zone = zone;
  wrap.mode = mode;
  if (o.disabled) wrap.disabled = true;
  paint();
  return wrap;
}
