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

Esto genera `abletoncli.exe` en esta carpeta.

## Uso

```powershell
.\abletoncli.exe play
.\abletoncli.exe stop
.\abletoncli.exe tempo          # leer tempo actual
.\abletoncli.exe tempo 128      # fijar tempo a 128 BPM
.\abletoncli.exe tracks
.\abletoncli.exe clips 0        # nombres de clips del track 0
.\abletoncli.exe scene 1        # disparar la escena 1
.\abletoncli.exe raw /live/song/get/tempo
.\abletoncli.exe --json tempo   # salida JSON
```

## Protocolo

- Envía OSC al puerto **11000** (donde escucha AbletonOSC).
- Para consultas (`get`), abre el puerto **11001** y espera la respuesta.

## Comandos

| Comando        | Dirección OSC                         | Respuesta |
|----------------|---------------------------------------|-----------|
| `play`         | `/live/song/start_playing`            | no        |
| `stop`         | `/live/song/stop_playing`             | no        |
| `tempo`        | `/live/song/get/tempo`                | sí        |
| `tempo <bpm>`  | `/live/song/set/tempo`                | no        |
| `tracks`       | `/live/song/get/track_names`          | sí        |
| `clips <n>`    | `/live/track/get/clips/name`          | sí        |
| `scene <n>`    | `/live/scene/fire`                    | no        |
| `raw <addr>`   | la dirección dada                     | según `/get/` |
