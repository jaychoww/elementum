import asyncio
import json
from jsonrpcserver import method, Result, Success, dispatch

settings = {
    "download_path": "download/",
    "library_path": "library/",
    "torrents_path": "torrents/",
    "download_storage": 0, # 0: file 1: memory
    "skip_burst_search": 1,
    "auto_memory_size": 1,
    'auto_adjust_memory_size': 1,
'auto_memory_size_strategy': 0,
'memory_size': 0,
'auto_kodi_buffer_size': 1,
'auto_adjust_buffer_size': 1,
'min_candidate_size': 0,
'min_candidate_show_size': 0,
'buffer_timeout': 5,
'buffer_size': 10,
'end_buffer_size': 10,
'max_upload_rate': 0,
'max_download_rate': 0,
'autoload_torrents': 1,
'autoload_torrents_paused': False,
'spoof_user_agent': False,
'limit_after_buffering': False,
'download_file_strategy': 2,  # DownloadFileAll
'keep_downloading': 1,
'keep_files_playing': 1,
'keep_files_finished': 1,
'use_torrent_history': 1,
'torrent_history_size': 10,
'use_fanart_tv': False,
'disable_bg_progress': 0,
'disable_bg_progress_playback': 0,
'force_use_trakt': False,
'use_cache_selection': 1,
'use_cache_search': 1,
'use_cache_torrents': 1,
'cache_search_duration': 10,
'results_per_page': 10,
'show_files_watched': 1,
'greeting_enabled': 1,
'enable_overlay_status': 1,
'silent_stream_start': False,
'autoyes_enabled': False,
'autoyes_timeout': 5,
'choose_stream_auto_movie': False,
'choose_stream_auto_show': False,
'choose_stream_auto_search': False,
'force_link_type': False,
'use_original_title': 1,
'use_anime_en_title': False,
'use_lowest_release_date': False,
'add_specials': False,
'add_episode_numbers': False, 'unaired_seasons': False,
'unaired_episodes': False,
'show_episodes_on_release_day': False,
'show_unwatched_episodes_number': False,
'seasons_all': False,
'seasons_order': 0, 
'seasons_specials': False,
'playback_percent': 1,
'smart_episode_start': False,
'smart_episode_match': False,
'smart_episode_choose': False,
'library_enabled': False,
'library_sync_enabled': False,
'library_sync_playback_enabled': False,
'library_update': 0,
'strm_language': False,
'library_nfo_movies': False,
'library_nfo_shows': False,
'seed_forever': 1,
'share_ratio_limit': 0,
'seed_time_ratio_limit': 0,
'seed_time_limit': 0,
'disable_upload': False,
'disable_lsd': False,
'disable_dht': False,
'disable_tcp': False,
'disable_utp': False,
'disable_upnp': False,
'encryption_policy': 0,
'listen_port_min': 61000,
'listen_port_max': 62000,
'listen_interfaces': "",
'listen_autodetect_ip': 1,
'listen_autodetect_port': 1,
'outgoing_interfaces': "",
'tuned_storage': False,
'disk_cache_size': 0,
'use_libtorrent_config': 1,
'use_libtorrent_logging': 1,
'use_libtorrent_deadline': False,
'use_libtorrent_pauseresume': False,
'libtorrent_profile': 0,
'magnet_resolve_timeout': 5,
'add_extra_trackers': 1,
'remove_original_trackers': False,
'modify_trackers_strategy': 1,
'connections_limit': 0,
'conntracker_limit': 0,
'conntracker_limit_auto': 0,
# 'session_save'
# 'trakt_scrobble'
# 'autoscrape_is_enabled'
# 'autoscrape_library_enabled'
# 'autoscrape_strategy'
# 'autoscrape_strategy_expect'
# 'autoscrape_per_hours'
# 'autoscrape_limit_movies'
# 'autoscrape_interval'
# 'trakt_username'
# 'trakt_token'
# 'trakt_refresh_token'
# 'trakt_token_expiry'
# 'trakt_sync_enabled'
# 'trakt_sync_playback_enabled'
# 'trakt_sync_frequency_min'
# 'trakt_sync_collections'
# 'trakt_sync_watchlist'
# 'trakt_sync_userlists'
# 'trakt_sync_playback_progress'
# 'trakt_sync_hidden'
# 'trakt_sync_watched'
# 'trakt_sync_watchedback'
# 'trakt_sync_added_movies'
# 'trakt_sync_added_movies_location'
# 'trakt_sync_added_movies_list'
# 'trakt_sync_added_shows'
# 'trakt_sync_added_shows_location'
# 'trakt_sync_added_shows_list'
# 'trakt_sync_removed_movies'
# 'trakt_sync_removed_movies_location'
# 'trakt_sync_removed_movies_list'
# 'trakt_sync_removed_shows'
# 'trakt_sync_removed_shows_location'
# 'trakt_sync_removed_shows_list'
# 'trakt_progress_unaired'
# 'trakt_progress_sort'
# 'trakt_progress_date_format'
# 'trakt_progress_color_date'
# 'trakt_progress_color_show'
# 'trakt_progress_color_episode'
# 'trakt_progress_color_unaired'
# 'trakt_calendars_date_format'
# 'trakt_calendars_color_date'
# 'trakt_calendars_color_show'
# 'trakt_calendars_color_episode'
# 'trakt_calendars_color_unaired'
# 'library_update_frequency'
# 'library_update_delay'
# 'library_auto_scan'
# 'play_resume_action'
# 'play_resume_back'
# 'tmdb_api_key'
# 'tmdb_show_use_prod_company_as_studio'
# 'osdb_user'
# 'osdb_pass'
# 'osdb_language'
# 'osdb_auto_language'
# 'osdb_auto_load'
# 'osdb_auto_load_count'
# 'osdb_auto_load_delete'
# 'osdb_auto_load_skipexists'
# 'osdb_included_enabled'
# 'osdb_included_skipexists'
# 'sorting_mode_movies'
# 'sorting_mode_shows'
# 'resolution_preference_movies'
# 'resolution_preference_shows'
# 'percentage_additional_seeders'
# 'custom_provider_timeout_enabled'
# 'custom_provider_timeout'
# 'internal_dns_enabled'
# 'internal_dns_skip_ipv6'
# 'internal_proxy_enabled'
# 'internal_proxy_logging'
# 'internal_proxy_logging_body'
# 'proxy_type'
# 'proxy_enabled'
# 'proxy_host'
# 'proxy_port'
# 'proxy_login'
# 'proxy_password'
# 'use_proxy_http'
# 'use_proxy_tracker'
# 'use_proxy_download'
# 'completed_move'
# 'completed_movies_path'
# 'completed_shows_path'
# 'local_only_client'
'log_level': 5,  # debug
# 'internal_dns_opennic'
    


}


