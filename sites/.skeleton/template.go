// Code generated by gentemplate.sh; DO NOT EDIT
package _skeleton

import (
	_ "embed"
	"net/http"
)

//go:embed template/readme.html
var md56be30c9b1530fa67c37ba634b869e506 []byte

// TemplateContents generated bindings for template files
var TemplateContents = map[string][]byte{
	"template/readme.html": md56be30c9b1530fa67c37ba634b869e506, // template/readme.html
}

// TemplateHandlers generated handler bindings
var TemplateHandlers = map[string]func(res http.ResponseWriter, req *http.Request){
	"/readme":  ReadmeHandler,
	"/readme/": ReadmeHandler,
}
