package main

import (
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"
)

const consolePrefix = "/v2/console/base"

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	routes := []route{
		{consolePrefix + "/languages", http.MethodGet, s.languageList},
		{consolePrefix + "/languages", http.MethodPost, s.languageWrite(true)},
		{consolePrefix + "/languages/coverage", http.MethodGet, s.languageCoverage},
		{consolePrefix + "/languages/{id}", http.MethodGet, s.languageGet},
		{consolePrefix + "/languages/{id}", http.MethodPut, s.languageWrite(false)},
		{consolePrefix + "/languages/{id}", http.MethodDelete, s.languageDelete},
		{consolePrefix + "/categories/key-types", http.MethodGet, s.keyTypes(categoryKeyTypes)},
		{consolePrefix + "/tags/key-types", http.MethodGet, s.keyTypes(tagKeyTypes)},
		{consolePrefix + "/brand", http.MethodGet, s.brandGet},
		{consolePrefix + "/brand", http.MethodPut, s.brandPut},
		{consolePrefix + "/brand/public", http.MethodGet, s.brandPublic},
		{consolePrefix + "/setup", http.MethodGet, s.setupGet},
		{consolePrefix + "/setup", http.MethodPut, s.setupPut},
		{consolePrefix + "/setup/claim", http.MethodPost, s.setupClaim},
		{consolePrefix + "/account-rules", http.MethodGet, s.rulesGet},
		{consolePrefix + "/account-rules", http.MethodPut, s.rulesPut},
		{consolePrefix + "/app-versions", http.MethodGet, s.appVersionList},
		{consolePrefix + "/app-versions", http.MethodPost, s.appVersionWrite(true)},
		{consolePrefix + "/app-versions/catalog", http.MethodGet, s.appVersionCatalog},
		{consolePrefix + "/app-versions/{id}", http.MethodGet, s.appVersionGet},
		{consolePrefix + "/app-versions/{id}", http.MethodPut, s.appVersionWrite(false)},
		{consolePrefix + "/app-versions/{id}", http.MethodDelete, s.appVersionDelete},
		{consolePrefix + "/countries", http.MethodGet, s.countryList},
		{consolePrefix + "/countries", http.MethodPost, s.countryWrite(true)},
		{consolePrefix + "/countries/{id}", http.MethodGet, s.countryGet},
		{consolePrefix + "/countries/{id}", http.MethodPut, s.countryWrite(false)},
		{consolePrefix + "/countries/{id}", http.MethodDelete, s.countryDelete},
		{consolePrefix + "/players", http.MethodGet, s.playerList},
		{consolePrefix + "/players/tags", http.MethodPost, s.playerTagsBulk},
		{consolePrefix + "/players/{id}", http.MethodGet, s.playerGet},
		{consolePrefix + "/players/{id}/tags", http.MethodPut, s.playerTagsPut},
		{consolePrefix + "/players/{id}/restrictions", http.MethodPost, s.restrictionCreate},
		{consolePrefix + "/restrictions", http.MethodGet, s.restrictionList},
		{consolePrefix + "/restrictions/{id}", http.MethodGet, s.restrictionGet},
		{consolePrefix + "/restrictions/{id}/lift", http.MethodPost, s.restrictionLift},
		{consolePrefix + "/chat-words", http.MethodGet, s.wordList},
		{consolePrefix + "/chat-words", http.MethodPost, s.wordWrite(true)},
		{consolePrefix + "/chat-words/test", http.MethodPost, s.wordTest},
		{consolePrefix + "/chat-words/{id}", http.MethodGet, s.wordGet},
		{consolePrefix + "/chat-words/{id}", http.MethodPut, s.wordWrite(false)},
		{consolePrefix + "/chat-words/{id}", http.MethodDelete, s.wordDelete},
		{consolePrefix + "/report-titles", http.MethodGet, s.reportTitleList},
		{consolePrefix + "/report-titles", http.MethodPost, s.reportTitleWrite(true)},
		{consolePrefix + "/report-titles/{id}", http.MethodGet, s.reportTitleGet},
		{consolePrefix + "/report-titles/{id}", http.MethodPut, s.reportTitleWrite(false)},
		{consolePrefix + "/report-titles/{id}", http.MethodDelete, s.reportTitleDelete},
		{consolePrefix + "/reports", http.MethodGet, s.reportList},
		{consolePrefix + "/reports/summary", http.MethodGet, s.reportSummary},
		{consolePrefix + "/reports/{id}", http.MethodGet, s.reportGet},
		{consolePrefix + "/reports/{id}/resolve", http.MethodPost, s.reportResolve},
		{consolePrefix + "/support/summary", http.MethodGet, s.supportSummary},
		{consolePrefix + "/support/heartbeat", http.MethodPost, s.supportHeartbeat},
		{consolePrefix + "/support/agents", http.MethodGet, s.supportAgents},
		{consolePrefix + "/support/players", http.MethodGet, s.supportPlayers},
		{consolePrefix + "/support/settings", http.MethodGet, s.supportSettingsGet},
		{consolePrefix + "/support/settings", http.MethodPut, s.supportSettingsPut},
		{consolePrefix + "/support/departments", http.MethodGet, s.departmentList},
		{consolePrefix + "/support/departments", http.MethodPost, s.departmentWrite(true)},
		{consolePrefix + "/support/departments/{id}", http.MethodGet, s.departmentGet},
		{consolePrefix + "/support/departments/{id}", http.MethodPut, s.departmentWrite(false)},
		{consolePrefix + "/support/departments/{id}", http.MethodDelete, s.departmentDelete},
		{consolePrefix + "/support/canned", http.MethodGet, s.cannedList},
		{consolePrefix + "/support/canned", http.MethodPost, s.cannedWrite(true)},
		{consolePrefix + "/support/canned/{id}", http.MethodGet, s.cannedGet},
		{consolePrefix + "/support/canned/{id}", http.MethodPut, s.cannedWrite(false)},
		{consolePrefix + "/support/canned/{id}", http.MethodDelete, s.cannedDelete},
		{consolePrefix + "/support/uploads", http.MethodPost, s.consoleSupportUpload},
		{consolePrefix + "/support/tickets", http.MethodGet, s.supportTicketList},
		{consolePrefix + "/support/tickets", http.MethodPost, s.supportTicketCreate},
		{consolePrefix + "/support/tickets/{id}", http.MethodGet, s.supportTicketGet},
		{consolePrefix + "/support/tickets/{id}/messages", http.MethodPost, s.supportTicketMessage},
		{consolePrefix + "/support/tickets/{id}/status", http.MethodPost, s.supportTicketStatus},
		{consolePrefix + "/support/tickets/{id}/assign", http.MethodPost, s.supportTicketAssign},
		{consolePrefix + "/support/chats", http.MethodGet, s.supportChatList},
		{consolePrefix + "/support/chats/{id}", http.MethodGet, s.supportChatGet},
		{consolePrefix + "/support/chats/{id}/take", http.MethodPost, s.supportChatTake},
		{consolePrefix + "/support/chats/{id}/messages", http.MethodPost, s.supportChatMessage},
		{consolePrefix + "/support/chats/{id}/typing", http.MethodPost, s.supportChatTyping},
		{consolePrefix + "/support/chats/{id}/read", http.MethodPost, s.supportChatRead},
		{consolePrefix + "/support/chats/{id}/transfer", http.MethodPost, s.supportChatTransfer},
		{consolePrefix + "/support/chats/{id}/close", http.MethodPost, s.supportChatClose},
		{consolePrefix + "/remote-config", http.MethodGet, s.configList},
		{consolePrefix + "/remote-config", http.MethodPost, s.configWrite(true)},
		{consolePrefix + "/remote-config/history", http.MethodGet, s.configRecentHistory},
		{consolePrefix + "/remote-config/preview", http.MethodPost, s.configPreview},
		{consolePrefix + "/remote-config/{id}", http.MethodGet, s.configGet},
		{consolePrefix + "/remote-config/{id}", http.MethodPut, s.configWrite(false)},
		{consolePrefix + "/remote-config/{id}", http.MethodDelete, s.configDelete},
		{consolePrefix + "/remote-config/{id}/history", http.MethodGet, s.configHistory},
		{consolePrefix + "/remote-config/{id}/rollback", http.MethodPost, s.configRollback},
		{consolePrefix + "/activity", http.MethodGet, s.activityList},
		{consolePrefix + "/activity/summary", http.MethodGet, s.activitySummary},
	}
	for _, res := range []resource{categoriesResource, tagsResource} {
		base := consolePrefix + "/" + res.name
		routes = append(routes,
			route{base, http.MethodGet, s.taxonomyList(res)},
			route{base, http.MethodPost, s.taxonomyWrite(res, true)},
			route{base + "/{id}", http.MethodGet, s.taxonomyGet(res)},
			route{base + "/{id}", http.MethodPut, s.taxonomyWrite(res, false)},
			route{base + "/{id}", http.MethodDelete, s.taxonomyDelete(res)},
		)
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}
