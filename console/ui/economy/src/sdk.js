ICONS['sdk'] = '<path d="M16 18l6-6-6-6M8 6l-6 6 6 6"/>';
PH_ICONS['sdk'] = '<path d="M69.12,94.15,28.5,128l40.62,33.85a8,8,0,1,1-10.24,12.29l-48-40a8,8,0,0,1,0-12.29l48-40a8,8,0,0,1,10.24,12.3Zm176,27.7-48-40a8,8,0,1,0-10.24,12.3L227.5,128l-40.62,33.85a8,8,0,1,0,10.24,12.29l48-40a8,8,0,0,0,0-12.29ZM162.73,32.48a8,8,0,0,0-10.25,4.79l-64,176a8,8,0,0,0,4.79,10.26A8.14,8.14,0,0,0,96,224a8,8,0,0,0,7.52-5.27l64-176A8,8,0,0,0,162.73,32.48Z"/>';

var SDK_REPO = 'https://github.com/copilot-rast/nakama-farsi-unity.git';
var SDK_WEB = 'https://github.com/copilot-rast/nakama-farsi-unity';
var SDK_PACKAGE = 'com.nakamafarsi.unity';

function sdkSection(n, title, body) {
  return card(h('div', { class: 'p-4 flex flex-col gap-3', 'data-sdk-step': String(n) },
    h('div', { class: 'flex items-center gap-2' },
      h('span', { class: 'eco-sdk-n', text: fmtNum(n) }),
      h('h2', { class: 'font-semibold', text: title })),
    body));
}

function sdkText(key) {
  return h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 eco-pre-line', text: t(key) });
}

function sdkList(keys) {
  return h('ul', { class: 'eco-sdk-list text-sm text-secondary-700 dark:text-secondary-200' }, keys.map(function (k) { return h('li', { text: t(k) }); }));
}

function sdkCode(text) {
  var btn = button({ variant: 'muted', size: 'sm', icon: 'copy', label: t('sdk.copy'), onclick: function () {
    copyText(text).then(function () { toast('success', t('up.copied')); }, function () { toast('error', t('up.copyFailed')); });
  } });
  btn.setAttribute('data-up-copy', '');
  return h('div', { class: 'eco-up-code eco-sdk-code', 'data-l10n-skip': '' },
    h('pre', { class: 'font-mono text-xs', dir: 'ltr', text: text }),
    h('div', { class: 'eco-sdk-copy' }, btn));
}

function connCopy(text, label) {
  var btn = button({ variant: 'muted', size: 'sm', icon: 'copy', label: t('sdk.copy'), onclick: function () {
    copyText(text).then(function () { toast('success', t('up.copied')); }, function () { toast('error', t('up.copyFailed')); });
  } });
  btn.setAttribute('data-conn-copy', '');
  btn.setAttribute('aria-label', t('cn.copyOf', { name: label }));
  return btn;
}

function connRow(name, label, value, copy) {
  var val = value == null
    ? h('span', { class: 'eco-conn-ask', text: t(name === 'key' ? 'cn.key.hidden' : 'cn.host.ask') })
    : h('code', { class: 'eco-conn-val font-mono text-xs', dir: 'ltr', 'data-l10n-skip': '', text: String(value) });
  return h('div', { class: 'eco-conn-row', 'data-conn': name },
    h('dt', { class: 'eco-conn-label', text: label }),
    h('dd', { class: 'eco-conn-cell' }, val, value != null && copy ? connCopy(String(value), label) : null));
}

function connectionView(c) {
  c = c || {};
  var rows = h('dl', { class: 'eco-conn-list' },
    connRow('url', t('cn.url'), c.hostSet ? c.url : null, true),
    c.hostSet ? connRow('host', t('cn.host'), c.host, true) : null,
    c.hostSet ? connRow('port', t('cn.port'), String(c.port), false) : null,
    c.hostSet ? connRow('ssl', t('cn.ssl'), String(!!c.ssl), false) : null,
    connRow('key', t('cn.key'), c.keyShown ? c.serverKey : null, true));
  return h('div', { class: 'eco-conn flex flex-col gap-2', 'data-connection': '', 'data-key-shown': String(!!c.keyShown), 'data-host-set': String(!!c.hostSet) },
    rows,
    c.keyShown && c.defaultKey ? h('div', { 'data-default-key': '' }, notice('warning', t('cn.default'))) : null,
    c.hostSet ? null : h('div', { 'data-host-missing': '' }, notice('info', t('cn.host.missing'))));
}

