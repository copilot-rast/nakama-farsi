package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type lotteryPrizeToolInput struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type" jsonschema:"cash, brand_credit, planet_points, or discount_coupon"`
	ValueRial   string `json:"valueRial" jsonschema:"Nonnegative whole rial amount"`
	Priority    int    `json:"priority" jsonschema:"Unique priority from 1 to 50"`
	ImageAsset  string `json:"imageAsset,omitempty"`
}
type lotteryDraftToolInput struct {
	PlanetID          string                  `json:"planetId"`
	Title             string                  `json:"title"`
	ShortDescription  string                  `json:"shortDescription,omitempty"`
	LongDescription   string                  `json:"longDescription,omitempty"`
	BannerAsset       string                  `json:"bannerAsset,omitempty"`
	StartsAt          string                  `json:"startsAt" jsonschema:"RFC3339 start time"`
	DrawAt            string                  `json:"drawAt" jsonschema:"RFC3339 future draw time"`
	TicketPricePoints int                     `json:"ticketPricePoints"`
	PerUserTicketCap  *int                    `json:"perUserTicketCap,omitempty"`
	PerGroupTicketCap *int                    `json:"perGroupTicketCap,omitempty"`
	Prizes            []lotteryPrizeToolInput `json:"prizes"`
}
type lotteryRefToolInput struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
}
type lotteryClientInput struct {
	PlanetID string `json:"planetId"`
}
type lotteryFinishedToolInput struct {
	PlanetID string `json:"planetId"`
	Range    string `json:"range,omitempty" jsonschema:"all, m1, m3, or y"`
}
type lotteryCancelToolInput struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
	Reason    string `json:"reason"`
}
type lotteryListToolInput struct {
	PlanetID string `json:"planetId"`
	Status   string `json:"status,omitempty"`
	Search   string `json:"search,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}
type lotteryUpdateToolInput struct {
	PlanetID               string                   `json:"planetId"`
	LotteryID              string                   `json:"lotteryId"`
	Title                  *string                  `json:"title,omitempty"`
	ShortDescription       *string                  `json:"shortDescription,omitempty"`
	LongDescription        *string                  `json:"longDescription,omitempty"`
	BannerAsset            *string                  `json:"bannerAsset,omitempty"`
	ClearShortDescription  bool                     `json:"clearShortDescription,omitempty"`
	ClearLongDescription   bool                     `json:"clearLongDescription,omitempty"`
	ClearBannerAsset       bool                     `json:"clearBannerAsset,omitempty"`
	StartsAt               *string                  `json:"startsAt,omitempty"`
	DrawAt                 *string                  `json:"drawAt,omitempty"`
	TicketPricePoints      *int                     `json:"ticketPricePoints,omitempty"`
	PerUserTicketCap       *int                     `json:"perUserTicketCap,omitempty"`
	PerGroupTicketCap      *int                     `json:"perGroupTicketCap,omitempty"`
	ClearPerUserTicketCap  bool                     `json:"clearPerUserTicketCap,omitempty"`
	ClearPerGroupTicketCap bool                     `json:"clearPerGroupTicketCap,omitempty"`
	Prizes                 *[]lotteryPrizeToolInput `json:"prizes,omitempty"`
}

func lotteryUpdatePayload(in lotteryUpdateToolInput) map[string]any {
	out := map[string]any{"planetId": in.PlanetID, "lotteryId": in.LotteryID}
	if in.Title != nil {
		out["title"] = *in.Title
	}
	if in.ShortDescription != nil {
		out["shortDescription"] = *in.ShortDescription
	}
	if in.ClearShortDescription {
		out["shortDescription"] = nil
	}
	if in.LongDescription != nil {
		out["longDescription"] = *in.LongDescription
	}
	if in.ClearLongDescription {
		out["longDescription"] = nil
	}
	if in.BannerAsset != nil {
		out["bannerAsset"] = *in.BannerAsset
	}
	if in.ClearBannerAsset {
		out["bannerAsset"] = nil
	}
	if in.StartsAt != nil {
		out["startsAt"] = *in.StartsAt
	}
	if in.DrawAt != nil {
		out["drawAt"] = *in.DrawAt
	}
	if in.TicketPricePoints != nil {
		out["ticketPricePoints"] = *in.TicketPricePoints
	}
	if in.PerUserTicketCap != nil {
		out["perUserTicketCap"] = *in.PerUserTicketCap
	}
	if in.ClearPerUserTicketCap {
		out["perUserTicketCap"] = nil
	}
	if in.PerGroupTicketCap != nil {
		out["perGroupTicketCap"] = *in.PerGroupTicketCap
	}
	if in.ClearPerGroupTicketCap {
		out["perGroupTicketCap"] = nil
	}
	if in.Prizes != nil {
		out["prizes"] = *in.Prizes
	}
	return out
}

func addLotteryTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_meta", Description: "Read the administered planet's point name, ticket price limits and permitted lottery prize types."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryClientInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_meta", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_client_active", Description: "List active and drawing raffles on an active planet, with prize summary and the caller's group ticket count."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryClientInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_lottery_client_active", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_client_finished", Description: "List finished and cancelled raffles on an active planet, with Persian dates, winning groups and the caller's winning group."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryFinishedToolInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_lottery_client_finished", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_client_detail", Description: "Read the full raffle lifecycle, prize pool, participating groups and, after drawing, winning tickets and the caller's share."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryRefToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_client_detail", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_draft_create", Description: "Create a draft lottery with prizes, schedule and ticket caps as an active brand administrator."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryDraftToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_draft_create", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_activate", Description: "Activate a valid draft lottery as an active brand administrator; ticket issuance remains subject to the start and draw times."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryRefToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_activate", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_list", Description: "List the administered planet's lottery drafts and published lotteries with status and search filters."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryListToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_list", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_get", Description: "Read a lottery and its prize slots as an active brand administrator."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryRefToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_get", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_pool", Description: "Read each group's ticket counts and odds in an administered lottery."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryRefToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_pool", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_update", Description: "Edit a draft lottery; an active lottery accepts only text and banner edits. Clear flags remove optional text or caps."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryUpdateToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_update", lotteryUpdatePayload(input))
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_cancel", Description: "Cancel a draft or active lottery with a reason; purchases are refunded and earned tickets move to another open lottery or await the next one."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryCancelToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_lottery_brand_cancel", input)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_lottery_brand_winners", Description: "Read a brand lottery's recorded winners, ticket serials, delivery status and member prize-share snapshots."},
		func(ctx context.Context, _ *mcp.CallToolRequest, input lotteryRefToolInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_lottery_brand_winners", input)
			return nil, out, err
		})
}
