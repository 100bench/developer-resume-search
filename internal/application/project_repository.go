package application

import (
	"devsearch-go/internal/domain"

	"github.com/google/uuid"
)

// ProjectRepository defines the interface for project data operations.
type ProjectRepository interface {
	FindAllProjects(searchQuery string, page, limit int) ([]domain.Project, int64, error)
	FindProjectByID(id uuid.UUID) (*domain.Project, error)
	CreateProject(project *domain.Project) error
	UpdateProject(project *domain.Project) error
	DeleteProject(id uuid.UUID) error
	FindOrCreateTag(tagName string) (*domain.Tag, error)
	AssociateTagWithProject(project *domain.Project, tag *domain.Tag) error
	ClearProjectTags(project *domain.Project) error
}
