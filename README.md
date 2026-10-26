# abletoncli

Control [Ableton Live](https://www.ableton.com) desde la terminal usando
[AbletonOSC](https://github.com/ideoforms/AbletonOSC) por OSC.

Escrito en Go, **sin dependencias externas**: el encode/decode de OSC está
implementado a mano sobre la librería estándar.

## Requisitos

1. Ableton Live 11 o superior.
2. El Remote Script **AbletonOSC** instalado y seleccionado:
   - Carpeta: `Documents\Ableton\User Library\Remote Scripts\AbletonOSC`
   - En Live: `Preferences > Link / Tempo / MIDI > Control Surface = AbletonOSC`.
   - Debe aparecer el mensaje *"AbletonOSC: Listening for OSC on port 11000"*.
3. Para crear clips en Arrangement, instala el AbletonOSC modificado de
   `./abletonosc` (rama `feature/arrangement-clips`). `arr-midi` funciona desde
   Live 11 usando un slot temporal; `arr-audio` requiere Live 12.0.5 o superior.
   Copia el contenido de esa carpeta a `Documents\Ableton\User Library\Remote Scripts\AbletonOSC`
   y reinicia Live; no copies solo su subcarpeta `abletonosc`.

## Compilar

```powershell
go build -o abletoncli.exe .
```

## Uso

```powershell
.\abletoncli.exe [flags] <comando> [args]
```

Flags principales:

| Flag | Descripción | Default |
|------|-------------|---------|
| `--json` | Salida JSON (JSON-lines al hacer streaming) | off |
| `--host` | Host de AbletonOSC | `127.0.0.1` |
| `--port` | Puerto de escucha de AbletonOSC | `11000` |
| `--response-port` | Puerto de respuesta a enlazar | `11001` |
| `--timeout` | Timeout de respuesta (ms) | `2000` |
| `--duration` | Duración de streaming (ms) | `10000` |
| `--wait` | Esperar respuesta incluso en comandos de acción | off |

## Comandos

### Song

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `play` / `stop` | `/live/song/start_playing` / `stop_playing` | no |
| `undo` / `redo` | `/live/song/undo` / `redo` | no |
| `tap` | `/live/song/tap_tempo` | no |
| `tempo [bpm]` | `/live/song/get|set tempo` | sí |
| `metronome [on\|off]` | `/live/song/get|set metronome` | sí |
| `loop [on\|off]` | `/live/song/get|set loop` | sí |
| `record-mode [on\|off]` | `/live/song/get|set record_mode` | sí |
| `signature [num [den]]` | signature numerator/denominator | sí |
| `create-track <audio\|midi\|return> [index]` | `create_*_track` | no |
| `create-scene [index]` | `/live/song/create_scene` | no |
| `cuepoints` | `/live/song/get/cue_points` | sí |
| `cue <index\|name>` | `/live/song/cue_point/jump` | no |
| `trackdata <min> <max> <prop...>` | `/live/song/get/track_data` | sí |
| `export` | `/live/song/export/structure` | no |
| `beat` | `start_listen/beat` (stream) | sí |
| `listen <addr> [args]` | `start_listen` (stream) | sí |

### Track

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `tracks` | `/live/song/get/track_names` | sí |
| `arr-clips <track>` | `/live/track/get/arrangement_clips/{name,start_time,length}` | sí |
| `arr-midi <track> <start> <end>` | `/live/track/create_midi_clip` | no |
| `arr-audio <track> <file> <position>` | `/live/track/create_audio_clip` | no |
| `arr-clip-delete <track> <index>` | `/live/arrangement_clip/delete` | no |
| `arr-clip-name <track> <index> [name]` | get/set name | sí al consultar |
| `arr-clip-notes <track> <index> [...]` | `/live/arrangement_clip/get/notes` | sí |
| `arr-clip-add-notes <track> <index> ...` | `/live/arrangement_clip/add/notes` | no |
| `arr-clip-remove-notes <track> <index> [...]` | `/live/arrangement_clip/remove/notes` | no |
| `volume <track> [val]` | `/live/track/get|set volume` | sí |
| `pan <track> [val]` | `/live/track/get|set panning` | sí |
| `send <track> <send> [val]` | `/live/track/get|set send` | sí |
| `mute` / `solo` / `arm <track> [on\|off]` | get/set correspondiente | sí |
| `track-name <track> [name]` | `/live/track/get|set name` | sí |
| `track-color <track> [index]` | `/live/track/get|set color_index` | sí |
| `track-fold <track> [on\|off]` | `/live/track/get|set fold_state` | sí |
| `devices <track>` | `/live/track/get/devices/name` | sí |
| `out-type` / `out-channel <track> [name]` | output routing | sí |
| `in-type` / `in-channel <track> [name]` | input routing | sí |
| `out-types` / `out-channels` / `in-types` / `in-channels <track>` | routing disponible | sí |

### Clip Slot

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `slot-fire` / `slot-stop <track> <clip>` | `/live/clip_slot/fire` / `stop` | no |
| `slot-create <track> <clip> [length]` | `/live/clip_slot/create_clip` | no |
| `slot-delete <track> <clip>` | `/live/clip_slot/delete_clip` | no |
| `slot-has-clip` / `slot-playing <track> <clip>` | get correspondiente | sí |

### Clip

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `clip-fire` / `clip-stop <track> <clip>` | `/live/clip/fire` / `stop` | no |
| `clip-name <track> <clip> [name]` | get/set name | sí |
| `clip-gain <track> <clip> [db]` | get/set gain | sí |
| `clip-pitch <track> <clip> [semitones]` | get/set pitch_coarse | sí |
| `clip-muted` / `clip-loop` / `clip-warp <track> <clip> [on\|off]` | get/set | sí |
| `clip-color <track> <clip> [index]` | get/set color_index | sí |
| `clip-length <track> <clip>` | get length | sí |
| `clip-notes <track> <clip> [pitchStart pitchSpan timeStart timeSpan]` | get notes | sí |
| `clip-add-notes <track> <clip> <pitch start duration velocity mute>...` | add notes | no |
| `clip-remove-notes <track> <clip> [...]` | remove notes | no |

### Arrangement

Los índices de clips de Arrangement corresponden al orden actual de
`track.arrangement_clips`; vuelve a consultar `arr-clips` después de crear o
borrar clips antes de reutilizar un índice. Los tiempos y duraciones se expresan
en beats. `arr-midi` recibe inicio y fin (no duración). `arr-audio` requiere la
ruta absoluta de un archivo de audio compatible. La edición de notas usa tiempos
relativos al clip.

```powershell
.\abletoncli.exe arr-midi 0 0 16
.\abletoncli.exe arr-clips 0
.\abletoncli.exe arr-clip-name 0 0 "Intro"
.\abletoncli.exe arr-clip-add-notes 0 0 60 0 1 100 off 64 0 1 100 off
```

La interfaz Arrangement requiere tanto este CLI actualizado como el Remote
Script modificado en `./abletonosc`. En Live 11, `arr-midi` usa un slot de
Session temporal y lo elimina después de duplicarlo al Arrangement. `arr-audio`
requiere Live 12.0.5+. `arr-midi` crea un clip vacío y `arr-clip-add-notes` lo rellena.

### Chord

| Comando | Descripción |
|---------|-------------|
| `chord <track> <clip> <spec> [dur] [velocity]` | Crea un clip MIDI y escribe acordes/notas |

`<spec>` acepta:

- Notas explícitas separadas por coma: `C4,E4,G4`
- Símbolos de acorde: `Cmaj7`, `Am`, `G7`, `F#m7`, `Bdim`, `Csus4`…
- Progresión (acordes separados por espacio): `"C Am F G"`
- Una nota sola: `C4`

`dur` es la duración de cada acorde en beats (default 4). La longitud del clip
es `num_acordes × dur`. `velocity` va de 0 a 127 (default 100).

Ejemplos:

```powershell
.\abletoncli.exe chord 0 0 Cmaj7
.\abletoncli.exe chord 0 0 "C Am F G"          # progresión, 4 beats cada uno
.\abletoncli.exe chord 0 0 "C Am F G" 1        # 1 beat por acorde
.\abletoncli.exe chord 0 0 "C4,E4,G4" 2 110    # notas explícitas, dur 2, vel 110
```

Notas:

- `chord` crea un clip nuevo en el slot; si ya existía uno, bórralo antes con
  `slot-delete`.
- Una letra sola (`C`) es un acorde mayor; para una nota sola usa la octava
  explícita (`C4`).
- `C7` se interpreta como acorde de séptima dominante, no como la nota C7.
- Middle C = `C4` = 60.

Calidades de acorde soportadas: mayor (`C`), menor (`Cm`), `6`, `m6`, `7`,
`maj7`, `m7`, `m7b5`, `dim`, `aug`, `sus2`, `sus4`.

### Scene

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `scene <index>` | `/live/scene/fire` | no |
| `scene-fire-selected <index>` | `/live/scene/fire_as_selected` | no |
| `scene-name <scene> [name]` | get/set name | sí |
| `scene-color <scene> [index]` | get/set color_index | sí |
| `scene-tempo <scene> [bpm]` | get/set tempo | sí |
| `scene-signature <scene> [num [den]]` | time signature | sí |

### Device

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `device-num-params <track> <device>` | get num_parameters | sí |
| `device-params <track> <device>` | get parameters/name | sí |
| `device-params-value` / `-min` / `-max <track> <device>` | get parameters | sí |
| `device-param <track> <device> <param> [value]` | get/set parameter/value | sí |
| `device-param-name <track> <device> <param>` | get parameter/name | sí |

### View

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `select-track` / `select-scene [index]` | get/set selected | sí |
| `select-clip` / `select-device [track] [scene\|device]` | get/set selected | sí |

### Otros

| Comando | Dirección OSC | Respuesta |
|---------|---------------|-----------|
| `map-cc <track> <device> <param> <channel> <cc>` | `/live/midimap/map_cc` | no |
| `version` | `/live/application/get/version` | sí |
| `cpu` | `/live/application/get/average_process_usage` | sí |
| `raw <addr> [args]` | la dirección dada | según `/get/` |

## Protocolo

- Envía OSC al puerto **11000** (donde escucha AbletonOSC).
- Para consultas (`get`), abre el puerto **11001** y espera la respuesta.
- AbletonOSC envía mensajes `/live/error` cuando algo falla; la CLI los
  muestra como error en stderr y sale con código 1.

## Notas

- El puerto de respuesta **11001 es de un solo proceso a la vez**: no lances
  varias consultas en paralelo (es una limitación de AbletonOSC, no de la CLI).
- `export` escribe el JSON de estructura de la canción en el directorio temporal
  del sistema.
- `beat` y `listen` abren el puerto de respuesta durante `--duration` ms y
  van imprimiendo los mensajes entrantes.
