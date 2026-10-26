from . import client, wait_one_tick
import math


@pytest.fixture
def arrangement_midi_clip(client):
    track_id = 0  # The integration-test set's first track is MIDI.
    song_length = client.query("/live/song/get/song_length")[0]
    start_time = float(math.ceil((song_length + 8) / 8) * 8)
    end_time = start_time + 4
    client.send_message("/live/track/create_midi_clip", (track_id, start_time, end_time))
    wait_one_tick()

    starts = client.query("/live/track/get/arrangement_clips/start_time", (track_id,))[1:]
    clip_index = next(i for i, value in enumerate(starts) if abs(value - start_time) < 0.001)
    try:
        yield track_id, clip_index
    finally:
        client.send_message("/live/arrangement_clip/delete", (track_id, clip_index))
        wait_one_tick()


def test_arrangement_midi_clip_notes_and_name(client, arrangement_midi_clip):
    track_id, clip_index = arrangement_midi_clip
    client.send_message("/live/arrangement_clip/set/name", (track_id, clip_index, "OSC Arrangement Test"))
    client.send_message(
        "/live/arrangement_clip/add/notes",
        (track_id, clip_index, 60, 0.0, 1.0, 100, False),
    )
    wait_one_tick()

    assert client.query("/live/arrangement_clip/get/name", (track_id, clip_index)) == (
        track_id, clip_index, "OSC Arrangement Test"
    )
    assert client.query("/live/arrangement_clip/get/notes", (track_id, clip_index)) == (
        track_id, clip_index, 60, 0.0, 1.0, 100, False
    )

    client.send_message("/live/arrangement_clip/remove/notes", (track_id, clip_index))
    wait_one_tick()
    assert client.query("/live/arrangement_clip/get/notes", (track_id, clip_index)) == (
        track_id, clip_index
    )
