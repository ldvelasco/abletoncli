package main

import (
	"bytes"
	"net"
	"testing"
	"time"
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

func TestStream(t *testing.T) {
	// Discover a free response port.
	tmp, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	respPort := tmp.LocalAddr().(*net.UDPAddr).Port
	tmp.Close()

	// Mock AbletonOSC server on an ephemeral port.
	mock, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()
	mockPort := mock.LocalAddr().(*net.UDPAddr).Port

	startAddrCh := make(chan string, 1)
	go func() {
		buf := make([]byte, 65536)
		n, _, _ := mock.ReadFromUDP(buf)
		gotAddr, _, _ := parseMessage(buf[:n])
		startAddrCh <- gotAddr
		reply := buildMessage("/live/song/get/beat", []OscValue{{Type: 'i', I: 42}})
		_, _ = mock.WriteToUDP(reply, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: respPort})
	}()

	c := &Client{Host: "127.0.0.1", Port: mockPort, ResponsePort: respPort, Timeout: time.Second}

	var gotAddr string
	var got []OscValue
	err = c.Stream("/live/song/start_listen/beat", nil, "/live/song/stop_listen/beat", nil, time.Second,
		func(addr string, vals []OscValue) error {
			gotAddr = addr
			got = vals
			return nil
		})
	if err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if gotAddr != "/live/song/get/beat" {
		t.Fatalf("got addr %q", gotAddr)
	}
	if len(got) != 1 || got[0].I != 42 {
		t.Fatalf("got %+v", got)
	}
	if s := <-startAddrCh; s != "/live/song/start_listen/beat" {
		t.Fatalf("start addr = %q", s)
	}
}
