package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/config"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	db     *sql.DB
	cfg    *config.Config
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize Zap logger
	if cfg.Server.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Configuration loaded successfully",
		zap.String("environment", cfg.Server.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database connection
	db, err = sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Configure database connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Successfully connected to database")

	// Set Gin mode
	if cfg.Server.Environment == "production" {
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
	r.GET("/api/v1/users", handleGetUsers)

	logger.Info("Starting server", zap.String("port", cfg.Server.Port))
	if err := r.Run(":" + cfg.Server.Port); err != nil {
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
	// Check database connection
	if err := db.Ping(); err != nil {
		c.JSON(503, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
		})
		return
	}

	c.JSON(200, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}

func handleStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "healthcare-backend",
		"version": "1.0.0",
	})
}

func handleGetUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, email, name, created_at FROM users ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		logger.Error("Failed to query users", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	type User struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
			logger.Error("Failed to scan user", zap.Error(err))
			continue
		}
		users = append(users, u)
	}

	c.JSON(200, gin.H{
		"users": users,
		"count": len(users),
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
