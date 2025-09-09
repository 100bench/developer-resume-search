package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devsearch-go/internal/application"
	"devsearch-go/internal/domain"
	"devsearch-go/internal/infrastructure/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	ProjectUseCase *application.ProjectUseCase
	UserUseCase    *application.UserUseCase // Added for user-related operations
}

// GetProjects handles fetching all projects
func (h *Handler) GetProjects(c *gin.Context) {
	// Parse query parameters for search, page, and limit
	searchQuery := c.Query("q")
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	projects, _, err := h.ProjectUseCase.GetProjects(searchQuery, page, limit)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Failed to fetch projects")
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.JSON(http.StatusOK, projects)
}

// GetProject handles fetching a single project by ID
func (h *Handler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid project ID")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	project, err := h.ProjectUseCase.GetProjectByID(id)
	if err != nil {
		log.Printf("Project not found for ID %s: %v", idStr, err)
		utils.SetFlashMessage(c, utils.FlashError, "Project not found")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	c.JSON(http.StatusOK, project)
}

// CreateProject handles creating a new project
func (h *Handler) CreateProject(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to create project")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	profile, err := h.UserUseCase.GetProfileByID(userID)
	if err != nil {
		log.Printf("Profile not found for authenticated user %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Profile not found for authenticated user")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	demoLink := c.PostForm("demo_link")
	sourceLink := c.PostForm("source_link")
	tagsStr := c.PostForm("tags")

	project := domain.Project{
		OwnerID:     profile.ID,
		Title:       title,
		Description: description,
		DemoLink:    demoLink,
		SourceLink:  sourceLink,
	}

	// Handle featured image upload
	file, err := c.FormFile("featured_image")
	var filename string
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			log.Printf("Failed to open image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to open image file")
			c.Redirect(http.StatusFound, "/create-project")
			return
		}
		defer src.Close()

		// Ensure the media/projects directory exists
		if _, err := os.Stat("./media/projects"); os.IsNotExist(err) {
			os.MkdirAll("./media/projects", os.ModePerm)
		}

		filename = fmt.Sprintf("projects/%s%s", uuid.New().String(), filepath.Ext(file.Filename))
		dst, err := os.Create(filepath.Join(".", "media", filename))
		if err != nil {
			log.Printf("Failed to save image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to save image file")
			c.Redirect(http.StatusFound, "/create-project")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("Failed to copy image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to copy image file")
			c.Redirect(http.StatusFound, "/create-project")
			return
		}
		project.FeaturedImage = filename
	} else if err != nil && err != http.ErrMissingFile {
		log.Printf("Failed to get file: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, fmt.Sprintf("Failed to get file: %v", err))
		c.Redirect(http.StatusFound, "/create-project")
		return
	}

	tagNames := strings.Split(tagsStr, ",")
	if err := h.ProjectUseCase.CreateProject(&project, tagNames); err != nil {
		log.Printf("Failed to create project for user %s: %v", userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to create project")
		c.Redirect(http.StatusFound, "/create-project")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Project was added successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// UpdateProject handles updating an existing project
func (h *Handler) UpdateProject(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to update project")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid project ID")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	project, err := h.ProjectUseCase.GetProjectByID(id)
	if err != nil {
		log.Printf("Project not found or unauthorized for user %s, project %s: %v", userID.String(), idStr, err)
		utils.SetFlashMessage(c, utils.FlashError, "Project not found or you don't have permission")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	// Ensure the authenticated user is the owner of the project
	profile, err := h.UserUseCase.GetProfileByID(userID)
	if err != nil || project.OwnerID != profile.ID {
		utils.SetFlashMessage(c, utils.FlashError, "You don't have permission to edit this project")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	project.Title = c.PostForm("title")
	project.Description = c.PostForm("description")
	project.DemoLink = c.PostForm("demo_link")
	project.SourceLink = c.PostForm("source_link")
	tagsStr := c.PostForm("tags")

	// Handle featured image upload
	file, err := c.FormFile("featured_image")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			log.Printf("Failed to open image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to open image file")
			c.Redirect(http.StatusFound, fmt.Sprintf("/update-project/%s", idStr))
			return
		}
		defer src.Close()

		// Ensure the media/projects directory exists
		if _, err := os.Stat("./media/projects"); os.IsNotExist(err) {
			os.MkdirAll("./media/projects", os.ModePerm)
		}

		filename := fmt.Sprintf("projects/%s%s", uuid.New().String(), filepath.Ext(file.Filename))
		dst, err := os.Create(filepath.Join(".", "media", filename))
		if err != nil {
			log.Printf("Failed to save image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to save image file")
			c.Redirect(http.StatusFound, fmt.Sprintf("/update-project/%s", idStr))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Printf("Failed to copy image file: %v", err)
			utils.SetFlashMessage(c, utils.FlashError, "Failed to copy image file")
			c.Redirect(http.StatusFound, fmt.Sprintf("/update-project/%s", idStr))
			return
		}
		project.FeaturedImage = filename
	} else if err != nil && err != http.ErrMissingFile {
		log.Printf("Failed to get file: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, fmt.Sprintf("Failed to get file: %v", err))
		c.Redirect(http.StatusFound, fmt.Sprintf("/update-project/%s", idStr))
		return
	}

	tagNames := strings.Split(tagsStr, ",")
	if err := h.ProjectUseCase.UpdateProject(project, tagNames); err != nil {
		log.Printf("Failed to update project %s for user %s: %v", idStr, userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to update project")
		c.Redirect(http.StatusFound, fmt.Sprintf("/update-project/%s", idStr))
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Project was updated successfully!")
	c.Redirect(http.StatusFound, "/account")
}

// DeleteProject handles deleting a project
func (h *Handler) DeleteProject(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	if userIDStr == nil {
		utils.SetFlashMessage(c, utils.FlashError, "User not authenticated")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.Printf("Invalid user ID in session: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to delete project")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.SetFlashMessage(c, utils.FlashError, "Invalid project ID")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	project, err := h.ProjectUseCase.GetProjectByID(id)
	if err != nil {
		log.Printf("Project not found or unauthorized for user %s, project %s: %v", userID.String(), idStr, err)
		utils.SetFlashMessage(c, utils.FlashError, "Project not found or you don't have permission")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	// Ensure the authenticated user is the owner of the project
	profile, err := h.UserUseCase.GetProfileByID(userID)
	if err != nil || project.OwnerID != profile.ID {
		utils.SetFlashMessage(c, utils.FlashError, "You don't have permission to delete this project")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	if err := h.ProjectUseCase.DeleteProject(id); err != nil {
		log.Printf("Failed to delete project %s for user %s: %v", idStr, userID.String(), err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to delete project")
		c.Redirect(http.StatusFound, "/account")
		return
	}

	utils.SetFlashMessage(c, utils.FlashSuccess, "Project deleted successfully!")
	c.Redirect(http.StatusFound, "/account")
}
