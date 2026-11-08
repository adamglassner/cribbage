# Cribbage

A lighthouse-cottage cribbage game by Ironite, written in Go with
[Ebitengine](https://ebitengine.org). Play against the regulars at the Keeper's
Cottage, share one computer with up to three friends, or play together online.
It ships on its own and inside the [Ironite Launcher](https://github.com/adamglassner/launcher).

## How to play

- Two or three players each keep their own score; four play as partners across the table.
- Two players get six cards and lay away two; three or four get five and lay away one (with three, a card off the pack fills the crib). The crib is the dealer's.
- The starter is cut and turned up; a jack is "his heels", 2 for the dealer.
- Pegging: play in turn without passing 31. 15 or 31 is 2, a pair 2, a pair royal 6, a double pair royal 12, runs 1 a card. Say "Go" when you can't play; the last card pegs 1.
- The show: hands then the crib, with the starter: fifteens, pairs, runs, flushes and his nobs.
- First to 121 (or 61) wins. Below 91 (31) is a skunk, below 61 at 121 a double skunk.
- F2 starts a new game, M toggles music, F11 toggles full screen.

### Ways to play

- **Deal** – you against the cottage regulars. Every game makes up new ones, each with a face, and a name and manner that fit it.
- **Play Together → Same Computer** – 2 to 4 people take turns at one PC with a "pass the computer" screen between turns.
- **Play Together → Online** – one player hosts, friends join by IP address (TCP port 47261). The host runs the table and sends each guest only what they can see. A guest who leaves is replaced by a computer player.

Saves and settings are kept in `%APPDATA%\Ironite\Cribbage\save.json`, outside the install folder.

## Play

The easiest way is **[Ironite Launcher](https://github.com/adamglassner/launcher)**: it downloads
Cribbage from this repository's releases, installs it and keeps it up to date.

Or download from the [latest release](../../releases/latest): `Cribbage-Setup-v<version>.exe`
installs it with Start menu and desktop shortcuts, or unzip `Cribbage-<version>-windows.zip`
anywhere and run `Cribbage.exe`.

## Code

| File | What it is |
|---|---|
| `main.go` | Window setup and the entry point. |
| `game.go` | Game state, card zones laid out from each player's seat, the hand script (deal, crib, cut, pegging, the show), results. |
| `internal/crib/` | The rules, scoring and the computer players (no graphics; has tests). |
| `board.go` | The cribbage board, the running count, score pops and the tally box. |
| `net.go` | Online play: host/guest protocol (JSON lines over TCP). |
| `setup.go`, `setup_ui.go` | Play Together: seating, lobby, pass screen, text fields, face picker. |
| `faces.go`, `cameo.go` | The cottage regulars: random faces, fitting names and table talk, and their portraits drawn in Go. |
| `ui.go`, `panels.go` | The table, top bar, crib box and banner; title, rules, credits and result panels. |
| `draw.go`, `glyphs.go`, `assets.go` | Vector drawing, text and suit symbols, embedded art. |
| `sound.go`, `save.go`, `common.go` | Music and sound effects, the save file, shared bits. |
| `assets/` | Card faces, props, the pine table, fonts, music and sound effects. |
| `tools/bake/` | Draws the card art and store art and synthesizes the music and sounds into `assets/` (`node tools/bake/bake.mjs`, needs Playwright and ffmpeg; set `CHROME_PATH` to use an installed Chrome). |
| `tools/mkres/` | Writes the Windows icon and version info into the .exe. |
| `installer/` | Inno Setup scripts for the standalone `Cribbage-Setup-v<version>.exe`. |

`headless_test_main.go` and `sound_stub.go` are only used by test builds:
`go run -tags headless .` plays two-, three- and four-handed games against
the computer, two online games between a host and two guests over TCP on this
computer, and a shared-computer game through the pass screen.
`go test ./internal/crib` checks the scoring and plays a few hundred computer games.

## Building

You need Go 1.24 or newer and bash (Git Bash works on Windows). No C compiler is needed.

```bash
./build.sh
```

This writes `dist/Cribbage/Cribbage.exe` with its README, CREDITS, CHANGELOG and license files.
The version comes from the top entry of `CHANGELOG.md`. For the installer, install
Inno Setup 6.3 or newer, then: `ISCC.exe /DGameVersion=<version> installer\CribbageSetup.iss`

## Releasing

Add a new version section at the top of `CHANGELOG.md` and push to `main`. The
release workflow builds the game, publishes `Cribbage-<version>-windows.zip` as a
GitHub release (Ironite Launcher installs and updates from it), then builds and
attaches the standalone `Cribbage-Setup-v<version>.exe`.

(c) 2026 Ironite
