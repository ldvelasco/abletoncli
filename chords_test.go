package main

import (
	"reflect"
	"testing"
)

func TestNoteNameToMIDI(t *testing.T) {
	cases := map[string]int{
		"C4":   60,
		"A4":   69,
		"F#4":  66,
		"Bb4":  70,
		"Eb3":  51,
		"G9":   127,
		"c4":   60,
		"C#4":  61,
		"B3":   59,
	}
	for name, want := range cases {
		got, err := noteNameToMIDI(name)
		if err != nil {
			t.Errorf("noteNameToMIDI(%q) error: %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("noteNameToMIDI(%q) = %d, want %d", name, got, want)
		}
	}
}

func TestNoteNameToMIDIInvalid(t *testing.T) {
	for _, name := range []string{"", "C", "H4", "C", "C#", "X#4"} {
		if _, err := noteNameToMIDI(name); err == nil {
			t.Errorf("noteNameToMIDI(%q) expected error", name)
		}
	}
}

func TestParseChordSymbol(t *testing.T) {
	cases := []struct {
		symbol string
		want   []int
	}{
		{"C", []int{60, 64, 67}},
		{"Cmaj7", []int{60, 64, 67, 71}},
		{"Am", []int{69, 72, 76}},
		{"C#m7", []int{61, 64, 68, 71}},
		{"G7", []int{67, 71, 74, 77}},
		{"Bdim", []int{71, 74, 77}},
		{"Fsus4", []int{65, 70, 72}},
	}
	for _, c := range cases {
		got, err := parseChordSymbol(c.symbol, 4)
		if err != nil {
			t.Errorf("parseChordSymbol(%q) error: %v", c.symbol, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseChordSymbol(%q) = %v, want %v", c.symbol, got, c.want)
		}
	}
}

func TestParseChordSpec(t *testing.T) {
	cases := []struct {
		spec string
		want [][]int
	}{
		{"C4,E4,G4", [][]int{{60, 64, 67}}},
		{"Cmaj7", [][]int{{60, 64, 67, 71}}},
		{"C4", [][]int{{60}}},
		{"C Am F G", [][]int{{60, 64, 67}, {69, 72, 76}, {65, 69, 72}, {67, 71, 74}}},
		{"C7", [][]int{{60, 64, 67, 70}}},
		{"C4,E4,G4 F4,A4,C5", [][]int{{60, 64, 67}, {65, 69, 72}}},
	}
	for _, c := range cases {
		got, err := parseChordSpec(c.spec)
		if err != nil {
			t.Errorf("parseChordSpec(%q) error: %v", c.spec, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseChordSpec(%q) = %v, want %v", c.spec, got, c.want)
		}
	}
}
