package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"code.google.com/p/go.tools/imports"
)

var dev = os.Getenv("PROD") != "true"

const version = "0.2.4"
const userAgent = "jpillora/go-sandbox:" + version
const sandboxDomain = "www.go-sandbox.com"
const playgroundDomain = "play.golang.org"

// const domain = "http://echo.jpillora.com"

//Sandbox is an HTTP server
type Sandbox struct {
	server      *http.Server
	fileHandler http.Handler
	importsOpts *imports.Options
	log         func(string, ...interface{})
	stats       struct {
		Compiles uint64
		Imports  uint64
		Shares   uint64
		Uptime   string
	}
}

//New creates a new sandbox
func New() *Sandbox {
	s := &Sandbox{}
	s.stats.Uptime = time.Now().UTC().Format(time.RFC822)
	s.fileHandler = FileHandler
	s.importsOpts = &imports.Options{AllErrors: true, TabWidth: 4, Comments: true}
	s.log = log.New(os.Stdout, "sandbox: ", 0).Printf
	return s
}

//proxy this request onto play.golang
func (s *Sandbox) playgroundProxy(w http.ResponseWriter, r *http.Request) {
	target := "http://" + playgroundDomain + r.URL.Path
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

	//increment stats
	if strings.HasPrefix(r.URL.Path, "/compile") {
		s.stats.Compiles++
	} else {
		s.stats.Shares++
	}
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
	s.stats.Imports++
}

func (s *Sandbox) getVersion(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(version))
}

func (s *Sandbox) getStats(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(s.stats)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func (s *Sandbox) redirect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://"+sandboxDomain+r.URL.Path)
	w.WriteHeader(302)
	w.Write([]byte("Redirecting..."))
}

//ListenAndServe and sandbox API and frontend
func (s *Sandbox) ListenAndServe(addr string) error {

	r := mux.NewRouter()
	//playground proxy endpoints
	r.HandleFunc("/compile", s.playgroundProxy).Methods("POST")
	r.HandleFunc("/share", s.playgroundProxy).Methods("POST")
	r.HandleFunc("/p/{key}", s.playgroundProxy).Methods("GET")
	//server endpoints
	r.HandleFunc("/imports", s.imports).Methods("POST")
	r.HandleFunc("/version", s.getVersion).Methods("GET")
	r.HandleFunc("/stats", s.getStats).Methods("GET")
	//static files
	r.Handle("/static/{rest:.*}", s.fileHandler).Methods("GET")
	//index
	r.Handle("/", s.fileHandler).Methods("GET").Schemes("https")
	//force all GET http -> https
	r.HandleFunc("/", s.redirect).Methods("GET")

	server := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.log("Listening at %s...", addr)
	return server.ListenAndServe()
}
