package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

// The source draw algorithm uses HMAC-SHA256(seed, decimal counter), taking
// the first 48 bits with rejection sampling. Serial order is bytewise.
type drawTicket struct{ ID, Serial, GroupID string }
type drawPrize struct {
	ID       string
	Priority int
}
type drawPick struct {
	Priority int    `json:"priority"`
	PrizeID  string `json:"prizeId"`
	Index    int    `json:"index"`
	TicketID string `json:"ticketId"`
	Serial   string `json:"serial"`
	GroupID  string `json:"groupId"`
}
type drawOutcome struct {
	Picks             []drawPick `json:"picks"`
	UnawardedPrizeIDs []string   `json:"unawardedPrizeIds"`
	TicketDigest      string     `json:"ticketDigest"`
	TicketCount       int        `json:"ticketCount"`
}

func drawLottery(seed string, tickets []drawTicket, prizes []drawPrize) (drawOutcome, error) {
	pool := append([]drawTicket(nil), tickets...)
	sort.Slice(pool, func(i, j int) bool { return pool[i].Serial < pool[j].Serial })
	serials := make([]string, len(pool))
	for i, t := range pool {
		serials[i] = t.Serial
	}
	digest := sha256.Sum256([]byte(strings.Join(serials, "\n")))
	out := drawOutcome{Picks: []drawPick{}, UnawardedPrizeIDs: []string{}, TicketDigest: hex.EncodeToString(digest[:]), TicketCount: len(pool)}
	ordered := append([]drawPrize(nil), prizes...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Priority < ordered[j].Priority })
	var counter uint64
	for _, p := range ordered {
		if len(pool) == 0 {
			out.UnawardedPrizeIDs = append(out.UnawardedPrizeIDs, p.ID)
			continue
		}
		index, err := drawUniformIndex(seed, &counter, len(pool))
		if err != nil {
			return drawOutcome{}, err
		}
		t := pool[index]
		out.Picks = append(out.Picks, drawPick{p.Priority, p.ID, index, t.ID, t.Serial, t.GroupID})
		pool = append(pool[:index], pool[index+1:]...)
	}
	return out, nil
}

func drawUniformIndex(seed string, counter *uint64, n int) (int, error) {
	if n <= 0 || uint64(n) > 1<<48 {
		return 0, errors.New("invalid draw pool size")
	}
	const space uint64 = 1 << 48
	limit := space - space%uint64(n)
	for {
		mac := hmac.New(sha256.New, []byte(seed))
		mac.Write([]byte(strconv.FormatUint(*counter, 10)))
		*counter++
		sum := mac.Sum(nil)
		var buf [8]byte
		copy(buf[2:], sum[:6])
		x := binary.BigEndian.Uint64(buf[:])
		if x < limit {
			return int(x % uint64(n)), nil
		}
	}
}

type shareInput struct {
	UserID        string
	MemberTickets int
}
type shareSplit struct {
	UserID                string `json:"userId"`
	IndividualTicketCount int    `json:"individualTicketCount"`
	SharePercent          string `json:"sharePercent"`
	ShareValueRial        string `json:"shareValueRial"`
}

// Group tickets affect the draw odds but do not affect the member shares.
// Remainder rials go to largest fractional remainders, then user ID.
func splitLotteryPrize(value *big.Int, members []shareInput) ([]shareSplit, error) {
	if value == nil || value.Sign() < 0 {
		return nil, errors.New("invalid prize value")
	}
	rows := append([]shareInput(nil), members...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].UserID < rows[j].UserID })
	if len(rows) == 0 {
		return []shareSplit{}, nil
	}
	byTickets := false
	for _, m := range rows {
		if m.MemberTickets > 0 {
			byTickets = true
		}
	}
	weights := make([]*big.Int, len(rows))
	total := new(big.Int)
	for i, m := range rows {
		n := int64(1)
		if byTickets {
			n = int64(m.MemberTickets)
			if n < 0 {
				n = 0
			}
		}
		weights[i] = big.NewInt(n)
		total.Add(total, weights[i])
	}
	type remainder struct {
		index int
		value *big.Int
	}
	parts := make([]*big.Int, len(rows))
	rem := make([]remainder, len(rows))
	allocated := new(big.Int)
	for i, w := range weights {
		x := new(big.Int).Mul(value, w)
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(x, total, r)
		parts[i] = q
		rem[i] = remainder{i, r}
		allocated.Add(allocated, q)
	}
	left := new(big.Int).Sub(value, allocated)
	sort.Slice(rem, func(i, j int) bool {
		c := rem[i].value.Cmp(rem[j].value)
		if c == 0 {
			return rows[rem[i].index].UserID < rows[rem[j].index].UserID
		}
		return c > 0
	})
	for _, r := range rem {
		if left.Sign() == 0 {
			break
		}
		if weights[r.index].Sign() > 0 {
			parts[r.index].Add(parts[r.index], big.NewInt(1))
			left.Sub(left, big.NewInt(1))
		}
	}
	result := make([]shareSplit, len(rows))
	for i, m := range rows {
		// Round weight/total*100 to four decimals, half up.
		scaled := new(big.Int).Mul(weights[i], big.NewInt(1_000_000))
		scaled.Mul(scaled, big.NewInt(2))
		scaled.Add(scaled, total)
		scaled.Quo(scaled, new(big.Int).Mul(total, big.NewInt(2)))
		whole, frac := new(big.Int), new(big.Int)
		whole.QuoRem(scaled, big.NewInt(10_000), frac)
		result[i] = shareSplit{m.UserID, m.MemberTickets, whole.String() + "." + leftPad4(frac.String()), parts[i].String()}
	}
	return result, nil
}

func leftPad4(s string) string {
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}
