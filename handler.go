package main

import (
	"log"
	"net/http"
	"time"
)

type PageData[V any] struct {
	FileResolver  func(string) string
	Title         string
	IncludeHeader bool
	FormFields    map[string]ValidationResult
	Data          V
}

func NewPageData[V any](value V, req *http.Request) PageData[V] {
	return PageData[V]{
		Title:         appName,
		IncludeHeader: IsHTMX(req),
		FileResolver: func(name string) string {
			return StaticFileNames[name]
		},
		FormFields: make(map[string]ValidationResult),
		Data:       value,
	}
}

type ValidationResult struct {
	Value   string
	Errors  []string
	Name    string
	IsValid bool
}

func IndexHandler(res http.ResponseWriter, req *http.Request) {
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

func faviconHandler(res http.ResponseWriter, req *http.Request) {
	// Write the correct http content-type header to the response
	res.Header().Set("Content-Type", "image/x-icon")

	// We need to be able to change this, but we still need to cache it
	res.Header().Set("Cache-Control", "public, max-age=3600")
	res.Header().Set("Expires", time.Now().Add(3600*time.Second).Format(http.TimeFormat))

	// Write the content of the requested file to the response
	count, err := res.Write(FaviconICO)

	if err != nil || count == 0 {
		NotFoundHandler(res, req)
	}
}

func staticHandler(res http.ResponseWriter, req *http.Request) {
	if len(StaticFiles[req.RequestURI]) > 0 {
		// Write the correct http content-type header to the response
		res.Header().Set("Content-Type", StaticFilesContentType[req.RequestURI])

		// Cache static files for a year. This is fine from a cache busting perspective because the filenames
		// contain the md5 hash of the file
		res.Header().Set("Cache-Control", "public, max-age=31536000")
		res.Header().Set("Expires", time.Now().Add(365*24*time.Hour).Format(http.TimeFormat))

		// Write the content of the requested file to the response
		count, err := res.Write(StaticFiles[req.RequestURI])

		if err != nil || count == 0 {
			NotFoundHandler(res, req)
		}
	} else {
		NotFoundHandler(res, req)
	}
}

func SendChatHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	if req.Method != http.MethodPost {
		NotFoundHandler(res, req)
		return
	}

	err := tmpl.ExecuteTemplate(res, "template/fragment/message.html", struct {
		Message string
		Time    string
		Class   string
	}{
		Message: req.FormValue("message"),
		Time:    "now",
		Class:   "me",
	})
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
}

// NotFoundHandler HTTP handler for URI /not_found
func NotFoundHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)
	res.WriteHeader(http.StatusNotFound)

	data := NewPageData("", req)
	data.IncludeHeader = IsHTMX(req)

	err := tmpl.ExecuteTemplate(res, "template/not_found.html", data)
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
}

// RegisterHandler HTTP handler for URI /register
func RegisterHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	data := NewPageData("", req)
	data.IncludeHeader = IsHTMX(req)

	err := tmpl.ExecuteTemplate(res, "template/register.html", data)
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}
}

// LoginHandler HTTP handler for URI /login
func LoginHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	data := NewPageData("", req)
	data.IncludeHeader = IsHTMX(req)

	err := tmpl.ExecuteTemplate(res, "template/login.html", data)
	if err != nil {
		println("ERROR: " + err.Error())
	}
}

// ForgotHandler HTTP handler for URI /forgot
func ForgotHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	data := NewPageData("", req)
	data.IncludeHeader = IsHTMX(req)

	err := tmpl.ExecuteTemplate(res, "template/forgot.html", data)
	if err != nil {
		println("ERROR: " + err.Error())
	}
}

// ChatHandler HTTP handler for URI /chat
func ChatHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	data := NewPageData("", req)
	data.IncludeHeader = IsHTMX(req)

	err := tmpl.ExecuteTemplate(res, "template/chat.html", data)
	if err != nil {
		println("ERROR: " + err.Error())
	}
}
