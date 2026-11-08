package crib

import (
	"math/rand/v2"
	"testing"
)

func cs(ids ...string) []Card {
	var out []Card
	for _, id := range ids {
		c, ok := ParseID(id)
		if !ok {
			panic(id)
		}
		out = append(out, c)
	}
	return out
}

func TestShow(t *testing.T) {
	cases := []struct {
		hand    []string
		starter string
		crib    bool
		want    int
	}{
		{[]string{"H5", "D5", "S5", "C11"}, "C5", false, 29},
		{[]string{"H5", "D5", "S5", "C5"}, "H13", false, 28},
		{[]string{"H7", "D7", "S8", "C8"}, "H9", false, 24},
		{[]string{"H1", "D2", "S3", "C4"}, "H13", false, 8},                    // a run of 4, A+4+K and 2+3+K
		{[]string{"H2", "H4", "H6", "H8"}, "S13", false, 4},                    // a flush of 4
		{[]string{"H2", "H4", "H6", "H8"}, "S13", true, 0},                     // the crib needs five for a flush
		{[]string{"H2", "H4", "H6", "H8"}, "H13", true, 5},                     // ...and gets them
		{[]string{"H11", "D2", "S4", "C6"}, "H9", false, 1 + 4},                // nobs, 2+4+9 and 6+9
		{[]string{"H10", "D12", "S13", "C11"}, "H5", false, 8 + 3 + 1},         // 4 fifteens, run of 3 (10-J-Q... 10 J Q K is 4)
		{[]string{"H3", "D3", "S4", "C5"}, "H5", false, 2*2 + 2 + 2 + 3*4 + 2}, // see below
		{[]string{"H2", "D4", "S6", "C8"}, "H13", false, 0},                    // nineteen: nothing
		{[]string{"H9", "D9", "S13", "C12"}, "H1", false, 2},                   // just the pair
	}
	// 10-J-Q-K with a 5: four fifteens (8) and a run of four (4), nobs if J matches: JC, starter H5: no nobs
	cases[8].want = 8 + 4
	// 3,3,4,5,5: fifteens 5+5+... 3+3+4+5 (x2) = 4, 5+5+... 5+5 =10 no; 3+3+4+5: two 5s -> 2 ways = 4.
	// pairs: 3s and 5s = 4. runs 3-4-5: 2*2 = 4 runs of 3 = 12. total 20
	cases[9].want = 20
	for _, c := range cases {
		st, _ := ParseID(c.starter)
		got, parts := ScoreHand(cs(c.hand...), st, c.crib)
		if got != c.want {
			t.Errorf("%v + %s (crib %v): got %d want %d: %s", c.hand, c.starter, c.crib, got, c.want, Describe(parts))
		}
	}
}

func TestPeg(t *testing.T) {
	cases := []struct {
		pile []string
		want int
	}{
		{[]string{"H5", "S10"}, 2},
		{[]string{"H7", "S7"}, 2},
		{[]string{"H7", "S7", "D7"}, 6},
		{[]string{"H7", "S7", "D7", "C7"}, 12},
		{[]string{"H3", "S5", "D4"}, 3},
		{[]string{"H3", "S5", "D4", "C6"}, 4},
		{[]string{"H3", "S4", "D3"}, 0},
		{[]string{"H10", "S10", "D10", "C1"}, 2},
		{[]string{"H4", "S5", "D6"}, 2 + 3},      // fifteen and a run
		{[]string{"H13", "S12", "D11", "C1"}, 2}, // 31; K Q J A isn't a run
		{[]string{"H2", "S3", "D2"}, 0},
	}
	for _, c := range cases {
		got, parts := PegScore(cs(c.pile...))
		if got != c.want {
			t.Errorf("%v: got %d want %d (%v)", c.pile, got, c.want, parts)
		}
	}
}

func TestDescribe(t *testing.T) {
	st, _ := ParseID("C5")
	_, parts := ScoreHand(cs("H5", "D5", "S5", "C11"), st, false)
	if d := Describe(parts); d == "" || d[:10] != "Fifteen 2," {
		t.Errorf("describe: %q", d)
	}
}

func TestSkunk(t *testing.T) {
	if Skunk(90, 121) != 1 || Skunk(91, 121) != 0 || Skunk(60, 121) != 2 || Skunk(30, 61) != 1 || Skunk(31, 61) != 0 {
		t.Error("skunk lines")
	}
}

