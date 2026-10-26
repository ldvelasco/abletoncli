package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

type options struct {
	json         bool
	host         string
	port         int
	responsePort int
	timeout      time.Duration
	duration     time.Duration
	wait         bool
}

const usageText = `abletoncli - control Ableton Live via AbletonOSC (OSC)

Usage:
  abletoncli [flags] <command> [args]

Song:
  play / stop / undo / redo / tap
  tempo [bpm]            Get or set tempo
  metronome [on|off]     Get or set the metronome
  loop [on|off]          Get or set the song loop switch
  record-mode [on|off]   Get or set record mode
  signature [num [den]]  Get or set time signature
  create-track <audio|midi|return> [index]
  create-scene [index]
  cuepoints              List cue points (name, time)
  cue <index|name>       Jump to a cue point
  trackdata <min> <max> <property...>
  export                 Export song structure to a JSON file
  beat                   Stream beat messages (listener)
  listen <addr> [args]   Stream a start_listen address

Track:
  tracks                 List track names
  arr-clips <track>      List Arrangement clips (name, start and length)
  arr-midi <track> <start> <end>
                         Create a MIDI clip in Arrangement (Live 11+)
  arr-audio <track> <file> <position>
                         Add an audio file to Arrangement (Live 12.0.5+)
  arr-clip-delete <track> <index>
  arr-clip-name <track> <index> [name]
  arr-clip-notes <track> <index> [pitchStart pitchSpan timeStart timeSpan]
  arr-clip-add-notes <track> <index> <pitch start duration velocity mute>...
  arr-clip-remove-notes <track> <index> [pitchStart pitchSpan timeStart timeSpan]
  volume <track> [val]   Get or set track volume
  pan <track> [val]      Get or set track panning
  send <track> <send> [val]
  mute / solo / arm <track> [on|off]
  track-name <track> [name]
  track-color <track> [index]
  track-fold <track> [on|off]
  devices <track>        List devices on a track
  out-type / out-channel <track> [name]
  in-type / in-channel <track> [name]
  out-types / out-channels / in-types / in-channels <track>

Clip Slot:
  slot-fire / slot-stop <track> <clip>
  slot-create <track> <clip> [length]
  slot-delete <track> <clip>
  slot-has-clip / slot-playing <track> <clip>

Clip:
  clip-fire / clip-stop <track> <clip>
  clip-name <track> <clip> [name]
  clip-gain <track> <clip> [db]
  clip-pitch <track> <clip> [semitones]
  clip-muted / clip-loop / clip-warp <track> <clip> [on|off]
  clip-color <track> <clip> [index]
  clip-length <track> <clip>
  clip-notes <track> <clip> [pitchStart pitchSpan timeStart timeSpan]
  clip-add-notes <track> <clip> <pitch start duration velocity mute>...
  clip-remove-notes <track> <clip> [pitchStart pitchSpan timeStart timeSpan]

Chord:
  chord <track> <clip> <spec> [dur] [velocity]
                         Create a clip and write chords/notes
                         spec examples: "C4,E4,G4" (notes), "Cmaj7" (symbol),
                         "C Am F G" (progression, one chord per dur)

Scene:
  scene <index>          Fire a scene
  scene-fire-selected <index>
  scene-name <scene> [name]
  scene-color <scene> [index]
  scene-tempo <scene> [bpm]
  scene-signature <scene> [num [den]]

Device:
  device-num-params <track> <device>
  device-params / device-params-value / device-params-min / device-params-max <track> <device>
  device-param <track> <device> <param> [value]
  device-param-name <track> <device> <param>

View:
  select-track / select-scene [index]
  select-clip / select-device [track] [scene|device]

Other:
  map-cc <track> <device> <param> <channel> <cc>
  version                Live version (major minor)
  cpu                    Average process usage
  raw <addr> [args]      Send a raw OSC address

Flags:
  --json            Output results as JSON (JSON lines when streaming)
  --host HOST       AbletonOSC host (default 127.0.0.1)
  --port PORT       AbletonOSC listen port (default 11000)
  --response-port P Reply port to bind (default 11001)
  --timeout MS      Reply timeout in milliseconds (default 2000)
  --duration MS     Stream duration in milliseconds (default 10000)
  --wait            Wait for a reply even for action commands
`

