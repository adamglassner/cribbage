// The cottage regulars who might pull up a chair. Every face is made up
// fresh from a seed, and its name, title and way of talking always fit the
// face: a grandmother with silver hair is "Nana", a man in a peaked cap is
// "Cap'n", a lobsterman in his oilskins is "Skipper". (The drawing itself is
// in cameo.go.)
package main

import (
	"image/color"
	"math/rand/v2"
	"strconv"
	"strings"
)

// Face is all a seat needs to share over the network: the look, the name and
// the manner are all worked out from it. As text it is "f:<seed>" or "m:<seed>".
type Face struct {
	Female bool
	Seed   uint64
}

func (f Face) String() string {
	if f.Female {
		return "f:" + strconv.FormatUint(f.Seed, 10)
	}
	return "m:" + strconv.FormatUint(f.Seed, 10)
}

func parseFace(s string) (Face, bool) {
	if len(s) < 3 || s[1] != ':' || (s[0] != 'f' && s[0] != 'm') {
		return Face{}, false
	}
	n, err := strconv.ParseUint(s[2:], 10, 64)
	if err != nil {
		return Face{}, false
	}
	return Face{Female: s[0] == 'f', Seed: n}, true
}

func faceOK(s string) bool { _, ok := parseFace(s); return ok }

/* ---------------- the look ---------------- */

type Look struct {
	Lady, Stern, Smile      bool
	Skin, Hair, Brow, Lip   color.RGBA
	HairStyle               string // bob, bun, waves, curls | short, side, bald, curly
	Top                     string // stripes, blouse, sweater, peacoat, slicker, cardigan
	Coat, Dark, Accent      color.RGBA
	Stripe                  color.RGBA
	Neck                    string // scarf, pearls, kerchief, ""
	NeckColor               color.RGBA
	Hat                     string // beanie, cloche, cap, souwester, captain, ""
	HatColor                color.RGBA
	Mustache                string // pencil, handlebar, walrus, ""
	Beard                   bool
	Glasses                 string // round, cateye, ""
	Mole, Earrings, Freckle bool
}

func rgb(s string) color.RGBA {
	s = strings.TrimPrefix(s, "#")
	v := func(i int) uint8 { n, _ := strconv.ParseUint(s[i:i+2], 16, 8); return uint8(n) }
	return color.RGBA{v(0), v(2), v(4), 255}
}

func shade(c color.RGBA, k float64) color.RGBA {
	f := func(v uint8) uint8 { return uint8(min(255, float64(v)*k)) }
	return color.RGBA{f(c.R), f(c.G), f(c.B), 255}
}

func choose[T any](r *rand.Rand, a ...T) T { return a[r.IntN(len(a))] }

var skinTones = []string{"#f6dbc4", "#f2d4b6", "#f1d0b5", "#eec3a2", "#e3b48e", "#d6a27a", "#c98f66", "#a8704a", "#8a5a3a", "#6e452c"}

func (f Face) rng(salt uint64) *rand.Rand { return rand.New(rand.NewPCG(f.Seed, salt)) }

