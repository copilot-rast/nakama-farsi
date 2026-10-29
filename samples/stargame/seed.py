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


def find(token, resource, match, q):
    res = call("GET", f"/v2/console/economy/{resource}?limit=100&q=" + urllib.parse.quote(q), token=token)
    for item in res.get("items", []):
        if match(item):
            return item
    return None


def upsert(token, resource, body, field="key"):
    value = body[field]
    existing = find(token, resource, lambda i: i.get(field) == value, value)
    if existing:
        return call("PUT", f"/v2/console/economy/{resource}/{existing['id']}", body, token)
    return call("POST", f"/v2/console/economy/{resource}", body, token)


def main():
    token = login()

    coin = find(token, "currencies", lambda i: i["key"] == "coin", "coin")
    coin_body = {"key": "coin", "name": {"fa": "سکه", "en": "Coin"}, "type": "soft", "defaultValue": 50, "iconUrl": "", "metadata": {}, "sort": 1}
    if coin:
        call("PUT", f"/v2/console/economy/currencies/{coin['id']}", coin_body, token)
    else:
        call("POST", "/v2/console/economy/currencies", coin_body, token)

    packs_cat = upsert(token, "categories", {"kind": "product", "key": "star_coin_packs", "name": {"fa": "بسته‌های سکه", "en": "Coin packs"}, "sort": 1})
    skins_cat = upsert(token, "categories", {"kind": "product", "key": "star_skins", "name": {"fa": "اسکین‌ها", "en": "Skins"}, "sort": 2})

    def coin_prize(key, amount, fa):
        return upsert(token, "prizes", {
            "key": key, "name": {"fa": fa}, "message": {"fa": f"{amount} سکه به کیف پولت اضافه شد"},
            "categoryId": None, "tagIds": [], "isRepeatable": True, "metadata": {},
            "items": [{"type": "currency", "key": "coin", "action": "add", "value": amount}],
        })

    def skin_prize(key, skin, fa):
        return upsert(token, "prizes", {
            "key": key, "name": {"fa": fa}, "message": {"fa": f"{fa} مال تو شد"},
            "categoryId": None, "tagIds": [], "isRepeatable": False, "metadata": {},
            "items": [{"type": "storage", "key": skin, "action": "give", "value": {"skin": skin, "name": fa}}],
        })

    p100 = coin_prize("star_coins_100", 100, "۱۰۰ سکه")
    p500 = coin_prize("star_coins_500", 500, "۵۰۰ سکه")
    p1200 = coin_prize("star_coins_1200", 1200, "۱۲۰۰ سکه")
    s_gold = skin_prize("star_skin_gold", "skin_gold", "سبد طلایی")
    s_neon = skin_prize("star_skin_neon", "skin_neon", "سبد نئونی")
    s_galaxy = skin_prize("star_skin_galaxy", "skin_galaxy", "سبد کهکشانی")

    def pack(key, fa, desc, price, prize, sort):
        return upsert(token, "products", {
            "key": key, "name": {"fa": fa}, "description": {"fa": desc}, "isActive": True,
            "categoryId": packs_cat["id"], "tagIds": [], "priceToman": price, "coinPrice": None, "discountPercent": 0,
            "paymentMethods": ["zarinpal", "zibal", "bazaar", "myket"],
            "sku": {"bazaar": key, "myket": key}, "isSubscription": False, "prizeId": prize["id"],
            "iconUrl": "", "coverUrl": "", "minLevel": 0, "metadata": {}, "sort": sort,
        })

    def skin(key, fa, desc, coins, prize, sort):
        return upsert(token, "products", {
            "key": key, "name": {"fa": fa}, "description": {"fa": desc}, "isActive": True,
            "categoryId": skins_cat["id"], "tagIds": [], "priceToman": None,
            "coinPrice": {"currencyKey": "coin", "amount": coins}, "discountPercent": 0,
            "paymentMethods": ["in_app"], "sku": {}, "isSubscription": False, "prizeId": prize["id"],
            "iconUrl": "", "coverUrl": "", "minLevel": 0, "metadata": {}, "sort": sort,
        })

    pack("star_pack_small", "بسته کوچک", "۱۰۰ سکه", 10000, p100, 1)
    pack("star_pack_medium", "بسته متوسط", "۵۰۰ سکه", 40000, p500, 2)
    pack("star_pack_large", "بسته بزرگ", "۱۲۰۰ سکه", 90000, p1200, 3)
    skin("star_skin_gold", "سبد طلایی", "سبدی براق از طلا", 80, s_gold, 11)
    skin("star_skin_neon", "سبد نئونی", "سبدی که در تاریکی می‌درخشد", 200, s_neon, 12)
    skin("star_skin_galaxy", "سبد کهکشانی", "سبدی پر از ستاره و رنگ", 400, s_galaxy, 13)

    upsert(token, "gifts", {
        "code": "STAR100", "name": "ستاره‌چین: ۱۰۰ سکه هدیه", "type": "reward", "prizeId": p100["id"], "discountPercent": None,
        "isPromo": True, "perUserLimit": 1, "maxTotalUses": None, "validFrom": None, "expireAt": None, "isActive": True,
    }, field="code")
    upsert(token, "gifts", {
        "code": "HALF50", "name": "ستاره‌چین: ۵۰٪ تخفیف", "type": "discount", "prizeId": None, "discountPercent": 50,
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
