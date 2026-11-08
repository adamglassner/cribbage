// Online play, the same way Euchre and Go Fish do it. One player hosts: their game
// runs the table and sends each guest a snapshot of what that guest can see
// (their own hand face up, everyone else's face down, cards known only by a
// token that changes every hand) plus the bubbles, banners and sounds,
// worded for their seat. Guests only send back their bids and plays.
// Messages are JSON, one per line, over TCP.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	)

const (
	netPort  = 47261
	protoVer = 1
)

type Msg struct {
	T     string      `json:"t"`
	Name  string      `json:"name,omitempty"`
	Face  string      `json:"face,omitempty"`
	Ver   int         `json:"ver,omitempty"`
	Seat  int         `json:"seat,omitempty"`
	Why   string      `json:"why,omitempty"`
	Lobby []lobbySeat `json:"lobby,omitempty"`
	N     int         `json:"n,omitempty"` // how many play at the table
	Snap  *snap       `json:"snap,omitempty"`
	Ev    *event      `json:"ev,omitempty"`
	Res   *wireRes    `json:"res,omitempty"`
	Ask   *wireAsk    `json:"ask,omitempty"`
	Act   *wireAct    `json:"act,omitempty"`
}

type lobbySeat struct {
	Name string `json:"name"`
	Face string `json:"face"`
	Kind Kind   `json:"kind"`
	On   bool   `json:"on"` // a chair at this table
}

// wireAsk is a decision the host needs from a guest.
type wireAsk struct {
	K      string   `json:"k"` // "crib" or "card"
	Q      string   `json:"q,omitempty"`
	Sub    string   `json:"sub,omitempty"`
	Btns   []askBtn `json:"btns,omitempty"`
	Need   int      `json:"need,omitempty"` // cards to pick (passing)
	Prompt string   `json:"prompt,omitempty"`
	Toks   []int    `json:"toks,omitempty"` // the cards they may pick
}

type wireAct struct {
	Btn  int   `json:"btn"`
	Tok  int   `json:"tok"`
	Toks []int `json:"toks,omitempty"` // the cards laid away
	Drop bool  `json:"-"`              // the guest left
}

type remoteWait struct {
	seat int
	ask  *wireAsk
	ch   chan wireAct
}

type snapCard struct {
	T  int    `json:"t"`
	Z  zone   `json:"z"`
	Sp int    `json:"sp,omitempty"`
	ID string `json:"id,omitempty"`
	W  bool   `json:"w,omitempty"`
}

type snapPlayer struct {
	Name, Full, Face string
	Kind             Kind
	Turn, On         bool
}

type snap struct {
	Gen, Hand, HandNo int
	Viewer, Dealer    int
	Act               []int
	Score, Prev       [4]int
	Target, Count     int
	Pegging           bool
	Winner            int
	P                 [4]snapPlayer
	C                 []snapCard
	Fresh, Finished   bool
}

type wireRes struct {
	Kick, Title, Sub string
	Rows             []resRow
	Won              bool
	Tally            tallyRec
}

func encode(m Msg) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Print("net: ", err)
		return nil
	}
	return append(b, '\n')
}

// conn is one end of a connection with a buffered writer goroutine.
type conn struct {
	c      net.Conn
	out    chan []byte
	closed bool
}

func newConn(c net.Conn) *conn {
	k := &conn{c: c, out: make(chan []byte, 1024)}
	go func() {
		for b := range k.out {
			_ = c.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if _, err := c.Write(b); err != nil {
				_ = c.Close()
				for range k.out { // drain
				}
				return
			}
		}
		_ = c.Close()
	}()
	return k
}

func (k *conn) send(m Msg) {
	if k == nil || k.closed {
		return
	}
	b := encode(m)
	select {
	case k.out <- b:
	default: // hopelessly behind: give up on this connection
		k.close()
	}
}

func (k *conn) close() {
	if !k.closed {
		k.closed = true
		close(k.out)
	}
}

/* ---------------- hosting ---------------- */

type peer struct {
	*conn
	seat int
	name string
}

type server struct {
	ln    net.Listener
	peers []*peer
	addrs []string
	open  bool // taking new players (not mid-game)
}

// joinOrder is where friends sit at a table of n: across from you first, then to your left and right.
func joinOrder(n int) []int {
	switch n {
	case 3:
		return []int{1, 3}
	case 4:
		return []int{2, 1, 3}
	}
	return []int{2}
}

