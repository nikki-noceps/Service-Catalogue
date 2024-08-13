package main

import (
	"context"
	"log"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/logger"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"nikki-noceps/serviceCatalogue/presentation"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// https://www.reddit.com/r/golang/comments/199954n/what_is_gomaxprocs_actually_used_for/
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTSTP)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		osCall := <-c
		log.Printf("[OS INTERRUPT] system call: %+v", osCall)
		presentation.ToggleHealthCheck = true
		time.Sleep(5 * time.Second)
		cancel()
	}()

	cfg, err := config.Load("./config.yml")
	if err != nil {
		logger.FATAL("failed to load config", tag.NewErrorTag(err))
		return
	}
	logger.INFO("loaded config", tag.NewAnyTag("config", cfg))

	_, err = presentation.NewRouter(ctx, cfg)

}
