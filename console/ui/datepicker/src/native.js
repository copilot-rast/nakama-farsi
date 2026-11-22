var ND_NATIVE_TYPES = { date: 'date', 'datetime-local': 'datetime', time: 'time' };
var ND_NATIVE_SCOPE = '.form-mode';
var ndNative = { observer: null, items: [] };
var ndValueDesc = typeof HTMLInputElement !== 'undefined' ? Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value') : null;

function nativeToWire(raw, mode) {
  if (!raw) return null;
  if (mode === 'datetime') {
    var m = /^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2})/.exec(raw);
    return m ? m[1] + 'T' + m[2] + ':00.000Z' : null;
  }
  if (mode === 'time') return raw.slice(0, 5);
  return raw.slice(0, 10);
}

function wireToNative(v, mode, withSeconds) {
  if (!v) return '';
  if (mode === 'datetime') return new Date(v).toISOString().slice(0, 16);
  if (mode === 'time') return v + (withSeconds ? ':00' : '');
  return v;
}

function upgradeNative(el) {
  if (el.__ndUpgraded || !ndValueDesc) return;
  var mode = ND_NATIVE_TYPES[(el.getAttribute('type') || '').toLowerCase()];
  if (!mode || !el.closest(ND_NATIVE_SCOPE) || el.closest('.nd-field, [data-nd-skip]')) return;
  el.__ndUpgraded = true;
  var withSeconds = mode === 'time' && el.getAttribute('step') === '1';
  var syncing = false;
  var label = el.getAttribute('aria-label') || el.getAttribute('placeholder') || null;
  var lab = el.id && document.querySelector('label[for="' + (window.CSS && CSS.escape ? CSS.escape(el.id) : el.id) + '"]');
  if (lab) label = lab.textContent.trim() || label;
  var picker = ndInput({
    mode: mode,
    value: nativeToWire(ndValueDesc.get.call(el), mode),
    label: label,
    inputClass: el.className,
    zoneLine: mode !== 'time',
    onchange: function (v) {
      syncing = true;
      try {
        ndValueDesc.set.call(el, wireToNative(v, mode, withSeconds));
        el.dispatchEvent(new Event('input', { bubbles: true }));
        el.dispatchEvent(new Event('change', { bubbles: true }));
      } finally { syncing = false; }
    }
  });
  picker.classList.add('nd-native');
  picker.setAttribute('data-nd-native', el.getAttribute('type'));
  Object.defineProperty(el, 'value', {
    configurable: true,
    get: function () { return ndValueDesc.get.call(el); },
    set: function (v) {
      ndValueDesc.set.call(el, v);
      if (!syncing) picker.set(nativeToWire(v, mode), { silent: true });
    }
  });
  el.classList.add('nd-native-source');
  el.setAttribute('aria-hidden', 'true');
  el.setAttribute('tabindex', '-1');
  picker.disabled = el.disabled;
  el.parentNode.insertBefore(picker, el.nextSibling);
  ndNative.items.push({ el: el, picker: picker });
}

function scanNative(root) {
  if (!root || !root.querySelectorAll) return;
  if (root.matches && root.matches('input[type="date"], input[type="datetime-local"], input[type="time"]')) upgradeNative(root);
  Array.prototype.forEach.call(root.querySelectorAll('input[type="date"], input[type="datetime-local"], input[type="time"]'), upgradeNative);
}

function tidyNative() {
  ndNative.items = ndNative.items.filter(function (it) {
    if (it.el.isConnected) {
      if (it.picker.disabled !== it.el.disabled) it.picker.disabled = it.el.disabled;
      if (it.picker.parentNode !== it.el.parentNode || it.el.nextSibling !== it.picker) it.el.parentNode.insertBefore(it.picker, it.el.nextSibling);
      return true;
    }
    it.picker.remove();
    return false;
  });
}

function watchNative() {
  if (ndNative.observer || typeof MutationObserver === 'undefined' || !document.body) return;
  var pending = false;
  ndNative.observer = new MutationObserver(function (list) {
    for (var i = 0; i < list.length; i++) {
      var r = list[i];
      if (r.type === 'attributes') { pending = true; continue; }
      for (var j = 0; j < r.addedNodes.length; j++) if (r.addedNodes[j].nodeType === 1) scanNative(r.addedNodes[j]);
      if (r.removedNodes.length) pending = true;
    }
    if (pending && ndNative.items.length) { pending = false; tidyNative(); }
  });
  ndNative.observer.observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['disabled'] });
  scanNative(document.body);
}