// hostGame starts listening and opens the lobby. You sit at seat 0; the
// other seats are computer folks until friends take them.
func (g *game) hostGame() {
	g.leaveOnline()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", netPort))
	if err != nil {
		g.showMsg("Couldn't open a game on port " + strconv.Itoa(netPort) + ". Is another copy of Cribbage already hosting?\n\n" + err.Error())
		return
	}
	g.closePanels()
	g.server = &server{ln: ln, addrs: localAddrs(), open: true}
	g.mode, g.viewer, g.hidden = ModeHost, 0, false
	g.resetDeck(false)
	g.started, g.finished = false, false
	g.players = [4]*player{}
	g.act = seatsFor(g.tableSize())
	g.setPlayer(0, KindLocal, g.setup.face, cleanName(g.setup.name, "Host"))
	g.fillFolks()
	g.showLobby = true
	g.sendLobby()
	srv := g.server
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go g.handshake(srv, c)
		}
	}()
}

func localAddrs() []string {
	var out []string
	ifs, _ := net.InterfaceAddrs()
	for _, a := range ifs {
		if ip, ok := a.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() && !ip.IP.IsLinkLocalUnicast() {
			out = append(out, ip.IP.String())
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}

func cleanName(s, def string) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) > 14 {
		s = string([]rune(s)[:14])
	}
	if s == "" {
		return def
	}
	return s
}

func (g *game) handshake(srv *server, c net.Conn) {
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	dec := json.NewDecoder(bufio.NewReader(c))
	var m Msg
	k := newConn(c)
	if err := dec.Decode(&m); err != nil || m.T != "hello" {
		k.close()
		return
	}
	_ = c.SetReadDeadline(time.Time{})
	g.mu.Lock()
	if g.server != srv {
		g.mu.Unlock()
		k.close()
		return
	}
	reject, seat := "", -1
	switch {
	case m.Ver != protoVer:
		reject = "That game is running a different version of Cribbage. Update both copies to the same version and try again."
	case !srv.open:
		reject = "That game is in the middle of a hand. Ask the host to let you in after this game."
	default:
		for _, i := range joinOrder(g.n()) {
			if p := g.players[i]; p == nil || p.kind == KindAI {
				seat = i
				break
			}
		}
		if seat < 0 {
			reject = "That table is full: every seat is taken."
		}
	}
	if reject != "" {
		k.send(Msg{T: "reject", Why: reject})
		k.close()
		g.mu.Unlock()
		return
	}
	name := cleanName(m.Name, "Friend")
	for _, p := range g.players { // two people with the same name get told apart
		if p != nil && p.kind != KindAI && strings.EqualFold(p.name, name) {
			name = cleanName(name, "Friend") + " 2"
		}
	}
	f := m.Face
	if !faceOK(f) {
		f = pickerFaces[g.rand(len(pickerFaces))]
	}
	pr := &peer{conn: k, seat: seat, name: name}
	srv.peers = append(srv.peers, pr)
	g.setPlayer(seat, KindRemote, f, name).peer = pr
	g.fillFolks()
	k.send(Msg{T: "welcome", Seat: seat})
	g.sendLobby()
	if g.started && g.finished { // between games: show them the table as it is
		pr.send(Msg{T: "snap", Snap: g.snapFor(seat, false)})
	}
	g.mu.Unlock()

	for {
		var m Msg
		if err := dec.Decode(&m); err != nil {
			break
		}
		if m.T == "bye" {
			break
		}
		g.mu.Lock()
		if m.T == "act" && m.Act != nil && g.players[pr.seat].peer == pr {
			if w := g.remote[pr.seat]; w != nil && validAct(w.ask, m.Act) {
				delete(g.remote, pr.seat)
				w.ch <- *m.Act
			}
		}
		g.mu.Unlock()
	}
	g.mu.Lock()
	g.dropPeer(pr)
	g.mu.Unlock()
}

func validAct(a *wireAsk, r *wireAct) bool {
	in := func(t int) bool {
		for _, x := range a.Toks {
			if x == t {
				return true
			}
		}
		return false
	}
	if a.K == "crib" {
		seen := map[int]bool{}
		for _, t := range r.Toks {
			if !in(t) || seen[t] {
				return false
			}
			seen[t] = true
		}
		return len(r.Toks) == a.Need
	}
	return in(r.Tok)
}