function sdkConnection(slot) {
  clear(slot).appendChild(h('div', { class: 'flex flex-col gap-2' }, bar('60%'), bar('80%')));
  baseApi('GET', '/handoff/connection').then(function (res) {
    clear(slot).appendChild(connectionView(res && res.connection));
  }, function (err) {
    clear(slot).appendChild(notice('error', t('cn.loadFailed') + ' ' + errorText(err)));
  });
}

function sdkPage(container) {
  var frame = pageFrame(t('nav.sdk'), t('desc.sdk'));
  clear(container).appendChild(frame.el);
  var manifest = '"' + SDK_PACKAGE + '": "' + SDK_REPO + '?path=/Packages/Nakama#dev"';
  var setup = 'var client = NakamaFarsiRunner.Initialize(NakamaFarsiConfig.Load());\nawait client.LoginWithDeviceAsync(SystemInfo.deviceUniqueIdentifier);\n\nvar boards = await client.Rpcs.Leaderboardx.ListAsync();\nvar tasks  = await client.Rpcs.Economy.TasksListAsync();';
  var phone = 'var sent   = await client.Phone.RequestCodeAsync("09121234567");\nvar result = await client.Phone.VerifyAsync(sent.Phone, code);';
  var pay = 'await NakamaFarsiStore.SetupAsync(client);\nvar store = await NakamaFarsiStore.BuyAsync(productId);\n\nvar web = await NakamaFarsiGateway.BuyAsync(productId, PaymentMethods.ZarinPal);\nvar coins = await client.Payments.BuyAsync(productId, PaymentMethods.Coins);';
  var chat = 'var room = await client.Chat.JoinRoomAsync("lobby");\nawait client.Chat.SendTextAsync(room.Id, "salam");\nawait client.Chat.SendStickerAsync(room.Id, "happy/smile");';
  var connSlot = h('div', { 'data-sdk-connection': '' });
  var repoLink = h('a', { href: SDK_WEB, target: '_blank', rel: 'noopener', class: 'eco-link font-mono text-sm', dir: 'ltr', text: SDK_WEB });
  var box = h('div', { class: 'flex flex-col gap-4 eco-sdk', 'data-sdk-page': '' },
    sdkSection(1, t('sdk.what'), h('div', { class: 'flex flex-col gap-3' },
      sdkText('sdk.what.text'),
      sdkList(['sdk.f.login', 'sdk.f.rpc', 'sdk.f.pay', 'sdk.f.store', 'sdk.f.chat', 'sdk.f.game']),
      h('div', { class: 'flex flex-wrap items-center gap-2 text-sm' }, h('span', { class: 'text-secondary-500', text: t('sdk.repo') }), repoLink))),
    sdkSection(2, t('sdk.need'), sdkList(['sdk.need.unity', 'sdk.need.git'])),
    sdkSection(3, t('sdk.install'), h('div', { class: 'flex flex-col gap-3' }, sdkText('sdk.install.text'), sdkCode(manifest), notice('warning', t('sdk.install.warn'))))
    , sdkSection(4, t('sdk.config'), h('div', { class: 'flex flex-col gap-3' }, sdkText('sdk.config.text'), sdkList(['sdk.config.url', 'sdk.config.key', 'sdk.config.lang', 'sdk.config.link']), connSlot))
    , sdkSection(5, t('sdk.start'), h('div', { class: 'flex flex-col gap-3' }, sdkText('sdk.start.text'), sdkCode(setup), sdkText('sdk.phone.text'), sdkCode(phone)))
    , sdkSection(6, t('sdk.pay'), h('div', { class: 'flex flex-col gap-3' }, sdkText('sdk.pay.text'), sdkCode(pay), sdkList(['sdk.pay.bazaar', 'sdk.pay.switch', 'sdk.pay.test'])))
    , sdkSection(7, t('sdk.chat'), h('div', { class: 'flex flex-col gap-3' }, sdkText('sdk.chat.text'), sdkCode(chat)))
    , sdkSection(8, t('sdk.errors'), sdkText('sdk.errors.text')));
  frame.content.appendChild(box);
  sdkConnection(connSlot);
}

CUSTOM_PAGES['sdk'] = function (root) { sdkPage(root); };
