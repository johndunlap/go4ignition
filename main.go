package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var tmpl = template.New("base")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(req *http.Request) bool {
		// TODO: Possibly rethink this in production. Reject connections without a recognized domain in
		//  the Host header?
		return true
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	parseTemplates()
	InitDb()
	RunMigrations(migrations)

	go func() {
		http.HandleFunc("/favicon.ico", faviconHandler)
		http.HandleFunc("/ws", wsHandler)
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/static/", staticHandler)
		http.HandleFunc("/send-chat", sendChatHandler)

		log.Println("Starting server: http://localhost:8888")
		err := http.ListenAndServe(":8888", nil)

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	<-sigc
	log.Println("\nReceived interrupt signal, shutting down...")
	os.Exit(0)
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	// Obnoxiously, this pattern is treated as a wildcard by the HTTP server, so we have to manually return 404 from
	// this handler when the request URI isn't "/"
	if req.URL.Path != "/" {
		NotFoundHandler(res, req)
		return
	}

	SetHeaders(res, req)

	data := NewPageData("", req)

	err := tmpl.ExecuteTemplate(res, "template/index.html", data)
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	err = conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Println(err)
		return
	}

	conn.SetPongHandler(func(string) error {
		err = conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Println(err)
		}
		return nil
	})

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Received message:", string(message))

		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func requestHandler(res http.ResponseWriter, req *http.Request) {
	mux := http.NewServeMux()

	for key := range TemplateHandlers {
		if key[len(key)-1:] == "/" && key != "/" {
			continue
		}
		mux.HandleFunc(key, TemplateHandlers[key])
	}

	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/ws", wsHandler)
	mux.ServeHTTP(res, req)
}

func parseTemplates() {
	var err error

	for name, contents := range TemplateContents {
		tmpl, err = tmpl.New(name).Parse(string(contents))
		if err != nil {
			log.Fatalf("Error parsing header: %v", err)
		}
	}
}

type Person struct {
	Name string
	Age  int
}

func (p Person) Save() {
	// Add Person-specific save logic here (or just a debug print for now)
	log.Println("Saving Person:", p.Name)
}

type Inventory struct {
	Item  string
	Count int
}

func (i Inventory) Save() {
	// Add Inventory-specific save logic here
	log.Println("Saving Inventory item:", i.Item)
}

func SetHeaders(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html")
	res.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	res.Header().Set("Pragma", "no-cache")
	res.Header().Set("Expires", "0")

	// https://htmx.org/docs/#caching
	if IsHTMX(req) {
		res.Header().Set("Vary", "HX-Request")
	}
}

func IsHTMX(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
