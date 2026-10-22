package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func testGateway(t *testing.T) (*gateway, *httptest.Server) {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/account":
			_, _ = w.Write([]byte(`{"user":{"id":"user-1","username":"test"},"wallet":"{}"}`))
		case "/v2/rpc/jazebe_brand_list":
			if r.Method != http.MethodPost || r.URL.RawQuery != "unwrap" {
				http.Error(w, "bad RPC request", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"brands":[]}`))
		case "/v2/rpc/jazebe_planet_memberships":
			if r.Method != http.MethodPost || r.URL.RawQuery != "unwrap" {
				http.Error(w, "bad RPC request", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"memberships":[{"planetId":"planet-1","joinPath":"join_button"}]}`))
		case "/v2/rpc/jazebe_brand_planets":
			_, _ = w.Write([]byte(`{"planets":[{"id":"draft-1","status":"setup"}]}`))
		case "/v2/rpc/jazebe_brand_planet_get":
			_, _ = w.Write([]byte(`{"planet":{"id":"draft-1","status":"setup"},"settings":{"ticketPriceMin":1}}`))
		case "/v2/rpc/jazebe_brand_planet_update":
			_, _ = w.Write([]byte(`{"planet":{"id":"draft-1","status":"setup","name":"Updated"},"settings":{"ticketPriceMin":1}}`))
		case "/v2/rpc/jazebe_planet_publish_check":
			_, _ = w.Write([]byte(`{"planetId":"draft-1","status":"setup","missing":["active_mission"]}`))
		case "/v2/rpc/jazebe_media_mission_draft_create":
			_, _ = w.Write([]byte(`{"id":"mission-1","planetId":"draft-1","status":"draft","functionalType":"media_view"}`))
		case "/v2/rpc/jazebe_media_mission_start":
			_, _ = w.Write([]byte(`{"missionId":"mission-1","startedAt":"2026-10-22T12:00:00Z","completableAt":"2026-10-22T12:00:30Z"}`))
		case "/v2/rpc/jazebe_point_balance":
			_, _ = w.Write([]byte(`{"planetId":"draft-1","balance":25}`))
		case "/v2/rpc/jazebe_brand_team_list":
			_, _ = w.Write([]byte(`{"members":[{"id":"user-1","role":"main"}]}`))
		case "/v2/rpc/jazebe_lottery_brand_get":
			_, _ = w.Write([]byte(`{"id":"lottery-1","status":"draft","prizes":[{"type":"cash","valueRial":"1000"}]}`))
		case "/v2/rpc/jazebe_lottery_client_active":
			_, _ = w.Write([]byte(`[{"id":"lottery-1","status":"active","groupTickets":2}]`))
		case "/v2/rpc/jazebe_lottery_client_finished":
			_, _ = w.Write([]byte(`[{"id":"lottery-2","status":"finished","winners":["group"]}]`))
		case "/v2/rpc/jazebe_lottery_client_detail":
			_, _ = w.Write([]byte(`{"id":"lottery-2","status":"finished","winningTickets":["T-1"]}`))
		case "/v2/rpc/jazebe_lottery_brand_winners":
			_, _ = w.Write([]byte(`[{"id":"winner-1","ticketSerial":"T-1","shares":[]}]`))
		case "/v2/rpc/jazebe_game_list":
			_, _ = w.Write([]byte(`[{"id":"game-1","engine":"tap","freeInd":2}]`))
		case "/v2/rpc/jazebe_brand_game_report":
			_, _ = w.Write([]byte(`{"plays":2,"freePlays":1,"paidPlays":1,"ticketsIssued":3}`))
		case "/v2/rpc/jazebe_referral_stats":
			_, _ = w.Write([]byte(`{"issuedLinks":1,"registered":1,"qualified":2}`))
		case "/v2/rpc/jazebe_profile_overview":
			_, _ = w.Write([]byte(`{"user":{"name":"player","phone":""},"planet":null,"medals":[],"verifyStatus":"none"}`))
		case "/v2/rpc/jazebe_profile_public":
			_, _ = w.Write([]byte(`{"name":"member","level":1,"medals":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(upstream.Close)
	g, err := newGateway(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	return g, upstream
}

func TestMCPDiscoveryAndInvocation(t *testing.T) {
	g, _ := testGateway(t)
	ctx := context.Background()
	s := newServer(g, "valid-token")
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	if _, err := s.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 112 {
		t.Fatalf("tools: %v, %v", tools, err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "nakama_get_account"})
	if err != nil || result.IsError {
		t.Fatalf("call: %v, %v", result, err)
	}
	encoded, _ := json.Marshal(result.StructuredContent)
	if !strings.Contains(string(encoded), "user-1") {
		t.Fatalf("wrong account: %s", encoded)
	}
	games, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_game_list", Arguments: map[string]any{"planetId": "planet-1"}})
	if err != nil || games.IsError {
		t.Fatalf("game list: %v, %v", games, err)
	}
	encoded, _ = json.Marshal(games.StructuredContent)
	if !strings.Contains(string(encoded), "game-1") {
		t.Fatalf("wrong game list: %s", encoded)
	}
	report, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_brand_game_report", Arguments: map[string]any{"planetId": "planet-1", "gameId": "game-1"}})
	if err != nil || report.IsError {
		t.Fatalf("game report: %v, %v", report, err)
	}
	encoded, _ = json.Marshal(report.StructuredContent)
	if !strings.Contains(string(encoded), "paidPlays") {
		t.Fatalf("wrong game report: %s", encoded)
	}
	referrals, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_referral_stats", Arguments: map[string]any{"planetId": "planet-1"}})
	if err != nil || referrals.IsError {
		t.Fatalf("referral stats: %v, %v", referrals, err)
	}
	encoded, _ = json.Marshal(referrals.StructuredContent)
	if !strings.Contains(string(encoded), "qualified") {
		t.Fatalf("wrong referral stats: %s", encoded)
	}
	profile, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_profile_overview"})
	if err != nil || profile.IsError {
		t.Fatalf("profile overview: %v, %v", profile, err)
	}
	encoded, _ = json.Marshal(profile.StructuredContent)
	if !strings.Contains(string(encoded), "verifyStatus") {
		t.Fatalf("wrong owner profile: %s", encoded)
	}
	public, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_profile_public", Arguments: map[string]any{"planetId": "planet-1", "userId": "user-2"}})
	if err != nil || public.IsError {
		t.Fatalf("public profile: %v, %v", public, err)
	}
	encoded, _ = json.Marshal(public.StructuredContent)
	if strings.Contains(string(encoded), "phone") {
		t.Fatalf("public profile leaked private field: %s", encoded)
	}
	lottery, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_lottery_brand_get", Arguments: map[string]any{"planetId": "planet-1", "lotteryId": "lottery-1"}})
	if err != nil || lottery.IsError {
		t.Fatalf("lottery read: %v, %v", lottery, err)
	}
	encoded, _ = json.Marshal(lottery.StructuredContent)
	if !strings.Contains(string(encoded), "lottery-1") {
		t.Fatalf("wrong lottery: %s", encoded)
	}
	active, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_lottery_client_active", Arguments: map[string]any{"planetId": "planet-1"}})
	if err != nil || active.IsError {
		t.Fatalf("client lottery list: %v, %v", active, err)
	}
	encoded, _ = json.Marshal(active.StructuredContent)
	if !strings.Contains(string(encoded), "lottery-1") {
		t.Fatalf("wrong client lottery list: %s", encoded)
	}
	detail, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_lottery_client_detail", Arguments: map[string]any{"planetId": "planet-1", "lotteryId": "lottery-2"}})
	if err != nil || detail.IsError {
		t.Fatalf("client lottery detail: %v, %v", detail, err)
	}
	encoded, _ = json.Marshal(detail.StructuredContent)
	if !strings.Contains(string(encoded), "winningTickets") {
		t.Fatalf("wrong client detail: %s", encoded)
	}
	result, err = session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_start_media_mission", Arguments: map[string]any{"planetId": "draft-1", "missionId": "mission-1"}})
	if err != nil || result.IsError {
		t.Fatalf("media start MCP call: %v, %v", result, err)
	}
	encoded, _ = json.Marshal(result.StructuredContent)
	if !strings.Contains(string(encoded), "completableAt") {
		t.Fatalf("wrong media start: %s", encoded)
	}
	brandResult, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_list_brands"})
	if err != nil || brandResult.IsError {
		t.Fatalf("brand list: %v, %v", brandResult, err)
	}
	memberships, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_list_my_planets"})
	if err != nil || memberships.IsError {
		t.Fatalf("membership list: %v, %v", memberships, err)
	}
	memberJSON, _ := json.Marshal(memberships.StructuredContent)
	if !strings.Contains(string(memberJSON), "planet-1") {
		t.Fatalf("wrong memberships: %s", memberJSON)
	}
	drafts, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_list_brand_planets", Arguments: map[string]any{"brandId": "brand-1"}})
	if err != nil || drafts.IsError {
		t.Fatalf("draft list: %v, %v", drafts, err)
	}
	draftJSON, _ := json.Marshal(drafts.StructuredContent)
	if !strings.Contains(string(draftJSON), "draft-1") {
		t.Fatalf("wrong drafts: %s", draftJSON)
	}
	draft, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_get_brand_planet", Arguments: map[string]any{"planetId": "draft-1"}})
	if err != nil || draft.IsError {
		t.Fatalf("draft get: %v, %v", draft, err)
	}
	draftDetail, _ := json.Marshal(draft.StructuredContent)
	if !strings.Contains(string(draftDetail), "ticketPriceMin") {
		t.Fatalf("settings missing: %s", draftDetail)
	}
	updated, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_update_brand_planet", Arguments: map[string]any{"planetId": "draft-1", "name": "Updated"}})
	if err != nil || updated.IsError {
		t.Fatalf("draft update: %v, %v", updated, err)
	}
	check, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "jazebe_check_planet_publish", Arguments: map[string]any{"planetId": "draft-1"}})
	if err != nil || check.IsError {
		t.Fatalf("publish check: %v, %v", check, err)
	}
	checkJSON, _ := json.Marshal(check.StructuredContent)
	if !strings.Contains(string(checkJSON), "active_mission") {
		t.Fatalf("publish prerequisite missing: %s", checkJSON)
	}
	resources, err := session.ListResources(ctx, nil)
	if err != nil || len(resources.Resources) != 1 {
		t.Fatalf("resources: %v, %v", resources, err)
	}
	read, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "nakama://me/account"})
	if err != nil || len(read.Contents) != 1 || !strings.Contains(read.Contents[0].Text, "user-1") {
		t.Fatalf("resource read: %v, %v", read, err)
	}
	prompts, err := session.ListPrompts(ctx, nil)
	if err != nil || len(prompts.Prompts) != 1 {
		t.Fatalf("prompts: %v, %v", prompts, err)
	}
}

func TestHTTPAuthorizationAndOrigin(t *testing.T) {
	g, _ := testGateway(t)
	handler := httpHandler(g)
	for _, tc := range []struct {
		name, authorization, origin string
		want                        int
	}{
		{"missing session", "", "", http.StatusUnauthorized},
		{"invalid session", "Bearer bad-token", "", http.StatusUnauthorized},
		{"cross origin", "Bearer valid-token", "https://evil.example", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://localhost/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
			req.Header.Set("Authorization", tc.authorization)
			req.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status %d, want %d", w.Code, tc.want)
			}
		})
	}
}

type tokenTransport struct{ base http.RoundTripper }

func (t tokenTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	clone := r.Clone(r.Context())
	clone.Header.Set("Authorization", "Bearer valid-token")
	return t.base.RoundTrip(clone)
}

func TestStreamableHTTPClient(t *testing.T) {
	g, _ := testGateway(t)
	httpServer := httptest.NewServer(httpHandler(g))
	defer httpServer.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "http-test", Version: "1"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             httpServer.URL,
		HTTPClient:           &http.Client{Transport: tokenTransport{base: http.DefaultTransport}},
		DisableStandaloneSSE: true,
		MaxRetries:           -1,
	}
	ctx := context.Background()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	tools, err := session.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 112 {
		t.Fatalf("HTTP discovery: %v, %v", tools, err)
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "nakama_get_account"})
	if err != nil || result.IsError {
		t.Fatalf("HTTP invocation: %v, %v", result, err)
	}
}
