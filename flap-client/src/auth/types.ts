export interface JWTPayload {
    sub?: string;
    username?: string;
    exp: number;
}

declare global {
    interface Window {
        authLatency: number;
        initiatorLatency: number;
        gameOverScreenShown: boolean;
        firstFrameReceived: boolean;
        gameId: string;
    }
}
