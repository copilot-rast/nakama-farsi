var QZ_IMPORT_SAMPLE = {
  packs: [{ key: 'general-1', name: { fa: 'دانش عمومی', en: 'General knowledge' }, questionCount: 10, passCorrect: 8, questionSeconds: 30 }],
  questions: [
    { key: 'gk-001', packKey: 'general-1', type: 'options', question: { fa: 'پایتخت فرانسه کجاست؟', en: 'Capital of France?' },
      options: [{ id: 0, text: { fa: 'لندن', en: 'London' } }, { id: 1, text: { fa: 'پاریس', en: 'Paris' }, correct: true }, { id: 2, text: { fa: 'رم', en: 'Rome' } }],
      explanation: { fa: 'پاریس پایتخت فرانسه است.' }, difficulty: 1, hints: [{ name: { fa: 'راهنما' }, text: { fa: 'شهر برج ایفل' }, price: { currencyKey: 'coin', amount: 50 } }] },
    { packKey: 'general-1', type: 'text', question: { en: 'Largest ocean on Earth?' }, answers: { en: ['Pacific', 'Pacific Ocean'], fa: ['آرام', 'اقیانوس آرام'] } }],
  rows: [{ 'Pack Key': 'general-1', 'Type': 'options', 'Question': 'Which numbers are even?', 'Option 1': '1', 'Option 2': '2', 'Option 3': '3', 'Option 4': '4', 'Correct Options': '2,4' }]
};

function importPage(root) {
  var frame = pageFrame('import', t('desc.import'));
  clear(root).appendChild(frame.el);
  var text = input('', { name: 'import-json', ltr: true, mono: true, multiline: true, rows: 14, placeholder: JSON.stringify(QZ_IMPORT_SAMPLE, null, 2) });
  var file = h('input', { type: 'file', accept: '.json,application/json', name: 'import-file', class: 'text-sm' });
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    var reader = new FileReader();
    reader.onload = function () { text.value = String(reader.result || ''); };
    reader.readAsText(f);
  });
  var packSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'import-pack', 'aria-label': t('im.pack') });
  var lang = input('fa', { name: 'import-lang', ltr: true });
  var result = h('div', { 'data-quiz-import-result': '' });
  var exportPack = h('select', { class: QZ_INPUT + ' eco-select', name: 'export-pack', 'aria-label': t('im.exportPack') });
  function run(dry, btn) {
    clear(result);
    var doc;
    try { doc = JSON.parse(text.value || '{}'); } catch (e) { result.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
    if (Array.isArray(doc)) doc = { rows: doc };
    if (packSel.value) doc.packKey = packSel.value;
    if (lang.value.trim()) doc.lang = lang.value.trim();
    doc.dryRun = dry;
    busy(btn, true);
    api('POST', '/quiz/import', doc).then(function (out) {
      busy(btn, false);
      var c = out.counts || {};
      var line = ['missions', 'packs', 'questions'].map(function (k) { var x = c[k] || {}; return t('im.count.' + k, { c: fmtNum(x.created || 0), u: fmtNum(x.updated || 0) }); }).join(' · ');
      result.appendChild(notice('success', (dry ? t('im.dryOk') : t('im.done')) + ' ' + line));
      qzRefs = null;
    }, function (err) {
      busy(btn, false);
      result.appendChild(notice('error', qzErrText(err)));
      var errs = err.meta && err.meta.errors;
      if (errs && errs.length) {
        result.appendChild(table([{ label: t('im.section') }, { label: t('im.row') }, { label: t('im.field') }, { label: t('im.error') }], errs.map(function (e, i) {
          return { id: String(i), cells: [h('span', { text: e.section }), h('span', { text: digits(String(e.index + 1)) }), mono(e.field), h('span', { text: has('code.' + e.code) ? t('code.' + e.code) : e.code })] };
        })));
      }
    });
  }
  var dryBtn = button({ label: t('im.dry'), data: { 'quiz-action': 'import-dry' }, onclick: function () { run(true, dryBtn); } });
  var goBtn = button({ label: t('im.go'), variant: 'primary', data: { 'quiz-action': 'import' }, onclick: function () { run(false, goBtn); } });
  var sampleBtn = button({ label: t('im.sample'), onclick: function () { text.value = JSON.stringify(QZ_IMPORT_SAMPLE, null, 2); } });
  var expBtn = button({ label: t('im.export'), data: { 'quiz-action': 'export' }, onclick: function () {
    busy(expBtn, true);
    api('GET', '/quiz/export' + qs({ packId: exportPack.value })).then(function (d) {
      busy(expBtn, false);
      var blob = new Blob([JSON.stringify(d, null, 2)], { type: 'application/json' });
      var a = h('a', { href: URL.createObjectURL(blob), download: 'quiz-export.json' });
      document.body.appendChild(a); a.click(); a.remove();
      toast('success', t('im.exported', { n: fmtNum((d.questions || []).length) }));
    }, function (err) { busy(expBtn, false); toast('error', qzErrText(err)); });
  } });
  frame.content.appendChild(section(t('im.title'), h('div', {},
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('im.help') }),
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('im.pack'), packSel, t('hint.importPack')), field(t('im.lang'), lang, t('hint.importLang'))),
    field(t('im.file'), file), field(t('im.json'), text), h('div', { class: 'flex flex-wrap gap-2 mb-3' }, sampleBtn, dryBtn, goBtn), result), 'import'));
  frame.content.appendChild(section(t('im.exportTitle'), h('div', {}, h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('im.exportHelp') }),
    h('div', { class: 'flex flex-wrap gap-2 items-end' }, field(t('im.exportPack'), exportPack), expBtn)), 'export'));
  Promise.all([api('GET', '/quiz/packs?limit=200'), loadLanguages()]).then(function (res) {
    var packs = res[0].items || [];
    [{ value: '', label: t('im.packFromFile') }].concat(packs.map(function (p) { return { value: p.key, label: textOf(p.name, p.key) }; })).forEach(function (o) { packSel.appendChild(h('option', { value: o.value, text: o.label })); });
    [{ value: '', label: t('all') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })).forEach(function (o) { exportPack.appendChild(h('option', { value: o.value, text: o.label })); });
  });
}
