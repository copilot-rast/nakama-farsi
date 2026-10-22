// Command jazebe_route_audit reconciles the route inventory with a maintained,
// deliberately conservative machine-readable parity file. It never infers
// implementation parity from a matching RPC name.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type route struct {
	ID                  string   `json:"id"`
	SourceFile          string   `json:"sourceFile"`
	SourceHandler       string   `json:"sourceHandler"`
	ControllerGroup     string   `json:"controllerGroup"`
	ControllerPath      string   `json:"controllerPath"`
	MethodPath          string   `json:"methodPath"`
	Method              string   `json:"method"`
	Path                string   `json:"path"`
	RequestContract     string   `json:"requestContract"`
	ResponseContract    string   `json:"responseContract"`
	ServiceDependencies []string `json:"serviceDependencies"`
	Authorization       string   `json:"authorization"`
	Persistence         []string `json:"persistence"`
	DestinationEntry    string   `json:"destinationEntry"`
	DestinationFile     string   `json:"destinationFile"`
	Implementation      string   `json:"implementationStatus"`
	Differences         string   `json:"behavioralDifferences"`
	Tests               []string `json:"testReferences"`
	RuntimeEvidence     string   `json:"runtimeEvidence"`
	FrontendCoverage    string   `json:"frontendCoverage"`
}

var decorator = regexp.MustCompile(`@(Get|Post|Put|Patch|Delete)\((?:'([^']*)'|"([^"]*)")?\)`)
var controller = regexp.MustCompile(`@Controller\((?:'([^']*)'|"([^"]*)")?\)`)
var handler = regexp.MustCompile(`(?m)^\s*(?:async\s+)?([A-Za-z][A-Za-z0-9_]*)\s*\(`)
var rpcName = regexp.MustCompile("`(jazebe_[a-z0-9_]+)`")
var registration = regexp.MustCompile(`\{"(jazebe_[a-z0-9_]+)",\s*([A-Za-z][A-Za-z0-9_]*)\}`)

func readInventory(path string) ([]route, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result []route
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "| JZ-ROUTE-") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 8 {
			return nil, fmt.Errorf("malformed inventory row: %s", line)
		}
		field := func(i int) string { return strings.TrimSpace(parts[i]) }
		note := field(6)
		state := "UNAUDITED"
		if strings.HasPrefix(note, "PARTIAL:") {
			state = "PARTIAL"
		}
		if strings.HasPrefix(note, "NOT_STARTED:") {
			state = "NOT_STARTED"
		}
		if strings.HasPrefix(note, "BLOCKED:") {
			state = "BLOCKED"
		}
		path := field(3)
		if field(4) != "" {
			path += "/" + field(4)
		}
		result = append(result, route{ID: field(1), Method: field(2), Path: path, ControllerPath: field(3), MethodPath: field(4),
			SourceFile: field(5), ControllerGroup: field(5), Implementation: state, Differences: note,
			ServiceDependencies: []string{}, Persistence: []string{}, Tests: []string{}, FrontendCoverage: "UNAUDITED"})
	}
	return result, nil
}

func sourceHandlers(sourceRoot string, sourceFile string) (map[string][]string, error) {
	data, err := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(sourceFile)))
	if err != nil {
		return nil, err
	}
	text := string(data)
	matches := decorator.FindAllStringSubmatchIndex(text, -1)
	controllers := controller.FindAllStringSubmatchIndex(text, -1)
	result := map[string][]string{}
	for i, match := range matches {
		part := func(a, b int) string {
			if a < 0 || b < 0 {
				return ""
			}
			return text[a:b]
		}
		method := strings.ToUpper(part(match[2], match[3]))
		path := part(match[4], match[5])
		if path == "" {
			path = part(match[6], match[7])
		}
		controllerPath := ""
		for _, c := range controllers {
			if c[0] > match[0] {
				break
			}
			controllerPath = part(c[2], c[3])
			if controllerPath == "" {
				controllerPath = part(c[4], c[5])
			}
		}
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		body := text[match[1]:end]
		found := handler.FindStringSubmatch(body)
		if len(found) == 2 {
			key := controllerPath + ":" + method + ":" + path
			result[key] = append(result[key], found[1])
		}
	}
	return result, nil
}

