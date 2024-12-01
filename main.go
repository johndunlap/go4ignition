package main

import (
	_ "embed"
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed template/index.html
var indexTpl []byte

//go:embed static/js/websocket.js
var websocketJs []byte

var TemplateContents = map[string][]byte{
	"template/index.html": indexTpl,
}

var tmpl = template.New("base")

var TemplateHandlers = map[string]func(res http.ResponseWriter, req *http.Request){
	"/": IndexHandler,
}

var StaticFiles = map[string][]byte{
	"/static/js/websocket.js": websocketJs,
}

var StaticFilesContentType = map[string]string{
	"/static/js/websocket.js": "text/javascript",
}

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	parseTemplates()

	go func() {
		http.Handle("/ws", websocket.Handler(websocketHandler))
		http.HandleFunc("/", requestHandler)

		fmt.Printf("Starting server: http://localhost:8888\n")
		err := http.ListenAndServe(":8888", nil)

		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	<-sigc
	fmt.Println("\nReceived interrupt signal, shutting down...")
	os.Exit(0)
}

func websocketHandler(ws *websocket.Conn) {
	fmt.Println("Websocket client connected")
	for {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			fmt.Println("Error receiving message:", err)
			break
		}
		fmt.Printf("Received: %s\n", msg)
		if err := websocket.Message.Send(ws, "ACK: "+msg); err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}
	fmt.Println("Websocket client disconnected")
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
	mux.Handle("/ws", websocket.Handler(websocketHandler))
	mux.ServeHTTP(res, req)
}

func staticHandler(res http.ResponseWriter, req *http.Request) {
	if len(StaticFiles[req.RequestURI]) > 0 {
		res.Header().Set("Content-Type", StaticFilesContentType[req.RequestURI])
		res.Header().Set("Cache-Control", "public, max-age=31536000")
		res.Header().Set("Expires", time.Now().Add(365*24*time.Hour).Format(http.TimeFormat))
		count, err := res.Write(StaticFiles[req.RequestURI])

		if err != nil || count == 0 {
			NotFoundHandler(res, req)
		}
	} else {
		NotFoundHandler(res, req)
	}
}

func IndexHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		NotFoundHandler(res, req)
		return
	}

	SetHeaders(res, req)

	err := tmpl.ExecuteTemplate(res, "template/index.html", nil)
	if err != nil {
		println("ERROR: " + err.Error())
	}
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

func NotFoundHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)
	res.WriteHeader(http.StatusNotFound)

	err := tmpl.ExecuteTemplate(res, "template/not_found.html", nil)
	if err != nil {
		println("ERROR: " + err.Error())
	}
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
