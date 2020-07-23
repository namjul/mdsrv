package main

import (
	"bytes"
	"flag"
	"fmt"
	"text/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/russross/blackfriday"
)

type server struct {
	addr string
	dir  string
	tmpl *template.Template
	cert string
	key  string
}

func html(w http.ResponseWriter, content string, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func text(w http.ResponseWriter, content string, status int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(content)), 10))
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func (s server) handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "readme"
	}

	if !strings.HasSuffix(r.URL.Path, ".md") {
		r.URL.Path += ".md"
	}

	b, err := ioutil.ReadFile(filepath.Join(s.dir, r.URL.Path))

	if err != nil {
		if os.IsNotExist(err) {
			text(w, "Not Found", http.StatusNotFound)
			return
		}

		fmt.Fprintf(os.Stderr, "%s\n", err)

		text(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") || s.tmpl == nil {
		text(w, string(b), http.StatusOK)
		return
	}

	buf := &bytes.Buffer{}

	md := blackfriday.Run(b, blackfriday.WithExtensions(blackfriday.CommonExtensions))

	data := struct {
		Document string
	}{
		Document: string(md),
	}

	if err := s.tmpl.Execute(buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)

		text(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	html(w, buf.String(), http.StatusOK)
}

func (s server) serve() error {
	if s.cert != "" && s.key != "" {
		return http.ListenAndServeTLS(s.addr, s.cert, s.key, http.HandlerFunc(s.handle))
	}
	return http.ListenAndServe(s.addr, http.HandlerFunc(s.handle))
}

func main() {
	var (
		addr string
		dir  string
		tmpl string
		cert string
		key  string
	)

	flag.StringVar(&addr, "addr", ":8080", "the address to serve on")
	flag.StringVar(&dir, "dir", ".", "the directory to serve")
	flag.StringVar(&tmpl, "tmpl", "", "the template file for the documents")
	flag.StringVar(&cert, "cert", "", "the server certificate to use for TLS")
	flag.StringVar(&key, "key", "", "the server key to use for TLS")
	flag.Parse()

	srv := server{
		addr: addr,
		dir:  dir,
		cert: cert,
		key:  key,
	}

	if tmpl != "" {
		b, err := ioutil.ReadFile(tmpl)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to read template file: %s\n", os.Args[0], err)
			os.Exit(1)
		}

		t, err := template.New(tmpl).Parse(string(b))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to parse template: %s\n", os.Args[0], err)
			os.Exit(1)
		}
		srv.tmpl = t
	}

	info, err := os.Stat(dir)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to stat document directory: %s\n", os.Args[0], err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s: document directory is not a directory\n", os.Args[0])
		os.Exit(1)
	}

	if err := srv.serve(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
