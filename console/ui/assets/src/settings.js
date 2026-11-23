function assetsSettingsPage(root) {
  settingsPage(root, 'settings', '/assets', function (cur, defs, d) {
    var e = cur.entries, hs = cur.history, p = cur.packs;
    var kb = function (v) { return numInput(Math.round((v || 0) / 1024), ''); };
    var mb = function (v) { return numInput(Math.round((v || 0) / 1048576), ''); };
    var maxEntry = kb(e.maxEntryBytes); maxEntry.name = 'entries.maxEntryKb';
    var maxPush = numInput(e.maxPushItems, 'entries.maxPushItems'), maxFetch = numInput(e.maxFetchItems, 'entries.maxFetchItems'), maxEntries = numInput(e.maxEntries, 'entries.maxEntries');
    var conflict = select(e.conflictMode, [{ value: 'reject', label: t('se.conflict.reject') }, { value: 'skip', label: t('se.conflict.skip') }], { name: 'entries.conflictMode' });
    var force = check(e.allowForce, t('se.allowForce'), 'entries.allowForce');
    var access = select(e.access, [{ value: 'session', label: t('se.access.session') }, { value: 'public', label: t('se.access.public') }], { name: 'entries.access' });
    var cache = numInput(e.cacheSeconds, 'entries.cacheSeconds');
    var keepV = numInput(hs.keepVersions, 'history.keepVersions'), keepD = numInput(hs.keepDays, 'history.keepDays'), keepDel = numInput(hs.keepDeletedDays, 'history.keepDeletedDays');
    var platforms = input((p.platforms || []).join(', '), { name: 'packs.platforms', ltr: true });
    var part = kb(p.partBytes); part.name = 'packs.partKb';
    var maxFile = mb(p.maxFileBytes); maxFile.name = 'packs.maxFileMb';
    var maxPack = mb(p.maxPackBytes); maxPack.name = 'packs.maxPackMb';
    var maxFiles = numInput(p.maxFiles, 'packs.maxFiles'), uploadH = numInput(p.uploadHours, 'packs.uploadHours'), keepI = numInput(p.keepInactive, 'packs.keepInactive');
    var fromEditor = check(p.activateFromEditor, t('se.activateFromEditor'), 'packs.activateFromEditor');
    var single = check(p.singleActive, t('se.singleActive'), 'packs.singleActive');
    var pub = check(p.publicIndex, t('se.publicIndex'), 'packs.publicIndex');
    var idxCache = numInput(p.indexCacheSeconds, 'packs.indexCacheSeconds');
    var base = input(p.baseUrl, { name: 'packs.baseUrl', ltr: true, placeholder: d.serverBaseUrl || 'https://cdn.example.com' });
    var redirect = check(p.redirectToStorage, t('se.redirect'), 'packs.redirectToStorage');
    var pattern = input(p.catalogPattern, { name: 'packs.catalogPattern', ltr: true, mono: true });
    var reqHash = check(p.requireHash, t('se.requireHash'), 'packs.requireHash');
    var cl = cur.client || {};
    var reqOnline = check(cl.requireOnline, t('se.requireOnline'), 'client.requireOnline');
    var maxAge = numInput(cl.maxCacheAgeHours || 0, 'client.maxCacheAgeHours');
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    var info = h('div', { class: 'flex flex-col gap-1 text-sm' }, h('div', { text: tr('se.storage', {}, { d: d.storage }) }),
      d.publicStorage ? h('div', { text: t('se.publicStorage') }) : null, h('div', { dir: 'auto' }, tr('se.scopes', {}, { s: (d.scopes || []).join(', ') })));
    return {
      el: [section(t('se.info'), info, 'info'),
        section(t('se.entries'), h('div', {}, grid(field(t('se.maxEntryKb'), maxEntry), field(t('se.maxPushItems'), maxPush), field(t('se.maxFetchItems'), maxFetch), field(t('se.maxEntries'), maxEntries)),
          field(t('se.conflictMode'), conflict), force, grid(field(t('se.access'), access), field(t('se.cacheSeconds'), cache))), 'entries'),
        section(t('se.history'), grid(field(t('se.keepVersions'), keepV), field(t('se.keepDays'), keepD), field(t('se.keepDeletedDays'), keepDel)), 'history'),
        section(t('se.packs'), h('div', {}, field(t('se.platforms'), platforms), grid(field(t('se.partKb'), part, t('se.partHint')), field(t('se.maxFiles'), maxFiles), field(t('se.maxFileMb'), maxFile), field(t('se.maxPackMb'), maxPack),
          field(t('se.uploadHours'), uploadH), field(t('se.keepInactive'), keepI)), fromEditor, single, pub, field(t('se.indexCacheSeconds'), idxCache, null, 'as-num'),
          field(t('se.baseUrl'), base), redirect, field(t('se.catalogPattern'), pattern), reqHash), 'packs'),
        section(t('se.client'), h('div', {}, reqOnline, field(t('se.maxCacheAgeHours'), maxAge, t('se.maxCacheAgeHint'), 'as-num')), 'client')],
      read: function (c) {
        c.entries = { maxEntryBytes: numberValue(maxEntry) * 1024, maxPushItems: numberValue(maxPush), maxFetchItems: numberValue(maxFetch), maxEntries: numberValue(maxEntries) || 0,
          conflictMode: conflict.value, allowForce: force.input.checked, access: access.value, cacheSeconds: numberValue(cache) || 0 };
        c.history = { keepVersions: numberValue(keepV) || 0, keepDays: numberValue(keepD) || 0, keepDeletedDays: numberValue(keepDel) || 0 };
        c.packs = { platforms: platforms.value.split(/[,،\s]+/).map(function (x) { return x.trim(); }).filter(Boolean), partBytes: numberValue(part) * 1024,
          maxFileBytes: numberValue(maxFile) * 1048576, maxPackBytes: numberValue(maxPack) * 1048576, maxFiles: numberValue(maxFiles), uploadHours: numberValue(uploadH),
          keepInactive: numberValue(keepI) || 0, activateFromEditor: fromEditor.input.checked, singleActive: single.input.checked, publicIndex: pub.input.checked,
          indexCacheSeconds: numberValue(idxCache) || 0, baseUrl: base.value.trim(), redirectToStorage: redirect.input.checked, catalogPattern: pattern.value.trim(), requireHash: reqHash.input.checked };
        c.client = { requireOnline: reqOnline.input.checked, maxCacheAgeHours: numberValue(maxAge) || 0 };
        return null;
      }
    };
  });
}
