select 'pm', method, is_active, md5(config::text), updated_at from economy_payment_method order by 2;
select 'pv', channel, kind, is_active, is_default, sandbox, md5(config::text), updated_at from publication_provider order by 2,3;
select 'cur', key, type, name::text, default_value, sort, icon_url from economy_currency order by sort, key;
select 'lang', code, is_active, is_default, sort from base_language order by code;
select 'H-pm', md5(coalesce(string_agg(method||is_active::text||config::text, '|' order by method),'')) from economy_payment_method;
select 'H-pv', md5(coalesce(string_agg(id::text||is_active::text||is_default::text||sandbox::text||config::text, '|' order by id),'')) from publication_provider;
select 'H-lang', md5(coalesce(string_agg(code||name::text||is_active::text||is_default::text||sort::text, '|' order by code),'')) from base_language;
select 'H-cur', md5(coalesce(string_agg(id::text||key||name::text||default_value::text||icon_url, '|' order by id),'')) from economy_currency;
