// Play Together: who sits where on a shared computer, and the name and face
// you use online. (The screens for these are in setup_ui.go.)
package main

import (
	"fmt"
	"strings"
)

type setupState struct {
	hotN       int // people sharing this computer (2-4)
	hotNames   [4]string
	hotFaces   [4]string
	name, face string
	addr       string
}

func newSetup(sv *saveData) setupState {
	s := sv.Settings
	st := setupState{hotN: s.HotN, name: s.Name, face: s.Face, addr: s.Host}
	if st.hotN < 2 || st.hotN > 4 {
		st.hotN = 2
	}
	for i := 0; i < 4; i++ {
		st.hotNames[i] = fmt.Sprintf("Player %d", i+1)
		st.hotFaces[i] = pickerFaces[[]int{0, 1, 2, 3}[i]]
		if i < len(s.HotNames) && strings.TrimSpace(s.HotNames[i]) != "" {
			st.hotNames[i] = s.HotNames[i]
		}
		if i < len(s.HotFaces) && faceOK(s.HotFaces[i]) {
			st.hotFaces[i] = s.HotFaces[i]
		}
	}
	if strings.TrimSpace(st.name) == "" {
		st.name = "Player"
	}
	if !faceOK(st.face) {
		st.face = pickerFaces[0]
	}
	return st
}

// keepSetup saves the setup choices with the settings.
func (g *game) keepSetup() {
	s := &g.save.Settings
	st := &g.setup
	s.HotN, s.Name, s.Face, s.Host = st.hotN, st.name, st.face, st.addr
	s.HotNames, s.HotFaces = st.hotNames[:], st.hotFaces[:]
	g.persist()
}

// hotSeats says which seat each person sharing the computer takes. Two
// people sit across from each other; three or four go round the table.
func (st *setupState) hotSeats() []int { return seatsFor(st.hotN) }

func (g *game) startHotseat() {
	g.leaveOnline()
	g.keepSetup()
	st := &g.setup
	g.mode, g.viewer = ModeHotseat, 0
	g.players = [4]*player{}
	g.act = seatsFor(st.hotN)
	for k, seat := range st.hotSeats() {
		name := cleanName(st.hotNames[k], fmt.Sprintf("Player %d", k+1))
		for _, p := range g.players { // two people with the same name get told apart
			if p != nil && strings.EqualFold(p.name, name) {
				name = cleanName(name, "Player") + " " + itoa(k+1)
			}
		}
		g.setPlayer(seat, KindLocal, st.hotFaces[k], name)
	}
	g.fillFolks()
	g.closePanels()
	g.showTitle = false
	g.startAudio()
	g.newGame()
}

func (g *game) startAudio() {
	snd.setVolume(float64(g.save.Settings.Vol) / 100)
	snd.setSfx(g.save.Settings.Sfx)
	snd.setMusic(g.save.Settings.Music)
	snd.startMusic()
}

// goPlay is "Deal" on the title card: you and the regulars it shows.
func (g *game) goPlay() {
	g.leaveOnline()
	g.mode = ModeSolo
	g.closePanels()
	g.showTitle = false
	g.startAudio()
	g.newGame()
}

// newDeal is New Game / F2 / Play Again: the same table, dealt again.
func (g *game) newDeal() {
	switch g.mode {
	case ModeClient:
		g.showMsg("Only the host can deal a new game.")
	case ModeHost:
		g.showTitle = false
		g.startHosted()
	case ModeHotseat:
		g.closePanels()
		g.showTitle, g.showResult = false, false
		g.newGame()
	default:
		g.goPlay()
	}
}

func (g *game) toMenu() {
	g.showResult = false
	g.showTitle = true
}

func (g *game) closePanels() {
	g.showTogether, g.showHotseat, g.showOnline, g.showLobby = false, false, false, false
	g.message, g.focus = "", ""
}

func (g *game) showMsg(s string) { g.message = s }
