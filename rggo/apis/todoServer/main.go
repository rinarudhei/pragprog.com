package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	host := flag.String("h", "localhost", "Server host")
	port := flag.String("p", "8080", "Server port")
	todoFile := flag.String("f", "todoServer.json", "todo JSON file")
	flag.Parse()

	s := http.Server{
		Addr:         net.JoinHostPort(*host, *port),
		Handler:      newMux(*todoFile),
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
