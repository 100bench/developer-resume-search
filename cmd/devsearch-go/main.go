package main

import (
	"html/template"
	"log"
	"os"

	"devsearch-go/internal/application"
	"devsearch-go/internal/domain"
	"devsearch-go/internal/infrastructure"
	"devsearch-go/internal/infrastructure/middleware"
	"devsearch-go/internal/infrastructure/utils"
	"devsearch-go/internal/interfaces/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&domain.User{}, &domain.Profile{}, &domain.Skill{}, &domain.Message{}, &domain.Project{}, &domain.Tag{}, &domain.Review{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}

	// Initialize repositories
	projectRepo := &infrastructure.GormProjectRepository{DB: db}
	userRepo := &infrastructure.GormUserRepository{DB: db}
	profileRepo := &infrastructure.GormProfileRepository{DB: db}
	skillRepo := &infrastructure.GormSkillRepository{DB: db}
	messageRepo := &infrastructure.GormMessageRepository{DB: db}

	// Initialize use cases
	projectUseCase := application.NewProjectUseCase(projectRepo)
	userUseCase := application.NewUserUseCase(userRepo, profileRepo, skillRepo, messageRepo)

	// Initialize HTTP handlers
	h := &http.Handler{ProjectUseCase: projectUseCase, UserUseCase: userUseCase}

	router := gin.Default()

	// Configure sessions
	cookieStore := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	router.Use(sessions.Sessions("devsearch_session", cookieStore))

	// Register custom template functions
	router.SetFuncMap(template.FuncMap{
		"pluralize":    utils.Pluralize,
		"sliceString":  utils.SliceString,
		"linebreaksbr": utils.Linebreaksbr,
	})

	// Load HTML templates
	t := template.New("").Funcs(template.FuncMap{
		"pluralize":    utils.Pluralize,
		"sliceString":  utils.SliceString,
		"linebreaksbr": utils.Linebreaksbr,
	})
	template.Must(t.ParseGlob("templates/**/*.html"))
	router.SetHTMLTemplate(t)

	// Serve static files
	router.Static("/static", "."+string(os.PathSeparator)+"static")
	router.Static("/media", "."+string(os.PathSeparator)+"media")

	// Project API routes
	api := router.Group("/api")
	{
		api.GET("/projects", h.GetProjects)
		// The following handlers are commented out because they are not defined on http.Handler
		// api.GET("/projects/:id", h.GetProject)
		// api.POST("/projects", h.CreateProject)
		// api.PUT("/projects/:id", h.UpdateProject)
		// api.DELETE("/projects/:id", h.DeleteProject)
	}

	// User API routes
	userAPI := router.Group("/api")
	{
		userAPI.GET("/profiles", h.GetProfiles)
		userAPI.GET("/profiles/:id", h.GetUserProfile)
		userAPI.POST("/register", h.RegisterUser)
		userAPI.POST("/login", h.LoginUser)
		userAPI.POST("/logout", h.LogoutUser)
		userAPI.GET("/account", h.GetUserAccount)
		userAPI.PUT("/account/:id", h.UpdateUserAccount)
		userAPI.POST("/skills", h.CreateSkill)
		userAPI.PUT("/skills/:id", h.UpdateSkill)
		userAPI.DELETE("/skills/:id", h.DeleteSkill)
		userAPI.GET("/inbox", h.GetInbox)
		userAPI.GET("/messages/:id", h.GetMessage)
		userAPI.POST("/messages", h.CreateMessage)
	}

	// Project HTML routes
	router.GET("/", h.RenderProjectsPage) // This will render the projects list page
	router.GET("/projects", h.RenderProjectsPage)
	router.GET("/project/:id", h.RenderSingleProjectPage)

	authRequired := router.Group("/")
	authRequired.Use(middleware.AuthRequired())
	{
		authRequired.GET("/create-project", h.RenderCreateProjectPage)
		authRequired.POST("/create-project", h.CreateProject)
		authRequired.GET("/update-project/:id", h.RenderUpdateProjectPage)
		authRequired.POST("/update-project/:id", h.UpdateProject)
		authRequired.GET("/delete-project/:id", h.RenderDeleteProjectPage)
		authRequired.POST("/delete-project/:id", h.DeleteProject)

		authRequired.GET("/account", h.RenderAccountPage)
		authRequired.GET("/edit-account", h.RenderEditAccountPage)
		authRequired.POST("/edit-account", h.UpdateUserAccount)
		authRequired.GET("/create-skill", h.RenderCreateSkillPage)
		authRequired.POST("/create-skill", h.CreateSkill)
		authRequired.GET("/update-skill/:id", h.RenderUpdateSkillPage)
		authRequired.POST("/update-skill/:id", h.UpdateSkill)
		authRequired.GET("/delete-skill/:id", h.RenderDeleteSkillPage)
		authRequired.POST("/delete-skill/:id", h.DeleteSkill)
		authRequired.GET("/inbox", h.RenderInboxPage)
		authRequired.GET("/message/:id", h.RenderMessagePage)
		authRequired.GET("/create-message/:id", h.RenderCreateMessagePage)
		authRequired.POST("/create-message/:id", h.CreateMessage)
	}

	// Public User HTML routes
	router.GET("/profiles", h.RenderProfilesPage)
	router.GET("/profile/:id", h.RenderUserProfilePage)
	router.GET("/login", h.RenderLoginRegisterPage)
	router.POST("/login", h.LoginUser)
	router.GET("/register", h.RenderLoginRegisterPage)
	router.POST("/register", h.RegisterUser)
	router.GET("/logout", h.LogoutUser)

	log.Println("Attempting to run server...")
	log.Println("Server starting on :8080")
	router.Run(":8080")
}
