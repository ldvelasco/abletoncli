from typing import Any, Tuple

import Live

from .handler import AbletonOSCHandler


class ArrangementClipHandler(AbletonOSCHandler):
    """OSC access to clips on a track's Arrangement timeline.

    Arrangement clip indices refer to the current order of ``track.arrangement_clips``.
    Re-query the track after inserting or deleting clips before reusing an index.
    """

    def __init__(self, manager):
        super().__init__(manager)
        self.class_identifier = "arrangement_clip"

    def init_api(self):
        def create_arrangement_clip_callback(func, *args):
            def callback(params: Tuple[Any]):
                if len(params) < 2:
                    raise ValueError("Arrangement clip API requires <track_id> <clip_index>")

                track_index, clip_index = int(params[0]), int(params[1])
                if track_index < 0 or track_index >= len(self.song.tracks):
                    raise ValueError("track index out of range: %d" % track_index)

                track = self.song.tracks[track_index]
                clips = track.arrangement_clips
                if clip_index < 0 or clip_index >= len(clips):
                    raise ValueError("arrangement clip index out of range: %d" % clip_index)

                clip = clips[clip_index]
                rv = func(clip, *args, tuple(params[2:]))
                if rv is not None:
                    return (track_index, clip_index, *rv)

            return callback

        properties_r = [
            "end_time",
            "is_audio_clip",
            "is_midi_clip",
            "length",
            "name",
            "start_time",
        ]
        properties_rw = [
            "color",
            "color_index",
            "gain",
            "loop_end",
            "loop_start",
            "looping",
            "muted",
            "name",
            "pitch_coarse",
            "pitch_fine",
            "warping",
        ]

        for prop in properties_r + properties_rw:
            self.osc_server.add_handler(
                "/live/arrangement_clip/get/%s" % prop,
                create_arrangement_clip_callback(self._get_property, prop),
            )

        for prop in properties_rw:
            self.osc_server.add_handler(
                "/live/arrangement_clip/set/%s" % prop,
                create_arrangement_clip_callback(self._set_property, prop),
            )

        def get_notes(clip, params: Tuple[Any] = ()):
            if len(params) == 4:
                pitch_start, pitch_span, time_start, time_span = params
            elif len(params) == 0:
                pitch_start, pitch_span, time_start, time_span = 0, 127, -8192, 16384
            else:
                raise ValueError("Expected no note range or <pitch_start pitch_span time_start time_span>")

            notes = clip.get_notes_extended(pitch_start, pitch_span, time_start, time_span)
            values = []
            for note in notes:
                values.extend((note.pitch, note.start_time, note.duration, note.velocity, note.mute))
            return tuple(values)

        def add_notes(clip, params: Tuple[Any] = ()):
            if not params or len(params) % 5:
                raise ValueError("Note data must be groups of 5: pitch start duration velocity mute")

            notes = []
            for offset in range(0, len(params), 5):
                pitch, start_time, duration, velocity, mute = params[offset:offset + 5]
                notes.append(Live.Clip.MidiNoteSpecification(
                    start_time=start_time,
                    duration=duration,
                    pitch=int(pitch),
                    velocity=int(velocity),
                    mute=bool(mute),
                ))
            clip.add_new_notes(tuple(notes))

        def remove_notes(clip, params: Tuple[Any] = ()):
            if len(params) == 4:
                pitch_start, pitch_span, time_start, time_span = params
            elif len(params) == 0:
                pitch_start, pitch_span, time_start, time_span = 0, 127, -8192, 16384
            else:
                raise ValueError("Expected no note range or <pitch_start pitch_span time_start time_span>")
            clip.remove_notes_extended(pitch_start, pitch_span, time_start, time_span)

        self.osc_server.add_handler(
            "/live/arrangement_clip/get/notes",
            create_arrangement_clip_callback(get_notes),
        )
        self.osc_server.add_handler(
            "/live/arrangement_clip/add/notes",
            create_arrangement_clip_callback(add_notes),
        )
        self.osc_server.add_handler(
            "/live/arrangement_clip/remove/notes",
            create_arrangement_clip_callback(remove_notes),
        )

        def delete_arrangement_clip(params: Tuple[Any]):
            if len(params) != 2:
                raise ValueError("Expected <track_id> <clip_index>")
            track_index, clip_index = int(params[0]), int(params[1])
            if track_index < 0 or track_index >= len(self.song.tracks):
                raise ValueError("track index out of range: %d" % track_index)
            track = self.song.tracks[track_index]
            clips = track.arrangement_clips
            if clip_index < 0 or clip_index >= len(clips):
                raise ValueError("arrangement clip index out of range: %d" % clip_index)
            track.delete_clip(clips[clip_index])

        self.osc_server.add_handler("/live/arrangement_clip/delete", delete_arrangement_clip)
