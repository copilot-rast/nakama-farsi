package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InitModule is loaded by Nakama's existing Go runtime. All callers use
// Nakama's authenticated user ID; this module creates no parallel user records.
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if _, err := mediaUploadPerMinute(); err != nil {
		return err
	}
	if _, _, err := shopRateConfig(); err != nil {
		return err
	}
	for _, entry := range jazebeRPCEntries() {
		if err := initializer.RegisterRpc(entry.name, nakamaRPCError(guardJazebeRPC(entry.fn))); err != nil {
			return err
		}
	}
	if err := registerMediaHTTP(initializer, db); err != nil {
		return err
	}
	if err := registerReferralHTTP(initializer, db); err != nil {
		return err
	}
	if err := startLotteryDrawWorker(ctx, logger, db); err != nil {
		return err
	}
	if err := startProgressionWorker(ctx, logger, db); err != nil {
		return err
	}
	if err := startMedalBackfillWorker(ctx, logger, db); err != nil {
		return err
	}
	if err := startReferralCompletionWorker(ctx, logger, db); err != nil {
		return err
	}
	if err := startAccountRevocationWorker(ctx, logger, db, nk); err != nil {
		return err
	}
	logger.Info("Jazebe native integration RPCs registered")
	return nil
}

type jazebeRPCEntry struct {
	name string
	fn   jazebeRPC
}

