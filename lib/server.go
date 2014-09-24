package sandbox

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"code.google.com/p/go.tools/imports"
)

const domain = "http://echo.jpillora.com"

// const domain = "http://play.golang.org/"

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
	s.importsOpts = &imports.Options{AllErrors: true, TabWidth: 4}
	s.log = log.New(os.Stdout, "sandbox: ", 0).Printf
	return s
}

//ListenAndServe and sandbox API and frontend
func (s *Sandbox) ListenAndServe(addr string) error {
	server := &http.Server{
		Addr:           addr,
		Handler:        s,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.log("Listening at %s...", server.Addr)
	return server.ListenAndServe()
}

func (s *Sandbox) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		s.fileHandler.ServeHTTP(w, r)
		return
	}

	//only accept post from here
	if r.Method != "POST" {
		w.WriteHeader(500)
		w.Write([]byte("Invalid request"))
		return
	}

	req, _ := http.NewRequest(r.Method, domain+r.URL.Path, r.Body)
	req.Header.Set("User-Agent", "jpillora/go-sandbox")
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("post failed " + err.Error()))
		return
	}
	defer resp.Body.Close()

	resbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("post body read failed " + err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(resbody))
}

// https://godoc.org/code.google.com/p/go.tools/imports
func (s *Sandbox) imports(code []byte, w http.ResponseWriter) {
	newCode, err := imports.Process("prog.go", code, s.importsOpts)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write(newCode)
}
