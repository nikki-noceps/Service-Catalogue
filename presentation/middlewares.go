package presentation

import (
	"fmt"
	"net/http"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/logger"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"runtime/debug"
	"strings"
	"time"

	nice "github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var requestIDHeaderKey = "x-request-id"

type APIErrorResponse struct {
	Error     string    `json:"error"`
	TimeStamp time.Time `json:"timestamp"`
	RequestId string    `json:"requestId"`
}

// Logger is a middleware to log the incoming request.
func loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %v | %3d | %v ms | %15s | %s %#v\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency.Milliseconds(),
			param.ClientIP,
			param.Method,
			param.Path,
		)
	})
}

// CORS middleware for preflight request from the browser. Breaks the OPTIONS request with 204 status code
// Also can be extended to allow only certain origins for preflight requests to ensure this service only serves
// service to service calls
func CORSMiddleware(c *gin.Context) {
	// Replace with comma seperated values to implement strict cross origin.
	// Currently set to allow all domain calls to the server
	origin := "*"

	allowedCorsHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Authorization",
		"Accept",
		"X-Auth-Client-ID",
	}

	corsHeaders := map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     strings.Join(allowedCorsHeaders, ", "),
		"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, PATCH, DELETE",
		"Access-Control-Allow-Origin":      origin,
	}

	for key, value := range corsHeaders {
		c.Writer.Header().Set(key, value)
	}

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}

	c.Next()
}

// CustomContextInit creates a CustomContext out of c.Request.Context() and replace
// c.Request.Context() with created CustomContext. This is a gin compatible middleware.
func CustomContextInit(serviceName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get requestID from X-RequestID header
		requestID := ctx.Request.Header.Get(requestIDHeaderKey)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// get traceID from current span
		span := trace.SpanFromContext(ctx.Request.Context())
		span.SetAttributes(attribute.KeyValue{Key: "request_id", Value: attribute.StringValue(requestID)})

		traceID := span.SpanContext().TraceID().String()

		// initialise an attributed logger
		lg := logger.WITH(
			tag.NewAnyTag("service.name", serviceName),
			tag.NewAnyTag("requestId", requestID),
		)

		// initialise the CustomContext
		cctx := context.NewCustomContext(&context.CustomContextConfig{
			RequestID: requestID,
			TraceID:   traceID,
			Logger:    lg,
			Ctx:       ctx.Request.Context(),
		})

		// replace current request context with CustomContext
		ctx.Request = ctx.Request.WithContext(cctx)
		ctx.Next()
	}
}

func ErrorMiddleware(c *gin.Context) {
	c.Next()
	if len(c.Errors) == 0 {
		return
	}
	err := c.Errors.Last().Err
	cctx := context.CustomContextFromContext(c.Request.Context())
	c.JSON(c.Writer.Status(), &APIErrorResponse{
		Error:     err.Error(),
		TimeStamp: time.Now(),
		RequestId: cctx.RequestID(),
	})
}

func PanicRecovery() gin.HandlerFunc {
	panicHandler := func(c *gin.Context, err interface{}) {
		cctx := context.CustomContextFromContext(c.Request.Context())

		cctx.Logger().ERROR("PANIC_OCCURRED", tag.NewAnyTag("trace", string(debug.Stack())), tag.NewAnyTag("err", err))

		c.JSON(http.StatusInternalServerError, nil)
	}
	return nice.Recovery(panicHandler)
}
