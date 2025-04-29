import './style.css';
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from './protos/auth/auth_pb';
import * as engine from './protos/game_engine/game_engine_pb.js';
import { createClient } from "@connectrpc/connect";
import { create, toBinary } from "@bufbuild/protobuf"
import { sizeDelimitedEncode, sizeDelimitedDecodeStream  } from "@bufbuild/protobuf/wire"
import { jwtDecode } from "jwt-decode";
import * as wkt  from "@bufbuild/protobuf/wkt";

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

            await connectToWebTransport(response.jwtToken)
        } else {
            if (errorDiv) errorDiv.textContent = 'Authentication failed';
        }
    } catch (error) {
        console.error('Authentication error:', error);
        if (errorDiv) errorDiv.textContent = 'Authentication failed. Please try again.';
    }
});

export const connectToWebTransport = async (jwt) => {
    try {
        const url = import.meta.env.VITE_WEBTRANSPORT_URL + "?token=" + jwt;
        const transport = new WebTransport(url);

        await transport.ready;
        console.log('WebTransport connection established');

        const stream = await transport.createBidirectionalStream();
        const writer = stream.writable.getWriter();

        // Send a message to the server
        const inputReq = create(engine.GameEngineInputReqSchema, {
            gameId: "idk",
            key: engine.Key.SPACE
        })

        const inputBin = sizeDelimitedEncode(engine.GameEngineInputReqSchema, inputReq)

        await writer.write(inputBin);
        console.log(`Sent: ${JSON.stringify(inputReq)}`);

        // Read a message from the server
        // const { value, done } = await reader.read();
        //
        for await (const msg of sizeDelimitedDecodeStream(wkt.EmptySchema, stream.readable)) {
            console.log(msg)
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
