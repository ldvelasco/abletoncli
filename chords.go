package main

import (
	"fmt"
	"strconv"
	"strings"
)

// defaultChordOctave is the octave used when a chord symbol (e.g. "Cmaj7")
// has no explicit octave. Middle C is C4 = 60.
const defaultChordOctave = 4

var letterSemitone = map[byte]int{
	'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11,
}

var chordQualities = map[string][]int{
	"":     {0, 4, 7},
	"m":    {0, 3, 7},
	"min":  {0, 3, 7},
	"maj":  {0, 4, 7},
	"6":    {0, 4, 7, 9},
	"m6":   {0, 3, 7, 9},
	"7":    {0, 4, 7, 10},
	"maj7": {0, 4, 7, 11},
	"m7":   {0, 3, 7, 10},
	"m7b5": {0, 3, 6, 10},
	"dim":  {0, 3, 6},
	"aug":  {0, 4, 8},
	"sus2": {0, 2, 7},
	"sus4": {0, 5, 7},
}

// noteNameToMIDI converts a note name such as "C4", "F#3" or "Bb4" to a MIDI
// note number. Middle C is C4 = 60.
func noteNameToMIDI(name string) (int, error) {
	if len(name) < 2 {
		return 0, fmt.Errorf("invalid note name %q", name)
	}
	letter := name[0]
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	semi, ok := letterSemitone[letter]
	if !ok {
		return 0, fmt.Errorf("invalid note name %q (letter must be A-G)", name)
	}
	rest := name[1:]
	acc := 0
	if len(rest) > 0 {
		switch rest[0] {
		case '#':
			acc = 1
			rest = rest[1:]
		case 'b':
			acc = -1
			rest = rest[1:]
		}
	}
	octave, err := strconv.Atoi(rest)
	if err != nil {
		return 0, fmt.Errorf("invalid octave in note %q", name)
	}
	midi := (octave+1)*12 + semi + acc
	if midi < 0 || midi > 127 {
		return 0, fmt.Errorf("note %q out of MIDI range (0-127)", name)
	}
	return midi, nil
}

// parseChordSymbol expands a chord symbol such as "Cmaj7" or "F#m7" into a
// list of MIDI note numbers at the given octave.
func parseChordSymbol(token string, octave int) ([]int, error) {
	if token == "" {
		return nil, fmt.Errorf("empty chord symbol")
	}
	letter := token[0]
	if letter >= 'a' && letter <= 'z' {
		letter -= 'a' - 'A'
	}
	semi, ok := letterSemitone[letter]
	if !ok {
		return nil, fmt.Errorf("invalid chord root %q", token)
	}
	rest := token[1:]
	acc := 0
	if len(rest) > 0 {
		switch rest[0] {
		case '#':
			acc = 1
			rest = rest[1:]
		case 'b':
			acc = -1
			rest = rest[1:]
		}
	}
	quality := strings.ToLower(rest)
	intervals, ok := chordQualities[quality]
	if !ok {
		return nil, fmt.Errorf("unknown chord quality %q in %q", quality, token)
	}
	root := (octave+1)*12 + semi + acc
	notes := make([]int, len(intervals))
	for i, iv := range intervals {
		notes[i] = root + iv
	}
	return notes, nil
}

// parseChordSpec parses a chord spec into a list of chords, where each chord
// is a list of MIDI note numbers.
//
// A spec is a whitespace-separated list of chords. Each chord is either an
// explicit set of note names separated by commas ("C4,E4,G4") or a chord
// symbol ("Cmaj7", "Am", "G7"). A lone note name ("C4") is a single note.
func parseChordSpec(spec string) ([][]int, error) {
	tokens := strings.Fields(spec)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty chord spec")
	}
	chords := make([][]int, 0, len(tokens))
	for _, tok := range tokens {
		if strings.Contains(tok, ",") {
			parts := strings.Split(tok, ",")
			notes := make([]int, 0, len(parts))
			for _, p := range parts {
				n, err := noteNameToMIDI(strings.TrimSpace(p))
				if err != nil {
					return nil, err
				}
				notes = append(notes, n)
			}
			chords = append(chords, notes)
			continue
		}
		if notes, err := parseChordSymbol(tok, defaultChordOctave); err == nil {
			chords = append(chords, notes)
			continue
		}
		if n, err := noteNameToMIDI(tok); err == nil {
			chords = append(chords, []int{n})
			continue
		}
		return nil, fmt.Errorf("cannot parse chord %q (use note names like C4,E4,G4 or chord symbols like Cmaj7)", tok)
	}
	return chords, nil
}
