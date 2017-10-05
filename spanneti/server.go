package spanneti

import (
	"net/http"
	"time"
)

func (s *spanneti) startServer() {
	mux := http.NewServeMux()
	for name, plugin := range s.plugins {
		if plugin.httpHandler != nil {
			mux.Handle("/"+name+"/", http.StripPrefix("/"+name, plugin.httpHandler))
		}
	}

	srv := &http.Server{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		Handler:      mux,
		Addr:         ":8080",
	}
	go srv.ListenAndServe()
}
