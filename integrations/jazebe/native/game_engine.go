package main

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
)

const gameMinRoundMs = 1500
const gameMaxRoundMs = 10 * 60 * 1000
const gameMaxRoundsPerMinute = 12

func gameTapScore(hits, durationSec int) (float64, error) {
	if durationSec < 3 || durationSec > 600 || hits < 0 || float64(hits) > float64(durationSec)*4 {
		return 0, errors.New("impossible tap result")
	}
	return math.Min(1, float64(hits)/(float64(durationSec)*1.2)), nil
}

func gameTicketsForScore(min, max int, score float64) int {
	if max == 0 {
		return 0
	}
	score = math.Max(0, math.Min(1, score))
	return int(math.Round(float64(min) + float64(max-min)*score))
}

func gameRollBoxes(min, max int) ([3]int, error) {
	var boxes [3]int
	if min < 0 || max < min || max > 10000 {
		return boxes, errors.New("invalid prize range")
	}
	for i := range boxes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
		if err != nil {
			return boxes, err
		}
		boxes[i] = min + int(n.Int64())
	}
	return boxes, nil
}

func gameRoundOutcome(engine string, boxes [3]int, box, hits, durationSec, elapsedMs, prizeMin, prizeMax int) (float64, int, error) {
	if elapsedMs < gameMinRoundMs || elapsedMs > gameMaxRoundMs {
		return 0, 0, errors.New("round outside allowed duration")
	}
	switch engine {
	case "box":
		if box < 0 || box > 2 {
			return 0, 0, errors.New("invalid box")
		}
		tickets := boxes[box]
		if tickets < prizeMin || tickets > prizeMax {
			return 0, 0, errors.New("stored box reward outside prize range")
		}
		if prizeMax == prizeMin {
			return 1, tickets, nil
		}
		return float64(tickets-prizeMin) / float64(prizeMax-prizeMin), tickets, nil
	case "tap":
		if elapsedMs < durationSec*1000 {
			return 0, 0, errors.New("tap round too fast")
		}
		score, err := gameTapScore(hits, durationSec)
		if err != nil {
			return 0, 0, err
		}
		return score, gameTicketsForScore(prizeMin, prizeMax, score), nil
	default:
		return 0, 0, errors.New("unsupported game engine")
	}
}
