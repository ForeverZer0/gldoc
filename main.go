package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/ForeverZer0/gldoc/ref"
)

var specs []ref.Spec

// cacheDir returns the path to the cache where the application stores documentation sources.
func cacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); len(xdg) > 0 {
		return filepath.Join(xdg, "gldoc")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "gldoc")
}

// dirNames returns the names of child directories in the registry based on OpenGL API/version.
func dirNames(gles bool, version float64) []string {
	var names []string
	if gles {
		switch {
		case version == 0 || version > 3.1:
			names = append(names, "es3")
			fallthrough
		case version > 3.0:
			names = append(names, "es3.1")
			fallthrough
		case version > 2.0:
			names = append(names, "es3.0")
			fallthrough
		case version > 1.0:
			names = append(names, "es2.0")
			fallthrough
		default:
			names = append(names, "es1.0")
		}
	} else {
		if version == 0 || version > 2.1 {
			names = append(names, "gl4")
		}
		names = append(names, "gl2.1")
	}

	return names
}

// clone uses git to clone the OpenGL-Refpages repository into the local cache.
func clone(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		const repo = "https://github.com/KhronosGroup/OpenGL-Refpages.git"
		git := exec.Command("git", "clone", repo, path)
		git.Stdout = os.Stdout
		git.Stderr = os.Stderr
		return git.Run()
	}
	// Already exists
	return nil
}

// load loads and parses all required documentation sources for the specified API/version.
func load(gles bool, version float64) {
	// Use a map for collecting only unique names
	var entries, funcs map[string]bool
	entries = make(map[string]bool)
	funcs = make(map[string]bool)

	fmt.Println("Loading documentation sources... ")
	dirs := dirNames(gles, version)
	base := cacheDir()
	if err := clone(base); err != nil {
		fmt.Printf("Repo cloning failed: %s\n", err)
		os.Exit(1)
	}

	for _, dir := range dirs {
		spec, err := ref.LoadSpec(base, dir)
		if err != nil {
			fmt.Printf("Failed to load specification: %s\n", err)
			os.Exit(1)
		}
		specs = append(specs, spec)
		for _, entry := range spec.Entries {
			entries[entry.Name] = true
			for _, fn := range entry.Funcs {
				funcs[fn.Name] = true
			}
		}
	}

	fmt.Printf("Loaded %d source entries for %d functions\n", len(entries), len(funcs))
}

// jsonArray formats an array of strings into
func jsonArray(values []string) []byte {
	var sb bytes.Buffer
	sb.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Quote(value))
	}
	sb.WriteByte(']')
	return sb.Bytes()
}

// serveFunc implements the "/{func}" route.
func serveFunc(w http.ResponseWriter, r *http.Request) {
	var entry *ref.Entry
	var fn ref.Function
	var ok bool

	name := r.PathValue("name")
	for _, spec := range specs {
		if entry, ok = spec.Entries[name]; ok {
			fn, ok = entry.Func(name)
		}
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
	fmt.Fprintf(&sb, `],"seealso":%s,"errors":%s}`, jsonArray(entry.SeeAlso), jsonArray(entry.Errors))
	w.Write(sb.Bytes())
}

func start(addr string, handler http.Handler) {
	fmt.Println("Starting server... ")

	ready := make(chan bool)
	go func() {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("Failed to start server: %s\n", err)
			os.Exit(1)
		}

		ready <- true
		if err = http.Serve(listener, handler); err != nil {
			fmt.Printf("Server closed unexpectedly: %s\n", err)
			os.Exit(1)
		}
	}()

	<-ready
	close(ready)
	fmt.Printf("Awaiting requests at %s\nPress Ctrl+C to cancel\n", addr)
}

func main() {
	var host string
	var port int
	var gles bool
	var version float64

	flag.IntVar(&port, "port", 8888, "port to serve HTTP requests on")
	flag.StringVar(&host, "host", "localhost", "address to serve HTTP requests on")
	flag.BoolVar(&gles, "gles", false, "load documentation for GLES API")
	flag.Float64Var(&version, "version", 0, "target version for the OpenGL API to document")
	flag.Parse()

	load(gles, version)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{name}", serveFunc)
	mux.HandleFunc("GET /{name}/", serveFunc)
	start(fmt.Sprintf("%s:%d", host, port), mux)

	<-sigs
	fmt.Println("\rServer stopped")
}
