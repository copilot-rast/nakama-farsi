package main

import (
	"strings"
	"time"
)

var jalaliMonthNames = []string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}

// Jazebe's dates use the Persian calendar in Asia/Tehran. The arithmetic
// conversion is independent of the host locale; UTC instants become Tehran
// civil dates before conversion.
func lotteryDateFa(at time.Time) string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		loc = time.FixedZone("IR", 3*3600+1800)
	}
	civil := at.In(loc)
	jy, jm, jd := gregorianToJalali(civil.Year(), int(civil.Month()), civil.Day())
	return faNumber(jd) + " " + jalaliMonthNames[jm-1] + " " + faNumber(jy)
}

func gregorianToJalali(year, month, day int) (int, int, int) {
	var jy, gy int
	if year > 1600 {
		jy = 979
		gy = year - 1600
	} else {
		jy = 0
		gy = year - 621
	}
	gy2 := gy
	if month > 2 {
		gy2++
	}
	monthDays := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	days := 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 - 80 + day + monthDays[month-1]
	jy += 33 * (days / 12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	if days < 186 {
		return jy, 1 + days/31, 1 + days%31
	}
	return jy, 7 + (days-186)/30, 1 + (days-186)%30
}

func lotteryMonthsAgo(at, now time.Time) int {
	days := int(now.Sub(at).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days / 30
}

func lotteryRialFa(value string) string {
	if value == "" {
		value = "0"
	}
	var b strings.Builder
	for i, r := range value {
		if i > 0 && (len(value)-i)%3 == 0 {
			b.WriteRune('٬')
		}
		if r >= '0' && r <= '9' {
			b.WriteRune(persianDigits[r-'0'])
		} else {
			b.WriteRune(r)
		}
	}
	return b.String() + " ریال"
}

func lotteryDrawInstantFa(at time.Time) string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		loc = time.FixedZone("IR", 3*3600+1800)
	}
	c := at.In(loc)
	year, month, day := gregorianToJalali(c.Year(), int(c.Month()), c.Day())
	return faNumber(year) + "/" + faPad2(month) + "/" + faPad2(day) + " · " + faPad2(c.Hour()) + ":" + faPad2(c.Minute()) + ":" + faPad2(c.Second())
}

func faPad2(n int) string {
	if n < 10 {
		return "۰" + faNumber(n)
	}
	return faNumber(n)
}

func lotteryTrackingCode(id string) string {
	id = strings.ReplaceAll(id, "-", "")
	if len(id) > 6 {
		id = id[:6]
	}
	return "RF-" + strings.ToUpper(id)
}