// jazebeRPCEntries lists every RPC InitModule registers, in registration order.
func jazebeRPCEntries() []jazebeRPCEntry {
	return []jazebeRPCEntry{
		{"jazebe_brand_create", brandCreate},
		{"jazebe_brand_list", brandList},
		{"jazebe_profile_bootstrap", profileBootstrap},
		{"jazebe_profile_terms_accept", profileTermsAccept},
		{"jazebe_profile_name_available", profileNameAvailable},
		{"jazebe_profile_setup", profileSetup},
		{"jazebe_profile_rename", profileRename},
		{"jazebe_profile_avatar_set", profileSetAvatar},
		{"jazebe_profile_overview", profileOverview},
		{"jazebe_profile_public", profilePublic},
		{"jazebe_profile_report", profileReport},
		{"jazebe_operator_report_list", operatorReportList},
		{"jazebe_operator_report_review", operatorReportReview},
		{"jazebe_operator_report_resolve", operatorReportResolve},
		{"jazebe_operator_user_list", operatorUserList},
		{"jazebe_operator_user_get", operatorUserGet},
		{"jazebe_operator_user_audit", operatorUserAudit},
		{"jazebe_operator_user_suspend", operatorAccountSuspend},
		{"jazebe_operator_user_unsuspend", operatorAccountUnsuspend},
		{"jazebe_operator_user_reactivate", operatorAccountReactivate},
		{"jazebe_operator_user_revocation_retry", operatorAccountRevocationRetry},
		{"jazebe_operator_user_constellation", operatorUserConstellation},
		{"jazebe_planet_create", planetCreate},
		{"jazebe_planet_list", planetList},
		{"jazebe_brand_planets", brandPlanets},
		{"jazebe_brand_planet_get", brandPlanetGet},
		{"jazebe_brand_planet_update", brandPlanetUpdate},
		{"jazebe_planet_publish_check", planetPublishCheck},
		{"jazebe_planet_publish_request", planetRequestPublish},
		{"jazebe_planet_publish_withdraw", planetWithdrawPublish},
		{"jazebe_planet_reject", planetReject},
		{"jazebe_planet_approve", planetApprove},
		{"jazebe_planet_suspend", planetSuspend},
		{"jazebe_planet_unsuspend", planetUnsuspend},
		{"jazebe_planet_archive", planetArchive},
		{"jazebe_brand_planet_categories", brandPlanetCategories},
		{"jazebe_operator_planet_category_list", operatorPlanetCategoryList},
		{"jazebe_operator_planet_category_create", operatorPlanetCategoryCreate},
		{"jazebe_operator_planet_category_update", operatorPlanetCategoryUpdate},
		{"jazebe_operator_planet_category_delete", operatorPlanetCategoryDelete},
		{"jazebe_media_mission_draft_create", mediaMissionDraftCreate},
		{"jazebe_media_mission_activate", mediaMissionActivate},
		{"jazebe_media_mission_start", mediaMissionStart},
		{"jazebe_media_mission_complete", mediaMissionComplete},
		{"jazebe_media_mission_review", mediaMissionReview},
		{"jazebe_mission_review", mediaMissionReview},
		{"jazebe_mission_review_queue", missionReviewQueue},
		{"jazebe_media_mission_list", mediaMissionList},
		{"jazebe_media_mission_history", mediaMissionHistory},
		{"jazebe_mission_history", mediaMissionHistory},
		{"jazebe_mission_history_item", missionHistoryItem},
		{"jazebe_media_mission_withdraw", mediaMissionWithdraw},
		{"jazebe_mission_withdraw", mediaMissionWithdraw},
		{"jazebe_code_mission_draft_create", codeMissionDraftCreate},
		{"jazebe_code_mission_activate", codeMissionActivate},
		{"jazebe_code_mission_generate", codeMissionGenerate},
		{"jazebe_code_mission_redeem", codeMissionRedeem},
		{"jazebe_code_mission_codes", codeMissionCodes},
		{"jazebe_content_mission_draft_create", contentMissionDraftCreate},
		{"jazebe_content_mission_activate", contentMissionActivate},
		{"jazebe_content_mission_submit", contentMissionSubmit},
		{"jazebe_media_upload_begin", mediaUploadBegin},
		{"jazebe_media_upload_delete", mediaUploadDelete},
		{"jazebe_mission_list", missionList},
		{"jazebe_mission_detail", missionDetail},
		{"jazebe_mission_admin_list", missionAdminList},
		{"jazebe_mission_admin_detail", missionAdminDetail},
		{"jazebe_mission_admin_suspend", missionAdminSuspend},
		{"jazebe_mission_admin_unsuspend", missionAdminUnsuspend},
		{"jazebe_mission_status", missionSetStatus},
		{"jazebe_point_balance", pointBalance},
		{"jazebe_wallet_summary", walletSummary},
		{"jazebe_wallet_transactions", walletTransactions},
		{"jazebe_wallet_adjust", walletAdjust},
		{"jazebe_brand_team_list", brandTeamList},
		{"jazebe_group_create", groupCreate},
		{"jazebe_group_meta", groupMeta},
		{"jazebe_group_name_availability", groupNameAvailability},
		{"jazebe_group_join", groupJoin},
		{"jazebe_group_membership_state", groupMembershipState},
		{"jazebe_group_cancel_request", groupCancelRequest},
		{"jazebe_group_leave", groupLeave},
		{"jazebe_group_decide_request", groupDecideRequest},
		{"jazebe_group_request_queue", groupRequestQueue},
		{"jazebe_group_decide_many", groupDecideMany},
		{"jazebe_group_invite_create", groupInviteMint},
		{"jazebe_group_invite_preview", groupInvitePreview},
		{"jazebe_group_invite_join", groupInviteJoin},
		{"jazebe_group_invite_revoke", groupInviteRevoke},
		{"jazebe_group_update", groupUpdate},
		{"jazebe_group_set_role", groupSetRole},
		{"jazebe_group_leadership_offer", groupLeadershipOffer},
		{"jazebe_group_leadership_respond", groupRespondLeadership},
		{"jazebe_group_leadership_withdraw", groupWithdrawLeadership},
		{"jazebe_group_remove_member", groupRemoveMember},
		{"jazebe_group_disband", groupDisband},
		{"jazebe_group_members", groupMembers},
		{"jazebe_group_activity", groupActivity},
		{"jazebe_group_my_request", groupMyRequest},
		{"jazebe_group_list", groupList},
		{"jazebe_group_mine", groupMine},
		{"jazebe_group_detail", groupDetail},
		{"jazebe_group_history", groupHistory},
		{"jazebe_group_chat_list", groupChatList},
		{"jazebe_group_chat_send", groupChatSend},
		{"jazebe_group_chat_mark_read", groupChatMarkRead},
		{"jazebe_group_chat_delete", groupChatDelete},
		{"jazebe_group_chat_report", groupChatReport},
		{"jazebe_group_chat_subscribe", groupChatSubscribe},
		{"jazebe_group_chat_unsubscribe", groupChatUnsubscribe},
		{"jazebe_group_chat_moderate", groupChatModerate},
		{"jazebe_group_chat_reports", groupChatReports},
		{"jazebe_group_chat_dismiss_report", groupChatDismissReport},
		{"jazebe_lottery_draft_create", lotteryDraftCreate},
		{"jazebe_lottery_activate", lotteryActivate},
		{"jazebe_lottery_brand_list", lotteryBrandList},
		{"jazebe_lottery_brand_get", lotteryBrandGet},
		{"jazebe_lottery_brand_pool", lotteryBrandPool},
		{"jazebe_lottery_brand_update", lotteryBrandUpdate},
		{"jazebe_lottery_brand_cancel", lotteryCancel},
		{"jazebe_lottery_client_active", lotteryClientActive},
		{"jazebe_lottery_client_finished", lotteryClientFinished},
		{"jazebe_lottery_client_detail", lotteryClientDetail},
		{"jazebe_lottery_brand_draw", lotteryBrandDraw},
		{"jazebe_lottery_admin_replay", lotteryAdminReplay},
		{"jazebe_lottery_admin_list", lotteryAdminList},
		{"jazebe_lottery_admin_audit", lotteryAdminAudit},
		{"jazebe_lottery_admin_tickets", lotteryAdminTickets},
		{"jazebe_lottery_admin_report", lotteryAdminReport},
		{"jazebe_lottery_brand_winners", lotteryBrandWinners},
		{"jazebe_lottery_brand_delivery", lotteryBrandDelivery},
		{"jazebe_lottery_brand_meta", lotteryBrandMeta},
		{"jazebe_shop_ticket_create", shopTicketCreate},
		{"jazebe_shop_ticket_purchase", shopTicketPurchase},
		{"jazebe_shop_purchases", shopPurchases},
		{"jazebe_shop_products", shopCatalogList},
		{"jazebe_shop_product", shopCatalogGet},
		{"jazebe_shop_coupon_create", couponProductCreate},
		{"jazebe_shop_coupon_generate", couponGenerate},
		{"jazebe_shop_coupon_purchase", couponPurchase},
		{"jazebe_shop_digital_create", digitalProductCreate},
		{"jazebe_shop_digital_purchase", digitalPurchase},
		{"jazebe_shop_purchase", shopPurchase},
		{"jazebe_shop_brand_list", shopBrandList},
		{"jazebe_shop_brand_get", shopBrandGet},
		{"jazebe_shop_brand_status", shopBrandSetStatus},
		{"jazebe_shop_brand_lotteries", shopBrandLotteryOptions},
		{"jazebe_shop_brand_coupon_inventory", shopBrandCouponInventory},
		{"jazebe_shop_brand_coupons", shopBrandCouponList},
		{"jazebe_shop_brand_sales", shopBrandSales},
		{"jazebe_shop_brand_orders", shopBrandOrders},
		{"jazebe_shop_brand_product_orders", shopBrandProductOrders},
		{"jazebe_shop_brand_update", shopBrandUpdate},
		{"jazebe_shop_admin_list", shopAdminList},
		{"jazebe_shop_admin_get", shopAdminGet},
		{"jazebe_shop_admin_suspend", shopAdminSuspend},
		{"jazebe_shop_admin_unsuspend", shopAdminUnsuspend},
		{"jazebe_medal_library_list", medalLibraryList},
		{"jazebe_medal_library_meta", medalLibraryMeta},
		{"jazebe_medal_library_create", medalLibraryCreate},
		{"jazebe_medal_library_get", medalLibraryGet},
		{"jazebe_medal_library_update", medalLibraryUpdate},
		{"jazebe_medal_library_activate", medalLibraryActivate},
		{"jazebe_medal_library_deactivate", medalLibraryDeactivate},
		{"jazebe_medal_library_history", medalLibraryHistory},
		{"jazebe_brand_medals", medalBrandList},
		{"jazebe_brand_medal_meta", medalBrandMeta},
		{"jazebe_brand_medal_library", medalBrandLibrary},
		{"jazebe_brand_medal_add", medalBrandAdd},
		{"jazebe_brand_medal_update", medalBrandUpdate},
		{"jazebe_brand_medal_status", medalBrandStatus},
		{"jazebe_brand_medal_backfill_status", medalBrandBackfillStatus},
		{"jazebe_medal_wall", medalWall},
		{"jazebe_game_start", gameStart},
		{"jazebe_game_list", gameClientList},
		{"jazebe_referral_my_code", referralMyCode},
		{"jazebe_notification_list", notificationList},
		{"jazebe_notification_unread_count", notificationUnreadCount},
		{"jazebe_notification_read_all", notificationReadAll},
		{"jazebe_notification_mark_read", notificationMarkRead},
		{"jazebe_notification_remove", notificationRemove},
		{"jazebe_notification_clear", notificationClear},
		{"jazebe_settings_preferences_get", settingsPreferencesGet},
		{"jazebe_settings_preferences_update", settingsPreferencesUpdate},
		{"jazebe_settings_planet_order_get", settingsPlanetOrderGet},
		{"jazebe_settings_planet_order_update", settingsPlanetOrderUpdate},
		{"jazebe_referral_stats", referralStats},
		{"jazebe_referral_accept", referralAccept},
		{"jazebe_referral_mission_link", referralMissionLink},
		{"jazebe_invite_mission_draft_create", inviteMissionDraftCreate},
		{"jazebe_invite_mission_activate", inviteMissionActivate},
		{"jazebe_game_claim", gameClaim},
		{"jazebe_brand_game_enable", gameBrandEnable},
		{"jazebe_brand_game_update", gameBrandUpdate},
		{"jazebe_brand_game_status", gameBrandStatus},
		{"jazebe_brand_game_report", gameBrandReport},
		{"jazebe_brand_game_get", gameBrandGet},
		{"jazebe_brand_game_list", gameBrandList},
		{"jazebe_brand_game_catalog", gameBrandCatalog},
		{"jazebe_brand_game_meta", gameBrandMeta},
		{"jazebe_game_template_list", gameTemplateList},
		{"jazebe_game_template_get", gameTemplateGet},
		{"jazebe_game_template_create", gameTemplateCreate},
		{"jazebe_game_template_update", gameTemplateUpdate},
		{"jazebe_game_template_suspend", gameTemplateSuspend},
		{"jazebe_game_template_unsuspend", gameTemplateUnsuspend},
		{"jazebe_game_admin_list", gameAdminList},
		{"jazebe_game_admin_get", gameAdminGet},
		{"jazebe_game_admin_suspend", gameAdminSuspend},
		{"jazebe_game_admin_unsuspend", gameAdminUnsuspend},
		{"jazebe_platform_policy_get", platformPolicyGet},
		{"jazebe_platform_policy_update", platformPolicyUpdate},
		{"jazebe_brand_platform_policy_get", brandPlatformPolicyGet},
		{"jazebe_planet_join", planetJoin},
		{"jazebe_planet_memberships", planetMemberships},
		{"jazebe_planet_overview", planetOverview},
		{"jazebe_planet_constellation", planetConstellation},
		{"jazebe_planet_missions_preview", planetMissionsPreview},
		{"jazebe_planet_groups_preview", planetGroupsPreview},
		{"jazebe_planet_raffle_preview", planetRafflePreview},
		{"jazebe_planet_discover", planetDiscover},
	}
}

