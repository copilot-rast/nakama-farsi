package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Operator brand management (source planet-admin.service.ts brands section)
// and the brand-admin context (brand-me.service.ts). jazebe_brand.name is the
// trade name; is_active is the tenant switch contract_status drives.

const brandRowColumns = `b.id,b.name,b.legal_name,b.national_company_id,b.contact_phone,b.contact_email,b.contact_address,
	b.contract_status,b.contract_signed_at,b.is_active,b.owner_id,b.created_at,b.updated_at,
	(SELECT count(*) FROM jazebe_planet p WHERE p.brand_id=b.id)`

func scanBrandRow(row interface{ Scan(...any) error }) (map[string]any, error) {
	var id, name, contract, owner string
	var legal, company, phone, email, address sql.NullString
	var signed sql.NullTime
	var active bool
	var created, updated time.Time
	var planets int
	if err := row.Scan(&id, &name, &legal, &company, &phone, &email, &address, &contract, &signed, &active, &owner, &created, &updated, &planets); err != nil {
		return nil, err
	}
	var signedAt any
	if signed.Valid {
		signedAt = signed.Time
	}
	return map[string]any{"id": id, "tradeName": name, "legalName": nullable(legal), "nationalCompanyId": nullable(company),
		"contactPhone": nullable(phone), "contactEmail": nullable(email), "contactAddress": nullable(address),
		"contractStatus": contract, "contractSignedAt": signedAt, "isActive": active, "ownerUserId": owner,
		"createdAt": created, "updatedAt": updated, "planetCount": planets}, nil
}

func brandAdmins(ctx context.Context, db *sql.DB, brandID string) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `SELECT user_id,full_name,role,is_active FROM jazebe_brand_admin WHERE brand_id=$1 ORDER BY created_at ASC`, brandID)
	if err != nil {
		return nil, status.Error(codes.Internal, "brand administrators unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, name, role string
		var active bool
		if err := rows.Scan(&id, &name, &role, &active); err != nil {
			return nil, status.Error(codes.Internal, "brand administrators unavailable")
		}
		// The member is identified by the Nakama user ID; no phone identity exists.
		out = append(out, map[string]any{"id": id, "userId": id, "phone": nil, "fullName": name, "role": role, "isActive": active})
	}
	return out, rows.Err()
}