// dropPeer hands a departed guest's seat to a computer player.
func (g *game) dropPeer(pr *peer) {
	pr.close()
	srv := g.server
	if srv == nil {
		return
	}
	keep := srv.peers[:0]
	for _, q := range srv.peers {
		if q != pr {
			keep = append(keep, q)
		}
	}
	srv.peers = keep
	p := g.players[pr.seat]
	if p == nil || p.peer != pr {
		return
	}
	if g.started && !g.finished {
		// a computer player takes over their cards, under the same name
		p.peer, p.kind = nil, KindAI
		f := randomFolks(g.rng, 1, g.facesInUse(), g.namesInUse())
		if len(f) == 1 {
			p.face = f[0].face
		}
		g.emitAll(event{K: "banner", Text: p.name + " left", Sub: "A computer player takes their cards.", Blue: true})
		if w := g.remote[pr.seat]; w != nil {
			delete(g.remote, pr.seat)
			w.ch <- wireAct{Drop: true}
		}
		g.sync()
	} else {
		g.players[pr.seat] = nil
		g.fillFolks()
	}
	g.sendLobby()
}

func (g *game) facesInUse() map[string]bool {
	m := map[string]bool{}
	for _, p := range g.players {
		if p != nil {
			m[p.face] = true
		}
	}
	return m
}

func (g *game) namesInUse() map[string]bool {
	m := map[string]bool{}
	for _, p := range g.players {
		if p != nil {
			m[p.name] = true
		}
	}
	return m
}

func (g *game) sendLobby() {
	if g.server == nil {
		return
	}
	seats := g.lobbySeats()
	g.lobby = seats
	for _, pr := range g.server.peers {
		pr.send(Msg{T: "lobby", Lobby: seats, N: g.n()})
	}
}

// snapFor builds what the person in seat v gets to see.
func (g *game) snapFor(v int, fresh bool) *snap {
	s := &snap{Gen: g.gameID, Hand: g.handsTot, HandNo: g.handNo, Viewer: v, Dealer: g.dealer, Act: g.act, Score: g.score, Prev: g.prev,
		Target: g.target, Count: g.count, Pegging: g.pegging, Winner: g.winner, Fresh: fresh, Finished: g.finished}
	for i, p := range g.players {
		if p == nil {
			continue
		}
		k := p.kind
		if i == v {
			k = KindLocal
		} else if k == KindLocal {
			k = KindRemote // the host, as a guest sees them
		}
		s.P[i] = snapPlayer{Name: p.name, Full: p.full, Face: p.face, Kind: k, Turn: p.isTurn, On: true}
	}
	for _, c := range g.deck {
		if c.zn.K == zNone {
			continue
		}
		mine := c.zn.K == zHand && c.zn.P == v
		sc := snapCard{T: c.tok, Z: c.zn, Sp: c.speed, W: c.win}
		if mine || c.zn.faceUp() {
			sc.ID = c.id
		}
		s.C = append(s.C, sc)
	}
	return s
}

func (srv *server) sendSnaps(g *game, fresh bool) {
	for _, pr := range srv.peers {
		pr.send(Msg{T: "snap", Snap: g.snapFor(pr.seat, fresh)})
	}
}

func (srv *server) sendResults(g *game) {
	srv.open = true
	for _, pr := range srv.peers {
		r := g.resultFor(pr.seat)
		pr.send(Msg{T: "res", Res: &wireRes{Kick: r.kick, Title: r.title, Sub: r.sub, Rows: r.rows, Won: r.won,
			Tally: g.tallies[pr.seat]}})
	}
	g.sendLobby()
}

// startHosted deals the first (or next) game for everyone at the table.
func (g *game) startHosted() {
	if g.server == nil {
		return
	}
	g.closePanels()
	g.newGame()
}

// remoteAsk sends a guest a decision and waits for their answer. It reports
// false if they left (a computer player decides instead).
func (g *game) remoteAsk(i int, a *wireAsk) (wireAct, bool) {
	if !g.remoteSend(i, a) {
		return wireAct{}, false
	}
	return g.remoteWait(i)
}

