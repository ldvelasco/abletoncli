import Live
from typing import Tuple
from .handler import AbletonOSCHandler

class ApplicationHandler(AbletonOSCHandler):
    def init_api(self):
        #--------------------------------------------------------------------------------
        # Generic callbacks
        #--------------------------------------------------------------------------------
        def get_version(_) -> Tuple:
            application = Live.Application.get_application()
            return application.get_major_version(), application.get_minor_version()
        self.osc_server.add_handler("/live/application/get/version", get_version)
        self.osc_server.send("/live/startup")

        def get_average_process_usage(_) -> Tuple:
            application = Live.Application.get_application()
            return application.average_process_usage,
        self.osc_server.add_handler("/live/application/get/average_process_usage", get_average_process_usage)
        self.osc_server.send("/live/application/get/average_process_usage")

        def find_browser_items(category, name):
            categories = ("audio_effects", "drums", "instruments", "midi_effects", "plugins", "samples", "sounds")
            if category not in categories:
                raise ValueError("Unsupported Browser category %r" % category)
            browser = Live.Application.get_application().browser
            pending = [getattr(browser, category)]
            matches = []
            while pending:
                item = pending.pop()
                if item.is_loadable and name.casefold() in item.name.casefold():
                    matches.append(item)
                pending.extend(item.children)
            return matches

        def load_browser_item(track_index, category, name):
            if track_index < 0 or track_index >= len(self.song.tracks):
                raise ValueError("track index out of range: %d" % track_index)

            matches = find_browser_items(category, name)
            if not matches:
                raise ValueError("No loadable item matching %r found in Browser category %r" % (name, category))
            if len(matches) > 1:
                uris = ", ".join(item.uri for item in matches)
                raise ValueError("Browser query %r is ambiguous; matching URIs: %s" % (name, uris))

            target_track = self.song.tracks[track_index]
            previous_track = self.song.view.selected_track
            try:
                self.song.view.selected_track = target_track
                Live.Application.get_application().browser.load_item(matches[0])
            finally:
                if previous_track is not None:
                    self.song.view.selected_track = previous_track
            return track_index, matches[0].name

        def get_plugin_matches(params):
            if len(params) != 1:
                raise ValueError("Expected one plugin name")
            matches = find_browser_items("plugins", str(params[0]))
            return tuple(uri for item in matches for uri in (item.name, item.uri))

        def load_plugin(params):
            if len(params) != 2:
                raise ValueError("Expected <track_id> <plugin_name>")
            track_index, plugin_name = int(params[0]), str(params[1])
            return load_browser_item(track_index, "plugins", plugin_name)

        def get_browser_matches(params):
            if len(params) != 2:
                raise ValueError("Expected <category> <name_query>")
            category, name = str(params[0]), str(params[1])
            matches = find_browser_items(category, name)
            return tuple(value for item in matches for value in (item.name, item.uri))

        def load_browser_item_handler(params):
            if len(params) != 3:
                raise ValueError("Expected <track_id> <category> <name_query>")
            return load_browser_item(int(params[0]), str(params[1]), str(params[2]))

        self.osc_server.add_handler("/live/browser/get/plugin_matches", get_plugin_matches)
        self.osc_server.add_handler("/live/browser/load_plugin", load_plugin)
        self.osc_server.add_handler("/live/browser/get/matches", get_browser_matches)
        self.osc_server.add_handler("/live/browser/load_item", load_browser_item_handler)
