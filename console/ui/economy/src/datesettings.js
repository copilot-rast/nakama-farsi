function dsZones(current) {
  var list = TIMEZONES.slice();
  if (current && list.indexOf(current) === -1) list.unshift(current);
  return list.map(function (z) { return { value: z, label: z === 'UTC' ? 'UTC' : z.replace(/_/g, ' ') }; });
}

function dsFields(current) {
  return [
    { name: 'timezone', type: 'select', label: 'f.ds.timezone', hint: 'f.ds.timezone.hint', required: true, noPlaceholder: true, options: function () { return dsZones(current); } },
    { name: 'calendarNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', 'data-ds-calendar-note': '', text: t('ds.calendarNote', { calendar: window.NakamaDate ? window.NakamaDate.calendarName() : '' }) }); } }
  ];
}

function dateSettingsPage(container) {
  var frame = tabFrame(t('desc.datetime'));
  clear(container).appendChild(frame.el);
  var box = card(h('div', { class: 'p-4 md:p-6' }, formSkeleton({ fields: dsFields('') })));
  box.setAttribute('data-settings-form', 'datetime');
  frame.content.appendChild(box);
  var body = box.firstChild;

  function load() {
    baseApi('GET', '/console-settings').then(function (data) {
      if (!box.isConnected) return;
      paint(data || {});
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }

  function paint(data) {
    var zone = data.timezone || (window.NakamaDate ? window.NakamaDate.defaultZone : 'Asia/Tehran');
    var values = { timezone: zone };
    var top = h('div', {});
    var form = buildForm(dsFields(zone), values, {});
    var lead = form.el.firstElementChild;
    if (lead && lead.classList.contains('border-t')) lead.classList.add('eco-flush');
    var now = h('p', { class: 'text-sm text-secondary-700 dark:text-secondary-200 mb-4', 'data-ds-now': '' });
    function paintNow(z) {
      now.textContent = t('ds.now', { time: window.NakamaDate ? window.NakamaDate.format(new Date().toISOString(), { zone: z, style: 'long' }) : fmtDate(new Date().toISOString()), zone: z });
    }
    paintNow(zone);
    form.el.addEventListener('change', function () { var p = form.payload(); if (p.timezone) paintNow(p.timezone); });
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    saveBtn.setAttribute('data-settings-save', '');
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } },
      top, now, form.el,
      data.updatedBy ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-3', text: t('up.savedBy', { who: data.updatedBy, when: fmtDate(data.updatedAt) }) }) : null,
      h('div', { class: 'flex flex-wrap justify-end gap-2 pt-2' }, saveBtn),
      h('button', { type: 'submit', hidden: true })));

    function save() {
      clear(top);
      if (!form.validate()) return;
      setBusy(saveBtn, true, t('saving'));
      baseApi('PUT', '/console-settings', { timezone: form.payload().timezone }).then(function (res) {
        if (window.NakamaDate && res && res.timezone) window.NakamaDate.setZone(res.timezone);
        toast('success', t('saved'));
        paint(res || data);
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
      });
    }
  }

  load();
}
