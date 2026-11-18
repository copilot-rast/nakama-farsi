package main

import "net/http"

type knownLanguage struct {
	Code       string            `json:"code"`
	NativeName string            `json:"nativeName"`
	Direction  string            `json:"direction"`
	Name       map[string]string `json:"name"`
}

var knownLanguages = []knownLanguage{
	{"fa", "فارسی", "rtl", map[string]string{"fa": "فارسی", "en": "Persian"}},
	{"en", "English", "ltr", map[string]string{"fa": "انگلیسی", "en": "English"}},
	{"ar", "العربية", "rtl", map[string]string{"fa": "عربی", "en": "Arabic"}},
	{"tr", "Türkçe", "ltr", map[string]string{"fa": "ترکی استانبولی", "en": "Turkish"}},
	{"az", "Azərbaycanca", "ltr", map[string]string{"fa": "آذربایجانی", "en": "Azerbaijani"}},
	{"ku", "Kurdî", "ltr", map[string]string{"fa": "کردی", "en": "Kurdish"}},
	{"ur", "اردو", "rtl", map[string]string{"fa": "اردو", "en": "Urdu"}},
	{"ps", "پښتو", "rtl", map[string]string{"fa": "پشتو", "en": "Pashto"}},
	{"ru", "Русский", "ltr", map[string]string{"fa": "روسی", "en": "Russian"}},
	{"de", "Deutsch", "ltr", map[string]string{"fa": "آلمانی", "en": "German"}},
	{"fr", "Français", "ltr", map[string]string{"fa": "فرانسوی", "en": "French"}},
	{"es", "Español", "ltr", map[string]string{"fa": "اسپانیایی", "en": "Spanish"}},
	{"it", "Italiano", "ltr", map[string]string{"fa": "ایتالیایی", "en": "Italian"}},
	{"pt", "Português", "ltr", map[string]string{"fa": "پرتغالی", "en": "Portuguese"}},
	{"hi", "हिन्दी", "ltr", map[string]string{"fa": "هندی", "en": "Hindi"}},
	{"zh", "中文", "ltr", map[string]string{"fa": "چینی", "en": "Chinese"}},
	{"ja", "日本語", "ltr", map[string]string{"fa": "ژاپنی", "en": "Japanese"}},
	{"ko", "한국어", "ltr", map[string]string{"fa": "کره‌ای", "en": "Korean"}},
}

func (s *service) languageKnown(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": knownLanguages})
}
