package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"sync"
	"time"

	cr "github.com/adamglassner/cribbage/internal/crib"
)

/* ---------------- who's at the table ---------------- */

// Kind says who plays a seat.
type Kind int

const (
	KindAI     Kind = iota // a computer player
	KindLocal              // someone at this computer
	KindRemote             // someone playing over the network
)

// Mode is how this game is being played.
type Mode int

const (
	ModeSolo    Mode = iota // you and the cottage regulars
	ModeHotseat             // 2-4 people sharing this computer
	ModeHost                // hosting an online game
	ModeClient              // playing in someone else's online game
)

// Seats as drawn: the viewer always sits at the bottom (S), the next player
// to their left (W), across (N) and the last to their right (E).
const (
	seatS = 'S'
	seatW = 'W'
	seatN = 'N'
	seatE = 'E'
)

var seatOrder = [4]byte{seatS, seatW, seatN, seatE}

// seatsFor is which of the four chairs a table of n uses, in playing order:
// two players sit across from each other, three leave the far chair empty.
func seatsFor(n int) []int {
	switch n {
	case 3:
		return []int{0, 1, 3}
	case 4:
		return []int{0, 1, 2, 3}
	}
	return []int{0, 2}
}

type player struct {
	i          int
	kind       Kind
	face       string // a Face as text, "" for none
	name, full string // "Ruth", "Nana Ruth"
	hand       []*cardV
	say        string
	sayAt      time.Time
	sayEnd     time.Time
	turnAt     time.Time
	isTurn     bool
	peer       *peer
}

/* ---------------- where a card is ---------------- */

type zoneK int

const (
	zNone    zoneK = iota
	zDeck          // the pack (N = height in the pile)
	zHand          // in player P's hand, position N
	zCrib          // laid away into the crib (N = order)
	zStarter       // the starter, cut and turned on top of the pack
	zPile          // played in the current count by P (N = order)
	zPlayed        // played by P earlier this hand, turned down (N = order)
	zShow          // laid out to be counted (N = position; the starter is 4)
	zDone          // counted, back in front of player P (N = order)
)

type zone struct {
	K zoneK   `json:"k"`
	P int     `json:"p,omitempty"`
	N int     `json:"n,omitempty"`
	J float64 `json:"j,omitempty"` // a little random rotation
}

// faceUp says whether a card in a zone is shown to everyone.
func (z zone) faceUp() bool {
	switch z.K {
	case zStarter, zPile, zShow, zDone:
		return true
	}
	return false
}

type cardV struct {
	c   cr.Card
	id  string // "H11"; "" when this computer doesn't know the face
	idx int
	tok int // what guests call this card (shuffled every hand)
	zn  zone
	// position animation (table coordinates, card centre)
	fx, fy, fr, fs, tx, ty, tr, ts float64
	t0                             time.Time
	dur                            float64
	placed                         bool
	// flip animation: 0 = face up, 180 = face down
	fa0, fa1 float64
	ft0      time.Time
	fdur     float64
	z        int
	speed    int // 0 normal, 1 fast, 2 instant
	pick     bool
	dim      bool
	win      bool // part of what's being counted
	sel      bool // chosen for the crib
	lift     float64
	sl       float64 // selection lift (eased)
	x, y     float64
}

func (c *cardV) pos(now time.Time) (x, y, r, s float64) {
	if c.dur <= 0 {
		return c.tx, c.ty, c.tr, c.ts
	}
	t := easeCard(now.Sub(c.t0).Seconds() / c.dur)
	l := func(a, b float64) float64 { return a + (b-a)*t }
	return l(c.fx, c.tx), l(c.fy, c.ty), l(c.fr, c.tr), l(c.fs, c.ts)
}

func (c *cardV) angle(now time.Time) float64 {
	if c.fdur <= 0 {
		return c.fa1
	}
	t := easeInOut(now.Sub(c.ft0).Seconds() / c.fdur)
	return c.fa0 + (c.fa1-c.fa0)*t
}

/* ---------------- game state ---------------- */

type askBtn struct {
	Label string `json:"label"`
	Sub   string `json:"sub,omitempty"`
	Cls   string `json:"cls,omitempty"`
}

// askT is a question in the box above your hand. With need > 0 it also asks
// you to pick that many cards from all (the cards for the crib).
type askT struct {
	q, sub string
	btns   []askBtn
	need   int
	all    []*cardV
	at     time.Time
	ch     chan askRes
}

type askRes struct {
	btn   askBtn
	idx   int
	cards []*cardV
}

type waitT struct {
	ok  []*cardV
	all []*cardV
	ch  chan *cardV
}

type bannerT struct {
	title, small string
	blue         bool
	at, until    time.Time
}

type particle struct {
	k                        int // 0 peg spark, 1 spark, 2 confetti
	x, y, vx, vy, life, d, s float64
	rot, vr                  float64
	col                      int
}

// pop is a little "Fifteen for 2" that floats up from the cards.
type pop struct {
	text string
	seat int
	at   time.Time
}

// tallyT is the box beside the show that adds a hand up as it's counted.
type tallyT struct {
	title  string
	lines  []tallyLine
	total  int
	done   bool
	at     time.Time
	crib   bool
	nineteen bool
}

type tallyLine struct {
	Kind string `json:"k"`
	N    int    `json:"n"` // how many of them
	Pts  int    `json:"p"`
}

type resRow struct {
	Name  string   `json:"n"`
	Faces []string `json:"f,omitempty"`
	Score int      `json:"s"`
	Place int      `json:"p"`
	Me    bool     `json:"m,omitempty"`
	Skunk int      `json:"k,omitempty"`
}

type resultT struct {
	kick, title, sub string
	rows             []resRow
	rec              string
	won              bool
}

// tallyRec is one seat's haul this game, for guests' records.
type tallyRec struct{ Hands, BestHand, Skunks, TwentyNines int }

type game struct {
	mu sync.Mutex

	save    saveData
	mode    Mode
	viewer  int  // whose seat the table is drawn from
	hidden  bool // shared computer: hands face down between turns
	players [4]*player // nil for an empty chair
	act     []int      // the chairs in use, in playing order
	deck    []*cardV
	byCard  map[cr.Card]*cardV
	byTok   [52]*cardV

	dealer   int // a chair
	handNo   int // hands dealt this game (the first is 0)
	pile     []*cardV
	count    int
	pegging  bool
	score    [4]int // by side
	prev     [4]int // the back peg, by side
	target   int
	winner   int // side, -1 while playing
	handsTot int // hands dealt since the program started (for guests' snapshots)

	gameID   int
	cancel   chan struct{}
	finished bool
	started  bool
	wait     *waitT
	ask      *askT
	remote   map[int]*remoteWait
	prompt   string
	promptAt time.Time
	ban      bannerT
	pace     float64
	parts    []particle
	pops     []pop
	tally    *tallyT

	tallies [4]tallyRec

	// layout, in logical pixels (table space)
	Wd, Ht, cw, ch float64

	// overlays
	showTitle, showRules, showCredits, showResult bool
	showTogether, showHotseat, showOnline         bool
	showLobby                                     bool
	passTo                                        int
	passCh                                        chan struct{}
	message                                       string
	result                                        resultT
	resultAt                                      time.Time
	ui                                            uiState

	setup  setupState
	focus  string // text field being typed in
	server *server
	client *client
	lobby  []lobbySeat
	cliEnd chan struct{} // closed to drop a guest's pending ask

	nextFolks []folk // who sits down for the next single-player game

	rng *rand.Rand
}

