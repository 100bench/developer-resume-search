package http

import (
	"log"
	"net/http"
	"strconv"

	"devsearch-go/internal/infrastructure/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) RenderProjectsPage(c *gin.Context) {
	// Get authenticated user ID for template rendering
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	// Search logic
	searchQuery := c.Query("search_query")

	// Pagination logic
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit := 3 // Items per page, consistent with Django project

	projects, totalProjects, err := h.ProjectUseCase.GetProjects(searchQuery, page, limit)
	if err != nil {
		log.Printf("Error fetching projects: %v", err)
		utils.SetFlashMessage(c, utils.FlashError, "Failed to load projects")
		c.Redirect(http.StatusFound, "/")
		return
	}

	pagination := utils.Paginate(c, int(totalProjects), limit)

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Projects = projects
	data.SearchQuery = searchQuery
	data.Pagination = pagination
	c.HTML(http.StatusOK, "projects.html", data)
}

func (h *Handler) RenderSingleProjectPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

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

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Project = *project
	c.HTML(http.StatusOK, "single-project.html", data)
}

func (h *Handler) RenderCreateProjectPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

	data := utils.GetTemplateData(c, isAuthenticated)
	data.FormTitle = "Create Project"
	c.HTML(http.StatusOK, "form-template.html", data)
}

func (h *Handler) RenderUpdateProjectPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

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

	// Ensure the authenticated user is the owner of the project
	profile, err := h.UserUseCase.GetProfileByID(uuid.MustParse(userIDStr.(string)))
	if err != nil || project.OwnerID != profile.ID {
		utils.SetFlashMessage(c, utils.FlashError, "You don't have permission to edit this project")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.FormTitle = "Update Project"
	data.Project = *project
	c.HTML(http.StatusOK, "form-template.html", data)
}

func (h *Handler) RenderDeleteProjectPage(c *gin.Context) {
	session := sessions.Default(c)
	userIDStr := session.Get("userID")
	isAuthenticated := userIDStr != nil

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

	// Ensure the authenticated user is the owner of the project
	profile, err := h.UserUseCase.GetProfileByID(uuid.MustParse(userIDStr.(string)))
	if err != nil || project.OwnerID != profile.ID {
		utils.SetFlashMessage(c, utils.FlashError, "You don't have permission to delete this project")
		c.Redirect(http.StatusFound, "/projects")
		return
	}

	data := utils.GetTemplateData(c, isAuthenticated)
	data.Object = project
	c.HTML(http.StatusOK, "delete.html", data)
}
