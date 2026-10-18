// Converts the console stylesheet to direction-agnostic CSS.
//
// Physical properties (margin-left, left, text-align: right, ...) become their
// logical equivalents, which render identically in LTR and mirror in RTL
// without the cascade conflicts that per-rule [dir=rtl] overrides would cause.
// Horizontal translations cannot be expressed logically, so for those an
// html[dir=rtl] override with the negated offset is appended.
import postcss from 'postcss';

const PROPERTY_MAP = {
  'margin-left': 'margin-inline-start',
  'margin-right': 'margin-inline-end',
  'padding-left': 'padding-inline-start',
  'padding-right': 'padding-inline-end',
  'scroll-margin-left': 'scroll-margin-inline-start',
  'scroll-margin-right': 'scroll-margin-inline-end',
  'scroll-padding-left': 'scroll-padding-inline-start',
  'scroll-padding-right': 'scroll-padding-inline-end',
  left: 'inset-inline-start',
  right: 'inset-inline-end',
  'border-left': 'border-inline-start',
  'border-right': 'border-inline-end',
  'border-left-width': 'border-inline-start-width',
  'border-right-width': 'border-inline-end-width',
  'border-left-color': 'border-inline-start-color',
  'border-right-color': 'border-inline-end-color',
  'border-left-style': 'border-inline-start-style',
  'border-right-style': 'border-inline-end-style',
  'border-top-left-radius': 'border-start-start-radius',
  'border-top-right-radius': 'border-start-end-radius',
  'border-bottom-left-radius': 'border-end-start-radius',
  'border-bottom-right-radius': 'border-end-end-radius',
};

const VALUE_MAP = {
  'text-align': { left: 'start', right: 'end' },
  float: { left: 'inline-start', right: 'inline-end' },
  clear: { left: 'inline-start', right: 'inline-end' },
};

// Box shorthands whose 4-value form is asymmetric horizontally.
const BOX_SHORTHANDS = {
  margin: ['margin-block', 'margin-inline'],
  padding: ['padding-block', 'padding-inline'],
  inset: ['inset-block', 'inset-inline'],
  'border-width': ['border-block-width', 'border-inline-width'],
  'border-style': ['border-block-style', 'border-inline-style'],
  'border-color': ['border-block-color', 'border-inline-color'],
};

// Third-party widgets that manage their own direction (code editors) are left alone.
export const DEFAULT_EXCLUDE = /\.(monaco|codicon|mtk\d|view-line|xterm|iPadShowKeyboard|context-view|quick-input|suggest-widget|parameter-hints)/;

// splitValues splits a CSS value on top-level whitespace (ignoring calc() etc).
export function splitValues(value) {
  const out = [];
  let depth = 0;
  let current = '';
  for (const ch of value.trim()) {
    if (ch === '(') depth++;
    if (ch === ')') depth--;
    if (/\s/.test(ch) && depth === 0) {
      if (current) out.push(current);
      current = '';
    } else {
      current += ch;
    }
  }
  if (current) out.push(current);
  return out;
}

export function negate(value) {
  const v = value.trim();
  if (v === '0' || /^0[a-z%]*$/.test(v)) return v;
  if (v.startsWith('-')) {
    const rest = v.slice(1);
    return /^[\d.]/.test(rest) ? rest : `calc(-1 * ${v})`;
  }
  if (/^[\d.]/.test(v)) return `-${v}`;
  return `calc(-1 * ${v})`;
}

// negateTransform flips the horizontal component of translate functions.
export function negateTransform(value) {
  let changed = false;
  const out = value.replace(/\b(translateX|translate|translate3d)\(([^()]*(?:\([^()]*\)[^()]*)*)\)/g, (all, fn, args) => {
    if (/var\(--tw-translate-x\)/.test(args)) return all; // handled through the custom property
    const parts = args.split(',');
    const flipped = negate(parts[0]);
    if (flipped === parts[0].trim()) return all;
    changed = true;
    parts[0] = flipped;
    return `${fn}(${parts.join(',')})`;
  });
  return changed ? out : null;
}

