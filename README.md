# abletoncli

Control [Ableton Live](https://www.ableton.com) from your terminal using
[AbletonOSC](https://github.com/ideoforms/AbletonOSC) over OSC.

Written in Go, **with no external dependencies**: OSC encoding and decoding
are implemented using the standard library.

## Requirements

1. Ableton Live 11 or later.
2. Install and select the **AbletonOSC** Remote Script:
   - Folder: `Documents\Ableton\User Library\Remote Scripts\AbletonOSC`
   - In Live: `Preferences > Link / Tempo / MIDI > Control Surface = AbletonOSC`.
   - Live should display *"AbletonOSC: Listening for OSC on port 11000"*.
3. Arrangement clip commands require the modified AbletonOSC source included in
   `./abletonosc`. `arr-midi` works in Live 11 using a temporary Session slot;
   `arr-audio` requires Live 12.0.5 or later. Copy the contents of that folder to
   `Documents\Ableton\User Library\Remote Scripts\AbletonOSC` and restart Live.
   Do not copy only its inner `abletonosc` package.

## Build

```powershell
go build -o abletoncli.exe .
```

## Usage

```powershell
.\abletoncli.exe [flags] <command> [args]
```

Main flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--json` | JSON output (JSON Lines when streaming) | off |
| `--host` | AbletonOSC host | `127.0.0.1` |
| `--port` | AbletonOSC listening port | `11000` |
| `--response-port` | Response port to bind | `11001` |
| `--timeout` | Response timeout (ms) | `2000` |
| `--duration` | Streaming duration (ms) | `10000` |
| `--wait` | Wait for a response, including for action commands | off |

## Commands

### Song

| Command | OSC address | Response |
|---------|---------------|-----------|
| `play` / `stop` | `/live/song/start_playing` / `stop_playing` | no |
| `undo` / `redo` | `/live/song/undo` / `redo` | no |
| `tap` | `/live/song/tap_tempo` | no |
| `tempo [bpm]` | `/live/song/get|set tempo` | yes |
| `metronome [on\|off]` | `/live/song/get|set metronome` | yes |
| `loop [on\|off]` | `/live/song/get|set loop` | yes |
| `record-mode [on\|off]` | `/live/song/get|set record_mode` | yes |
| `signature [num [den]]` | Signature numerator/denominator | yes |
| `create-track <audio\|midi\|return> [index]` | `create_*_track` | no |
| `create-scene [index]` | `/live/song/create_scene` | no |
| `cuepoints` | `/live/song/get/cue_points` | yes |
| `cue <index\|name>` | `/live/song/cue_point/jump` | no |
| `trackdata <min> <max> <prop...>` | `/live/song/get/track_data` | yes |
| `export` | `/live/song/export/structure` | no |
| `beat` | `start_listen/beat` (stream) | yes |
| `listen <addr> [args]` | `start_listen` (stream) | yes |

### Track

| Command | OSC address | Response |
|---------|---------------|-----------|
| `tracks` | `/live/song/get/track_names` | yes |
| `arr-clips <track>` | `/live/track/get/arrangement_clips/{name,start_time,length}` | yes |
| `arr-midi <track> <start> <end>` | `/live/track/create_midi_clip` | no |
| `arr-audio <track> <file> <position>` | `/live/track/create_audio_clip` | no |
| `arr-clip-delete <track> <index>` | `/live/arrangement_clip/delete` | no |
| `arr-clip-name <track> <index> [name]` | Get/set name | yes when querying |
| `arr-clip-notes <track> <index> [...]` | `/live/arrangement_clip/get/notes` | yes |
| `arr-clip-add-notes <track> <index> ...` | `/live/arrangement_clip/add/notes` | no |
| `arr-clip-remove-notes <track> <index> [...]` | `/live/arrangement_clip/remove/notes` | no |
| `volume <track> [val]` | `/live/track/get|set volume` | yes |
| `pan <track> [val]` | `/live/track/get|set panning` | yes |
| `send <track> <send> [val]` | `/live/track/get|set send` | yes |
| `mute` / `solo` / `arm <track> [on\|off]` | Get/set corresponding property | yes |
| `track-name <track> [name]` | `/live/track/get|set name` | yes |
| `track-color <track> [index]` | `/live/track/get|set color_index` | yes |
| `track-fold <track> [on\|off]` | `/live/track/get|set fold_state` | yes |
| `devices <track>` | `/live/track/get/devices/name` | yes |
| `out-type` / `out-channel <track> [name]` | Output routing | yes |
| `in-type` / `in-channel <track> [name]` | Input routing | yes |
| `out-types` / `out-channels` / `in-types` / `in-channels <track>` | Available routing options | yes |

### Clip Slot

| Command | OSC address | Response |
|---------|---------------|-----------|
| `slot-fire` / `slot-stop <track> <clip>` | `/live/clip_slot/fire` / `stop` | no |
| `slot-create <track> <clip> [length]` | `/live/clip_slot/create_clip` | no |
| `slot-delete <track> <clip>` | `/live/clip_slot/delete_clip` | no |
| `slot-has-clip` / `slot-playing <track> <clip>` | Corresponding get command | yes |

### Clip

| Command | OSC address | Response |
|---------|---------------|-----------|
| `clip-fire` / `clip-stop <track> <clip>` | `/live/clip/fire` / `stop` | no |
| `clip-name <track> <clip> [name]` | Get/set name | yes |
| `clip-gain <track> <clip> [db]` | Get/set gain | yes |
| `clip-pitch <track> <clip> [semitones]` | Get/set pitch_coarse | yes |
| `clip-muted` / `clip-loop` / `clip-warp <track> <clip> [on\|off]` | Get/set | yes |
| `clip-color <track> <clip> [index]` | Get/set color_index | yes |
| `clip-length <track> <clip>` | Get length | yes |
| `clip-notes <track> <clip> [pitchStart pitchSpan timeStart timeSpan]` | Get notes | yes |
| `clip-add-notes <track> <clip> <pitch start duration velocity mute>...` | add notes | no |
| `clip-remove-notes <track> <clip> [...]` | remove notes | no |

### Arrangement

Arrangement clip indices match the current order of `track.arrangement_clips`.
Query `arr-clips` again after creating or deleting clips before reusing an index.
Times and lengths are expressed in beats. `arr-midi` takes a start and end time
(not a duration). `arr-audio` requires the absolute path to a compatible audio
file. Note editing uses positions relative to the clip.

```powershell
.\abletoncli.exe arr-midi 0 0 16
.\abletoncli.exe arr-clips 0
.\abletoncli.exe arr-clip-name 0 0 "Intro"
.\abletoncli.exe arr-clip-add-notes 0 0 60 0 1 100 off 64 0 1 100 off
```

Arrangement editing requires both this updated CLI and the modified Remote Script
in `./abletonosc`. In Live 11, `arr-midi` uses a temporary Session slot and
deletes it after duplicating the clip into Arrangement. `arr-audio` requires Live
12.0.5 or later. `arr-midi` creates an empty clip; use `arr-clip-add-notes` to
populate it.

### Chord

| Command | Description |
|---------|-------------|
| `chord <track> <clip> <spec> [dur] [velocity]` | Creates a MIDI clip and writes chords/notes |

`<spec>` supports:

- Explicit comma-separated notes: `C4,E4,G4`
- Chord symbols: `Cmaj7`, `Am`, `G7`, `F#m7`, `Bdim`, `Csus4`…
- Progression (space-separated chords): `"C Am F G"`
- A single note: `C4`

`dur` is the length of each chord in beats (default 4). The clip length is
`number_of_chords × dur`. `velocity` ranges from 0 to 127 (default 100).

Examples:

```powershell
.\abletoncli.exe chord 0 0 Cmaj7
.\abletoncli.exe chord 0 0 "C Am F G"          # progression, 4 beats each
.\abletoncli.exe chord 0 0 "C Am F G" 1        # 1 beat per chord
.\abletoncli.exe chord 0 0 "C4,E4,G4" 2 110    # explicit notes, duration 2, velocity 110
```

Notes:

- `chord` creates a new clip in the slot. If one already exists, delete it first
  with `slot-delete`.
- A single letter (`C`) is a major chord. To specify a single note, include the
  octave explicitly (`C4`).
- `C7` is interpreted as a dominant seventh chord, not as the note C7.
- Middle C = `C4` = 60.

Supported chord qualities: major (`C`), minor (`Cm`), `6`, `m6`, `7`,
`maj7`, `m7`, `m7b5`, `dim`, `aug`, `sus2`, `sus4`.

### Scene

| Command | OSC address | Response |
|---------|---------------|-----------|
| `scene <index>` | `/live/scene/fire` | no |
| `scene-fire-selected <index>` | `/live/scene/fire_as_selected` | no |
| `scene-name <scene> [name]` | Get/set name | yes |
| `scene-color <scene> [index]` | Get/set color_index | yes |
| `scene-tempo <scene> [bpm]` | Get/set tempo | yes |
| `scene-signature <scene> [num [den]]` | Time signature | yes |

### Device

| Command | OSC address | Response |
|---------|---------------|-----------|
| `device-num-params <track> <device>` | Get num_parameters | yes |
| `device-params <track> <device>` | Get parameters/name | yes |
| `device-params-value` / `-min` / `-max <track> <device>` | Get parameters | yes |
| `device-param <track> <device> <param> [value]` | Get/set parameter/value | yes |
| `device-param-name <track> <device> <param>` | Get parameter/name | yes |

### View

| Command | OSC address | Response |
|---------|---------------|-----------|
| `select-track` / `select-scene [index]` | Get/set selection | yes |
| `select-clip` / `select-device [track] [scene\|device]` | Get/set selection | yes |

### Other Commands

| Command | OSC address | Response |
|---------|---------------|-----------|
| `map-cc <track> <device> <param> <channel> <cc>` | `/live/midimap/map_cc` | no |
| `version` | `/live/application/get/version` | yes |
| `cpu` | `/live/application/get/average_process_usage` | yes |
| `raw <addr> [args]` | The specified address | Depends on `/get/` |

## Protocol

- The CLI sends OSC messages to port **11000**, where AbletonOSC listens.
- For queries (`get`), it binds to port **11001** and waits for a response.
- When an error occurs, AbletonOSC sends a `/live/error` message. The CLI prints
  it to stderr and exits with code 1.

## Notes

- Response port **11001 can be used by only one process at a time**. Do not run
  multiple queries in parallel; this is an AbletonOSC limitation, not a CLI one.
- `export` writes the song structure JSON to the system temporary directory.
- `beat` and `listen` bind to the response port for `--duration` ms and print
  incoming messages as they arrive.
