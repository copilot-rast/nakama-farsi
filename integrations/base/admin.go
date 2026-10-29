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
