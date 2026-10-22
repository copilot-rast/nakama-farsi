package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

var errInsufficientPoints = errors.New("insufficient points")

type pointMovement struct {
	WalletID      string
	TransactionID string
	Balance       int64
	Applied       bool
}

// applyPointMovement is the only writer of Jazebe's per-planet point wallet.
// The caller owns the SQL transaction, so purchases, mission completions and
// adjustments can commit or roll back with their side effects.
func applyPointMovement(ctx context.Context, tx *sql.Tx, planetID, userID, direction string, amount int64, source, sourceID, description string, at time.Time) (pointMovement, error) {
	if amount <= 0 || sourceID == "" || !(direction == "credit" || direction == "debit") {
		return pointMovement{}, errors.New("invalid point movement")
	}
	walletID := uuid.Must(uuid.NewV4()).String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_point_wallet (id,planet_id,user_id) VALUES ($1,$2,$3)
		ON CONFLICT (planet_id,user_id) DO NOTHING`, walletID, planetID, userID); err != nil {
		return pointMovement{}, err
	}
	var current int64
	if err := tx.QueryRowContext(ctx, `SELECT id,balance FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2 FOR UPDATE`, planetID, userID).Scan(&walletID, &current); err != nil {
		return pointMovement{}, err
	}
	var existingID string
	var existingBalance int64
	err := tx.QueryRowContext(ctx, `SELECT id,balance_after FROM jazebe_point_transaction
		WHERE wallet_id=$1 AND source=$2 AND source_id=$3`, walletID, source, sourceID).Scan(&existingID, &existingBalance)
	if err == nil {
		return pointMovement{WalletID: walletID, TransactionID: existingID, Balance: current, Applied: false}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return pointMovement{}, err
	}
	next := current + amount
	if direction == "debit" {
		next = current - amount
		if next < 0 {
			return pointMovement{}, errInsufficientPoints
		}
	}
	if next < 0 {
		return pointMovement{}, errors.New("point balance overflow")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_point_wallet SET balance=$2,last_transaction_at=$3 WHERE id=$1`, walletID, next, at); err != nil {
		return pointMovement{}, err
	}
	transactionID := uuid.Must(uuid.NewV4()).String()
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_point_transaction
		(id,wallet_id,planet_id,user_id,direction,amount,balance_after,source,source_id,description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, transactionID, walletID, planetID, userID, direction, amount, next, source, sourceID, description); err != nil {
		return pointMovement{}, err
	}
	return pointMovement{WalletID: walletID, TransactionID: transactionID, Balance: next, Applied: true}, nil
}
