// Package crib holds the rules of Cribbage and the computer players.
// It has no graphics or sound, so it can be tested on its own.
package crib

import (
	"sort"
	"strconv"
)

// Suit is one of 'S', 'H', 'D', 'C'.
type Suit byte

const (
	Spades   Suit = 'S'
	Hearts   Suit = 'H'
	Diamonds Suit = 'D'
	Clubs    Suit = 'C'
)

var Suits = []Suit{Clubs, Diamonds, Hearts, Spades}

func (s Suit) Name() string {
	switch s {
	case Spades:
		return "Spades"
	case Hearts:
		return "Hearts"
	case Diamonds:
		return "Diamonds"
	case Clubs:
		return "Clubs"
	}
	return ""
}

func (s Suit) Symbol() string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	}
	return ""
}

func (s Suit) Red() bool { return s == Hearts || s == Diamonds }

// Card ranks run 1..13: the ace is low (1), then 2-10, J=11, Q=12, K=13.
type Card struct {
	S Suit
	R int
}

func (c Card) ID() string { return string(c.S) + strconv.Itoa(c.R) }

// ParseID turns "H11" back into a card.
func ParseID(id string) (Card, bool) {
	if len(id) < 2 {
		return Card{}, false
	}
	s := Suit(id[0])
	if s != Spades && s != Hearts && s != Diamonds && s != Clubs {
		return Card{}, false
	}
	r, err := strconv.Atoi(id[1:])
	if err != nil || r < 1 || r > 13 {
		return Card{}, false
	}
	return Card{s, r}, true
}