func destinations(root string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "integrations", "jazebe", "native", "module.go"))
	if err != nil {
		return nil, err
	}
	funcs := map[string]string{}
	for _, match := range registration.FindAllStringSubmatch(string(data), -1) {
		funcs[match[1]] = match[2]
	}
	files, err := filepath.Glob(filepath.Join(root, "integrations", "jazebe", "native", "*.go"))
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for rpc, fn := range funcs {
		needle := "func " + fn + "("
		for _, file := range files {
			data, e := os.ReadFile(file)
			if e != nil {
				return nil, e
			}
			if strings.Contains(string(data), needle) {
				relative, _ := filepath.Rel(root, file)
				result[rpc] = filepath.ToSlash(relative)
				break
			}
		}
	}
	return result, nil
}

func syncRoutes(root, sourceRoot string, inventory []route, existing []route) ([]route, error) {
	old := map[string]route{}
	for _, r := range existing {
		old[r.ID] = r
	}
	dest, err := destinations(root)
	if err != nil {
		return nil, err
	}
	cache := map[string]map[string][]string{}
	positions := map[string]int{}
	for i := range inventory {
		r := &inventory[i]
		if prior, ok := old[r.ID]; ok && prior.SourceFile == r.SourceFile && prior.Method == r.Method && prior.Path == r.Path {
			keep := *r
			*r = prior
			r.Implementation = keep.Implementation
			r.Differences = keep.Differences
		}
		r.SourceHandler = ""
		if _, ok := cache[r.SourceFile]; !ok {
			found, e := sourceHandlers(sourceRoot, r.SourceFile)
			if e != nil {
				return nil, fmt.Errorf("%s: %w", r.SourceFile, e)
			}
			cache[r.SourceFile] = found
		}
		if r.SourceHandler == "" {
			// Match by occurrence order within this controller for duplicate routes.
			key := r.SourceFile + ":" + r.Method + ":" + r.Path
			matches := cache[r.SourceFile][r.ControllerPath+":"+r.Method+":"+r.MethodPath]
			if positions[key] < len(matches) {
				r.SourceHandler = matches[positions[key]]
			}
			positions[key]++
		}
		if r.DestinationEntry == "" {
			if m := rpcName.FindStringSubmatch(r.Differences); len(m) == 2 {
				r.DestinationEntry = m[1]
			}
		}
		if r.DestinationEntry != "" && r.DestinationFile == "" {
			r.DestinationFile = dest[r.DestinationEntry]
		}
	}
	return inventory, nil
}

