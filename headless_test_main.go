//go:build headless

// A test build with no window or sound: plays whole games with the computer
// at every seat (two, three and four handed), online games between a host
// and guests over real TCP on this computer, and a shared-computer game
// through the pass screen. Run: go run -tags headless .
package main

import (
	"fmt"
	"os"
	"time"

	cr "github.com/adamglassner/cribbage/internal/crib"
)

type uiState struct{}

func check(ok bool, f string, a ...any) {
	if !ok {
		fmt.Printf("FAIL: "+f+"\n", a...)
		os.Exit(1)
	}
}

func waitFor(what string, d time.Duration, f func() bool) {
	end := time.Now().Add(d)
	for time.Now().Before(end) {
		if f() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	check(false, "timed out waiting for %s", what)
}

func locked[T any](g *game, f func() T) T {
	g.mu.Lock()
	defer g.mu.Unlock()
	return f()
}

// fitsFace checks a regular's name was made for their face.
func fitsFace(face, name string) bool {
	f, ok := parseFace(face)
	if !ok {
		return false
	}
	for _, n := range mannerNames[f.mannerKey()].names {
		if n == name {
			return true
		}
	}
	return false
}

// tableOK checks every card is somewhere sensible: the right number in each
// hand and the crib, and pegs never past the target.
func tableOK(g *game) {
	n := 0
	for _, c := range g.deck {
		if c.zn.K != zNone {
			n++
		}
	}
	check(n == 0 || n == 52, "%d cards on the table", n)
	for s := 0; s < g.sides(); s++ {
		check(g.score[s] <= g.target && g.prev[s] <= g.score[s], "side %d pegs %d/%d", s, g.prev[s], g.score[s])
	}
	crib := 0
	for _, c := range g.deck {
		if c.zn.K == zCrib {
			crib++
		}
	}
	check(crib <= 4, "crib has %d", crib)
	check(g.count <= 31, "count %d", g.count)
}

func soloGame(players int) *game {
	g := newGameState()
	g.mu.Lock()
	g.save.Settings.Players = players
	g.save.Settings.Target = 121
	g.rollNext()
	g.goPlay()
	check(g.n() == players, "table of %d has %d", players, g.n())
	seen := map[string]bool{}
	for _, i := range g.act[1:] {
		p := g.players[i]
		check(p.kind == KindAI && fitsFace(p.face, p.name), "seat %d: %q doesn't fit %s", i, p.name, p.face)
		check(!seen[p.face] && !seen[p.name], "repeated face or name")
		seen[p.face], seen[p.name] = true, true
		fmt.Printf("%d players, seat %d: %s (%s)\n", players, i, p.full, p.face)
	}
	g.mu.Unlock()
	maxCrib := 0
	waitFor(fmt.Sprintf("a %d-player game", players), 5*time.Minute, func() bool {
		return locked(g, func() bool {
			tableOK(g)
			c := 0
			for _, x := range g.deck {
				if x.zn.K == zCrib {
					c++
				}
			}
			maxCrib = max(maxCrib, c)
			return g.showResult
		})
	})
	g.mu.Lock()
	defer g.mu.Unlock()
	check(maxCrib == 4, "the crib never filled (%d)", maxCrib)
	check(g.winner >= 0 && g.score[g.winner] == g.target, "winner %d scores %v", g.winner, g.score)
	for s := 0; s < g.sides(); s++ {
		if s != g.winner {
			check(g.score[s] < g.target, "two winners: %v", g.score)
		}
	}
	check(g.save.Stats.Played >= 1 && g.save.Stats.Hands > 0, "record %+v", g.save.Stats)
	check(len(g.result.rows) == cr.Sides(players), "result rows %d", len(g.result.rows))
	fmt.Printf("%d players: %v after %d hands, %q / %s\n", players, g.score[:g.sides()], g.handNo, g.result.title, g.result.sub)
	return g
}

func main() {
	os.Setenv("CRIBBAGE_DATA", os.TempDir()+"/cribbage-headless-test")
	os.RemoveAll(os.TempDir() + "/cribbage-headless-test")
	autopilot, testFast = true, true

	// 0. faces: names always fit, ladies are ladies
	for i := 0; i < 2000; i++ {
		f := Face{Female: i%2 == 0, Seed: uint64(i)*7919 + 3}
		n, full, ok := f.nameFor(map[string]bool{})
		check(ok && fitsFace(f.String(), n), "face %v got %q", f, full)
		l := f.look()
		check(l.Lady == f.Female, "a lady's face must look like a lady")
		if !f.Female {
			check(l.Lip.A == 0 && l.HairStyle != "bob", "a gentleman's face got a lady's look")
		}
		g2, ok := parseFace(f.String())
		check(ok && g2 == f, "face round trip")
	}

	// 1. single player, two, three and four handed
	for _, n := range []int{2, 3, 4} {
		soloGame(n)
	}

	// 2. online: a host and two guests at a table of three, then four
	host := newGameState()
	host.setup.name, host.setup.face = "Hosty", pickerFaces[1]
	host.mu.Lock()
	host.save.Settings.Players = 3
	host.hostGame()
	check(host.server != nil, "host should be listening")
	check(host.n() == 3, "host table of 3")
	host.mu.Unlock()
	var guests []*game
	for i, nm := range []string{"Ann", "Bo"} {
		c := newGameState()
		c.save.Stats = defaultSave().Stats // everyone here shares one save folder
		c.setup.name, c.setup.face = nm, pickerFaces[2+i]
		c.mu.Lock()
		c.joinGame("127.0.0.1")
		c.mu.Unlock()
		guests = append(guests, c)
		waitFor(nm+" to join", 10*time.Second, func() bool {
			return locked(host, func() bool { return len(host.server.peers) == i+1 }) && locked(c, func() bool { return c.mode == ModeClient })
		})
	}
	host.mu.Lock()
	check(host.players[1].name == "Ann" && host.players[3].name == "Bo" && host.players[2] == nil, "join order at a table of three")
	host.mu.Unlock()
	// a fourth chair: the guests keep their seats and a regular fills the last one
	host.mu.Lock()
	host.setTableSize(4)
	check(host.n() == 4 && host.players[2] != nil && host.players[2].kind == KindAI, "table of four")
	host.setTableSize(3)
	check(host.n() == 3 && host.players[2] == nil, "back to three")
	host.mu.Unlock()
	time.Sleep(150 * time.Millisecond)
	for _, c := range guests {
		c.mu.Lock()
		check(len(c.lobby) == 4 && c.n() == 3, "guest should have the lobby")
		c.mu.Unlock()
	}

	for round := 1; round <= 2; round++ {
		host.mu.Lock()
		if round == 1 {
			host.startHosted()
		} else {
			host.newDeal()
		}
		host.mu.Unlock()
		// while it plays, guests should always know their own cards and nobody else's
		done, crib := false, 0
		for !done {
			time.Sleep(10 * time.Millisecond)
			for _, c := range guests {
				c.mu.Lock()
				me := c.viewer
				for _, card := range c.deck {
					if card.zn.K == zHand && card.zn.P == me {
						check(card.id != "", "guest doesn't know a card in their own hand")
					}
					if card.zn.K == zHand && card.zn.P != me {
						check(card.id == "" || card.fa1 == 180, "guest can see someone else's card")
					}
					if card.zn.K == zCrib {
						check(card.fa1 == 180, "guest can see the crib")
					}
				}
				if c.ask != nil && c.ask.need == 1 {
					crib++
				}
				c.mu.Unlock()
			}
			done = locked(host, func() bool { tableOK(host); return host.showResult })
		}
		check(crib > 0, "guests were never asked for the crib")
		waitFor("guests' results", 10*time.Second, func() bool {
			for _, c := range guests {
				if !locked(c, func() bool { return c.showResult }) {
					return false
				}
			}
			return true
		})
		hs := host.score
		for _, c := range guests {
			c.mu.Lock()
			v := c.viewer
			check(c.score == hs, "guest scores %v vs host %v", c.score, hs)
			check(len(c.result.rows) == 3, "guest result rows")
			check(c.save.Stats.Played == round, "guest record: %d games", c.save.Stats.Played)
			fmt.Printf("round %d: %s sees %q (%s) %v\n", round, c.players[v].name, c.result.title, c.result.sub, c.score[:3])
			c.mu.Unlock()
		}
	}

	// 3. a guest leaves mid-game: a computer player takes over and the game finishes
	host.mu.Lock()
	host.newDeal()
	host.mu.Unlock()
	time.Sleep(400 * time.Millisecond)
	bo := guests[1]
	bo.mu.Lock()
	bo.leaveOnline()
	bo.mu.Unlock()
	waitFor("game after a guest left", 5*time.Minute, func() bool { return locked(host, func() bool { return host.showResult }) })
	host.mu.Lock()
	check(host.players[3].kind == KindAI, "Bo's seat should be a computer now")
	check(len(host.server.peers) == 1, "one guest left")
	host.leaveOnline()
	host.mu.Unlock()
	waitFor("Ann to hear the host closed", 10*time.Second, func() bool {
		return locked(guests[0], func() bool { return guests[0].mode == ModeSolo && guests[0].message != "" })
	})

	// 4. same computer: two people, answered through the real pass screen,
	// crib boxes and card picks
	autopilot = false
	hs := newGameState()
	hs.mu.Lock()
	hs.save.Settings.Target = 61
	hs.setup.hotN = 2
	hs.setup.hotNames = [4]string{"Al", "Bea", "", ""}
	hs.startHotseat()
	check(hs.players[0].name == "Al" && hs.players[2].name == "Bea" && hs.players[1] == nil && hs.players[3] == nil, "hot-seat seating")
	hs.mu.Unlock()
	passScreens, turns, cribs := 0, map[int]int{}, map[int]int{}
	for !locked(hs, func() bool { return hs.showResult }) {
		time.Sleep(3 * time.Millisecond)
		hs.mu.Lock()
		switch {
		case hs.passCh != nil:
			for _, c := range hs.deck {
				check(c.zn.K != zHand || c.fa1 == 180, "a hand is showing behind the pass screen")
			}
			passScreens++
			hs.passCh <- struct{}{}
			hs.passCh = nil
		case hs.ask != nil:
			a := hs.ask
			hs.ask = nil
			check(hs.players[hs.viewer].kind == KindLocal && a.need == 2, "crib box for a computer seat")
			for _, c := range a.all {
				check(c.zn.P == hs.viewer, "laying away from someone else's hand")
			}
			cribs[hs.viewer]++
			perm := hs.rng.Perm(len(a.all))
			a.ch <- askRes{btn: a.btns[0], cards: []*cardV{a.all[perm[0]], a.all[perm[1]]}}
		case hs.wait != nil:
			w := hs.wait
			hs.wait = nil
			for _, c := range w.all {
				check(c.zn.P == hs.viewer, "picking from someone else's hand")
			}
			for _, c := range w.ok {
				check(hs.count+c.c.Value() <= 31, "offered a card past 31")
			}
			turns[hs.viewer]++
			w.ch <- w.ok[hs.rng.IntN(len(w.ok))]
		}
		tableOK(hs)
		hs.mu.Unlock()
	}
	check(passScreens > 3 && turns[0] > 0 && turns[2] > 0 && cribs[0] > 0 && cribs[2] > 0, "hot-seat turns %v cribs %v screens %d", turns, cribs, passScreens)
	fmt.Printf("same computer: %d pass screens, turns %v, cribs %v, %q %s\n", passScreens, turns, cribs, hs.result.title, hs.result.sub)
	fmt.Println("PASS")
}
