package main

import (
	"net/http"
	_ "net/http/pprof"
)

// StartPProfListen start debug pprof listening.
func StartPProfListen(addr string) {
	go func() {
		log.Println("pprof listening:", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Panicln("pprof listening:", err)
		}
	}()
}
