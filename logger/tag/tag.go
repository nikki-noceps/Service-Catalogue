package tag

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapTag struct {
	field zap.Field
}

func (zt ZapTag) Field() zap.Field {
	return zt.field
}

func (zt ZapTag) Key() string {
	return zt.field.Key
}

func (zt ZapTag) Value() interface{} {
	enc := zapcore.NewMapObjectEncoder()
	zt.field.AddTo(enc)

	for _, val := range enc.Fields {
		return val
	}

	return nil
}

func NewAnyTag(key string, value any) ZapTag {
	return ZapTag{
		field: zap.Any(key, value),
	}
}

// key is already `error`
func NewErrorTag(value error) ZapTag {
	return ZapTag{
		field: zap.Error(value),
	}
}
