package presentation

import (
	"context"
	"net/http"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/handlers"
	"nikki-noceps/serviceCatalogue/logger"
	"nikki-noceps/serviceCatalogue/services"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// This toggles the health check route notifying the load balancer to stop sending traffic.
	// The load balancer health checks also need to be tweeked to fully operationalize graceful shutdown
	ToggleHealthCheck = false
)

func NewRouter(ctx context.Context, cfg *config.Configuration, svc *services.Service) (*gin.Engine, error) {
	// initiate logger with log level
	if !strings.Contains(cfg.App.LogLevel, "info") {
		logger.ReplaceGlobalZapLogger(logger.NewZapLogger(cfg.App.LogLevel))
	}

	switch cfg.App.Environment {
	case "local", "dev", "stg":
		logger.INFO("gin running in debug mode")
		gin.SetMode(gin.DebugMode)
	case "prd":
		logger.INFO("gin running in release mode")
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(
		CORSMiddleware,
		CustomContextInit("catalogue"),
		loggerMiddleware(),
		ErrorMiddleware,
	)
	setupRoutes(ctx, router, handlers.NewHandler(svc))
	return router, nil
}

func setupRoutes(ctx context.Context, router *gin.Engine, handler *handlers.Handler) {
	router.GET("/health", func(c *gin.Context) {
		if ToggleHealthCheck {
			c.String(http.StatusInternalServerError, "Server Shutting Down")
		}
		c.String(http.StatusOK, "Working!")
	})
	router.GET("/serviceCatalogue/search", handler.SearchSvcCatalogue)
}
