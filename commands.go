package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// commandResult describes one OSC message to send to AbletonOSC.
// A command can expand to several results (e.g. setting time signature
// numerator and denominator). At most the final results are queries.
type commandResult struct {
	address string
	args    []OscValue
	query   bool   // wait for a reply
	trim    int    // number of leading reply args to drop when printing
	label   string // optional label prefix for the printed value
	stream  *streamSpec
}

// streamSpec describes a start_listen/stream/stop_listen flow.
type streamSpec struct {
	startAddr string
	startArgs []OscValue
	stopAddr  string
	stopArgs  []OscValue
}

// --- argument helpers -------------------------------------------------------

func parseBoolWord(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "on", "yes", "1", "enable", "enabled":
		return true, nil
	case "false", "off", "no", "0", "disable", "disabled":
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean %q (use on/off or true/false)", s)
}

func oscBoolWord(s string) (OscValue, error) {
	b, err := parseBoolWord(s)
	if err != nil {
		return OscValue{}, err
	}
	if b {
		return OscValue{Type: 'T'}, nil
	}
	return OscValue{Type: 'F'}, nil
}

func oscInt(s, name string) (OscValue, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return OscValue{}, fmt.Errorf("invalid %s %q", name, s)
	}
	if i < math.MinInt32 || i > math.MaxInt32 {
		return OscValue{}, fmt.Errorf("%s %q out of int32 range", name, s)
	}
	return OscValue{Type: 'i', I: int32(i)}, nil
}

func oscFloat(s, name string) (OscValue, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return OscValue{}, fmt.Errorf("invalid %s %q", name, s)
	}
	return OscValue{Type: 'f', F: float32(f)}, nil
}

func oscStr(s string) OscValue { return OscValue{Type: 's', S: s} }

// inferValue guesses an OSC argument type from its text representation.
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