var (
	errCancel = fmt.Errorf("cancelled")
	errWon    = fmt.Errorf("someone pegged out")
)

func newGameState() *game {
	seed := uint64(time.Now().UnixNano())
	g := &game{rng: rand.New(rand.NewPCG(seed, seed>>7^0x5eed)), pace: 1, showTitle: true, winner: -1, remote: map[int]*remoteWait{}}
	g.save = loadSave()
	g.setup = newSetup(&g.save)
	g.target = g.save.Settings.Target
	g.resetDeck(false)
	g.soloTable()
	g.Wd, g.Ht = 1400, 820
	g.measure()
	return g
}

// resetDeck makes the 52 cards. A guest's cards start out unknown.
func (g *game) resetDeck(guest bool) {
	g.deck = nil
	g.byCard = map[cr.Card]*cardV{}
	for i, c := range cr.Deck() {
		v := &cardV{c: c, id: c.ID(), idx: i, tok: i, ts: 1, fs: 1, fa1: 180}
		if guest {
			v.id = ""
		}
		g.deck = append(g.deck, v)
		g.byCard[c] = v
		g.byTok[i] = v
	}
}

func (g *game) rand(n int) int      { return g.rng.IntN(n) }
func pickS[T any](g *game, a []T) T { return a[g.rand(len(a))] }

/* pacing */
var paceOf = map[string]float64{"relaxed": 1.35, "normal": 1, "brisk": 0.6}

func (g *game) T(k string) float64 {
	v := map[string]float64{"deal": 70, "move": 420, "play": 520, "think": 650, "call": 620, "quip": 1900, "banner": 1700}[k]
	v *= g.pace
	if testFast {
		v *= 0.04
	}
	return v
}

// sleep releases the table while waiting; a new game cancels the old one.
func (g *game) sleep(ms float64) {
	id, cancel := g.gameID, g.cancel
	g.mu.Unlock()
	select {
	case <-time.After(time.Duration(ms * float64(time.Millisecond))):
	case <-cancel:
	}
	g.mu.Lock()
	if id != g.gameID {
		panic(errCancel)
	}
}

// waitOn releases the table until ch delivers (or the game is cancelled).
func waitOn[T any](g *game, ch chan T) T {
	id, cancel := g.gameID, g.cancel
	g.mu.Unlock()
	var v T
	select {
	case v = <-ch:
	case <-cancel:
	}
	g.mu.Lock()
	if id != g.gameID {
		panic(errCancel)
	}
	return v
}

/* ---------------- seats and sides ---------------- */

func (g *game) n() int { return len(g.act) }

// k is a chair's place in playing order.
func (g *game) k(i int) int {
	for k, s := range g.act {
		if s == i {
			return k
		}
	}
	return 0
}

// next is the chair after i in playing order.
func (g *game) next(i int) int { return g.act[(g.k(i)+1)%g.n()] }

func (g *game) side(i int) int { return cr.SideOf(g.k(i), g.n()) }
func (g *game) sides() int     { return cr.Sides(g.n()) }

// seated is everyone at the table, in playing order.
func (g *game) seated() []*player {
	var out []*player
	for _, i := range g.act {
		if p := g.players[i]; p != nil {
			out = append(out, p)
		}
	}
	return out
}

// sideSeats is the chairs on side s.
func (g *game) sideSeats(s int) []int {
	var out []int
	for _, i := range g.act {
		if g.side(i) == s {
			out = append(out, i)
		}
	}
	return out
}

/* ---------------- names, as each seat sees them ---------------- */

// who is how player i is named to the person in seat v ("you" or a name).
func (g *game) who(i, v int) string {
	if i == v && g.mode != ModeHotseat {
		return "you"
	}
	return g.players[i].name
}

func (g *game) Who(i, v int) string {
	s := g.who(i, v)
	if s == "you" {
		return "You"
	}
	return s
}

// whose: "your" or "Ruth's".
func (g *game) whose(i, v int) string {
	if g.who(i, v) == "you" {
		return "your"
	}
	return g.players[i].name + "'s"
}

func (g *game) Whose(i, v int) string {
	s := g.whose(i, v)
	if s == "your" {
		return "Your"
	}
	return s
}

// verb picks "take" or "takes" for who(i, v).
func (g *game) verb(i, v int, you, them string) string {
	if g.who(i, v) == "you" {
		return you
	}
	return them
}

// sideName is how side s is named to the person in seat v ("You", "Ruth", "You & Eli").
func (g *game) sideName(s, v int) string {
	var ns []string
	seats := g.sideSeats(s)
	sort.SliceStable(seats, func(a, b int) bool { return seats[a] == v })
	for _, i := range seats {
		ns = append(ns, g.Who(i, v))
	}
	return strings.Join(ns, " & ")
}

func (g *game) sideFull(s, v int) string {
	var ns []string
	seats := g.sideSeats(s)
	sort.SliceStable(seats, func(a, b int) bool { return seats[a] == v })
	for _, i := range seats {
		nm := g.players[i].full
		if g.who(i, v) == "you" {
			nm = "You"
		}
		ns = append(ns, nm)
	}
	return strings.Join(ns, " & ")
}

// mySide reports whether side s includes seat v (as a person, not on a shared computer).
func (g *game) mySide(s, v int) bool {
	return g.mode != ModeHotseat && v >= 0 && v < 4 && g.players[v] != nil && g.side(v) == s
}

/* ---------------- layout ---------------- */

func (g *game) measure() {
	g.ch = math.Max(92, math.Min(math.Min(g.Ht*0.2, g.Wd*0.115), 190))
	g.cw = g.ch * cardW / cardH
}

func (g *game) cy() float64 { return g.Ht * 0.47 }

func (g *game) seatOf(i int) byte { return seatOrder[((i-g.viewer)%4+4)%4] }

func (g *game) anchor(seat byte) (float64, float64) {
	switch seat {
	case seatS:
		return g.Wd / 2, g.Ht - g.ch*0.58 - 4
	case seatN:
		return g.Wd / 2, g.ch*0.42 + 2
	case seatW:
		return g.ch*0.5 + 4, g.cy()
	}
	return g.Wd - g.ch*0.5 - 4, g.cy()
}

func (g *game) handSpread(seat byte) float64 {
	switch seat {
	case seatS:
		return math.Min(g.cw*0.62, g.Wd*0.5/7)
	case seatN:
		return math.Min(g.cw*0.36, g.Wd*0.34/7)
	}
	return math.Min(g.cw*0.34, g.Ht*0.3/7)
}

func (g *game) handPos(seat byte, i, n int) (x, y, r, s float64) {
	ax, ay := g.anchor(seat)
	k := float64(i) - float64(n-1)/2
	sp := g.handSpread(seat)
	switch seat {
	case seatS:
		return ax + k*sp, ay + k*k*1.6, k * 2.8, 1
	case seatN:
		return ax - k*sp, ay - k*k*0.6, 180 - k*2.4, 0.8
	}
	dir := 1.0
	if seat == seatE {
		dir = -1
	}
	return ax - dir*k*k*0.6, ay + dir*k*sp, dir*90 + dir*k*2.4, 0.8
}

