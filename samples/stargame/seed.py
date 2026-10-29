import json
import os
import sys
import urllib.parse
import urllib.request

BASE = os.environ.get("CONSOLE_URL", "http://127.0.0.1:7451")
ENV_FILE = os.path.expanduser("~/.config/nakama-farsi/mac-dev.env")


def read_env():
    out = {}
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            k, v = line.split("=", 1)
            k = k.replace("export ", "").strip()
            out[k] = v.strip().strip('"').strip("'")
    return out


def call(method, path, body=None, token=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE + path, data=data, method=method)
    req.add_header("content-type", "application/json")
    if token:
        req.add_header("authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(req) as r:
            raw = r.read()
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        sys.exit(f"{method} {path} -> {e.code} {e.read().decode()}")


def login():
    env = read_env()
    user = os.environ.get("CONSOLE_USER", env.get("CONSOLE_USER"))
    password = os.environ.get("CONSOLE_PASSWORD", env.get("CONSOLE_PASSWORD"))
    return call("POST", "/v2/console/authenticate", {"username": user, "password": password})["token"]


def find(token, path, match, q, extra=""):
    res = call("GET", f"{path}?limit=100{extra}&q=" + urllib.parse.quote(q), token=token)
    for item in res.get("items", []):
        if match(item):
            return item
    return None


def upsert(token, resource, body, field="key"):
    path = f"/v2/console/economy/{resource}"
    value = body[field]
    existing = find(token, path, lambda i: i.get(field) == value, value)
    if existing:
        return call("PUT", f"{path}/{existing['id']}", body, token)
    return call("POST", path, body, token)


def upsert_category(token, body):
    path = "/v2/console/base/categories"
    existing = find(token, path, lambda i: i.get("key") == body["key"] and i.get("keyType") == body["keyType"], body["key"], "&keyType=" + body["keyType"])
    if existing:
        return call("PUT", f"{path}/{existing['id']}", body, token)
    return call("POST", path, body, token)


def ensure_arabic(token):
    res = call("GET", "/v2/console/base/languages?limit=100", token=token)
    if any(i.get("code") == "ar" for i in res.get("items", [])):
        return
    call("POST", "/v2/console/base/languages", {
        "code": "ar", "name": {"fa": "عربی", "en": "Arabic", "ar": "العربية"}, "nativeName": "العربية",
        "direction": "rtl", "iconUrl": "", "isActive": True, "isDefault": False, "sort": 3,
    }, token)
    print("added language ar")


def main():
    token = login()
    ensure_arabic(token)

    coin_body = {"key": "coin", "name": {"fa": "سکه", "en": "Coin", "ar": "عملة"}, "type": "soft", "defaultValue": 50, "iconUrl": "", "metadata": {}, "sort": 1}
    coin = find(token, "/v2/console/economy/currencies", lambda i: i["key"] == "coin", "coin")
    if coin:
        call("PUT", f"/v2/console/economy/currencies/{coin['id']}", coin_body, token)
    else:
        call("POST", "/v2/console/economy/currencies", coin_body, token)

    packs_cat = upsert_category(token, {
        "keyType": "product", "key": "star_coin_packs",
        "name": {"fa": "بسته‌های سکه", "en": "Coin packs", "ar": "حزم العملات"},
        "description": {"fa": "سکه بخر و زودتر سبد تازه بگیر", "en": "Buy coins and unlock new baskets sooner", "ar": "اشترِ العملات واحصل على سلال جديدة أسرع"},
        "iconUrl": "", "metadata": {}, "sort": 1,
    })
    skins_cat = upsert_category(token, {
        "keyType": "product", "key": "star_skins",
        "name": {"fa": "اسکین‌ها", "en": "Skins", "ar": "المظاهر"},
        "description": {"fa": "سبدهای تازه برای بازی", "en": "New baskets for the game", "ar": "سلال جديدة للعبة"},
        "iconUrl": "", "metadata": {}, "sort": 2,
    })

    def coin_prize(key, amount, name):
        return upsert(token, "prizes", {
            "key": key, "name": name,
            "message": {
                "fa": f"{amount} سکه به کیف پولت اضافه شد",
                "en": f"{amount} coins were added to your wallet",
                "ar": f"تمت إضافة {amount} عملة إلى محفظتك",
            },
            "categoryId": None, "tagIds": [], "isRepeatable": True, "metadata": {},
            "items": [{"type": "currency", "key": "coin", "action": "add", "value": amount}],
        })

    def skin_prize(key, skin, name):
        return upsert(token, "prizes", {
            "key": key, "name": name,
            "message": {"fa": f"{name['fa']} مال تو شد", "en": f"The {name['en']} is yours", "ar": f"{name['ar']} أصبحت لك"},
            "categoryId": None, "tagIds": [], "isRepeatable": False, "metadata": {},
            "items": [{"type": "storage", "key": skin, "action": "give", "value": {"skin": skin, "name": name["fa"]}}],
        })

    gold = {"fa": "سبد طلایی", "en": "Golden basket", "ar": "السلة الذهبية"}
    neon = {"fa": "سبد نئونی", "en": "Neon basket", "ar": "سلة النيون"}
    galaxy = {"fa": "سبد کهکشانی", "en": "Galaxy basket", "ar": "سلة المجرة"}

    p100 = coin_prize("star_coins_100", 100, {"fa": "۱۰۰ سکه", "en": "100 coins", "ar": "100 عملة"})
    p500 = coin_prize("star_coins_500", 500, {"fa": "۵۰۰ سکه", "en": "500 coins", "ar": "500 عملة"})
    p1200 = coin_prize("star_coins_1200", 1200, {"fa": "۱۲۰۰ سکه", "en": "1200 coins", "ar": "1200 عملة"})
    s_gold = skin_prize("star_skin_gold", "skin_gold", gold)
    s_neon = skin_prize("star_skin_neon", "skin_neon", neon)
    s_galaxy = skin_prize("star_skin_galaxy", "skin_galaxy", galaxy)

    def pack(key, name, desc, price, prize, sort):
        return upsert(token, "products", {
            "key": key, "name": name, "description": desc, "isActive": True,
            "categoryId": packs_cat["id"], "tagIds": [], "priceToman": price, "coinPrice": None, "discountPercent": 0,
            "paymentMethods": ["zarinpal", "zibal", "bazaar", "myket"],
            "sku": {"bazaar": key, "myket": key}, "isSubscription": False, "prizeId": prize["id"],
            "iconUrl": "", "coverUrl": "", "minLevel": 0, "metadata": {}, "sort": sort,
        })

    def skin(key, name, desc, coins, prize, sort):
        return upsert(token, "products", {
            "key": key, "name": name, "description": desc, "isActive": True,
            "categoryId": skins_cat["id"], "tagIds": [], "priceToman": None,
            "coinPrice": {"currencyKey": "coin", "amount": coins}, "discountPercent": 0,
            "paymentMethods": ["in_app"], "sku": {}, "isSubscription": False, "prizeId": prize["id"],
            "iconUrl": "", "coverUrl": "", "minLevel": 0, "metadata": {}, "sort": sort,
        })

    pack("star_pack_small", {"fa": "بسته کوچک", "en": "Small pack", "ar": "حزمة صغيرة"}, {"fa": "۱۰۰ سکه", "en": "100 coins", "ar": "100 عملة"}, 10000, p100, 1)
    pack("star_pack_medium", {"fa": "بسته متوسط", "en": "Medium pack", "ar": "حزمة متوسطة"}, {"fa": "۵۰۰ سکه", "en": "500 coins", "ar": "500 عملة"}, 40000, p500, 2)
    pack("star_pack_large", {"fa": "بسته بزرگ", "en": "Large pack", "ar": "حزمة كبيرة"}, {"fa": "۱۲۰۰ سکه", "en": "1200 coins", "ar": "1200 عملة"}, 90000, p1200, 3)
    skin("star_skin_gold", gold, {"fa": "سبدی براق از طلا", "en": "A shiny basket made of gold", "ar": "سلة لامعة من الذهب"}, 80, s_gold, 11)
    skin("star_skin_neon", neon, {"fa": "سبدی که در تاریکی می‌درخشد", "en": "A basket that glows in the dark", "ar": "سلة تتوهج في الظلام"}, 200, s_neon, 12)
    skin("star_skin_galaxy", galaxy, {"fa": "سبدی پر از ستاره و رنگ", "en": "A basket full of stars and color", "ar": "سلة مليئة بالنجوم والألوان"}, 400, s_galaxy, 13)

    upsert(token, "gifts", {
        "code": "STAR100", "name": {"fa": "ستاره‌چین: ۱۰۰ سکه هدیه", "en": "Star Catcher: 100 free coins", "ar": "صائد النجوم: 100 عملة هدية"},
        "type": "reward", "prizeId": p100["id"], "discountPercent": None,
        "isPromo": True, "perUserLimit": 1, "maxTotalUses": None, "validFrom": None, "expireAt": None, "isActive": True,
    }, field="code")
    upsert(token, "gifts", {
        "code": "HALF50", "name": {"fa": "ستاره‌چین: ۵۰٪ تخفیف", "en": "Star Catcher: 50% off", "ar": "صائد النجوم: خصم 50٪"},
        "type": "discount", "prizeId": None, "discountPercent": 50,
        "isPromo": True, "perUserLimit": 1, "maxTotalUses": None, "validFrom": None, "expireAt": None, "isActive": True,
    }, field="code")

    res = call("GET", "/v2/console/economy/products?limit=100", token=token)
    for p in res.get("items", []):
        if p["isActive"] and p["key"].split("_")[0] in ("st", "gw", "inapp", "ui"):
            body = {k: v for k, v in p.items() if k not in ("id", "createdAt", "updatedAt")}
            body["isActive"] = False
            call("PUT", f"/v2/console/economy/products/{p['id']}", body, token)
            print("deactivated", p["key"])
    print("ok")


if __name__ == "__main__":
    main()
