package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ForeverZer0/gldoc/ref"
	"github.com/ForeverZer0/gldoc/util"
	"github.com/spf13/cobra"
)

var (
	srvHost string
	srvPort int
	apiName string
	apiVers float32
	specs   []ref.Spec
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func find(name string) (*ref.Entry, bool) {
	for _, spec := range specs {
		if entry, ok := spec.Entries[name]; ok {
			return entry, true
		}
	}
	return nil, false
}

var rootCmd = &cobra.Command{
	Use:   "gldoc",
	Short: "GLdoc is a local HTTP server to provide simplified OpenGL documentation in JSON format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := loadSrc(apiName, apiVers)
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /entry/{name}", serveEntry)
		mux.HandleFunc("GET /entry/{name}/", serveEntry)

		mux.HandleFunc("GET /{name}", serveFunc)
		mux.HandleFunc("GET /{name}/", serveFunc)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		log.Println("Starting server...")
		go func() {
			err = http.ListenAndServe(fmt.Sprintf("%s:%d", srvHost, srvPort), mux)
			log.Fatalf("Server closed unexpectedly: %s\n", err)
			os.Exit(1)
		}()

		log.Printf("Awaiting requests at %s:%d (Ctrl+C to cancel)\n", srvHost, srvPort)
		<-sigs
		log.Writer().Write([]byte{'\r'})
		log.Println("Server stopped")
		return nil
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringVar(&srvHost, "host", "localhost", "address the server will handle requests on")
	flags.IntVar(&srvPort, "port", 8888, "port the server will handle requests on")
	flags.StringVar(&apiName, "api", "gl", "target OpenGL API (\"gl\" or \"gles\")")
	flags.Float32Var(&apiVers, "version", 0, "target version for the OpenGL API, or 0 for any/latest")
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
	fmt.Fprintf(&sb, `{"name":"%s","desc":%s,"args":{`, name, strconv.Quote(entry.Desc))
	for i, arg := range fn.Args {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `"%s":%s`, arg, strconv.Quote(entry.Params[arg]))
	}
	fmt.Fprintf(&sb, `},"seealso":%s,"errors":%s}`, util.JsonArray(entry.SeeAlso), util.JsonArray(entry.Errors))
	w.Write(sb.Bytes())
}

func loadSrc(api string, version float32) error {
	dirs, err := util.DirNames(api, version)
	if err != nil {
		return err
	}

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
