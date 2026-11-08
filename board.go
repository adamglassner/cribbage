//go:build !headless

// The cribbage board in the top-left corner, the running count beside the
// pegging, the "Fifteen for 2" pops, and the tally box that adds up a hand
// in the show.
package main

import (
	"image/color"
	"math"
	"strings"
	"time"
)

// pegCols are the sides' pegs: lighthouse red, harbour blue, bottle green.
var pegCols = []color.NRGBA{hexc("c8372d", 1), hexc("2f6ea8", 1), hexc("2f8a52", 1), hexc("d9a62a", 1)}

type boardGeo struct {
	x, y, w, h  float64
	nameW       float64
	hx0, dx, gp float64 // first hole x, hole step, extra gap between groups of five
	rowH, band  float64
	top         float64 // y of the first band
}

func (g *game) boardGeo() boardGeo {
	var b boardGeo
	b.w = clamp(g.Wd*0.4, 360, 560)
	b.x, b.y = 14, 14
	b.nameW = 108
	b.rowH, b.band = 10, 30
	b.top = b.y + 26
	b.h = 26 + b.band*float64(g.sides()) + 10
	b.hx0 = b.x + 12 + b.nameW + 22
	inner := b.x + b.w - 14 - b.hx0
	b.dx = inner / (59 + 11*0.8)
	b.gp = b.dx * 0.8
	return b
}

// holeX is the x of the n-th hole (1-60) along a row, left to right.
func (b boardGeo) holeX(n int) float64 {
	return b.hx0 + float64(n-1)*b.dx + float64((n-1)/5)*b.gp
}

// holeXY is where a peg on point n sits for side s: out along the lower row
// (1-60), back along the upper one (61-120), home in the finish hole (121).
// The back peg starts in the second start hole.
func (b boardGeo) holeXY(s, n int, back bool) (float64, float64) {
	upper := b.top + float64(s)*b.band + 7
	lower := upper + b.rowH
	switch {
	case n <= 0:
		if back {
			return b.hx0 - 15, lower
		}
		return b.hx0 - 8, lower
	case n <= 60:
		return b.holeX(n), lower
	case n <= 120:
		return b.holeX(121 - n), upper
	}
	return b.hx0 - 12, upper
}

// pegAt eases between holes so a peg glides along its track.
func (b boardGeo) pegAt(s int, f float64, back bool, target int) (float64, float64) {
	f = math.Min(f, float64(target))
	lo := int(math.Floor(f))
	t := f - float64(lo)
	x0, y0 := b.holeXY(s, lo, back)
	if t < 1e-6 {
		return x0, y0
	}
	x1, y1 := b.holeXY(s, lo+1, back)
	return x0 + (x1-x0)*t, y0 + (y1-y0)*t
}

