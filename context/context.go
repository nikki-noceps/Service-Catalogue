package context

import (
	"nikki-noceps/serviceCatalogue/logger"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context"
)

const (
	keyRequestID = "requestID"
	keyUserID    = "userID"
	keyTraceID   = "traceID"
	keyLogger    = "logger"
)

// CustomContextConfig is a group of options for CustomContext.
//
// Warning: New fields may be added into it in future releases.
type CustomContextConfig struct {
	RequestID string
	TraceID   string
	Logger    logger.Logger
	Ctx       context.Context
}

// CustomContext stores in and retrieves from underlying context,
// request scoped metadata including the logger for the request
// coming in Prism ecosystem. It also satisfies the standard
// context.Context interface.
//
// Warning: New methods may be added into it in future releases.
//
// All fields that were previously stored in `CustomContext` are
// now stored in `ctx`. This decision was made so developers don't
// have to carefully transfer values when extending `CustomContext`
// for custom context behaviour. Going forward, any new attributes
// should be added in `ctx`.
type CustomContext struct {
	ctx context.Context
	mu  *sync.RWMutex

	// keys will hold key-value pairs for this CustomContext
	// also map are passed around as reference this will not get copied
	// when CustomContext will be copied e.g. in passing between functions
	keys map[string]string
}

// NewCustomContext returns a new CustomContext.
func NewCustomContext(cfg *CustomContextConfig) CustomContext {
	rctx := context.Background()
	if cfg.Ctx != nil {
		rctx = cfg.Ctx
	}

	// If logger is provided in configuration, it will be added in `rctx`.
	// Otherwise if `rctx` doesn't contain a logger, a new logger will be set.
	if cfg.Logger != nil {
		rctx = context.WithValue(rctx, keyLogger, cfg.Logger)
	} else if rctx.Value(keyLogger) == nil {
		rctx = context.WithValue(rctx, keyLogger, logger.WITH())
	}

	// Setting various attributes in context if provided in `cfg`.
	if cfg.TraceID != "" {
		rctx = context.WithValue(rctx, keyTraceID, cfg.TraceID)
	}
	if cfg.RequestID != "" {
		rctx = context.WithValue(rctx, keyRequestID, cfg.RequestID)
	}

	bctx := CustomContext{
		ctx:  rctx,
		mu:   &sync.RWMutex{},
		keys: map[string]string{},
	}

	return bctx
}

// UserID returns UserID associated with the request otherwise returns "".
func (b CustomContext) UserID() string {
	id, _ := b.ctx.Value(keyUserID).(string)
	return id
}

// RequestID returns RequestID associated with the request otherwise returns "".
func (b CustomContext) RequestID() string {
	id, _ := b.ctx.Value(keyRequestID).(string)
	return id
}

// TraceID returns TraceID associated with the request otherwise returns "".
func (b CustomContext) TraceID() string {
	id, _ := b.ctx.Value(keyTraceID).(string)
	return id
}

// Get fetches a key-val pair if its store in the custom context.
func (b CustomContext) Get(key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.keys[key]
}

// Set stores a key-val pair in the custom context.
func (b CustomContext) Set(key, value string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.keys[key] = value
}

// Logger return logger associated with the request. If CustomContext contains a valid
// and sampled span inside it the returned logger will have the spanID as attribute in it.
func (b CustomContext) Logger() logger.Logger {
	// get the current span
	span := trace.SpanFromContext(b.ctx)
	var lg = logger.WITH()
	if ctxLogger, ok := b.ctx.Value(keyLogger).(logger.Logger); ok {
		lg = ctxLogger
	}
	if span.IsRecording() {
		sctx := span.SpanContext()
		if sctx.IsValid() && sctx.IsSampled() {
			return lg.WITH(tag.NewAnyTag("span.id", sctx.SpanID().String()))
		}
	}
	return lg
}

// WithContext creates a shallow copy of CustomContext with underlying context
// changed to ctx.
func (b CustomContext) WithContext(ctx context.Context) CustomContext {
	return NewCustomContext(&CustomContextConfig{
		RequestID: b.RequestID(),
		TraceID:   b.TraceID(),
		Logger:    b.Logger(),
		Ctx:       ctx,
	})
}

// Deadline calls underlying context's Deadline method.
func (b CustomContext) Deadline() (deadline time.Time, ok bool) {
	return b.ctx.Deadline()
}

// Done calls underlying context's Done method.
func (b CustomContext) Done() <-chan struct{} {
	return b.ctx.Done()
}

// Err calls underlying context's Err method.
func (b CustomContext) Err() error {
	return b.ctx.Err()
}

// Value calls underlying context's Value method.
func (b CustomContext) Value(key interface{}) interface{} {
	return b.ctx.Value(key)
}

// CustomContextFromContext returns CustomContext from given context.Context, if it fails
// returns an empty CustomContext.
func CustomContextFromContext(ctx context.Context) CustomContext {
	if bctx, ok := ctx.(CustomContext); ok {
		return bctx
	}

	bctxRaw := ctx.Value("CustomContext")
	if bctx, ok := bctxRaw.(CustomContext); ok {
		return bctx
	}

	return NewCustomContext(&CustomContextConfig{Ctx: ctx})
}

func StoreCustomContextInContext(ctx context.Context, bctx CustomContext) context.Context {
	return context.WithValue(ctx, "CustomContext", bctx)
}

// WithValue returns a copy of `parent` with `value` associated to `key`.
func WithValue(parent CustomContext, key, value interface{}) CustomContext {
	parent.ctx = context.WithValue(parent, key, value)
	return parent
}

// WithCancel returns a copy of `parent` with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the `parent` context's Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithCancel(parent CustomContext) (CustomContext, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(parent)
	parent.ctx = ctx
	return parent, cancelFunc
}

// WithDeadline returns a copy of the `parent` with the deadline adjusted
// to be no later than `d`. If the `parent`'s deadline is already earlier than `d`,
// WithDeadline(`parent`, `d`) is semantically equivalent to `parent`. The returned
// context's Done channel is closed when the deadline expires, when the returned
// cancel function is called, or when the `parent` context's Done channel is
// closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithDeadline(parent CustomContext, d time.Time) (CustomContext, context.CancelFunc) {
	ctx, cancelFunc := context.WithDeadline(parent, d)
	parent.ctx = ctx
	return parent, cancelFunc
}

// WithTimeout returns WithDeadline(`parent`, time.Now().Add(`timeout`)).
func WithTimeout(parent CustomContext, timeout time.Duration) (CustomContext, context.CancelFunc) {
	ctx, cancelFunc := context.WithTimeout(parent, timeout)
	parent.ctx = ctx
	return parent, cancelFunc
}
