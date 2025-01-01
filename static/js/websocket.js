/**
 * TODO: The idea here is to use the same websocket for all communication with the server.
 */
class Bus {
    constructor(url) {
        this.attempt = 0;
        this.url = url;
        this.topics = {};
        this.connect();
    }

    connect() {
        this.socket = new WebSocket(this.url);
        this.socket.onopen = () => this.onopen();
        this.socket.onclose = (event) => this.onclose(event);
        this.socket.onerror = (event) => this.onerror(event);
        this.socket.onmessage = (event) => this.onmessage(event);
    }

    onopen() {
        console.log('Connected');
        this.attempt = 0;
    }

    onclose(event) {
        console.log('Closed:', event.code, event.reason);
        let delay = Math.min(1000 * 2 ** this.attempt++, 30000);
        console.log('Reconnect attempt ' + this.attempt + ' in ' + delay + 'ms')
        setTimeout(() => this.connect(), delay);
    }

    onerror(event) {
        console.error('Error:', event);
    }

    onmessage(event) {
        const message = this.decode(new Uint8Array(event.data));
        if (message.topic && this.topics[message.topic]) {
            this.topics[message.topic].forEach(handler => handler(message.data));
        }
    }

    sub(topic, handler) {
        if (!this.topics[topic]) {
            this.topics[topic] = [];
        }

        this.topics[topic].push(handler);
    }

    encode(topicName, mimeType, payload) {
        if (typeof topicName !== 'string' || topicName.length === 0) {
            throw new Error("Topic name must be a non-empty string.");
        }
        if (typeof mimeType !== 'string' || mimeType.length === 0) {
            throw new Error("MIME type must be a non-empty string.");
        }

        const topicBytes = new TextEncoder().encode(topicName); // Encode the topic name as UTF-8
        const mimeTypeBytes = new TextEncoder().encode(mimeType); // Encode the MIME type as UTF-8
        const payloadBytes = new TextEncoder().encode(payload); // Encode the payload as UTF-8

        if (topicBytes.length > 65535) {
            throw new Error("Topic name is too long (max 65535 bytes).");
        }
        if (mimeTypeBytes.length > 65535) {
            throw new Error("MIME type is too long (max 65535 bytes).");
        }

        // Prepare the message buffer
        const messageBuffer = new Uint8Array(4 + topicBytes.length + mimeTypeBytes.length + payloadBytes.length);

        // Write the topic length (2 bytes, big-endian)
        messageBuffer[0] = (topicBytes.length >> 8) & 0xff; // High byte
        messageBuffer[1] = topicBytes.length & 0xff;        // Low byte

        // Write the topic bytes
        messageBuffer.set(topicBytes, 2);

        // Write the MIME type length (2 bytes, big-endian)
        const mimeTypeStart = 2 + topicBytes.length;
        messageBuffer[mimeTypeStart] = (mimeTypeBytes.length >> 8) & 0xff; // High byte
        messageBuffer[mimeTypeStart + 1] = mimeTypeBytes.length & 0xff;    // Low byte

        // Write the MIME type bytes
        messageBuffer.set(mimeTypeBytes, mimeTypeStart + 2);

        // Write the payload bytes
        const payloadStart = mimeTypeStart + 2 + mimeTypeBytes.length;
        messageBuffer.set(payloadBytes, payloadStart);

        return messageBuffer;
    }

    send(topicName, mimeType, payload) {
        if (this.socket.readyState !== WebSocket.OPEN) {
            throw new Error("WebSocket is not open. Cannot send message.");
        }

        const message = this.encode(topicName, mimeType, payload);
        this.socket.send(message);
    }

    decode(messageBuffer) {
        // Read the topic length (2 bytes, big-endian)
        const topicLength = (messageBuffer[0] << 8) | messageBuffer[1];
        const topicBytes = messageBuffer.slice(2, 2 + topicLength);
        const topicName = new TextDecoder().decode(topicBytes); // Decode topic name

        // Read the MIME type length (2 bytes, big-endian)
        const mimeTypeLength = (messageBuffer[2 + topicLength] << 8) | messageBuffer[2 + topicLength + 1];
        const mimeTypeBytes = messageBuffer.slice(2 + topicLength + 2, 2 + topicLength + 2 + mimeTypeLength);
        const mimeType = new TextDecoder().decode(mimeTypeBytes); // Decode MIME type

        // Read the payload bytes
        const payloadBytes = messageBuffer.slice(2 + topicLength + 2 + mimeTypeLength);
        const payload = new TextDecoder().decode(payloadBytes); // Decode payload

        return { topic: topicName, mimeType: mimeType, data: payload };
    }
}

const bus = new Bus("ws://" + window.location.hostname + ":" + window.location.port + "/ws");
