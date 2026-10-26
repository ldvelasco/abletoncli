package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// commandResult describes what to send to AbletonOSC.
type commandResult struct {
	address string
	args    []OscValue
	query   bool // wait for a reply
	trim    int  // number of leading reply args to drop when printing
}

// resolveCommand maps a friendly command to an OSC address and arguments.
func resolveCommand(cmd string, args []string) (commandResult, error) {
	switch cmd {
	case "play":
		return commandResult{address: "/live/song/start_playing"}, nil
	case "stop":
		return commandResult{address: "/live/song/stop_playing"}, nil

	case "tempo":
		if len(args) == 0 {
			return commandResult{address: "/live/song/get/tempo", query: true}, nil
		}
		f, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return commandResult{}, fmt.Errorf("invalid tempo %q", args[0])
		}
		return commandResult{
			address: "/live/song/set/tempo",
			args:    []OscValue{{Type: 'f', F: float32(f)}},
		}, nil

	case "tracks":
		return commandResult{address: "/live/song/get/track_names", query: true}, nil

	case "clips":
		if len(args) == 0 {
			return commandResult{}, fmt.Errorf("clips requires a track index")
		}
		i, err := strconv.Atoi(args[0])
		if err != nil {
			return commandResult{}, fmt.Errorf("invalid track index %q", args[0])
		}
		return commandResult{
			address: "/live/track/get/clips/name",
			args:    []OscValue{{Type: 'i', I: int32(i)}},
			query:   true,
			trim:    1, // drop the echoed track index
		}, nil

	case "scene":
		if len(args) == 0 {
			return commandResult{}, fmt.Errorf("scene requires a scene index")
		}
		i, err := strconv.Atoi(args[0])
		if err != nil {
			return commandResult{}, fmt.Errorf("invalid scene index %q", args[0])
		}
		return commandResult{
			address: "/live/scene/fire",
			args:    []OscValue{{Type: 'i', I: int32(i)}},
		}, nil

	case "raw":
		if len(args) == 0 {
			return commandResult{}, fmt.Errorf("raw requires an OSC address")
		}
		addr := args[0]
		vals := make([]OscValue, 0, len(args)-1)
		for _, a := range args[1:] {
			vals = append(vals, inferValue(a))
		}
		return commandResult{
			address: addr,
			args:    vals,
			query:   strings.Contains(addr, "/get/"),
		}, nil

	default:
		return commandResult{}, fmt.Errorf("unknown command %q (run with no arguments for help)", cmd)
	}
}

// inferValue guesses an OSC argument type from its text representation:
// bool keywords, integers, floats, otherwise a string.
func inferValue(s string) OscValue {
	switch s {
	case "true":
		return OscValue{Type: 'T'}
	case "false":
		return OscValue{Type: 'F'}
	case "null", "nil", "None":
		return OscValue{Type: 'N'}
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		if i >= math.MinInt32 && i <= math.MaxInt32 {
			return OscValue{Type: 'i', I: int32(i)}
		}
		return OscValue{Type: 'h', I64: i}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return OscValue{Type: 'f', F: float32(f)}
	}
	return OscValue{Type: 's', S: s}
}
