package sandbox

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"code.google.com/p/go.tools/imports"
)

const version = "0.2.1"
const userAgent = "jpillora/go-sandbox:" + version
const domain = "http://play.golang.org"

// const domain = "http://echo.jpillora.com"

//Sandbox is an HTTP server
type Sandbox struct {
	server      *http.Server
	fileHandler http.Handler
	importsOpts *imports.Options
	log         func(string, ...interface{})
}

//New creates a new sandbox
func New() *Sandbox {
	s := &Sandbox{}
	s.fileHandler = FileHandler
	s.importsOpts = &imports.Options{AllErrors: true, TabWidth: 4, Comments: true}
	s.log = log.New(os.Stdout, "sandbox: ", 0).Printf
	return s
}

//proxy this request onto play.golang
func (s *Sandbox) playgroundProxy(w http.ResponseWriter, r *http.Request) {
	target := domain + r.URL.Path
	req, _ := http.NewRequest(r.Method, target, r.Body)
	req.Header = r.Header
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Could not contact play.golang.org: %s", err)
		return
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// https://godoc.org/code.google.com/p/go.tools/imports
func (s *Sandbox) imports(w http.ResponseWriter, r *http.Request) {
	code, _ := ioutil.ReadAll(r.Body)
	newCode, err := imports.Process("prog.go", code, s.importsOpts)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(newCode)
}

func (s *Sandbox) xdomainProxy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
	<!DOCTYPE HTML>
	<script src="//cdn.rawgit.com/jpillora/xdomain/0.6.15/dist/0.6/xdomain.min.js" master="http://go-sandbox.jpillora.com"></script>
	`))
}

func (s *Sandbox) getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(version))
}

//ListenAndServe and sandbox API and frontend
func (s *Sandbox) ListenAndServe(addr string) error {

	r := mux.NewRouter()
	r.HandleFunc("/compile", s.playgroundProxy).Methods("POST")
	r.HandleFunc("/share", s.playgroundProxy).Methods("POST")
	r.HandleFunc("/p/{key}", s.playgroundProxy).Methods("GET")
	r.HandleFunc("/imports", s.imports).Methods("POST")
	r.HandleFunc("/version", s.getVersion).Methods("GET")
	r.HandleFunc("/proxy.html", s.xdomainProxy).Methods("GET")
	r.Handle("/static/{rest:.*}", s.fileHandler).Methods("GET")
	r.Handle("/", s.fileHandler).Methods("GET")

	server := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.log("Listening at %s...", server.Addr)
	return server.ListenAndServe()
}