// remoteSend asks a guest something without waiting (passing asks everyone at once).
func (g *game) remoteSend(i int, a *wireAsk) bool {
	p := g.players[i]
	if p.peer == nil || p.peer.closed {
		return false
	}
	g.remote[i] = &remoteWait{seat: i, ask: a, ch: make(chan wireAct, 1)}
	p.peer.send(Msg{T: "ask", Ask: a})
	return true
}

func (g *game) remoteWait(i int) (wireAct, bool) {
	w := g.remote[i]
	if w == nil {
		return wireAct{}, false
	}
	r := waitOn(g, w.ch)
	return r, !r.Drop
}

/* ---------------- joining ---------------- */

type client struct {
	*conn
	addr string
}

// joinGame connects to a host ("192.168.1.20" or "host:port").
func (g *game) joinGame(addr string) {
	g.leaveOnline()
	addr = strings.TrimSpace(addr)
	if addr == "" {
		g.showMsg("Type the host's address first (it's shown on their screen).")
		return
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, strconv.Itoa(netPort))
	}
	cl := &client{addr: addr}
	g.client = cl
	g.showMsg("Heading over to " + addr + " …")
	name, f := cleanName(g.setup.name, "Friend"), g.setup.face
	go func() {
		c, err := net.DialTimeout("tcp", addr, 8*time.Second)
		g.mu.Lock()
		if g.client != cl {
			g.mu.Unlock()
			if err == nil {
				_ = c.Close()
			}
			return
		}
		if err != nil {
			g.client = nil
			g.showMsg("Couldn't reach a game at " + addr + ".\n\nCheck the address, and that the host's firewall lets Cribbage through.")
			g.mu.Unlock()
			return
		}
		cl.conn = newConn(c)
		cl.send(Msg{T: "hello", Name: name, Face: f, Ver: protoVer})
		g.mu.Unlock()
		dec := json.NewDecoder(bufio.NewReader(c))
		for {
			var m Msg
			if err := dec.Decode(&m); err != nil {
				break
			}
			g.mu.Lock()
			if g.client == cl {
				g.clientMsg(&m)
			}
			g.mu.Unlock()
		}
		g.mu.Lock()
		if g.client == cl {
			g.client = nil
			g.leaveOnline()
			g.toMenu()
			g.showMsg("Lost the connection to the host.")
		}
		g.mu.Unlock()
	}()
}

func (cl *client) send(m Msg) {
	if cl != nil && cl.conn != nil {
		cl.conn.send(m)
	}
}

// becomeGuest sets this computer up to show someone else's table.
func (g *game) becomeGuest(seat int) {
	g.gameID++
	if g.cancel != nil {
		close(g.cancel)
	}
	g.cancel = make(chan struct{})
	g.dropClientAsk()
	g.mode, g.viewer, g.hidden = ModeClient, seat, false
	g.started, g.finished = false, false
	g.resetDeck(true)
	g.score, g.prev = [4]int{}, [4]int{}
	g.players = [4]*player{}
	g.closePanels()
	g.showTitle, g.showResult = false, false
	g.showLobby = true
}

func (g *game) clientMsg(m *Msg) {
	switch m.T {
	case "welcome":
		if g.mode != ModeClient {
			g.becomeGuest(m.Seat)
		} else {
			g.viewer = m.Seat
		}
	case "lobby":
		g.lobby = m.Lobby
		if m.N >= 2 && m.N <= 4 {
			g.act = seatsFor(m.N)
		}
	case "reject":
		g.dropClient()
		g.showMsg(m.Why)
	case "closed":
		g.dropClient()
		g.leaveOnline()
		g.toMenu()
		g.showMsg("The host closed the game.")
	case "snap":
		if m.Snap != nil {
			g.applySnap(m.Snap)
		}
	case "ev":
		if m.Ev != nil {
			g.apply(m.Ev)
		}
	case "ask":
		if m.Ask != nil {
			g.clientAsk(m.Ask)
		}
	case "res":
		if r := m.Res; r != nil {
			g.dropClientAsk()
			g.finished = true
			st := &g.save.Stats
			st.Hands += r.Tally.Hands
			st.BestHand = max(st.BestHand, r.Tally.BestHand)
			st.TwentyNines += r.Tally.TwentyNines
			g.recordGame(r.Won, r.Tally.Skunks)
			g.result = resultT{kick: r.Kick, title: r.Title, sub: r.Sub, rows: r.Rows, won: r.Won, rec: g.recordText()}
			g.showResult, g.resultAt = true, time.Now()
		}
	}
}

