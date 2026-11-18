ICONS['server-rules'] = '<path d="M12 2l8 4v6c0 5-3.5 9-8 10-4.5-1-8-5-8-10V6z"/><path d="M9 12l2 2 4-4"/>';
PH_ICONS['server-rules'] = '<path d="M208,40H48A16,16,0,0,0,32,56v56c0,52.72,25.52,84.67,46.93,102.19,23.06,18.86,46,25.27,47,25.53a8,8,0,0,0,4.2,0c1-.26,23.91-6.67,47-25.53C198.48,196.67,224,164.72,224,112V56A16,16,0,0,0,208,40Zm0,72c0,37.07-13.66,67.16-40.6,89.42A129.3,129.3,0,0,1,128,223.62a128.25,128.25,0,0,1-38.92-21.81C61.82,179.51,48,149.3,48,112l0-56,160,0ZM82.34,141.66a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35a8,8,0,0,1,11.32,11.32l-56,56a8,8,0,0,1-11.32,0Z"/>';

function srTextRows(name, label, hint, add, empty, max, when) {
  return { name: name, type: 'rows', label: label, hint: hint, addLabel: add, emptyText: empty, max: max, empty: [], when: when,
    newRow: function () { return { v: '' }; },
    columns: [{ name: 'v', type: 'text', ltr: true, label: label, required: true }] };
}

var SR_FIELDS = [
  { type: 'heading', label: 'sr.h.maintenance' },
  { name: 'maintenance.enabled', type: 'bool', inline: true, label: 'f.sr.mt.enabled', hint: 'f.sr.mt.enabled.hint' },
  tr('maintenance.message', { label: 'f.sr.mt.message', hint: 'f.sr.mt.message.hint', multiline: true, rows: 2, max: 500, optional: true }),
  { name: 'maintenance.until', type: 'datetime', label: 'f.sr.mt.until', hint: 'f.sr.mt.until.hint', optional: true },
  { name: 'maintenance.allowTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.sr.mt.tags', hint: 'f.sr.mt.tags.hint', optional: true, empty: [] },
  srTextRows('maintenance.allowRpcs', 'f.sr.mt.rpcs', 'f.sr.mt.rpcs.hint', 'sr.rpc.add', 'sr.rpc.empty', 100),
  { type: 'heading', label: 'sr.h.storage' },
  { type: 'row', fields: [
    { name: 'storage.blockWrites', type: 'bool', inline: true, label: 'f.sr.st.writes', hint: 'f.sr.st.writes.hint' },
    { name: 'storage.blockDeletes', type: 'bool', inline: true, label: 'f.sr.st.deletes' }
  ] },
  srTextRows('storage.allowCollections', 'f.sr.st.allow', 'f.sr.st.allow.hint', 'sr.col.add', 'sr.col.empty', 100),
  { type: 'heading', label: 'sr.h.delete' },
  { name: 'accountDelete.enabled', type: 'bool', inline: true, label: 'f.sr.ad.enabled', hint: 'f.sr.ad.enabled.hint' },
  { type: 'row', fields: [
    { name: 'accountDelete.mode', type: 'select', label: 'f.sr.ad.mode', required: true, options: function () { return [{ value: 'grace', label: t('sr.ad.grace') }, { value: 'immediate', label: t('sr.ad.immediate') }]; } },
    { name: 'accountDelete.graceDays', type: 'number', label: 'f.sr.ad.days', min: 1, max: 365, required: true, when: function (v) { return (v.accountDelete || {}).mode !== 'immediate'; } }
  ] },
  { type: 'heading', label: 'sr.h.docs' },
  { type: 'row', fields: [
    { name: 'docs.enabled', type: 'bool', inline: true, label: 'f.sr.docs.enabled', hint: 'f.sr.docs.enabled.hint' },
    { name: 'docs.consoleLogin', type: 'bool', inline: true, label: 'f.sr.docs.login', hint: 'f.sr.docs.login.hint', when: function (v) { return (v.docs || {}).enabled !== false; } }
  ] },
  { type: 'heading', label: 'sr.h.handoff' },
  { type: 'row', fields: [
    { name: 'handoff.enabled', type: 'bool', inline: true, label: 'f.sr.handoff.enabled', hint: 'f.sr.handoff.enabled.hint' },
    { name: 'handoff.audience', type: 'select', label: 'f.sr.handoff.audience', hint: 'f.sr.handoff.audience.hint', required: true, options: function () { return [{ value: 'all', label: t('sr.handoff.all') }, { value: 'admins', label: t('sr.handoff.admins') }]; }, when: function (v) { return (v.handoff || {}).enabled !== false; } }
  ] }
];

function srCall(method, path, body) {
  return baseApi(method, path, body).then(function (res) {
    var v = Object.assign({}, res || {});
    v.maintenance = Object.assign({ enabled: false, message: {}, allowRpcs: [], allowTagIds: [] }, v.maintenance || {});
    v.maintenance.allowRpcs = (v.maintenance.allowRpcs || []).map(function (x) { return { v: x }; });
    v.maintenance.message = v.maintenance.message || {};
    v.storage = Object.assign({ blockWrites: false, blockDeletes: false, allowCollections: [] }, v.storage || {});
    v.storage.allowCollections = (v.storage.allowCollections || []).map(function (x) { return { v: x }; });
    v.accountDelete = Object.assign({ enabled: true, mode: 'grace', graceDays: 14 }, v.accountDelete || {});
    v.docs = Object.assign({ enabled: true, consoleLogin: false }, v.docs || {});
    v.handoff = Object.assign({ enabled: true, audience: 'all' }, v.handoff || {});
    if (method === 'PUT') handoffLoad(true).catch(function () { return null; });
    return v;
  });
}

function serverRulesPage(container) {
  var frame = pageFrame(t('nav.server-rules'), t('desc.server-rules'));
  clear(container).appendChild(frame.el);
  var page = apiPage(srCall, '/server-rules', null, SR_FIELDS, {
    id: 'server-rules',
    onchange: function (v, extra) {
      clear(extra);
      if (v.maintenance && v.maintenance.enabled) extra.appendChild(notice('warning', t('sr.mt.on')));
      if (v.pendingDeletions) extra.appendChild(notice('info', t('sr.ad.pending', { n: fmtNum(v.pendingDeletions) })));
    },
    fixPayload: function (p) {
      delete p.updatedAt;
      delete p.updatedBy;
      delete p.pendingDeletions;
      p.maintenance.allowRpcs = (p.maintenance.allowRpcs || []).map(function (r) { return (r.v || '').trim(); }).filter(Boolean);
      p.maintenance.allowTagIds = p.maintenance.allowTagIds || [];
      p.maintenance.until = p.maintenance.until || null;
      p.storage.allowCollections = (p.storage.allowCollections || []).map(function (r) { return (r.v || '').trim(); }).filter(Boolean);
      if (!p.accountDelete.graceDays) p.accountDelete.graceDays = 14;
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['server-rules'] = function (root) { serverRulesPage(root); };