// look builds a regular's appearance from their face. Ladies and gentlemen
// draw from different wardrobes, so a face always matches its name.
func (f Face) look() Look {
	r := f.rng(0xca7e)
	l := Look{Skin: rgb(choose(r, skinTones...)), Stern: r.IntN(10) < 3}
	l.Smile = !l.Stern && r.IntN(10) < 6
	l.Freckle = r.IntN(10) < 2
	grey := r.IntN(10) < 3
	slicker := "#e8b422"
	if f.Female {
		l.Lady = true
		if grey {
			l.Hair = rgb(choose(r, "#b8b6bc", "#d6d2cc", "#9c9aa2", "#e8e4de"))
			l.Brow = shade(l.Hair, .7)
		} else {
			l.Hair = rgb(choose(r, "#2a1e18", "#4a2e1c", "#6b3f22", "#8a5a30", "#b5651d", "#d8b46a", "#e6c88a", "#1c1414", "#a8461e"))
		}
		l.HairStyle = choose(r, "bob", "bob", "bun", "waves", "curls")
		l.Lip = rgb(choose(r, "#b3283a", "#c0392b", "#9c2a40", "#c25a5a", "#a0404a"))
		l.Top = choose(r, "stripes", "blouse", "blouse", "sweater", "cardigan", "slicker")
		l.Coat = rgb(choose(r, "#1f3a5a", "#7a2a24", "#2f5a4a", "#c9a24c", "#3a4a6a", "#e8e0cc", "#a8463a", "#3a6a8a", "#6a7a5a", "#5a3a4a"))
		if l.Top == "slicker" {
			l.Coat = rgb(slicker)
		}
		l.Stripe = rgb(choose(r, "#1f3a5a", "#1f3a5a", "#a3242b"))
		l.Accent = rgb(choose(r, "#f4ecd8", "#d8e4ec", "#d9b44a", "#ffffff", "#b3283a"))
		switch n := r.IntN(10); {
		case n < 3:
			l.Neck = "scarf"
			l.NeckColor = rgb(choose(r, "#b3283a", "#e8b04a", "#3a7a9a", "#7a3a7a", "#2f7a5a", "#e07a5f"))
		case n < 5:
			l.Neck = "pearls"
		}
		switch n := r.IntN(100); {
		case l.Top == "slicker" && n < 70:
			l.Hat = "souwester"
			l.HatColor = rgb(slicker)
		case n < 26:
			l.Hat = "beanie"
			l.HatColor = rgb(choose(r, "#b3283a", "#1f3a5a", "#e8b04a", "#2f5a4a", "#f2ece0", "#6a3a5a"))
		case n < 44:
			l.Hat = "cloche"
			l.HatColor = rgb(choose(r, "#e9d9b0", "#1f3a5a", "#7a2a24", "#c9a27a", "#5a3a4a", "#2f5a4a"))
		}
		l.Earrings = r.IntN(10) < 5
		l.Mole = r.IntN(10) < 2
		if r.IntN(100) < 18 {
			l.Glasses = "cateye"
		}
		if grey && r.IntN(3) == 0 {
			l.Glasses = "round"
		}
		if l.Hat != "" && l.HairStyle == "bun" { // a hat would hide the bun: let the hair show
			l.HairStyle = choose(r, "bob", "waves", "curls")
		}
	} else {
		if grey {
			l.Hair = rgb(choose(r, "#e8e4dc", "#c9c4bc", "#a8a4a0"))
			l.Brow = shade(l.Hair, .75)
		} else {
			l.Hair = rgb(choose(r, "#2a1e18", "#4a2e1c", "#6b3f22", "#8a5a30", "#b5651d", "#c8a060", "#1c1414"))
		}
		l.HairStyle = choose(r, "short", "side", "side", "curly")
		if grey && r.IntN(2) == 0 {
			l.HairStyle = "bald"
		}
		l.Top = choose(r, "stripes", "peacoat", "peacoat", "sweater", "sweater", "cardigan", "slicker")
		l.Coat = rgb(choose(r, "#1f3a5a", "#3b2a1e", "#2f4a3a", "#4a2a2a", "#2a2a3a", "#5a4a3a", "#6b3a28", "#3a3a4a", "#e8e0cc", "#7a6a50"))
		if l.Top == "peacoat" {
			l.Coat = rgb(choose(r, "#1c2a44", "#24304a", "#2a2a30", "#3a2a22"))
		}
		if l.Top == "slicker" {
			l.Coat = rgb(choose(r, slicker, slicker, "#d8562a"))
		}
		l.Stripe = rgb(choose(r, "#1f3a5a", "#1f3a5a", "#a3242b"))
		l.Accent = rgb(choose(r, "#f4ecd8", "#ffffff", "#e8e0c8"))
		switch l.Top {
		case "peacoat":
			if r.IntN(2) == 0 {
				l.Neck = "scarf"
				l.NeckColor = rgb(choose(r, "#b3283a", "#e8e0cc", "#3a5a3a", "#6a2a2a"))
			}
		case "slicker":
			l.Hat = "souwester"
			l.HatColor = l.Coat
		case "stripes", "sweater":
			if r.IntN(3) == 0 {
				l.Neck = "kerchief"
				l.NeckColor = rgb(choose(r, "#b3283a", "#e8b04a", "#3a7a9a", "#2a2a3a"))
			}
		}
		if l.Hat == "" {
			switch n := r.IntN(100); {
			case n < 22:
				l.Hat = "captain"
				l.HatColor = rgb(choose(r, "#1c2a44", "#1e1e24", "#f2efe6"))
			case n < 46:
				l.Hat = "beanie"
				l.HatColor = rgb(choose(r, "#1c2a44", "#3a3a3a", "#6e1f2a", "#5a4a3a", "#2f4a3a"))
			case n < 60:
				l.Hat = "cap"
				l.HatColor = rgb(choose(r, "#5a4a3a", "#3a3a4a", "#6b5a3a", "#4a3a2a"))
			}
		}
		switch n := r.IntN(100); {
		case n < 14:
			l.Mustache = "pencil"
		case n < 30:
			l.Mustache = "handlebar"
		case n < 48:
			l.Mustache = "walrus"
		}
		l.Beard = r.IntN(100) < 32
		if r.IntN(100) < 22 {
			l.Glasses = "round"
		}
	}
	l.Dark = shade(l.Coat, .62)
	return l
}