function isExcluded(rule, exclude) {
  return exclude.test(rule.selector || '');
}

function insideKeyframes(node) {
  for (let p = node.parent; p; p = p.parent) {
    if (p.type === 'atrule' && /keyframes$/.test(p.name)) return true;
  }
  return false;
}

export function convertDeclaration(decl) {
  const prop = decl.prop.toLowerCase();
  if (PROPERTY_MAP[prop]) {
    decl.prop = PROPERTY_MAP[prop];
    return true;
  }
  if (VALUE_MAP[prop]) {
    const v = decl.value.trim().toLowerCase().replace(/\s*!important$/, '');
    const mapped = VALUE_MAP[prop][v];
    if (mapped) {
      decl.value = decl.value.replace(new RegExp(v, 'i'), mapped);
      return true;
    }
    return false;
  }
  if (BOX_SHORTHANDS[prop]) {
    const values = splitValues(decl.value.replace(/\s*!important$/i, ''));
    if (values.length !== 4 || values[1] === values[3]) return false;
    const [block, inline] = BOX_SHORTHANDS[prop];
    decl.cloneBefore({ prop: block, value: `${values[0]} ${values[2]}` });
    decl.cloneBefore({ prop: inline, value: `${values[3]} ${values[1]}` });
    decl.remove();
    return true;
  }
  if (prop === 'border-radius') {
    const raw = decl.value.replace(/\s*!important$/i, '');
    if (raw.includes('/')) return false;
    const values = splitValues(raw);
    let tl, tr, br, bl;
    if (values.length === 4) [tl, tr, br, bl] = values;
    else if (values.length === 3) [tl, tr, br, bl] = [values[0], values[1], values[2], values[1]];
    else return false;
    if (tl === tr && br === bl) return false;
    decl.cloneBefore({ prop: 'border-start-start-radius', value: tl });
    decl.cloneBefore({ prop: 'border-start-end-radius', value: tr });
    decl.cloneBefore({ prop: 'border-end-end-radius', value: br });
    decl.cloneBefore({ prop: 'border-end-start-radius', value: bl });
    decl.remove();
    return true;
  }
  return false;
}

export function scopeSelector(selector, scope) {
  return selector
    .split(/,(?![^(]*\))/)
    .map((s) => {
      const t = s.trim();
      if (/^(html|:root)\b/.test(t)) return t.replace(/^(html|:root)/, scope);
      return `${scope} ${t}`;
    })
    .join(',');
}

// toLogical returns the converted stylesheet plus statistics.
export function toLogical(css, { exclude = DEFAULT_EXCLUDE, rtlScope = 'html[dir="rtl"]' } = {}) {
  const root = postcss.parse(css);
  const stats = { converted: 0, rtlOverrides: 0, excludedRules: 0 };
  const overrides = [];
  root.walkRules((rule) => {
    if (insideKeyframes(rule)) return;
    if (isExcluded(rule, exclude)) {
      stats.excludedRules++;
      return;
    }
    const flipped = [];
    rule.walkDecls((decl) => {
      if (convertDeclaration(decl)) stats.converted++;
      const prop = decl.prop.toLowerCase();
      if (prop === '--tw-translate-x') {
        // Always emitted, even for 0, so state variants keep their precedence
        // relative to the other (more specific) RTL overrides.
        flipped.push({ prop, value: negate(decl.value) });
      } else if (prop === 'transform') {
        const v = negateTransform(decl.value);
        if (v) flipped.push({ prop, value: v, important: decl.important });
      }
    });
    if (flipped.length) {
      const clone = postcss.rule({ selector: scopeSelector(rule.selector, rtlScope) });
      flipped.forEach((d) => clone.append(postcss.decl(d)));
      overrides.push({ parent: rule.parent, rule: clone });
      stats.rtlOverrides++;
    }
  });
  // Overrides go at the end of the same container (media query) so they win.
  for (const { parent, rule } of overrides) parent.append(rule);
  return { css: root.toString(), stats };
}
