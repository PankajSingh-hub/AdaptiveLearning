package main

import (
	"log"
	"time"
	"adaptive-assessment/internal/apidocs"
	"adaptive-assessment/internal/coding"

	"adaptive-assessment/internal/config"
	"adaptive-assessment/internal/database"
	"adaptive-assessment/internal/handler"
	"adaptive-assessment/internal/middleware"
	"adaptive-assessment/internal/models"
	"adaptive-assessment/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)
	database.Seed(db)

	authSvc := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiryHrs)
	reportSvc := service.NewReportService(db)
	interviewSvc := service.NewInterviewService(db, service.NewAdaptiveEngine(), reportSvc)
	questionSvc := service.NewQuestionService(db)
	assessmentSvc := service.NewAssessmentService(db)
	analyticsSvc := service.NewAnalyticsService(db, reportSvc)
	candidateSvc := service.NewCandidateService(db, reportSvc)
	codingSvc := service.NewCodingService(db, coding.NewLocalExecutor(2))

	authH := handler.NewAuthHandler(authSvc)
	interviewH := handler.NewInterviewHandler(interviewSvc)
	questionH := handler.NewQuestionHandler(questionSvc)
	assessmentH := handler.NewAssessmentHandler(assessmentSvc)
	reportH := handler.NewReportHandler(reportSvc)
	analyticsH := handler.NewAnalyticsHandler(analyticsSvc)
	candidateH := handler.NewCandidateHandler(candidateSvc)
codingH := handler.NewCodingHandler(codingSvc)
		r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.AllowedOrigin},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	apidocs.Register(r)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	protected := api.Group("", middleware.Auth(authSvc))
	protected.GET("/auth/me", authH.Me)

	// Any logged-in user (handlers apply ownership rules)
	protected.GET("/assessments", assessmentH.List)
	protected.GET("/assessments/:id", assessmentH.Get)
	protected.GET("/reports", reportH.List)
	protected.GET("/reports/:id", reportH.Get)
	protected.GET("/analytics/candidate/:id", analyticsH.Candidate)

	candidate := protected.Group("/interviews", middleware.RequireRole(string(models.RoleCandidate)))
	candidate.POST("/start", interviewH.Start)
	candidate.GET("/:id", interviewH.Get)
	candidate.POST("/:id/answer", interviewH.Answer)
	candidate.POST("/:id/finish", interviewH.Finish)
	cod := protected.Group("/coding", middleware.RequireRole(string(models.RoleCandidate)))

cod.GET("/languages", codingH.Languages)
cod.POST("/start", codingH.Start)
cod.GET("/sessions/:id", codingH.Session)
cod.GET("/sessions/:id/problems/:order", codingH.Problem)
cod.POST("/sessions/:id/problems/:order/run", codingH.Run)
cod.POST("/sessions/:id/problems/:order/submit", codingH.Submit)
cod.POST("/sessions/:id/finish", codingH.Finish)

	admin := protected.Group("", middleware.RequireRole(string(models.RoleAdmin)))
	admin.GET("/questions", questionH.List)
	admin.GET("/questions/:id", questionH.Get)
	admin.POST("/questions", questionH.Create)
	admin.PUT("/questions/:id", questionH.Update)
	admin.DELETE("/questions/:id", questionH.Delete)
	admin.POST("/assessments", assessmentH.Create)
	admin.PUT("/assessments/:id", assessmentH.Update)
	admin.DELETE("/assessments/:id", assessmentH.Delete)
	admin.GET("/candidates", candidateH.List)
	admin.GET("/candidates/:id", candidateH.Get)
	admin.GET("/analytics/admin", analyticsH.Admin)

	log.Printf("server running on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}