package main

import (
	"sort"
	"strconv"
)

func bracketSize(n int) int {
	size := 2
	for size < n {
		size *= 2
	}
	return size
}

func roundCount(size int) int {
	r := 0
	for s := size; s > 1; s /= 2 {
		r++
	}
	return r
}

func seedOrder(size int) []int {
	order := []int{1, 2}
	for len(order) < size {
		n := len(order) * 2
		next := make([]int, 0, n)
		for _, s := range order {
			next = append(next, s, n+1-s)
		}
		order = next
	}
	return order
}

type entrant struct {
	UserID   string
	Trophies int
	Joined   int64
}

func seedEntrants(list []entrant, mode string, rnd func(int) int) []entrant {
	out := append([]entrant(nil), list...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Joined != out[j].Joined {
			return out[i].Joined < out[j].Joined
		}
		return out[i].UserID < out[j].UserID
	})
	if mode == "trophies" {
		sort.SliceStable(out, func(i, j int) bool { return out[i].Trophies > out[j].Trophies })
		return out
	}
	for i := len(out) - 1; i > 0; i-- {
		j := rnd(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

type pairing struct {
	Slot  int
	SeedA int
	SeedB int
}

func firstRound(size, players int) []pairing {
	order := seedOrder(size)
	out := make([]pairing, 0, size/2)
	for i := 0; i < size; i += 2 {
		a, b := order[i], order[i+1]
		if a > players {
			a = 0
		}
		if b > players {
			b = 0
		}
		if a == 0 && b != 0 {
			a, b = b, 0
		}
		out = append(out, pairing{Slot: i / 2, SeedA: a, SeedB: b})
	}
	return out
}

func roundKey(round, rounds int) string {
	switch rounds - round {
	case 0:
		return "final"
	case 1:
		return "semi"
	case 2:
		return "quarter"
	}
	return "r" + strconv.Itoa(1<<(rounds-round+1))
}

func nextSlot(round, slot int) (int, int, string) {
	side := "a"
	if slot%2 == 1 {
		side = "b"
	}
	return round + 1, slot / 2, side
}

func placeOf(lostRound, rounds int) int {
	if lostRound >= rounds {
		return 2
	}
	return (1 << (rounds - lostRound)) + 1
}

func winsNeeded(bestOf int) int {
	if bestOf < 1 {
		bestOf = 1
	}
	return bestOf/2 + 1
}

func poolShare(total int64, percent int) int64 {
	if total <= 0 || percent <= 0 {
		return 0
	}
	return total * int64(percent) / 100
}
