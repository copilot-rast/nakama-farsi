package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type shopTicketCreateToolInput struct {
	PlanetID         string `json:"planetId"`
	Name             string `json:"name"`
	ShortDescription string `json:"shortDescription,omitempty"`
	LongDescription  string `json:"longDescription,omitempty"`
	ImageAsset       string `json:"imageAsset,omitempty"`
	PricePoints      int    `json:"pricePoints"`
	Stock            *int   `json:"stock,omitempty"`
	Terms            string `json:"terms,omitempty"`
	DeliveryNote     string `json:"deliveryNote,omitempty"`
	Ticket           struct {
		LotteryID string `json:"lotteryId"`
	} `json:"ticket"`
}
type shopTicketPurchaseToolInput struct {
	PlanetID       string `json:"planetId"`
	ProductID      string `json:"productId"`
	IdempotencyKey string `json:"idempotencyKey" jsonschema:"Caller-generated stable key of 8 to 128 allowed characters; reuse only for retries"`
}

type shopPurchasesToolInput struct {
	PlanetID string `json:"planetId"`
	Type     string `json:"type,omitempty" jsonschema:"all, ticket, coupon, or digital"`
	Range    string `json:"range,omitempty" jsonschema:"all, month (30 days), or quarter (120 days)"`
}
type shopProductsToolInput struct {
	PlanetID string `json:"planetId"`
	Type     string `json:"type,omitempty" jsonschema:"all, ticket, coupon, or digital"`
	Sort     string `json:"sort,omitempty" jsonschema:"fresh, cheap, or expensive"`
}
type shopPlanetToolInput struct {
	PlanetID string `json:"planetId"`
}
type shopProductToolInput struct {
	PlanetID  string `json:"planetId"`
	ProductID string `json:"productId"`
}
type shopBrandListToolInput struct {
	PlanetID string `json:"planetId"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Status   string `json:"status,omitempty"`
	Type     string `json:"type,omitempty"`
	Search   string `json:"search,omitempty"`
}
type shopSalesToolInput struct {
	PlanetID string `json:"planetId"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Type     string `json:"type,omitempty"`
}

func addShopTools(s *mcp.Server, g *gateway, token string) {
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_brand_list", Description: "List shop products and inventory for a planet as an active brand administrator."}, func(ctx context.Context, _ *mcp.CallToolRequest, in shopBrandListToolInput) (*mcp.CallToolResult, map[string]any, error) {
		out, err := g.rpc(ctx, token, "jazebe_shop_brand_list", in)
		return nil, out, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_brand_get", Description: "Read one brand shop product, including inventory and configuration, as an active brand administrator."}, func(ctx context.Context, _ *mcp.CallToolRequest, in shopProductToolInput) (*mcp.CallToolResult, map[string]any, error) {
		out, err := g.rpc(ctx, token, "jazebe_shop_brand_get", in)
		return nil, out, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_brand_lotteries", Description: "List draft and active lotteries eligible for a brand shop ticket product."}, func(ctx context.Context, _ *mcp.CallToolRequest, in shopPlanetToolInput) (*mcp.CallToolResult, []any, error) {
		out, err := g.rpcArray(ctx, token, "jazebe_shop_brand_lotteries", in)
		return nil, out, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_brand_coupon_inventory", Description: "Read available, issued, and expired coupon counts for a brand shop product; raw codes are not returned."}, func(ctx context.Context, _ *mcp.CallToolRequest, in shopProductToolInput) (*mcp.CallToolResult, map[string]any, error) {
		out, err := g.rpc(ctx, token, "jazebe_shop_brand_coupon_inventory", in)
		return nil, out, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_brand_sales", Description: "Read aggregate orders, points, buyers, top products, and product-type sales for a brand planet."}, func(ctx context.Context, _ *mcp.CallToolRequest, in shopSalesToolInput) (*mcp.CallToolResult, map[string]any, error) {
		out, err := g.rpc(ctx, token, "jazebe_shop_brand_sales", in)
		return nil, out, err
	})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_products", Description: "List currently available ticket, coupon, and digital products on an active planet, with stock and caller-specific ticket caps."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopProductsToolInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_shop_products", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_product", Description: "Read one currently available ticket, coupon, or digital product."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopProductToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_product", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_purchases", Description: "Read your own planet shop purchase archive, including delivered ticket, coupon, and digital order details."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopPurchasesToolInput) (*mcp.CallToolResult, []any, error) {
			out, err := g.rpcArray(ctx, token, "jazebe_shop_purchases", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_coupon_purchase", Description: "Buy one pre-generated coupon with planet points atomically; reuse the idempotency key for retries."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopTicketPurchaseToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_coupon_purchase", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_digital_purchase", Description: "Buy one digital product with planet points atomically; reuse the idempotency key for retries."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopTicketPurchaseToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_digital_purchase", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_purchase", Description: "Purchase one available ticket, coupon, or digital product with planet points. Reuse the idempotency key for retries."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopTicketPurchaseToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_purchase", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_ticket_create", Description: "Create an active ticket product for a draft or active lottery as a brand administrator."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopTicketCreateToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_ticket_create", in)
			return nil, out, err
		})
	mcp.AddTool(s, &mcp.Tool{Name: "jazebe_shop_ticket_purchase", Description: "Buy one lottery ticket with planet points atomically; reuse the idempotency key for retries."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in shopTicketPurchaseToolInput) (*mcp.CallToolResult, map[string]any, error) {
			out, err := g.rpc(ctx, token, "jazebe_shop_ticket_purchase", in)
			return nil, out, err
		})
}
