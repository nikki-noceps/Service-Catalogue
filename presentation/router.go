package presentation

import (
	"context"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

var ToggleHealthCheck = false

func NewRouter(ctx context.Context, cfg *config.Configuration) (*gin.Engine, error) {
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
	)
	return router, nil
}
