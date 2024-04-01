package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ForeverZer0/gldoc/ref"
	"github.com/ForeverZer0/gldoc/util"
)

var specs []ref.Spec

func find(name string) (*ref.Entry, bool) {
	for _, spec := range specs {
		if entry, ok := spec.Entries[name]; ok {
			return entry, true
		}
	}
	return nil, false
}

func loadSrc(gles bool, version float64) error {
	dirs := util.DirNames(gles, version)
	base := util.RepoPath()
	if err := util.CloneRepo(base); err != nil {
		return err
	}

	for _, dir := range dirs {
		spec, err := ref.LoadSpec(base, dir)
		if err != nil {
			return err
		}
		specs = append(specs, spec)
	}
	return nil
}

// serveEntry implements the "/entry/{entry}" route.
func serveEntry(w http.ResponseWriter, r *http.Request) {
	if entry, ok := find(r.PathValue("name")); ok {
		enc := json.NewEncoder(w)
		enc.Encode(entry)
		return
	}
	http.Error(w, "invalid function name", http.StatusBadRequest)
}

// serveFunc implements the "/{func}" route.
func serveFunc(w http.ResponseWriter, r *http.Request) {
	var entry *ref.Entry
	var fn ref.Function
	var ok bool

	name := r.PathValue("name")
	if entry, ok = find(name); ok {
		fn, ok = entry.Func(name)
	}

	if !ok {
		http.Error(w, "invalid function name", http.StatusBadRequest)
		return
	}

	var sb bytes.Buffer
	fmt.Fprintf(&sb, `{"name":"%s","desc":%s,"args":[`, name, strconv.Quote(entry.Desc))
	for i, argName := range fn.Args {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `{"name":"%s","desc":%s}`, argName, strconv.Quote(entry.Params[argName]))
	}
	fmt.Fprintf(&sb, `],"seealso":%s,"errors":%s}`, util.JsonArray(entry.SeeAlso), util.JsonArray(entry.Errors))
	w.Write(sb.Bytes())
}

func main() {
	var host string
	var port int
	var gles bool
	var version float64

	flag.IntVar(&port, "port", 8888, "port to serve HTTP requests on")
	flag.StringVar(&host, "host", "localhost", "address to serve HTTP requests on")
	flag.BoolVar(&gles, "gles", false, "documentation for OpenGLES API")
	flag.Float64Var(&version, "version", 0, "target version for the OpenGL API to document")

	fmt.Println("Loading documentation sources...")
	err := loadSrc(gles, version)
	if err != nil {
		fmt.Printf("error: failed to load sources: %s", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /entry/{name}", serveEntry)
	mux.HandleFunc("GET /entry/{name}/", serveEntry)

	mux.HandleFunc("GET /{name}", serveFunc)
	mux.HandleFunc("GET /{name}/", serveFunc)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Starting server...")
	go func() {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), mux)
		fmt.Printf("error: Server closed unexpectedly: %s\n", err)
		os.Exit(1)
	}()

	fmt.Printf("Awaiting requests at %s:%d (Ctrl+C to cancel)\n", host, port)
	<-sigs
	fmt.Println("\rServer stopped")
}