// toward moves (x, y) a fraction f of the way from the table centre to seat's anchor.
func (g *game) toward(seat byte, f float64) (float64, float64) {
	ax, ay := g.anchor(seat)
	return g.Wd/2 + (ax-g.Wd/2)*f, g.cy() + (ay-g.cy())*f
}

func (g *game) packPos() (float64, float64) { return g.Wd/2 - g.cw*2.25, g.cy() }

func seatRot(seat byte) float64 {
	switch seat {
	case seatN:
		return 180
	case seatW:
		return 90
	case seatE:
		return -90
	}
	return 0
}

// cribPos is where the crib sits: by the dealer, to their right.
func (g *game) cribPos() (float64, float64, float64) {
	seat := g.seatOf(g.dealer)
	ax, ay := g.anchor(seat)
	switch seat {
	case seatS:
		return ax + 3.2*g.handSpread(seat) + g.cw*1.05, ay + g.ch*0.04, 4
	case seatN:
		return ax - 3.2*g.handSpread(seat) - g.cw*1.0, ay + g.ch*0.06, 176
	}
	// the side chairs keep it above their hand, but clear of the cribbage
	// board (top left) and the hand box (top right)
	y := math.Max(ay-3*g.handSpread(seat)-g.cw*1.25, 14+36+30*float64(g.sides())+g.cw*0.4+10)
	if seat == seatW {
		return ax + g.ch*0.12, y, 94
	}
	return ax - g.ch*0.12, y, -86
}

// zonePos is where a card in a zone sits, as the viewer sees the table.
func (g *game) zonePos(c *cardV) (x, y, r, s float64, up bool, z int) {
	zn := c.zn
	N := float64(zn.N)
	switch zn.K {
	case zDeck:
		px, py := g.packPos()
		return px + N*0.1, py - N*0.18, zn.J, 0.9, false, 20 + zn.N
	case zStarter:
		px, py := g.packPos()
		return px + 52*0.1 + 3, py - 52*0.18 - 2, zn.J, 0.9, true, 80
	case zHand:
		p := g.players[zn.P]
		if p == nil {
			return c.tx, c.ty, c.tr, c.ts, false, c.z
		}
		x, y, r, s := g.handPos(g.seatOf(zn.P), zn.N, len(p.hand))
		return x, y, r, s, zn.P == g.viewer && !g.hidden, 100 + zn.N
	case zCrib:
		x, y, r := g.cribPos()
		return x + N*1.5, y - N*1.5, r + (N-1.5)*3 + zn.J, 0.62, false, 60 + zn.N
	case zPile:
		n := 0
		for _, x := range g.deck {
			if x.zn.K == zPile {
				n++
			}
		}
		step := math.Min(g.cw*0.42, g.Wd*0.34/math.Max(1, float64(n)))
		x0 := g.Wd/2 - g.cw*0.85
		// lean a little toward whoever played it
		px, py := g.toward(g.seatOf(zn.P), 0.08)
		dx, dy := px-g.Wd/2, py-g.cy()
		return x0 + N*step + dx*0.25, g.cy() + dy*0.6, zn.J + (seatRot(g.seatOf(zn.P)))*0.03, 0.92, true, 300 + zn.N
	case zPlayed, zDone:
		seat := g.seatOf(zn.P)
		tx, ty := g.toward(seat, 0.52)
		rot := seatRot(seat)
		var dx, dy float64
		switch seat {
		case seatS:
			tx -= g.cw * 1.6
			dx = N * g.cw * 0.18
		case seatN:
			tx += g.cw * 1.6
			dx = -N * g.cw * 0.18
		case seatW:
			ty -= g.cw * 1.0
			dy = N * g.cw * 0.18
		default:
			ty -= g.cw * 1.0
			dy = N * g.cw * 0.18
		}
		up := zn.K == zDone
		return tx + dx, ty + dy, rot + zn.J, 0.5, up, 30 + zn.N
	case zShow:
		y := g.cy() - g.ch*0.04
		if zn.N == 4 { // the starter, set a little apart
			return g.Wd/2 + g.cw*1.7, y - g.ch*0.06, 3, 1.02, true, 420
		}
		return g.Wd/2 - g.cw*2.3 + N*g.cw*0.92, y, (N - 1.5) * 1.5, 1.02, true, 400 + zn.N
	}
	return c.tx, c.ty, c.tr, c.ts, false, c.z
}

// layoutCards moves every card to where its zone says (animated unless instant).
func (g *game) layoutCards(instant bool) {
	for _, c := range g.deck {
		if c.zn.K == zNone {
			continue
		}
		x, y, r, s, up, z := g.zonePos(c)
		sp := c.speed
		if instant {
			c.speed = 2
		}
		if !c.placed || x != c.tx || y != c.ty || r != c.tr || s != c.ts || z != c.z {
			g.place(c, x, y, r, z, s)
		}
		g.flip(c, up && c.id != "")
		c.speed = sp
	}
}

func (g *game) place(c *cardV, x, y, r float64, z int, scale float64) {
	now := time.Now()
	c.fx, c.fy, c.fr, c.fs = c.pos(now)
	c.tx, c.ty, c.tr, c.ts = x, y, r, scale
	c.x, c.y = x, y
	c.t0 = now
	switch c.speed {
	case 0:
		c.dur = 0.42
	case 1:
		c.dur = 0.26
	default:
		c.dur = 0
	}
	if !c.placed { // first placement: appear in place rather than fly in from the corner
		c.dur, c.placed = 0, true
	}
	if z >= 0 {
		c.z = z
	}
}

func (g *game) flip(c *cardV, up bool) {
	now := time.Now()
	target := 180.0
	if up {
		target = 0
	}
	if c.fa1 == target {
		return
	}
	c.fa0, c.fa1, c.ft0 = c.angle(now), target, now
	c.fdur = 0.36
	if c.speed == 2 {
		c.fdur = 0
	}
}

// setHand gives every card in p's hand its zone (after the hand changes).
func (g *game) setHand(p *player) {
	for i, c := range p.hand {
		c.zn = zone{K: zHand, P: p.i, N: i}
	}
}

func cardsOf(h []*cardV) []cr.Card {
	out := make([]cr.Card, len(h))
	for i, c := range h {
		out[i] = c.c
	}
	return out
}

func (g *game) sortHand(p *player) {
	cs := cardsOf(p.hand)
	cr.SortHand(cs)
	for i, c := range cs {
		p.hand[i] = g.byCard[c]
	}
	g.setHand(p)
}

// sync lays the table out here and sends every guest their view of it.
func (g *game) sync() {
	g.layoutCards(false)
	if g.server != nil {
		g.server.sendSnaps(g, false)
	}
}

func (g *game) setTurn(i int) {
	for _, p := range g.players {
		if p == nil {
			continue
		}
		on := p.i == i
		if on && !p.isTurn {
			p.turnAt = time.Now()
		}
		p.isTurn = on
	}
}

func (g *game) setPrompt(t string) {
	if t != g.prompt {
		g.promptAt = time.Now()
	}
	g.prompt = t
}

func (g *game) banner(title, small string, blue bool, ms float64) {
	if ms == 0 {
		ms = 1700 * g.pace
		if testFast {
			ms *= 0.2
		}
	}
	now := time.Now()
	g.ban = bannerT{title, small, blue, now, now.Add(time.Duration(ms) * time.Millisecond)}
}

/* ---------------- events everyone sees (each worded for their seat) ---------------- */