settings_list = [{"key": k, "type":type(v).__name__ , "value":str(v)} for k,v in settings.items()]
# print(settings_list)

@method
def GetAddonInfo() -> Result:
    # print("call GetAddonInfo")
    res = {
        "id": "jay",
        "path": "/tmp",
        "home": "home",
        "Profile": "profile"
        
    }
    return Success(res)

@method
def TranslatePath(path) -> Result:
    print(f"TranslatePath:{path}")
    if path.startswith("special://"):
        path = path[9:]
    return Success("/tmp/" + path)

@method
def GetPlatform() -> Result:
    return Success(
        {
            "OS": "linux",
            "Arch": "x86_64"
        }
    )


@method
def GetAllSettings() -> Result:
    return Success(settings_list)

@method
def GetSetting(id) -> Result:
    print(f"GetSetting: {id}")
    return Success(str(settings[id]))

@method
def AddonSettings(addonID) -> Result:
    return Success("")

@method
def AddonSettingsOpened(addonID) -> Result:
    return Success(True)


@method
def GetLanguage(format, withRegion):
    print(f"Get Language: {format}, {withRegion}" )
    return Success("English-UK")

    



async def handle_echo(reader, writer):
    
    data = await reader.read(1000)

    # print(f"raw data req:{data}")
    message = json.loads(data.decode())
    # print(f"message:{message}")
    res = bytes(dispatch(data), encoding='utf-8')
    # print(f"Send: {res}")
    writer.write(res)
    await writer.drain()

    # print("Close the connection")
    writer.close()


async def main():
    server = await asyncio.start_server(
        handle_echo, '127.0.0.1', 65221)

    addr = server.sockets[0].getsockname()
    print(f'Serving on {addr}')

    async with server:
        await server.serve_forever()

asyncio.run(main())
