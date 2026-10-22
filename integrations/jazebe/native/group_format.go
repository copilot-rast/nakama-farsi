package main

import (
	"strconv"
	"strings"
	"time"
)

var groupMemberColors = []string{"#F7931E", "#4ADE80", "#2D78D4", "#60A5FA", "#A78BFA", "#F472B6", "#FBBF24", "#34D399", "#388EF8", "#FF8A8A"}
var persianDigits = []rune("۰۱۲۳۴۵۶۷۸۹")

func groupMemberColor(id string) string {
	var h uint32
	for _, r := range id {
		h = h*31 + uint32(r)
	}
	return groupMemberColors[h%uint32(len(groupMemberColors))]
}

func groupInitial(name string) string {
	for _, r := range strings.TrimSpace(name) {
		return string(r)
	}
	return "؟"
}

func faNumber(value int) string {
	var b strings.Builder
	for _, r := range strconv.Itoa(value) {
		if r >= '0' && r <= '9' {
			b.WriteRune(persianDigits[r-'0'])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func groupRelativeFa(date, now time.Time) string {
	minutes := int(now.Sub(date).Minutes())
	if minutes < 1 {
		return "همین حالا"
	}
	if minutes < 60 {
		return faNumber(minutes) + " دقیقه پیش"
	}
	hours := minutes / 60
	if hours < 24 {
		return faNumber(hours) + " ساعت پیش"
	}
	days := hours / 24
	if days == 1 {
		return "دیروز"
	}
	if days < 7 {
		return faNumber(days) + " روز پیش"
	}
	weeks := days / 7
	if weeks < 5 {
		return faNumber(weeks) + " هفته پیش"
	}
	months := days / 30
	if months < 12 {
		return faNumber(months) + " ماه پیش"
	}
	return faNumber(days/365) + " سال پیش"
}
