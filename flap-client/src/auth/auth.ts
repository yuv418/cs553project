import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthResponse, AuthService } from '../protos/auth/auth_pb';
import { InitiatorService } from '../protos/initiator/initiator_pb';
import { createClient } from "@connectrpc/connect";
import { jwtDecode } from "jwt-decode";
import { JWTPayload } from './types';

export const authTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_AUTH_SERVICE_URL,
    useBinaryFormat: true,
});

export const initiatorTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_INITIATOR_SERVICE_URL,
    useBinaryFormat: true,
});

export const authClient = createClient(AuthService, authTransport);
export const initiatorClient = createClient(InitiatorService, initiatorTransport);

export async function handleLogin(username: string, password: string): Promise<AuthResponse | null> {
    try {
        let now = performance.now()
        const response = await authClient.authenticate({
            username,
            password
        });
        let elapsed = performance.now()

        window.authLatency = elapsed - now

        if (response.jwtToken) {
            localStorage.setItem('auth_token', response.jwtToken);
            return response;
        }
        return null;
    } catch (error) {
        console.error('Authentication error:', error);
        return null;
    }
}

export function decodeToken(token: string): { username: string; } {
    const decoded = jwtDecode<JWTPayload>(token);
    return {
        username: decoded.sub || decoded.username || 'Player'
    };
}

export function isLoggedIn(): boolean {
    const token = localStorage.getItem('auth_token');

    // Check if the token is valid
    if (token) {
        const decoded = jwtDecode<JWTPayload>(token);
        const currentTime = Math.floor(Date.now() / 1000);
        return decoded.exp ? decoded.exp > currentTime : false;
    }

    // If no token is found, the user is not logged in
    return false;
}
