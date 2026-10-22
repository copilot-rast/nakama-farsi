package main

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type referralStatsInput struct {
	PlanetID string `json:"planetId" jsonschema:"Planet UUID whose referral counts the authenticated user owns"`
}

func addReferralTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_referral_stats", Description: "Read your own issued, registered and qualified mission invitation counts on a planet."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in referralStatsInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_referral_stats", in)
			return nil, out, err
		})
}
