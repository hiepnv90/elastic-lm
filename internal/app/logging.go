package app

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new logger instance.
// The type of logger instance will be different with different application running modes.
func NewLogger(logLevel zapcore.Level) *zap.Logger {
	atom := zap.NewAtomicLevelAt(logLevel)

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.CallerKey = "caller"

	encoder := zapcore.NewConsoleEncoder(cfg)

	return zap.New(zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atom), zap.AddCaller())
}