type brand struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type planet struct {
	ID        string    `json:"id"`
	BrandID   string    `json:"brandId"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

var planetSlug = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,62}[a-z0-9])?$`)

func planetName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 128 {
		return "", status.Error(codes.InvalidArgument, "planet name must contain 2 to 128 Unicode characters")
	}
	return name, nil
}

func validatedSlug(value string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(value))
	if !planetSlug.MatchString(slug) {
		return "", status.Error(codes.InvalidArgument, "invalid planet slug")
	}
	return slug, nil
}

func userID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || id == "" {
		return "", status.Error(codes.Unauthenticated, "Nakama user session required")
	}
	if _, err := uuid.FromString(id); err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid Nakama user identity")
	}
	return id, nil
}

func decode(payload string, out any) error {
	if len(payload) > 8192 {
		return status.Error(codes.InvalidArgument, "request too large")
	}
	if payload == "" {
		payload = "{}"
	}
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return status.Error(codes.InvalidArgument, "invalid JSON payload")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return status.Error(codes.InvalidArgument, "multiple JSON values are not allowed")
	}
	return nil
}

func requiredName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if !utf8.ValidString(name) || name == "" || utf8.RuneCountInString(name) > 120 {
		return "", status.Error(codes.InvalidArgument, "name must contain 1 to 120 Unicode characters")
	}
	return name, nil
}

