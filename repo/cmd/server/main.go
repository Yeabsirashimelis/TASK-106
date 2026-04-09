package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/handler"
	"github.com/eaglepoint/authapi/internal/middleware"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/eaglepoint/authapi/internal/router"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	// Connect to database with retry
	var db *sqlx.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sqlx.Connect("postgres", cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Printf("waiting for database... attempt %d/30: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("connected to database")

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("migrations complete")

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	loginAttemptRepo := repository.NewLoginAttemptRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	seasonRepo := repository.NewSeasonRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	venueRepo := repository.NewVenueRepository(db)
	matchRepo := repository.NewMatchRepository(db)
	assignmentRepo := repository.NewMatchAssignmentRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	outlineRepo := repository.NewCourseOutlineRepository(db)
	membershipRepo := repository.NewCourseMembershipRepository(db)
	resourceRepo := repository.NewResourceRepository(db)
	resourceVersionRepo := repository.NewResourceVersionRepository(db)
	resourceTagRepo := repository.NewResourceTagRepository(db)
	sensitiveWordRepo := repository.NewSensitiveWordRepository(db)
	moderationReviewRepo := repository.NewModerationReviewRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// Initialize services
	tokenService := service.NewTokenService(cfg)
	deviceService := service.NewDeviceService(deviceRepo, cfg)
	authService := service.NewAuthService(accountRepo, refreshTokenRepo, loginAttemptRepo, tokenService, deviceService, cfg)
	accountService := service.NewAccountService(accountRepo, refreshTokenRepo, cfg)
	auditService := service.NewAuditService(auditLogRepo)
	auditService.SetConfig(cfg)
	authService.SetAuditService(auditService)
	metricsCollector := service.NewMetrics()
	observabilityService := service.NewObservabilityService(db, metricsCollector)
	seasonService := service.NewSeasonService(seasonRepo, teamRepo, venueRepo, auditService)
	matchService := service.NewMatchService(matchRepo, seasonRepo, teamRepo, venueRepo, assignmentRepo, auditService)
	courseService := service.NewCourseService(courseRepo, outlineRepo, membershipRepo, auditService)
	resourceService := service.NewResourceService(resourceRepo, resourceVersionRepo, resourceTagRepo, courseRepo, membershipRepo, auditService, cfg)
	moderationService := service.NewModerationService(sensitiveWordRepo, moderationReviewRepo, auditService)
	reportService := service.NewReportService(reportRepo, auditService, cfg)
	reviewService := service.NewReviewService(reviewRepo, auditService)

	// Wire disposition callbacks for review write-back to originating entities
	reviewService.RegisterDisposition("course", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		if decision == "Approved" {
			return courseRepo.UpdateStatus(ctx, entityID, models.CourseStatusPublished)
		}
		return nil
	})
	reviewService.RegisterDisposition("resource", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		if decision == "Approved" {
			return resourceRepo.UpdateVisibility(ctx, entityID, models.VisibilityEnrolled)
		}
		return nil
	})
	reviewService.RegisterDisposition("match", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		if decision == "Approved" {
			return matchRepo.UpdateStatus(ctx, entityID, models.MatchScheduled, nil)
		}
		return nil
	})

	paymentService := service.NewPaymentService(paymentRepo, auditService, cfg)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	seasonHandler := handler.NewSeasonHandler(seasonService)
	matchHandler := handler.NewMatchHandler(matchService)
	courseHandler := handler.NewCourseHandler(courseService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	moderationHandler := handler.NewModerationHandler(moderationService)
	reportHandler := handler.NewReportHandler(reportService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	auditHandler := handler.NewAuditHandler(auditService, observabilityService)

	// Initialize middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	writeLimiter := middleware.NewWriteLimiter(60, time.Minute) // 60 writes/min per account

	// Setup Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.LoggerWithConfig(echomw.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","method":"${method}","uri":"${uri}","status":${status},"latency_ms":${latency_human},"bytes_out":${bytes_out},"remote_ip":"${remote_ip}"}` + "\n",
	}))
	e.Use(echomw.Recover())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Ensure storage directory exists
	os.MkdirAll(cfg.StoragePath, 0750)

	// Setup routes
	router.Setup(e, authHandler, accountHandler, seasonHandler, matchHandler, courseHandler, resourceHandler, moderationHandler, reportHandler, reviewHandler, paymentHandler, auditHandler, tokenService, rateLimiter, writeLimiter, metricsCollector)

	// Graceful shutdown
	go func() {
		if err := e.Start(cfg.ServerPort); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func runMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating migration driver: %w", err)
	}

	// Determine migrations path
	migrationsPath := getMigrationsPath()
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}

func getMigrationsPath() string {
	// Check for MIGRATIONS_PATH env var first
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		return p
	}
	// Try relative to working directory
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}
	// Try relative to executable
	ex, _ := os.Executable()
	return filepath.Join(filepath.Dir(ex), "migrations")
}
