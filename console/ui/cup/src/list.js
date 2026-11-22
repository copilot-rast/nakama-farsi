var CU_STATUS_COLOR = { draft: 'secondary', upcoming: 'primary', signup: 'primary', running: 'warning', finished: 'success', cancelled: 'secondary' };
var CU_STATUSES = ['draft', 'upcoming', 'signup', 'running', 'finished', 'cancelled'];

function statusTag(s) {
  return tag(t('st.' + s), CU_STATUS_COLOR[s] || 'secondary');
}

function roundName(key) {
  if (!key) return '';
  var m = /^r(\d+)$/.exec(key);
  return m ? t('rk.of', { n: Number(m[1]) }) : t('rk.' + key);
}

function cupName(c) {
  return textOf(c && c.name, c && (c.templateKey || (c.id || '').slice(0, 8)));
}

function viewHref(id) {
  return '#/cups/list/' + encodeURIComponent(id);
}

function cupsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 25, total: 0, status: CU_STATUSES.indexOf(q.status) !== -1 ? q.status : '', q: '', template: q.template || '' };
  var frame = pageFrame('cups/list', t('desc.cups/list'), [
    button({ label: t('cu.new'), variant: 'primary', data: { 'cup-action': 'new-cup' }, onclick: function () { cupEditor(null, function (c) { location.hash = viewHref(c.id); }); } })]);
  clear(root).appendChild(frame.el);
  var status = select(state.status, [{ value: '', label: t('all') }].concat(opt('st.', CU_STATUSES)), { name: 'status', label: t('f.state') });
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cu-filters' }, field(t('f.state'), status), field(t('search'), search, t('cu.hint.search'))));
  if (state.template) frame.content.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm' }, h('span', { text: t('cu.byTemplate') }), mono(state.template),
    h('a', { href: '#/cups/list', class: 'eco-link', 'data-cup-clear-template': '', text: t('cu.allCups') })));
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/cups/list' + qs({ status: state.status, q: state.q, template: state.template, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      if (!items.length) { listBox.appendChild(emptyCard(state.q || state.status ? t('empty') : t('cu.empty'))); return; }
      listBox.appendChild(h('div', { 'data-cup-list': 'cups' }, table([
        { label: t('f.name') }, { label: t('f.state') }, { label: t('ru.mode') }, { label: t('cu.players') }, { label: t('cu.starts') }, { label: t('cu.round') }
      ], items.map(function (c) {
        var r = c.rules || {};
        return { id: c.id, item: c, cells: [
          h('a', { href: viewHref(c.id), class: 'flex flex-col min-w-0', 'data-cup-open': c.id },
            h('span', { class: 'font-medium eco-link', dir: 'auto', text: cupName(c) }),
            h('span', { class: 'text-xs text-secondary-500', text: rulesSummary(r) })),
          statusTag(c.status),
          h('span', { text: t('mode.' + ((r.timing && r.timing.mode) || 'window')) }),
          h('span', { class: 'tabular-nums', 'data-cup-entrants': '', text: t('cu.entrantsV', { n: fmtNum(c.status === 'running' || c.status === 'finished' ? c.players : c.entrants || 0), cap: fmtNum(r.capacity || 0) }) }),
          h('span', { class: 'text-xs', text: fmtDate(c.startsAt) }),
          h('span', { class: 'text-xs', text: c.status === 'running' && c.currentRound ? t('cu.roundV', { n: c.currentRound, of: c.rounds }) : c.status === 'finished' ? t('st.finished') : t('none') })
        ] };
      }), function (c) { location.hash = viewHref(c.id); })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', cuErrText(err))); });
  }
  reload();
}

function soon(minutes) {
  var d = new Date(Date.now() + minutes * 60000);
  d.setSeconds(0, 0);
  return d.toISOString();
}

function cupEditor(c, done) {
  var editing = !!(c && c.id);
  var open = !editing || c.status === 'draft' || c.status === 'upcoming' || c.status === 'signup';
  var body = h('div', { 'data-cup-form': 'cup' }, loading());
  var save = button({ label: editing ? t('save') : t('cu.create'), variant: 'primary', data: { 'cup-action': 'save-cup' }, disabled: true });
  sidePanel(editing ? t('cu.edit') : t('cu.new'), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'cup');
  Promise.all([loadRefs('cups', true), loadLanguages()]).then(function (res) {
    var refs = res[0];
    clear(body);
    var name = translated(t('f.name'), 'name', c && c.name);
    var desc = translated(t('f.description'), 'description', c && c.description, { multiline: true });
    var icon = input(c && c.iconUrl || '', { name: 'iconUrl', ltr: true, placeholder: 'https://' });
    var signupAt = localInput(c ? c.signupAt : soon(0), 'signupAt');
    var startsAt = localInput(c ? c.startsAt : soon(60), 'startsAt');
    var announceAt = localInput(c ? c.announceAt : null, 'announceAt');
    var publish = editing ? null : check(false, t('cu.publish'), 'publish');
    var rules = rulesForm(c ? c.rules : refs.defaults, { locked: !open });
    var msg = h('div', { 'data-cup-msg': '' });
    body.appendChild(name);
    body.appendChild(desc);
    body.appendChild(field(t('f.icon'), icon));
    var times = grid(field(t('cu.signupAt'), signupAt, t('cu.hint.signupAt')), field(t('cu.startsAt'), startsAt, t('cu.hint.startsAt')));
    body.appendChild(times);
    body.appendChild(field(t('cu.announceAt'), announceAt, t('cu.hint.announceAt')));
    if (!open) {
      [signupAt, startsAt, announceAt].forEach(function (x) { x.disabled = true; });
      body.appendChild(notice('warning', t('cu.lockedRules')));
      body.appendChild(h('div', { class: 'h-3' }));
    }
    if (publish) body.appendChild(publish);
    rules.el.forEach(function (s) { body.appendChild(s); body.appendChild(h('div', { class: 'h-3' })); });
    body.appendChild(msg);
    save.disabled = false;
    save.addEventListener('click', function () {
      clear(msg);
      var payload = { name: name.read(), description: desc.read(), iconUrl: icon.value.trim() };
      if (!Object.keys(payload.name).length) { msg.appendChild(notice('error', t('cu.needName'))); return; }
      if (open) {
        var rr = rules.read();
        if (rr.error) { msg.appendChild(notice('error', rr.error)); return; }
        payload.rules = rr.rules;
        payload.signupAt = signupAt.read();
        payload.startsAt = startsAt.read();
        payload.announceAt = announceAt.read();
        if (!payload.signupAt || !payload.startsAt) { msg.appendChild(notice('error', t('cu.needTimes'))); return; }
      }
      if (publish) payload.publish = publish.input.checked;
      busy(save, true);
      api(editing ? 'PUT' : 'POST', '/cups/list' + (editing ? '/' + c.id : ''), payload).then(function (out) {
        closePanel();
        toast('success', editing ? t('saved') : t('cu.created'));
        if (done) done(out.cup, out);
      }, function (err) { busy(save, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', cuErrText(err))); });
}
