package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type profileOverviewInput struct {
	PlanetID string `json:"planetId,omitempty" jsonschema:"Optional planet UUID; defaults to your latest active membership"`
}

type profilePublicInput struct {
	PlanetID string `json:"planetId" jsonschema:"Planet UUID shared by the caller and target"`
	UserID   string `json:"userId" jsonschema:"Target Nakama user UUID"`
}

func addProfileTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_profile_overview", Description: "Read your own Jazebe profile, planet XP and medal wall, and verification status."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in profileOverviewInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_profile_overview", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_profile_public", Description: "Read another active member's public profile on a planet you both belong to; private fields are excluded."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in profilePublicInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_profile_public", in)
			return nil, out, err
		})
}
