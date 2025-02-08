package _skeleton

import (
	"net/http"
	"time"
)

type PageData[V any] struct {
	FileResolver       func(string) string
	Title              string
	Description        string
	SearchKeywords     string
	OGTitle            string
	OGDescription      string
	OGImage            string
	OGURL              string
	TwitterTitle       string
	TwitterDescription string
	TwitterImage       string
	CanonicalLink      string
	IncludeHeader      bool
	Data               V
}

func NewPageData[V any](value V, req *http.Request) PageData[V] {
	return PageData[V]{
		Title:         "go4ignition",
		IncludeHeader: IsHTMX(req),
		FileResolver: func(name string) string {
			return StaticFileNames[name]
		},
		Data: value,
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
	// https://htmx.org/attributes/hx-request/
	return req.Header.Get("HX-Request") == "true"
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

// NotFoundHandler HTTP handler for resources which do not exist
func NotFoundHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)
	res.WriteHeader(http.StatusNotFound)

	err := tmpl.ExecuteTemplate(res, "template/not_found.html", nil)
	if err != nil {
		println("ERROR: " + err.Error())
	}
}

// ReadmeHandler HTTP handler for URI /readme
func ReadmeHandler(res http.ResponseWriter, req *http.Request) {
	SetHeaders(res, req)

	data := NewPageData("", req)

	err := tmpl.ExecuteTemplate(res, "template/readme.html", data)
	if err != nil {
		println("ERROR: " + err.Error())
	}
}

func StaticFileHandler(res http.ResponseWriter, req *http.Request) {
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
