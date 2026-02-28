package middleware

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLoggingMiddleware logs all incoming requests with detailed information
// Validates: Requirements 12.1
func RequestLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Get user ID from context if available (for authenticated requests)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Log request details
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("user_id", userID),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Time("timestamp", startTime),
		}

		// Add request ID if available
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Log at appropriate level based on status code
		status := c.Writer.Status()
		if status >= 500 {
			logger.Error("Request completed with server error", fields...)
		} else if status >= 400 {
			logger.Warn("Request completed with client error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}
	}
}

// ErrorLoggingMiddleware logs errors with stack traces and request context
// Validates: Requirements 12.2
func ErrorLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				// Log error with context
				logger.Error("Request error occurred",
					zap.Error(err.Err),
					zap.Uint64("error_type", uint64(err.Type)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.Stack("stack_trace"),
				)
			}
		}
	}
}

// RecoveryMiddleware recovers from panics and logs them with stack traces
// Validates: Requirements 12.2
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with full context
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.Stack("stack_trace"),
				)

				// Return 500 error
				c.JSON(500, gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already set in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = generateRequestID()
		}

		// Store in context
		c.Set("request_id", requestID)

		// Add to response header
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production, use UUID or similar
	return time.Now().Format("20060102150405.000000")
}

// SlowQueryLoggingMiddleware logs database queries that exceed a threshold
// This is a placeholder - actual implementation would be in the repository layer
// Validates: Requirements 12.5
func SlowQueryLoggingMiddleware(logger *zap.Logger, threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store threshold in context for repository layer to use
		c.Set("slow_query_threshold", threshold)
		c.Set("slow_query_logger", logger)
		c.Next()
	}
}
