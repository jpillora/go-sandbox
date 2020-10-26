package sandbox

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

var dir, _ = os.Getwd()

func handler(w http.ResponseWriter, req *http.Request) {

	//get path
	p := req.URL.Path
	if p == "/" {
		p = "/index.html"
	}

	p = strings.ReplaceAll(p, "..", "")

	if dev {
		fmt.Printf("GET: %s\n", p)
	}

	//lookup asset
	b, err := ioutil.ReadFile(dir + p)

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	//lookup mimetype
	t := mime.TypeByExtension(path.Ext(p))
	if t != "" {
		w.Header().Set("Content-Type", t)
	}

	if dev {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0
		w.Header().Set("Expires", "0")                                         // Proxies
	}
	w.WriteHeader(200)
	//write body
	w.Write(b)
}

//Handler handles all files
var FileHandler = http.HandlerFunc(handler)
