package infrastructure

import (
	"devsearch-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormProjectRepository implements the application.ProjectRepository interface using GORM.
type GormProjectRepository struct {
	DB *gorm.DB
}

// FindAllProjects retrieves all projects with optional search and pagination.
func (r *GormProjectRepository) FindAllProjects(searchQuery string, page, limit int) ([]domain.Project, int64, error) {
	var projects []domain.Project
	query := r.DB.Preload("Owner").Preload("Tags")

	if searchQuery != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	var totalProjects int64
	query.Model(&domain.Project{}).Count(&totalProjects)

	offset := (page - 1) * limit
	err := query.Order("vote_ratio DESC, vote_total DESC, title ASC").Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}
	return projects, totalProjects, nil
}

// FindProjectByID retrieves a single project by its ID.
func (r *GormProjectRepository) FindProjectByID(id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	if err := r.DB.Preload("Owner").Preload("Tags").Preload("Reviews.Owner").First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject creates a new project.
func (r *GormProjectRepository) CreateProject(project *domain.Project) error {
	return r.DB.Create(project).Error
}

// UpdateProject updates an existing project.
func (r *GormProjectRepository) UpdateProject(project *domain.Project) error {
	return r.DB.Save(project).Error
}

// DeleteProject deletes a project by its ID.
func (r *GormProjectRepository) DeleteProject(id uuid.UUID) error {
	return r.DB.Delete(&domain.Project{}, "id = ?", id).Error
}

// FindOrCreateTag finds a tag by name or creates a new one if it doesn't exist.
func (r *GormProjectRepository) FindOrCreateTag(tagName string) (*domain.Tag, error) {
	var tag domain.Tag
	if err := r.DB.Where("name = ?", tagName).FirstOrCreate(&tag, domain.Tag{Name: tagName}).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// AssociateTagWithProject associates a tag with a project.
func (r *GormProjectRepository) AssociateTagWithProject(project *domain.Project, tag *domain.Tag) error {
	return r.DB.Model(project).Association("Tags").Append(tag)
}

// ClearProjectTags clears all tags associated with a project.
func (r *GormProjectRepository) ClearProjectTags(project *domain.Project) error {
	return r.DB.Model(project).Association("Tags").Clear()
}
