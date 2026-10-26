package main

import (
	"bytes"
	"testing"
)

func TestBuildParseRoundTrip(t *testing.T) {
	args := []OscValue{
		{Type: 's', S: "hola"},
		{Type: 'i', I: -42},
		{Type: 'f', F: 123.5},
		{Type: 'd', D: 3.14159},
		{Type: 'T'},
		{Type: 'F'},
		{Type: 'N'},
	}
	msg := buildMessage("/live/test", args)
	addr, got, err := parseMessage(msg)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if addr != "/live/test" {
		t.Fatalf("addr = %q, want /live/test", addr)
	}
	if len(got) != len(args) {
		t.Fatalf("len = %d, want %d", len(got), len(args))
	}
	if got[0].S != "hola" || got[1].I != -42 || got[2].F != 123.5 ||
		got[3].D != 3.14159 || got[4].Type != 'T' || got[5].Type != 'F' || got[6].Type != 'N' {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestBuildNoArgs(t *testing.T) {
	msg := buildMessage("/live/song/start_playing", nil)
	addr, args, err := parseMessage(msg)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if addr != "/live/song/start_playing" {
		t.Fatalf("addr = %q", addr)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %d", len(args))
	}
}

func TestBuildMatchesPythonOscFormat(t *testing.T) {
	// "/live/song/set/tempo" (20 bytes) + 4 nulls, tag ",f" + 2 nulls,
	// then float32 123.0 = 0x42f60000 big-endian.
	msg := buildMessage("/live/song/set/tempo", []OscValue{{Type: 'f', F: 123.0}})
	expected := []byte("/live/song/set/tempo\x00\x00\x00\x00,f\x00\x00\x42\xf6\x00\x00")
	if !bytes.Equal(msg, expected) {
		t.Fatalf("build mismatch:\n got % x\nwant % x", msg, expected)
	}
}

func TestOscStringPadding(t *testing.T) {
	var b bytes.Buffer
	oscString(&b, "ab")  // 2 chars -> pad 2 -> 4 total
	if b.Len() != 4 {
		t.Fatalf("len = %d, want 4", b.Len())
	}
	b.Reset()
	oscString(&b, "abcd") // 4 chars -> pad 4 -> 8 total (terminator included)
	if b.Len() != 8 {
		t.Fatalf("len = %d, want 8", b.Len())
	}
	b.Reset()
	oscString(&b, "") // 0 chars -> pad 4 -> 4 total
	if b.Len() != 4 {
		t.Fatalf("len = %d, want 4", b.Len())
	}
}
