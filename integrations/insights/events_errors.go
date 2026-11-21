package main

import "github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"

func errorDefaults() map[string]apierr.Default {
	return map[string]apierr.Default{
		"ANALYTICS_BACKLOG_FULL": {Status: 503, RetryAfter: 30, Messages: map[string]string{
			"en": "Analytics is catching up; keep the events and send them again later.",
			"fa": "آنالیتیکس پشت سر هم کار دارد؛ رویدادها را نگه دارید و بعداً دوباره بفرستید."}},
		"ANALYTICS_DISABLED": {Status: 503, Messages: map[string]string{
			"en": "Analytics events are turned off on this server.",
			"fa": "ثبت رویدادهای آنالیتیکس روی این سرور خاموش است."}},
	}
}