func requiredUUID(value string) (string, error) {
	id, err := uuid.FromString(value)
	if err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid ID")
	}
	return id.String(), nil
}

func encode(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", status.Error(codes.Internal, "could not encode response")
	}
	return string(data), nil
}

func brandCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	operatorID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(operatorID, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "Jazebe operator role required")
	}
	var input struct {
		Name        string `json:"name"`
		OwnerUserID string `json:"ownerUserId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	name, err := requiredName(input.Name)
	if err != nil {
		return "", err
	}
	ownerID, err := requiredUUID(input.OwnerUserID)
	if err != nil {
		return "", err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return "", status.Error(codes.Internal, "could not create ID")
	}
	result := brand{ID: id.String(), Name: name}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	defer tx.Rollback()
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO jazebe_brand (id, owner_id, name) VALUES ($1, $2, $3) RETURNING created_at`,
		result.ID, ownerID, name).Scan(&result.CreatedAt); err != nil {
		logger.Error("Jazebe brand creation failed: %v", err)
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_brand_member (brand_id,user_id,role) VALUES ($1,$2,'owner')`, result.ID, ownerID); err != nil {
		logger.Error("Jazebe brand owner membership failed: %v", err)
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_brand_admin (brand_id,user_id,role,full_name)
		VALUES ($1,$2,'main',$3)`, result.ID, ownerID, name); err != nil {
		logger.Error("Jazebe brand main admin creation failed: %v", err)
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,brand_id,params)
		VALUES ('brand.create','operator',$1,'brand',$2,$2,jsonb_build_object('ownerUserId',$3::text,'name',$4::text))`,
		operatorID, result.ID, ownerID, name); err != nil {
		logger.Error("Jazebe brand audit creation failed: %v", err)
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand creation failed")
	}
	return encode(result)
}

func isOperator(userID, rawAllowlist string) bool {
	for _, candidate := range strings.Split(rawAllowlist, ",") {
		if strings.EqualFold(strings.TrimSpace(candidate), userID) {
			return true
		}
	}
	return false
}

func brandList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct{}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT b.id, b.name, a.role, b.created_at FROM jazebe_brand b
		JOIN jazebe_brand_admin a ON a.brand_id = b.id
		WHERE a.user_id = $1 AND a.is_active ORDER BY b.name ASC LIMIT 100`, ownerID)
	if err != nil {
		logger.Error("Jazebe brand list failed: %v", err)
		return "", status.Error(codes.Internal, "could not list brands")
	}
	defer rows.Close()
	items := make([]brand, 0)
	for rows.Next() {
		var item brand
		if err := rows.Scan(&item.ID, &item.Name, &item.Role, &item.CreatedAt); err != nil {
			return "", status.Error(codes.Internal, "could not list brands")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "could not list brands")
	}
	return encode(map[string]any{"brands": items})
}

func planetCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		BrandID          string `json:"brandId"`
		Name             string `json:"name"`
		Slug             string `json:"slug"`
		ShortDescription string `json:"shortDescription"`
		CategoryID       string `json:"categoryId"`
		LatinName        string `json:"latinName"`
		PointName        string `json:"pointName"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	brandID, err := requiredUUID(input.BrandID)
	if err != nil {
		return "", err
	}
	name, err := planetName(input.Name)
	if err != nil {
		return "", err
	}
	slug, err := validatedSlug(input.Slug)
	if err != nil {
		return "", err
	}
	if !utf8.ValidString(input.ShortDescription) || utf8.RuneCountInString(input.ShortDescription) > 512 ||
		!utf8.ValidString(input.LatinName) || utf8.RuneCountInString(input.LatinName) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid planet description or Latin name")
	}
	pointName := strings.TrimSpace(input.PointName)
	if pointName == "" {
		pointName = "امتیاز"
	}
	if !utf8.ValidString(pointName) || utf8.RuneCountInString(pointName) > 32 {
		return "", status.Error(codes.InvalidArgument, "invalid point name")
	}
	var categoryID any
	if input.CategoryID != "" {
		categoryID, err = requiredUUID(input.CategoryID)
		if err != nil {
			return "", err
		}
	}
	id, err := uuid.NewV4()
	if err != nil {
		return "", status.Error(codes.Internal, "could not create ID")
	}
	result := planet{ID: id.String(), BrandID: brandID, Name: name, Slug: slug, Status: "setup"}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	defer tx.Rollback()
	var owned bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_brand_admin
		WHERE brand_id = $1 AND user_id = $2 AND is_active AND role IN ('main','technical'))`, brandID, ownerID).Scan(&owned); err != nil {
		logger.Error("Jazebe brand ownership lookup failed: %v", err)
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	if !owned {
		return "", status.Error(codes.PermissionDenied, "active main or technical brand role required")
	}
	// Serialize same-brand creations so the source's case-insensitive name rule holds.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774102, hashtext($1))`, brandID); err != nil {
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	var duplicate bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE brand_id = $1 AND lower(name) = lower($2))`, brandID, name).Scan(&duplicate); err != nil {
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	if duplicate {
		return "", status.Error(codes.AlreadyExists, "planet name already exists in this brand")
	}
	if categoryID != nil {
		var valid bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet_category WHERE id = $1 AND is_active)`, categoryID).Scan(&valid); err != nil {
			return "", status.Error(codes.Internal, "planet category lookup failed")
		}
		if !valid {
			return "", status.Error(codes.NotFound, "active planet category not found")
		}
	}
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet
		(id, brand_id, name, slug, status, short_description, category_id, latin_name, point_name, created_by)
		VALUES ($1, $2, $3, $4, 'setup', $5, $6, $7, $8, $9) RETURNING created_at`,
		result.ID, brandID, name, slug, strings.TrimSpace(input.ShortDescription), categoryID, strings.TrimSpace(input.LatinName), pointName, ownerID).Scan(&result.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "planet name or slug already exists")
		}
		logger.Error("Jazebe planet creation failed: %v", err)
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_planet_settings (planet_id) VALUES ($1)`, result.ID); err != nil {
		logger.Error("Jazebe planet settings creation failed: %v", err)
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,brand_id,planet_id,params)
		VALUES ('planet.create','brand_admin',$1,'planet',$2,$3,$2,jsonb_build_object('name',$4::text,'slug',$5::text))`,
		ownerID, result.ID, brandID, name, slug); err != nil {
		logger.Error("Jazebe planet creation audit failed: %v", err)
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error("Jazebe planet creation commit failed: %v", err)
		return "", status.Error(codes.Internal, "planet creation failed")
	}
	return encode(result)
}

func planetList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := userID(ctx); err != nil {
		return "", err
	}
	var input struct{}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx,
		`SELECT id, brand_id, name, created_at FROM jazebe_planet WHERE status = 'active' ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		logger.Error("Jazebe planet list failed: %v", err)
		return "", status.Error(codes.Internal, "could not list planets")
	}
	defer rows.Close()
	items := make([]planet, 0)
	for rows.Next() {
		var item planet
		if err := rows.Scan(&item.ID, &item.BrandID, &item.Name, &item.CreatedAt); err != nil {
			return "", status.Error(codes.Internal, "could not list planets")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "could not list planets")
	}
	return encode(map[string]any{"planets": items})
}

