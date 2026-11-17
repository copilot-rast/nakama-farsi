package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/heroiclabs/nakama/v3/tools/rpcschema/gen"
)

func main() {
	root := flag.String("root", ".", "repository root")
	out := flag.String("out", "", "folder for rpc-schema.json and codes.json")
	openapi := flag.String("openapi", "", "write the OpenAPI document to this file")
	extra := flag.String("modules", "", "modules as name=dir,name=dir; empty means every plugin found under integrations/ and samples/")
	flag.Func("buildflag", "extra go build flag for type loading, repeatable (match the plugin builds to reuse the build cache)", func(v string) error {
		gen.BuildFlags = append(gen.BuildFlags, v)
		return nil
	})
	flag.Parse()
	modules, err := gen.Discover(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if strings.TrimSpace(*extra) != "" {
		modules = append([]gen.Module{}, gen.DefaultModules...)
		for _, part := range strings.Split(*extra, ",") {
			name, dir, ok := strings.Cut(strings.TrimSpace(part), "=")
			if ok && name != "" && dir != "" {
				modules = append(modules, gen.Module{Name: name, Dir: dir})
			}
		}
	}
	if *out != "" {
		if err := gen.Write(*root, *out, modules); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if *openapi != "" {
		if err := gen.WriteOpenAPI(*root, *openapi, modules); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