// TestGames plays a few hundred whole games between computer players
// (two, three and four handed) with the full rules and checks they finish.
func TestGames(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for g := 0; g < 300; g++ {
		n := 2 + g%3
		d := DealFor(n)
		sides := Sides(n)
		score := make([]int, sides)
		dealer, hands := 0, 0
		won := -1
		add := func(k, pts int) {
			if won >= 0 {
				return
			}
			s := SideOf(k, n)
			score[s] += pts
			if score[s] >= 121 {
				won = s
			}
		}
		for won < 0 {
			hands++
			if hands > 200 {
				t.Fatal("game never ends")
			}
			deck := Deck()
			r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
			h := make([][]Card, n)
			k := 0
			for i := 0; i < n; i++ {
				h[i] = append([]Card(nil), deck[k:k+d.Cards]...)
				k += d.Cards
			}
			var crib []Card
			for i := 0; i < n; i++ {
				own := SideOf(i, n) == SideOf(dealer, n)
				away := ChooseDiscard(h[i], d.Discard, own)
				if len(away) != d.Discard {
					t.Fatalf("discarded %d", len(away))
				}
				for _, c := range away {
					h[i] = remove(h[i], c)
					crib = append(crib, c)
				}
			}
			for j := 0; j < d.FromPack; j++ {
				crib = append(crib, deck[k])
				k++
			}
			if len(crib) != 4 {
				t.Fatalf("crib has %d", len(crib))
			}
			starter := deck[k]
			for i := range h {
				if len(h[i]) != 4 {
					t.Fatalf("hand has %d", len(h[i]))
				}
			}
			if starter.R == 11 {
				add(dealer, 2)
			}
			// pegging
			held := make([][]Card, n)
			for i := range h {
				held[i] = append([]Card(nil), h[i]...)
			}
			var pile, seen []Card
			turn := (dealer + 1) % n
			last := -1
			passed := make([]bool, n)
			for won < 0 {
				left := 0
				for i := range held {
					left += len(held[i])
				}
				if left == 0 {
					if len(pile) > 0 && Count(pile) != 31 {
						add(last, 1)
					}
					break
				}
				ok := Playable(held[turn], Count(pile))
				if len(ok) == 0 {
					passed[turn] = true
					all := true
					for i := range passed {
						all = all && (passed[i] || len(held[i]) == 0)
					}
					if all {
						if Count(pile) != 31 && last >= 0 {
							add(last, 1)
						}
						pile = nil
						for i := range passed {
							passed[i] = false
						}
						turn = (last + 1) % n
						continue
					}
					turn = (turn + 1) % n
					continue
				}
				c := ChoosePeg(PegState{Hand: held[turn], Pile: pile, Seen: append(append([]Card(nil), seen...), starter)})
				found := false
				for _, x := range ok {
					found = found || x == c
				}
				if !found {
					t.Fatalf("illegal peg %v at %d", c, Count(pile))
				}
				held[turn] = remove(held[turn], c)
				pile = append(pile, c)
				seen = append(seen, c)
				pts, _ := PegScore(pile)
				add(turn, pts)
				last = turn
				if Count(pile) == 31 {
					pile = nil
					for i := range passed {
						passed[i] = false
					}
				}
				turn = (turn + 1) % n
			}
			// the show
			for j := 1; j <= n && won < 0; j++ {
				i := (dealer + j) % n
				pts, _ := ScoreHand(h[i], starter, false)
				add(i, pts)
			}
			if won < 0 {
				pts, _ := ScoreHand(crib, starter, true)
				add(dealer, pts)
			}
			dealer = (dealer + 1) % n
		}
	}
}

func remove(h []Card, c Card) []Card {
	for i, x := range h {
		if x == c {
			return append(append([]Card(nil), h[:i]...), h[i+1:]...)
		}
	}
	panic("not in hand")
}

func TestAIKeepsGoodHands(t *testing.T) {
	// with 5-5-5-J and two junk cards, keep the 29 shape
	h := cs("H5", "D5", "S5", "C11", "H2", "S9")
	away := ChooseDiscard(h, 2, false)
	for _, c := range away {
		if c.R == 5 || c.R == 11 {
			t.Errorf("threw away %v", away)
		}
	}
}
