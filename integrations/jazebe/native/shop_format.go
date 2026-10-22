package main

import "time"

func shopUntilFa(at, now time.Time) string {
	diff := at.Sub(now)
	if diff <= 0 {
		return "به پایان رسیده"
	}
	minutes := int(diff / time.Minute)
	if minutes < 60 {
		if minutes < 1 {
			minutes = 1
		}
		return faNumber(minutes) + " دقیقه دیگر"
	}
	hours := minutes / 60
	if hours < 24 {
		return faNumber(hours) + " ساعت دیگر"
	}
	days, rest := hours/24, hours%24
	if rest > 0 {
		return faNumber(days) + " روز و " + faNumber(rest) + " ساعت دیگر"
	}
	return faNumber(days) + " روز دیگر"
}

func shopWhenFa(at, now time.Time) string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		loc = time.FixedZone("IR", 3*3600+1800)
	}
	day := at.In(loc).Format("2006-01-02")
	clock := faPad2(at.In(loc).Hour()) + ":" + faPad2(at.In(loc).Minute())
	if day == now.In(loc).Format("2006-01-02") {
		return "امروز · " + clock
	}
	if day == now.Add(-24*time.Hour).In(loc).Format("2006-01-02") {
		return "دیروز · " + clock
	}
	return groupRelativeFa(at, now)
}

func shopTicketRecord(id, productID, name string, paid int, purchased time.Time, lotteryID, lotteryTitle string, now time.Time) map[string]any {
	return map[string]any{"id": id, "pid": productID, "name": name, "type": "ticket", "paid": paid, "when": shopWhenFa(purchased, now), "mo": lotteryMonthsAgo(purchased, now), "code": nil, "raffle": lotteryTitle, "raffleId": lotteryID}
}
