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
	wait         bool
}

const usageText = `abletoncli - control Ableton Live via AbletonOSC (OSC)

Usage:
  abletoncli [flags] <command> [args]

Commands:
  play              Start playback
  stop              Stop playback
  tempo [bpm]       Get the current tempo, or set it (e.g. "tempo 128")
  tracks            List track names
  clips <track>     List clip names on a track (0-based index)
  scene <index>     Fire a scene (0-based index)
  raw <addr> [args] Send a raw OSC address (e.g. raw /live/song/get/tempo)

Flags:
  --json            Output results as JSON
  --host HOST       AbletonOSC host (default 127.0.0.1)
  --port PORT       AbletonOSC listen port (default 11000)
  --response-port P Reply port to bind (default 11001)
  --timeout MS      Reply timeout in milliseconds (default 2000)
  --wait            Wait for a reply even for action commands

Examples:
  abletoncli play
  abletoncli tempo 128
  abletoncli tracks
  abletoncli clips 0
  abletoncli raw "/live/song/get/tempo"
  abletoncli --json tempo
`

type jsonResult struct {
	Address string        `json:"address"`
	Args    []interface{} `json:"args"`
}

func main() {
	opts := options{}
	timeoutMs := 2000

	fs := flag.NewFlagSet("abletoncli", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	fs.BoolVar(&opts.json, "json", false, "output results as JSON")
	fs.StringVar(&opts.host, "host", "127.0.0.1", "AbletonOSC host")
	fs.IntVar(&opts.port, "port", 11000, "AbletonOSC listen port")
	fs.IntVar(&opts.responsePort, "response-port", 11001, "reply port to bind")
	fs.IntVar(&timeoutMs, "timeout", 2000, "reply timeout in milliseconds")
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

	cmd := fs.Arg(0)
	args := fs.Args()[1:]

	res, err := resolveCommand(cmd, args)
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

	if res.query || opts.wait {
		address, vals, err := client.Query(res.address, res.args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if res.trim > 0 {
			if len(vals) > res.trim {
				vals = vals[res.trim:]
			} else {
				vals = nil
			}
		}
		printResult(opts.json, address, vals)
		return
	}

	if err := client.Send(res.address, res.args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printResult(asJSON bool, address string, vals []OscValue) {
	if asJSON {
		anyVals := make([]interface{}, len(vals))
		for i, v := range vals {
			anyVals[i] = v.toJSON()
		}
		b, err := json.Marshal(jsonResult{Address: address, Args: anyVals})
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		return
	}
	for _, v := range vals {
		fmt.Println(v.String())
	}
}