func (g *game) drawBoard(p pen) {
	b := g.boardGeo()
	u := &g.ui
	// the board: a plank of cherry with rounded ends
	p.soft(b.x, b.y+8, b.w, b.h, 12, 22, 0.4)
	p.rr(b.x, b.y+3, b.w, b.h, 12, hexc("3a1e0e", 1))
	p.rr(b.x, b.y, b.w, b.h, 12, hexc("5a2e14", 1))
	p.vgrad(b.x+2, b.y+2, b.w-4, b.h-4, hexc("a8643a", 1), hexc("7a4220", 1))
	for i := 0; i < 9; i++ { // wood grain
		yy := b.y + 6 + float64(i)*(b.h-12)/8
		p.rect(b.x+8, yy, b.w-16, 0.8, color.NRGBA{60, 28, 10, 40})
	}
	p.rrStroke(b.x, b.y, b.w, b.h, 12, 2, hexc("2a1408", 1))
	tf := slab(12)
	tl := "GAME TO " + itoa(g.target)
	p.spaced(tl, tf, b.x+14, base(tf, b.y+6, 16), 0.12, hexc("f6e2c0", 0.9), alignL)
	p.text("121", italic(11, 700), b.hx0-12, base(italic(11, 700), b.y+6, 16), hexc("f6e2c0", 0.8), alignC)
	for _, n := range []int{30, 60} {
		p.text(itoa(n), italic(11, 700), b.holeX(n), base(italic(11, 700), b.y+6, 16), hexc("f6e2c0", 0.65), alignC)
	}
	// skunk line (91 on the way back, 31 in a short game)
	sk := 91
	if g.target <= 61 {
		sk = 31
	}
	for s := 0; s < g.sides(); s++ {
		v := g.viewer
		top := b.top + float64(s)*b.band
		// the side's name and score
		pc := pegCols[s%len(pegCols)]
		p.rr(b.x+8, top+1, b.nameW, b.band-4, 6, color.NRGBA{40, 18, 6, 120})
		p.circle(b.x+18, top+b.band/2-1, 5, pc)
		nm := g.sideName(s, v)
		nf := body(13, 700)
		for richWidth(nm, nf) > b.nameW-58 && len([]rune(nm)) > 3 {
			r := []rune(nm)
			nm = strings.TrimRight(string(r[:len(r)-2]), " &") + "…"
		}
		p.text(nm, nf, b.x+28, base(nf, top+1, b.band-4), hexc("fbf2e2", 1), alignL)
		sf := slab(17)
		p.text(itoa(g.score[s]), sf, b.x+8+b.nameW-8, base(sf, top+1, b.band-4)+1, hexc("f2d68e", 1), alignR)
		// its track: a groove with two rows of holes
		ty := top + 3
		p.rr(b.hx0-20, ty, b.x+b.w-14-(b.hx0-20)+6, b.rowH*2+8, 7, color.NRGBA{40, 18, 6, 70})
		p.rect(b.hx0-4, ty+b.rowH+3.5, b.x+b.w-14-b.hx0+6, 1.2, colA(pc, 0.55))
		for n := 1; n <= 120; n++ {
			x, y := b.holeXY(s, n, false)
			p.circle(x, y, 1.7, color.NRGBA{30, 12, 4, 230})
		}
		for _, back := range []bool{false, true} {
			x, y := b.holeXY(s, 0, back)
			p.circle(x, y, 1.7, color.NRGBA{30, 12, 4, 230})
		}
		fx, fy := b.holeXY(s, 121, false)
		p.circle(fx, fy, 3, color.NRGBA{30, 12, 4, 230})
		p.circle(fx, fy, 1.8, hexc("d9b45a", 1))
		// the skunk line
		sx, _ := b.holeXY(s, sk, false)
		sx2, _ := b.holeXY(s, sk-1, false)
		lx := (sx + sx2) / 2
		if sk == 31 {
			lx = (b.holeX(30) + b.holeX(31)) / 2
			p.rect(lx-0.6, ty+b.rowH+3, 1.2, b.rowH+5, hexc("f6e2c0", 0.55))
		} else {
			p.rect(lx-0.6, ty, 1.2, b.rowH+5, hexc("f6e2c0", 0.55))
		}
		// the pegs: back peg, then front peg
		bx, by := b.pegAt(s, u.pegB[s], true, g.prev[s])
		if g.prev[s] == 0 && g.score[s] > 0 {
			bx, by = b.holeXY(s, 0, true)
		}
		fx, fy = b.pegAt(s, u.pegF[s], false, g.score[s])
		for _, pt := range [][2]float64{{bx, by}, {fx, fy}} {
			p.circle(pt[0]+0.8, pt[1]+1.6, 4.2, color.NRGBA{0, 0, 0, 90})
			p.circle(pt[0], pt[1], 4, lighten(pc, 0.6))
			p.circle(pt[0], pt[1], 3.2, pc)
			p.circle(pt[0]-1, pt[1]-1.1, 1.2, color.NRGBA{255, 255, 255, 150})
		}
	}
}

// drawCount shows the running count while the cards are pegged.
func (g *game) drawCount(p pen) {
	if !g.pegging || g.finished {
		return
	}
	x, y := g.Wd/2-g.cw*0.85-g.cw*0.5, g.cy()+g.ch*0.62
	nf := slab(26)
	lbl := "COUNT"
	w := math.Max(84, measure(itoa(g.count), nf)+spacedWidth(lbl, slab(11), 0.12)+40)
	h := 40.0
	p.soft(x, y+6, w, h, 20, 14, 0.4)
	p.rr(x, y, w, h, 20, cBrown)
	p.rr(x+2, y+2, w-4, h-4, 18, cPaper)
	p.spaced(lbl, slab(11), x+16, base(slab(11), y, h)-1, 0.12, cMuted, alignL)
	col := cInk
	if g.count == 15 || g.count == 31 {
		col = cBarn
	}
	p.text(itoa(g.count), nf, x+w-16, base(nf, y, h)+1, col, alignR)
}

