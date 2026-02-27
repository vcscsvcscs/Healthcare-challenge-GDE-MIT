package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	// Initialize Zap logger
	var err error
	if os.Getenv("ENV") == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Set Gin mode
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.New()

	// Middleware
	r.Use(ginZapLogger(logger))
	r.Use(gin.Recovery())

	// Routes
	r.GET("/", handleRoot)
	r.GET("/health", handleHealth)
	r.GET("/api/v1/status", handleStatus)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "Healthcare Backend API",
		"version": "1.0.0",
	})
}

func handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
	})
}

func handleStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "healthcare-backend",
		"version": "1.0.0",
	})
}

// ginZapLogger returns a gin.HandlerFunc middleware that logs requests using Zap
func ginZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log request details
		logger.Info("Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)
	}
}
