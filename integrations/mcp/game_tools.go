package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type gamePlanetInput struct {
	PlanetID string `json:"planetId" jsonschema:"Planet UUID"`
}

type gameReportInput struct {
	PlanetID string `json:"planetId" jsonschema:"Administered planet UUID"`
	GameID   string `json:"gameId" jsonschema:"Game UUID"`
}

type gameBrandListInput struct {
	PlanetID string `json:"planetId"`
	Status   string `json:"status,omitempty"`
	Search   string `json:"search,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

func addGameTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_game_list", Description: "List the authenticated player's active games and current play quotas on a planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gamePlanetInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_game_list", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_game_report", Description: "Read session counts, paid plays, tickets and player totals for a game administered by the caller's brand."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gameReportInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_game_report", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_game_catalog", Description: "Browse the platform game templates with availability for an administered planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gamePlanetInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_brand_game_catalog", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_game_meta", Description: "Read point name and selectable lotteries for configuring games on an administered planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gamePlanetInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_game_meta", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_game_list", Description: "List configured planet games and their aggregated play statistics for an authorized brand administrator."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gameBrandListInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_game_list", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_brand_game_get", Description: "Read one configured planet game with its status, prize settings and play statistics for an authorized brand administrator."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in gameReportInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_brand_game_get", in)
			return nil, out, err
		})
}
