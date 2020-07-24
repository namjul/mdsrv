package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var (
	version string
	build   string
)

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

func resolvePath(dir, path string) (string, error) {
	path = strings.Replace(path, "..", "", -1)

	if path == "/" {
		path = "readme.md"
	}

	realpath := filepath.Join(dir, path)

	info, err := os.Stat(realpath)

	if err != nil {
		if os.IsNotExist(err) {
			if !strings.HasSuffix(path, ".md") {
				return realpath + ".md", nil
			}
			return path, nil
		}
		return "", nil
	}

	if info.IsDir() {
		return filepath.Join(realpath, "readme.md"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return realpath + ".md", nil
	}
	return realpath, nil
}

func documentHandler(log *log.Logger, dir string, t *template.Template) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realpath, err := resolvePath(dir, r.URL.Path)

		if err != nil {
			if os.IsNotExist(err) {
				text(w, "Not Found", http.StatusNotFound)
				return
			}
			log.Println("ERROR", err)
			text(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		b, err := ioutil.ReadFile(realpath)

		if err != nil {
			if os.IsNotExist(err) {
				text(w, "Not Found", http.StatusNotFound)
				return
			}
			log.Println("ERROR", err)
			text(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		if strings.HasPrefix(r.Header.Get("Accept"), "text/plain") || t == nil {
			text(w, string(b), http.StatusOK)
			return
		}

		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		)

		mdBuf := bytes.Buffer{}

		if err := md.Convert(b, &mdBuf); err != nil {
			log.Println("ERROR", err)
			text(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		name := strings.TrimSuffix(filepath.Base(realpath), ".md")
		title := strings.Title(strings.Replace(name, "-", " ", -1))

		data := struct {
			Title    string
			Path     string
			Document string
		}{
			Title:    title,
			Path:     r.URL.Path,
			Document: mdBuf.String(),
		}

		buf := bytes.Buffer{}

		if err := t.Execute(&buf, data); err != nil {
			log.Println("ERROR", err)
			text(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		html(w, buf.String(), http.StatusOK)
	})
}

func serve(srv *http.Server, cert, key string) error {
	if cert != "" && key != "" {
		return srv.ListenAndServeTLS(cert, key)
	}
	return srv.ListenAndServe()
}

func main() {
	var (
		addr        string
		assets      string
		dir         string
		tmpl        string
		cert        string
		key         string
		logname     string
		showversion bool
	)

	flag.StringVar(&addr, "addr", ":8080", "the address to serve on")
	flag.StringVar(&assets, "assets", "", "the directory of assets to serve")
	flag.StringVar(&dir, "dir", ".", "the directory to serve")
	flag.StringVar(&tmpl, "tmpl", "", "the template file for the documents")
	flag.StringVar(&cert, "cert", "", "the server certificate to use for TLS")
	flag.StringVar(&key, "key", "", "the server key to use for TLS")
	flag.StringVar(&logname, "log", "/dev/stdout", "the file to log errors to")
	flag.BoolVar(&showversion, "version", false, "show the version")
	flag.Parse()

	if showversion {
		fmt.Println(version, "-", build)
		return
	}

	var t *template.Template

	if tmpl != "" {
		b, err := ioutil.ReadFile(tmpl)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to read template file: %s\n", os.Args[0], err)
			os.Exit(1)
		}

		t, err = template.New(tmpl).Parse(string(b))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: failed to parse template: %s\n", os.Args[0], err)
			os.Exit(1)
		}
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

	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC)

	if logname != "/dev/stdout" {
		f, err := os.OpenFile(logname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0644))

		if err != nil {
			log.Println("ERROR", "failed to open log file", err, "using stdout")
		} else {
			defer f.Close()

			log.SetOutput(f)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", documentHandler(log, dir, t))

	if assets != "" {
		mux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(assets))))
	}


	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := serve(srv, cert, key); err != nil {
			if err != http.ErrServerClosed {
				log.Println("ERROR", err)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("INFO  ", "serving markdown documents in", dir, "on", addr)

	if assets != "" {
		log.Println("INFO  ", "serving assets from", dir)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	sig := <-c

	srv.Shutdown(ctx)

	log.Println("INFO  ", "received signal", sig, "shutting down")
}
