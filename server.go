package opm

import (
	sdtContext "context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	DefaultHost  = "localhost"
	DefaultPort  = 8080
	WriteTimeout = 45 * time.Second
	ReadTimeout  = 45 * time.Second
)

type Server struct {
	Name         string
	Host         string
	Port         int
	Addr         string
	Handler      http.Handler
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func (s *Server) Run() {
	httphost := DefaultHost
	if s.Host != "" {
		httphost = s.Host
	}

	httpport := DefaultPort
	if s.Port > 0 {
		httpport = s.Port
	}

	if s.Addr == "" {
		s.Addr = fmt.Sprintf("%s:%d", httphost, httpport)
	}

	if s.WriteTimeout < 1 {
		s.WriteTimeout = WriteTimeout
	}

	if s.ReadTimeout < 1 {
		s.ReadTimeout = ReadTimeout
	}

	srv := &http.Server{
		Addr:         s.Addr,
		WriteTimeout: s.WriteTimeout,
		ReadTimeout:  s.ReadTimeout,
		IdleTimeout:  time.Second * 60,
		Handler:      s.Handler,
	}

	go func() {
		fmt.Printf("Server running %s\n", s.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Println("commencing server shutdown...")
			} else {
				log.Panicf("Listen and serve: %v\n", err)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := sdtContext.WithTimeout(sdtContext.Background(), time.Second*15)
	defer cancel()

	srv.Shutdown(ctx)
	log.Printf("shutdown serve: %s", s.Name)
	os.Exit(0)
}