// resolveCommand maps a friendly command to one or more OSC messages.
func resolveCommand(cmd string, args []string) ([]commandResult, error) {
	switch cmd {
	// ------------------------------------------------------------------ Song
	case "play":
		return []commandResult{{address: "/live/song/start_playing"}}, nil
	case "stop":
		return []commandResult{{address: "/live/song/stop_playing"}}, nil
	case "undo":
		return []commandResult{{address: "/live/song/undo"}}, nil
	case "redo":
		return []commandResult{{address: "/live/song/redo"}}, nil
	case "tap":
		return []commandResult{{address: "/live/song/tap_tempo"}}, nil
	case "export":
		return []commandResult{{address: "/live/song/export/structure"}}, nil

	case "tempo":
		if len(args) == 0 {
			return []commandResult{{address: "/live/song/get/tempo", query: true}}, nil
		}
		f, err := oscFloat(args[0], "tempo")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/song/set/tempo", args: []OscValue{f}}}, nil

	case "metronome":
		if len(args) == 0 {
			return []commandResult{{address: "/live/song/get/metronome", query: true}}, nil
		}
		b, err := oscBoolWord(args[0])
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/song/set/metronome", args: []OscValue{b}}}, nil

	case "loop":
		if len(args) == 0 {
			return []commandResult{{address: "/live/song/get/loop", query: true}}, nil
		}
		b, err := oscBoolWord(args[0])
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/song/set/loop", args: []OscValue{b}}}, nil

	case "record-mode":
		if len(args) == 0 {
			return []commandResult{{address: "/live/song/get/record_mode", query: true}}, nil
		}
		b, err := oscBoolWord(args[0])
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/song/set/record_mode", args: []OscValue{b}}}, nil

	case "signature":
		if len(args) == 0 {
			return []commandResult{
				{address: "/live/song/get/signature_numerator", query: true, label: "numerator"},
				{address: "/live/song/get/signature_denominator", query: true, label: "denominator"},
			}, nil
		}
		if len(args) != 2 {
			return nil, fmt.Errorf("signature expects [numerator denominator] (or no args to read)")
		}
		num, err := oscInt(args[0], "numerator")
		if err != nil {
			return nil, err
		}
		den, err := oscInt(args[1], "denominator")
		if err != nil {
			return nil, err
		}
		return []commandResult{
			{address: "/live/song/set/signature_numerator", args: []OscValue{num}},
			{address: "/live/song/set/signature_denominator", args: []OscValue{den}},
		}, nil

	case "create-track":
		if len(args) == 0 {
			return nil, fmt.Errorf("create-track requires a type (audio|midi|return)")
		}
		switch strings.ToLower(args[0]) {
		case "return":
			return []commandResult{{address: "/live/song/create_return_track"}}, nil
		case "audio", "midi":
			if len(args) < 2 {
				return nil, fmt.Errorf("create-track %s requires an index", args[0])
			}
			idx, err := oscInt(args[1], "track index")
			if err != nil {
				return nil, err
			}
			addr := "/live/song/create_midi_track"
			if strings.EqualFold(args[0], "audio") {
				addr = "/live/song/create_audio_track"
			}
			return []commandResult{{address: addr, args: []OscValue{idx}}}, nil
		default:
			return nil, fmt.Errorf("unknown track type %q (use audio|midi|return)", args[0])
		}

	case "create-scene":
		if len(args) == 0 {
			return nil, fmt.Errorf("create-scene requires an index")
		}
		idx, err := oscInt(args[0], "scene index")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/song/create_scene", args: []OscValue{idx}}}, nil

	case "cuepoints":
		return []commandResult{{address: "/live/song/get/cue_points", query: true}}, nil

	case "cue":
		if len(args) == 0 {
			return nil, fmt.Errorf("cue requires a cue point index or name")
		}
		return []commandResult{{address: "/live/song/cue_point/jump", args: []OscValue{inferValue(args[0])}}}, nil

	case "trackdata":
		if len(args) < 3 {
			return nil, fmt.Errorf("trackdata requires <min> <max> <property...> (e.g. trackdata 0 -1 track.name clip.name)")
		}
		min, err := oscInt(args[0], "min track")
		if err != nil {
			return nil, err
		}
		max, err := oscInt(args[1], "max track")
		if err != nil {
			return nil, err
		}
		vals := []OscValue{min, max}
		for _, p := range args[2:] {
			vals = append(vals, oscStr(p))
		}
		return []commandResult{{address: "/live/song/get/track_data", args: vals, query: true}}, nil

	case "beat":
		return []commandResult{{stream: &streamSpec{
			startAddr: "/live/song/start_listen/beat",
			stopAddr:  "/live/song/stop_listen/beat",
		}}}, nil

	case "listen":
		if len(args) == 0 {
			return nil, fmt.Errorf("listen requires an OSC address (e.g. /live/song/start_listen/beat)")
		}
		start := args[0]
		if strings.Contains(start, "/get/") {
			start = strings.Replace(start, "/get/", "/start_listen/", 1)
		}
		if !strings.Contains(start, "/start_listen/") {
			return nil, fmt.Errorf("listen address must be a start_listen address (got %q)", args[0])
		}
		stop := strings.Replace(start, "/start_listen/", "/stop_listen/", 1)
		var vals []OscValue
		for _, a := range args[1:] {
			vals = append(vals, inferValue(a))
		}
		return []commandResult{{stream: &streamSpec{
			startAddr: start,
			startArgs: vals,
			stopAddr:  stop,
			stopArgs:  vals,
		}}}, nil

	// ----------------------------------------------------------------- Track
	case "tracks":
		return []commandResult{{address: "/live/song/get/track_names", query: true}}, nil

	case "volume":
		return trackFloatProp("/live/track/get/volume", "/live/track/set/volume", "volume", args)
	case "pan":
		return trackFloatProp("/live/track/get/panning", "/live/track/set/panning", "panning", args)
	case "mute":
		return trackBoolProp("/live/track/get/mute", "/live/track/set/mute", args)
	case "solo":
		return trackBoolProp("/live/track/get/solo", "/live/track/set/solo", args)
	case "arm":
		return trackBoolProp("/live/track/get/arm", "/live/track/set/arm", args)
	case "track-fold":
		return trackBoolProp("/live/track/get/fold_state", "/live/track/set/fold_state", args)

	case "send":
		if len(args) < 2 {
			return nil, fmt.Errorf("send requires <track> <send-id>")
		}
		idx, err := oscInt(args[0], "track index")
		if err != nil {
			return nil, err
		}
		sid, err := oscInt(args[1], "send id")
		if err != nil {
			return nil, err
		}
		if len(args) == 2 {
			return []commandResult{{address: "/live/track/get/send", args: []OscValue{idx, sid}, query: true, trim: 2}}, nil
		}
		v, err := oscFloat(args[2], "send value")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/track/set/send", args: []OscValue{idx, sid, v}}}, nil

	case "track-name":
		return trackStrProp("/live/track/get/name", "/live/track/set/name", "name", args)
	case "track-color":
		return trackIntProp("/live/track/get/color_index", "/live/track/set/color_index", "color index", args)

	case "devices":
		if len(args) == 0 {
			return nil, fmt.Errorf("devices requires a track index")
		}
		idx, err := oscInt(args[0], "track index")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/track/get/devices/name", args: []OscValue{idx}, query: true, trim: 1}}, nil

	case "out-type":
		return trackStrProp("/live/track/get/output_routing_type", "/live/track/set/output_routing_type", "output routing type", args)
	case "out-channel":
		return trackStrProp("/live/track/get/output_routing_channel", "/live/track/set/output_routing_channel", "output routing channel", args)
	case "in-type":
		return trackStrProp("/live/track/get/input_routing_type", "/live/track/set/input_routing_type", "input routing type", args)
	case "in-channel":
		return trackStrProp("/live/track/get/input_routing_channel", "/live/track/set/input_routing_channel", "input routing channel", args)

	case "out-types":
		return trackListProp("/live/track/get/available_output_routing_types", args)
	case "out-channels":
		return trackListProp("/live/track/get/available_output_routing_channels", args)
	case "in-types":
		return trackListProp("/live/track/get/available_input_routing_types", args)
	case "in-channels":
		return trackListProp("/live/track/get/available_input_routing_channels", args)

	// -------------------------------------------------------------- Clip Slot
	case "slot-fire":
		return clipSlotAction("/live/clip_slot/fire", "fire", args)
	case "slot-stop":
		return clipSlotAction("/live/clip_slot/stop", "stop", args)
	case "slot-delete":
		return clipSlotAction("/live/clip_slot/delete_clip", "delete", args)

	case "slot-create":
		r, err := clipSlotIndexed(args, "create")
		if err != nil {
			return nil, err
		}
		length := OscValue{Type: 'f', F: 4.0}
		if len(args) > 2 {
			length, err = oscFloat(args[2], "clip length")
			if err != nil {
				return nil, err
			}
		}
		r.address = "/live/clip_slot/create_clip"
		r.args = append(r.args, length)
		return []commandResult{r}, nil

	case "slot-has-clip":
		return clipSlotGet("/live/clip_slot/get/has_clip", "has-clip", args)
	case "slot-playing":
		return clipSlotGet("/live/clip_slot/get/is_playing", "playing", args)

	// ------------------------------------------------------------------ Clip
	case "clip-fire":
		return clipAction("/live/clip/fire", "fire", args)
	case "clip-stop":
		return clipAction("/live/clip/stop", "stop", args)
	case "clip-name":
		return clipStrProp("/live/clip/get/name", "/live/clip/set/name", "name", args)
	case "clip-gain":
		return clipFloatProp("/live/clip/get/gain", "/live/clip/set/gain", "gain", args)
	case "clip-pitch":
		return clipIntProp("/live/clip/get/pitch_coarse", "/live/clip/set/pitch_coarse", "pitch", args)
	case "clip-muted":
		return clipBoolProp("/live/clip/get/muted", "/live/clip/set/muted", args)
	case "clip-loop":
		return clipBoolProp("/live/clip/get/looping", "/live/clip/set/looping", args)
	case "clip-warp":
		return clipBoolProp("/live/clip/get/warping", "/live/clip/set/warping", args)
	case "clip-color":
		return clipIntProp("/live/clip/get/color_index", "/live/clip/set/color_index", "color index", args)

	case "clip-length":
		r, err := clipIndexed(args, "clip-length")
		if err != nil {
			return nil, err
		}
		r.address = "/live/clip/get/length"
		r.query = true
		return []commandResult{r}, nil

	case "clip-notes":
		r, err := clipIndexed(args, "clip-notes")
		if err != nil {
			return nil, err
		}
		if len(args) > 2 {
			if len(args) != 6 {
				return nil, fmt.Errorf("clip-notes optional args: <pitchStart> <pitchSpan> <timeStart> <timeSpan>")
			}
			pitchStart, err := oscInt(args[2], "pitch start")
			if err != nil {
				return nil, err
			}
			pitchSpan, err := oscInt(args[3], "pitch span")
			if err != nil {
				return nil, err
			}
			timeStart, err := oscFloat(args[4], "time start")
			if err != nil {
				return nil, err
			}
			timeSpan, err := oscFloat(args[5], "time span")
			if err != nil {
				return nil, err
			}
			r.args = append(r.args, pitchStart, pitchSpan, timeStart, timeSpan)
		}
		r.address = "/live/clip/get/notes"
		r.query = true
		return []commandResult{r}, nil

	case "clip-add-notes":
		r, err := clipIndexed(args, "clip-add-notes")
		if err != nil {
			return nil, err
		}
		rest := args[2:]
		if len(rest) == 0 || len(rest)%5 != 0 {
			return nil, fmt.Errorf("clip-add-notes note data must be groups of 5 (pitch start duration velocity mute)")
		}
		for i := 0; i < len(rest); i += 5 {
			pitch, err := oscInt(rest[i], "pitch")
			if err != nil {
				return nil, err
			}
			start, err := oscFloat(rest[i+1], "start")
			if err != nil {
				return nil, err
			}
			dur, err := oscFloat(rest[i+2], "duration")
			if err != nil {
				return nil, err
			}
			vel, err := oscInt(rest[i+3], "velocity")
			if err != nil {
				return nil, err
			}
			mute, err := oscBoolWord(rest[i+4])
			if err != nil {
				return nil, err
			}
			r.args = append(r.args, pitch, start, dur, vel, mute)
		}
		r.address = "/live/clip/add/notes"
		return []commandResult{r}, nil

	case "clip-remove-notes":
		r, err := clipIndexed(args, "clip-remove-notes")
		if err != nil {
			return nil, err
		}
		if len(args) > 2 {
			if len(args) != 6 {
				return nil, fmt.Errorf("clip-remove-notes optional args: <pitchStart> <pitchSpan> <timeStart> <timeSpan>")
			}
			pitchStart, err := oscInt(args[2], "pitch start")
			if err != nil {
				return nil, err
			}
			pitchSpan, err := oscInt(args[3], "pitch span")
			if err != nil {
				return nil, err
			}
			timeStart, err := oscFloat(args[4], "time start")
			if err != nil {
				return nil, err
			}
			timeSpan, err := oscFloat(args[5], "time span")
			if err != nil {
				return nil, err
			}
			r.args = append(r.args, pitchStart, pitchSpan, timeStart, timeSpan)
		}
		r.address = "/live/clip/remove/notes"
		return []commandResult{r}, nil

	// ----------------------------------------------------------------- Scene
	case "scene":
		if len(args) == 0 {
			return nil, fmt.Errorf("scene requires a scene index")
		}
		idx, err := oscInt(args[0], "scene index")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/scene/fire", args: []OscValue{idx}}}, nil

	case "scene-fire-selected":
		if len(args) == 0 {
			return nil, fmt.Errorf("scene-fire-selected requires a scene index")
		}
		idx, err := oscInt(args[0], "scene index")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/scene/fire_as_selected", args: []OscValue{idx}}}, nil

	case "scene-name":
		return sceneStrProp("/live/scene/get/name", "/live/scene/set/name", "name", args)
	case "scene-color":
		return sceneIntProp("/live/scene/get/color_index", "/live/scene/set/color_index", "color index", args)
	case "scene-tempo":
		return sceneFloatProp("/live/scene/get/tempo", "/live/scene/set/tempo", "tempo", args)

	case "scene-signature":
		if len(args) == 0 {
			return nil, fmt.Errorf("scene-signature requires a scene index")
		}
		idx, err := oscInt(args[0], "scene index")
		if err != nil {
			return nil, err
		}
		if len(args) == 1 {
			return []commandResult{
				{address: "/live/scene/get/time_signature_numerator", args: []OscValue{idx}, query: true, trim: 1, label: "numerator"},
				{address: "/live/scene/get/time_signature_denominator", args: []OscValue{idx}, query: true, trim: 1, label: "denominator"},
			}, nil
		}
		if len(args) != 3 {
			return nil, fmt.Errorf("scene-signature expects <scene> [numerator denominator]")
		}
		num, err := oscInt(args[1], "numerator")
		if err != nil {
			return nil, err
		}
		den, err := oscInt(args[2], "denominator")
		if err != nil {
			return nil, err
		}
		return []commandResult{
			{address: "/live/scene/set/time_signature_numerator", args: []OscValue{idx, num}},
			{address: "/live/scene/set/time_signature_denominator", args: []OscValue{idx, den}},
		}, nil

	// ---------------------------------------------------------------- Device
	case "device-num-params":
		return deviceProp("/live/device/get/num_parameters", "device-num-params", args)
	case "device-params":
		return deviceProp("/live/device/get/parameters/name", "device-params", args)
	case "device-params-value":
		return deviceProp("/live/device/get/parameters/value", "device-params-value", args)
	case "device-params-min":
		return deviceProp("/live/device/get/parameters/min", "device-params-min", args)
	case "device-params-max":
		return deviceProp("/live/device/get/parameters/max", "device-params-max", args)
	case "device-param-name":
		return deviceParamProp("/live/device/get/parameter/name", "device-param-name", args)

	case "device-param":
		if len(args) < 3 {
			return nil, fmt.Errorf("device-param requires <track> <device> <param>")
		}
		idx, err := oscInt(args[0], "track index")
		if err != nil {
			return nil, err
		}
		didx, err := oscInt(args[1], "device index")
		if err != nil {
			return nil, err
		}
		pidx, err := oscInt(args[2], "parameter index")
		if err != nil {
			return nil, err
		}
		if len(args) == 3 {
			return []commandResult{{address: "/live/device/get/parameter/value", args: []OscValue{idx, didx, pidx}, query: true, trim: 3}}, nil
		}
		v, err := oscFloat(args[3], "parameter value")
		if err != nil {
			return nil, err
		}
		return []commandResult{{address: "/live/device/set/parameter/value", args: []OscValue{idx, didx, pidx, v}}}, nil

	// ------------------------------------------------------------------ View
	case "select-track":
		return viewProp("/live/view/get/selected_track", "/live/view/set/selected_track", 1, "select-track", args)
	case "select-scene":
		return viewProp("/live/view/get/selected_scene", "/live/view/set/selected_scene", 1, "select-scene", args)
	case "select-clip":
		return viewProp("/live/view/get/selected_clip", "/live/view/set/selected_clip", 2, "select-clip", args)
	case "select-device":
		return viewProp("/live/view/get/selected_device", "/live/view/set/selected_device", 2, "select-device", args)

	// ---------------------------------------------------------------- MidiMap
	case "map-cc":
		if len(args) != 5 {
			return nil, fmt.Errorf("map-cc requires <track> <device> <param> <channel> <cc>")
		}
		var vals []OscValue
		for i, a := range args {
			v, err := oscInt(a, "argument")
			if err != nil {
				return nil, fmt.Errorf("map-cc argument %d: %v", i+1, err)
			}
			vals = append(vals, v)
		}
		return []commandResult{{address: "/live/midimap/map_cc", args: vals}}, nil

	// ----------------------------------------------------------- Application
	case "version":
		return []commandResult{{address: "/live/application/get/version", query: true}}, nil
	case "cpu":
		return []commandResult{{address: "/live/application/get/average_process_usage", query: true}}, nil

	// ------------------------------------------------------------------- Raw
	case "raw":
		if len(args) == 0 {
			return nil, fmt.Errorf("raw requires an OSC address")
		}
		addr := args[0]
		vals := make([]OscValue, 0, len(args)-1)
		for _, a := range args[1:] {
			vals = append(vals, inferValue(a))
		}
		return []commandResult{{address: addr, args: vals, query: strings.Contains(addr, "/get/")}}, nil

	default:
		return nil, fmt.Errorf("unknown command %q (run with no arguments for help)", cmd)
	}
}

