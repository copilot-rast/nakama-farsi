package main

import (
	"regexp"
	"strconv"
	"strings"
)

var jsonPathSegment = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_\-]*)?((\[[0-9]{1,4}\])*)$`)
var jsonPathIndex = regexp.MustCompile(`\[([0-9]{1,4})\]`)

type pathStep struct {
	key   string
	index int
}

func parseJSONPath(path string) ([]pathStep, bool) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return nil, true
	}
	path = strings.TrimPrefix(strings.TrimPrefix(path, "$"), ".")
	if len(path) > 200 {
		return nil, false
	}
	var out []pathStep
	for _, seg := range strings.Split(path, ".") {
		m := jsonPathSegment.FindStringSubmatch(seg)
		if m == nil || (m[1] == "" && m[2] == "") {
			return nil, false
		}
		if m[1] != "" {
			out = append(out, pathStep{key: m[1], index: -1})
		}
		for _, ix := range jsonPathIndex.FindAllStringSubmatch(m[2], -1) {
			n, _ := strconv.Atoi(ix[1])
			out = append(out, pathStep{index: n})
		}
	}
	return out, true
}

func validJSONPath(path string) bool {
	_, ok := parseJSONPath(path)
	return ok
}

func lookupJSONPath(v any, path string) (any, bool) {
	steps, ok := parseJSONPath(path)
	if !ok {
		return nil, false
	}
	cur := v
	for _, st := range steps {
		if st.index >= 0 && st.key == "" {
			arr, ok := cur.([]any)
			if !ok || st.index >= len(arr) {
				return nil, false
			}
			cur = arr[st.index]
			continue
		}
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		if cur, ok = obj[st.key]; !ok {
			return nil, false
		}
	}
	return cur, cur != nil
}