type event struct {
	K    string      `json:"k"`
	P    int         `json:"p,omitempty"`
	N    int         `json:"n,omitempty"`
	Text string      `json:"text,omitempty"`
	Sub  string      `json:"sub,omitempty"`
	Blue bool        `json:"blue,omitempty"`
	Ms   float64     `json:"ms,omitempty"`
	L    []tallyLine `json:"l,omitempty"`
}

// emit shows an event here and sends it to every guest, worded for their seat.
func (g *game) emit(f func(v int) *event) {
	if e := f(g.viewer); e != nil {
		g.apply(e)
	}
	if g.server != nil {
		for _, pr := range g.server.peers {
			if e := f(pr.seat); e != nil {
				pr.send(Msg{T: "ev", Ev: e})
			}
		}
	}
}

func (g *game) emitAll(e event) { g.emit(func(int) *event { c := e; return &c }) }

func (g *game) sound(name string) { g.emitAll(event{K: "snd", Text: name}) }

func (g *game) apply(e *event) {
	switch e.K {
	case "snd":
		snd.play(e.Text)
	case "say":
		if e.P < 0 || e.P > 3 || g.players[e.P] == nil {
			return
		}
		p := g.players[e.P]
		now := time.Now().Add(time.Duration(e.Ms) * time.Millisecond)
		p.say, p.sayAt, p.sayEnd = e.Text, now, now.Add(time.Duration(g.T("quip")+500)*time.Millisecond)
	case "banner":
		g.banner(e.Text, e.Sub, e.Blue, e.Ms)
	case "pop":
		if e.P >= 0 && e.P < 4 {
			g.pops = append(g.pops, pop{text: e.Text, seat: e.P, at: time.Now()})
		}
	case "spark":
		x, y := g.pileEnd()
		g.burst(x, y, 1, e.N)
	case "confetti":
		g.burst(0, 0, 2, e.N)
	case "prompt":
		g.setPrompt(e.Text)
	case "tally":
		g.tally = &tallyT{title: e.Text, at: time.Now(), crib: e.Blue}
	case "tallyAdd":
		if t := g.tally; t != nil {
			t.lines, t.total = e.L, e.N
		}
	case "tallyEnd":
		if t := g.tally; t != nil {
			t.lines, t.total, t.done = e.L, e.N, true
			t.nineteen = e.N == 0
		}
	case "tallyOff":
		g.tally = nil
	}
}

// pileEnd is where the last card played sits.
func (g *game) pileEnd() (float64, float64) {
	var last *cardV
	for _, c := range g.deck {
		if c.zn.K == zPile && (last == nil || c.zn.N > last.zn.N) {
			last = c
		}
	}
	if last == nil {
		return g.Wd / 2, g.cy()
	}
	return last.tx, last.ty
}

// say has a computer player say one of their lines; {p} becomes about's name.
func (g *game) say(p *player, kind string, chance float64, about int, delay float64) {
	if p == nil {
		return
	}
	f, ok := parseFace(p.face)
	if p.kind != KindAI || !ok || g.rng.Float64() > chance {
		return
	}
	ls := f.lines(kind)
	if len(ls) == 0 {
		return
	}
	line := pickS(g, ls)
	if about >= 0 && about < 4 && g.players[about] != nil {
		n := g.players[about].name
		if n == "You" || n == "" {
			line = strings.ReplaceAll(line, ", {p}", "")
		}
		line = strings.ReplaceAll(line, "{p}", n)
	}
	g.emitAll(event{K: "say", P: p.i, Text: line, Ms: delay})
}

// speak puts words in anyone's bubble ("Go!").
func (g *game) speak(i int, text string) { g.emitAll(event{K: "say", P: i, Text: text}) }

/* ---------------- effects ---------------- */

func (g *game) burst(x, y float64, kind int, n int) {
	R := g.rng.Float64
	for i := 0; i < n; i++ {
		switch kind {
		case 1:
			a, v := R()*math.Pi*2, 1.5+R()*4
			g.parts = append(g.parts, particle{k: 1, x: x, y: y, vx: math.Cos(a) * v, vy: math.Sin(a)*v - 1, life: 1, d: 0.02 + R()*0.02, s: 2 + R()*3})
		case 2:
			g.parts = append(g.parts, particle{k: 2, x: R() * g.Wd, y: -20 - R()*g.Ht*0.5, vx: (R() - .5) * 2, vy: 2 + R()*2.5, life: 1, d: 0.003, s: 6 + R()*6, rot: R() * 6, vr: (R() - .5) * .25, col: g.rand(6)})
		}
	}
}

/* ---------------- decisions: here, over the network, or the computer ---------------- */

func (g *game) chooseCard(ok, all []*cardV, text string) *cardV {
	g.setPrompt(text)
	in := func(c *cardV) bool {
		for _, x := range ok {
			if x == c {
				return true
			}
		}
		return false
	}
	for _, c := range all {
		c.pick, c.dim = in(c), !in(c)
	}
	w := &waitT{ok: ok, all: all, ch: make(chan *cardV, 1)}
	g.wait = w
	defer func() {
		for _, x := range all {
			x.pick, x.dim = false, false
		}
		g.setPrompt("")
	}()
	return waitOn(g, w.ch)
}

func (g *game) askUser(a *askT) askRes {
	a.at, a.ch = time.Now(), make(chan askRes, 1)
	g.ask = a
	for _, c := range a.all {
		c.pick, c.sel = true, false
	}
	defer func() {
		for _, c := range a.all {
			c.pick = false
		}
	}()
	return waitOn(g, a.ch)
}

// localTurn is true when someone at this computer decides for seat i.
func (g *game) localTurn(i int) bool { return g.players[i].kind == KindLocal && !autopilot }

// ensureView turns a shared computer over to the player in seat i: a
// "pass the computer" screen while everyone's cards are face down, then the
// table turns to their seat.
func (g *game) ensureView(i int) {
	if g.mode != ModeHotseat || (g.viewer == i && !g.hidden) {
		return
	}
	if g.viewer != i {
		g.hidden = true
		g.sync()
		g.passTo = i
		ch := make(chan struct{}, 1)
		g.passCh = ch
		func() {
			defer func() { g.passCh = nil }()
			waitOn(g, ch)
		}()
		g.viewer = i
	}
	g.hidden = false
	g.sync()
	g.sleep(g.T("move"))
}

func (g *game) hideAfter() {
	if g.mode == ModeHotseat {
		g.hidden = true
		g.sync()
	}
}

func (g *game) decideCard(i int, ok []*cardV, text string, think float64, ai func() *cardV) *cardV {
	p := g.players[i]
	switch {
	case g.localTurn(i):
		g.ensureView(i)
		c := g.chooseCard(ok, p.hand, text)
		return c
	case p.kind == KindRemote:
		a := &wireAsk{K: "card", Prompt: text}
		for _, c := range ok {
			a.Toks = append(a.Toks, c.tok)
		}
		if r, okk := g.remoteAsk(i, a); okk {
			if c := g.byTok[r.Tok]; c != nil {
				for _, x := range ok {
					if x == c {
						return c
					}
				}
			}
		}
	}
	g.sleep(think)
	return ai()
}

/* ---------------- the regulars ---------------- */

// tableSize is how many play a single-player game.
func (g *game) tableSize() int {
	n := g.save.Settings.Players
	if n < 2 || n > 4 {
		n = 2
	}
	return n
}

