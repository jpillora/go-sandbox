package sandbox

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"golang.org/x/tools/imports"
)

var dev = os.Getenv("DEV") != ""

const version = "0.2.7"
const userAgent = "go-sandbox/" + version
const sandboxDomain = "go-sandbox.com"
const playgroundDomain = "play.golang.org"

var client = &http.Client{
	Transport: &http.Transport{
		DisableCompression: true,
	},
}

//Sandbox is an HTTP server
type Sandbox struct {
	server      *http.Server
	fileHandler http.Handler
	importsOpts *imports.Options
	logf        func(string, ...interface{})
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
	s.logf = log.New(os.Stdout, "sandbox: ", 0).Printf
	return s
}

//proxy this request onto play.golang
func (s *Sandbox) playgroundProxy(w http.ResponseWriter, r *http.Request) {
	target := "https://" + playgroundDomain + r.URL.Path
	req, _ := http.NewRequest(r.Method, target, r.Body)
	req.Header = r.Header
	req.Header.Set("User-Agent", userAgent)

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

type Reply struct {
	Errors        string
	CompileErrors string `json:"compile_errors,omitempty"`
	Events        []*Event
	NewCode       string
}

type Event struct {
	Delay         int
	Kind, Message string
}

//modified copy of playgroundProxy()
func (s *Sandbox) importsCompile(w http.ResponseWriter, r *http.Request) {
	//prepare reply
	reply := &Reply{}
	w.Header().Set("Content-Type", "application/json")

	code, _ := ioutil.ReadAll(r.Body)
	newCode, err := imports.Process("main.go", code, s.importsOpts)
	s.stats.Imports++
	if err != nil {
		reply.Errors = err.Error()
		b, _ := json.Marshal(reply)
		w.Write(b)
		return
	}

	v := url.Values{}
	v.Set("version", "2")
	v.Set("body", string(newCode))
	form := v.Encode()
	body := strings.NewReader(form)
	req, err := http.NewRequest("POST", "https://"+playgroundDomain+"/compile", body)
	if err != nil {
		http.Error(w, "playground request error", http.StatusBadGateway)
		return
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "playground gateway error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	//optional de-gzip
	reader := resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(reader)
		if err != nil {
			http.Error(w, "playground gzip error", http.StatusBadGateway)
			return
		}
		defer gr.Close()
		reader = gr
	}
	//copy JSON into buffer
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		http.Error(w, "playground data error", http.StatusBadGateway)
		return
	}
	//if necessary, include updated code
	if bytes.Compare(code, newCode) != 0 {
		if err := json.Unmarshal(b, reply); err != nil {
			http.Error(w, "playground unmarshal error", http.StatusBadGateway)
			return
		}
		reply.NewCode = string(newCode)
		b, _ = json.Marshal(reply)
	}
	//write response!
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(resp.StatusCode)
	gw := gzip.NewWriter(w)
	gw.Write(b)
	gw.Close()
	s.stats.Compiles++
	s.logf("compile #%d success", s.stats.Compiles)
}

func (s *Sandbox) imports(w http.ResponseWriter, r *http.Request) {
	code, _ := ioutil.ReadAll(r.Body)
	newCode, err := imports.Process("main.go", code, s.importsOpts)
	s.stats.Imports++
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(newCode)
	s.logf("goimports #%d success", s.stats.Imports)
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
	r.HandleFunc("/importscompile", s.importsCompile).Methods("POST")
	r.HandleFunc("/imports", s.imports).Methods("POST")
	r.HandleFunc("/version", s.getVersion).Methods("GET")
	r.HandleFunc("/stats", s.getStats).Methods("GET")
	//static files
	r.Handle("/static/{rest:.*}", s.fileHandler).Methods("GET")
	//redirect from old domain
	r.HandleFunc("/", s.redirect).Host("go-sandbox.jpillora.com").Methods("GET")
	r.HandleFunc("/", s.redirect).Host("www.go-sandbox.com").Methods("GET")
	//index
	r.Handle("/", s.fileHandler).Methods("GET")
	server := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.logf("Listening at %s...", addr)
	return server.ListenAndServe()
}
