import './style.css'

const connectToWebTransport = async () => {
    try {
        const url = 'https://localhost:4433/webtransport';
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

connectToWebTransport();