// brandPlanets is the brand panel's draft-aware list, scoped to active admins.
func brandPlanets(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		BrandID string `json:"brandId"`
		Status  string `json:"status"`
		Page    int    `json:"page"`
		Limit   int    `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	brandID, err := requiredUUID(input.BrandID)
	if err != nil {
		return "", err
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Page < 1 || input.Limit < 1 || input.Limit > 100 || input.Page > 100000 {
		return "", status.Error(codes.InvalidArgument, "invalid pagination")
	}
	switch input.Status {
	case "", "setup", "pending_review", "active", "suspended", "archived":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid planet status")
	}
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_planet p JOIN jazebe_brand b ON b.id = p.brand_id
		JOIN jazebe_brand_admin a ON a.brand_id=b.id AND a.user_id=$2 AND a.is_active
		WHERE b.id = $1 AND ($3 = '' OR p.status = $3)`, brandID, ownerID, input.Status).Scan(&total); err != nil {
		logger.Error("Jazebe brand planet count failed: %v", err)
		return "", status.Error(codes.Internal, "could not list brand planets")
	}
	rows, err := db.QueryContext(ctx, `SELECT p.id, p.brand_id, p.name, COALESCE(p.slug, ''), p.status, p.created_at
		FROM jazebe_planet p JOIN jazebe_brand b ON b.id = p.brand_id
		JOIN jazebe_brand_admin a ON a.brand_id=b.id AND a.user_id=$2 AND a.is_active
		WHERE b.id = $1 AND ($3 = '' OR p.status = $3)
		ORDER BY p.created_at DESC LIMIT $4 OFFSET $5`, brandID, ownerID, input.Status, input.Limit, (input.Page-1)*input.Limit)
	if err != nil {
		logger.Error("Jazebe brand planet list failed: %v", err)
		return "", status.Error(codes.Internal, "could not list brand planets")
	}
	defer rows.Close()
	items := make([]planet, 0)
	for rows.Next() {
		var item planet
		if err := rows.Scan(&item.ID, &item.BrandID, &item.Name, &item.Slug, &item.Status, &item.CreatedAt); err != nil {
			return "", status.Error(codes.Internal, "could not list brand planets")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "could not list brand planets")
	}
	pages := (total + input.Limit - 1) / input.Limit
	return encode(map[string]any{"planets": items, "meta": map[string]int{"total": total, "page": input.Page, "limit": input.Limit, "pages": pages}})
}