func operatorBrandRow(ctx context.Context, db *sql.DB, brandID string) (map[string]any, error) {
	row, err := scanBrandRow(db.QueryRowContext(ctx, `SELECT `+brandRowColumns+` FROM jazebe_brand b WHERE b.id=$1`, brandID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "brand not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "brand unavailable")
	}
	if row["admins"], err = brandAdmins(ctx, db, brandID); err != nil {
		return nil, err
	}
	return row, nil
}

func brandIDFrom(payload string) (string, error) {
	var in struct {
		BrandID string `json:"brandId"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	return requiredUUID(in.BrandID)
}

func operatorBrandList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in struct {
		Search         string `json:"search"`
		ContractStatus string `json:"contractStatus"`
		Page           int    `json:"page"`
		Limit          int    `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	in.Search = strings.TrimSpace(in.Search)
	if !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 64 {
		return "", status.Error(codes.InvalidArgument, "invalid brand search")
	}
	switch in.ContractStatus {
	case "", "active", "suspended", "terminated":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid contract status")
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid brand page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(in.Search) + "%"
	filter := ` FROM jazebe_brand b WHERE ($1::text='' OR b.name ILIKE $2 OR b.legal_name ILIKE $2) AND ($3::text='' OR b.contract_status=$3::text)`
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*)`+filter, in.Search, pattern, in.ContractStatus).Scan(&total); err != nil {
		return "", status.Error(codes.Internal, "brands unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT `+brandRowColumns+filter+` ORDER BY b.created_at DESC, b.id DESC LIMIT $4 OFFSET $5`,
		in.Search, pattern, in.ContractStatus, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "brands unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		row, err := scanBrandRow(rows)
		if err != nil {
			return "", status.Error(codes.Internal, "brands unavailable")
		}
		items = append(items, row)
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "brands unavailable")
	}
	rows.Close()
	for _, row := range items {
		if row["admins"], err = brandAdmins(ctx, db, row["id"].(string)); err != nil {
			return "", err
		}
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func operatorBrandGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	id, err := brandIDFrom(payload)
	if err != nil {
		return "", err
	}
	row, err := operatorBrandRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func brandAudit(ctx context.Context, tx *sql.Tx, action, actor, brandID string, params map[string]any) error {
	raw, _ := json.Marshal(params)
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,brand_id,params)
		VALUES($1,'operator',$2,'brand',$3,$3,$4::jsonb)`, action, actor, brandID, string(raw)); err != nil {
		return status.Error(codes.Internal, "brand audit failed")
	}
	return nil
}

// operatorBrandUpdate patches the commercial record (source UpdateBrandDto).
func operatorBrandUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	var in map[string]json.RawMessage
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	var id string
	if json.Unmarshal(in["brandId"], &id) != nil {
		return "", status.Error(codes.InvalidArgument, "brandId required")
	}
	if id, err = requiredUUID(id); err != nil {
		return "", err
	}
	delete(in, "brandId")
	columns := map[string]string{"tradeName": "name", "legalName": "legal_name", "nationalCompanyId": "national_company_id",
		"contactPhone": "contact_phone", "contactEmail": "contact_email", "contactAddress": "contact_address", "contractSignedAt": "contract_signed_at"}
	limits := map[string]int{"tradeName": 256, "legalName": 256, "nationalCompanyId": 32, "contactPhone": 32, "contactEmail": 320, "contactAddress": 2000}
	sets := []string{}
	args := []any{id}
	changed := []string{}
	for key, raw := range in {
		column, ok := columns[key]
		if !ok {
			return "", status.Error(codes.InvalidArgument, "unsupported brand field: "+key)
		}
		var value *string
		if json.Unmarshal(raw, &value) != nil {
			return "", status.Error(codes.InvalidArgument, "invalid "+key)
		}
		var arg any
		if value != nil {
			v := strings.TrimSpace(*value)
			switch {
			case key == "contractSignedAt":
				at, err := time.Parse(time.RFC3339, v)
				if err != nil {
					return "", status.Error(codes.InvalidArgument, "invalid contractSignedAt")
				}
				arg = at
			case key == "contactEmail":
				if _, err := mail.ParseAddress(v); err != nil || utf8.RuneCountInString(v) > limits[key] {
					return "", status.Error(codes.InvalidArgument, "invalid contactEmail")
				}
				arg = v
			default:
				if !utf8.ValidString(v) || utf8.RuneCountInString(v) > limits[key] || key == "tradeName" && utf8.RuneCountInString(v) < 2 {
					return "", status.Error(codes.InvalidArgument, "invalid "+key)
				}
				arg = v
			}
		} else if key == "tradeName" {
			return "", status.Error(codes.InvalidArgument, "tradeName cannot be null")
		}
		args = append(args, arg)
		sets = append(sets, column+"=$"+stringInt(len(args)))
		changed = append(changed, key)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand update unavailable")
	}
	defer tx.Rollback()
	// Column names come from the fixed map above, never from the payload.
	res, err := tx.ExecContext(ctx, `UPDATE jazebe_brand SET `+strings.Join(append(sets, "updated_at=now()"), ",")+` WHERE id=$1`, args...)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "national company ID already registered")
		}
		return "", status.Error(codes.Internal, "brand update failed")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", status.Error(codes.NotFound, "brand not found")
	}
	if err := brandAudit(ctx, tx, "brand.update", actor, id, map[string]any{"fields": changed}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand update failed")
	}
	row, err := operatorBrandRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

// operatorBrandContract sets the contract state (PRD 12.3.4). A suspended or
// terminated contract deactivates the brand, which takes its planets and
// brand-panel routes down; termination is final.
func operatorBrandContract(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		BrandID string `json:"brandId"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.BrandID, err = requiredUUID(in.BrandID); err != nil {
		return "", err
	}
	if in.Status != "active" && in.Status != "suspended" && in.Status != "terminated" {
		return "", status.Error(codes.InvalidArgument, "invalid contract status")
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "reason must be 3 to 512 characters")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "contract update unavailable")
	}
	defer tx.Rollback()
	var from string
	err = tx.QueryRowContext(ctx, `SELECT contract_status FROM jazebe_brand WHERE id=$1 FOR UPDATE`, in.BrandID).Scan(&from)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "brand not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "contract update unavailable")
	}
	if from == "terminated" && in.Status != "terminated" {
		return "", status.Error(codes.FailedPrecondition, "a terminated contract cannot be reopened")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_brand SET contract_status=$2,is_active=$3,updated_at=now() WHERE id=$1`,
		in.BrandID, in.Status, in.Status == "active"); err != nil {
		return "", status.Error(codes.Internal, "contract update failed")
	}
	if err := brandAudit(ctx, tx, "brand.contract", actor, in.BrandID, map[string]any{"from": from, "to": in.Status, "reason": in.Reason}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "contract update failed")
	}
	row, err := operatorBrandRow(ctx, db, in.BrandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

// operatorBrandAdminAdd assigns an existing user (source addBrandAdmin by
// userId; the phone variant needs a trusted phone identity). A terminated
// brand takes no new administrators.
func operatorBrandAdminAdd(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	in.FullName = strings.TrimSpace(in.FullName)
	if !utf8.ValidString(in.FullName) || utf8.RuneCountInString(in.FullName) < 2 || utf8.RuneCountInString(in.FullName) > 128 {
		return "", status.Error(codes.InvalidArgument, "full name must be 2 to 128 characters")
	}
	if in.Role == "" {
		in.Role = "marketing"
	}
	if !validBrandRole(in.Role) {
		return "", status.Error(codes.InvalidArgument, "invalid brand role")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand administrator add unavailable")
	}
	defer tx.Rollback()
	var contract string
	err = tx.QueryRowContext(ctx, `SELECT contract_status FROM jazebe_brand WHERE id=$1`, in.BrandID).Scan(&contract)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "brand not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "brand unavailable")
	}
	if contract == "terminated" {
		return "", status.Error(codes.FailedPrecondition, "brand contract terminated")
	}
	if err = lockXact(ctx, tx, "jz-brand-team:"+in.BrandID); err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	target, _, err := brandTeamUser(ctx, tx, in)
	if err != nil {
		return "", err
	}
	if target == "" {
		return "", status.Error(codes.NotFound, "user not found")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_brand_member(brand_id,user_id,role) VALUES($1,$2,'admin') ON CONFLICT (brand_id,user_id) DO NOTHING`, in.BrandID, target); err != nil {
		return "", status.Error(codes.Internal, "brand membership failed")
	}
	var active sql.NullBool
	if err := tx.QueryRowContext(ctx, `SELECT is_active FROM jazebe_brand_admin WHERE brand_id=$1 AND user_id=$2`, in.BrandID, target).Scan(&active); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	reactivated := false
	switch {
	case active.Valid && active.Bool:
		return "", status.Error(codes.AlreadyExists, "user is already a brand administrator")
	case active.Valid:
		reactivated = true
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_brand_admin SET is_active=true,role=$3,full_name=$4 WHERE brand_id=$1 AND user_id=$2`, in.BrandID, target, in.Role, in.FullName)
	default:
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_brand_admin(brand_id,user_id,role,full_name,is_active) VALUES($1,$2,$3,$4,true)`, in.BrandID, target, in.Role, in.FullName)
	}
	if err != nil {
		logger.Error("Jazebe operator brand admin add failed: %v", err)
		return "", status.Error(codes.Internal, "brand administrator add failed")
	}
	if err := brandAudit(ctx, tx, "brand.admin_add", actor, in.BrandID, map[string]any{"userId": target, "role": in.Role, "reactivated": reactivated, "userCreated": false}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand administrator add failed")
	}
	row, err := operatorBrandRow(ctx, db, in.BrandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func operatorBrandAdminChange(ctx context.Context, db *sql.DB, payload string, change func(context.Context, *sql.Tx, string, brandTeamInput) error) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	in, err := decodeBrandTeam(payload)
	if err != nil {
		return "", err
	}
	if in.MemberID, err = requiredUUID(in.MemberID); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	defer tx.Rollback()
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand WHERE id=$1)`, in.BrandID).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "brand unavailable")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "brand not found")
	}
	if err := change(ctx, tx, actor, in); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "brand team update failed")
	}
	row, err := operatorBrandRow(ctx, db, in.BrandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func operatorBrandAdminRole(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return operatorBrandAdminChange(ctx, db, payload, func(ctx context.Context, tx *sql.Tx, actor string, in brandTeamInput) error {
		if !validBrandRole(in.Role) {
			return status.Error(codes.InvalidArgument, "invalid brand role")
		}
		return brandTeamChangeRole(ctx, tx, "operator", actor, in.BrandID, in.MemberID, in.Role)
	})
}

func operatorBrandAdminDeactivate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return operatorBrandAdminChange(ctx, db, payload, func(ctx context.Context, tx *sql.Tx, actor string, in brandTeamInput) error {
		return brandTeamSetActiveTx(ctx, tx, "operator", actor, in.BrandID, in.MemberID, false)
	})
}

func operatorBrandAdminReactivate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return operatorBrandAdminChange(ctx, db, payload, func(ctx context.Context, tx *sql.Tx, actor string, in brandTeamInput) error {
		return brandTeamSetActiveTx(ctx, tx, "operator", actor, in.BrandID, in.MemberID, true)
	})
}

// brandMe lists every brand the caller actively administers (source
// brand-me.service.ts), with the role's panel actions.
func brandMe(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT b.id,b.name,a.role,a.full_name,b.contract_status,b.is_active FROM jazebe_brand_admin a
		JOIN jazebe_brand b ON b.id=a.brand_id WHERE a.user_id=$1 AND a.is_active ORDER BY b.name ASC`, user)
	if err != nil {
		return "", status.Error(codes.Internal, "brands unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, name, role, fullName, contract string
		var active bool
		if err := rows.Scan(&id, &name, &role, &fullName, &contract, &active); err != nil {
			return "", status.Error(codes.Internal, "brands unavailable")
		}
		items = append(items, map[string]any{"brandId": id, "tradeName": name, "role": role, "fullName": fullName,
			"contractStatus": contract, "isActive": active, "actions": brandActions(role)})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "brands unavailable")
	}
	return encode(map[string]any{"items": items})
}
