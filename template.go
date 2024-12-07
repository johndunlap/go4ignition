// Generated by gentemplate.sh
package main

import (
	_ "embed"
  "net/http"
)

//go:embed template/register.html
var md5d41d8cd98f00b204e9800998ecf8427e []byte

//go:embed template/index.html
var md5229c0df79c3c029fe8d0bbac997d8d43 []byte

//go:embed template/fragment/header.html
var md58c7becc66f7ffad96533054873aabd5e []byte

//go:embed template/fragment/typing.html
var md5632006067794e49ff3d59b7d67212756 []byte

//go:embed template/fragment/footer.html
var md5308065b5078a49f986fc3c9f9b66e5d3 []byte

//go:embed template/fragment/message.html
var md54bfb6b3bc3f44cb86b69ec72b7161e99 []byte

//go:embed template/not_found.html
var md5c3e7441b19141943f20e2dcd8026aa5c []byte

// TemplateContents generated bindings for template files
var TemplateContents = map[string][]byte{
  "template/register.html": md5d41d8cd98f00b204e9800998ecf8427e, // template/register.html
  "template/index.html": md5229c0df79c3c029fe8d0bbac997d8d43, // template/index.html
  "template/fragment/header.html": md58c7becc66f7ffad96533054873aabd5e, // template/fragment/header.html
  "template/fragment/typing.html": md5632006067794e49ff3d59b7d67212756, // template/fragment/typing.html
  "template/fragment/footer.html": md5308065b5078a49f986fc3c9f9b66e5d3, // template/fragment/footer.html
  "template/fragment/message.html": md54bfb6b3bc3f44cb86b69ec72b7161e99, // template/fragment/message.html
  "template/not_found.html": md5c3e7441b19141943f20e2dcd8026aa5c, // template/not_found.html
}

// TemplateHandlers generated handler bindings
var TemplateHandlers = map[string]func(res http.ResponseWriter, req *http.Request){
  "/register": RegisterHandler,
  "/register/": RegisterHandler,
  "/": IndexHandler,
  "/not_found": NotFoundHandler,
  "/not_found/": NotFoundHandler,
}