// --- command builder helpers ------------------------------------------------

// trackIndexed parses a required track index and returns a base result.
func trackIndexed(args []string, name string) (commandResult, error) {
	if len(args) == 0 {
		return commandResult{}, fmt.Errorf("%s requires a track index", name)
	}
	idx, err := oscInt(args[0], "track index")
	if err != nil {
		return commandResult{}, err
	}
	return commandResult{args: []OscValue{idx}, trim: 1}, nil
}

func trackFloatProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := trackIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscFloat(args[1], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

func trackBoolProp(getAddr, setAddr string, args []string) ([]commandResult, error) {
	r, err := trackIndexed(args, "track")
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	b, err := oscBoolWord(args[1])
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, b)
	return []commandResult{r}, nil
}

func trackStrProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := trackIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	r.address = setAddr
	r.args = append(r.args, oscStr(args[1]))
	return []commandResult{r}, nil
}

func trackIntProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := trackIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscInt(args[1], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

func trackListProp(getAddr string, args []string) ([]commandResult, error) {
	r, err := trackIndexed(args, "list")
	if err != nil {
		return nil, err
	}
	r.address = getAddr
	r.query = true
	return []commandResult{r}, nil
}

// clipSlotIndexed parses track+clip indices and returns a base result.
func clipSlotIndexed(args []string, name string) (commandResult, error) {
	if len(args) < 2 {
		return commandResult{}, fmt.Errorf("%s requires <track> <clip>", name)
	}
	t, err := oscInt(args[0], "track index")
	if err != nil {
		return commandResult{}, err
	}
	c, err := oscInt(args[1], "clip index")
	if err != nil {
		return commandResult{}, err
	}
	return commandResult{args: []OscValue{t, c}, trim: 2}, nil
}

func clipSlotAction(addr, name string, args []string) ([]commandResult, error) {
	r, err := clipSlotIndexed(args, name)
	if err != nil {
		return nil, err
	}
	r.address = addr
	return []commandResult{r}, nil
}

func clipSlotGet(addr, name string, args []string) ([]commandResult, error) {
	r, err := clipSlotIndexed(args, name)
	if err != nil {
		return nil, err
	}
	r.address = addr
	r.query = true
	return []commandResult{r}, nil
}

// clipIndexed parses track+clip indices and returns a base result.
func clipIndexed(args []string, name string) (commandResult, error) {
	return clipSlotIndexed(args, name)
}

func clipAction(addr, name string, args []string) ([]commandResult, error) {
	r, err := clipIndexed(args, name)
	if err != nil {
		return nil, err
	}
	r.address = addr
	return []commandResult{r}, nil
}

func clipFloatProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := clipIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 2 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscFloat(args[2], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

func clipBoolProp(getAddr, setAddr string, args []string) ([]commandResult, error) {
	r, err := clipIndexed(args, "clip")
	if err != nil {
		return nil, err
	}
	if len(args) == 2 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	b, err := oscBoolWord(args[2])
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, b)
	return []commandResult{r}, nil
}

func clipStrProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := clipIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 2 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	r.address = setAddr
	r.args = append(r.args, oscStr(args[2]))
	return []commandResult{r}, nil
}

func clipIntProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := clipIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 2 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscInt(args[2], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

// sceneIndexed parses a scene index and returns a base result (trim 1).
func sceneIndexed(args []string, name string) (commandResult, error) {
	if len(args) == 0 {
		return commandResult{}, fmt.Errorf("%s requires a scene index", name)
	}
	idx, err := oscInt(args[0], "scene index")
	if err != nil {
		return commandResult{}, err
	}
	return commandResult{args: []OscValue{idx}, trim: 1}, nil
}

func sceneStrProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := sceneIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	r.address = setAddr
	r.args = append(r.args, oscStr(args[1]))
	return []commandResult{r}, nil
}

func sceneIntProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := sceneIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscInt(args[1], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

func sceneFloatProp(getAddr, setAddr, name string, args []string) ([]commandResult, error) {
	r, err := sceneIndexed(args, name)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		r.address = getAddr
		r.query = true
		return []commandResult{r}, nil
	}
	v, err := oscFloat(args[1], name)
	if err != nil {
		return nil, err
	}
	r.address = setAddr
	r.args = append(r.args, v)
	return []commandResult{r}, nil
}

// deviceProp handles device-list queries that take <track> <device>.
func deviceProp(addr, name string, args []string) ([]commandResult, error) {
	r, err := clipSlotIndexed(args, name)
	if err != nil {
		return nil, err
	}
	r.address = addr
	r.query = true
	return []commandResult{r}, nil
}

// deviceParamProp handles single device parameter queries (trim 3).
func deviceParamProp(addr, name string, args []string) ([]commandResult, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("%s requires <track> <device> <param>", name)
	}
	r, err := clipSlotIndexed(args, name)
	if err != nil {
		return nil, err
	}
	pidx, err := oscInt(args[2], "parameter index")
	if err != nil {
		return nil, err
	}
	r.args = append(r.args, pidx)
	r.address = addr
	r.query = true
	r.trim = 3
	return []commandResult{r}, nil
}

// viewProp handles view get/set with N optional args.
func viewProp(getAddr, setAddr string, n int, name string, args []string) ([]commandResult, error) {
	if len(args) == 0 {
		return []commandResult{{address: getAddr, query: true}}, nil
	}
	if len(args) != n {
		return nil, fmt.Errorf("%s expects %d argument(s), got %d", name, n, len(args))
	}
	var vals []OscValue
	for i, a := range args {
		v, err := oscInt(a, "index")
		if err != nil {
			return nil, fmt.Errorf("%s argument %d: %v", name, i+1, err)
		}
		vals = append(vals, v)
	}
	return []commandResult{{address: setAddr, args: vals}}, nil
}
