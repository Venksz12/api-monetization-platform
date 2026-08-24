package observability

import "go.uber.org/zap"

func NewLogger() *zap.Logger {
	l, err := zap.NewProduction()
	if err != nil {
		return zap.NewNop()
	}
	return l
}