func (g *game) dropClient() {
	cl := g.client
	g.client = nil
	if cl != nil && cl.conn != nil {
		cl.close()
	}
	g.dropClientAsk()
}

// dropClientAsk forgets a question the host asked us.
func (g *game) dropClientAsk() {
	if g.cliEnd != nil {
		close(g.cliEnd)
		g.cliEnd = nil
	}
	g.ask, g.wait = nil, nil
	for _, c := range g.deck {
		c.pick, c.dim, c.sel = false, false, false
	}
	g.setPrompt("")
}

func (g *game) applySnap(s *snap) {
	if s.Gen != g.gameID || s.Fresh {
		g.gameID = s.Gen
		g.dropClientAsk()
		g.closePanels()
		g.showTitle, g.showResult = false, false
		g.parts, g.ban = nil, bannerT{}
	}
	if s.Hand != g.handsTot { // new tokens every hand
		g.handsTot = s.Hand
		for _, c := range g.deck {
			c.id = ""
		}
	}
	g.started, g.finished = true, s.Finished
	g.viewer, g.dealer, g.score, g.prev, g.target = s.Viewer&3, s.Dealer&3, s.Score, s.Prev, s.Target
	g.handNo, g.count, g.pegging, g.winner = s.HandNo, s.Count, s.Pegging, s.Winner
	if len(s.Act) >= 2 && len(s.Act) <= 4 {
		g.act = seatsFor(len(s.Act))
	}
	for i, sp := range s.P {
		if !sp.On {
			g.players[i] = nil
			continue
		}
		p := g.players[i]
		if p == nil {
			p = &player{}
			g.players[i] = p
		}
		p.i, p.name, p.full, p.face, p.kind = i, sp.Name, sp.Full, sp.Face, sp.Kind
		if !faceOK(p.face) {
			p.face = ""
		}
		if sp.Turn && !p.isTurn {
			p.turnAt = time.Now()
		}
		p.isTurn = sp.Turn
	}
	seen := map[int]bool{}
	for _, sc := range s.C {
		if sc.T < 0 || sc.T >= 52 {
			continue
		}
		c := g.byTok[sc.T]
		seen[sc.T] = true
		c.zn, c.speed, c.win = sc.Z, sc.Sp, sc.W
		if c.zn.P < 0 || c.zn.P > 3 || g.players[c.zn.P] == nil {
			c.zn.P = g.viewer
		}
		if sc.ID != "" {
			c.id = sc.ID
		}
	}
	for t, c := range g.byTok {
		if !seen[t] {
			c.zn = zone{}
		}
	}
	g.handsFromZones()
	g.layoutCards(false)
}

// clientAsk shows the host's question here and sends the answer back.
func (g *game) clientAsk(a *wireAsk) {
	g.dropClientAsk()
	end := make(chan struct{})
	g.cliEnd = end
	cl := g.client
	var ok []*cardV
	for _, t := range a.Toks {
		if t >= 0 && t < 52 {
			ok = append(ok, g.byTok[t])
		}
	}
	me := g.players[g.viewer]
	if a.K == "crib" {
		at := &askT{q: a.Q, sub: a.Sub, btns: a.Btns, need: a.Need, all: ok, at: time.Now(), ch: make(chan askRes, 1)}
		if len(at.btns) == 0 {
			at.btns = []askBtn{{Label: "Pass"}}
		}
		g.ask = at
		for _, c := range ok {
			c.pick, c.sel = true, false
		}
		go func() {
			select {
			case r := <-at.ch:
				var toks []int
				for _, c := range r.cards {
					toks = append(toks, c.tok)
				}
				cl.send(Msg{T: "act", Act: &wireAct{Toks: toks}})
				g.mu.Lock()
				for _, c := range at.all {
					c.pick, c.sel = false, false
				}
				g.mu.Unlock()
			case <-end:
			}
		}()
		if autopilot {
			go func() {
				time.Sleep(30 * time.Millisecond)
				g.mu.Lock()
				defer g.mu.Unlock()
				if g.ask == at && len(ok) >= at.need {
					g.ask = nil
					perm := g.rng.Perm(len(ok))
					var cs []*cardV
					for _, k := range perm[:at.need] {
						cs = append(cs, ok[k])
					}
					at.ch <- askRes{btn: at.btns[0], cards: cs}
				}
			}()
		}
		return
	}
	for _, c := range me.hand {
		in := false
		for _, x := range ok {
			in = in || x == c
		}
		c.pick, c.dim = in, !in
	}
	w := &waitT{ok: ok, all: me.hand, ch: make(chan *cardV, 1)}
	g.wait = w
	g.setPrompt(a.Prompt)
	go func() {
		select {
		case c := <-w.ch:
			cl.send(Msg{T: "act", Act: &wireAct{Tok: c.tok}})
			g.mu.Lock()
			for _, x := range w.all {
				x.pick, x.dim = false, false
			}
			g.setPrompt("")
			g.mu.Unlock()
		case <-end:
		}
	}()
	if autopilot && len(ok) > 0 {
		go func() {
			time.Sleep(30 * time.Millisecond)
			g.mu.Lock()
			defer g.mu.Unlock()
			if g.wait == w {
				g.wait = nil
				w.ch <- ok[g.rng.IntN(len(ok))]
			}
		}()
	}
}

