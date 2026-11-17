package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
)

const codeFailedPrecondition = 9

func (s *service) isClan(ctx context.Context, groupID string) bool {
	groupID = strings.TrimSpace(groupID)
	if !uuidPattern.MatchString(groupID) {
		return false
	}
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan WHERE group_id::text=$1)`, strings.ToLower(groupID)).Scan(&ok)
	return ok
}

func (s *service) guard(ctx context.Context, groupID string) error {
	if s.isClan(ctx, groupID) {
		return runtime.NewError("CLAN_USE_RPC: clans change only through the clan_* RPCs", codeFailedPrecondition)
	}
	return nil
}

func guardOf[T any](s *service, id func(T) string) hookchain.BeforeFn[T] {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in T) (T, error) {
		var zero T
		if err := s.guard(ctx, id(in)); err != nil {
			return zero, err
		}
		return in, nil
	}
}

func (s *service) registerGroupGuards(initializer runtime.Initializer) error {
	steps := []func() error{
		func() error {
			return initializer.RegisterBeforeUpdateGroup(hookchain.Before("UpdateGroup", guardOf(s, func(in *api.UpdateGroupRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeDeleteGroup(hookchain.Before("DeleteGroup", guardOf(s, func(in *api.DeleteGroupRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeJoinGroup(hookchain.Before("JoinGroup", guardOf(s, func(in *api.JoinGroupRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeLeaveGroup(hookchain.Before("LeaveGroup", guardOf(s, func(in *api.LeaveGroupRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeAddGroupUsers(hookchain.Before("AddGroupUsers", guardOf(s, func(in *api.AddGroupUsersRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeKickGroupUsers(hookchain.Before("KickGroupUsers", guardOf(s, func(in *api.KickGroupUsersRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeBanGroupUsers(hookchain.Before("BanGroupUsers", guardOf(s, func(in *api.BanGroupUsersRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforePromoteGroupUsers(hookchain.Before("PromoteGroupUsers", guardOf(s, func(in *api.PromoteGroupUsersRequest) string { return in.GetGroupId() })))
		},
		func() error {
			return initializer.RegisterBeforeDemoteGroupUsers(hookchain.Before("DemoteGroupUsers", guardOf(s, func(in *api.DemoteGroupUsersRequest) string { return in.GetGroupId() })))
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
