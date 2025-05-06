/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_AUTH_SERVICE_URL: string
    readonly VITE_INITIATOR_SERVICE_URL: string
    readonly VITE_SCORE_SERVICE_URL: string
    readonly VITE_WEBTRANSPORT_GAME_URL: string
    readonly VITE_WEBTRANSPORT_MUSIC_URL: string
    readonly VITE_DEBUG: boolean
}
