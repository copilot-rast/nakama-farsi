var SKY = {
  light: { top: '#dfe7f3', mid: '#e9edf3', low: '#f4ece6', wash: 'rgb(170 196 236 / .55)', wash2: 'rgb(248 212 192 / .55)', glow: 'rgb(255 222 190 / .85)', sun: 'rgb(255 246 234 / .9)', sunR: 48, h1: '#d5dcea', h2: '#ccd5e6', h3: '#c5cfe2' },
  dark: { top: '#0b0e1a', mid: '#111528', low: '#1a1630', wash: 'rgb(70 96 190 / .30)', wash2: 'rgb(120 70 150 / .18)', glow: 'rgb(255 214 150 / .20)', sun: 'rgb(255 236 200 / .55)', sunR: 34, h1: '#171a2e', h2: '#121424', h3: '#0d0f1a' }
};

function skyStars() {
  var s = 7;
  var rnd = function () { s = (s * 16807) % 2147483647; return s / 2147483647; };
  var out = '';
  for (var i = 0; i < 140; i++) {
    var cx = (rnd() * 1440).toFixed(1);
    var cy = (rnd() * 560).toFixed(1);
    var r = rnd() < 0.85 ? 0.7 : 1.2;
    out += '<circle cx="' + cx + '" cy="' + cy + '" r="' + r + '" fill="#fff" opacity="' + (0.25 + rnd() * 0.45).toFixed(2) + '"/>';
  }
  return out;
}

function skySvg(dark) {
  var c = dark ? SKY.dark : SKY.light;
  var stop = function (id, cx, cy, r, color) {
    return '<radialGradient id="' + id + '" cx="' + cx + '" cy="' + cy + '" r="' + r + '"><stop offset="0" stop-color="' + color + '"/><stop offset="1" stop-color="' + color + '" stop-opacity="0"/></radialGradient>';
  };
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false"><defs>' +
    '<linearGradient id="spk" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="' + c.top + '"/><stop offset=".6" stop-color="' + c.mid + '"/><stop offset="1" stop-color="' + c.low + '"/></linearGradient>' +
    stop('spw1', 0.82, 0.2, 0.5, c.wash) + stop('spw2', 0.35, 1, 0.55, c.wash2) + stop('spg', 0.22, 0.14, 0.42, c.glow) +
    '<filter id="sph" x="-10%" y="-10%" width="120%" height="120%"><feGaussianBlur stdDeviation="18"/></filter>' +
    '<filter id="sps" x="-150%" y="-150%" width="400%" height="400%"><feGaussianBlur stdDeviation="34"/></filter></defs>' +
    '<rect width="1440" height="900" fill="url(#spk)"/><rect width="1440" height="900" fill="url(#spw1)"/><rect width="1440" height="900" fill="url(#spw2)"/><rect width="1440" height="900" fill="url(#spg)"/>' +
    (dark ? skyStars() : '') +
    '<circle cx="300" cy="128" r="' + c.sunR + '" fill="' + c.sun + '" filter="url(#sps)"/>' +
    '<g filter="url(#sph)"><path d="M0 690 C160 640 300 650 440 676 C600 706 720 640 880 630 C1040 620 1180 668 1440 640 L1440 900 L0 900Z" fill="' + c.h1 + '"/>' +
    '<path d="M0 760 C200 720 360 744 520 760 C700 778 860 720 1040 716 C1200 712 1320 744 1440 734 L1440 900 L0 900Z" fill="' + c.h2 + '"/>' +
    '<path d="M0 836 C240 806 480 830 720 828 C960 826 1200 800 1440 812 L1440 900 L0 900Z" fill="' + c.h3 + '"/></g></svg>';
}

var skyEl = null;
function paintSky() {
  if (!document.body) return;
  var mode = root.classList.contains('dark') ? 'dark' : 'light';
  if (!skyEl) {
    skyEl = el('div', { class: 'sp-sky', 'aria-hidden': 'true', 'data-l10n-skip': '' });
  }
  if (skyEl.getAttribute('data-theme') !== mode) {
    skyEl.setAttribute('data-theme', mode);
    skyEl.innerHTML = skySvg(mode === 'dark');
  }
  if (skyEl.parentNode !== document.body || document.body.firstChild !== skyEl) document.body.insertBefore(skyEl, document.body.firstChild);
}
