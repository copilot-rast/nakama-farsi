package main

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type shopCatalogInput struct {
	PlanetID  string `json:"planetId"`
	ProductID string `json:"productId"`
	Type      string `json:"type"`
	Sort      string `json:"sort"`
}

func validateShopCatalog(in shopCatalogInput) error {
	switch in.Type {
	case "", "all", "ticket", "coupon", "digital":
	default:
		return status.Error(codes.InvalidArgument, "invalid store type")
	}
	switch in.Sort {
	case "", "fresh", "cheap", "expensive":
	default:
		return status.Error(codes.InvalidArgument, "invalid store sort")
	}
	return nil
}

type shopCatalogTicket struct {
	view  map[string]any
	price int
	fresh int64
}

// Ticket catalog follows Jazebe's sale window and caller-specific ticket caps.
// Coupon and digital products are introduced by their separate inventory flows.
func shopTicketCatalog(ctx context.Context, db *sql.DB, planetID, actor, productID, sortBy string, now time.Time) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `SELECT p.id,p.name,p.price_points,p.stock,p.freshness_rank,p.created_at,COALESCE(p.short_description,''),COALESCE(p.long_description,p.short_description,''),COALESCE(p.terms,''),COALESCE(p.delivery_note,''),l.id,l.title,l.draw_at,l.per_user_ticket_cap,l.per_group_ticket_cap,(SELECT count(*) FROM jazebe_lottery_ticket t WHERE t.lottery_id=l.id AND t.creator_user_id=$2 AND t.status='active'),COALESCE((SELECT gs.ticket_count FROM jazebe_lottery_group_stat gs WHERE gs.lottery_id=l.id AND gs.group_id=gm.group_id),0),gm.group_id IS NOT NULL FROM jazebe_shop_product p JOIN jazebe_lottery l ON l.id::text=p.type_config->>'lotteryId' AND l.planet_id=p.planet_id LEFT JOIN LATERAL (SELECT m.group_id FROM jazebe_group_membership m WHERE m.planet_id=p.planet_id AND m.user_id=$2 AND m.status='active' LIMIT 1) gm ON true WHERE p.planet_id=$1 AND p.status='active' AND p.type='ticket' AND l.status='active' AND l.starts_at <= $4 AND l.draw_at > $4 AND ($3::uuid IS NULL OR p.id=$3) ORDER BY p.created_at DESC,p.id DESC`, planetID, actor, sql.NullString{String: productID, Valid: productID != ""}, now)
	if err != nil {
		return nil, status.Error(codes.Internal, "store catalog unavailable")
	}
	defer rows.Close()
	products := []shopCatalogTicket{}
	for rows.Next() {
		var id, name, short, long, terms, deliver, lotteryID, lotteryTitle string
		var price, userTickets, groupTickets int
		var stock, userCap, groupCap sql.NullInt64
		var rank int64
		var created, draw time.Time
		var inGroup bool
		if err = rows.Scan(&id, &name, &price, &stock, &rank, &created, &short, &long, &terms, &deliver, &lotteryID, &lotteryTitle, &draw, &userCap, &groupCap, &userTickets, &groupTickets, &inGroup); err != nil {
			return nil, status.Error(codes.Internal, "store catalog unavailable")
		}
		if rank == 0 {
			rank = created.Unix()
		}
		if deliver == "" {
			deliver = "افزودن به مخزن گروه"
		}
		var available any
		if stock.Valid {
			available = stock.Int64
		}
		view := map[string]any{"id": id, "type": "ticket", "name": name, "price": price, "stock": available, "fresh": rank, "short": short, "long": long, "terms": terms, "deliver": deliver, "raffle": lotteryTitle, "raffleId": lotteryID, "deadline": shopUntilFa(draw, now)}
		if userCap.Valid && int64(userTickets) >= userCap.Int64 || inGroup && groupCap.Valid && int64(groupTickets) >= groupCap.Int64 {
			view["capReached"] = true
		}
		products = append(products, shopCatalogTicket{view: view, price: price, fresh: rank})
	}
	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, "store catalog unavailable")
	}
	sort.SliceStable(products, func(i, j int) bool {
		switch sortBy {
		case "cheap":
			return products[i].price < products[j].price
		case "expensive":
			return products[i].price > products[j].price
		default:
			return products[i].fresh > products[j].fresh
		}
	})
	out := make([]map[string]any, 0, len(products))
	for _, p := range products {
		out = append(out, p.view)
	}
	return out, nil
}

