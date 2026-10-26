package main

import "testing"

func TestResolveArrangementCommands(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		address  string
		query    bool
		argCount int
	}{
		{name: "arr-midi", args: []string{"2", "8", "16"}, address: "/live/track/create_midi_clip", argCount: 3},
		{name: "arr-audio", args: []string{"1", `C:\audio\loop.wav`, "4"}, address: "/live/track/create_audio_clip", argCount: 3},
		{name: "arr-clip-name", args: []string{"3", "0"}, address: "/live/arrangement_clip/get/name", query: true, argCount: 2},
		{name: "arr-clip-add-notes", args: []string{"0", "1", "60", "0", "1", "100", "off"}, address: "/live/arrangement_clip/add/notes", argCount: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := resolveCommand(tt.name, tt.args)
			if err != nil {
				t.Fatalf("resolveCommand() error = %v", err)
			}
			if len(results) != 1 {
				t.Fatalf("got %d results, want 1", len(results))
			}
			got := results[0]
			if got.address != tt.address || got.query != tt.query || len(got.args) != tt.argCount {
				t.Fatalf("got %+v, want address=%q query=%t args=%d", got, tt.address, tt.query, tt.argCount)
			}
		})
	}
}

func TestResolveArrangementClipList(t *testing.T) {
	results, err := resolveCommand("arr-clips", []string{"4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d queries, want 3", len(results))
	}
	want := []string{
		"/live/track/get/arrangement_clips/name",
		"/live/track/get/arrangement_clips/start_time",
		"/live/track/get/arrangement_clips/length",
	}
	for i, result := range results {
		if result.address != want[i] || !result.query || len(result.args) != 1 || result.args[0].I != 4 {
			t.Errorf("result %d = %+v", i, result)
		}
	}
}
