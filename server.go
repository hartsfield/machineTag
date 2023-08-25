package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"
)

// bolt is where the server is configured and routes are registered
func bolt() (ctx context.Context, srv *http.Server) {
	template.Must(templates.ParseGlob("internal/pages/*/*"))
	template.Must(templates.ParseGlob("internal/components/*/*"))
	template.Must(templates.ParseGlob("internal/shared/*/*"))

	var mux *http.ServeMux
	mux = http.NewServeMux()
	mux.Handle("/", checkAuth(http.HandlerFunc(home)))
	mux.HandleFunc("/ranked", getByRanked)
	mux.HandleFunc("/chron", getByChron)
	mux.HandleFunc("/post/", viewPost)
	mux.HandleFunc("/submitForm", handleForm)
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	srv = serverFromConf(mux)
	ctx, cancelCtx := context.WithCancel(context.Background())

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
		cancelCtx()
	}()

	return
}

// serverFromConf returns a *http.Server with a pre-defined configuration
func serverFromConf(mux *http.ServeMux) *http.Server {
	return &http.Server{
		Addr:              servicePort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
}
