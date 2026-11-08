package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type saveData struct {
	Settings struct {
		Music  bool   `json:"music"`
		Sfx    bool   `json:"sfx"`
		Vol    int    `json:"vol"`
		Target  int    `json:"target"`
		Players int    `json:"players"`
		Speed   string `json:"speed"`
		// Play Together
		Name     string   `json:"name,omitempty"`
		Face     string   `json:"face,omitempty"`
		Host     string   `json:"host,omitempty"`
		HotN     int      `json:"hotN,omitempty"`
		HotNames []string `json:"hotNames,omitempty"`
		HotFaces []string `json:"hotFaces,omitempty"`
	} `json:"settings"`
	Stats struct {
		Played      int `json:"played"`
		Won         int `json:"won"`
		Lost        int `json:"lost"`
		Hands       int `json:"hands"`
		Skunks      int `json:"skunks"`   // games won by a skunk or more
		BestHand    int `json:"bestHand"` // the best hand counted
		TwentyNines int `json:"twentyNines"`
		Streak      int `json:"streak"`
		BestStreak  int `json:"bestStreak"`
	} `json:"stats"`
}

func defaultSave() saveData {
	var s saveData
	s.Settings.Music, s.Settings.Sfx, s.Settings.Vol = true, true, 60
	s.Settings.Target, s.Settings.Players, s.Settings.Speed = 121, 2, "normal"
	return s
}

// dataDir is %APPDATA%\Ironite\Cribbage on Windows.
func dataDir() string {
	if d := os.Getenv("CRIBBAGE_DATA"); d != "" { // test builds
		return d
	}
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "Ironite", productName)
}

func loadSave() saveData {
	s := defaultSave()
	if dir := dataDir(); dir != "" {
		if b, err := os.ReadFile(filepath.Join(dir, "save.json")); err == nil {
			_ = json.Unmarshal(b, &s)
		}
	}
	if s.Settings.Target != 61 && s.Settings.Target != 121 {
		s.Settings.Target = 121
	}
	if s.Settings.Players < 2 || s.Settings.Players > 4 {
		s.Settings.Players = 2
	}
	if _, ok := paceOf[s.Settings.Speed]; !ok {
		s.Settings.Speed = "normal"
	}
	s.Settings.Vol = min(100, max(0, s.Settings.Vol))
	return s
}

func (g *game) persist() {
	dir := dataDir()
	if dir == "" {
		return
	}
	b, err := json.MarshalIndent(g.save, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "save.json")
	if os.WriteFile(path+".tmp", b, 0o644) == nil {
		_ = os.Rename(path+".tmp", path)
	}
}
