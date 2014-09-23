package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// const domain = "http://echo.jpillora.com"
const domain = "http://play.golang.org/"

//Sandbox is an HTTP server
type Sandbox struct {
	server      *http.Server
	fileHandler http.Handler
	log         func(string, ...interface{})
}

//New creates a new sandbox
func New() *Sandbox {
	s := &Sandbox{}
	s.fileHandler = FileHandler
	s.log = log.New(os.Stdout, "sandbox: ", 0).Printf
	return s
}

//ListenAndServe and sandbox API and frontend
func (s *Sandbox) ListenAndServe(port int) error {
	server := &http.Server{
		Addr:           fmt.Sprintf("0.0.0.0:%d", port),
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

	compile := r.URL.Path == "/compile"
	share := r.URL.Path == "/share"

	//read request request
	if r.Method != "POST" || (!share && !compile) {
		w.WriteHeader(500)
		w.Write([]byte("Invalid request"))
		return
	}

	code, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("post body read failed " + err.Error()))
		return
	}

	if compile {
		s.compile(code, w)
	} else {
		s.share(code, w)
	}
}

func (s *Sandbox) compile(code []byte, w http.ResponseWriter) {
	form := url.Values{"version": {"2"}, "body": {string(code)}}
	s.playgroundProxy(
		"/compile",
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		bytes.NewBufferString(form.Encode()),
		w,
	)
}

func (s *Sandbox) share(code []byte, w http.ResponseWriter) {
	s.playgroundProxy(
		"/share",
		map[string]string{},
		bytes.NewBuffer(code),
		w,
	)
}

func (s *Sandbox) playgroundProxy(endpoint string, headers map[string]string, reader io.Reader, w http.ResponseWriter) {
	req, _ := http.NewRequest("POST", domain+endpoint, reader)
	req.Header.Set("User-Agent", "jpillora/go-sandbox")

	//set all headers
	for k, v := range headers {
		req.Header.Set(k, v)
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
