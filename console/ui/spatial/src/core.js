var SP_KEY = 'nakama.spatial';
var SP_RAIL_KEY = 'nakama.spatial.rail';
var root = document.documentElement;

function spEnabled() {
  var m = /[?&]spatial=([01])\b/.exec(location.search) || /[?&]spatial=([01])\b/.exec(location.hash);
  try {
    if (m && m[1] === '0') localStorage.setItem(SP_KEY, '0');
    else if (m && m[1] === '1') localStorage.removeItem(SP_KEY);
    return localStorage.getItem(SP_KEY) !== '0';
  } catch (e) {
    return !(m && m[1] === '0');
  }
}

if (!spEnabled()) return;
root.classList.add('sp');

var mqReduce = window.matchMedia ? window.matchMedia('(prefers-reduced-motion: reduce)') : { matches: false };
var mqMobile = window.matchMedia ? window.matchMedia('(max-width: 767px)') : { matches: false };

function reduced() {
  return !!mqReduce.matches;
}

function isRtl() {
  return (root.getAttribute('dir') || 'ltr') === 'rtl';
}

function nc() {
  return window.NakamaConsole || null;
}

function el(tag, props) {
  var e = document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'text') e.textContent = v;
      else if (k === 'html') e.innerHTML = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') e.addEventListener(k.slice(2), v);
      else e.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) {
    var c = arguments[i];
    if (c === null || c === undefined || c === false) continue;
    e.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
  }
  return e;
}

var SP_ICON_PATHS = {
  star: '<path d="M11.525 2.295a.53.53 0 0 1 .95 0l2.31 4.679a2.123 2.123 0 0 0 1.595 1.16l5.166.756a.53.53 0 0 1 .294.904l-3.736 3.638a2.123 2.123 0 0 0-.611 1.878l.882 5.14a.53.53 0 0 1-.771.56l-4.618-2.428a2.122 2.122 0 0 0-1.973 0L6.396 21.01a.53.53 0 0 1-.77-.56l.881-5.139a2.122 2.122 0 0 0-.611-1.879L2.16 9.795a.53.53 0 0 1 .294-.906l5.165-.755a2.122 2.122 0 0 0 1.597-1.16z" fill="currentColor" stroke="none"/>',
  coins: '<circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="m16.71 13.88.7.71-2.82 2.82"/>',
  megaphone: '<path d="m3 11 18-5v12L3 14v-3z"/><path d="M11.6 16.8a3 3 0 1 1-5.8-1.6"/>',
  code: '<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>',
  help: '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>',
  gear: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  sliders: '<path d="M20 7h-9"/><path d="M14 17H5"/><circle cx="17" cy="17" r="3"/><circle cx="7" cy="7" r="3"/>',
  search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
  moon: '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>',
  chevron: '<path d="m9 18 6-6-6-6"/>',
  panel: '<rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/>',
  menu: '<path d="M4 6h16"/><path d="M4 12h16"/><path d="M4 18h16"/>',
  close: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
  external: '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
  activity: '<path d="M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2"/>',
  page: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/>'
};

function spIcon(name, cls) {
  var p = SP_ICON_PATHS[name] || SP_ICON_PATHS.page;
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="' + (cls || 'sp-ic') + (name === 'chevron' ? ' sp-ic-dir' : '') + '">' + p + '</svg>';
}

function currentHash() {
  var h = location.hash || '#/';
  return h === '#' ? '#/' : h;
}

function routePath(hash) {
  return (hash || currentHash()).replace(/^#/, '').split('?')[0] || '/';
}

function hrefActive(href, hash) {
  if (!href || href.indexOf('#/') !== 0) return false;
  var p = routePath(hash);
  var base = href.split('?')[0].replace(/^#/, '');
  if (base === '/') return p === '/';
  return p === base || p.indexOf(base + '/') === 0;
}

var spCleanups = [];
function onIdle(fn) {
  if (window.requestIdleCallback) window.requestIdleCallback(fn, { timeout: 300 });
  else setTimeout(fn, 60);
}