type jsonResult struct {
	Address string        `json:"address"`
	Label   string        `json:"label,omitempty"`
	Args    []interface{} `json:"args"`
}

type queryOutput struct {
	address string
	label   string
	vals    []OscValue
}

func main() {
	opts := options{}
	timeoutMs := 2000
	durationMs := 10000

	fs := flag.NewFlagSet("abletoncli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	fs.BoolVar(&opts.json, "json", false, "output results as JSON")
	fs.StringVar(&opts.host, "host", "127.0.0.1", "AbletonOSC host")
	fs.IntVar(&opts.port, "port", 11000, "AbletonOSC listen port")
	fs.IntVar(&opts.responsePort, "response-port", 11001, "reply port to bind")
	fs.IntVar(&timeoutMs, "timeout", 2000, "reply timeout in milliseconds")
	fs.IntVar(&durationMs, "duration", 10000, "stream duration in milliseconds")
	fs.BoolVar(&opts.wait, "wait", false, "wait for a reply even for action commands")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		os.Exit(2)
	}
	if fs.NArg() == 0 {
		fmt.Fprint(os.Stderr, usageText)
		os.Exit(1)
	}
	opts.timeout = time.Duration(timeoutMs) * time.Millisecond
	opts.duration = time.Duration(durationMs) * time.Millisecond

	cmd := fs.Arg(0)
	args := fs.Args()[1:]

	results, err := resolveCommand(cmd, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	client := &Client{
		Host:         opts.host,
		Port:         opts.port,
		ResponsePort: opts.responsePort,
		Timeout:      opts.timeout,
	}

	var outputs []queryOutput
	for _, res := range results {
		if res.stream != nil {
			if err := runStream(client, res.stream, opts); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			continue
		}

		if res.query || opts.wait {
			address, vals, err := client.Query(res.address, res.args)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			if address == "/live/error" {
				msg := "unknown AbletonOSC error"
				if len(vals) > 0 {
					msg = vals[0].String()
				}
				fmt.Fprintln(os.Stderr, "error:", msg)
				os.Exit(1)
			}
			vals = trimVals(vals, res.trim)
			outputs = append(outputs, queryOutput{address: address, label: res.label, vals: vals})
			continue
		}

		if err := client.Send(res.address, res.args); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}

	if len(outputs) > 0 {
		printOutputs(opts.json, outputs)
	}
}

func trimVals(vals []OscValue, trim int) []OscValue {
	if trim <= 0 {
		return vals
	}
	if len(vals) <= trim {
		return nil
	}
	return vals[trim:]
}

func runStream(client *Client, s *streamSpec, opts options) error {
	return client.Stream(s.startAddr, s.startArgs, s.stopAddr, s.stopArgs, opts.duration,
		func(addr string, vals []OscValue) error {
			if opts.json {
				b, err := json.Marshal(jsonResult{Address: addr, Args: toJSONSlice(vals)})
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
			for _, v := range vals {
				fmt.Println(v.String())
			}
			return nil
		})
}

func toJSONSlice(vals []OscValue) []interface{} {
	out := make([]interface{}, len(vals))
	for i, v := range vals {
		out[i] = v.toJSON()
	}
	return out
}

func printOutputs(asJSON bool, outputs []queryOutput) {
	if asJSON {
		if len(outputs) == 1 {
			o := outputs[0]
			b, err := json.Marshal(jsonResult{Address: o.address, Label: o.label, Args: toJSONSlice(o.vals)})
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Println(string(b))
			return
		}
		arr := make([]jsonResult, len(outputs))
		for i, o := range outputs {
			arr[i] = jsonResult{Address: o.address, Label: o.label, Args: toJSONSlice(o.vals)}
		}
		b, err := json.Marshal(arr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		return
	}

	for _, o := range outputs {
		if o.label != "" {
			parts := make([]string, len(o.vals))
			for i, v := range o.vals {
				parts[i] = v.String()
			}
			fmt.Printf("%s: %s\n", o.label, joinWords(parts))
			continue
		}
		for _, v := range o.vals {
			fmt.Println(v.String())
		}
	}
}

func joinWords(parts []string) string {
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += " "
		}
		s += p
	}
	return s
}
