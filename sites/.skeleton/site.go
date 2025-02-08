package _skeleton

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"html/template"
	"log"
	"net/http"
	"os/user"
	"server/sites"
	"strings"
)

var (
	dataDir        string
	dbPath         string
	DatabaseHandle *sql.DB
	// Parser for form values: https://gorilla.github.io/
	decoder = schema.NewDecoder()
	tmpl    = template.New("base")
	router  = mux.NewRouter()
)

func init() {
	parseTemplates()

	router.HandleFunc("/favicon.ico", faviconHandler)
	router.PathPrefix("/static/").HandlerFunc(StaticFileHandler)
	router.HandleFunc("/", ReadmeHandler)

	// When you create a new html file in the template folder, reload.sh will
	// automatically generate a handler, based on its file path, which gets registered here
	for key := range TemplateHandlers {
		router.HandleFunc(key, TemplateHandlers[key])
	}

	sites.RegisterInit(Init)
	sites.RegisterCleanup(Cleanup)
	sites.RegisterServe(Serve)
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

func Init(config *sites.Config) error {
	usr, err := user.Current()

	if err != nil {
		return fmt.Errorf("failed to find current user: %w", err)
	}

	name := strings.ToLower(sites.GetPackageName())
	dataDir = usr.HomeDir + "/.go4ignition/sites/" + name
	dbPath = dataDir + "/" + name + ".db"

	if !sites.DirExists(dataDir) {
		sites.CreateDir(dataDir)
		log.Println("Created data directory: " + dataDir)
	}

	DatabaseHandle, err = sites.InitDb(dbPath)

	if err != nil {
		return fmt.Errorf("failed to init database: %w", err)
	}

	return nil
}

func Cleanup() error {
	err := sites.CloseDb(DatabaseHandle)

	if err != nil {
		return fmt.Errorf("failed to close db: %w", err)
	}

	return nil
}

func Serve(res http.ResponseWriter, req *http.Request) {
	// Map path handlers in init; DO NOT add them here
	router.ServeHTTP(res, req)
}
