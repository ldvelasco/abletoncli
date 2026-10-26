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
