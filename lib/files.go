package sandbox

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
)

var devMode = os.Getenv("PROD") != "true"

var dir, _ = os.Getwd()

func handler(w http.ResponseWriter, req *http.Request) {

	//get path
	p := req.URL.Path
	if p == "/" {
		p = "/index.html"
	}

	chars := []rune(p)
	//strip initial slash
	p = string(chars[1:])

	var b []byte
	var err error
	//lookup asset
	if devMode {
		b, err = ioutil.ReadFile(dir + "/static/" + p)
	} /* else {
		b, err = Asset(p)
	}*/

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

	if devMode {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1
		w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0
		w.Header().Set("Expires", "0")                                         // Proxies
	}
	//write body
	w.Write(b)
}

//Handler handles all files
var FileHandler = http.HandlerFunc(handler)
