package crib

import "math"

// The computer players. They lay away the cards that keep the best hand on
// average over every possible starter, weighing what those cards are worth
// to the crib (theirs or not), and peg for points while trying not to leave
// an easy fifteen, thirty-one or pair royal for the next player.

// ChooseDiscard picks the n cards to lay away. ownCrib is true when the crib
// belongs to the player (or their partner).
func ChooseDiscard(hand []Card, n int, ownCrib bool) []Card {
	seen := map[Card]bool{}
	for _, c := range hand {
		seen[c] = true
	}
	var rest []Card
	for _, c := range Deck() {
		if !seen[c] {
			rest = append(rest, c)
		}
	}
	best, bestV := []Card(nil), math.Inf(-1)
	pick(len(hand), n, func(idx []int) {
		var keep, away []Card
		in := map[int]bool{}
		for _, i := range idx {
			in[i] = true
			away = append(away, hand[i])
		}
		for i, c := range hand {
			if !in[i] {
				keep = append(keep, c)
			}
		}
		v := 0.0
		for _, s := range rest {
			t, _ := ScoreHand(keep, s, false)
			v += float64(t)
		}
		v /= float64(len(rest))
		cv := cribGuess(away)
		if ownCrib {
			v += cv
		} else {
			v -= cv
		}
		// a little tie-break: keep low cards for pegging
		for _, c := range keep {
			if c.Value() <= 4 {
				v += 0.05
			}
		}
		if v > bestV {
			best, bestV = away, v
		}
	})
	return best
}

// pick calls f with every way to choose k of n indices.
func pick(n, k int, f func([]int)) {
	idx := make([]int, k)
	var rec func(start, d int)
	rec = func(start, d int) {
		if d == k {
			f(append([]int(nil), idx...))
			return
		}
		for i := start; i < n; i++ {
			idx[d] = i
			rec(i+1, d+1)
		}
	}
	rec(0, 0)
}

// cribGuess is a rough idea of what cards laid away add to a crib.
func cribGuess(away []Card) float64 {
	v := 0.0
	for _, c := range away {
		switch {
		case c.R == 5:
			v += 1.6
		case c.Value() == 10:
			v += 0.5
		case c.R == 11:
			v += 0.7
		}
	}
	if len(away) == 2 {
		a, b := away[0], away[1]
		switch {
		case a.Value()+b.Value() == 15:
			v += 2
		case a.R == b.R:
			v += 2
		}
		if d := a.R - b.R; d == 1 || d == -1 {
			v += 0.9
		} else if d == 2 || d == -2 {
			v += 0.4
		}
		if a.S == b.S {
			v += 0.15
		}
	}
	return v
}

// PegState is what a computer player knows when it's their turn to peg.
type PegState struct {
	Hand []Card // cards still in hand
	Pile []Card // the current series, oldest first
	Seen []Card // every card it has seen this hand (its own, the starter, played cards)
}

// ChoosePeg picks a card to play from those that fit under 31 (at least one must).
func ChoosePeg(st PegState) Card {
	count := Count(st.Pile)
	ok := Playable(st.Hand, count)
	if len(ok) == 1 {
		return ok[0]
	}
	// how many of each rank might still be out there
	left := map[int]int{}
	for r := 1; r <= 13; r++ {
		left[r] = 4
	}
	for _, c := range st.Seen {
		left[c.R]--
	}
	for _, c := range st.Hand {
		left[c.R]--
	}
	unseen := 0
	for r := 1; r <= 13; r++ {
		left[r] = max(0, left[r])
		unseen += left[r]
	}
	chance := func(f func(r int) bool) float64 { // that the next player holds one such card
		k := 0
		for r := 1; r <= 13; r++ {
			if f(r) {
				k += left[r]
			}
		}
		if unseen == 0 {
			return 0
		}
		p := float64(k) / float64(unseen)
		return 1 - math.Pow(1-p, 4)
	}
	best, bestV := ok[0], math.Inf(-1)
	for _, c := range ok {
		pile := append(append([]Card(nil), st.Pile...), c)
		pts, _ := PegScore(pile)
		v := float64(pts) * 1.0
		nc := count + c.Value()
		// what the next player could answer with
		risk := 0.0
		risk += 2 * chance(func(r int) bool { return nc+min(r, 10) == 15 })
		risk += 2 * chance(func(r int) bool { return nc+min(r, 10) == 31 })
		rp := 2.0
		if len(pile) >= 2 && pile[len(pile)-2].R == c.R {
			rp = 6
		}
		if nc+c.Value() <= 31 {
			risk += rp * chance(func(r int) bool { return r == c.R }) * 0.8
		}
		// leaving a run open
		if len(pile) >= 2 {
			a, b := pile[len(pile)-2].R, c.R
			if d := a - b; d == 1 || d == -1 || d == 2 || d == -2 {
				risk += 1.5 * chance(func(r int) bool {
					return nc+min(r, 10) <= 31 && isRun([]Card{{R: a}, {R: b}, {R: r}})
				})
			}
		}
		v -= risk * 0.75
		if len(st.Pile) == 0 { // leading
			if c.R == 5 {
				v -= 1
			}
			if c.Value() < 5 {
				v += 0.4
			}
		}
		// keep fives and low cards for later, spend tens when it's safe
		if c.R == 5 && nc != 15 {
			v -= 0.3
		}
		// a card that can't be followed past 31 by much is a good way to take the go
		if nc >= 27 {
			v += 0.2
		}
		if v > bestV {
			best, bestV = c, v
		}
	}
	return best
}
