package router

import (
	"github.com/eaglepoint/authapi/internal/handler"
	"github.com/eaglepoint/authapi/internal/middleware"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/labstack/echo/v4"
)

func Setup(
	e *echo.Echo,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	seasonHandler *handler.SeasonHandler,
	matchHandler *handler.MatchHandler,
	courseHandler *handler.CourseHandler,
	resourceHandler *handler.ResourceHandler,
	moderationHandler *handler.ModerationHandler,
	reportHandler *handler.ReportHandler,
	reviewHandler *handler.ReviewHandler,
	paymentHandler *handler.PaymentHandler,
	auditHandler *handler.AuditHandler,
	tokenService *service.TokenService,
	rateLimiter *middleware.RateLimiter,
	writeLimiter *middleware.WriteLimiter,
	metricsCollector *service.Metrics,
) {
	// Global metrics middleware
	e.Use(middleware.MetricsMiddleware(metricsCollector))

	// Public health/metrics endpoints
	e.GET("/health/detailed", auditHandler.DetailedHealth)
	e.GET("/metrics", auditHandler.GetMetrics)

	// Public routes (no JWT required)
	auth := e.Group("/auth")
	auth.Use(rateLimiter.Middleware())
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// Protected routes (JWT required)
	api := e.Group("/api",
		middleware.JWTAuth(tokenService),
		rateLimiter.Middleware(),
		writeLimiter.Middleware(),
	)

	// Logout requires authentication
	api.POST("/auth/logout", authHandler.Logout)

	// Account management - Administrator only
	accounts := api.Group("/accounts",
		middleware.RequireRoles(models.RoleAdministrator),
	)
	accounts.POST("", accountHandler.Create)
	accounts.GET("", accountHandler.List)
	accounts.PUT("/:id/status", accountHandler.UpdateStatus)

	// Get account - handler does self-access check internally
	api.GET("/accounts/:id", accountHandler.Get)

	// Change password - any authenticated user, handler enforces self-only
	api.PUT("/accounts/:id/password", accountHandler.ChangePassword)

	// Scheduling routes - Scheduler and Administrator
	schedulerRoles := middleware.RequireRoles(models.RoleScheduler, models.RoleAdministrator)

	// Seasons
	seasons := api.Group("/seasons")
	seasons.POST("", seasonHandler.CreateSeason, schedulerRoles)
	seasons.GET("", seasonHandler.ListSeasons)
	seasons.GET("/:id", seasonHandler.GetSeason)

	// Teams
	teams := api.Group("/teams")
	teams.POST("", seasonHandler.CreateTeam, schedulerRoles)
	teams.GET("/season/:season_id", seasonHandler.ListTeams)

	// Venues
	venues := api.Group("/venues")
	venues.POST("", seasonHandler.CreateVenue, schedulerRoles)
	venues.GET("", seasonHandler.ListVenues)

	// Matches
	matches := api.Group("/matches")
	matches.POST("", matchHandler.CreateMatch, schedulerRoles)
	matches.POST("/import", matchHandler.ImportSchedule, schedulerRoles)
	matches.POST("/generate", matchHandler.GenerateSchedule, schedulerRoles)
	matches.GET("", matchHandler.ListMatches)
	matches.GET("/:id", matchHandler.GetMatch)
	matches.PUT("/:id", matchHandler.UpdateMatch, schedulerRoles)
	matches.PUT("/:id/status", matchHandler.TransitionStatus, schedulerRoles)

	// Match Assignments
	assignments := api.Group("/assignments")
	assignments.POST("", matchHandler.CreateAssignment, schedulerRoles)
	assignments.GET("/match/:match_id", matchHandler.ListAssignments)
	assignments.PUT("/:id/reassign", matchHandler.ReassignAssignment, schedulerRoles)
	assignments.DELETE("/:id", matchHandler.DeleteAssignment, schedulerRoles)

	// Course routes - Instructor and Administrator can create/manage
	instructorRoles := middleware.RequireRoles(models.RoleInstructor, models.RoleAdministrator)

	courses := api.Group("/courses")
	courses.POST("", courseHandler.CreateCourse, instructorRoles)
	courses.GET("", courseHandler.ListCourses)
	courses.GET("/:id", courseHandler.GetCourse)
	courses.PUT("/:id", courseHandler.UpdateCourse, instructorRoles)

	// Course outline (tree structure)
	outline := api.Group("/outline-nodes")
	outline.POST("", courseHandler.CreateOutlineNode, instructorRoles)
	outline.GET("/course/:course_id", courseHandler.GetOutlineTree)
	outline.PUT("/:id", courseHandler.UpdateOutlineNode, instructorRoles)
	outline.DELETE("/:id", courseHandler.DeleteOutlineNode, instructorRoles)

	// Course memberships
	membersCourse := api.Group("/courses/:course_id/members")
	membersCourse.POST("", courseHandler.AddMember, instructorRoles)
	membersCourse.GET("", courseHandler.ListMembers)
	membersCourse.DELETE("/:id", courseHandler.RemoveMember, instructorRoles)

	// Resources
	resources := api.Group("/resources")
	resources.POST("", resourceHandler.CreateResource, instructorRoles)
	resources.GET("", resourceHandler.ListResources)
	resources.GET("/search", resourceHandler.SearchResources)
	resources.GET("/:id", resourceHandler.GetResource)
	resources.PUT("/:id", resourceHandler.UpdateResource, instructorRoles)

	// Resource versions (file upload/download)
	resources.POST("/:id/versions", resourceHandler.UploadVersion, instructorRoles)
	resources.GET("/:id/versions", resourceHandler.ListVersions)
	resources.GET("/versions/:version_id/download", resourceHandler.DownloadVersion)
	resources.GET("/versions/:version_id/preview", resourceHandler.PreviewVersion)

	// Moderation routes
	adminOnly := middleware.RequireRoles(models.RoleAdministrator)
	reviewerRoles := middleware.RequireRoles(models.RoleReviewer, models.RoleAdministrator)

	// Sensitive word dictionaries - Administrator only
	dicts := api.Group("/moderation/dictionaries", adminOnly)
	dicts.POST("", moderationHandler.CreateDictionary)
	dicts.GET("", moderationHandler.ListDictionaries)
	dicts.GET("/:id", moderationHandler.GetDictionary)
	dicts.PUT("/:id", moderationHandler.UpdateDictionary)
	dicts.DELETE("/:id", moderationHandler.DeleteDictionary)

	// Sensitive words - Administrator only
	dicts.POST("/:dict_id/words", moderationHandler.AddWord)
	dicts.POST("/:dict_id/words/bulk", moderationHandler.AddWords)
	dicts.GET("/:dict_id/words", moderationHandler.ListWords)
	api.DELETE("/moderation/words/:id", moderationHandler.DeleteWord, adminOnly)

	// Content check - any authenticated user
	api.POST("/moderation/check", moderationHandler.CheckContent)

	// Moderation reviews - Reviewer and Administrator
	reviews := api.Group("/moderation/reviews", reviewerRoles)
	reviews.POST("", moderationHandler.CreateReview)
	reviews.GET("", moderationHandler.ListReviews)
	reviews.GET("/:id", moderationHandler.GetReview)
	reviews.PUT("/:id/decide", moderationHandler.DecideReview)

	// Reports - any authenticated user can create; Reviewer/Admin manage
	api.POST("/reports", reportHandler.CreateReport)
	reportsAdmin := api.Group("/reports", reviewerRoles)
	reportsAdmin.GET("", reportHandler.ListReports)
	reportsAdmin.GET("/:id", reportHandler.GetReport)
	reportsAdmin.PUT("/:id/status", reportHandler.UpdateReportStatus)
	reportsAdmin.PUT("/:id/assign", reportHandler.AssignReport)

	// Report evidence and notes
	reportsAdmin.POST("/:id/evidence", reportHandler.UploadEvidence)
	reportsAdmin.GET("/:id/evidence", reportHandler.ListEvidence)
	reportsAdmin.GET("/evidence/:evidence_id/download", reportHandler.DownloadEvidence)
	reportsAdmin.POST("/:id/notes", reportHandler.AddNote)
	reportsAdmin.GET("/:id/notes", reportHandler.ListNotes)

	// Review workflow routes
	// Review configs - Administrator only
	reviewConfigs := api.Group("/reviews/configs", adminOnly)
	reviewConfigs.POST("", reviewHandler.CreateConfig)
	reviewConfigs.GET("", reviewHandler.ListConfigs)
	reviewConfigs.GET("/:id", reviewHandler.GetConfig)
	reviewConfigs.PUT("/:id", reviewHandler.UpdateConfig)
	reviewConfigs.DELETE("/:id", reviewHandler.DeleteConfig)

	// Review requests - Reviewer and Administrator manage; any authenticated can submit
	api.POST("/reviews/requests", reviewHandler.SubmitRequest)
	reviewRequests := api.Group("/reviews/requests", reviewerRoles)
	reviewRequests.GET("", reviewHandler.ListRequests)
	reviewRequests.GET("/by-entity", reviewHandler.ListByEntity)
	reviewRequests.GET("/:id", reviewHandler.GetRequest)
	reviewRequests.GET("/:id/follow-up-requests", reviewHandler.ListFollowUpRequests)
	reviewRequests.POST("/:id/resubmit", reviewHandler.ResubmitRequest)

	// My assignments - any authenticated reviewer
	api.GET("/reviews/my-assignments", reviewHandler.ListMyAssignments)

	// Review levels - Reviewer and Administrator
	reviewLevels := api.Group("/reviews/levels", reviewerRoles)
	reviewLevels.GET("/request/:request_id", reviewHandler.ListLevels)
	reviewLevels.PUT("/:id/assign", reviewHandler.AssignLevel)
	reviewLevels.PUT("/:id/decide", reviewHandler.DecideLevel)

	// Follow-up records on reviews
	reviewRequests.POST("/:id/follow-ups", reviewHandler.AddFollowUp)
	reviewRequests.GET("/:id/follow-ups", reviewHandler.ListFollowUps)

	// Payments & Reconciliation routes
	financeRoles := middleware.RequireRoles(models.RoleFinanceClerk, models.RoleAdministrator)

	// Payment ledger - Finance Clerk and Administrator
	payments := api.Group("/payments", financeRoles)
	payments.POST("", paymentHandler.CreatePayment)
	payments.GET("", paymentHandler.ListPayments)
	payments.GET("/failed-retriable", paymentHandler.ListFailedRetriable)
	payments.GET("/:id", paymentHandler.GetPayment)
	payments.GET("/account/:account_id", paymentHandler.ListByAccount)
	payments.PUT("/:id/sign", paymentHandler.SignPosting)
	payments.PUT("/:id/fail", paymentHandler.FailSettlement)
	payments.PUT("/:id/retry", paymentHandler.RetrySettlement)

	// Reconciliation - Finance Clerk and Administrator
	recon := api.Group("/reconciliation", financeRoles)
	recon.GET("/summary", paymentHandler.GetDailySummary)
	recon.GET("/summary/range", paymentHandler.GetSummaryRange)
	recon.POST("/reports", paymentHandler.GenerateReconciliation)
	recon.GET("/reports", paymentHandler.ListReconciliationReports)
	recon.GET("/reports/:id", paymentHandler.GetReconciliationReport)
	recon.GET("/reports/:id/csv", paymentHandler.DownloadReconciliationCSV)

	// Audit log routes - Auditor and Administrator
	auditorRoles := middleware.RequireRoles(models.RoleAuditor, models.RoleAdministrator)

	auditLogs := api.Group("/audit", auditorRoles)
	auditLogs.GET("/logs", auditHandler.QueryLogs)
	auditLogs.GET("/logs/export", auditHandler.ExportCSV)
	auditLogs.GET("/logs/by-entity", auditHandler.ListByEntity)
	auditLogs.GET("/logs/by-actor/:actor_id", auditHandler.ListByActor)
	auditLogs.GET("/logs/:id", auditHandler.GetAuditLog)
	auditLogs.GET("/logs/tier-counts", auditHandler.CountByTier)

	// Hash chain - read/verify for Auditor and Administrator
	auditLogs.GET("/hash-chain/verify", auditHandler.VerifyHashChain)
	auditLogs.GET("/hash-chain", auditHandler.ListHashChain)

	// Audit write operations - Administrator only
	auditAdmin := api.Group("/audit", adminOnly)
	auditAdmin.POST("/hash-chain/build", auditHandler.BuildHashChain)
	auditAdmin.POST("/purge-expired", auditHandler.PurgeExpired)
}
