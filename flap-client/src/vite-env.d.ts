/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_AUTH_SERVICE_URL: string
    readonly VITE_INITIATOR_SERVICE_URL: string
    readonly VITE_WEBTRANSPORT_URL: string
    readonly VITE_DEBUG: boolean
}
