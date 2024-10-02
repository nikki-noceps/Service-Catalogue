package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/internal/services"
	"nikki-noceps/serviceCatalogue/pkg/database"
	"nikki-noceps/serviceCatalogue/pkg/logger"
	"nikki-noceps/serviceCatalogue/pkg/logger/tag"
	"nikki-noceps/serviceCatalogue/pkg/migrations"
	"nikki-noceps/serviceCatalogue/pkg/presentation"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// https://www.reddit.com/r/golang/comments/199954n/what_is_gomaxprocs_actually_used_for/
	runtime.GOMAXPROCS(runtime.NumCPU())

	runMigration := flag.Bool("migrate", false, "Whether to run the index migration")

	// Parse the flags
	flag.Parse()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTSTP)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		osCall := <-c
		logger.INFO("[OS INTERRUPT] system call", tag.NewAnyTag("OScall", osCall))
		presentation.ToggleHealthCheck = true
		time.Sleep(5 * time.Second)
		cancel()
	}()

	cfg, esClient, err := loadDependencies(ctx)
	if err != nil {
		logger.FATAL("failed to load dependencies", tag.NewErrorTag(err))
		return
	}

	svc, err := services.NewService(ctx, esClient)
	if err != nil {
		logger.FATAL("failed to create service", tag.NewErrorTag(err))
		return
	}

	// Check if the migration flag is set
	if *runMigration {
		err := migrations.RunMigrations(ctx, cfg)
		if err != nil {
			log.Fatalf("Error migrating index: %s", err)
		}
		cancel()
		return
	}

	router, err := presentation.NewRouter(ctx, cfg, svc)
	if err != nil {
		logger.FATAL("failed to create router", tag.NewErrorTag(err))
		return
	}
	err = initServer(ctx, router, cfg)
	if err != nil {
		logger.FATAL("failed to start router", tag.NewErrorTag(err))
		return
	}
}

func initServer(ctx context.Context, router *gin.Engine, cfg *config.Configuration) error {
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.FATAL("Server closed.", tag.NewErrorTag(err))
		} else if err != nil {
			logger.ERROR("Server closed.", tag.NewErrorTag(err))
		}
	}()
	logger.INFO(fmt.Sprintf("Listening on %s", cfg.Server.Port))

	<-ctx.Done()

	logger.INFO("shutting down server")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server shutdown failed: %v", err)
		}
	}
	logger.INFO("server exited properly")
	return nil
}

func loadDependencies(ctx context.Context) (*config.Configuration, *database.ESClient, error) {
	cfg, err := config.Load("config.yml")
	if err != nil {
		return nil, nil, err
	}
	logger.INFO("loaded config", tag.NewAnyTag("config", cfg))

	esClient, err := database.InitESClient(cfg.ElasticSearch)
	if err != nil {
		return nil, nil, err
	}
	return cfg, esClient, nil
}
