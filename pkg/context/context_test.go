package context

import (
	"nikki-noceps/serviceCatalogue/pkg/logger"
	"nikki-noceps/serviceCatalogue/pkg/logger/tag"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const keyRandom = "random"

func TestCustomContext(t *testing.T) {
	expectedUserID := uuid.NewString()
	expectedMerchantID := uuid.NewString()
	expectedRequestID := uuid.NewString()
	expectedTraceID := uuid.NewString()
	expectedRandomValue := uuid.NewString()

	assertValues := func(cctx CustomContext) {
		assert.Equal(t, expectedUserID, cctx.UserID())
		assert.Equal(t, expectedRequestID, cctx.RequestID())
		assert.Equal(t, expectedTraceID, cctx.TraceID())
		assert.Equal(t, expectedRandomValue, cctx.Value(keyRandom).(string))
		cctx.Logger().INFO("message")
	}

	cctx := NewCustomContext(&CustomContextConfig{
		RequestID: expectedRequestID,
		TraceID:   expectedTraceID,
		Ctx:       context.WithValue(context.Background(), keyRandom, expectedRandomValue),
		Logger: logger.WITH(
			tag.NewAnyTag("merchantId", expectedMerchantID),
			tag.NewAnyTag("requestId", expectedRequestID),
			tag.NewAnyTag("traceId", expectedTraceID),
			tag.NewAnyTag("userId", expectedUserID),
			tag.NewAnyTag("randomId", expectedRandomValue),
		),
	})
	assertValues(cctx)

	expectedRandomValue = uuid.NewString()
	ctx := context.WithValue(cctx, keyRandom, expectedRandomValue)
	cctx = CustomContextFromContext(ctx)
	assertValues(cctx)

	expectedRandomValue = uuid.NewString()
	cctx = WithValue(cctx, keyRandom, expectedRandomValue)
	assertValues(cctx)

	cctx, cancelFunc := WithCancel(cctx)
	assertValues(cctx)
	assert.NoError(t, cctx.Err())
	cancelFunc()
	assert.Error(t, cctx.Err())
	select {
	case <-cctx.Done():
	default:
		t.FailNow()
	}
}

func TestCustomContextFromContext(t *testing.T) {
	t.Run("get context when given context is CustomContext", func(t *testing.T) {
		expectedRequestID := uuid.NewString()
		expectedTraceID := uuid.NewString()

		assertValues := func(cctx CustomContext) {
			assert.Equal(t, expectedRequestID, cctx.RequestID())
			assert.Equal(t, expectedTraceID, cctx.TraceID())
			cctx.Logger().INFO("message")
		}

		cctx := NewCustomContext(&CustomContextConfig{
			RequestID: expectedRequestID,
			TraceID:   expectedTraceID,
			Ctx:       context.Background(),
			Logger: logger.WITH(
				tag.NewAnyTag("requestId", expectedRequestID),
				tag.NewAnyTag("traceId", expectedTraceID),
			),
		})

		nctx := CustomContextFromContext(cctx)
		assertValues(nctx)
	})

	t.Run("get stored context when given context.Context which contains CustomContext already", func(t *testing.T) {

		expectedUserID := uuid.NewString()
		expectedMerchantID := uuid.NewString()
		expectedRequestID := uuid.NewString()
		expectedTraceID := uuid.NewString()

		assertValues := func(cctx CustomContext) {
			assert.Equal(t, expectedUserID, cctx.UserID())
			assert.Equal(t, expectedRequestID, cctx.RequestID())
			assert.Equal(t, expectedTraceID, cctx.TraceID())
			cctx.Logger().INFO("message")
		}

		parentCtx := context.Background()
		cctx := NewCustomContext(&CustomContextConfig{
			RequestID: expectedRequestID,
			TraceID:   expectedTraceID,
			Ctx:       parentCtx,
			Logger: logger.WITH(
				tag.NewAnyTag("merchantId", expectedMerchantID),
				tag.NewAnyTag("requestId", expectedRequestID),
				tag.NewAnyTag("traceId", expectedTraceID),
				tag.NewAnyTag("userId", expectedUserID),
			),
		})

		nctx := CustomContextFromContext(cctx)
		assertValues(nctx)

		parentCtx = StoreCustomContextInContext(parentCtx, nctx)

		ncctx := CustomContextFromContext(parentCtx)
		assertValues(ncctx)
	})
}

func TestCustomContext_Set_And_Get(t *testing.T) {
	cctx := NewCustomContext(&CustomContextConfig{Ctx: context.Background()})

	key := "key1"
	expectedVal := "val1"
	cctx.Set(key, expectedVal)

	val := cctx.Get(key)
	assert.Equal(t, expectedVal, val)
}
