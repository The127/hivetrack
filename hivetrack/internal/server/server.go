package server

import (
	"fmt"
	"net/http"

	"github.com/The127/ioc"
	"github.com/The127/mediatr"
	"github.com/gorilla/mux"
	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/config"
	"github.com/the127/hivetrack/internal/handlers"
	"github.com/the127/hivetrack/internal/middlewares"
	"github.com/the127/hivetrack/internal/repositories"
	"go.uber.org/zap"
)

// New creates and returns the configured HTTP handler (mux router).
func New(dp *ioc.DependencyProvider) http.Handler {
	cfg := ioc.GetDependency[*config.Config](dp)
	logger := ioc.GetDependency[*zap.Logger](dp)
	verifier := ioc.GetDependency[*authentication.OIDCVerifier](dp)
	med := ioc.GetDependency[mediatr.Mediator](dp)

	// Wire the handlers package logger for error logging in RespondError
	handlers.SetLogger(logger)

	r := mux.NewRouter()

	// Global middleware (order matters)
	r.Use(middlewares.RecoveryMiddleware(logger))
	r.Use(middlewares.LoggingMiddleware(logger))
	r.Use(middlewares.CORSMiddleware(cfg.Server.AllowedOrigins))

	// Inject a scoped DbContext into every request context
	r.Use(dbContextMiddleware(dp))

	// Public routes (no auth required)
	authH := handlers.NewAuthHandler(cfg)
	r.HandleFunc("/api/v1/auth/oidc-config", authH.GetOIDCConfig).Methods("GET")

	// Protected routes
	protected := r.PathPrefix("/api/v1").Subrouter()
	protected.Use(middlewares.AuthMiddleware(verifier, logger, cfg))

	// Users
	userH := handlers.NewUserHandler(med)
	protected.HandleFunc("/users/me", userH.GetMe).Methods("GET")
	protected.HandleFunc("/users", userH.ListUsers).Methods("GET")

	// Projects
	projectH := handlers.NewProjectHandler(med)
	protected.HandleFunc("/projects", projectH.ListProjects).Methods("GET")
	protected.HandleFunc("/projects", projectH.CreateProject).Methods("POST")
	protected.HandleFunc("/projects/{slug}", projectH.GetProject).Methods("GET")
	protected.HandleFunc("/projects/{id}", projectH.UpdateProject).Methods("PATCH")
	protected.HandleFunc("/projects/{id}", projectH.DeleteProject).Methods("DELETE")
	protected.HandleFunc("/projects/{slug}/members", projectH.AddMember).Methods("POST")
	protected.HandleFunc("/projects/{slug}/members/{user_id}", projectH.RemoveMember).Methods("DELETE")

	// Issues
	issueH := handlers.NewIssueHandler(med)
	protected.HandleFunc("/projects/{slug}/issues", issueH.ListIssues).Methods("GET")
	protected.HandleFunc("/projects/{slug}/issues", issueH.CreateIssue).Methods("POST")
	protected.HandleFunc("/projects/{slug}/issues/{number}", issueH.GetIssue).Methods("GET")
	protected.HandleFunc("/projects/{slug}/issues/{number}", issueH.UpdateIssue).Methods("PATCH")
	protected.HandleFunc("/projects/{slug}/issues/{number}", issueH.DeleteIssue).Methods("DELETE")
	protected.HandleFunc("/projects/{slug}/issues/{number}/triage", issueH.TriageIssue).Methods("POST")
	protected.HandleFunc("/projects/{slug}/issues/{number}/checklist", issueH.AddChecklistItem).Methods("POST")
	protected.HandleFunc("/projects/{slug}/issues/{number}/checklist/{item_id}", issueH.UpdateChecklistItem).Methods("PATCH")
	protected.HandleFunc("/projects/{slug}/issues/{number}/checklist/{item_id}", issueH.RemoveChecklistItem).Methods("DELETE")
	protected.HandleFunc("/me/issues", issueH.GetMyIssues).Methods("GET")

	// Comments
	commentH := handlers.NewCommentHandler(med)
	protected.HandleFunc("/projects/{slug}/issues/{number}/comments", commentH.ListComments).Methods("GET")
	protected.HandleFunc("/projects/{slug}/issues/{number}/comments", commentH.CreateComment).Methods("POST")
	protected.HandleFunc("/projects/{slug}/issues/{number}/comments/{comment_id}", commentH.UpdateComment).Methods("PATCH")
	protected.HandleFunc("/projects/{slug}/issues/{number}/comments/{comment_id}", commentH.DeleteComment).Methods("DELETE")

	// Sprints
	sprintH := handlers.NewSprintHandler(med)
	protected.HandleFunc("/projects/{slug}/sprints", sprintH.ListSprints).Methods("GET")
	protected.HandleFunc("/projects/{slug}/sprints", sprintH.CreateSprint).Methods("POST")
	protected.HandleFunc("/projects/{slug}/sprints/{id}", sprintH.UpdateSprint).Methods("PATCH")
	protected.HandleFunc("/projects/{slug}/sprints/{id}", sprintH.DeleteSprint).Methods("DELETE")

	// Milestones
	milestoneH := handlers.NewMilestoneHandler(med)
	protected.HandleFunc("/projects/{project_id}/milestones", milestoneH.ListMilestones).Methods("GET")

	// Labels
	labelH := handlers.NewLabelHandler(med)
	protected.HandleFunc("/projects/{project_id}/labels", labelH.ListLabels).Methods("GET")

	return r
}

// dbContextMiddleware creates a scoped DbContext per request and injects it into the context.
func dbContextMiddleware(dp *ioc.DependencyProvider) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scope := dp.NewScope()
			defer func() { _ = scope.Close() }()

			db := ioc.GetDependency[repositories.DbContext](scope)
			ctx := repositories.ContextWithDbContext(r.Context(), db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Serve starts the HTTP server and blocks.
func Serve(dp *ioc.DependencyProvider) error {
	cfg := ioc.GetDependency[*config.Config](dp)
	logger := ioc.GetDependency[*zap.Logger](dp)

	handler := New(dp)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	logger.Info("starting server", zap.String("addr", addr))
	return http.ListenAndServe(addr, handler)
}
