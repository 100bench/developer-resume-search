package application

import (
	"devsearch-go/internal/domain"

	"github.com/google/uuid"
)

// ProjectUseCase defines the business logic for projects.
type ProjectUseCase struct {
	ProjectRepo ProjectRepository
}

// NewProjectUseCase creates a new ProjectUseCase.
func NewProjectUseCase(projectRepo ProjectRepository) *ProjectUseCase {
	return &ProjectUseCase{
		ProjectRepo: projectRepo,
	}
}

// GetProjects retrieves all projects with optional search and pagination.
func (uc *ProjectUseCase) GetProjects(searchQuery string, page, limit int) ([]domain.Project, int64, error) {
	return uc.ProjectRepo.FindAllProjects(searchQuery, page, limit)
}

// GetProjectByID retrieves a single project by its ID.
func (uc *ProjectUseCase) GetProjectByID(id uuid.UUID) (*domain.Project, error) {
	return uc.ProjectRepo.FindProjectByID(id)
}

// CreateProject creates a new project, handling tags.
func (uc *ProjectUseCase) CreateProject(project *domain.Project, tagNames []string) error {
	if err := uc.ProjectRepo.CreateProject(project); err != nil {
		return err
	}

	for _, tagName := range tagNames {
		tag, err := uc.ProjectRepo.FindOrCreateTag(tagName)
		if err != nil {
			// Log the error but continue to create the project without this tag
			// In a real application, you might want more robust error handling
			continue
		}
		if err := uc.ProjectRepo.AssociateTagWithProject(project, tag); err != nil {
			// Log the error but continue
			continue
		}
	}
	return nil
}

// UpdateProject updates an existing project, handling tags.
func (uc *ProjectUseCase) UpdateProject(project *domain.Project, tagNames []string) error {
	// Clear existing tags
	if err := uc.ProjectRepo.ClearProjectTags(project); err != nil {
		// Log error but continue with project update
	}

	if err := uc.ProjectRepo.UpdateProject(project); err != nil {
		return err
	}

	// Add new tags
	for _, tagName := range tagNames {
		tag, err := uc.ProjectRepo.FindOrCreateTag(tagName)
		if err != nil {
			// Log the error but continue
			continue
		}
		if err := uc.ProjectRepo.AssociateTagWithProject(project, tag); err != nil {
			// Log the error but continue
			continue
		}
	}
	return nil
}

// DeleteProject deletes a project by its ID.
func (uc *ProjectUseCase) DeleteProject(id uuid.UUID) error {
	return uc.ProjectRepo.DeleteProject(id)
}