func RankName(r int) string {
	switch r {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return strconv.Itoa(r)
}

// RankWord names a rank in words ("ace", "seven", "king").
func RankWord(r int) string {
	return [...]string{"", "ace", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "jack", "queen", "king"}[r]
}

func (c Card) Label() string { return RankName(c.R) + c.S.Symbol() }

// Value is what a card counts toward fifteen and thirty-one: face cards are 10.
func (c Card) Value() int { return min(c.R, 10) }

// Deck returns the 52 cards.
func Deck() []Card {
	d := make([]Card, 0, 52)
	for _, s := range Suits {
		for r := 1; r <= 13; r++ {
			d = append(d, Card{s, r})
		}
	}
	return d
}

// SortHand orders a hand for display: by rank, then suit.
func SortHand(h []Card) {
	so := map[Suit]int{Clubs: 0, Diamonds: 1, Hearts: 2, Spades: 3}
	sort.SliceStable(h, func(a, b int) bool {
		if h[a].R != h[b].R {
			return h[a].R < h[b].R
		}
		return so[h[a].S] < so[h[b].S]
	})
}

/* ---------------- how many players, and who's on which side ---------------- */

// Deal is how a game for n players is dealt: cards to each player, cards each
// lays away into the crib, and cards the dealer adds to the crib from the pack.
type Deal struct {
	Cards, Discard, FromPack int
}

// DealFor: two players get six cards and lay away two; three get five, lay
// away one, and one card off the pack fills the crib; four get five and lay
// away one each.
func DealFor(n int) Deal {
	switch n {
	case 3:
		return Deal{5, 1, 1}
	case 4:
		return Deal{5, 1, 0}
	}
	return Deal{6, 2, 0}
}

// Sides is how many scores (and tracks on the board) n players keep: four
// players play as two partnerships.
func Sides(n int) int {
	if n == 4 {
		return 2
	}
	return n
}

// SideOf is the side of the k-th player round the table (k counts from 0).
func SideOf(k, n int) int {
	if n == 4 {
		return k % 2
	}
	return k
}

/* ---------------- scores ---------------- */

// Part is one scoring combination: "Fifteen", "Pair", "Run of 3" and so on.
type Part struct {
	Kind  string
	Pts   int
	Cards []Card
}

func combos(cs []Card, f func([]Card)) {
	n := len(cs)
	for mask := 1; mask < 1<<n; mask++ {
		var sub []Card
		for i := 0; i < n; i++ {
			if mask&(1<<i) != 0 {
				sub = append(sub, cs[i])
			}
		}
		f(sub)
	}
}

// ScoreHand counts a hand of four with the starter in the show. crib is true
// for the crib, which only scores a flush when all five cards match.
func ScoreHand(hand []Card, starter Card, crib bool) (int, []Part) {
	all := append(append([]Card(nil), hand...), starter)
	var parts []Part
	// fifteens
	combos(all, func(sub []Card) {
		if len(sub) < 2 {
			return
		}
		s := 0
		for _, c := range sub {
			s += c.Value()
		}
		if s == 15 {
			parts = append(parts, Part{"Fifteen", 2, sub})
		}
	})
	// pairs (three or four of a kind score as 3 or 6 pairs)
	byRank := map[int][]Card{}
	for _, c := range all {
		byRank[c.R] = append(byRank[c.R], c)
	}
	for r := 1; r <= 13; r++ {
		cs := byRank[r]
		switch len(cs) {
		case 2:
			parts = append(parts, Part{"Pair", 2, cs})
		case 3:
			parts = append(parts, Part{"Pair royal", 6, cs})
		case 4:
			parts = append(parts, Part{"Double pair royal", 12, cs})
		}
	}
	// runs: each longest stretch of ranks, once for every way to pick the cards
	for r := 1; r <= 13; {
		if len(byRank[r]) == 0 {
			r++
			continue
		}
		e := r
		for e+1 <= 13 && len(byRank[e+1]) > 0 {
			e++
		}
		if L := e - r + 1; L >= 3 {
			ways := [][]Card{nil}
			for k := r; k <= e; k++ {
				var next [][]Card
				for _, w := range ways {
					for _, c := range byRank[k] {
						next = append(next, append(append([]Card(nil), w...), c))
					}
				}
				ways = next
			}
			for _, w := range ways {
				parts = append(parts, Part{"Run of " + strconv.Itoa(L), L, w})
			}
		}
		r = e + 1
	}
	// flush
	same := true
	for _, c := range hand[1:] {
		same = same && c.S == hand[0].S
	}
	if same && len(hand) == 4 {
		switch {
		case starter.S == hand[0].S:
			parts = append(parts, Part{"Flush", 5, all})
		case !crib:
			parts = append(parts, Part{"Flush", 4, append([]Card(nil), hand...)})
		}
	}
	// his nobs: the jack of the starter's suit
	for _, c := range hand {
		if c.R == 11 && c.S == starter.S {
			parts = append(parts, Part{"His nobs", 1, []Card{c}})
		}
	}
	t := 0
	for _, p := range parts {
		t += p.Pts
	}
	return t, parts
}

// Count is the running total of a pegging series.
func Count(pile []Card) int {
	s := 0
	for _, c := range pile {
		s += c.Value()
	}
	return s
}

// Playable is the cards in hand that keep the count at 31 or under.
func Playable(hand []Card, count int) []Card {
	var out []Card
	for _, c := range hand {
		if count+c.Value() <= 31 {
			out = append(out, c)
		}
	}
	return out
}

// PegScore is what the last card of a pegging series scores (the series
// since the count last went back to zero, the new card last).
func PegScore(pile []Card) (int, []Part) {
	n := len(pile)
	if n == 0 {
		return 0, nil
	}
	var parts []Part
	switch Count(pile) {
	case 15:
		parts = append(parts, Part{"Fifteen", 2, nil})
	case 31:
		parts = append(parts, Part{"Thirty-one", 2, nil})
	}
	last := pile[n-1]
	k := 1
	for k < n && pile[n-1-k].R == last.R {
		k++
	}
	switch k {
	case 2:
		parts = append(parts, Part{"Pair", 2, pile[n-2:]})
	case 3:
		parts = append(parts, Part{"Pair royal", 6, pile[n-3:]})
	case 4:
		parts = append(parts, Part{"Double pair royal", 12, pile[n-4:]})
	}
	for L := n; L >= 3; L-- {
		if isRun(pile[n-L:]) {
			parts = append(parts, Part{"Run of " + strconv.Itoa(L), L, pile[n-L:]})
			break
		}
	}
	t := 0
	for _, p := range parts {
		t += p.Pts
	}
	return t, parts
}

func isRun(cs []Card) bool {
	lo, hi, seen := 99, 0, map[int]bool{}
	for _, c := range cs {
		if seen[c.R] {
			return false
		}
		seen[c.R] = true
		lo, hi = min(lo, c.R), max(hi, c.R)
	}
	return hi-lo == len(cs)-1
}

// Skunk says how badly a side lost a game to target: 1 for a skunk (under
// 91 in a game of 121, under 31 in a game of 61), 2 for a double skunk.
func Skunk(score, target int) int {
	if target <= 61 {
		if score < 31 {
			return 1
		}
		return 0
	}
	switch {
	case score < 61:
		return 2
	case score < 91:
		return 1
	}
	return 0
}

// Describe sums parts up the way players call them out: "Fifteen 2, fifteen 4, a pair is 6".
func Describe(parts []Part) string {
	s, run := "", 0
	for i, p := range parts {
		run += p.Pts
		w := ""
		switch p.Kind {
		case "Fifteen":
			w = "fifteen " + strconv.Itoa(run)
		case "Pair":
			w = "a pair is " + strconv.Itoa(run)
		case "Pair royal":
			w = "three of a kind is " + strconv.Itoa(run)
		case "Double pair royal":
			w = "four of a kind is " + strconv.Itoa(run)
		case "Flush":
			w = "a flush is " + strconv.Itoa(run)
		case "His nobs":
			w = "and nobs is " + strconv.Itoa(run)
		default:
			w = "a " + lower(p.Kind) + " is " + strconv.Itoa(run)
		}
		if i == 0 {
			w = upper(w)
		} else {
			s += ", "
		}
		s += w
	}
	return s
}

func upper(s string) string {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func lower(s string) string {
	if s == "" || s[0] < 'A' || s[0] > 'Z' {
		return s
	}
	return string(s[0]+32) + s[1:]
}