// soloTable seats you with the regulars the title card shows and picks who comes next time.
func (g *game) soloTable() {
	g.mode, g.viewer, g.hidden = ModeSolo, 0, false
	n := g.tableSize()
	g.act = seatsFor(n)
	if len(g.nextFolks) != n-1 {
		g.rollNext()
	}
	g.players = [4]*player{}
	g.players[0] = &player{i: 0, kind: KindLocal, name: "You", full: "You"}
	for k, seat := range g.act[1:] {
		f := g.nextFolks[k]
		g.players[seat] = &player{i: seat, kind: KindAI, face: f.face, name: f.name, full: f.full}
	}
	g.rollNext()
}

func (g *game) rollNext() {
	g.nextFolks = randomFolks(g.rng, g.tableSize()-1, map[string]bool{}, map[string]bool{"You": true})
}

// fillFolks gives every empty or computer seat at the table a random regular.
func (g *game) fillFolks() {
	faces, names := map[string]bool{}, map[string]bool{}
	for _, p := range g.players {
		if p != nil && p.kind != KindAI {
			faces[p.face], names[p.name] = true, true
		}
	}
	var open []int
	for _, i := range g.act {
		if p := g.players[i]; p == nil || p.kind == KindAI {
			open = append(open, i)
		}
	}
	for i := range g.players { // chairs not at this table stay empty
		in := false
		for _, s := range g.act {
			in = in || s == i
		}
		if !in {
			g.players[i] = nil
		}
	}
	if len(open) == 0 {
		return
	}
	fs := randomFolks(g.rng, len(open), faces, names)
	for k, i := range open {
		f := fs[k%len(fs)]
		g.players[i] = &player{i: i, kind: KindAI, face: f.face, name: f.name, full: f.full}
	}
}

func (g *game) setPlayer(i int, kind Kind, face, name string) *player {
	p := &player{i: i, kind: kind, face: face, name: name, full: name}
	g.players[i] = p
	return p
}

/* ---------------- a game ---------------- */

// newGame is called with the lock held.
func (g *game) newGame() {
	g.gameID++
	if g.cancel != nil {
		close(g.cancel)
	}
	g.cancel = make(chan struct{})
	g.finished, g.started = false, true
	g.pace = paceOf[g.save.Settings.Speed]
	if g.pace == 0 {
		g.pace = 1
	}
	if g.mode == ModeSolo { // new regulars every game
		g.soloTable()
	}
	g.target = g.save.Settings.Target
	g.showResult, g.wait, g.ask, g.passCh = false, nil, nil, nil
	g.remote = map[int]*remoteWait{}
	g.prompt = ""
	g.ban = bannerT{}
	g.parts, g.pops, g.tally = nil, nil, nil
	for _, p := range g.players {
		if p != nil {
			p.hand, p.isTurn, p.say = nil, false, ""
		}
	}
	if g.mode == ModeHotseat {
		g.viewer, g.hidden = 0, true
	}
	for _, c := range g.deck {
		c.speed, c.pick, c.dim, c.win, c.sel = 2, false, false, false, false
		c.zn = zone{}
	}
	g.score, g.prev = [4]int{}, [4]int{}
	g.tallies = [4]tallyRec{}
	g.handNo, g.winner = 0, -1
	g.dealer = g.act[g.rand(g.n())]
	g.pile, g.count, g.pegging = nil, 0, false
	if g.server != nil {
		g.server.open = false
		g.server.sendSnaps(g, true)
	}
	go g.run(g.gameID)
}

func (g *game) run(id int) {
	g.mu.Lock()
	defer func() {
		r := recover()
		if r == errWon {
			func() {
				defer func() {
					if r2 := recover(); r2 != nil && r2 != errCancel {
						g.mu.Unlock()
						panic(r2)
					}
				}()
				g.endGame()
			}()
		} else if r != nil && r != errCancel {
			g.mu.Unlock()
			panic(r)
		}
		g.mu.Unlock()
	}()
	if id != g.gameID {
		return
	}
	g.cutForDeal()
	for {
		g.playHand()
		g.handNo++
		g.dealer = g.next(g.dealer)
	}
}

// cutForDeal shows who deals first.
func (g *game) cutForDeal() {
	d := g.dealer
	g.emit(func(v int) *event {
		return &event{K: "banner", Text: g.Who(d, v) + " " + g.verb(d, v, "Deal", "Deals") + " First", Sub: "The deal goes round to the left after every hand.", Blue: true}
	})
	g.sleep(g.T("banner") * 0.9)
}

func (g *game) playHand() {
	g.pile, g.count, g.pegging = nil, 0, false
	g.tally = nil
	g.emitAll(event{K: "tallyOff"})
	g.handsTot++
	for _, p := range g.seated() {
		p.hand = nil
	}
	g.setTurn(-1)
	// guests know a card only by a token that changes every hand
	for i, t := range g.rng.Perm(52) {
		g.deck[i].tok = t
		g.byTok[t] = g.deck[i]
	}
	deck := append([]*cardV(nil), g.deck...)
	g.rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	for i, c := range deck {
		c.speed, c.win, c.pick, c.dim, c.sel = 1, false, false, false, false
		c.zn = zone{K: zDeck, N: 51 - i, J: (g.rng.Float64() - .5) * 4}
	}
	g.sync()
	g.sound("shuffle")
	d := g.dealer
	g.emit(func(v int) *event {
		return &event{K: "prompt", Text: g.Who(d, v) + " " + g.verb(d, v, "deal", "deals") + "…"}
	})
	g.sleep(g.T("move") * 1.6)
	dl := cr.DealFor(g.n())
	k := 0
	for r := 0; r < dl.Cards; r++ {
		for j := 1; j <= g.n(); j++ {
			p := g.players[g.act[(g.k(g.dealer)+j)%g.n()]]
			p.hand = append(p.hand, deck[k])
			k++
			g.setHand(p)
			g.sync()
			g.sound("deal")
			g.sleep(g.T("deal") * 1.5)
		}
	}
	for _, p := range g.seated() {
		g.sortHand(p)
	}
	for _, c := range deck {
		c.speed = 0
	}
	g.sync()
	g.emitAll(event{K: "prompt"})
	g.sleep(g.T("move"))
	g.discards(deck[k : k+dl.FromPack])
	g.cutStarter()
	g.pegPlay()
	g.show()
}

/* ---------------- laying away to the crib ---------------- */

// cribQuestion is the crib box for seat i.
func (g *game) cribQuestion(i int, need int) (string, string) {
	what := "two cards"
	if need == 1 {
		what = "one card"
	}
	whose := g.players[g.dealer].name + "'s crib"
	switch {
	case g.dealer == i:
		whose = "your crib"
	case g.side(g.dealer) == g.side(i):
		whose = "your partner " + g.players[g.dealer].name + "'s crib"
	}
	sub := "Click " + what + ", then Lay Away."
	if g.side(g.dealer) == g.side(i) {
		sub = "The crib is yours: lay away cards that score. " + sub
	} else {
		sub = "The crib isn't yours: give away as little as you can. " + sub
	}
	return "Lay away " + what + " for " + whose, sub
}