func validate(root, sourceRoot string, inventory, mapped []route) error {
	if len(inventory) != len(mapped) {
		return fmt.Errorf("inventory has %d routes, mapping has %d", len(inventory), len(mapped))
	}
	known := map[string]route{}
	for _, r := range inventory {
		if _, ok := known[r.ID]; ok {
			return fmt.Errorf("duplicate inventory ID %s", r.ID)
		}
		known[r.ID] = r
	}
	seen := map[string]bool{}
	sourceCache := map[string]map[string][]string{}
	for _, r := range mapped {
		if seen[r.ID] {
			return fmt.Errorf("duplicate mapping ID %s", r.ID)
		}
		seen[r.ID] = true
		base, ok := known[r.ID]
		if !ok {
			return fmt.Errorf("mapping ID %s absent from inventory", r.ID)
		}
		if r.SourceFile != base.SourceFile || r.Method != base.Method || r.Path != base.Path {
			return fmt.Errorf("source route changed for %s", r.ID)
		}
		if r.SourceHandler == "" {
			return fmt.Errorf("%s missing source handler", r.ID)
		}
		if sourceRoot != "" {
			if _, err := os.Stat(filepath.Join(sourceRoot, filepath.FromSlash(r.SourceFile))); err != nil {
				return fmt.Errorf("%s source: %w", r.ID, err)
			}
			if _, ok := sourceCache[r.SourceFile]; !ok {
				found, err := sourceHandlers(sourceRoot, r.SourceFile)
				if err != nil {
					return fmt.Errorf("%s source handlers: %w", r.ID, err)
				}
				sourceCache[r.SourceFile] = found
			}
			key := r.ControllerPath + ":" + r.Method + ":" + r.MethodPath
			matched := false
			for _, name := range sourceCache[r.SourceFile][key] {
				if name == r.SourceHandler {
					matched = true
					break
				}
			}
			if !matched {
				return fmt.Errorf("%s source controller/path/handler no longer matches %s", r.ID, key)
			}
		}
		switch r.Implementation {
		case "UNAUDITED", "NOT_STARTED", "PARTIAL", "IMPLEMENTED", "EQUIVALENT_NATIVE", "BLOCKED", "SOURCE_ONLY_NOT_APPLICABLE":
		default:
			return fmt.Errorf("%s invalid status %q", r.ID, r.Implementation)
		}
		if r.Implementation == "PARTIAL" || r.Implementation == "IMPLEMENTED" || r.Implementation == "EQUIVALENT_NATIVE" {
			if r.DestinationEntry == "" || r.DestinationFile == "" {
				return fmt.Errorf("%s mapped without destination evidence", r.ID)
			}
		}
		if r.DestinationFile != "" {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(r.DestinationFile))); err != nil {
				return fmt.Errorf("%s destination: %w", r.ID, err)
			}
		}
		for _, ref := range r.Tests {
			path := strings.TrimSuffix(ref, " (NOT_RUN)")
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
				return fmt.Errorf("%s test reference %q: %w", r.ID, ref, err)
			}
		}
		if r.Implementation == "IMPLEMENTED" || r.Implementation == "EQUIVALENT_NATIVE" {
			if r.SourceHandler == "" || r.RequestContract == "" || r.ResponseContract == "" || r.Authorization == "" || len(r.Tests) == 0 {
				return fmt.Errorf("%s claimed implemented without contract/test evidence", r.ID)
			}
		}
	}
	for id := range known {
		if !seen[id] {
			return fmt.Errorf("missing mapping for %s", id)
		}
	}
	return nil
}

func main() {
	syncFlag := flag.Bool("sync", false, "refresh inventory-linked fields while retaining manual contract evidence")
	offlineFlag := flag.Bool("offline", false, "validate mapping without the sibling read-only Jazebe source tree")
	flag.Parse()
	if *syncFlag && *offlineFlag {
		panic("-sync requires the source tree")
	}
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sourceRoot := filepath.Join(filepath.Dir(root), "jazebe")
	if *offlineFlag {
		sourceRoot = ""
	}
	inventoryPath := filepath.Join(root, "docs", "integration", "jazebe-http-routes.md")
	mappingPath := filepath.Join(root, "docs", "integration", "jazebe-route-parity.json")
	inventory, err := readInventory(inventoryPath)
	if err != nil {
		panic(err)
	}
	var mapped []route
	if data, e := os.ReadFile(mappingPath); e == nil {
		if err = json.Unmarshal(data, &mapped); err != nil {
			panic(err)
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		panic(e)
	}
	if *syncFlag {
		mapped, err = syncRoutes(root, sourceRoot, inventory, mapped)
		if err != nil {
			panic(err)
		}
		data, e := json.MarshalIndent(mapped, "", "  ")
		if e != nil {
			panic(e)
		}
		if e = os.WriteFile(mappingPath, append(data, '\n'), 0644); e != nil {
			panic(e)
		}
	}
	if err = validate(root, sourceRoot, inventory, mapped); err != nil {
		panic(err)
	}
	counts := map[string]int{}
	missingHandlers := 0
	missingIDs := make([]string, 0)
	for _, r := range mapped {
		counts[r.Implementation]++
		if r.SourceHandler == "" {
			missingHandlers++
			missingIDs = append(missingIDs, r.ID)
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sourceValidation := "FULL"
	if *offlineFlag {
		sourceValidation = "SKIPPED"
	}
	fmt.Printf("routes=%d source_validation=%s source_handlers_missing=%d", len(mapped), sourceValidation, missingHandlers)
	for _, k := range keys {
		fmt.Printf(" %s=%d", k, counts[k])
	}
	fmt.Println()
	if len(missingIDs) > 0 {
		fmt.Println("source_handler_missing_ids=" + strings.Join(missingIDs, ","))
	}
}
