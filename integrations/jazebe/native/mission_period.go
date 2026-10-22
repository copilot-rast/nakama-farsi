package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func supportedMissionRepeat(pattern, period string) (string, error) {
	if pattern == "" {
		pattern = "once"
	}
	if pattern != "once" && pattern != "periodic" {
		return "", status.Error(codes.InvalidArgument, "unsupported mission repeat pattern")
	}
	if pattern == "periodic" && period != "daily" && period != "weekly" && period != "monthly" {
		return "", status.Error(codes.InvalidArgument, "invalid mission repeat period")
	}
	if pattern == "once" && period != "" {
		return "", status.Error(codes.InvalidArgument, "repeat period requires periodic mission")
	}
	return pattern, nil
}

func supportedReviewedRepeat(pattern, period string) (string, error) {
	if pattern == "unlimited" {
		if period != "" {
			return "", status.Error(codes.InvalidArgument, "unlimited mission cannot have a repeat period")
		}
		return pattern, nil
	}
	return supportedMissionRepeat(pattern, period)
}

// Calendar keys match Jazebe's Tehran day, ISO week and month boundaries.
// The database stores UTC timestamps; only cycle identity uses local time.
func missionPeriodKey(pattern, period string, now time.Time) (string, error) {
	if pattern == "once" {
		return "", nil
	}
	if pattern == "unlimited" {
		var random [3]byte
		if _, err := rand.Read(random[:]); err != nil {
			return "", err
		}
		return "U-" + strings.ToUpper(strconv.FormatInt(now.UnixMilli(), 36)) + "-" + strings.ToUpper(hex.EncodeToString(random[:])), nil
	}
	if pattern != "periodic" {
		return "", fmt.Errorf("unsupported repeat pattern")
	}
	zone, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return "", err
	}
	local := now.In(zone)
	switch period {
	case "daily":
		return local.Format("2006-01-02"), nil
	case "weekly":
		year, week := local.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week), nil
	case "monthly":
		return local.Format("2006-01"), nil
	default:
		return "", fmt.Errorf("unsupported repeat period")
	}
}