func (g *game) discards(fromPack []*cardV) {
	need := cr.DealFor(g.n()).Discard
	var chosen [4][]*cardV
	// ask everyone playing over the network at once
	for _, i := range g.act {
		p := g.players[i]
		if p.kind == KindRemote && p.peer != nil {
			q, sub := g.cribQuestion(i, need)
			a := &wireAsk{K: "crib", Q: q, Sub: sub, Need: need, Btns: []askBtn{{Label: "Lay Away"}}}
			for _, c := range p.hand {
				a.Toks = append(a.Toks, c.tok)
			}
			g.remoteSend(i, a)
		}
	}
	g.setTurn(-1)
	for _, i := range g.act {
		p := g.players[i]
		if !g.localTurn(i) {
			continue
		}
		g.ensureView(i)
		q, sub := g.cribQuestion(i, need)
		r := g.askUser(&askT{q: q, sub: sub, btns: []askBtn{{Label: "Lay Away"}}, need: need, all: p.hand})
		chosen[i] = r.cards
		for _, c := range r.cards {
			c.sel = false
		}
		g.hideAfter()
	}
	for _, i := range g.act {
		p := g.players[i]
		if p.kind == KindRemote && p.peer != nil && chosen[i] == nil {
			g.setPrompt("Waiting for " + p.name + " to lay away…")
			if r, ok := g.remoteWait(i); ok && len(r.Toks) == need {
				var cs []*cardV
				seen := map[*cardV]bool{}
				for _, t := range r.Toks {
					if t >= 0 && t < 52 {
						c := g.byTok[t]
						if c.zn.K == zHand && c.zn.P == i && !seen[c] {
							cs, seen[c] = append(cs, c), true
						}
					}
				}
				if len(cs) == need {
					chosen[i] = cs
				}
			}
			g.setPrompt("")
		}
	}
	for _, i := range g.act {
		p := g.players[i]
		if len(chosen[i]) != need {
			chosen[i] = nil
			own := g.side(i) == g.side(g.dealer)
			for _, c := range cr.ChooseDiscard(cardsOf(p.hand), need, own) {
				chosen[i] = append(chosen[i], g.byCard[c])
			}
		}
	}
	// into the crib, by the dealer
	n := 0
	for _, i := range g.act {
		p := g.players[i]
		for _, c := range chosen[i] {
			p.hand = without(p.hand, c)
			c.zn = zone{K: zCrib, N: n, J: (g.rng.Float64() - .5) * 6}
			n++
		}
		g.setHand(p)
	}
	for _, c := range fromPack { // three players: one off the pack
		c.zn = zone{K: zCrib, N: n, J: (g.rng.Float64() - .5) * 6}
		n++
	}
	g.sync()
	g.sound("pass")
	g.sleep(g.T("move") * 1.6)
}

func without(h []*cardV, c *cardV) []*cardV {
	out := make([]*cardV, 0, len(h))
	for _, x := range h {
		if x != c {
			out = append(out, x)
		}
	}
	return out
}

/* ---------------- the cut ---------------- */

func (g *game) cutStarter() {
	var pack []*cardV
	for _, c := range g.deck {
		if c.zn.K == zDeck {
			pack = append(pack, c)
		}
	}
	// the player on the dealer's right cuts; any card of the pack (but the bottom few) can come up
	cutter := g.act[(g.k(g.dealer)+g.n()-1)%g.n()]
	st := pack[4+g.rand(len(pack)-8)]
	g.emit(func(v int) *event {
		return &event{K: "prompt", Text: g.Who(cutter, v) + " " + g.verb(cutter, v, "cut", "cuts") + " the pack…"}
	})
	g.sleep(g.T("move") * 1.4)
	st.zn = zone{K: zStarter, J: (g.rng.Float64() - .5) * 6}
	st.speed = 0
	g.sync()
	g.sound("flip")
	g.emitAll(event{K: "prompt"})
	g.sleep(g.T("move") * 1.2)
	if st.c.R == 11 {
		d := g.dealer
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "His Heels!", Sub: "A jack is cut: " + g.Who(d, v) + " " + g.verb(d, v, "peg", "pegs") + " 2.", Blue: !g.mySide(g.side(d), v)}
		})
		g.say(g.players[d], "heels", 1, -1, 0)
		g.peg(d, 2, "His heels")
		g.sleep(g.T("banner"))
	}
}

func (g *game) starter() *cardV {
	for _, c := range g.deck {
		if c.zn.K == zStarter || (c.zn.K == zShow && c.zn.N == 4) {
			return c
		}
	}
	return nil
}

/* ---------------- scoring ---------------- */

// peg moves side(i)'s pegs on; reaching the target ends the game at once.
func (g *game) peg(i, pts int, why string) {
	if pts <= 0 {
		return
	}
	s := g.side(i)
	g.prev[s] = g.score[s]
	g.score[s] = min(g.score[s]+pts, g.target)
	g.sync()
	g.sound("peg")
	if g.score[s] >= g.target {
		g.winner = s
		g.sleep(g.T("move") * 1.5)
		panic(errWon)
	}
}

/* ---------------- pegging ---------------- */

func (g *game) pegPlay() {
	g.pegging = true
	defer func() { g.pegging = false }()
	g.pile, g.count = nil, 0
	played := map[int]int{} // cards turned down per chair
	turn := g.next(g.dealer)
	last := -1
	said := map[int]bool{} // who has said "Go" this count
	var seen []cr.Card
	reset := func() {
		for _, c := range g.pile {
			c.zn = zone{K: zPlayed, P: c.zn.P, N: played[c.zn.P], J: (g.rng.Float64() - .5) * 6}
			played[c.zn.P]++
		}
		g.pile, g.count = nil, 0
		said = map[int]bool{}
		g.sync()
		g.sound("sweep")
		g.sleep(g.T("move"))
	}
	left := func() int {
		n := 0
		for _, p := range g.seated() {
			n += len(p.hand)
		}
		return n
	}
	for left() > 0 {
		p := g.players[turn]
		if len(p.hand) == 0 {
			if g.nobodyCanPlay() {
				g.goPoint(last)
				reset()
				turn = g.next(last)
				continue
			}
			turn = g.next(turn)
			continue
		}
		var ok []*cardV
		for _, c := range p.hand {
			if g.count+c.c.Value() <= 31 {
				ok = append(ok, c)
			}
		}
		if len(ok) == 0 {
			if !said[turn] {
				said[turn] = true
				g.setTurn(turn)
				g.sync()
				g.speak(turn, "Go!")
				g.sound("knock")
				g.sleep(g.T("play") * 1.1)
			}
			if g.nobodyCanPlay() {
				g.goPoint(last)
				reset()
				turn = g.next(last)
				continue
			}
			turn = g.next(turn)
			continue
		}
		g.setTurn(turn)
		g.sync()
		text := "Your lead · play any card"
		if len(g.pile) > 0 {
			text = "The count is " + itoa(g.count)
			if len(ok) < len(p.hand) {
				text += " · only cards up to " + itoa(31-g.count) + " fit"
			}
		}
		c := g.decideCard(turn, ok, text, g.T("think")*(0.7+g.rng.Float64()*0.5), func() *cardV {
			st := cr.PegState{Hand: cardsOf(p.hand), Pile: cardsOf(g.pile), Seen: append(append([]cr.Card(nil), seen...), g.starter().c)}
			return g.byCard[cr.ChoosePeg(st)]
		})
		fits := false
		for _, x := range ok {
			fits = fits || x == c
		}
		if !fits {
			panic("illegal play " + c.id)
		}
		p.hand = without(p.hand, c)
		g.pile = append(g.pile, c)
		g.count += c.c.Value()
		seen = append(seen, c.c)
		c.zn = zone{K: zPile, P: turn, N: len(g.pile) - 1, J: (g.rng.Float64() - .5) * 8}
		g.setHand(p)
		g.setTurn(-1)
		g.sync()
		g.sound("play")
		last = turn
		pts, parts := cr.PegScore(cardsOf(g.pile))
		if pts > 0 {
			g.callPeg(turn, pts, parts)
		} else {
			g.sleep(g.T("play"))
		}
		if g.count == 31 {
			g.sleep(g.T("move") * 0.6)
			reset()
		}
		turn = g.next(turn)
		if g.mode == ModeHotseat && left() > 0 {
			g.hideAfter()
		}
	}
	if len(g.pile) > 0 {
		g.goPoint(last)
	}
	g.sleep(g.T("move"))
	reset()
}