/* ---------------- names and manners ---------------- */

// manner is the kind of regular a face suggests; it picks the title and the turns of phrase.
func (f Face) manner() string {
	l := f.look()
	grey := l.Brow != (color.RGBA{}) // only silver hair sets the brow colour
	if f.Female {
		switch {
		case grey:
			return "nana"
		case l.Top == "slicker" || l.Hat == "beanie" || l.Top == "stripes":
			return "skipper_f"
		case l.Neck == "pearls" || l.Hat == "cloche":
			return "aunt"
		}
		return "miss"
	}
	switch {
	case l.Hat == "captain":
		return "capn"
	case l.Top == "slicker":
		return "skipper"
	case grey:
		return "old"
	case l.Glasses == "round":
		return "doc"
	case l.Hat == "beanie" || l.Top == "sweater":
		return "keeper"
	}
	return "young"
}

var mannerNames = map[string]struct {
	titles []string
	names  []string
}{
	"nana":      {[]string{"Nana", "Grammy"}, []string{"Ruth", "Edna", "Dot", "Mae", "Hazel", "Pearl", "Irene", "Agnes"}},
	"skipper_f": {[]string{"Skipper"}, []string{"Josie", "Tess", "Mags", "Frankie", "Bev", "Sal", "Jo", "Rosie"}},
	"aunt":      {[]string{"Aunt"}, []string{"Bea", "Hattie", "Lottie", "Winnie", "Marjorie", "Prudence", "Opal", "Lou"}},
	"miss":      {[]string{"Miss"}, []string{"Abigail", "Clara", "Nora", "Libby", "Molly", "Ada", "Lucy", "Emmeline"}},
	"capn":      {[]string{"Cap'n"}, []string{"Eli", "Amos", "Silas", "Jonah", "Ezra", "Caleb", "Thaddeus", "Obed"}},
	"skipper":   {[]string{"Skipper"}, []string{"Gus", "Hank", "Moe", "Butch", "Clem", "Dutch", "Lew", "Sonny"}},
	"old":       {[]string{"Old", "Grampa"}, []string{"Elmer", "Harlan", "Ike", "Otis", "Rufus", "Wendell", "Virgil", "Ansel"}},
	"doc":       {[]string{"Doc"}, []string{"Whitcomb", "Pennell", "Thayer", "Hallett", "Bartlett", "Crowell", "Snow", "Libby"}},
	"keeper":    {[]string{"Keeper"}, []string{"Asa", "Jeb", "Nate", "Ned", "Abner", "Seth", "Zeke", "Hollis"}},
	"young":     {[]string{"Young"}, []string{"Will", "Ben", "Sam", "Charlie", "Toby", "Nat", "Jamie", "Henry"}},
}

