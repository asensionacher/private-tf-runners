package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/private-tf-runners/server/internal/config"
	"github.com/private-tf-runners/server/internal/database"
	"github.com/private-tf-runners/server/internal/handlers"
	"github.com/private-tf-runners/server/internal/middleware"
	"github.com/private-tf-runners/server/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	authHandler := handlers.NewAuthHandler(db)
	stackHandler := handlers.NewStackHandler(db)
	runHandler := handlers.NewRunHandler(db)
	backendHandler := handlers.NewBackendHandler(db)
	runnerHandler := handlers.NewRunnerHandler(db)
	userHandler := handlers.NewUserHandler(db)

	rateLimiter := middleware.NewRateLimiter(cfg.Security.RateLimitWindow, cfg.Security.RateLimitMax)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CSP())
	router.Use(middleware.CORS())
	router.Use(middleware.NoCache())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.Use(middleware.RateLimit(rateLimiter))
			auth.GET("/csrf", authHandler.GetCSRFToken)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/me", authHandler.Me)

stacks := protected.Group("/stacks")
			{
				stacks.GET("", stackHandler.List, middleware.RequirePermission(models.PermissionStackRead))
				stacks.POST("", stackHandler.Create, middleware.RequirePermission(models.PermissionStackCreate))
				stacks.GET("/:id", stackHandler.Get, middleware.RequirePermission(models.PermissionStackRead))
				stacks.GET("/:id/refs", stackHandler.GetWithRefs, middleware.RequirePermission(models.PermissionStackRead))
				stacks.PUT("/:id", stackHandler.Update, middleware.RequirePermission(models.PermissionStackCreate))
				stacks.DELETE("/:id", stackHandler.Delete, middleware.RequirePermission(models.PermissionStackDelete))
				stacks.GET("/validate-repo", stackHandler.ValidateRepo, middleware.RequirePermission(models.PermissionStackCreate))
				stacks.GET("/:id/refetch-repo", stackHandler.RefetchRepo, middleware.RequirePermission(models.PermissionStackCreate))
				stacks.PUT("/:id/refs", stackHandler.SyncRefs, middleware.RequirePermission(models.PermissionStackCreate))
				stacks.GET("/:id/runs", runHandler.GetByStackID, middleware.RequirePermission(models.PermissionStackRead))
			}

			runs := protected.Group("/runs")
			{
				runs.GET("", runHandler.List, middleware.RequirePermission(models.PermissionRunRead))
				runs.GET("/:id", runHandler.Get, middleware.RequirePermission(models.PermissionRunRead))
				runs.POST("", runHandler.Create, middleware.RequirePermission(models.PermissionRunCreate))
				runs.POST("/:id/assign", runnerHandler.AssignRun, middleware.RequirePermission(models.PermissionRunCreate))
				runs.POST("/:id/approve", runnerHandler.ApproveRun, middleware.RequirePermission(models.PermissionRunCreate))
				runs.POST("/:id/reject", runnerHandler.RejectRun, middleware.RequirePermission(models.PermissionRunCreate))
				runs.PUT("/:id/plan-output", runnerHandler.UpdatePlanOutput, middleware.RequirePermission(models.PermissionRunCreate))
				runs.GET("/:id/wait", runnerHandler.WaitForRun, middleware.RequirePermission(models.PermissionRunRead))
			}

			runners := protected.Group("/runners")
			{
				runners.GET("", runnerHandler.List, middleware.RequirePermission(models.PermissionRunnerRead))
				runners.POST("", runnerHandler.Create, middleware.RequirePermission(models.PermissionRunnerCreate))
				runners.GET("/:id", runnerHandler.Get, middleware.RequirePermission(models.PermissionRunnerRead))
				runners.PUT("/:id", runnerHandler.Update, middleware.RequirePermission(models.PermissionRunnerCreate))
				runners.DELETE("/:id", runnerHandler.Delete, middleware.RequirePermission(models.PermissionRunnerDelete))
				runners.POST("/:id/reset-token", runnerHandler.ResetToken, middleware.RequirePermission(models.PermissionRunnerToken))
			}

			users := protected.Group("/users")
			{
				users.GET("", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.List)
				users.POST("", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.Create)
				users.GET("/:id", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.Get)
				users.PUT("/:id", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.Update)
				users.DELETE("/:id", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.Delete)
				users.POST("/:id/reset-password", middleware.RequirePermission(models.PermissionUserAdmin), userHandler.ResetPassword)
			}

			runnerAuth := middleware.NewRunnerAuthMiddleware(db)
			runnerPublic := api.Group("")
			runnerPublic.Use(runnerAuth.Handler())
			{
				runnerPublic.POST("/runners/:id/heartbeat", runnerHandler.Heartbeat)
				runnerPublic.GET("/runners/:id/runs", runnerHandler.GetRunnerRuns)
				runnerPublic.GET("/runner/stacks/:id", stackHandler.Get)
				runnerPublic.PUT("/runner/runs/:id/plan-output", runnerHandler.UpdatePlanOutput)
				runnerPublic.PUT("/runner/runs/:id/apply-output", runnerHandler.UpdateApplyOutput)
				runnerPublic.PUT("/runner/runs/:id/status", runnerHandler.UpdateRunStatus)
				runnerPublic.PUT("/runner/runs/:id/work-dir", runnerHandler.UpdateWorkDir)
				runnerPublic.GET("/runner/runs/:id/wait", runnerHandler.WaitForRun)
			}

			backends := protected.Group("/backends")
			{
				backends.GET("/schemas", backendHandler.Schemas)
				backends.GET("/schemas/:type", backendHandler.Schema)
				backends.GET("", backendHandler.List)
				backends.POST("", backendHandler.Create)
				backends.GET("/:id", backendHandler.Get)
				backends.PUT("/:id", backendHandler.Update)
				backends.DELETE("/:id", backendHandler.Delete)
			}
		}
	}

	frontendDir := "./frontend/public"
	router.Static("/static", frontendDir+"/static")

	router.NoRoute(func(c *gin.Context) {
		if _, err := os.Stat(frontendDir + c.Request.URL.Path); os.IsNotExist(err) {
			c.File(frontendDir + "/index.html")
		} else {
			c.File(frontendDir + c.Request.URL.Path)
		}
	})

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if err := db.Runner().MarkStaleOffline(90 * time.Second); err != nil {
				log.Printf("Failed to mark stale runners offline: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