func (g *game) nobodyCanPlay() bool {
	for _, p := range g.seated() {
		for _, c := range p.hand {
			if g.count+c.c.Value() <= 31 {
				return false
			}
		}
	}
	return true
}

// goPoint: one for the last card when nobody can go on (thirty-one already scored).
func (g *game) goPoint(last int) {
	if last < 0 || g.count == 31 || len(g.pile) == 0 {
		return
	}
	what := "Go for 1"
	if g.nobodyHasCards() {
		what = "Last card for 1"
	}
	g.emitAll(event{K: "pop", P: last, Text: what})
	g.peg(last, 1, what)
	g.sleep(g.T("play") * 0.9)
}

func (g *game) nobodyHasCards() bool {
	for _, p := range g.seated() {
		if len(p.hand) > 0 {
			return false
		}
	}
	return true
}

// callPeg announces what a pegging card scored and moves the pegs.
func (g *game) callPeg(i, pts int, parts []cr.Part) {
	var ws []string
	kinds := map[string]bool{}
	for _, pt := range parts {
		kinds[pt.Kind] = true
		switch pt.Kind {
		case "Thirty-one":
			ws = append(ws, "31 for 2")
		default:
			ws = append(ws, pt.Kind+" for "+itoa(pt.Pts))
		}
	}
	g.emitAll(event{K: "pop", P: i, Text: strings.Join(ws, " · ")})
	g.emitAll(event{K: "spark", N: 6 + pts*2})
	switch {
	case kinds["Pair royal"] || kinds["Double pair royal"]:
		g.sound("royal")
	case strings.HasPrefix(parts[len(parts)-1].Kind, "Run"):
		g.sound("run")
	default:
		g.sound("fifteen")
	}
	g.peg(i, pts, "")
	p := g.players[i]
	switch {
	case kinds["Pair royal"] || kinds["Double pair royal"]:
		g.say(p, "royal", 0.9, -1, 200)
	case pts >= 3 || kinds["Thirty-one"]:
		g.say(p, "peg", 0.5, -1, 200)
	case kinds["Pair"]:
		g.say(p, "pair", 0.25, -1, 200)
	default:
		g.say(p, "peg", 0.15, -1, 200)
	}
	g.sleep(g.T("play") * 1.4)
}

/* ---------------- the show ---------------- */

func (g *game) show() {
	g.setTurn(-1)
	st := g.starter()
	// everyone picks their cards back up
	for _, p := range g.seated() {
		p.hand = nil
	}
	for _, c := range g.deck {
		if c.zn.K == zPlayed {
			p := g.players[c.zn.P]
			p.hand = append(p.hand, c)
		}
	}
	for _, p := range g.seated() {
		sort.SliceStable(p.hand, func(a, b int) bool { return p.hand[a].zn.N < p.hand[b].zn.N })
	}
	g.emitAll(event{K: "banner", Text: "The Show", Sub: "Every hand is counted with the starter, then the crib.", Blue: true, Ms: g.T("banner") * 0.8})
	g.sleep(g.T("banner") * 0.8)
	for j := 1; j <= g.n(); j++ {
		i := g.act[(g.k(g.dealer)+j)%g.n()]
		p := g.players[i]
		g.countCards(i, p.hand, st, false)
		for n, c := range p.hand {
			c.zn = zone{K: zDone, P: i, N: n, J: (g.rng.Float64() - .5) * 6}
		}
		p.hand = nil
	}
	var crib []*cardV
	for _, c := range g.deck {
		if c.zn.K == zCrib {
			crib = append(crib, c)
		}
	}
	sort.SliceStable(crib, func(a, b int) bool { return crib[a].zn.N < crib[b].zn.N })
	g.countCards(g.dealer, crib, st, true)
	for n, c := range crib {
		c.zn = zone{K: zDone, P: g.dealer, N: 4 + n, J: (g.rng.Float64() - .5) * 6}
	}
	st.zn = zone{K: zStarter}
	g.emitAll(event{K: "tallyOff"})
	g.sync()
	g.persist()
	g.sleep(g.T("move") * 2)
}

// countCards lays four cards out with the starter and calls the score out.
func (g *game) countCards(i int, cards []*cardV, st *cardV, crib bool) {
	cs := cardsOf(cards)
	cr.SortHand(cs)
	for n, c := range cs {
		v := g.byCard[c]
		v.zn = zone{K: zShow, N: n}
		v.speed = 0
	}
	st.zn = zone{K: zShow, N: 4}
	g.emit(func(v int) *event {
		t := g.Whose(i, v) + " hand"
		if crib {
			t = g.Whose(i, v) + " crib"
		}
		return &event{K: "tally", Text: t, Blue: crib}
	})
	g.sync()
	g.sound("slide")
	g.sleep(g.T("move") * 1.6)
	pts, parts := cr.ScoreHand(cs, st.c, crib)
	var lines []tallyLine
	run := 0
	for _, pt := range parts {
		for _, c := range g.deck {
			c.win = false
		}
		for _, c := range pt.Cards {
			g.byCard[c].win = true
		}
		run += pt.Pts
		k := pt.Kind
		if strings.HasPrefix(k, "Run") {
			k = "Run of " + strings.TrimPrefix(k, "Run of ")
		}
		if n := len(lines); n > 0 && lines[n-1].Kind == k {
			lines[n-1].N++
			lines[n-1].Pts += pt.Pts
		} else {
			lines = append(lines, tallyLine{Kind: k, N: 1, Pts: pt.Pts})
		}
		g.emitAll(event{K: "tallyAdd", N: run, L: append([]tallyLine(nil), lines...)})
		g.sync()
		g.sound("count")
		ms := g.T("call")
		if pt.Kind == "Fifteen" {
			ms *= 0.8
		}
		g.sleep(ms)
	}
	for _, c := range g.deck {
		c.win = false
	}
	g.emitAll(event{K: "tallyEnd", N: pts, L: lines})
	g.sync()
	me := g.statSeat()
	if !crib {
		g.tallies[i].Hands++
		g.tallies[i].BestHand = max(g.tallies[i].BestHand, pts)
		if pts == 29 {
			g.tallies[i].TwentyNines++
		}
		if i == me {
			s := &g.save.Stats
			s.Hands++
			s.BestHand = max(s.BestHand, pts)
			if pts == 29 {
				s.TwentyNines++
			}
		}
	}
	p := g.players[i]
	switch {
	case pts == 29:
		g.sound("big")
		g.emit(func(v int) *event {
			return &event{K: "banner", Text: "Twenty-Nine!", Sub: "The perfect cribbage hand.", Blue: !g.mySide(g.side(i), v)}
		})
		g.emitAll(event{K: "confetti", N: 160})
	case pts >= 16:
		g.sound("big")
		g.say(p, "big", 1, -1, 0)
	case pts >= 10:
		g.sound("good")
		g.say(p, "big", 0.4, -1, 0)
	case pts == 0:
		g.sound("nineteen")
		g.say(p, "zero", 0.8, -1, 0)
		if crib {
			g.say(p, "zero", 0.4, -1, 0)
		}
	default:
		g.sound("good")
	}
	g.peg(i, pts, "")
	g.sleep(g.T("banner") * 0.75)
}

