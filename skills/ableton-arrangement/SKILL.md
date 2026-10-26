---
name: ableton-arrangement
description: Create and edit Ableton Live Arrangement clips, load browser plugins, and play a section in a loop using this repository's abletoncli and custom AbletonOSC.
---

# Ableton Arrangement workflow

Use this skill when asked to compose or edit a song in the Arrangement View using
the local `abletoncli` project. It assumes Ableton Live is open and AbletonOSC is
selected as a Control Surface.

## Important behavior

- Track and Arrangement-clip indices are zero-based. Re-query `arr-clips` after
  creating or deleting clips; Arrangement indices are their current order on a
  track and may change.
- OSC queries bind response port 11001. Run queries serially, not in parallel.
- Live 11.1 lacks `Track.create_midi_clip`. This AbletonOSC branch implements
  `/live/track/create_midi_clip` by creating a temporary clip in the first empty
  Session slot, duplicating it into Arrangement, and deleting the temporary clip.
  Verify the target track has an empty Session slot before relying on this fallback.
- `arr-midi` arguments are track, start beat, and **end beat**. Note positions are
  relative to their Arrangement clip. Audio import via `arr-audio` requires the
  direct API available in Live 12.0.5+.
- Changes affect the open set but do not save it. Leave saving to the user unless
  explicitly asked.

## Inspect before editing

Build the CLI if needed:

```powershell
go build -o abletoncli.exe .
```

Inspect tracks and identify an empty MIDI track and free timeline range:

```powershell
.\abletoncli.exe tracks
.\abletoncli.exe raw /live/track/get/has_midi_input <track>
.\abletoncli.exe arr-clips <track>
.\abletoncli.exe slot-has-clip <track> 0
```

Do not overwrite existing clips or instruments. If the requested track/range is
occupied, inspect alternatives or ask the user before replacing content.

## Create a chord clip

For example, a four-bar C–Am–F–G progression in a 16-beat clip:

```powershell
.\abletoncli.exe arr-midi <track> 0 16
.\abletoncli.exe arr-clips <track>
.\abletoncli.exe arr-clip-name <track> <arrangement-index> "C - Am - F - G"
.\abletoncli.exe arr-clip-add-notes <track> <arrangement-index> `
  60 0 4 100 off 64 0 4 100 off 67 0 4 100 off `
  57 4 4 100 off 60 4 4 100 off 64 4 4 100 off `
  53 8 4 100 off 57 8 4 100 off 60 8 4 100 off `
  55 12 4 100 off 59 12 4 100 off 62 12 4 100 off
```

These note groups are `(pitch start duration velocity mute)`, with MIDI pitches
C4/E4/G4, A3/C4/E4, F3/A3/C4, and G3/B3/D4. Verify the result:

```powershell
.\abletoncli.exe arr-clips <track>
.\abletoncli.exe arr-clip-notes <track> <arrangement-index>
```

## Add a Drum Rack boom-bap track

Create a separate MIDI track rather than changing an existing instrument track.
Search the `drums` Browser category for a complete kit and verify its device name
after loading. A full rack preset is safer than loading a raw sample onto the
selected track: in Live 11, `Browser.load_item` can insert a sample as a regular
Simpler instead of placing it on the selected Drum Rack pad.

```powershell
.\abletoncli.exe create-track midi -1
.\abletoncli.exe tracks
.\abletoncli.exe raw /live/browser/get/matches drums "808 Core Kit.adg"
.\abletoncli.exe raw /live/browser/load_item <new-track> drums "808 Core Kit.adg"
.\abletoncli.exe devices <new-track>
.\abletoncli.exe arr-midi <new-track> 0 16
.\abletoncli.exe arr-clips <new-track>
```

Loading a kit can take longer than the CLI's query timeout; issue it as an
action, then verify with `devices` rather than assuming a timeout means failure.
With the standard Drum Rack mapping, C1/MIDI 36 is kick, D1/38 is snare, and
F#1/42 is closed hi-hat. A four-bar boom-bap pattern can use kicks around beats
`0, 2.5, 4, 6.5, 8, 10.5, 12, 14.5`, snares on beats `1, 3, 5, 7, 9, 11, 13, 15`,
and 8th-note hats. Vary velocities, keep note durations short, and verify via
`arr-clip-notes` before playback.

## Add a bass line

Add another MIDI track and load a bass preset from the `instruments` category.
For a warm, sampled tone, `Upright Bass.adv` is available in the standard Browser:

```powershell
.\abletoncli.exe create-track midi -1
.\abletoncli.exe tracks
.\abletoncli.exe raw /live/browser/get/matches instruments "Upright Bass.adv"
.\abletoncli.exe raw /live/browser/load_item <new-track> instruments "Upright Bass.adv"
.\abletoncli.exe devices <new-track>
.\abletoncli.exe arr-midi <new-track> 0 16
.\abletoncli.exe arr-clips <new-track>
```

Loading a preset may take longer than the CLI query timeout; use it as an action
and confirm the device with `devices`. For C–Am–F–G, a simple 4-bar line can use
C2/G1, A1/E2, F1/C2, and G1/D2 on the corresponding bars, with short pickup
notes before chord changes. MIDI pitches are C2=36, G1=31, A1=33, E2=40, F1=29,
and D2=38. Add the notes to the new clip with `arr-clip-add-notes`, then verify
with `arr-clip-notes`.

## Load an installed plugin

The custom AbletonOSC Browser endpoints search loadable items in Live's Plugins
Browser and only load a unique match. Search first; inspect the returned item
name/URI and do not guess if there are ambiguous matches:

```powershell
.\abletoncli.exe raw /live/browser/get/plugin_matches "Wurli"
.\abletoncli.exe --wait raw /live/browser/load_plugin <track> "Wurli V2"
.\abletoncli.exe devices <track>
```

The load endpoint temporarily selects the target track, loads the item, then
restores the prior track selection. It does not remove existing devices.

## Start playback inside the loop

Set the loop to the clip's exact timeline range, enable it, then start playback.
In Live 11.1, `play`/`start_playing` can start from the existing insert marker
even if `current_song_time` was moved while stopped. To ensure playback enters
the loop, seek to the loop start **while transport is running**, then verify:

```powershell
.\abletoncli.exe raw /live/song/set/loop_start 0.0
.\abletoncli.exe raw /live/song/set/loop_length 16.0
.\abletoncli.exe loop on
.\abletoncli.exe play
.\abletoncli.exe raw /live/song/set/current_song_time 0.0
.\abletoncli.exe raw /live/song/get/current_song_time
.\abletoncli.exe raw /live/song/get/is_playing
```

Confirm `is_playing` is true and the reported position is within the loop range.
If transport is stopped or position is outside the loop, correct it rather than
claiming the loop is playing.

## Remote Script setup

Arrangement editing and plugin loading require the modified Remote Script from
`./abletonosc`, installed as `Documents\Ableton\User Library\Remote Scripts\AbletonOSC`.
After changing its files, `/live/api/reload` hot-reloads supported modules; if a
new route remains unknown, inspect the AbletonOSC log before restarting Live.
