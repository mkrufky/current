package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/vimeo/go-util/exit"
)

func main() {
	// for graceful shutdown
	defer exit.Recover()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for signals
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
		os.Exit(0)
	}()

	cfg, err := NewConfiguration()
	if err != nil {
		log.Printf("configuration fail: %v\n", err)
		os.Exit(1)
	}

	// display configuration
	// fmt.Printf("%v\n", cfg)

	var w HistoryManager

	if cfg.LocalDatastore {
		w = NewWorker(ctx)
	} else {
		// set up postgres
		w, err = NewPqDB(ctx, &pqDbInfo{
			host: cfg.DbHost,
			port: cfg.DbPort,
			user: cfg.DbUser,
			pass: cfg.DbPass,
			name: cfg.DbName,
		}, cfg.PingTimeout, cfg.PingInterval)
		if err != nil {
			log.Printf("db fail: %v\n", err)
			os.Exit(1)
		}
	}
	defer w.Close(ctx)

	e := echo.New()

	// initialize endpoints
	endpoints(e, w)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: e,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Printf("server shutdown: %v\n", err)
	}

	s.Shutdown(ctx)
}
