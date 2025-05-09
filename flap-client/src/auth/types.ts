export interface JWTPayload {
    sub?: string;
    username?: string;
    exp: number;
}

declare global {
    interface Window {
        authLatency: number;
        gameOverScreenShown: boolean;
    }
}
