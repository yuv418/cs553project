import './style.css';
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from './gen/auth/auth_pb';
import { createClient } from "@connectrpc/connect";
import { jwtDecode } from "jwt-decode";

interface JWTPayload {
    sub?: string;
    username?: string;
    exp: number;
}

const transport = createConnectTransport({
    baseUrl: import.meta.env.VITE_AUTH_SERVICE_URL,
    useBinaryFormat: true,
});

const client = createClient(AuthService, transport);

document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const username = (document.getElementById('username') as HTMLInputElement).value;
    const password = (document.getElementById('password') as HTMLInputElement).value;
    const errorDiv = document.getElementById('error-message');
    
    try {
        const response = await client.authenticate({
            username,
            password
        });
          if (response.jwtToken) {
            // Store the JWT token
            localStorage.setItem('auth_token', response.jwtToken);
            
            // Decode the JWT token
            const decoded = jwtDecode<JWTPayload>(response.jwtToken);
            const username = decoded.sub || decoded.username;
            
            // Hide the login form and show welcome message
            const loginContainer = document.querySelector('.login-container');
            if (loginContainer) {
                loginContainer.innerHTML = `
                    <div class="welcome-message">
                        <h2>Welcome, ${username}!</h2>
                        <p>You have successfully logged in.</p>
                    </div>
                `;
            }
        } else {
            if (errorDiv) errorDiv.textContent = 'Authentication failed';
        }
    } catch (error) {
        console.error('Authentication error:', error);
        if (errorDiv) errorDiv.textContent = 'Authentication failed. Please try again.';
    }
});

const connectToWebTransport = async () => {
    try {
        const url = import.meta.env.VITE_WEBTRANSPORT_URL;
        const transport = new WebTransport(url);

        await transport.ready;
        console.log('WebTransport connection established');

        const stream = await transport.createBidirectionalStream();
        const writer = stream.writable.getWriter();
        const reader = stream.readable.getReader();

        // Send a message to the server
        const encoder = new TextEncoder();
        const message = 'Hello, server!';
        await writer.write(encoder.encode(message));
        console.log(`Sent: ${message}`);

        // Read a message from the server
        const { value, done } = await reader.read();
        if (!done) {
            const decoder = new TextDecoder();
            console.log(`Received: ${decoder.decode(value)}`);
        }

        // Close the stream
        writer.close();
        console.log('Stream closed');

        // Close the WebTransport session
        transport.close();
        console.log('WebTransport session closed');
    } catch (error) {
        console.error('Error with WebTransport:', error);
    }
};