func (f Face) mannerKey() string { return f.manner() }

// nameFor gives the face a name (first name, and the full name with a title),
// avoiding names already at the table.
func (f Face) nameFor(avoid map[string]bool) (name, full string, ok bool) {
	mn := mannerNames[f.mannerKey()]
	r := f.rng(0x9a3e5)
	names := append([]string(nil), mn.names...)
	r.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })
	for _, n := range names {
		if !avoid[n] {
			return n, choose(r, mn.titles...) + " " + n, true
		}
	}
	return "", "", false
}

/* ---------------- table talk ---------------- */

// {p}: whoever the line is about.
var mannerLines = map[string]map[string][]string{
	"nana": {
		"peg":   {"Oh, that's lovely.", "Two for Nana!", "Every little bit, dear."},
		"pair":  {"A pair, how nice.", "Oh, a pair!"},
		"royal": {"Three of a kind! Goodness!", "Pair royal, dear. Six for me."},
		"big":   {"Look at that, a lovely hand.", "Oh my, count that twice!"},
		"zero":  {"Nineteen. Story of my life.", "Not a thing, dear."},
		"heels": {"His heels! Two for the dealer."},
		"win":   {"Who wants pie?"}, "lose": {"Well played, sweetheart."}, "skunk": {"Oh dear, a skunk. Sorry, love."}},
	"skipper_f": {
		"peg":   {"Two for me!", "Pegging along nicely.", "Ha! Got it."},
		"pair":  {"Pair!", "Snap!"},
		"royal": {"Pair royal! Six!", "Three of 'em!"},
		"big":   {"Now that's a catch!", "Haul it in!"},
		"zero":  {"Nineteen. Empty traps.", "Skunked on that one."},
		"heels": {"His heels! Lucky cut."},
		"win":   {"Out first! Ha!"}, "lose": {"Next tide, then."}, "skunk": {"Skunked ya!"}},
	"aunt": {
		"peg":   {"Thank you kindly.", "I'll take those.", "Two, if you please."},
		"pair":  {"A pair. Naturally.", "How fortunate."},
		"royal": {"Pair royal. Do try to keep up.", "Six, I believe."},
		"big":   {"A rather fine hand.", "Do count it, dear. Slowly."},
		"zero":  {"Nineteen. How tiresome.", "Nothing at all."},
		"heels": {"His heels. Two for me."},
		"win":   {"That will do nicely."}, "lose": {"Well played, everyone."}, "skunk": {"A skunk. How unfortunate. For you."}},
	"miss": {
		"peg":   {"Yes! Two!", "Ooh, points!", "Peg me!"},
		"pair":  {"A pair!", "Got you!"},
		"royal": {"Three in a row! Pair royal!", "Six! Did you see?"},
		"big":   {"Look at my hand!", "So many fifteens!"},
		"zero":  {"Nineteen? Oh no.", "Zero. Ugh."},
		"heels": {"His heels! Lucky me!"},
		"win":   {"I won! Again?"}, "lose": {"Okay, rematch!"}, "skunk": {"A skunk! Sorry!"}},
	"capn": {
		"peg":   {"Aye, two.", "Steady as she goes.", "Mark it, mate."},
		"pair":  {"Pair. Aye.", "Two of a kind."},
		"royal": {"Pair royal! Six, by thunder!", "Three of a kind! Hoist it up!"},
		"big":   {"A full hold!", "That's a fine catch."},
		"zero":  {"Nineteen. Becalmed.", "Nothing in the nets."},
		"heels": {"His heels! Two for the dealer."},
		"win":   {"Land ho! Pegged out!"}, "lose": {"We'll sail again."}, "skunk": {"Skunked, by thunder!"}},
	"skipper": {
		"peg":   {"Two for the skipper.", "Yup.", "Ayuh, two."},
		"pair":  {"Pair.", "Matched ya."},
		"royal": {"Pair royal. Wicked good.", "Six! Ayuh!"},
		"big":   {"Traps are full today.", "Wicked good hand."},
		"zero":  {"Nineteen. Wicked bad.", "Not a lobster in the trap."},
		"heels": {"His heels. Ayuh."},
		"win":   {"Ayuh. Pegged out."}, "lose": {"Can't win 'em all."}, "skunk": {"Skunked ya, bub."}},
	"old": {
		"peg":   {"Heh. Two.", "Still got it.", "Peg that."},
		"pair":  {"A pair. Heh.", "Matched."},
		"royal": {"Pair royal! Like in '52!", "Six, young'un."},
		"big":   {"Now that's a hand.", "Haven't seen that since the big storm."},
		"zero":  {"Nineteen. Bah.", "Nothin'."},
		"heels": {"His heels. Heh heh."},
		"win":   {"Not bad for an old salt!"}, "lose": {"Bah. Another round."}, "skunk": {"Skunked. Heh."}},
	"doc": {
		"peg":   {"Two, as expected.", "Precisely two.", "Noted: two."},
		"pair":  {"A pair. Probable.", "Ah, a pair."},
		"royal": {"Pair royal. Statistically pleasing.", "Six. Remarkable."},
		"big":   {"An elegant hand.", "Fascinating. Count it twice."},
		"zero":  {"Nineteen. Curious.", "Nothing. A null result."},
		"heels": {"His heels. A one-in-thirteen chance."},
		"win":   {"A tidy result."}, "lose": {"Well played. I shall review my notes."}, "skunk": {"A skunk, I'm afraid."}},
	"keeper": {
		"peg":   {"Two. Light's on.", "Steady.", "That'll do."},
		"pair":  {"Pair.", "Matched."},
		"royal": {"Pair royal! Bright as the lamp.", "Six!"},
		"big":   {"Now that shines.", "Good hand, that."},
		"zero":  {"Nineteen. Fog's rolled in.", "Dark as a storm."},
		"heels": {"His heels. Two for me."},
		"win":   {"Pegged out. Light the lamp!"}, "lose": {"Good game. Back to the tower."}, "skunk": {"Skunked, I'm afraid."}},
	"young": {
		"peg":   {"Two!", "Nice!", "Pegging!"},
		"pair":  {"Pair!", "Gotcha!"},
		"royal": {"Pair royal! Six!", "Three of 'em!"},
		"big":   {"Check this hand out!", "Count it!"},
		"zero":  {"Nineteen. Aw.", "Nothing. Aw man."},
		"heels": {"His heels! Sweet!"},
		"win":   {"I won! Wicked!"}, "lose": {"Aw. Again?"}, "skunk": {"A skunk! Whoa!"}},
}