// drawPops floats each "Fifteen for 2" up from the cards.
func (g *game) drawPops(p pen, now time.Time) {
	keep := g.pops[:0]
	for _, q := range g.pops {
		t := now.Sub(q.at).Seconds() / (1.9 * math.Max(0.6, g.pace))
		if t >= 1 {
			continue
		}
		keep = append(keep, q)
		a := math.Min(1, t*6) * math.Min(1, (1-t)*3)
		x, y := g.pileEnd()
		if !g.pegging {
			x, y = g.Wd/2, g.cy()
		}
		y -= g.ch*0.55 + 40*easeOut(t)
		f := slab(20)
		w := richWidth(q.text, f) + 28
		h := 34.0
		x = clamp(x, w/2+10, g.Wd-w/2-10)
		pc := pegCols[g.side(q.seat)%len(pegCols)]
		p.rr(x-w/2, y-h/2+3, w, h, 17, color.NRGBA{0, 0, 0, u8(0.35 * a)})
		p.rr(x-w/2, y-h/2, w, h, 17, colA(lighten(pc, 0.7), a))
		p.rr(x-w/2+2, y-h/2+2, w-4, h-4, 15, colA(pc, a))
		p.rich(q.text, f, x, base(f, y-h/2, h)+1, colA(cWhite, a), alignC)
	}
	g.pops = keep
}

// drawTally adds up the hand being counted in a box above the show.
func (g *game) drawTally(p pen, now time.Time) {
	t := g.tally
	if t == nil {
		return
	}
	k := math.Min(1, now.Sub(t.at).Seconds()/0.25)
	tf, cf, nf := slab(17), body(14, 700), slab(30)
	var chips []string
	for _, l := range t.lines {
		s := l.Kind
		switch {
		case l.Kind == "Fifteen" && l.N > 1:
			s = "Fifteens ×" + itoa(l.N)
		case l.Kind == "Pair" && l.N > 1:
			s = "Pairs ×" + itoa(l.N)
		case l.N > 1:
			s = l.Kind + " ×" + itoa(l.N)
		}
		chips = append(chips, s+"  "+itoa(l.Pts))
	}
	if t.done && len(t.lines) == 0 {
		chips = append(chips, "Nineteen! Not a point")
	}
	maxW := math.Min(600, g.Wd*0.5)
	// lay the chips out in rows
	var rows [][]string
	rw := 0.0
	for _, c := range chips {
		w := richWidth(c, cf) + 20
		if len(rows) == 0 || rw+w+6 > maxW-120 {
			rows = append(rows, nil)
			rw = 0
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], c)
		rw += w + 6
	}
	inner := measure(t.title, tf)
	for _, r := range rows {
		w := -6.0
		for _, c := range r {
			w += richWidth(c, cf) + 26
		}
		inner = math.Max(inner, w)
	}
	w := math.Max(280, inner+120)
	h := 12 + 24 + math.Max(1, float64(len(rows)))*30 + 10
	showTop := g.cy() - g.ch*0.04 - g.ch*0.51 - 12
	x, y := g.Wd/2-w/2, showTop-h
	y = math.Max(y, g.ch*0.42+2+g.ch*0.45)
	q := p.scaled(g.Wd/2, y+h/2, 0.9+0.1*easeCSS(k))
	q.soft(x, y+10, w, h, 12, 24, 0.4*k)
	fill := cBarn
	if t.crib {
		fill = cDenim
	}
	q.box(x, y, w, h, 12, 3, colA(cPaper, k), colA(cBrown, k), 4, 0)
	q.text(t.title, tf, x+16, base(tf, y+10, 24), colA(fill, k), alignL)
	cy := y + 10 + 26
	for _, r := range rows {
		cx := x + 16
		for _, c := range r {
			cw := richWidth(c, cf) + 20
			q.rr(cx, cy, cw, 25, 7, colA(lighten(fill, 1.0), 0.12*k))
			q.rrStroke(cx, cy, cw, 25, 7, 1, colA(fill, 0.5*k))
			q.rich(c, cf, cx+cw/2, base(cf, cy, 25), colA(cInk, k), alignC)
			cx += cw + 6
		}
		cy += 30
	}
	// the running total
	bw := 80.0
	bx := x + w - bw - 12
	q.rr(bx, y+10, bw, h-20, 10, colA(fill, k))
	q.text(itoa(t.total), nf, bx+bw/2, base(nf, y+10, h-20)+2, colA(cWhite, k), alignC)
	if t.done {
		q.text("TOTAL", slab(10), bx+bw/2, y+h-16, colA(hexc("f6e2c0", 1), k), alignC)
	}
}
