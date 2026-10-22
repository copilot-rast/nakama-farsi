package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type medalPlanetInput struct {
	PlanetID string `json:"planetId" jsonschema:"UUID of a planet administered by the authenticated brand economy role"`
}

type medalBackfillInput struct {
	PlanetID string `json:"planetId"`
	ID       string `json:"id" jsonschema:"Planet medal UUID"`
}

func addMedalTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_medal_wall", Description: "Read your own planet medal wall; earned medals remain visible after they are paused."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in medalPlanetInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_medal_wall", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_medals", Description: "List selected medals, conditions, rewards and acquisition counts for an administered planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in medalPlanetInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_brand_medals", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_medal_meta", Description: "Read the administered planet's medal reward limits and point name."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in medalPlanetInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_medal_meta", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_medal_library", Description: "Browse active operator medal definitions and see which are selected by an administered planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in medalPlanetInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_brand_medal_library", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_medal_backfill_status", Description: "Read durable catch-up progress and last error for a medal in an administered planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in medalBackfillInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_medal_backfill_status", in)
			return nil, out, err
		})
}