func (f Face) lines(kind string) []string { return mannerLines[f.mannerKey()][kind] }

/* ---------------- picking regulars ---------------- */

// folk is one computer player picked for a game.
type folk struct{ face, name, full string }

// randomFolks makes up n regulars with different names (a mix of ladies and gentlemen).
func randomFolks(rng *rand.Rand, n int, avoidFaces, avoidNames map[string]bool) []folk {
	var out []folk
	men := 0
	for tries := 0; len(out) < n && tries < 500; tries++ {
		f := Face{Female: rng.IntN(2) == 0, Seed: rng.Uint64()}
		if len(out) == n-1 && n > 1 && men == 0 {
			f.Female = false // always at least one gentleman and one lady
		} else if len(out) == n-1 && n > 1 && men == len(out) {
			f.Female = true
		}
		if avoidFaces[f.String()] {
			continue
		}
		name, full, ok := f.nameFor(avoidNames)
		if !ok {
			continue
		}
		avoidNames[name] = true
		if !f.Female {
			men++
		}
		out = append(out, folk{face: f.String(), name: name, full: full})
	}
	return out
}

// pickerFaces are the faces offered for your own player online and on a shared computer.
var pickerFaces = func() []string {
	var out []string
	r := rand.New(rand.NewPCG(2026, 11))
	for len(out) < 10 {
		f := Face{Female: len(out)%2 == 0, Seed: r.Uint64()}
		if f.mannerKey() == "skipper" && len(out) > 6 {
			continue
		}
		out = append(out, f.String())
	}
	return out
}()