func shopCouponCatalog(ctx context.Context, db *sql.DB, planetID, productID string, now time.Time) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `SELECT p.id,p.name,p.price_points,p.stock,p.freshness_rank,p.created_at,COALESCE(p.short_description,''),COALESCE(p.long_description,p.short_description,''),COALESCE(p.terms,''),COALESCE(p.delivery_note,''),p.type_config->>'expiresAt',count(c.id) FILTER(WHERE c.status='available' AND c.deleted_at IS NULL AND (c.expires_at IS NULL OR c.expires_at>$3)) FROM jazebe_shop_product p LEFT JOIN jazebe_coupon_code c ON c.shop_product_id=p.id WHERE p.planet_id=$1 AND p.status='active' AND p.type='coupon' AND ($2::uuid IS NULL OR p.id=$2) GROUP BY p.id ORDER BY p.created_at DESC,p.id DESC`, planetID, sql.NullString{String: productID, Valid: productID != ""}, now)
	if err != nil {
		return nil, status.Error(codes.Internal, "coupon catalog unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name, short, long, terms, deliver string
		var price int
		var stock sql.NullInt64
		var rank, available int64
		var created time.Time
		var rawExpiry sql.NullString
		if err = rows.Scan(&id, &name, &price, &stock, &rank, &created, &short, &long, &terms, &deliver, &rawExpiry, &available); err != nil {
			return nil, status.Error(codes.Internal, "coupon catalog unavailable")
		}
		if rank == 0 {
			rank = created.Unix()
		}
		if deliver == "" {
			deliver = "تحویل آنی · نمایش کد"
		}
		effective := available
		if stock.Valid && stock.Int64 < effective {
			effective = stock.Int64
		}
		view := map[string]any{"id": id, "type": "coupon", "name": name, "price": price, "stock": effective, "fresh": rank, "short": short, "long": long, "terms": terms, "deliver": deliver}
		if rawExpiry.Valid && rawExpiry.String != "" {
			expiry, err := time.Parse(time.RFC3339Nano, rawExpiry.String)
			if err != nil {
				return nil, status.Error(codes.Internal, "coupon expiration invalid")
			}
			view["expiry"] = lotteryDateFa(expiry)
		}
		out = append(out, view)
	}
	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, "coupon catalog unavailable")
	}
	return out, nil
}

func shopDigitalCatalog(ctx context.Context, db *sql.DB, planetID, productID string) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `SELECT id,name,price_points,stock,freshness_rank,created_at,COALESCE(short_description,''),COALESCE(long_description,short_description,''),COALESCE(terms,''),COALESCE(delivery_note,'') FROM jazebe_shop_product WHERE planet_id=$1 AND status='active' AND type='digital' AND ($2::uuid IS NULL OR id=$2) ORDER BY created_at DESC,id DESC`, planetID, sql.NullString{String: productID, Valid: productID != ""})
	if err != nil {
		return nil, status.Error(codes.Internal, "digital catalog unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name, short, long, terms, deliver string
		var price int
		var stock sql.NullInt64
		var rank int64
		var created time.Time
		if err = rows.Scan(&id, &name, &price, &stock, &rank, &created, &short, &long, &terms, &deliver); err != nil {
			return nil, status.Error(codes.Internal, "digital catalog unavailable")
		}
		if rank == 0 {
			rank = created.Unix()
		}
		if deliver == "" {
			deliver = "تحویل آنی · نمایش کد"
		}
		var available any
		if stock.Valid {
			available = stock.Int64
		}
		out = append(out, map[string]any{"id": id, "type": "digital", "name": name, "price": price, "stock": available, "fresh": rank, "short": short, "long": long, "terms": terms, "deliver": deliver})
	}
	if err = rows.Err(); err != nil {
		return nil, status.Error(codes.Internal, "digital catalog unavailable")
	}
	return out, nil
}

func sortShopViews(views []map[string]any, sortBy string) {
	sort.SliceStable(views, func(i, j int) bool {
		switch sortBy {
		case "cheap":
			return views[i]["price"].(int) < views[j]["price"].(int)
		case "expensive":
			return views[i]["price"].(int) > views[j]["price"].(int)
		default:
			return views[i]["fresh"].(int64) > views[j]["fresh"].(int64)
		}
	})
}

func shopCatalogList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopCatalogInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = validateShopCatalog(in); err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	now := time.Now().UTC()
	out := []map[string]any{}
	if in.Type == "" || in.Type == "all" || in.Type == "ticket" {
		tickets, err := shopTicketCatalog(ctx, db, in.PlanetID, actor, "", in.Sort, now)
		if err != nil {
			return "", err
		}
		out = append(out, tickets...)
	}
	if in.Type == "" || in.Type == "all" || in.Type == "coupon" {
		coupons, err := shopCouponCatalog(ctx, db, in.PlanetID, "", now)
		if err != nil {
			return "", err
		}
		out = append(out, coupons...)
	}
	if in.Type == "" || in.Type == "all" || in.Type == "digital" {
		digital, err := shopDigitalCatalog(ctx, db, in.PlanetID, "")
		if err != nil {
			return "", err
		}
		out = append(out, digital...)
	}
	if in.Type == "" || in.Type == "all" {
		sortShopViews(out, in.Sort)
	} else if in.Type == "coupon" || in.Type == "digital" {
		sortShopViews(out, in.Sort)
	}
	return encode(out)
}

func shopCatalogGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopCatalogInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.ProductID, err = requiredUUID(in.ProductID)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	out, err := shopTicketCatalog(ctx, db, in.PlanetID, actor, in.ProductID, "", time.Now().UTC())
	if err != nil {
		return "", err
	}
	if len(out) == 0 {
		out, err = shopCouponCatalog(ctx, db, in.PlanetID, in.ProductID, time.Now().UTC())
		if err != nil {
			return "", err
		}
	}
	if len(out) == 0 {
		out, err = shopDigitalCatalog(ctx, db, in.PlanetID, in.ProductID)
		if err != nil {
			return "", err
		}
	}
	if len(out) == 0 {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	return encode(out[0])
}