// statSeat is the seat whose record this computer keeps (none on a shared computer).
func (g *game) statSeat() int {
	if g.mode == ModeHotseat {
		return -1
	}
	return g.viewer
}

/* ---------------- the end ---------------- */

// places ranks the sides by score, highest first.
func (g *game) places() [4]int {
	var pl [4]int
	for s := 0; s < g.sides(); s++ {
		pl[s] = 1
		for t := 0; t < g.sides(); t++ {
			if g.score[t] > g.score[s] {
				pl[s]++
			}
		}
	}
	return pl
}

// resultFor is the end-of-game card for seat v.
func (g *game) resultFor(v int) resultT {
	pl := g.places()
	var r resultT
	for s := 0; s < g.sides(); s++ {
		row := resRow{Name: g.sideFull(s, v), Score: g.score[s], Place: pl[s], Me: g.mySide(s, v)}
		if s != g.winner {
			row.Skunk = cr.Skunk(g.score[s], g.target)
		}
		for _, i := range g.sideSeats(s) {
			row.Faces = append(row.Faces, g.players[i].face)
		}
		r.rows = append(r.rows, row)
	}
	sort.SliceStable(r.rows, func(a, b int) bool { return r.rows[a].Score > r.rows[b].Score })
	w := g.winner
	wn := g.sideName(w, v)
	wins := "wins"
	if len(g.sideSeats(w)) > 1 || wn == "You" {
		wins = "win"
	}
	sk := 0
	for s := 0; s < g.sides(); s++ {
		if s != w {
			sk = max(sk, cr.Skunk(g.score[s], g.target))
		}
	}
	skW := map[int]string{1: " with a skunk", 2: " with a double skunk"}[sk]
	if g.mode == ModeHotseat {
		r.kick, r.title, r.won = "Hot cocoa for the winner", "Game Over!", true
		r.sub = fmt.Sprintf("%s %s, pegging out at %d%s.", wn, wins, g.target, skW)
		return r
	}
	r.won = g.mySide(w, v)
	mine := g.side(v)
	switch {
	case r.won && sk > 0:
		r.kick, r.title, r.sub = "Fair winds", "A Skunk!", fmt.Sprintf("%s %s by a mile%s.", wn, wins, skW)
	case r.won:
		r.kick, r.title, r.sub = "Fair winds", "You Win!", fmt.Sprintf("%s pegged out first at %d.", wn, g.target)
	case cr.Skunk(g.score[mine], g.target) > 0:
		r.kick, r.title, r.sub = "Rough seas", "Skunked!", fmt.Sprintf("%s %s. %s didn't make it past the skunk line.", wn, wins, g.sideName(mine, v))
	default:
		r.kick, r.title, r.sub = "Better luck next tide", "So Close!", fmt.Sprintf("%s %s. %s finished on %d.", wn, wins, g.sideName(mine, v), g.score[mine])
	}
	return r
}

func ordinal(n int) string {
	return map[int]string{1: "first", 2: "second", 3: "third", 4: "fourth"}[n]
}

func (g *game) endGame() {
	g.finished = true
	g.pegging = false
	g.setTurn(-1)
	g.emitAll(event{K: "prompt"})
	me := g.statSeat()
	w := g.winner
	for s := 0; s < g.sides(); s++ {
		if s != w && cr.Skunk(g.score[s], g.target) > 0 {
			for _, i := range g.sideSeats(w) {
				g.tallies[i].Skunks++
			}
		}
	}
	if me >= 0 {
		won := g.side(me) == w
		g.recordGame(won, g.tallies[me].Skunks)
	}
	g.emit(func(v int) *event {
		return &event{K: "banner", Text: g.sideName(w, v) + " Pegged Out!", Sub: "Game to " + itoa(g.target) + ".", Blue: !g.mySide(w, v) && g.mode != ModeHotseat}
	})
	g.emit(func(v int) *event {
		if g.mySide(w, v) || g.mode == ModeHotseat {
			return &event{K: "snd", Text: "win"}
		}
		return &event{K: "snd", Text: "lose"}
	})
	g.emit(func(v int) *event {
		if g.mySide(w, v) || g.mode == ModeHotseat {
			return &event{K: "confetti", N: 160}
		}
		return nil
	})
	var wAI, lAI []*player
	for _, p := range g.seated() {
		if p.kind == KindAI {
			if g.side(p.i) == w {
				wAI = append(wAI, p)
			} else {
				lAI = append(lAI, p)
			}
		}
	}
	skunked := false
	for s := 0; s < g.sides(); s++ {
		skunked = skunked || s != w && cr.Skunk(g.score[s], g.target) > 0
	}
	if len(wAI) > 0 {
		k := "win"
		if skunked {
			k = "skunk"
		}
		g.say(pickS(g, wAI), k, 1, -1, 0)
	}
	if len(lAI) > 0 {
		g.say(pickS(g, lAI), "lose", 1, -1, 300)
	}
	ms := 1900.0
	if testFast {
		ms = 140
	}
	g.sleep(ms)
	g.result = g.resultFor(g.viewer)
	g.result.rec = g.recordText()
	g.showResult = true
	g.resultAt = time.Now()
	if g.server != nil {
		g.server.sendResults(g)
	}
}

func (g *game) recordGame(won bool, skunks int) {
	st := &g.save.Stats
	st.Played++
	st.Skunks += skunks
	if won {
		st.Won++
		st.Streak++
		st.BestStreak = max(st.BestStreak, st.Streak)
	} else {
		st.Lost++
		st.Streak = 0
	}
	g.persist()
}

func (g *game) recordText() string {
	s := g.save.Stats
	if s.Played == 0 {
		return "No games played yet."
	}
	t := fmt.Sprintf("Games: %d · Won: %d · Skunks: %d · Best hand: %d", s.Played, s.Won, s.Skunks, s.BestHand)
	if s.TwentyNines > 0 {
		t += fmt.Sprintf(" · 29s: %d", s.TwentyNines)
	}
	return t
}

// handsFromZones rebuilds each player's hand from the card zones (guests).
func (g *game) handsFromZones() {
	for _, p := range g.players {
		if p != nil {
			p.hand = p.hand[:0]
		}
	}
	for _, c := range g.deck {
		if c.zn.K == zHand && c.zn.P >= 0 && c.zn.P < 4 && g.players[c.zn.P] != nil {
			g.players[c.zn.P].hand = append(g.players[c.zn.P].hand, c)
		}
	}
	for _, p := range g.players {
		if p != nil {
			sort.Slice(p.hand, func(a, b int) bool { return p.hand[a].zn.N < p.hand[b].zn.N })
		}
	}
	g.pile = g.pile[:0]
	for _, c := range g.deck {
		if c.zn.K == zPile {
			g.pile = append(g.pile, c)
		}
	}
	sort.Slice(g.pile, func(a, b int) bool { return g.pile[a].zn.N < g.pile[b].zn.N })
}
