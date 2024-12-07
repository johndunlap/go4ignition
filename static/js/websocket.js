const Bus = class {
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
        const message = JSON.parse(event.data);
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
};

const bus = new Bus("ws://" + window.location.hostname + ":" + window.location.port + "/ws");
