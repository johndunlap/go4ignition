package sites

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var Port int

type Config interface {
}

var initRegistry = make(map[string]func(config *Config) error)
var cleanupRegistry = make(map[string]func() error)
var serveRegistry = make(map[string]func(http.ResponseWriter, *http.Request))

func init() {
	Port = *FlagInt("port", "The port to listen on", "G4I_PORT", 8002)
}

func RegisterInit(fn func(config *Config) error) {
	packageName := GetPackageName()
	_, exists := initRegistry[packageName]

	if exists {
		panic(fmt.Sprintf("init function for %s is already registered", packageName))
	}

	initRegistry[packageName] = fn
}

func RegisterCleanup(fn func() error) {
	packageName := GetPackageName()
	_, exists := cleanupRegistry[packageName]

	if exists {
		panic(fmt.Sprintf("cleanup function for %s is already registered", packageName))
	}

	cleanupRegistry[packageName] = fn
}

func RegisterServe(fn func(http.ResponseWriter, *http.Request)) {
	packageName := GetPackageName()
	_, exists := serveRegistry[packageName]

	if exists {
		panic(fmt.Sprintf("serve function for %s is already registered", packageName))
	}

	serveRegistry[packageName] = fn

	// Automatically add localhost.YOUR.SITE to support local development
	serveRegistry["localhost."+packageName] = fn

	fmt.Println("http://localhost." + packageName + ":" + strconv.Itoa(Port) + "/")
}

func Serve(domain string, w http.ResponseWriter, r *http.Request) {
	fn, exists := serveRegistry[domain]

	if exists {
		fn(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte("404 Not Found"))

		if err != nil {
			println("failed to write to response body" + err.Error())
		}
	}
}

func Serves(domain string) bool {
	_, ok := serveRegistry[domain]
	return ok
}

func Init(config *Config) error {
	for name, fn := range initRegistry {
		fmt.Println("Running init for", name)
		err := fn(config)

		if err != nil {
			return fmt.Errorf("failed to run %s init function: %v", name, err)
		}
	}

	return nil
}

func Cleanup() error {
	for name, fn := range cleanupRegistry {
		fmt.Println("Running cleanup for", name)

		err := fn()

		if err != nil {
			return fmt.Errorf("failed to run %s cleanup function: %v", name, err)
		}
	}

	return nil
}

func GetPackageName() string {
	_, filename, _, _ := runtime.Caller(2)
	dir := filepath.Dir(filename)
	parts := strings.Split(dir, "/")
	packageName := parts[len(parts)-1]
	return strings.ReplaceAll(packageName, "_", ".")
}