func brandPlanetGet(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var result []byte
	err = db.QueryRowContext(ctx, `SELECT jsonb_build_object(
		'planet', jsonb_build_object('id',p.id,'brandId',p.brand_id,'name',p.name,'slug',p.slug,
			'status',p.status,'shortDescription',p.short_description,'categoryId',p.category_id,
			'latinName',p.latin_name,'pointName',p.point_name,'brandColor',p.brand_color,
			'iconAsset',p.icon_asset,'glyph',p.glyph,'fansCount',p.fans_count,'createdAt',p.created_at),
		'settings', jsonb_build_object(
			'missionRewardPointMin',s.mission_reward_point_min,'missionRewardPointMax',s.mission_reward_point_max,
			'missionRewardTicketMax',s.mission_reward_ticket_max,'ticketPriceMin',s.ticket_price_min,
			'ticketPriceMax',s.ticket_price_max,'lotteryFeePercent',s.lottery_fee_percent,
			'groupMemberCapMin',s.group_member_cap_min,'groupMemberCapMax',s.group_member_cap_max,
			'groupMaxCount',s.group_max_count,'groupAdminCap',s.group_admin_cap,
			'groupCreationEnabled',s.group_creation_enabled,'allowedGroupModes',s.allowed_group_modes,
			'xpLevelThresholds',s.xp_level_thresholds,'freeGameDailyCap',s.free_game_daily_cap,
			'xpPerPurchase',s.xp_per_purchase,'xpPerGroupJoin',s.xp_per_group_join,
			'displayNameChangeLimitPerMonth',s.display_name_change_limit_per_month,
			'pointToTomanRate',s.point_to_toman_rate))
		FROM jazebe_planet p JOIN jazebe_brand b ON b.id = p.brand_id
		JOIN jazebe_planet_settings s ON s.planet_id = p.id
		JOIN jazebe_brand_admin a ON a.brand_id=b.id AND a.user_id=$2 AND a.is_active
		WHERE p.id = $1`, planetID, ownerID).Scan(&result)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		logger.Error("Jazebe brand planet read failed: %v", err)
		return "", status.Error(codes.Internal, "could not read planet")
	}
	return string(result), nil
}

func planetJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	memberID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Jazebe planet join transaction failed: %v", err)
		return "", status.Error(codes.Internal, "planet join failed")
	}
	defer tx.Rollback()
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE id = $1 AND status = 'active')`, planetID).Scan(&active); err != nil {
		logger.Error("Jazebe planet lookup failed: %v", err)
		return "", status.Error(codes.Internal, "planet lookup failed")
	}
	if !active {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	var insertedID string
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership (planet_id, user_id, join_path)
		VALUES ($1, $2, 'join_button') ON CONFLICT (planet_id, user_id) DO NOTHING RETURNING planet_id`, planetID, memberID).Scan(&insertedID)
	created := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		logger.Error("Jazebe planet join failed: %v", err)
		return "", status.Error(codes.Internal, "planet join failed")
	}
	if created {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count = fans_count + 1 WHERE id = $1`, planetID); err != nil {
			logger.Error("Jazebe planet fan count failed: %v", err)
			return "", status.Error(codes.Internal, "planet join failed")
		}
	}
	var joinPath string
	var joinedAt time.Time
	if err := tx.QueryRowContext(ctx, `SELECT join_path, joined_at FROM jazebe_planet_membership WHERE planet_id = $1 AND user_id = $2`, planetID, memberID).Scan(&joinPath, &joinedAt); err != nil {
		logger.Error("Jazebe membership read failed: %v", err)
		return "", status.Error(codes.Internal, "planet join failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error("Jazebe planet join commit failed: %v", err)
		return "", status.Error(codes.Internal, "planet join failed")
	}
	return encode(map[string]any{"planetId": planetID, "joined": created, "membership": map[string]any{"planetId": planetID, "joinPath": joinPath, "joinedAt": joinedAt}})
}

func planetMemberships(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	memberID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct{}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT p.id, p.brand_id, p.name, p.fans_count, m.join_path, m.joined_at, m.last_activity_at
		FROM jazebe_planet_membership m JOIN jazebe_planet p ON p.id = m.planet_id
		WHERE m.user_id = $1 AND m.is_active AND p.status = 'active'
		ORDER BY m.sort_order ASC NULLS LAST, m.last_activity_at DESC NULLS LAST, m.joined_at DESC LIMIT 100`, memberID)
	if err != nil {
		logger.Error("Jazebe membership list failed: %v", err)
		return "", status.Error(codes.Internal, "could not list memberships")
	}
	defer rows.Close()
	type membership struct {
		PlanetID       string       `json:"planetId"`
		BrandID        string       `json:"brandId"`
		Name           string       `json:"name"`
		Fans           int          `json:"fans"`
		JoinPath       string       `json:"joinPath"`
		JoinedAt       time.Time    `json:"joinedAt"`
		LastActivityAt sql.NullTime `json:"-"`
	}
	items := make([]map[string]any, 0)
	for rows.Next() {
		var m membership
		if err := rows.Scan(&m.PlanetID, &m.BrandID, &m.Name, &m.Fans, &m.JoinPath, &m.JoinedAt, &m.LastActivityAt); err != nil {
			return "", status.Error(codes.Internal, "could not list memberships")
		}
		item := map[string]any{"planetId": m.PlanetID, "brandId": m.BrandID, "name": m.Name, "fans": m.Fans, "joinPath": m.JoinPath, "joinedAt": m.JoinedAt, "lastActivityAt": nil}
		if m.LastActivityAt.Valid {
			item["lastActivityAt"] = m.LastActivityAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "could not list memberships")
	}
	return encode(map[string]any{"memberships": items})
}
