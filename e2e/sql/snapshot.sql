select 'H-pm', md5(coalesce(string_agg(method||is_active::text||config::text, '|' order by method),'')) from economy_payment_method;
select 'H-pv', md5(coalesce(string_agg(id::text||kind||is_active::text||is_default::text||sandbox::text||config::text, '|' order by id),'')) from publication_provider;
select 'H-lang', md5(coalesce(string_agg(code||name::text||native_name||direction||icon_url||is_active::text||is_default::text||sort::text, '|' order by code),'')) from base_language;
select 'H-cur', md5(coalesce(string_agg(id::text||key||name::text||type||default_value::text||icon_url||metadata::text||sort::text, '|' order by id),'')) from economy_currency;
select 'H-brand', md5(coalesce(string_agg(name::text||short_name||logo_url||logo_dark_url||mark_url||accent||welcome::text, '|'),'')) from base_brand;
select 'H-setup', md5(coalesce(string_agg(status||steps::text||first_opened_by||completed_by||coalesce(completed_at::text,''), '|'),'')) from base_setup;
