function templatesPage(root) {
  var frame = pageFrame('cups/templates', t('desc.cups/templates'), [
    button({ label: t('tp.new'), variant: 'primary', data: { 'cup-action': 'new-template' }, onclick: function () { templateEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/cups/templates'), loadLanguages()]).then(function (res) {
      var items = res[0].items || [];
      clear(listBox);
      if (!items.length) { listBox.appendChild(emptyCard(t('tp.empty'))); return; }
      listBox.appendChild(h('div', { 'data-cup-list': 'templates' }, table([
        { label: t('f.name') }, { label: t('f.state') }, { label: t('tp.schedule') }, { label: t('tp.nextStart') }, { label: t('tp.openCups') }, { label: t('f.updated') }
      ], items.map(function (x) {
        return { id: x.key, item: x, cells: [
          h('div', { class: 'flex flex-col min-w-0' }, h('span', { class: 'font-medium', dir: 'auto', text: textOf(x.name, x.key) }), mono(x.key), h('span', { class: 'text-xs text-secondary-500', text: rulesSummary(x.rules) })),
          tag(x.active ? t('active') : t('inactive'), x.active ? 'success' : 'secondary'),
          x.schedule ? mono(x.schedule) : h('span', { class: 'text-secondary-500', text: t('tp.manual') }),
          h('span', { class: 'text-xs', text: x.active && x.nextStart ? fmtDate(x.nextStart) : t('none') }),
          x.openCups ? h('a', { href: '#/cups/list?template=' + encodeURIComponent(x.key), class: 'eco-link tabular-nums', text: fmtNum(x.openCups) }) : h('span', { text: fmtNum(0) }),
          h('span', { class: 'text-xs', dir: 'auto', text: (x.updatedBy ? x.updatedBy + ' · ' : '') + fmtDate(x.updatedAt) })
        ] };
      }), function (x) { templateEditor(x, reload); })));
    }, function (err) { clear(listBox).appendChild(notice('error', cuErrText(err))); });
  }
  reload();
}

function templateEditor(x, done) {
  var editing = !!(x && x.id);
  var body = h('div', { 'data-cup-form': 'template' }, loading());
  var save = button({ label: t('save'), variant: 'primary', data: { 'cup-action': 'save-template' }, disabled: true });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (editing) {
    foot.unshift(button({ label: t('tp.runNow'), data: { 'cup-action': 'run-template' }, onclick: function () {
      confirmBox(t('tp.runConfirm'), t('tp.runNow')).then(function (ok) {
        if (!ok) return;
        api('POST', '/cups/templates/' + x.id + '/run', {}).then(function (out) { closePanel(); toast('success', t('cu.created')); location.hash = viewHref(out.cup.id); }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
    foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'cup-action': 'delete-template' }, onclick: function () {
      confirmBox(t('tp.deleteConfirm', { key: x.key }), t('delete')).then(function (ok) {
        if (!ok) return;
        api('DELETE', '/cups/templates/' + x.id).then(function () { closePanel(); toast('success', t('deleted')); done(); }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
  }
  sidePanel(editing ? t('tp.edit') : t('tp.new'), body, foot, 'template');
  Promise.all([loadRefs('cups', true), loadLanguages()]).then(function (res) {
    var refs = res[0];
    clear(body);
    var key = input(x ? x.key : '', { name: 'key', ltr: true, mono: true, placeholder: 'weekly_cup' });
    var name = translated(t('f.name'), 'name', x && x.name);
    var desc = translated(t('f.description'), 'description', x && x.description, { multiline: true });
    var icon = input(x && x.iconUrl || '', { name: 'iconUrl', ltr: true, placeholder: 'https://' });
    var active = check(x ? x.active : false, t('tp.active'), 'active');
    var schedule = input(x ? x.schedule : '0 18 * * 5', { name: 'schedule', ltr: true, mono: true, placeholder: '0 18 * * 5' });
    var signup = numInput(x ? x.signupMinutes : 1440, 'signupMinutes', t('tp.signupMinutes'));
    var announce = numInput(x ? x.announceMinutes : 0, 'announceMinutes', t('tp.announceMinutes'));
    var sort = numInput(x ? x.sort : 0, 'sort', t('f.sort'));
    var rules = rulesForm(x ? x.rules : refs.defaults);
    var msg = h('div', { 'data-cup-msg': '' });
    body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-4', text: t('tp.hint') }));
    body.appendChild(field(t('f.key'), key, t('hint.key')));
    body.appendChild(name);
    body.appendChild(desc);
    body.appendChild(field(t('f.icon'), icon));
    body.appendChild(active);
    body.appendChild(field(t('tp.schedule'), schedule, t('tp.hint.schedule')));
    body.appendChild(grid3(field(t('tp.signupMinutes'), signup, t('tp.hint.signup')), field(t('tp.announceMinutes'), announce, t('tp.hint.announce')), field(t('f.sort'), sort)));
    rules.el.forEach(function (s) { body.appendChild(s); body.appendChild(h('div', { class: 'h-3' })); });
    body.appendChild(msg);
    save.disabled = false;
    save.addEventListener('click', function () {
      clear(msg);
      var rr = rules.read();
      if (rr.error) { msg.appendChild(notice('error', rr.error)); return; }
      var n = [signup, announce, sort].map(function (el) { return numberValue(el); });
      if (n.some(function (v) { return v === null || isNaN(v); })) { msg.appendChild(notice('error', t('ru.badNumber', { field: t('tp.signupMinutes') }))); return; }
      var payload = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), active: active.input.checked, schedule: schedule.value.trim(),
        signupMinutes: n[0], announceMinutes: n[1], sort: n[2], rules: rr.rules };
      busy(save, true);
      api(editing ? 'PUT' : 'POST', '/cups/templates' + (editing ? '/' + x.id : ''), payload).then(function () { closePanel(); toast('success', t('saved')); done(); }, function (err) { busy(save, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', cuErrText(err))); });
}