// leaveOnline closes any online game we're hosting or in.
func (g *game) leaveOnline() {
	if srv := g.server; srv != nil {
		g.server = nil
		_ = srv.ln.Close()
		for _, pr := range srv.peers {
			pr.send(Msg{T: "closed"})
			pr.close()
		}
		for _, p := range g.players {
			if p != nil && p.kind == KindRemote {
				p.kind, p.peer = KindAI, nil
			}
		}
	}
	if g.mode == ModeHost {
		g.mode = ModeSolo
	}
	if cl := g.client; cl != nil {
		g.client = nil
		if cl.conn != nil {
			cl.send(Msg{T: "bye"})
			cl.close()
		}
	}
	if g.mode == ModeClient { // back to a table of our own
		g.gameID++
		if g.cancel != nil {
			close(g.cancel)
		}
		g.cancel = make(chan struct{})
		g.dropClientAsk()
		g.resetDeck(false)
		g.soloTable()
		g.started, g.finished = false, false
		g.score, g.prev = [4]int{}, [4]int{}
	}
	g.lobby = nil
	g.showLobby = false
}

// lobbyRows is the lobby list as this computer should show it (seat order).
func (g *game) lobbyRows() []lobbySeat {
	if g.mode == ModeHost || len(g.lobby) != 4 {
		return g.lobbySeats()
	}
	return g.lobby
}

func (g *game) lobbySeats() []lobbySeat {
	out := make([]lobbySeat, 4)
	for i, p := range g.players {
		if p != nil {
			out[i] = lobbySeat{Name: p.full, Face: p.face, Kind: p.kind, On: true}
		}
	}
	return out
}

// setTableSize changes how many play at a hosted table (between games).
// Friends already seated keep their chairs if the table still has them.
func (g *game) setTableSize(n int) {
	if g.server == nil || (g.started && !g.finished) {
		return
	}
	guests := 0
	for _, p := range g.players {
		if p != nil && p.kind == KindRemote {
			guests++
		}
	}
	if n-1 < guests {
		g.showMsg("Too many friends have joined for a table that small.")
		return
	}
	var keep []*player
	for _, i := range joinOrder(g.n()) {
		if p := g.players[i]; p != nil && p.kind == KindRemote {
			keep = append(keep, p)
		}
	}
	g.save.Settings.Players = n
	g.persist()
	host := g.players[0]
	g.players = [4]*player{}
	g.players[0] = host
	g.act = seatsFor(n)
	inTable := func(s int) bool {
		for _, x := range g.act {
			if x == s {
				return true
			}
		}
		return false
	}
	var moved []*player
	for _, p := range keep {
		if s := p.i; s != 0 && inTable(s) && g.players[s] == nil {
			g.players[s] = p
			p.peer.send(Msg{T: "welcome", Seat: s})
		} else {
			moved = append(moved, p)
		}
	}
	for _, p := range moved {
		s := -1
		for _, x := range joinOrder(n) {
			if g.players[x] == nil {
				s = x
				break
			}
		}
		if s < 0 {
			continue
		}
		p.i = s
		p.peer.seat = s
		g.players[s] = p
		p.peer.send(Msg{T: "welcome", Seat: s})
	}
	g.fillFolks()
	g.sendLobby()
}
